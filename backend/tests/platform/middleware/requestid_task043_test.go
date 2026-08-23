//go:build hubtest

package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/middleware"
)

func TestValidRequestID_W4A1002(t *testing.T) {
	assert.True(t, middleware.ExportValidRequestID(uuid.NewString()))
	assert.True(t, middleware.ExportValidRequestID("my-custom-request-id-123"))
	assert.False(t, middleware.ExportValidRequestID(""))
	assert.False(t, middleware.ExportValidRequestID("bad id with spaces"))
	assert.False(t, middleware.ExportValidRequestID(strings.Repeat("a", 129)))
	assert.False(t, middleware.ExportValidRequestID("evil<script>"))
}

func TestRequestID_RejectsOversizedHeader_W4A1002(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(c.Locals("requestID").(string))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", strings.Repeat("x", 200))
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	id := string(body)
	_, parseErr := uuid.Parse(id)
	assert.NoError(t, parseErr, "oversized client ID must be replaced with UUID")
	assert.Equal(t, id, resp.Header.Get("X-Request-ID"))
}
