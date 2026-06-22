package rbac_test

// Тесты middleware RBAC: интеграционные и unit-сценарии testframework.
import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/rbac"
)

// TestRequireAuth_MissingHeader проверяет отклонение при require auth отсутствующий header.
func TestRequireAuth_MissingHeader(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	}, rbac.RequireAuth())

	status, body := doGet(t, app, "/")
	assert.Equal(t, fiber.StatusUnauthorized, status)
	assert.Contains(t, body, "missing authorization header")
}

// TestRequireAuth_InvalidFormat проверяет отклонение при require auth некорректный форматирование.
func TestRequireAuth_InvalidFormat(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	}, rbac.RequireAuth())

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token abc")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestRequireAuth_EmptyBearerToken проверяет сценарий: require auth пустой bearer token.
func TestRequireAuth_EmptyBearerToken(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	}, rbac.RequireAuth())

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer ")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestRequireAuth_ValidBearerNext проверяет сценарий: require auth корректный bearer next.
func TestRequireAuth_ValidBearerNext(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		token, _ := c.Locals("token").(string)
		return c.SendString(token)
	}, rbac.RequireAuth())

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer jwt-token-123")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
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
