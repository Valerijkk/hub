package outbox_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/valerii/hub/backend/ent/outboxevent"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"github.com/valerii/hub/backend/internal/platform/outbox"
)

const (
	outboxTraceIDHex = "4bf92f3577b34da6a3ce929d0e0e4736"
	outboxSpanIDHex  = "00f067aa0ba902b7"
)

// spanCtx строит контекст с валидным span-контекстом — имитация REST-запроса,
// в рамках которого сервис пишет событие в outbox.
func spanCtx(t *testing.T) context.Context {
	t.Helper()
	traceID, err := trace.TraceIDFromHex(outboxTraceIDHex)
	require.NoError(t, err)
	spanID, err := trace.SpanIDFromHex(outboxSpanIDHex)
	require.NoError(t, err)
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	return trace.ContextWithSpanContext(context.Background(), sc)
}

// ---------------------------------------------------------------------------
// Publisher сохраняет trace-context в outbox-строке
// ---------------------------------------------------------------------------

func TestPublish_StoresTraceContext(t *testing.T) {
	client := newTestEntClient(t)
	p := outbox.NewPublisher()

	require.NoError(t, p.Publish(spanCtx(t), client, "trace.stored", map[string]any{"k": "v"}))

	evt, err := client.OutboxEvent.Query().
		Where(outboxevent.EventTypeEQ("trace.stored")).
		Only(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, evt.TraceContext, "trace context must be persisted with the event")
	assert.Contains(t, evt.TraceContext["traceparent"], outboxTraceIDHex)

	// Восстановление: сохранённая map -> тот же trace-id.
	restored := natsclient.ExtractTraceContextFromMap(context.Background(), evt.TraceContext)
	sc := trace.SpanContextFromContext(restored)
	require.True(t, sc.IsValid())
	assert.Equal(t, outboxTraceIDHex, sc.TraceID().String())
}

func TestPublish_NoSpanContext_LeavesTraceContextEmpty(t *testing.T) {
	client := newTestEntClient(t)
	p := outbox.NewPublisher()

	require.NoError(t, p.Publish(context.Background(), client, "trace.absent", map[string]any{"k": "v"}))

	evt, err := client.OutboxEvent.Query().
		Where(outboxevent.EventTypeEQ("trace.absent")).
		Only(context.Background())
	require.NoError(t, err)
	assert.Empty(t, evt.TraceContext, "no active span -> no trace context stored")
}

func TestPublishInTx_StoresTraceContext(t *testing.T) {
	client := newTestEntClient(t)
	p := outbox.NewPublisher()
	ctx := spanCtx(t)

	tx, err := client.Tx(context.Background())
	require.NoError(t, err)
	require.NoError(t, p.PublishInTx(ctx, tx, "trace.in_tx", map[string]any{"k": "v"}))
	require.NoError(t, tx.Commit())

	evt, err := client.OutboxEvent.Query().
		Where(outboxevent.EventTypeEQ("trace.in_tx")).
		Only(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, evt.TraceContext)
	assert.Contains(t, evt.TraceContext["traceparent"], outboxTraceIDHex)
}

// ---------------------------------------------------------------------------
// Worker переносит trace-context в NATS headers при publish
// ---------------------------------------------------------------------------

func TestProcessEvent_PublishesTraceparentHeader(t *testing.T) {
	client := newTestEntClient(t)
	nc := startTestNATSServer(t)
	w := outbox.NewWorker(client, nc)
	ctx := context.Background()

	sub, err := nc.Conn.SubscribeSync("project.outbox.trace.headers")
	require.NoError(t, err)
	t.Cleanup(func() { _ = sub.Unsubscribe() })

	// Событие с сохранённым trace-context (как его записал бы Publisher).
	evt, err := client.OutboxEvent.Create().
		SetEventType("project.outbox.trace.headers").
		SetPayload([]byte(`{"ok":true}`)).
		SetTraceContext(natsclient.TraceContextMap(spanCtx(t))).
		Save(ctx)
	require.NoError(t, err)

	outbox.ExportWorkerProcessEvent(w, ctx, evt)

	msg, err := sub.NextMsg(2 * time.Second)
	require.NoError(t, err)
	assert.Contains(t, msg.Header.Get("traceparent"), outboxTraceIDHex,
		"worker must forward the stored trace context into NATS headers")
	assert.Equal(t, evt.ID.String(), msg.Header.Get(natsclient.HeaderEventID),
		"worker must forward the outbox event id for consumer span attributes")

	// Payload не тронут — trace-context едет в headers, не в envelope.
	assert.JSONEq(t, `{"ok":true}`, string(msg.Data))

	// Consumer-сторона: headers -> тот же trace-id.
	extracted := natsclient.ExtractTraceContext(context.Background(), msg.Header)
	sc := trace.SpanContextFromContext(extracted)
	require.True(t, sc.IsValid())
	assert.Equal(t, outboxTraceIDHex, sc.TraceID().String(),
		"trace must stay unbroken across REST -> outbox -> NATS -> consumer")
}

func TestProcessEvent_NoTraceContext_PublishesWithoutTraceparent(t *testing.T) {
	client := newTestEntClient(t)
	nc := startTestNATSServer(t)
	w := outbox.NewWorker(client, nc)
	ctx := context.Background()

	sub, err := nc.Conn.SubscribeSync("project.outbox.trace.none")
	require.NoError(t, err)
	t.Cleanup(func() { _ = sub.Unsubscribe() })

	evt := insertOutboxEvent(t, client, "project.outbox.trace.none", []byte(`{"ok":true}`))
	outbox.ExportWorkerProcessEvent(w, ctx, evt)

	msg, err := sub.NextMsg(2 * time.Second)
	require.NoError(t, err)
	assert.Empty(t, msg.Header.Get("traceparent"),
		"legacy events without stored trace context must not invent one")
	assert.Equal(t, evt.ID.String(), msg.Header.Get(natsclient.HeaderEventID))

	// Событие помечено published — headers не ломают happy-path.
	reloaded, err := client.OutboxEvent.Get(ctx, evt.ID)
	require.NoError(t, err)
	assert.Equal(t, outboxevent.StatusPublished, reloaded.Status)
}
