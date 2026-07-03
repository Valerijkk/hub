package nats

import (
	"context"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/propagation"
)

// HeaderEventID — NATS header с ID outbox-события. Consumers кладут его в
// span-атрибуты, чтобы связать трейс с конкретной outbox-записью в БД.
const HeaderEventID = "Hub-Event-Id"

// natsPropagator — явный W3C-пропагатор (TraceContext + Baggage).
// Глобальный otel.GetTextMapPropagator() остаётся no-op, пока InitTracing не
// зарегистрирует настоящий (а это происходит только при заданном OTLP
// endpoint). Завязка на глобальный сделала бы перенос trace-context через
// NATS зависимым от конфигурации экспорта; явный пропагатор работает всегда
// и детерминирован в тестах.
var natsPropagator = propagation.NewCompositeTextMapPropagator(
	propagation.TraceContext{},
	propagation.Baggage{},
)

// headerCarrier — propagation.TextMapCarrier поверх nats.Header.
type headerCarrier nats.Header

func (h headerCarrier) Get(key string) string { return nats.Header(h).Get(key) }

func (h headerCarrier) Set(key, value string) { nats.Header(h).Set(key, value) }

func (h headerCarrier) Keys() []string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	return keys
}

// InjectTraceContext кладёт trace-context из ctx (traceparent/tracestate,
// baggage) в NATS message headers.
func InjectTraceContext(ctx context.Context, h nats.Header) {
	natsPropagator.Inject(ctx, headerCarrier(h))
}

// ExtractTraceContext возвращает контекст со span-контекстом, восстановленным
// из NATS message headers. Без traceparent в headers возвращает ctx как есть.
func ExtractTraceContext(ctx context.Context, h nats.Header) context.Context {
	if len(h) == 0 {
		return ctx
	}
	return natsPropagator.Extract(ctx, headerCarrier(h))
}

// TraceContextMap сериализует trace-context из ctx в плоскую map для
// персистентного хранения (outbox-строка). Возвращает nil, если активного
// span-контекста нет — колонка в БД остаётся пустой.
func TraceContextMap(ctx context.Context) map[string]string {
	carrier := propagation.MapCarrier{}
	natsPropagator.Inject(ctx, carrier)
	if len(carrier) == 0 {
		return nil
	}
	return carrier
}

// ExtractTraceContextFromMap восстанавливает span-контекст из map,
// сохранённой через TraceContextMap.
func ExtractTraceContextFromMap(ctx context.Context, m map[string]string) context.Context {
	if len(m) == 0 {
		return ctx
	}
	return natsPropagator.Extract(ctx, propagation.MapCarrier(m))
}
