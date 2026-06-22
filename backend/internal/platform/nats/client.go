// Package nats управляет подключением к NATS и JetStream-стримами.
package nats

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// StreamName — единый JetStream stream для всех модулей.
const StreamName = "EVENTS"

// WebhookFilterSubjects — JetStream filter для durable consumer webhook-dispatcher.
// Только subject'ы, которые могут триггерить исходящие webhook (см. UI AVAILABLE_EVENTS).
// Исключает notification.*, auth.session.*, calls.* — они не доставляются наружу (APIGW-03).
var WebhookFilterSubjects = []string{
	"project.*.issue.>",
	"tracker.*.issue.>",
	"project.created",
	"tms.run.completed.>",
	"tms.testrun.completed",
	"sprint.>",
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

	slog.Info("connected to NATS", "url", url)
	return &Client{Conn: nc, JetStream: js}, nil
}

// Close закрывает NATS-соединение.
func (c *Client) Close() {
	c.Conn.Close()
}

// Publish публикует событие в core NATS (эфемерное, fire-and-forget).
// Используется для: индикаторов набора, присутствия, обновлений UI в реальном времени.
func (c *Client) Publish(subject string, data []byte) error {
	return c.Conn.Publish(subject, data)
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
//   - calls.chat.> — намеренно НЕ durable: чат внутри звонка — это
//     real-time, реплей не нужен (см. handleCallChatEvent в ws/bridge.go).
func (c *Client) EnsureStreams(ctx context.Context) error {
	_, err := c.JetStream.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name: StreamName,
		Subjects: []string{
			"project.>",
			"tracker.>",
			"messenger.channel.>",
			"tms.>",
			"calls.channel.>",
			"calls.recording.>",
			"auth.session.>",
			"sprint.>",
			"notification.>",
		},
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

// Subscribe подписывается на core NATS subject (эфемерная подписка).
func (c *Client) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return c.Conn.Subscribe(subject, handler)
}
