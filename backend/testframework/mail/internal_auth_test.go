package mail_test
import (
	"github.com/valerii/hub/backend/internal/mail"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInternalTokenAuth(t *testing.T) {
	app := fiber.New()
	app.Use("/v1/send", mail.InternalTokenAuth("secret-token"))
	app.Post("/v1/send", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	t.Run("missing token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/send", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("wrong token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/send", nil)
		req.Header.Set(mail.ExportInternalTokenHeader(), "wrong")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("valid header token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/send", nil)
		req.Header.Set(mail.ExportInternalTokenHeader(), "secret-token")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("valid bearer token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/send", nil)
		req.Header.Set("Authorization", "Bearer secret-token")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("empty expected token", func(t *testing.T) {
		locked := fiber.New()
		locked.Use("/v1/send", mail.InternalTokenAuth(""))
		locked.Post("/v1/send", func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
		req := httptest.NewRequest("POST", "/v1/send", nil)
		req.Header.Set(mail.ExportInternalTokenHeader(), "anything")
		resp, err := locked.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	})
}
