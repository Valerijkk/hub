package apigateway_test
import (
	"github.com/valerii/hub/backend/internal/apigateway"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

func newHandlerTestEnv(t *testing.T) (*apigateway.Handler, *apigateway.APIKeyService, uuid.UUID) {
	t.Helper()
	dsn := "file:apigwhandler_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	u, err := client.User.Create().
		SetEmail("apigw-" + uuid.New().String()[:8] + "@example.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	repo := apigateway.NewEntRepository(client, apigatewayTestCipher(t))
	apiKeySvc := apigateway.NewAPIKeyService(repo)
	t.Cleanup(apiKeySvc.Close)

	return apigateway.NewHandler(apiKeySvc, nil), apiKeySvc, u.ID
}

func TestHandler_CreateAPIKey_Validation(t *testing.T) {
	h, _, userID := newHandlerTestEnv(t)

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	})
	app.Post("/user/api-keys", h.CreateAPIKey)

	t.Run("missing name", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/user/api-keys", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("success", func(t *testing.T) {
		body, _ := json.Marshal(apigateway.CreateAPIKeyRequest{Name: "ci-key"})
		req := httptest.NewRequest(http.MethodPost, "/user/api-keys", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

func TestHandler_RevokeAPIKey_InvalidID(t *testing.T) {
	h, _, userID := newHandlerTestEnv(t)

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	})
	app.Delete("/user/api-keys/:id", h.RevokeAPIKey)

	req := httptest.NewRequest(http.MethodDelete, "/user/api-keys/not-a-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_ListAPIKeys(t *testing.T) {
	h, svc, userID := newHandlerTestEnv(t)
	ctx := context.Background()

	_, err := svc.CreateAPIKey(ctx, userID, "key-one", nil)
	require.NoError(t, err)

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	})
	app.Get("/user/api-keys", h.ListAPIKeys)

	req := httptest.NewRequest(http.MethodGet, "/user/api-keys", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "key-one")
}

func TestParseID_AndGetUserID(t *testing.T) {
	t.Run("parseID invalid uuid", func(t *testing.T) {
		app := fiber.New()
		app.Get("/p/:projectId", func(c fiber.Ctx) error {
			_, err := apigateway.ExportParseID(c, "projectId")
			return err
		})
		req := httptest.NewRequest(http.MethodGet, "/p/not-a-uuid", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("getUserID missing", func(t *testing.T) {
		app2 := fiber.New()
		app2.Get("/u", func(c fiber.Ctx) error {
			_, err := apigateway.ExportGetUserID(c)
			return err
		})
		req := httptest.NewRequest(http.MethodGet, "/u", nil)
		resp, err := app2.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAPIKeyService_VerifyInvalidKey(t *testing.T) {
	dsn := "file:apigwsvc_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	svc := apigateway.NewAPIKeyService(apigateway.NewEntRepository(client, apigatewayTestCipher(t)))
	t.Cleanup(svc.Close)

	_, err := svc.VerifyAPIKey(context.Background(), "definitely-not-a-valid-key")
	require.Error(t, err)
	assert.ErrorIs(t, err, apigateway.ErrAPIKeyInvalid)
}

func TestAPIKeyService_ExpiredKeyRejected(t *testing.T) {
	dsn := "file:apigwexp_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	u, err := client.User.Create().
		SetEmail("exp-" + uuid.New().String()[:8] + "@example.com").
		Save(ctx)
	require.NoError(t, err)

	svc := apigateway.NewAPIKeyService(apigateway.NewEntRepository(client, apigatewayTestCipher(t)))
	t.Cleanup(svc.Close)

	past := time.Now().Add(-time.Hour)
	created, err := svc.CreateAPIKey(ctx, u.ID, "expired", &past)
	require.NoError(t, err)

	_, err = svc.VerifyAPIKey(ctx, created.Key)
	require.Error(t, err)
	assert.ErrorIs(t, err, apigateway.ErrAPIKeyExpired)
}
