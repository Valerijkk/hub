package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/aiprovider"
)

func TestAdminHandler_CreateProvider_InvalidJSON(t *testing.T) {
	app, _ := testAdminAIApp(t, true)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/ai/providers", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAdminHandler_CreateProvider_MissingFields(t *testing.T) {
	app, _ := testAdminAIApp(t, true)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/ai/providers",
		bytes.NewReader([]byte(`{"name":"","base_url":""}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAdminHandler_UpdateSettings_InvalidJSON(t *testing.T) {
	app, _ := testAdminAIApp(t, true)

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/ai/settings", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAdminHandler_UpdateProvider_InvalidID(t *testing.T) {
	app, _ := testAdminAIApp(t, true)

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/ai/providers/not-a-uuid",
		bytes.NewReader([]byte(`{"enabled":false}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAdminHandler_TestProvider_NotFound(t *testing.T) {
	app, _ := testAdminAIApp(t, true)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/ai/providers/"+uuid.New().String()+"/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestAdminHandler_UpdateSettings_PromptTooLong(t *testing.T) {
	app, _ := testAdminAIApp(t, true)

	longPrompt := strings.Repeat("x", 8001)
	body, _ := json.Marshal(ai.UpdateAdminSettingsRequest{SystemPrompt: &longPrompt})
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/ai/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAdminHandler_UpdateProvider_Success(t *testing.T) {
	app, _ := testAdminAIApp(t, true)

	createBody, _ := json.Marshal(ai.CreateProviderRequest{
		Name:         "Local LLM",
		ProviderType: string(aiprovider.ProviderTypeOllama),
		BaseURL:      "http://127.0.0.1:11434/v1",
		Models:       []string{"llama3"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/ai/providers", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var created ai.ProviderResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	enabled := false
	patchBody, _ := json.Marshal(ai.UpdateProviderRequest{Enabled: &enabled, Version: created.Version})
	req = httptest.NewRequest(http.MethodPatch, "/api/admin/ai/providers/"+created.ID, bytes.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAdminHandler_GetUsageHistory_CustomDays(t *testing.T) {
	app, _ := testAdminAIApp(t, true)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/ai/usage/history?days=30", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
