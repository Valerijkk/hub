package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/platform/config"
	safehttp "github.com/valerii/hub/backend/internal/platform/http"
)

// A2-005: чужой endpoint нельзя перепривязать/удалить.
var (
	ErrSubscriptionNotFound  = errors.New("push subscription not found")
	ErrSubscriptionForbidden = errors.New("push subscription access denied")
	// ErrInvalidPushEndpoint — endpoint не проходит NF-07 ValidateOutboundURL (→ HTTP 400).
	ErrInvalidPushEndpoint = errors.New("invalid push endpoint")
)

// pushHTTPClient — SafeHTTPClient с SafeDialContext, блокирующим SSRF на
// RFC1918/loopback/link-local IP (см. platform/http/safedialer.go).
// Используется в webpush.Options.HTTPClient вместо http.DefaultClient,
// чтобы атакующий не мог зарегистрировать subscription с
// endpoint = http://169.254.169.254/... и заставить бэк выстрелить во
// внутреннюю сеть. SEC-H3.
var pushHTTPClient = safehttp.NewSafeHTTPClient(10 * time.Second)

// PushSub — Web Push подписка без зависимости от ent (ASN-001).
type PushSub struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Endpoint  string
	P256dhKey string
	AuthKey   string
}

// PushSubscriptionStore — CRUD подписок для PushSender (TZ: без *ent.Client).
type PushSubscriptionStore interface {
	ListPushSubscriptions(ctx context.Context, userID uuid.UUID) ([]PushSub, error)
	DeletePushSubscription(ctx context.Context, id uuid.UUID) error
	UpdateOwnPushSubscription(ctx context.Context, userID uuid.UUID, endpoint, p256dhKey, authKey, userAgent string) (int, error)
	PushEndpointExists(ctx context.Context, endpoint string) (bool, error)
	CreatePushSubscription(ctx context.Context, userID uuid.UUID, endpoint, p256dhKey, authKey, userAgent string) error
	DeleteOwnPushSubscription(ctx context.Context, userID uuid.UUID, endpoint string) (int, error)
}

// pushMessage — внутреннее представление push-уведомления для очереди.
type pushMessage struct {
	userID uuid.UUID
	title  string
	body   string
	url    string
}

// pushPayload — JSON-тело push-уведомления, отправляемого в браузер.
type pushPayload struct {
	Title string `json:"title"`
	Body  string `json:"body,omitempty"`
	URL   string `json:"url,omitempty"`
}

// PushSender отправляет Web Push уведомления через VAPID.
// Использует буферизованный канал и воркер-горутину (аналогично TelegramSender).
// ASN-001: хранит PushSubscriptionStore, не *ent.Client.
type PushSender struct {
	store  PushSubscriptionStore
	cfg    config.WebPushConfig
	queue  chan pushMessage
	cancel context.CancelFunc
}

// NewPushSender создаёт нового отправителя push-уведомлений.
// Если VAPID-ключи не заданы — возвращает nil (nil-safe: Send на nil sender — no-op).
func NewPushSender(store PushSubscriptionStore, cfg config.WebPushConfig) *PushSender {
	if cfg.VAPIDPublicKey == "" || cfg.VAPIDPrivateKey == "" {
		slog.Info("VAPID ключи не заданы, push-уведомления отключены")
		return nil
	}

	return &PushSender{
		store: store,
		cfg:   cfg,
		queue: make(chan pushMessage, 200),
	}
}

// Start запускает воркер отправки push-уведомлений.
func (p *PushSender) Start(ctx context.Context) {
	if p == nil {
		return
	}

	workerCtx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	go func() {
		for {
			select {
			case <-workerCtx.Done():
				return
			case msg := <-p.queue:
				p.sendToUser(workerCtx, msg)
				// Rate limiting: 50мс между отправками (~20 сообщений/сек).
				// ctx-aware — при shutdown воркер прерывается мгновенно,
				// не дожидаясь оставшихся 50мс × queue depth.
				select {
				case <-workerCtx.Done():
					return
				case <-time.After(50 * time.Millisecond):
				}
			}
		}
	}()

	slog.Info("push sender запущен")
}

// Stop останавливает воркер отправки.
func (p *PushSender) Stop() {
	if p == nil || p.cancel == nil {
		return
	}
	p.cancel()
}

// Send добавляет push-уведомление в очередь отправки.
// Nil-safe: вызов на nil sender — no-op.
func (p *PushSender) Send(userID uuid.UUID, title, body, url string) {
	if p == nil {
		return
	}

	select {
	case p.queue <- pushMessage{userID: userID, title: title, body: body, url: url}:
	default:
		slog.Warn("очередь push-уведомлений переполнена, сообщение отброшено")
	}
}

// sendToUser находит все подписки пользователя и отправляет push каждому устройству.
func (p *PushSender) sendToUser(ctx context.Context, msg pushMessage) {
	if p.store == nil {
		return
	}
	subs, err := p.store.ListPushSubscriptions(ctx, msg.userID)
	if err != nil {
		slog.Error("ошибка загрузки push-подписок", "error", err, "user_id", msg.userID)
		return
	}

	if len(subs) == 0 {
		return
	}

	payload, err := json.Marshal(pushPayload{
		Title: msg.title,
		Body:  msg.body,
		URL:   msg.url,
	})
	if err != nil {
		slog.Error("ошибка сериализации push payload", "error", err)
		return
	}

	for _, sub := range subs {
		p.sendToSubscription(ctx, sub, payload)
	}
}

