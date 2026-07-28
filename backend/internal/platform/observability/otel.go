package observability

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"

	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	httpTracerName      = "hub/http"
	operationTracerName = "hub/operation"
)

var tracingActive atomic.Bool

// TracingConfig — экспорт трейсов OTLP (OBS-01 / FB-21).
type TracingConfig struct {
	ServiceName  string
	OTLPEndpoint string
}

// TracingEnabled — зарегистрирован ли реальный TracerProvider.
func TracingEnabled() bool {
	return tracingActive.Load()
}

// InitTracing регистрирует global TracerProvider при OTLPEndpoint.
// OTEL_EXPORTER_OTLP_ENDPOINT если cfg.OTLPEndpoint пуст.
func InitTracing(ctx context.Context, cfg TracingConfig) (func(context.Context) error, error) {
	endpoint := strings.TrimSpace(cfg.OTLPEndpoint)
	if endpoint == "" {
		endpoint = strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	}
	if endpoint == "" {
		return func(context.Context) error { return nil }, nil
	}

	serviceName := strings.TrimSpace(cfg.ServiceName)
	if serviceName == "" {
		serviceName = strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	}
	if serviceName == "" {
		serviceName = "hub-backend"
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp trace exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			attribute.String("service.name", serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	tracingActive.Store(true)

	slog.InfoContext(ctx, "OpenTelemetry tracing enabled",
		"service", serviceName,
		"otlp_endpoint", endpoint,
	)

	return tp.Shutdown, nil
}

// HTTPTraceMiddleware — server span на запрос и propagation context.
// W4-A3-001: только http.route (шаблон), без http.target с UUID.
// W4-A3-003: status/RecordError без сырого err.Error() (PII/internals).
func HTTPTraceMiddleware() fiber.Handler {
	tracer := otel.Tracer(httpTracerName)
	prop := otel.GetTextMapPropagator()

	return func(c fiber.Ctx) error {
		if c.Path() == metricsPath {
			return c.Next()
		}

		carrier := fiberHeaderCarrier{c: c}
		ctx := prop.Extract(c.UserContext(), carrier)

		route := httpRouteLabel(c)
		spanName := c.Method() + " " + route

		ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.route", route),
		)
		if requestID, ok := c.Locals("requestID").(string); ok && requestID != "" {
			span.SetAttributes(attribute.String("hub.request_id", requestID))
		}

		c.SetUserContext(ctx)

		err := c.Next()

		route = httpRouteLabel(c)

		status := c.Response().StatusCode()
		if status == 0 {
			status = fiber.StatusOK
		}
		if err != nil {
			if fe, ok := err.(*fiber.Error); ok && fe.Code > 0 {
				status = fe.Code
			} else if status < fiber.StatusBadRequest {
				status = fiber.StatusInternalServerError
			}
			recordRedactedSpanError(span, "handler_error")
		}
		span.SetAttributes(
			attribute.String("http.route", route),
			attribute.Int("http.response.status_code", status),
		)

		return err
	}
}

type fiberHeaderCarrier struct {
	c fiber.Ctx
}

func (h fiberHeaderCarrier) Get(key string) string {
	return h.c.Get(key)
}

func (h fiberHeaderCarrier) Set(key, value string) {
	h.c.Set(key, value)
}

func (h fiberHeaderCarrier) Keys() []string {
	keys := make([]string, 0, 8)
	h.c.Request().Header.VisitAll(func(k, _ []byte) {
		keys = append(keys, string(k))
	})
	return keys
}

// startOperationSpan — child span при активном tracing.
func startOperationSpan(ctx context.Context, operation string) (context.Context, trace.Span) {
	if !TracingEnabled() {
		return ctx, nil
	}
	return otel.Tracer(operationTracerName).Start(ctx, operation)
}

func finishOperationSpan(span trace.Span, err error) {
	if span == nil {
		return
	}
	if err != nil {
		recordRedactedSpanError(span, "operation_error")
	}
	span.End()
}

// recordRedactedSpanError пишет Error-status без сырого текста ошибки (W4-A3-003).
func recordRedactedSpanError(span trace.Span, code string) {
	span.RecordError(fmt.Errorf("%s", code))
	span.SetStatus(codes.Error, code)
}

// ResetTracingForTest — сброс tracing flag (только tests).
func ResetTracingForTest() {
	tracingActive.Store(false)
	otel.SetTracerProvider(nil)
}
