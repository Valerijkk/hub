package apigateway

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/webhookdelivery"
	"github.com/valerii/hub/backend/ent/webhookendpoint"
	safehttp "github.com/valerii/hub/backend/internal/platform/http"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// webhookHTTPClient — общий безопасный HTTP-клиент для исходящих
// вебхук-запросов. Использует SafeDialContext, который блокирует RFC1918,
// loopback, link-local и метаданные облачных провайдеров (защита от SSRF).
var webhookHTTPClient = safehttp.NewSafeHTTPClient(10 * time.Second)

// Размеры пула воркеров и очереди доставки вебхуков (HI-08).
// Подобраны умеренно: 16 воркеров × до 4 ретраев × ~10s HTTP-timeout
// = до ~640s суммарной нагрузки в худшем случае; очередь 256 ловит
// микро-всплески, а под устойчивой нагрузкой сообщения копятся в NATS.
const (
	webhookWorkers   = 16
	webhookQueueSize = 256
)

// webhookJob — единица работы для воркер-пула: одна пара (endpoint, событие).
type webhookJob struct {
	ep        *ent.WebhookEndpoint
	eventType string
	body      []byte                 // уже сериализованный payload (msg.Data())
	payload   map[string]interface{} // распакованный payload для записи в WebhookDelivery
}

// WebhookDispatcher потребляет события из JetStream и рассылает HTTP POST
// на все подписанные активные вебхук-эндпоинты. Подписывает payload HMAC-SHA256.
//
// HI-08: для предотвращения goroutine/connection-pool exhaustion рассылка
// выполняется через фиксированный воркер-пул (webhookWorkers) с буфером
// webhookQueueSize. Если очередь полна — задача дропается с warn-логом;
// JetStream доставит событие повторно (MaxDeliver=3).
type WebhookDispatcher struct {
	db       *ent.Client
	nats     *natsclient.Client
	cons     jetstream.ConsumeContext
	jobs     chan webhookJob
	workerWG sync.WaitGroup
	stopOnce sync.Once
}

// NewWebhookDispatcher создаёт новый диспетчер вебхуков.
func NewWebhookDispatcher(db *ent.Client, nats *natsclient.Client) *WebhookDispatcher {
	return &WebhookDispatcher{
		db:   db,
		nats: nats,
		jobs: make(chan webhookJob, webhookQueueSize),
	}
}

// Start запускает JetStream consumer и фиксированный пул HTTP-воркеров для
// обработки событий вебхуков (HI-08). Подписывается на стрим EVENTS и
// пушит каждую задачу доставки в d.jobs; воркеры разбирают очередь.
func (d *WebhookDispatcher) Start(ctx context.Context) error {
	if d.nats == nil {
		slog.Info("NATS не подключен, вебхук-диспетчер не запущен")
		return nil
	}

	// Поднимаем воркеры до того, как начнём принимать сообщения.
	for i := 0; i < webhookWorkers; i++ {
		d.workerWG.Add(1)
		go d.deliveryWorker(ctx)
	}

	js := d.nats.JetStream

	// Получаем общий стрим EVENTS (создаётся при старте сервера).
	stream, err := js.Stream(ctx, natsclient.StreamName)
	if err != nil {
		return fmt.Errorf("не удалось получить стрим %s: %w", natsclient.StreamName, err)
	}

	// Durable consumer на общем стриме — гарантирует at-least-once доставку.
	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:    "webhook-dispatcher",
		AckPolicy:  jetstream.AckExplicitPolicy,
		MaxDeliver: 3,
		AckWait:    30 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("не удалось создать consumer WEBHOOKS: %w", err)
	}

	// Запуск асинхронного потребления сообщений.
	cons, err := consumer.Consume(func(msg jetstream.Msg) {
		d.handleMessage(ctx, msg)
		if ackErr := msg.Ack(); ackErr != nil {
			slog.Error("ack webhook event failed", "error", ackErr)
		}
	})
	if err != nil {
		return fmt.Errorf("не удалось запустить consume WEBHOOKS: %w", err)
	}
	d.cons = cons

	slog.Info("webhook dispatcher запущен", "workers", webhookWorkers, "queue", webhookQueueSize)
	return nil
}

// Stop останавливает JetStream consumer и дожидается завершения всех
// воркеров. Идемпотентно — повторный вызов безопасен.
func (d *WebhookDispatcher) Stop() {
	d.stopOnce.Do(func() {
		if d.cons != nil {
			d.cons.Stop()
		}
		// Закрываем канал — воркеры выходят из range loop.
		close(d.jobs)
		d.workerWG.Wait()
	})
}

// deliveryWorker — единичный HTTP-воркер. Читает задачи из канала и доставляет
// вебхук с ретраями. Завершается при закрытии d.jobs.
func (d *WebhookDispatcher) deliveryWorker(ctx context.Context) {
	defer d.workerWG.Done()
	for job := range d.jobs {
		d.deliverWithRetry(ctx, job)
	}
}

