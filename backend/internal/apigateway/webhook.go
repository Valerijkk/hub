package apigateway

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/valerii/hub/backend/ent"
	safehttp "github.com/valerii/hub/backend/internal/platform/http"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// errWebhookQueueFull сигнализирует, что задачу доставки не удалось поставить
// в воркер-очередь (переполнена). Сообщение JetStream в этом случае НЕ
// подтверждается (NAK с задержкой) — событие будет доставлено повторно
// (MaxDeliver=3), что исключает «тихую» потерю данных (APIGW-02).
var errWebhookQueueFull = errors.New("webhook delivery queue full")

// webhookRedeliverDelay — задержка перед повторной доставкой JetStream-
// сообщения, когда очередь воркеров переполнена. Даёт воркерам время
// разгрести очередь до следующей попытки и предотвращает busy-loop.
const webhookRedeliverDelay = 5 * time.Second

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
	// maxWebhookResponseBodyBytes — верхняя граница тела ответа от вебхук-
	// эндпоинта, которая сохраняется в `webhook_delivery.response_body`.
	// Защищает от использования вебхуков как SSRF read primitive (CR-03)
	// и от раздувания таблицы доставок ответом мегабайтного размера.
	maxWebhookResponseBodyBytes = 4096
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
	repo     WebhookRepository
	nats     *natsclient.Client
	cons     jetstream.ConsumeContext
	jobs     chan webhookJob
	workerWG sync.WaitGroup
	inFlight sync.WaitGroup // APIGW-03: in-flight JetStream callbacks before close(jobs)
	stopOnce sync.Once
}

