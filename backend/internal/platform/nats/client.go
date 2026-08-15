// Package nats управляет подключением к NATS и JetStream-стримами.
package nats

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// StreamName — единый JetStream stream для всех модулей.
const StreamName = "EVENTS"

// StreamSubjects — канонический список subjects для EVENTS (ME-07).
// EnsureStreams и unit-тесты сверяются только с этим слайсом (A2W4-2-003).
var StreamSubjects = []string{
	"project.>",
	"tracker.>",
	"messenger.channel.>",
	"tms.>",
	"calls.channel.>",
	"calls.recording.>",
	"auth.session.>",
	"sprint.>",
	"notification.>",
	"audit.>", // C-W19-002 / TASK-178: GDPR Mongo erase retry (audit.gdpr.mongo_erase)
	"calendar.>",
}

// WebhookFilterSubjects — JetStream filter для durable consumer webhook-dispatcher.
// Только subject'ы, которые могут триггерить исходящие webhook (см. UI AVAILABLE_EVENTS).
// Исключает notification.*, auth.session.*, calls.* — они не доставляются наружу (APIGW-03).
// A3W4-5-003 (trim): sprint.> убран — sprint.completed нет в allowlist и нет
// NATS producer sprint.*; вернуть вместе с producer + allowlist.
var WebhookFilterSubjects = []string{
	"project.*.issue.>",
	"tracker.*.issue.>",
	"project.created",
	"tms.run.completed.>",
	"tms.testrun.completed",
	"messenger.channel.*.message_sent",
}

// Client оборачивает NATS-соединение с поддержкой JetStream.
type Client struct {
	Conn      *nats.Conn
	JetStream jetstream.JetStream
}

// NewClient подключается к NATS и инициализирует JetStream.
func NewClient(ctx context.Context, url string) (*Client, error) {
	nc, err := nats.Connect(url,
		nats.Name("hub-backend"),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, err
	}

	slog.Info("connected to NATS", "url", redactNATSURL(url))
	return &Client{Conn: nc, JetStream: js}, nil
}

// redactNATSURL скрывает userinfo в логах (A14-003 credentials in URL).
func redactNATSURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return raw
	}
	host := u.Host
	if host == "" {
		return "nats://***"
	}
	return u.Scheme + "://***:***@" + host
}

// Close закрывает NATS-соединение через Drain (W4-A2-006 / P3-BE-12):
// in-flight publishes успевают уйти; при таймауте или если Drain не закрыл — force Close.
func (c *Client) Close() {
	if c == nil || c.Conn == nil {
		return
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = c.Conn.Drain()
	}()
	select {
	case <-done:
	case <-time.After(natsDrainTimeout):
	}
	if !c.Conn.IsClosed() {
		c.Conn.Close()
	}
}

// natsDrainTimeout — сколько ждать Drain перед force Close на shutdown.
const natsDrainTimeout = 5 * time.Second

// ForceClose закрывает соединение без Drain (hubtest / аварийный путь).
func (c *Client) ForceClose() {
	if c == nil || c.Conn == nil {
		return
	}
	c.Conn.Close()
}

// Publish публикует событие в core NATS (эфемерное, fire-and-forget).
// Используется для: индикаторов набора, присутствия, обновлений UI в реальном времени.
func (c *Client) Publish(subject string, data []byte) error {
	return c.Conn.Publish(subject, data)
}

// publishFlushTimeout — сколько ждать подтверждения сервером после Publish
// (P1-BE-24). Без Flush сообщение может остаться только в клиентском буфере.
const publishFlushTimeout = 5 * time.Second

// afterCorePublishHook — только для hubtest: вставка между Publish и Flush,
// чтобы симулировать drop соединения до подтверждения сервером.
var afterCorePublishHook func()

// PublishMsg публикует подготовленное сообщение (в т.ч. с headers) в NATS и
// ждёт FlushTimeout / серверный ack буфера. Нужен outbox-worker'у: он несёт
// traceparent и event_id в headers и помечает status=published только после
// успешного возврата (P1-BE-24 — иначе crash между Publish и Flush теряет событие).
func (c *Client) PublishMsg(msg *nats.Msg) error {
	if err := c.Conn.PublishMsg(msg); err != nil {
		return err
	}
	if afterCorePublishHook != nil {
		afterCorePublishHook()
	}
	return c.Conn.FlushTimeout(publishFlushTimeout)
}

