package notification

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/pushsubscription"
	"github.com/valerii/hub/backend/internal/platform/config"
)

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
type PushSender struct {
	db     *ent.Client
	cfg    config.WebPushConfig
	queue  chan pushMessage
	cancel context.CancelFunc
}

// NewPushSender создаёт нового отправителя push-уведомлений.
// Если VAPID-ключи не заданы — возвращает nil (nil-safe: Send на nil sender — no-op).
func NewPushSender(db *ent.Client, cfg config.WebPushConfig) *PushSender {
	if cfg.VAPIDPublicKey == "" || cfg.VAPIDPrivateKey == "" {
		slog.Info("VAPID ключи не заданы, push-уведомления отключены")
		return nil
	}

	return &PushSender{
		db:    db,
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
				time.Sleep(50 * time.Millisecond)
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
	subs, err := p.db.PushSubscription.Query().
		Where(pushsubscription.UserID(msg.userID)).
		All(ctx)
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
func (p *PushSender) sendToSubscription(ctx context.Context, sub *ent.PushSubscription, payload []byte) {
	s := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dhKey,
			Auth:   sub.AuthKey,
		},
	}

	resp, err := webpush.SendNotification(payload, s, &webpush.Options{
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
		if err := p.db.PushSubscription.DeleteOne(sub).Exec(ctx); err != nil {
			slog.Error("ошибка удаления мёртвой подписки", "error", err)
		}
		return
	}

	if resp.StatusCode >= 400 {
		slog.Warn("push-уведомление не доставлено",
			"status", resp.StatusCode, "endpoint", sub.Endpoint, "user_id", sub.UserID)
	}
}

// SaveSubscription сохраняет или обновляет подписку на push-уведомления.
// При существующем endpoint — обновляет ключи и привязку к пользователю.
func (p *PushSender) SaveSubscription(ctx context.Context, userID uuid.UUID, endpoint, p256dhKey, authKey, userAgent string) error {
	if p == nil {
		return nil
	}

	// Проверяем, существует ли подписка с таким endpoint.
	existing, err := p.db.PushSubscription.Query().
		Where(pushsubscription.Endpoint(endpoint)).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return err
	}

	if existing != nil {
		// Обновляем существующую подписку (ключи могут измениться при переподписке).
		_, err = existing.Update().
			SetUserID(userID).
			SetP256dhKey(p256dhKey).
			SetAuthKey(authKey).
			SetUserAgent(userAgent).
			Save(ctx)
		return err
	}

	// Создаём новую подписку.
	_, err = p.db.PushSubscription.Create().
		SetUserID(userID).
		SetEndpoint(endpoint).
		SetP256dhKey(p256dhKey).
		SetAuthKey(authKey).
		SetUserAgent(userAgent).
		Save(ctx)
	return err
}

// RemoveSubscription удаляет подписку по endpoint.
func (p *PushSender) RemoveSubscription(ctx context.Context, endpoint string) error {
	if p == nil {
		return nil
	}

	_, err := p.db.PushSubscription.Delete().
		Where(pushsubscription.Endpoint(endpoint)).
		Exec(ctx)
	return err
}

// VAPIDPublicKey возвращает публичный VAPID-ключ для передачи на фронтенд.
func (p *PushSender) VAPIDPublicKey() string {
	if p == nil {
		return ""
	}
	return p.cfg.VAPIDPublicKey
}
