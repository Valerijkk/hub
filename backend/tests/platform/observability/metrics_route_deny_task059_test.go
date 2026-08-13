//go:build hubtest

package observability_test

// TASK-059 / P3-BE-03 / W4-A3-009:
// HTTP metrics route label = шаблон :id, не raw UUID path.
// Mutation: c.Path() вместо httpRouteLabel → scrape с UUID → FAIL.
// W4-A3-010 deny peer: уже в TestMetricsEndpoint_BlocksNonMatchingCIDR (metrics_allowlist_test).

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/observability"
)

func TestHTTPMetricsMiddleware_RouteIsTemplate_NotRawUUID(t *testing.T) {
	observability.ResetPrometheusMetrics()
	t.Cleanup(observability.ResetPrometheusMetrics)

	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	app := fiber.New()
	app.Use(observability.HTTPMetricsMiddleware())
	app.Get("/metrics", observability.MetricsHandler())
	app.Get("/api/items/:id", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/api/items/"+id.String(), nil)
	_, err := app.Test(req)
	require.NoError(t, err)

	scrape, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/metrics", nil))
	require.NoError(t, err)
	defer scrape.Body.Close()
	body, err := io.ReadAll(scrape.Body)
	require.NoError(t, err)
	text := string(body)

	require.Contains(t, text, `route="/api/items/:id"`, "metrics must use route template")
	require.NotContains(t, text, id.String(), "raw UUID must not appear in route label")
}

// TestMetricsEndpoint_DeniesOutsideAllowlist_WithProxyPeer — W4-A3-010:
// peer вне CIDR → 403; matching 10.x → 200.
func TestMetricsEndpoint_DeniesOutsideAllowlist_WithProxyPeer(t *testing.T) {
	app := fiber.New(fiber.Config{ProxyHeader: fiber.HeaderXForwardedFor})
	app.Get("/metrics", observability.MetricsEndpoint([]string{"10.0.0.0/8"}))

	reqDeny := httptest.NewRequest(fiber.MethodGet, "/metrics", nil)
	reqDeny.Header.Set(fiber.HeaderXForwardedFor, "192.168.1.1")
	respDeny, err := app.Test(reqDeny)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusForbidden, respDeny.StatusCode)

	reqOK := httptest.NewRequest(fiber.MethodGet, "/metrics", nil)
	reqOK.Header.Set(fiber.HeaderXForwardedFor, "10.1.2.3")
	respOK, err := app.Test(reqOK)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, respOK.StatusCode)
}