// FlushTimeout прокидывает nats.Conn.FlushTimeout (подтверждение буфера сервером).
func (c *Client) FlushTimeout(timeout time.Duration) error {
	return c.Conn.FlushTimeout(timeout)
}

// PublishMsgAck публикует через JetStream и ждёт PubAck (durable path).
// Использовать, когда subject входит в stream EVENTS.
func (c *Client) PublishMsgAck(ctx context.Context, msg *nats.Msg) error {
	ack, err := c.JetStream.PublishMsg(ctx, msg)
	if afterCorePublishHook != nil {
		afterCorePublishHook()
	}
	if err != nil {
		return err
	}
	if ack == nil {
		return errors.New("nats: empty PubAck")
	}
	return nil
}

// PublishWithTraceContext публикует данные в core NATS, перенося trace-context
// из ctx в NATS headers (W3C traceparent/tracestate). Хелпер для прямых
// publish в обход outbox, чтобы трейс не рвался на асинхронной границе.
// A3W4-2-003: Flush через PublishMsg — тот же контракт, что P1-BE-24
// (буфер подтверждён сервером; hook между Publish и Flush для hubtest).
func (c *Client) PublishWithTraceContext(ctx context.Context, subject string, data []byte) error {
	msg := nats.NewMsg(subject)
	msg.Data = data
	InjectTraceContext(ctx, msg.Header)
	return c.PublishMsg(msg)
}

// EnsureStreams создаёт единый JetStream stream EVENTS, агрегирующий события
// всех модулей. Каждый consumer (audit, webhook, notification) создаёт свой
// durable consumer на этом стриме.
//
// ME-07: явно перечисляем durable subjects.
//   - notification.> — уведомления должны переживать рестарт, иначе бэкенд
//     теряет очередь между рестартами и pending события не доедут.
//   - calls.recording.> — события готовности/завершения записи звонка
//     должны быть durable, чтобы webhook-диспатчер и аудит их видели.
//   - audit.> — C-W19-002 / TASK-178: compensating GDPR Mongo erase retry
//     (audit.gdpr.mongo_erase) после post-Commit fail eraser.
//   - calls.chat.> — намеренно НЕ durable: чат внутри звонка — это
//     real-time, реплей не нужен (см. handleCallChatEvent в ws/bridge.go).
func (c *Client) EnsureStreams(ctx context.Context) error {
	_, err := c.JetStream.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      StreamName,
		Subjects:  append([]string(nil), StreamSubjects...),
		Retention: jetstream.InterestPolicy,
		MaxAge:    7 * 24 * time.Hour,
	})
	if err != nil {
		return err
	}
	slog.Info("JetStream stream ensured", "stream", StreamName)
	return nil
}

// PublishJSON сериализует payload в JSON и публикует в NATS (fire-and-forget).
// Безопасен при nil receiver — пропускает публикацию с debug-логом.
func (c *Client) PublishJSON(subject string, payload interface{}) {
	if c == nil {
		slog.Debug("NATS не подключён, пропускаем публикацию", "subject", subject)
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("не удалось сериализовать событие", "subject", subject, "error", err)
		return
	}
	if err := c.Conn.Publish(subject, data); err != nil {
		slog.Error("не удалось опубликовать событие", "subject", subject, "error", err)
	}
}

// PublishJSONContext — PublishJSON с переносом trace-context из ctx в NATS
// headers. Использовать для событий, порождённых в рамках запроса, чтобы
// consumer продолжил трейс. Fire-and-forget, безопасен при nil receiver.
func (c *Client) PublishJSONContext(ctx context.Context, subject string, payload interface{}) {
	if c == nil {
		slog.Debug("NATS не подключён, пропускаем публикацию", "subject", subject)
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		slog.ErrorContext(ctx, "не удалось сериализовать событие", "subject", subject, "error", err)
		return
	}
	if err := c.PublishWithTraceContext(ctx, subject, data); err != nil {
		slog.ErrorContext(ctx, "не удалось опубликовать событие", "subject", subject, "error", err)
	}
}

// Subscribe подписывается на core NATS subject (эфемерная подписка).
func (c *Client) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return c.Conn.Subscribe(subject, handler)
}
