package observability_test
import (
	"github.com/valerii/hub/backend/internal/platform/observability"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

// TestMetricsAllowMiddleware_EmptyCIDRsForbidden — W4-A3-004: пусто → 403.
func TestMetricsAllowMiddleware_EmptyCIDRsForbidden(t *testing.T) {
	app := fiber.New()
	app.Use(observability.MetricsAllowMiddleware(nil))
	app.Get("/metrics", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestMetricsAllowMiddleware_AllowAllSentinel(t *testing.T) {
	app := fiber.New()
	app.Use(observability.MetricsAllowMiddleware(observability.MetricsAllowAllCIDRs))
	app.Get("/metrics", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestMetricsAllowMiddleware_AllowsMatchingIP(t *testing.T) {
	app := fiber.New(fiber.Config{ProxyHeader: fiber.HeaderXForwardedFor})
	app.Use(observability.MetricsAllowMiddleware([]string{"127.0.0.0/8"}))
	app.Get("/metrics", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/metrics", nil)
	req.Header.Set(fiber.HeaderXForwardedFor, "127.0.0.1")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestMetricsAllowMiddleware_DeniesInvalidCIDRList(t *testing.T) {
	app := fiber.New()
	app.Use(observability.MetricsAllowMiddleware([]string{"not-cidr"}))
	app.Get("/metrics", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
