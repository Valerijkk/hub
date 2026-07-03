package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const consumerTracerName = "hub/consumer"

// StartConsumerSpan открывает span "consume <subject>" для обработки
// NATS-сообщения. Ctx должен уже нести span-контекст продюсера
// (natsclient.ExtractTraceContext) — тогда span станет частью сквозного
// трейса REST -> outbox -> NATS -> consumer.
// При выключенном tracing возвращает nil span (нулевой оверхед);
// завершать через EndConsumerSpan — он nil-safe.
func StartConsumerSpan(ctx context.Context, subject, eventID string) (context.Context, trace.Span) {
	if !TracingEnabled() {
		return ctx, nil
	}
	ctx, span := otel.Tracer(consumerTracerName).Start(ctx, "consume "+subject,
		trace.WithSpanKind(trace.SpanKindConsumer))
	attrs := []attribute.KeyValue{
		attribute.String("messaging.system", "nats"),
		attribute.String("messaging.destination.name", subject),
	}
	if eventID != "" {
		attrs = append(attrs, attribute.String("hub.event_id", eventID))
	}
	span.SetAttributes(attrs...)
	return ctx, span
}

// EndConsumerSpan завершает consumer-span, помечая ошибку при провале.
// Безопасен для nil span (tracing выключен).
func EndConsumerSpan(span trace.Span, err error) {
	finishOperationSpan(span, err)
}
