package observability_test
import (
	"github.com/valerii/hub/backend/internal/platform/observability"
	"io"
	"net"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func TestMetricsEndpoint_AllowsWhenNoCIDRs(t *testing.T) {
	observability.ExportRecordOperation("metrics.test", 1, false)
	t.Cleanup(observability.ResetMetrics)

	app := fiber.New()
	app.Get("/metrics", observability.MetricsEndpoint(nil))

	req := httptest.NewRequest(fiber.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	require.Contains(t, string(body), "hub_operation_calls_total")
}

func Test_ipAllowed_respectsCIDR(t *testing.T) {
	nets := observability.ExportParseCIDRs([]string{"10.0.0.0/8", "203.0.113.0/24"})
	require.True(t, observability.ExportIpAllowed(net.ParseIP("10.1.2.3"), nets))
	require.True(t, observability.ExportIpAllowed(net.ParseIP("203.0.113.10"), nets))
	require.False(t, observability.ExportIpAllowed(net.ParseIP("198.51.100.4"), nets))
	require.False(t, observability.ExportIpAllowed(nil, nets))
}

func TestMetricsEndpoint_BlocksWhenNetsEmptyAfterInvalidCIDR(t *testing.T) {
	app := fiber.New()
	app.Get("/metrics", observability.MetricsEndpoint([]string{"not-a-cidr"}))

	req := httptest.NewRequest(fiber.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestMetricsEndpoint_BlocksNonMatchingCIDR(t *testing.T) {
	app := fiber.New()
	app.Get("/metrics", observability.MetricsEndpoint([]string{"10.0.0.0/8"}))

	req := httptest.NewRequest(fiber.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
