//go:build hubtest

package observability_test

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/valerii/hub/backend/internal/platform/observability"
)

// TestHTTPTraceMiddleware_NoRawPathTarget_W4A3001 — span без http.target с UUID.
func TestHTTPTraceMiddleware_NoRawPathTarget_W4A3001(t *testing.T) {
	observability.ResetTracingForTest()
	t.Cleanup(observability.ResetTracingForTest)

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	observability.ExportSetTracingActiveForTest(true)

	app := fiber.New()
	app.Use(observability.HTTPTraceMiddleware())
	app.Get("/api/issues/:id", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/api/issues/550e8400-e29b-41d4-a716-446655440000", nil)
	_, err := app.Test(req)
	require.NoError(t, err)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	for _, attr := range spans[0].Attributes() {
		require.NotEqual(t, attribute.Key("http.target"), attr.Key, "http.target must not be set")
		if attr.Key == "http.route" {
			require.NotContains(t, attr.Value.AsString(), "550e8400")
			require.Contains(t, attr.Value.AsString(), ":id")
		}
	}
}

// TestHTTPTraceMiddleware_RedactsErrorMessage_W4A3003 — status без err.Error() PII.
func TestHTTPTraceMiddleware_RedactsErrorMessage_W4A3003(t *testing.T) {
	observability.ResetTracingForTest()
	t.Cleanup(observability.ResetTracingForTest)

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	observability.ExportSetTracingActiveForTest(true)

	secret := "user@evil.example token=sekrit"
	app := fiber.New()
	app.Use(observability.HTTPTraceMiddleware())
	app.Get("/boom", func(c fiber.Ctx) error {
		return errors.New(secret)
	})

	_, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/boom", nil))
	require.NoError(t, err)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, codes.Error, spans[0].Status().Code)
	require.Equal(t, "handler_error", spans[0].Status().Description)
	require.NotContains(t, spans[0].Status().Description, secret)
}

// TestTimed_RedactsOperationError_W4A3003 — operation span без raw err text.
func TestTimed_RedactsOperationError_W4A3003(t *testing.T) {
	observability.ResetTracingForTest()
	t.Cleanup(observability.ResetTracingForTest)

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	observability.ExportSetTracingActiveForTest(true)

	secret := "sql: password=hunter2"
	_, _ = observability.Timed(context.Background(), "otel.fail.redact", func(ctx context.Context) (struct{}, error) {
		return struct{}{}, errors.New(secret)
	})

	spans := sr.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, codes.Error, spans[0].Status().Code)
	require.Equal(t, "operation_error", spans[0].Status().Description)
	require.NotContains(t, spans[0].Status().Description, "hunter2")
}

// TestHTTPMetricsMiddleware_UnmatchedNoUUID_W4A3002 — 404 → route=unmatched.
func TestHTTPMetricsMiddleware_UnmatchedNoUUID_W4A3002(t *testing.T) {
	observability.ResetPrometheusMetrics()
	t.Cleanup(observability.ResetPrometheusMetrics)

	app := fiber.New()
	app.Use(observability.HTTPMetricsMiddleware())
	app.Get("/metrics", observability.MetricsHandler())
	app.Get("/api/known", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	uuidPath := "/api/missing/550e8400-e29b-41d4-a716-446655440000"
	_, err := app.Test(httptest.NewRequest(fiber.MethodGet, uuidPath, nil))
	require.NoError(t, err)

	scrape, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/metrics", nil))
	require.NoError(t, err)
	defer scrape.Body.Close()
	body, _ := io.ReadAll(scrape.Body)
	text := string(body)

	require.NotContains(t, text, "550e8400")
	require.Contains(t, text, `route="`+observability.ExportUnmatchedRouteLabel()+`"`)
	require.False(t, strings.Contains(text, uuidPath), "raw UUID path must not appear as route label")
}
