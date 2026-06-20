package apigateway

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
)

func newWebhookHandlerEnv(t *testing.T) (*Handler, context.Context, uuid.UUID) {
	t.Helper()
	dsn := "file:apigwwebhook_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	proj, err := client.Project.Create().
		SetName("Webhook Project").
		SetKey("WHK").
		Save(ctx)
	require.NoError(t, err)

	webhookSvc := NewWebhookService(client)
	apiKeySvc := NewAPIKeyService(client)
	t.Cleanup(apiKeySvc.Close)

	return NewHandler(apiKeySvc, webhookSvc), ctx, proj.ID
}

func TestHandler_CreateEndpoint_Validation(t *testing.T) {
	h, _, projectID := newWebhookHandlerEnv(t)

	app := fiber.New()
	app.Post("/projects/:projectId/webhooks", h.CreateEndpoint)

	t.Run("invalid project id", func(t *testing.T) {
		body, _ := json.Marshal(CreateWebhookRequest{
			URL:    "https://example.com/hook",
			Events: []string{"issue.created"},
		})
		req := httptest.NewRequest(http.MethodPost, "/projects/not-uuid/webhooks", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("unsafe url rejected", func(t *testing.T) {
		body, _ := json.Marshal(CreateWebhookRequest{
			URL:    "file:///etc/passwd",
			Events: []string{"issue.created"},
		})
		req := httptest.NewRequest(http.MethodPost, "/projects/"+projectID.String()+"/webhooks", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestHandler_CreateEndpoint_Success(t *testing.T) {
	h, _, projectID := newWebhookHandlerEnv(t)

	app := fiber.New()
	app.Post("/projects/:projectId/webhooks", h.CreateEndpoint)

	body, _ := json.Marshal(CreateWebhookRequest{
		URL:    "https://example.com/hub-webhook",
		Events: []string{"issue.created", "message.sent"},
	})
	req := httptest.NewRequest(http.MethodPost, "/projects/"+projectID.String()+"/webhooks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestHandler_ListEndpoints_Empty(t *testing.T) {
	h, _, projectID := newWebhookHandlerEnv(t)

	app := fiber.New()
	app.Get("/projects/:projectId/webhooks", h.ListEndpoints)

	req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID.String()+"/webhooks", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var eps []WebhookResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&eps))
	assert.Empty(t, eps)
}

func TestHandler_DeleteEndpoint_NotFound(t *testing.T) {
	h, _, _ := newWebhookHandlerEnv(t)

	app := fiber.New()
	app.Delete("/webhooks/:id", h.DeleteEndpoint)

	req := httptest.NewRequest(http.MethodDelete, "/webhooks/"+uuid.New().String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestWebhookService_ListEndpoints_ReturnsCreated(t *testing.T) {
	h, ctx, projectID := newWebhookHandlerEnv(t)
	webhookSvc := h.webhookSvc

	created, err := webhookSvc.CreateEndpoint(ctx, projectID, CreateWebhookRequest{
		URL:    "https://hooks.example.com/events",
		Events: []string{"issue.updated"},
	})
	require.NoError(t, err)

	list, err := webhookSvc.ListEndpoints(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, created.ID, list[0].ID)
}

func TestWebhookService_UpdateEndpoint_InvalidURL(t *testing.T) {
	h, ctx, projectID := newWebhookHandlerEnv(t)
	webhookSvc := h.webhookSvc

	created, err := webhookSvc.CreateEndpoint(ctx, projectID, CreateWebhookRequest{
		URL:    "https://hooks.example.com/events",
		Events: []string{"issue.updated"},
	})
	require.NoError(t, err)

	badURL := "file:///etc/passwd"
	_, err = webhookSvc.UpdateEndpoint(ctx, uuid.MustParse(created.ID), UpdateWebhookRequest{
		URL: &badURL,
	})
	require.Error(t, err)
}

func TestWebhookService_ListDeliveries_Empty(t *testing.T) {
	h, ctx, projectID := newWebhookHandlerEnv(t)
	webhookSvc := h.webhookSvc

	created, err := webhookSvc.CreateEndpoint(ctx, projectID, CreateWebhookRequest{
		URL:    "https://hooks.example.com/events",
		Events: []string{"issue.updated"},
	})
	require.NoError(t, err)

	deliveries, err := webhookSvc.ListDeliveries(ctx, uuid.MustParse(created.ID))
	require.NoError(t, err)
	assert.Empty(t, deliveries)
}
