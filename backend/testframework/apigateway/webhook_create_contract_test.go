package apigateway_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/apigateway"
)

// TestWebhookCreate_SecretOnce_P1BE27 — Create без secret → one-time secret в JSON;
// List/повторный GET не отдают secret (P1-BE-27 / W4-A5-003).
func TestWebhookCreate_SecretOnce_P1BE27(t *testing.T) {
	env := newWebhookEnv(t)
	app := env.buildApp(t, env.ownerID)

	body, _ := json.Marshal(apigateway.CreateWebhookRequest{
		URL:    "https://hooks.example.com/secret-once",
		Events: []string{"issue.created"},
		// Secret пустой — сервис генерирует uuid.
	})
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+env.projectID.String()+"/webhooks/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created apigateway.CreateWebhookResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEmpty(t, created.ID)
	require.NotEmpty(t, created.Secret, "Create без secret обязан вернуть auto-secret один раз")

	// List не содержит secret.
	listReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+env.projectID.String()+"/webhooks/", nil)
	listResp, err := app.Test(listReq)
	require.NoError(t, err)
	defer listResp.Body.Close()
	require.Equal(t, http.StatusOK, listResp.StatusCode)

	var raw []map[string]interface{}
	require.NoError(t, json.NewDecoder(listResp.Body).Decode(&raw))
	require.Len(t, raw, 1)
	_, hasSecret := raw[0]["secret"]
	assert.False(t, hasSecret, "List не должен отдавать secret")
}

// TestWebhookCreate_EventsAllowlist_P1BE28 — events:[] и неизвестные имена → 400;
// allowlist (AVAILABLE_EVENTS + *) → 201 (P1-BE-28 / W4-A5-004).
func TestWebhookCreate_EventsAllowlist_P1BE28(t *testing.T) {
	env := newWebhookEnv(t)
	app := env.buildApp(t, env.ownerID)
	base := "/api/projects/" + env.projectID.String() + "/webhooks/"

	t.Run("empty events rejected", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"url":    "https://hooks.example.com/empty",
			"events": []string{},
		})
		req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("unknown event rejected", func(t *testing.T) {
		body, _ := json.Marshal(apigateway.CreateWebhookRequest{
			URL:    "https://hooks.example.com/unknown",
			Events: []string{"message.sent"},
		})
		req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("allowlist accepted", func(t *testing.T) {
		body, _ := json.Marshal(apigateway.CreateWebhookRequest{
			URL:    "https://hooks.example.com/ok",
			Events: []string{"issue.status_changed", "test_run.completed"},
		})
		req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("wildcard accepted", func(t *testing.T) {
		body, _ := json.Marshal(apigateway.CreateWebhookRequest{
			URL:    "https://hooks.example.com/star",
			Events: []string{"*"},
		})
		req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

// TestWebhookService_CreateEndpoint_SecretAndEvents — guard на слое сервиса
// (не только HTTP), чтобы мутация CreateEndpoint без validate/secret red'ила.
func TestWebhookService_CreateEndpoint_SecretAndEvents(t *testing.T) {
	env := newWebhookEnv(t)

	created, err := env.webhookSvc.CreateEndpoint(env.ctx, env.projectID, apigateway.CreateWebhookRequest{
		URL:    "https://hooks.example.com/svc",
		Events: []string{"issue.created"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.Secret)

	_, err = env.webhookSvc.CreateEndpoint(env.ctx, env.projectID, apigateway.CreateWebhookRequest{
		URL:    "https://hooks.example.com/svc-empty",
		Events: []string{},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, apigateway.ErrInvalidInput)

	_, err = env.webhookSvc.CreateEndpoint(env.ctx, env.projectID, apigateway.CreateWebhookRequest{
		URL:    "https://hooks.example.com/svc-bad",
		Events: []string{"not.a.real.event"},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, apigateway.ErrInvalidInput)

	_, err = env.webhookSvc.CreateEndpoint(env.ctx, env.projectID, apigateway.CreateWebhookRequest{
		URL:    "https://hooks.example.com/svc-sprint",
		Events: []string{"sprint.completed"},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, apigateway.ErrInvalidInput)
}