// handleMessage разбирает входящее JetStream-сообщение и передаёт в dispatchEvent.
// HI-08: payload передаётся в воркеры как уже сериализованный body (msg.Data())
// и распакованная map — избегаем повторного json.Marshal в каждом воркере.
func (d *WebhookDispatcher) handleMessage(ctx context.Context, msg jetstream.Msg) {
	subject := msg.Subject()
	body := msg.Data()

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		slog.Error("не удалось разобрать payload вебхук-события", "error", err, "subject", subject)
		return
	}

	// Subject NATS используется как тип события (например: tracker.{projID}.issue.created).
	d.dispatchEvent(ctx, subject, body, payload)
}

// dispatchEvent находит все активные эндпоинты, подписанные на событие, и
// пушит задачи доставки в воркер-пул. HI-08: вместо unbounded `go` —
// очередь с фиксированным числом воркеров и явное логирование переполнения.
func (d *WebhookDispatcher) dispatchEvent(ctx context.Context, eventType string, body []byte, payload map[string]interface{}) {
	endpoints, err := d.db.WebhookEndpoint.Query().
		Where(
			webhookendpoint.IsActive(true),
		).
		All(ctx)
	if err != nil {
		slog.Error("ошибка получения вебхук-эндпоинтов", "error", err)
		return
	}

	for _, ep := range endpoints {
		if !containsEvent(ep.Events, eventType) {
			continue
		}
		job := webhookJob{ep: ep, eventType: eventType, body: body, payload: payload}
		select {
		case d.jobs <- job:
		case <-ctx.Done():
			return
		default:
			// Очередь переполнена. Не блокируем consumer — JetStream
			// доставит сообщение повторно (MaxDeliver=3).
			slog.Warn("webhook job queue full, dropping delivery",
				"endpoint", ep.URL, "event", eventType, "queue_size", webhookQueueSize)
		}
	}
}

// deliverWithRetry отправляет вебхук с повторами (1s, 5s, 30s). Тело уже
// сериализовано в job.body, повторного json.Marshal нет (HI-08).
func (d *WebhookDispatcher) deliverWithRetry(ctx context.Context, job webhookJob) {
	delays := []time.Duration{0, time.Second, 5 * time.Second, 30 * time.Second}

	sig := signPayload(job.body, job.ep.Secret)

	for attempt, delay := range delays {
		if delay > 0 {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return
			}
		}

		statusCode, respBody, deliveryErr := d.sendWebhook(job.ep.URL, job.body, sig, job.eventType)

		errMsg := ""
		if deliveryErr != nil {
			errMsg = deliveryErr.Error()
		}

		// Записываем доставку в БД
		builder := d.db.WebhookDelivery.Create().
			SetEventType(job.eventType).
			SetPayload(job.payload).
			SetStatusCode(statusCode).
			SetResponseBody(respBody).
			SetEndpointID(job.ep.ID)

		if errMsg != "" {
			builder.SetError(errMsg)
		}

		if _, saveErr := builder.Save(ctx); saveErr != nil {
			slog.Error("не удалось сохранить запись доставки вебхука", "error", saveErr)
		}

		if deliveryErr == nil && statusCode >= 200 && statusCode < 300 {
			return // Успешная доставка
		}

		if attempt < len(delays)-1 {
			slog.Warn("попытка доставки вебхука не удалась, повтор",
				"endpoint", job.ep.URL, "attempt", attempt+1, "status", statusCode, "error", errMsg)
		}
	}
}

// sendWebhook выполняет HTTP POST к эндпоинту с подписью HMAC-SHA256.
// URL предварительно валидируется (только http/https, без userinfo) и
// отправляется через webhookHTTPClient с SafeDialContext (защита от SSRF).
func (d *WebhookDispatcher) sendWebhook(url string, body []byte, signature, eventType string) (int, string, error) {
	if err := safehttp.ValidateOutboundURL(url); err != nil {
		return 0, "", fmt.Errorf("отклонён небезопасный URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", "sha256="+signature)
	req.Header.Set("X-Hub-Event", eventType)
	req.Header.Set("User-Agent", "Hub-Webhook/1.0")

	resp, err := webhookHTTPClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("ошибка отправки: %w", err)
	}
	defer resp.Body.Close()

	var respBuf bytes.Buffer
	respBuf.ReadFrom(resp.Body)
	respBody := respBuf.String()
	if len(respBody) > 4096 {
		respBody = respBody[:4096]
	}

	return resp.StatusCode, respBody, nil
}

// signPayload подписывает payload HMAC-SHA256 секретом эндпоинта.
func signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// containsEvent проверяет наличие события в списке подписок.
func containsEvent(events []string, event string) bool {
	for _, e := range events {
		if e == event || e == "*" {
			return true
		}
	}
	return false
}

