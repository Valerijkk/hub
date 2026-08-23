package rbac_test

// Тесты middleware RBAC: extract-only CaptureBearerToken (не JWT auth).
import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/rbac"
)

// TestCaptureBearerToken_MissingHeader — нет Authorization → 401.
func TestCaptureBearerToken_MissingHeader(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	}, rbac.CaptureBearerToken())

	status, body := doGet(t, app, "/")
	assert.Equal(t, fiber.StatusUnauthorized, status)
	assert.Contains(t, body, "missing authorization header")
}

// TestCaptureBearerToken_InvalidFormat — не Bearer-схема → 401.
func TestCaptureBearerToken_InvalidFormat(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	}, rbac.CaptureBearerToken())

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token abc")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestCaptureBearerToken_EmptyBearerToken — "Bearer " без токена → 401.
func TestCaptureBearerToken_EmptyBearerToken(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	}, rbac.CaptureBearerToken())

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer ")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestCaptureBearerToken_ExtractOnlyNoJWTValidation — A2W6-A3-004/005:
// CaptureBearerToken только кладёт raw Bearer в Locals("token"); JWT не проверяет.
// Имя теста явно: extract-only, не ValidBearer/auth success.
func TestCaptureBearerToken_ExtractOnlyNoJWTValidation(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		token, _ := c.Locals("token").(string)
		return c.SendString(token)
	}, rbac.CaptureBearerToken())

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer jwt-token-123")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	require.NoError(t, err)
	assert.Equal(t, "jwt-token-123", string(body))
}

// TestRequireAuth_DeprecatedAliasOfCaptureBearerToken — legacy RequireAuth == CaptureBearerToken.
func TestRequireAuth_DeprecatedAliasOfCaptureBearerToken(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		token, _ := c.Locals("token").(string)
		return c.SendString(token)
	}, rbac.RequireAuth())

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer legacy-alias")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	require.NoError(t, err)
	assert.Equal(t, "legacy-alias", string(body))
}

// TestUserIDFromCtx проверяет сценарий: user id from ctx.
func TestUserIDFromCtx(t *testing.T) {
	app := fiber.New()
	userID := uuid.New().String()
	app.Get("/", func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.SendString(rbac.UserIDFromCtx(c))
	})

	status, body := doGet(t, app, "/")
	assert.Equal(t, fiber.StatusOK, status)
	assert.Equal(t, userID, body)

	app2 := fiber.New()
	app2.Get("/", func(c fiber.Ctx) error {
		return c.SendString(rbac.UserIDFromCtx(c))
	})
	status, body = doGet(t, app2, "/")
	assert.Equal(t, fiber.StatusOK, status)
	assert.Empty(t, body)
}

func doGet(t *testing.T, app *fiber.App, path string) (int, string) {
	t.Helper()

	req := httptest.NewRequest(fiber.MethodGet, path, nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	require.NoError(t, err)
	return resp.StatusCode, string(body)
}
