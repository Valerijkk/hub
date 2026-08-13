package dev_test

// W7-A5-001: /api/docs за auth+admin — unit-проверка порядка middleware Fiber v3
// через изолированный router (полный wiring требует DB).
// Mutation: снять requireAuth/RequireAdmin с /docs → anon 200 → FAIL.

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/dev"
)

func TestAPIDocs_RequiresAuthAdmin_W7A5001(t *testing.T) {
	app := fiber.New()
	api := app.Group("/api")

	requireAuth := func(c fiber.Ctx) error {
		if c.Get("X-Test-User-ID") == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}
		c.Locals("userID", c.Get("X-Test-User-ID"))
		return c.Next()
	}
	requireAdmin := func(c fiber.Ctx) error {
		if c.Get("X-Test-Role") != "admin" {
			return fiber.NewError(fiber.StatusForbidden, "insufficient system role")
		}
		return c.Next()
	}

	// Как в wireCore: middleware после handler в сигнатуре → выполняются до.
	api.Get("/docs", func(c fiber.Ctx) error {
		return c.Type("html").SendString(dev.BuildAPIDocsHTML(app))
	}, requireAuth, requireAdmin)
	api.Get("/docs/doc.json", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		body, err := dev.BuildAPIDocsJSON(app)
		if err != nil {
			return err
		}
		return c.SendString(body)
	}, requireAuth, requireAdmin)

	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	req.Header.Set("X-Test-User-ID", "u1")
	req.Header.Set("X-Test-Role", "member")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	req.Header.Set("X-Test-User-ID", "u1")
	req.Header.Set("X-Test-Role", "admin")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Hub Platform API")

	req = httptest.NewRequest(http.MethodGet, "/api/docs/doc.json", nil)
	req.Header.Set("X-Test-User-ID", "u1")
	req.Header.Set("X-Test-Role", "admin")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
