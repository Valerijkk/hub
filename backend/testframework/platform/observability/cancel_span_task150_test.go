//go:build hubtest

package observability_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/valerii/hub/backend/internal/platform/observability"
)

// TestTimed_CancelNotErrorStatus_A3W4_3_001 — cancel/deadline ≠ span Error / error_count.
func TestTimed_CancelNotErrorStatus_A3W4_3_001(t *testing.T) {
	observability.ResetTracingForTest()
	observability.ResetMetrics()
	t.Cleanup(func() {
		observability.ResetTracingForTest()
		observability.ResetMetrics()
	})

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	observability.ExportSetTracingActiveForTest(true)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := observability.Timed(ctx, "test.cancel", func(ctx context.Context) (int, error) {
		return 0, ctx.Err()
	})
	require.ErrorIs(t, err, context.Canceled)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.NotEqual(t, codes.Error, spans[0].Status().Code,
		"A3W4-3-001: context.Canceled must not set span Error-status")

	snap := observability.Snapshot()
	require.Len(t, snap, 1)
	assert.Equal(t, "test.cancel", snap[0].Operation)
	assert.Equal(t, int64(1), snap[0].Count)
	assert.Equal(t, int64(0), snap[0].ErrorCount,
		"A3W4-3-001: cancel must not increment error_count")
}

func TestTimed_DeadlineNotErrorStatus_A3W4_3_001(t *testing.T) {
	observability.ResetTracingForTest()
	observability.ResetMetrics()
	t.Cleanup(func() {
		observability.ResetTracingForTest()
		observability.ResetMetrics()
	})

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	observability.ExportSetTracingActiveForTest(true)

	_, err := observability.Timed(context.Background(), "test.deadline", func(ctx context.Context) (int, error) {
		return 0, context.DeadlineExceeded
	})
	require.ErrorIs(t, err, context.DeadlineExceeded)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.NotEqual(t, codes.Error, spans[0].Status().Code)

	snap := observability.Snapshot()
	require.Len(t, snap, 1)
	assert.Equal(t, int64(0), snap[0].ErrorCount)
}

func TestEndConsumerSpan_CancelNotErrorStatus_A3W4_3_001(t *testing.T) {
	observability.ResetTracingForTest()
	t.Cleanup(observability.ResetTracingForTest)

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	observability.ExportSetTracingActiveForTest(true)

	_, span := observability.StartConsumerSpan(context.Background(), "hub.test.event", "evt-1")
	observability.EndConsumerSpan(span, context.Canceled)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.NotEqual(t, codes.Error, spans[0].Status().Code,
		"A3W4-3-001: EndConsumerSpan(cancel) must not set Error-status")
}

func TestTimed_RealErrorStillErrorStatus_A3W4_3_001(t *testing.T) {
	observability.ResetTracingForTest()
	observability.ResetMetrics()
	t.Cleanup(func() {
		observability.ResetTracingForTest()
		observability.ResetMetrics()
	})

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	observability.ExportSetTracingActiveForTest(true)

	boom := errors.New("boom")
	_, err := observability.Timed(context.Background(), "test.realerr", func(ctx context.Context) (int, error) {
		return 0, boom
	})
	require.ErrorIs(t, err, boom)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status().Code)

	snap := observability.Snapshot()
	require.Len(t, snap, 1)
	assert.Equal(t, int64(1), snap[0].ErrorCount)
}
