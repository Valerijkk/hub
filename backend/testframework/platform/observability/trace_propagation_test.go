package observability_test

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"github.com/valerii/hub/backend/internal/platform/observability"
)

const (
	testTraceIDHex = "4bf92f3577b34da6a3ce929d0e0e4736"
	testSpanIDHex  = "00f067aa0ba902b7"
)

// remoteSpanCtx строит контекст с валидным remote span-контекстом —
// имитация trace-context входящего REST-запроса.
func remoteSpanCtx(t *testing.T) context.Context {
	t.Helper()
	traceID, err := trace.TraceIDFromHex(testTraceIDHex)
	require.NoError(t, err)
	spanID, err := trace.SpanIDFromHex(testSpanIDHex)
	require.NoError(t, err)
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})
	return trace.ContextWithSpanContext(context.Background(), sc)
}

// ---------------------------------------------------------------------------
// Round-trip: inject в NATS headers -> extract -> тот же trace-id
// ---------------------------------------------------------------------------

func TestTraceContext_HeaderRoundTrip(t *testing.T) {
	ctx := remoteSpanCtx(t)

	header := nats.Header{}
	natsclient.InjectTraceContext(ctx, header)
	require.NotEmpty(t, header.Get("traceparent"), "traceparent header must be injected")

	extracted := natsclient.ExtractTraceContext(context.Background(), header)
	sc := trace.SpanContextFromContext(extracted)
	require.True(t, sc.IsValid())
	assert.Equal(t, testTraceIDHex, sc.TraceID().String(), "trace id must survive round-trip")
	assert.Equal(t, testSpanIDHex, sc.SpanID().String(), "span id must survive round-trip")
	assert.True(t, sc.IsRemote(), "extracted span context is remote by definition")
}

func TestTraceContext_ExtractEmptyHeaderIsNoop(t *testing.T) {
	ctx := natsclient.ExtractTraceContext(context.Background(), nats.Header{})
	assert.False(t, trace.SpanContextFromContext(ctx).IsValid())
}

// ---------------------------------------------------------------------------
// Round-trip через map (персистентное хранение в outbox-строке)
// ---------------------------------------------------------------------------

func TestTraceContext_MapRoundTrip(t *testing.T) {
	ctx := remoteSpanCtx(t)

	m := natsclient.TraceContextMap(ctx)
	require.NotNil(t, m)
	require.Contains(t, m, "traceparent")

	restored := natsclient.ExtractTraceContextFromMap(context.Background(), m)
	sc := trace.SpanContextFromContext(restored)
	require.True(t, sc.IsValid())
	assert.Equal(t, testTraceIDHex, sc.TraceID().String())
}

func TestTraceContextMap_NilWithoutSpanContext(t *testing.T) {
	assert.Nil(t, natsclient.TraceContextMap(context.Background()),
		"no active span context -> nil map, outbox column stays empty")
}

// ---------------------------------------------------------------------------
// Consumer span: родитель из сообщения, атрибуты subject/event_id
// ---------------------------------------------------------------------------

func TestStartConsumerSpan_ParentFromMessage(t *testing.T) {
	observability.ResetTracingForTest()
	t.Cleanup(observability.ResetTracingForTest)

	sr := tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr)))
	observability.ExportSetTracingActiveForTest(true)

	// Producer-сторона: trace-context уходит в headers.
	header := nats.Header{}
	natsclient.InjectTraceContext(remoteSpanCtx(t), header)
	header.Set(natsclient.HeaderEventID, "evt-42")

	// Consumer-сторона: extract + span.
	ctx := natsclient.ExtractTraceContext(context.Background(), header)
	_, span := observability.StartConsumerSpan(ctx, "tracker.p1.issue.created", header.Get(natsclient.HeaderEventID))
	require.NotNil(t, span)
	observability.EndConsumerSpan(span, nil)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	got := spans[0]
	assert.Equal(t, "consume tracker.p1.issue.created", got.Name())
	assert.Equal(t, trace.SpanKindConsumer, got.SpanKind())
	assert.Equal(t, testTraceIDHex, got.SpanContext().TraceID().String(),
		"consumer span must continue the producer trace")
	assert.Equal(t, testSpanIDHex, got.Parent().SpanID().String(),
		"producer span must be the parent")

	attrs := map[string]string{}
	for _, kv := range got.Attributes() {
		attrs[string(kv.Key)] = kv.Value.AsString()
	}
	assert.Equal(t, "nats", attrs["messaging.system"])
	assert.Equal(t, "tracker.p1.issue.created", attrs["messaging.destination.name"])
	assert.Equal(t, "evt-42", attrs["hub.event_id"])
}

func TestStartConsumerSpan_TracingDisabled(t *testing.T) {
	observability.ResetTracingForTest()
	t.Cleanup(observability.ResetTracingForTest)

	ctx, span := observability.StartConsumerSpan(context.Background(), "any.subject", "")
	assert.Nil(t, span, "no span when tracing is disabled")
	assert.NotNil(t, ctx)
	// EndConsumerSpan обязан быть nil-safe.
	require.NotPanics(t, func() { observability.EndConsumerSpan(nil, nil) })
}
