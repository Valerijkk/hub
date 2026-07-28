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

func testAIService(t *testing.T) (*ai.Service, uuid.UUID) {
	t.Helper()
	userID := uuid.New()
	repo := &mockRepo{
		providers: []*ent.AIProvider{
			{
				ID: uuid.New(), Name: "System", BaseURL: "https://api.deepseek.com/v1",
				Models: []string{"deepseek-chat"}, Enabled: true, Scope: aiprovider.ScopeSystem,
			},
		},
		allowExternal: map[uuid.UUID]bool{userID: true},
	}
	svc := ai.NewService(repo, &mockAdapter{resp: &ai.ChatCompletionResponse{
		Choices: []struct {
			Index        int         `json:"index"`
			Message      ai.ChatMessage `json:"message"`
			FinishReason string      `json:"finish_reason"`
		}{{Message: ai.ChatMessage{Content: "ok"}}},
	}}, nil)
	return svc, userID
}

func testAIHandlerApp(t *testing.T) (*fiber.App, uuid.UUID) {
	t.Helper()
	svc, userID := testAIService(t)
	h := ai.NewHandler(svc)
	h.SetBotUserID(uuid.New())

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, userID
}

func TestHandler_RegisterRoutes_BotAndModels(t *testing.T) {
	app, _ := testAIHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/ai/bot", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/ai/models", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var models ai.ModelsListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&models))
	require.Len(t, models.Models, 1)
}

func TestHandler_GetUsageToday(t *testing.T) {
	app, _ := testAIHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/ai/usage/today", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_GetPreferences(t *testing.T) {
	app, _ := testAIHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/ai/preferences", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_UpdatePreferences_InvalidBody(t *testing.T) {
	app, _ := testAIHandlerApp(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/ai/preferences", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_GetBotInfo_NotConfigured(t *testing.T) {
	svc, userID := testAIService(t)
	h := ai.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	req := httptest.NewRequest(http.MethodGet, "/api/ai/bot", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
}

func TestUserIDFromCtx(t *testing.T) {
	id := uuid.New()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", id.String())
		return c.Next()
	})
	app.Get("/ok", func(c fiber.Ctx) error {
		got, err := ai.ExportUserIDFromCtx(c)
		require.NoError(t, err)
		assert.Equal(t, id, got)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	app2 := fiber.New()
	app2.Get("/missing", func(c fiber.Ctx) error {
		_, err := ai.ExportUserIDFromCtx(c)
		return err
	})
	req = httptest.NewRequest(http.MethodGet, "/missing", nil)
	resp, err = app2.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestSecondsUntilMidnightUTC(t *testing.T) {
	sec := ai.ExportSecondsUntilMidnightUTC()
	assert.NotEmpty(t, sec)
	assert.NotEqual(t, "0", sec)
}

func TestHandler_AssistJQL_InvalidBody(t *testing.T) {
	app, _ := testAIHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/ai/jql/assist", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_SuggestComposer_InvalidBody(t *testing.T) {
	app, _ := testAIHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/ai/composer/suggest", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
