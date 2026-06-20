package apigateway

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

func newAPIKeyTestService(t *testing.T) (*APIKeyService, uuid.UUID, string) {
	t.Helper()

	dsn := "file:apikeytest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	u, err := client.User.Create().
		SetEmail("apikey-" + uuid.New().String()[:8] + "@example.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	svc := NewAPIKeyService(client)
	t.Cleanup(svc.Close)

	resp, err := svc.CreateAPIKey(ctx, u.ID, "integration-test", nil)
	require.NoError(t, err)

	return svc, u.ID, resp.Key
}

// TestAPIKeyAuth_ValidKey_SetsUserID — smoke: middleware принимает X-API-Key
// и прокидывает userID в Locals для downstream handlers.
func TestAPIKeyAuth_ValidKey_SetsUserID(t *testing.T) {
	svc, userID, rawKey := newAPIKeyTestService(t)

	app := fiber.New()
	app.Use(APIKeyAuth(svc))
	app.Get("/api/protected", func(c fiber.Ctx) error {
		return c.SendString(c.Locals("userID").(string))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("X-API-Key", rawKey)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, userID.String(), string(body))
}

// TestAPIKeyAuth_InvalidKey_Rejected — недействительный ключ не должен
// проходить middleware (401).
func TestAPIKeyAuth_InvalidKey_Rejected(t *testing.T) {
	svc, _, _ := newAPIKeyTestService(t)

	app := fiber.New()
	app.Use(APIKeyAuth(svc))
	app.Get("/api/protected", func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("X-API-Key", "invalid-key-value")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestAPIKeyAuth_NoHeader_PassesThrough — без X-API-Key middleware не блокирует
// запрос (JWT/other auth может обработать дальше).
func TestAPIKeyAuth_NoHeader_PassesThrough(t *testing.T) {
	svc, _, _ := newAPIKeyTestService(t)

	app := fiber.New()
	app.Use(APIKeyAuth(svc))
	app.Get("/api/open", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/open", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "ok", string(body))
}

// TestAPIKeyAuth_MiddlewareOrder_SEC01 — SEC-01: в Fiber Use влияет только на
// routes, зарегистрированные ПОСЛЕ middleware. APIKeyAuth после route = dead code.
func TestAPIKeyAuth_MiddlewareOrder_SEC01(t *testing.T) {
	svc, userID, rawKey := newAPIKeyTestService(t)

	handler := func(c fiber.Ctx) error {
		uid, _ := c.Locals("userID").(string)
		return c.SendString(uid)
	}

	// Wrong order: route registered before APIKeyAuth (pre-SEC-01 bug).
	wrongApp := fiber.New()
	wrongApp.Get("/api/wrong", handler)
	wrongApp.Use(APIKeyAuth(svc))

	reqWrong := httptest.NewRequest(http.MethodGet, "/api/wrong", nil)
	reqWrong.Header.Set("X-API-Key", rawKey)
	respWrong, err := wrongApp.Test(reqWrong)
	require.NoError(t, err)
	defer respWrong.Body.Close()

	wrongBody, _ := io.ReadAll(respWrong.Body)
	assert.Equal(t, "", string(wrongBody),
		"APIKeyAuth after route must not set userID (SEC-01 dead code)")

	// Correct order: APIKeyAuth before route (production main.go pattern).
	correctApp := fiber.New()
	correctApp.Use(APIKeyAuth(svc))
	correctApp.Get("/api/correct", handler)

	reqCorrect := httptest.NewRequest(http.MethodGet, "/api/correct", nil)
	reqCorrect.Header.Set("X-API-Key", rawKey)
	respCorrect, err := correctApp.Test(reqCorrect)
	require.NoError(t, err)
	defer respCorrect.Body.Close()

	assert.Equal(t, http.StatusOK, respCorrect.StatusCode)
	correctBody, _ := io.ReadAll(respCorrect.Body)
	assert.Equal(t, userID.String(), string(correctBody),
		"APIKeyAuth before route must authenticate via X-API-Key")
}
