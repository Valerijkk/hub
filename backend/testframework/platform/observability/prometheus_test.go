package observability_test
import (
	"github.com/valerii/hub/backend/internal/platform/observability"
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestMetricsHandler_ExportsOperationMetrics(t *testing.T) {
	observability.ResetMetrics()
	t.Cleanup(observability.ResetMetrics)

	observability.ExportRecordOperation("prom.test", 42, false)

	app := fiber.New()
	app.Get("/metrics", observability.MetricsHandler())

	req := httptest.NewRequest(fiber.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "hub_operation_calls_total") {
		t.Fatalf("expected hub_operation_calls_total in scrape, got: %s", text[:min(len(text), 500)])
	}
	if !strings.Contains(text, `operation="prom.test"`) {
		t.Fatalf("expected prom.test label in scrape")
	}
}

func TestHTTPMetricsMiddleware_RecordsRequest(t *testing.T) {
	observability.ResetPrometheusMetrics()
	t.Cleanup(observability.ResetPrometheusMetrics)

	app := fiber.New()
	app.Use(observability.HTTPMetricsMiddleware())
	app.Get("/metrics", observability.MetricsHandler())
	app.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("pong")
	})

	if _, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/ping", nil)); err != nil {
		t.Fatalf("ping: %v", err)
	}

	scrape, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/metrics", nil))
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	defer scrape.Body.Close()

	body, _ := io.ReadAll(scrape.Body)
	text := string(body)
	if !strings.Contains(text, "hub_http_requests_total") {
		t.Fatalf("expected hub_http_requests_total, sample: %s", text[:min(len(text), 500)])
	}
}

func TestTimed_RecordsPrometheus(t *testing.T) {
	observability.ResetMetrics()
	t.Cleanup(observability.ResetMetrics)

	ctx := context.Background()
	_, _ = observability.Timed(ctx, "prom.timed", func(ctx context.Context) (struct{}, error) {
		return struct{}{}, nil
	})

	app := fiber.New()
	app.Get("/metrics", observability.MetricsHandler())

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/metrics", nil))
	if err != nil {
		t.Fatalf("scrape: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), `operation="prom.timed"`) {
		t.Fatalf("expected prom.timed in prometheus output")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