// sendToSubscription отправляет push одной подписке.
// При 404/410 ответе удаляет мёртвую подписку из БД.
//
// SEC-H3: defense-in-depth — endpoint валидируется ещё раз перед отправкой,
// даже если подписка прошла регистрацию (на случай, если запись попала в БД
// мимо SaveSubscription — например, через миграцию или legacy-данные).
// Сам исходящий запрос идёт через pushHTTPClient с SafeDialContext —
// это финальный рубеж против DNS-rebinding на приватные IP.
func (p *PushSender) sendToSubscription(ctx context.Context, sub PushSub, payload []byte) {
	if err := safehttp.ValidateOutboundURL(sub.Endpoint); err != nil {
		slog.Warn("отказ отправки push: небезопасный endpoint",
			"error", err, "endpoint", sub.Endpoint, "user_id", sub.UserID)
		if p.store != nil {
			if err := p.store.DeletePushSubscription(ctx, sub.ID); err != nil {
				slog.Error("ошибка удаления небезопасной push-подписки", "error", err)
			}
		}
		return
	}

	s := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dhKey,
			Auth:   sub.AuthKey,
		},
	}

	resp, err := webpush.SendNotification(payload, s, &webpush.Options{
		HTTPClient:      pushHTTPClient,
		VAPIDPublicKey:  p.cfg.VAPIDPublicKey,
		VAPIDPrivateKey: p.cfg.VAPIDPrivateKey,
		Subscriber:      p.cfg.VAPIDSubject,
		TTL:             60, // 60 секунд TTL
	})
	if err != nil {
		slog.Error("ошибка отправки push-уведомления",
			"error", err, "endpoint", sub.Endpoint, "user_id", sub.UserID)
		return
	}
	defer resp.Body.Close()

	// Удаляем мёртвую подписку при 404 (Not Found) или 410 (Gone).
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		slog.Info("удаление мёртвой push-подписки",
			"endpoint", sub.Endpoint, "status", resp.StatusCode, "user_id", sub.UserID)
		if p.store != nil {
			if err := p.store.DeletePushSubscription(ctx, sub.ID); err != nil {
				slog.Error("ошибка удаления мёртвой подписки", "error", err)
			}
		}
		return
	}

	if resp.StatusCode >= 400 {
		slog.Warn("push-уведомление не доставлено",
			"status", resp.StatusCode, "endpoint", sub.Endpoint, "user_id", sub.UserID)
	}
}

// SaveSubscription сохраняет или обновляет подписку на push-уведомления.
// Update только для Endpoint AND UserID=caller — без SetUserID (A2-005 hijack).
//
// SEC-H3: endpoint валидируется ValidateOutboundURL (http/https-only, без
// userinfo, непустой host). Это отбрасывает javascript:/file:/data:/gopher:
// схемы и URL с user:pass@host. IP-проверка (RFC1918/loopback/link-local)
// выполняется в момент отправки через SafeDialContext — здесь её делать
// бессмысленно, т.к. DNS может ребиндиться между регистрацией и отправкой.
func (p *PushSender) SaveSubscription(ctx context.Context, userID uuid.UUID, endpoint, p256dhKey, authKey, userAgent string) error {
	if p == nil {
		return nil
	}
	if p.store == nil {
		return fmt.Errorf("push subscription store not configured")
	}

	if err := safehttp.ValidateOutboundURL(endpoint); err != nil {
		// W7-A1-002: sentinel → handler мапит в 400, не 500.
		return fmt.Errorf("%w: %v", ErrInvalidPushEndpoint, err)
	}

	// Mutation: снять UserID из Where → чужой endpoint перезаписывается → FAIL (A2-005).
	n, err := p.store.UpdateOwnPushSubscription(ctx, userID, endpoint, p256dhKey, authKey, userAgent)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	// Endpoint уже занят другим пользователем — без rebind.
	taken, err := p.store.PushEndpointExists(ctx, endpoint)
	if err != nil {
		return err
	}
	if taken {
		return ErrSubscriptionForbidden
	}

	return p.store.CreatePushSubscription(ctx, userID, endpoint, p256dhKey, authKey, userAgent)
}

// RemoveSubscription удаляет подписку по Endpoint AND UserID=caller (A2-005).
func (p *PushSender) RemoveSubscription(ctx context.Context, userID uuid.UUID, endpoint string) error {
	if p == nil {
		return nil
	}
	if p.store == nil {
		return fmt.Errorf("push subscription store not configured")
	}

	// Mutation: снять UserID из Where → чужой unsubscribe проходит → FAIL (A2-005).
	n, err := p.store.DeleteOwnPushSubscription(ctx, userID, endpoint)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrSubscriptionNotFound
	}
	return nil
}

// VAPIDPublicKey возвращает публичный VAPID-ключ для передачи на фронтенд.
func (p *PushSender) VAPIDPublicKey() string {
	if p == nil {
		return ""
	}
	return p.cfg.VAPIDPublicKey
}