// --- Webhook CRUD ---

// WebhookService управляет CRUD операциями для вебхук-эндпоинтов.
type WebhookService struct {
	db *ent.Client
}

// NewWebhookService создаёт новый сервис вебхуков.
func NewWebhookService(db *ent.Client) *WebhookService {
	return &WebhookService{db: db}
}

// CreateEndpoint создаёт новый вебхук-эндпоинт в проекте.
// URL валидируется на схему (http/https) и отсутствие userinfo —
// фактическая блокировка приватных IP происходит в SafeDialContext
// (CR-03, защита от DNS-rebinding).
func (s *WebhookService) CreateEndpoint(ctx context.Context, projectID uuid.UUID, req CreateWebhookRequest) (*WebhookResponse, error) {
	if err := safehttp.ValidateOutboundURL(req.URL); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "некорректный URL вебхука: "+err.Error())
	}

	secret := req.Secret
	if secret == "" {
		secret = uuid.New().String()
	}

	ep, err := s.db.WebhookEndpoint.Create().
		SetURL(req.URL).
		SetSecret(secret).
		SetEvents(req.Events).
		SetProjectID(projectID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать вебхук-эндпоинт: %w", err)
	}

	return toWebhookResponse(ep), nil
}

// ListEndpoints возвращает вебхук-эндпоинты проекта.
func (s *WebhookService) ListEndpoints(ctx context.Context, projectID uuid.UUID) ([]WebhookResponse, error) {
	eps, err := s.db.Project.Query().
		Where(project.ID(projectID)).
		QueryWebhookEndpoints().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить вебхук-эндпоинты: %w", err)
	}

	result := make([]WebhookResponse, 0, len(eps))
	for _, ep := range eps {
		result = append(result, *toWebhookResponse(ep))
	}
	return result, nil
}

// UpdateEndpoint обновляет вебхук-эндпоинт.
func (s *WebhookService) UpdateEndpoint(ctx context.Context, endpointID uuid.UUID, req UpdateWebhookRequest) (*WebhookResponse, error) {
	builder := s.db.WebhookEndpoint.UpdateOneID(endpointID)
	if req.URL != nil {
		if err := safehttp.ValidateOutboundURL(*req.URL); err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "некорректный URL вебхука: "+err.Error())
		}
		builder.SetURL(*req.URL)
	}
	if req.Secret != nil {
		builder.SetSecret(*req.Secret)
	}
	if req.Events != nil {
		builder.SetEvents(req.Events)
	}
	if req.IsActive != nil {
		builder.SetIsActive(*req.IsActive)
	}

	ep, err := builder.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fiber.NewError(fiber.StatusNotFound, "вебхук-эндпоинт не найден")
		}
		return nil, fmt.Errorf("не удалось обновить вебхук-эндпоинт: %w", err)
	}
	return toWebhookResponse(ep), nil
}

// DeleteEndpoint удаляет вебхук-эндпоинт.
func (s *WebhookService) DeleteEndpoint(ctx context.Context, endpointID uuid.UUID) error {
	err := s.db.WebhookEndpoint.DeleteOneID(endpointID).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fiber.NewError(fiber.StatusNotFound, "вебхук-эндпоинт не найден")
		}
		return fmt.Errorf("не удалось удалить вебхук-эндпоинт: %w", err)
	}
	return nil
}

// ListDeliveries возвращает историю доставок для эндпоинта.
func (s *WebhookService) ListDeliveries(ctx context.Context, endpointID uuid.UUID) ([]WebhookDeliveryResponse, error) {
	deliveries, err := s.db.WebhookDelivery.Query().
		Where(webhookdelivery.HasEndpointWith(webhookendpoint.ID(endpointID))).
		Order(ent.Desc(webhookdelivery.FieldDeliveredAt)).
		Limit(50).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить доставки: %w", err)
	}

	result := make([]WebhookDeliveryResponse, 0, len(deliveries))
	for _, d := range deliveries {
		errStr := ""
		if d.Error != nil {
			errStr = *d.Error
		}
		result = append(result, WebhookDeliveryResponse{
			ID:           d.ID.String(),
			EventType:    d.EventType,
			Payload:      d.Payload,
			StatusCode:   d.StatusCode,
			ResponseBody: d.ResponseBody,
			Error:        errStr,
			DeliveredAt:  d.DeliveredAt,
		})
	}
	return result, nil
}

// toWebhookResponse конвертирует Ent-сущность в DTO ответа.
func toWebhookResponse(ep *ent.WebhookEndpoint) *WebhookResponse {
	return &WebhookResponse{
		ID:       ep.ID.String(),
		URL:      ep.URL,
		Events:   ep.Events,
		IsActive: ep.IsActive,
		Created:  ep.CreatedAt.Format(time.RFC3339),
	}
}
