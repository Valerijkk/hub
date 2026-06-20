package rbac

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireAuth_MissingHeader(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	}, RequireAuth())

	status, body := doGet(t, app, "/")
	assert.Equal(t, fiber.StatusUnauthorized, status)
	assert.Contains(t, body, "missing authorization header")
}

func TestRequireAuth_InvalidFormat(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	}, RequireAuth())

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token abc")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRequireAuth_EmptyBearerToken(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	}, RequireAuth())

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer ")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRequireAuth_ValidBearerNext(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		token, _ := c.Locals("token").(string)
		return c.SendString(token)
	}, RequireAuth())

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer jwt-token-123")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestUserIDFromCtx(t *testing.T) {
	app := fiber.New()
	userID := uuid.New().String()
	app.Get("/", func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.SendString(UserIDFromCtx(c))
	})

	status, body := doGet(t, app, "/")
	assert.Equal(t, fiber.StatusOK, status)
	assert.Equal(t, userID, body)

	app2 := fiber.New()
	app2.Get("/", func(c fiber.Ctx) error {
		return c.SendString(UserIDFromCtx(c))
	})
	status, body = doGet(t, app2, "/")
	assert.Equal(t, fiber.StatusOK, status)
	assert.Empty(t, body)
}
