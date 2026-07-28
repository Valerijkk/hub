//go:build hubtest

package apigateway

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/valerii/hub/backend/ent"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// ExportHashKey возвращает SHA-256 hex хеш API-ключа для hubtest.
func ExportHashKey(key string) string {
	return hashKey(key)
}

// ExportSignPayload подписывает webhook payload для hubtest.
func ExportSignPayload(payload []byte, secret string) string {
	return signPayload(payload, secret)
}

// ExportContainsEvent проверяет наличие события в списке для hubtest.
func ExportContainsEvent(events []string, event string) bool {
	return containsEvent(events, event)
}

// ExportWebhookSendWebhook вызывает sendWebhook для hubtest.
func ExportWebhookSendWebhook(d *WebhookDispatcher, url string, body []byte, signature, eventType string) (int, string, error) {
	return d.sendWebhook(url, body, signature, eventType)
}

// ExportSetWebhookHTTPClient подменяет webhookHTTPClient для hubtest.
func ExportSetWebhookHTTPClient(c *http.Client) *http.Client {
	prev := webhookHTTPClient
	webhookHTTPClient = c
	return prev
}

// ExportResetWebhookHTTPClient восстанавливает webhookHTTPClient для hubtest.
func ExportResetWebhookHTTPClient(prev *http.Client) {
	webhookHTTPClient = prev
}

// ExportMaxWebhookResponseBodyBytes возвращает лимит тела ответа webhook для hubtest.
func ExportMaxWebhookResponseBodyBytes() int {
	return maxWebhookResponseBodyBytes
}

// ExportParseID извлекает UUID из path param для hubtest.
func ExportParseID(c fiber.Ctx, name string) (uuid.UUID, error) {
	return parseID(c, name)
}

// ExportGetUserID извлекает userID из Fiber Locals для hubtest.
func ExportGetUserID(c fiber.Ctx) (uuid.UUID, error) {
	return getUserID(c)
}

// ExportHandlerWebhookService возвращает webhook service handler для hubtest.
func ExportHandlerWebhookService(h *Handler) *WebhookService {
	return h.webhookSvc
}

// ExportWebhookDispatcherHandleMessage вызывает handleMessage для hubtest.
func ExportWebhookDispatcherHandleMessage(d *WebhookDispatcher, ctx context.Context, msg jetstream.Msg) error {
	return d.handleMessage(ctx, msg)
}

// ExportWebhookDispatcherDeliverWithRetry вызывает deliverWithRetry для hubtest.
func ExportWebhookDispatcherDeliverWithRetry(d *WebhookDispatcher, ctx context.Context, ep *ent.WebhookEndpoint, eventType string, body []byte, payload map[string]interface{}) {
	d.deliverWithRetry(ctx, webhookJob{ep: ep, eventType: eventType, body: body, payload: payload})
}

// ExportWebhookDispatcherInitJobs инициализирует jobs channel для hubtest.
func ExportWebhookDispatcherInitJobs(d *WebhookDispatcher, buf int) {
	d.jobs = make(chan webhookJob, buf)
}

// ExportWebhookDispatcherJobAvailable проверяет наличие job в очереди для hubtest.
func ExportWebhookDispatcherJobAvailable(d *WebhookDispatcher) bool {
	select {
	case <-d.jobs:
		return true
	default:
		return false
	}
}

// ExportWebhookDispatcherDrainJobEndpointIDs снимает все jobs из очереди и
// возвращает ID эндпоинтов (для denial/scope тестов).
func ExportWebhookDispatcherDrainJobEndpointIDs(d *WebhookDispatcher) []uuid.UUID {
	var ids []uuid.UUID
	for {
		select {
		case job := <-d.jobs:
			ids = append(ids, job.ep.ID)
		default:
			return ids
		}
	}
}

// ExportExtractEventProjectID извлекает project UUID из subject/payload для hubtest.
func ExportExtractEventProjectID(subject string, payload map[string]interface{}) (uuid.UUID, bool) {
	return extractEventProjectID(subject, payload)
}

// ExportWebhookDispatcherDispatchEvent вызывает dispatchEvent для hubtest.
func ExportWebhookDispatcherDispatchEvent(d *WebhookDispatcher, ctx context.Context, eventType string, body []byte, payload map[string]interface{}) error {
	return d.dispatchEvent(ctx, eventType, body, payload)
}

// ExportErrWebhookQueueFull возвращает sentinel-ошибку переполнения очереди для hubtest.
func ExportErrWebhookQueueFull() error {
	return errWebhookQueueFull
}

// ExportWebhookFilterSubjects возвращает JetStream filter subjects для hubtest.
func ExportWebhookFilterSubjects() []string {
	return natsclient.WebhookFilterSubjects
}
