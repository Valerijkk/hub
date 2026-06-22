package observability_test
import (
	"github.com/valerii/hub/backend/internal/platform/observability"
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestInitTracing_NoEndpointIsNoop(t *testing.T) {
	observability.ResetTracingForTest()
	t.Cleanup(observability.ResetTracingForTest)

	shutdown, err := observability.InitTracing(context.Background(), observability.TracingConfig{})
	require.NoError(t, err)
	require.False(t, observability.TracingEnabled())

	require.NoError(t, shutdown(context.Background()))
}

func TestTimed_CreatesSpanWhenTracingEnabled(t *testing.T) {
	observability.ResetTracingForTest()
	t.Cleanup(observability.ResetTracingForTest)

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	observability.ExportSetTracingActiveForTest(true)

	ctx := context.Background()
	_, err := observability.Timed(ctx, "otel.timed", func(ctx context.Context) (int, error) {
		return 1, nil
	})
	require.NoError(t, err)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, "otel.timed", spans[0].Name())
}

func TestTimed_RecordsErrorOnSpan(t *testing.T) {
	observability.ResetTracingForTest()
	t.Cleanup(observability.ResetTracingForTest)

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	observability.ExportSetTracingActiveForTest(true)

	testErr := errors.New("boom")
	_, _ = observability.Timed(context.Background(), "otel.fail", func(ctx context.Context) (struct{}, error) {
		return struct{}{}, testErr
	})

	spans := sr.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, codes.Error, spans[0].Status().Code)
}

func TestHTTPTraceMiddleware_SkipsMetricsPath(t *testing.T) {
	observability.ResetTracingForTest()
	t.Cleanup(observability.ResetTracingForTest)

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	observability.ExportSetTracingActiveForTest(true)

	app := fiber.New()
	app.Use(observability.HTTPTraceMiddleware())
	app.Get("/metrics", func(c fiber.Ctx) error {
		return c.SendString("metrics")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/metrics", nil)
	_, err := app.Test(req)
	require.NoError(t, err)
	require.Empty(t, sr.Ended())
}

func TestHTTPTraceMiddleware_ExtractsTraceContext(t *testing.T) {
	observability.ResetTracingForTest()
	t.Cleanup(observability.ResetTracingForTest)

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	observability.ExportSetTracingActiveForTest(true)

	app := fiber.New()
	app.Use(observability.HTTPTraceMiddleware())
	app.Get("/api/ping", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/api/ping", nil)
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	_, err := app.Test(req)
	require.NoError(t, err)
	require.Len(t, sr.Ended(), 1)
}

func TestHTTPTraceMiddleware_CreatesServerSpan(t *testing.T) {
	observability.ResetTracingForTest()
	t.Cleanup(observability.ResetTracingForTest)

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	observability.ExportSetTracingActiveForTest(true)

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("requestID", "req-123")
		return c.Next()
	})
	app.Use(observability.HTTPTraceMiddleware())
	app.Get("/api/ping", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/api/ping", nil)
	_, err := app.Test(req)
	require.NoError(t, err)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	require.Contains(t, spans[0].Name(), "GET")
}
