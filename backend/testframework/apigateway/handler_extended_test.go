package apigateway_test
import (
	"github.com/valerii/hub/backend/internal/apigateway"
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

func setupAPIGatewayApp(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID) {
	t.Helper()
	env := newWebhookEnv(t)
	app := env.buildApp(t, env.ownerID)
	return app, env.ownerID, env.projectID
}

func TestHandler_RegisterRoutes_RevokeAPIKeySuccess(t *testing.T) {
	dsn := "file:apigwroutes_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	u, err := client.User.Create().
		SetEmail("routes-" + uuid.New().String()[:8] + "@example.com").
		Save(ctx)
	require.NoError(t, err)

	repo := apigateway.NewEntRepository(client, apigatewayTestCipher(t))
	apiKeySvc := apigateway.NewAPIKeyService(repo)
	t.Cleanup(apiKeySvc.Close)
	h := apigateway.NewHandler(apiKeySvc, apigateway.NewWebhookService(repo))

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", u.ID.String())
		return c.Next()
	}
	// API-ключевые маршруты не требуют RBAC; webhook CRUD здесь не тестируется.
	h.RegisterRoutes(api, authMw, nil)

	createBody, _ := json.Marshal(apigateway.CreateAPIKeyRequest{Name: "revoke-me"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/api-keys/", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created apigateway.APIKeyResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	req = httptest.NewRequest(http.MethodDelete, "/api/user/api-keys/"+created.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/user/api-keys/", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var keys []apigateway.APIKeyResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&keys))
	assert.Empty(t, keys)
}

func TestHandler_UpdateEndpointAndListDeliveries(t *testing.T) {
	app, _, projectID := setupAPIGatewayApp(t)

	createBody, _ := json.Marshal(apigateway.CreateWebhookRequest{
		URL:    "https://example.com/hook",
		Events: []string{"issue.created"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/webhooks/", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var ep apigateway.WebhookResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&ep))

	newURL := "https://example.com/hook-v2"
	updateBody, _ := json.Marshal(apigateway.UpdateWebhookRequest{URL: &newURL})
	req = httptest.NewRequest(http.MethodPut, "/api/webhooks/"+ep.ID, bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/webhooks/"+ep.ID+"/deliveries", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, "/api/webhooks/"+ep.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestAPIKeyService_RevokeAPIKey_NotFound(t *testing.T) {
	repo, userID, ctx := newAPIKeyRepo(t)
	svc := apigateway.NewAPIKeyService(repo)
	t.Cleanup(svc.Close)

	err := svc.RevokeAPIKey(ctx, userID, uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, apigateway.ErrAPIKeyNotFound)
}

func TestAPIKeyService_RevokeAPIKey_Success(t *testing.T) {
	repo, userID, ctx := newAPIKeyRepo(t)
	svc := apigateway.NewAPIKeyService(repo)
	t.Cleanup(svc.Close)

	created, err := svc.CreateAPIKey(ctx, userID, "temp", nil)
	require.NoError(t, err)

	keyID, err := uuid.Parse(created.ID)
	require.NoError(t, err)
	require.NoError(t, svc.RevokeAPIKey(ctx, userID, keyID))

	list, err := svc.ListAPIKeys(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, list)
}