// NewWebhookDispatcher создаёт новый диспетчер вебхуков.
func NewWebhookDispatcher(repo WebhookRepository, nats *natsclient.Client) *WebhookDispatcher {
	return &WebhookDispatcher{
		repo: repo,
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
	// APIGW-03: FilterSubjects ограничивает поток только webhook-релевантными
	// subject'ами — notification/auth/calls не вызывают ListActiveWebhookEndpoints.
	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:         "webhook-dispatcher",
		FilterSubjects:  natsclient.WebhookFilterSubjects,
		AckPolicy:       jetstream.AckExplicitPolicy,
		MaxDeliver:      3,
		AckWait:         30 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("не удалось создать consumer WEBHOOKS: %w", err)
	}

	// Запуск асинхронного потребления сообщений.
	//
	// APIGW-02: подтверждаем (Ack) сообщение ТОЛЬКО после успешной постановки
	// всех задач доставки в воркер-очередь. Если очередь переполнена или
	// произошла транзиентная ошибка — отправляем NAK с задержкой, и JetStream
	// доставит событие повторно (MaxDeliver=3). Раньше Ack вызывался безусловно,
	// из-за чего переполнение очереди приводило к «тихой» потере событий.
	cons, err := consumer.Consume(func(msg jetstream.Msg) {
		d.inFlight.Add(1)
		defer d.inFlight.Done()
		if handleErr := d.handleMessage(ctx, msg); handleErr != nil {
			if nakErr := msg.NakWithDelay(webhookRedeliverDelay); nakErr != nil {
				slog.Error("nak webhook event failed", "error", nakErr, "cause", handleErr)
			}
			return
		}
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
		// Дождаться завершения callback'ов consumer до close(jobs) — иначе
		// dispatchEvent может писать в закрытый канал (DATA RACE под -race).
		d.inFlight.Wait()
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
//
// Возвращает nil, если событие успешно обработано (можно Ack), и ошибку, если
// постановка в очередь не удалась (нужен NAK для повторной доставки, APIGW-02).
// Некорректный JSON (poison message) считается обработанным — повтор не поможет,
// поэтому возвращается nil, и сообщение подтверждается.
func (d *WebhookDispatcher) handleMessage(ctx context.Context, msg jetstream.Msg) error {
	subject := msg.Subject()
	body := msg.Data()

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		slog.Error("не удалось разобрать payload вебхук-события", "error", err, "subject", subject)
		return nil
	}

	// Subject NATS используется как тип события (например: tracker.{projID}.issue.created).
	return d.dispatchEvent(ctx, subject, body, payload)
}

// dispatchEvent находит все активные эндпоинты, подписанные на событие, и
// пушит задачи доставки в воркер-пул. HI-08: вместо unbounded `go` —
// очередь с фиксированным числом воркеров.
//
// APIGW-02: возвращает ошибку, если хотя бы одну задачу не удалось поставить
// в очередь (переполнение) либо при транзиентной ошибке получения эндпоинтов.
// Вызывающий код в этом случае не подтверждает сообщение, и JetStream доставит
// его повторно — это исключает «тихую» потерю событий при всплесках нагрузки.
func (d *WebhookDispatcher) dispatchEvent(ctx context.Context, eventType string, body []byte, payload map[string]interface{}) error {
	endpoints, err := d.repo.ListActiveWebhookEndpoints(ctx)
	if err != nil {
		slog.Error("ошибка получения вебхук-эндпоинтов", "error", err)
		return fmt.Errorf("получение активных вебхук-эндпоинтов: %w", err)
	}

	// TASK-012: project scope. Событие без extractable projectID не доставляем
	// ни одному project-scoped эндпоинту (fail-closed). Все Hub webhook
	// endpoints привязаны к проекту (edge Required); org-level hooks нет.
	eventProjectID, hasProject := extractEventProjectID(eventType, payload)

	for _, ep := range endpoints {
		if !containsEvent(ep.Events, eventType) {
			continue
		}
		if !hasProject {
			// Subject без project UUID (tms.*, sprint.>, messenger.*, project.created)
			// и без project_id в payload — не шлём в project hooks.
			continue
		}
		epProjectID, ok := webhookEndpointProjectID(ep)
		if !ok || epProjectID != eventProjectID {
			continue
		}
		job := webhookJob{ep: ep, eventType: eventType, body: body, payload: payload}
		select {
		case d.jobs <- job:
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Очередь переполнена. Не блокируем consumer, но и НЕ теряем
			// событие: возвращаем ошибку, чтобы сообщение было повторно
			// доставлено JetStream (MaxDeliver=3) вместо безусловного Ack.
			slog.Warn("webhook job queue full, requesting redelivery",
				"endpoint", ep.URL, "event", eventType, "queue_size", webhookQueueSize)
			return errWebhookQueueFull
		}
	}
	return nil
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

		// Уникальный X-Hub-Delivery на каждую попытку (A2W4-5-003) —
		// consumer отличает retry от «нового» события.
		deliveryID := uuid.New().String()
		statusCode, respBody, deliveryErr := d.sendWebhook(ctx, job.ep.URL, job.body, sig, job.eventType, deliveryID)

		errMsg := ""
		if deliveryErr != nil {
			errMsg = deliveryErr.Error()
		}

		if saveErr := d.repo.CreateWebhookDelivery(ctx, CreateWebhookDeliveryInput{
			EventType:    job.eventType,
			Payload:      job.payload,
			StatusCode:   statusCode,
			ResponseBody: respBody,
			EndpointID:   job.ep.ID,
			ErrorMsg:     errMsg,
		}); saveErr != nil {
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
// ctx прокидывается в NewRequestWithContext (A2W4-5-002): cancel/timeout
// прерывает Do, а не ждёт полный client Timeout × retries.
// deliveryID — значение заголовка X-Hub-Delivery (UUID на попытку).
func (d *WebhookDispatcher) sendWebhook(ctx context.Context, url string, body []byte, signature, eventType, deliveryID string) (int, string, error) {
	if err := safehttp.ValidateOutboundURL(url); err != nil {
		return 0, "", fmt.Errorf("отклонён небезопасный URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", "sha256="+signature)
	req.Header.Set("X-Hub-Event", eventType)
	req.Header.Set("X-Hub-Delivery", deliveryID)
	req.Header.Set("User-Agent", "Hub-Webhook/1.0")

	resp, err := webhookHTTPClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("ошибка отправки: %w", err)
	}
	defer resp.Body.Close()

	// BE-H2: read at most maxWebhookResponseBodyBytes от ответа эндпоинта.
	// Раньше использовали respBuf.ReadFrom(resp.Body) — это читало тело
	// целиком (вплоть до OOM) и только потом обрезало. io.LimitReader
	// гарантирует, что больше N байт мы не аллоцируем.
	var respBuf bytes.Buffer
	_, _ = io.Copy(&respBuf, io.LimitReader(resp.Body, maxWebhookResponseBodyBytes))
	respBody := respBuf.String()

	return resp.StatusCode, respBody, nil
}

// signPayload подписывает payload HMAC-SHA256 секретом эндпоинта.
func signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// webhookEventAliases — UI short names (AVAILABLE_EVENTS) → альтернативные
// NATS subject / prefix, когда suffix-match недостаточен
// (test_run.completed ↔ tms.testrun.completed / tms.run.completed.<id>).
var webhookEventAliases = map[string][]string{
	"test_run.completed": {
		"tms.testrun.completed",
		"tms.run.completed",
	},
	// Tracker публикует смену колонки/статуса как issue.moved.
	"issue.status_changed": {
		"issue.moved",
	},
}

// containsEvent проверяет подписку эндпоинта на NATS subject / UI short name.
// Match: exact, "*", suffix ("."+sub), first+last token (sprint.<id>.completed),
// либо alias-таблица webhookEventAliases.
func containsEvent(events []string, event string) bool {
	for _, sub := range events {
		if eventMatchesSubscription(sub, event) {
			return true
		}
	}
	return false
}

func eventMatchesSubscription(sub, event string) bool {
	if sub == "*" || sub == event {
		return true
	}
	// project.<uuid>.issue.created ← issue.created
	if strings.HasSuffix(event, "."+sub) {
		return true
	}
	// sprint.<uuid>.completed ← sprint.completed (первый и последний токен)
	subParts := strings.Split(sub, ".")
	eventParts := strings.Split(event, ".")
	if len(subParts) == 2 && len(eventParts) >= 3 &&
		eventParts[0] == subParts[0] && eventParts[len(eventParts)-1] == subParts[1] {
		return true
	}
	for _, alias := range webhookEventAliases[sub] {
		if event == alias ||
			strings.HasSuffix(event, "."+alias) ||
			strings.HasPrefix(event, alias+".") {
			return true
		}
	}
	return false
}

// extractEventProjectID извлекает project UUID из subject
// (project.<uuid>.… / tracker.<uuid>.…) либо из payload (project_id / projectId).
// Если извлечь нельзя — ok=false (fail-closed для project-scoped delivery).
func extractEventProjectID(subject string, payload map[string]interface{}) (uuid.UUID, bool) {
	parts := strings.Split(subject, ".")
	if len(parts) >= 2 && (parts[0] == "project" || parts[0] == "tracker") {
		if id, err := uuid.Parse(parts[1]); err == nil {
			return id, true
		}
	}
	if payload == nil {
		return uuid.Nil, false
	}
	for _, key := range []string{"project_id", "projectId"} {
		raw, ok := payload[key]
		if !ok || raw == nil {
			continue
		}
		switch v := raw.(type) {
		case string:
			if id, err := uuid.Parse(v); err == nil {
				return id, true
			}
		case uuid.UUID:
			if v != uuid.Nil {
				return v, true
			}
		}
	}
	return uuid.Nil, false
}

// webhookEndpointProjectID возвращает ID проекта эндпоинта из eager-loaded edge.
func webhookEndpointProjectID(ep *ent.WebhookEndpoint) (uuid.UUID, bool) {
	if ep == nil || ep.Edges.Project == nil {
		return uuid.Nil, false
	}
	return ep.Edges.Project.ID, true
}

// --- CRUD вебхуков ---

// allowedWebhookEvents — UI short names (AVAILABLE_EVENTS из
// frontend/src/modules/settings/webhooks.tsx) + wildcard "*".
// Bind validate-теги Fiber не enforce'ит — проверка в сервисе (P1-BE-28).
var allowedWebhookEvents = map[string]struct{}{
	"*":                    {},
	"issue.created":        {},
	"issue.updated":        {},
	"issue.status_changed": {},
	"test_run.completed":   {},
	"sprint.completed":     {},
}

// validateWebhookEvents требует непустой список и только allowlist-имена.
func validateWebhookEvents(events []string) error {
	if len(events) == 0 {
		return newInvalidInputError("events: требуется хотя бы одно событие")
	}
	for _, e := range events {
		if _, ok := allowedWebhookEvents[e]; !ok {
			return newInvalidInputError("events: неизвестное событие " + e)
		}
	}
	return nil
}

// WebhookService управляет CRUD операциями для вебхук-эндпоинтов.
type WebhookService struct {
	repo WebhookRepository
}

// NewWebhookService создаёт новый сервис вебхуков.
func NewWebhookService(repo WebhookRepository) *WebhookService {
	return &WebhookService{repo: repo}
}

// CreateEndpoint создаёт новый вебхук-эндпоинт в проекте.
// URL валидируется на схему (http/https) и отсутствие userinfo —
// фактическая блокировка приватных IP происходит в SafeDialContext
// (CR-03, защита от DNS-rebinding).
// Автогенерированный (или переданный) secret возвращается в ответе один раз.
func (s *WebhookService) CreateEndpoint(ctx context.Context, projectID uuid.UUID, req CreateWebhookRequest) (*CreateWebhookResponse, error) {
	if err := safehttp.ValidateOutboundURL(req.URL); err != nil {
		return nil, newInvalidInputError("некорректный URL вебхука: " + err.Error())
	}
	if err := validateWebhookEvents(req.Events); err != nil {
		return nil, err
	}

	secret := req.Secret
	if secret == "" {
		secret = uuid.New().String()
	}

	ep, err := s.repo.CreateWebhookEndpoint(ctx, projectID, req.URL, secret, req.Events)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать вебхук-эндпоинт: %w", err)
	}

	return &CreateWebhookResponse{
		WebhookResponse: *toWebhookResponse(ep),
		Secret:          secret,
	}, nil
}

// ListEndpoints возвращает вебхук-эндпоинты проекта.
func (s *WebhookService) ListEndpoints(ctx context.Context, projectID uuid.UUID) ([]WebhookResponse, error) {
	eps, err := s.repo.ListWebhookEndpointsByProject(ctx, projectID)
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
	input := UpdateWebhookEndpointInput{
		URL:      req.URL,
		Secret:   req.Secret,
		Events:   req.Events,
		IsActive: req.IsActive,
	}
	if req.URL != nil {
		if err := safehttp.ValidateOutboundURL(*req.URL); err != nil {
			return nil, newInvalidInputError("некорректный URL вебхука: " + err.Error())
		}
	}
	if req.Events != nil {
		if err := validateWebhookEvents(req.Events); err != nil {
			return nil, err
		}
	}

	ep, err := s.repo.UpdateWebhookEndpoint(ctx, endpointID, input)
	if err != nil {
		if IsNotFound(err) {
			return nil, ErrWebhookEndpointNotFound
		}
		return nil, fmt.Errorf("не удалось обновить вебхук-эндпоинт: %w", err)
	}
	return toWebhookResponse(ep), nil
}

// DeleteEndpoint удаляет вебхук-эндпоинт.
func (s *WebhookService) DeleteEndpoint(ctx context.Context, endpointID uuid.UUID) error {
	err := s.repo.DeleteWebhookEndpoint(ctx, endpointID)
	if err != nil {
		if IsNotFound(err) {
			return ErrWebhookEndpointNotFound
		}
		return fmt.Errorf("не удалось удалить вебхук-эндпоинт: %w", err)
	}
	return nil
}

// EndpointProjectID возвращает projectID проекта, которому принадлежит эндпоинт
// (APIGW-01). Используется обработчиком для RBAC-проверки на маршрутах
// /webhooks/:id, где projectId отсутствует в пути.
func (s *WebhookService) EndpointProjectID(ctx context.Context, endpointID uuid.UUID) (uuid.UUID, error) {
	return s.repo.WebhookEndpointProjectID(ctx, endpointID)
}

// ListDeliveries возвращает историю доставок для эндпоинта.
func (s *WebhookService) ListDeliveries(ctx context.Context, endpointID uuid.UUID) ([]WebhookDeliveryResponse, error) {
	deliveries, err := s.repo.ListWebhookDeliveries(ctx, endpointID, 50)
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
