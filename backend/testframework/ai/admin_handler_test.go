package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
)

func testAdminAIApp(t *testing.T, isAdmin bool) (*fiber.App, uuid.UUID) {
	t.Helper()
	userID := uuid.New()
	repo := &mockRepo{providers: []*ent.AIProvider{}}
	svc := ai.NewService(repo, &mockAdapter{resp: &ai.ChatCompletionResponse{
		Choices: []struct {
			Index        int         `json:"index"`
			Message      ai.ChatMessage `json:"message"`
			FinishReason string      `json:"finish_reason"`
		}{{Message: ai.ChatMessage{Content: "pong"}}},
	}}, nil)
	svc.SetSettingsStore(&memSettings{data: map[string]string{}})

	isAdminFn := func(_ fiber.Ctx, id uuid.UUID) (bool, error) {
		return isAdmin && id == userID, nil
	}
	h := ai.NewAdminHandler(svc, isAdminFn)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, userID
}

func TestAdminHandler_ProvidersAndSettings(t *testing.T) {
	app, _ := testAdminAIApp(t, true)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/ai/providers", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := json.Marshal(ai.CreateProviderRequest{
		Name:         "DeepSeek",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "https://api.deepseek.com/v1",
		Models:       []string{"deepseek-chat"},
	})
	req = httptest.NewRequest(http.MethodPost, "/api/admin/ai/providers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created ai.ProviderResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	req = httptest.NewRequest(http.MethodGet, "/api/admin/ai/settings", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	prompt := "Custom system"
	settingsBody, _ := json.Marshal(ai.UpdateAdminSettingsRequest{SystemPrompt: &prompt})
	req = httptest.NewRequest(http.MethodPatch, "/api/admin/ai/settings", bytes.NewReader(settingsBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/admin/ai/providers/"+created.ID+"/test", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/admin/ai/usage/summary", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/admin/ai/usage/history?days=14", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAdminHandler_Forbidden(t *testing.T) {
	app, _ := testAdminAIApp(t, false)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/ai/providers", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
