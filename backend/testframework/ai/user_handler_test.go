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

func testUserAIApp(t *testing.T) (*fiber.App, uuid.UUID) {
	t.Helper()
	userID := uuid.New()
	repo := &mockRepo{providers: []*ent.AIProvider{}}
	svc := ai.NewService(repo, &mockAdapter{}, nil)
	h := ai.NewUserHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, userID
}

func TestUserHandler_ProviderCRUD(t *testing.T) {
	app, _ := testUserAIApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/ai/my/providers", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := json.Marshal(ai.CreateProviderRequest{
		Name:         "BYOK",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "https://api.openai.com/v1",
		Models:       []string{"gpt-4o-mini"},
	})
	req = httptest.NewRequest(http.MethodPost, "/api/ai/my/providers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created ai.ProviderResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	enabled := false
	patch, _ := json.Marshal(ai.UpdateProviderRequest{Enabled: &enabled})
	req = httptest.NewRequest(http.MethodPatch, "/api/ai/my/providers/"+created.ID, bytes.NewReader(patch))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUserHandler_UpdateProvider_InvalidID(t *testing.T) {
	app, _ := testUserAIApp(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/ai/my/providers/not-uuid", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
