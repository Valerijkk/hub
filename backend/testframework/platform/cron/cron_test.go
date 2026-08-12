//go:build hubtest

package cron_test

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/platform/cron"
)

type cronStopLogHandler struct {
	name    string
	stopped chan struct{}
	once    sync.Once
}

func (h *cronStopLogHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *cronStopLogHandler) Handle(_ context.Context, record slog.Record) error {
	if record.Message != "cron task stopped" {
		return nil
	}

	matched := false
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == "name" && attr.Value.String() == h.name {
			matched = true
		}
		return true
	})
	if matched {
		h.once.Do(func() { close(h.stopped) })
	}
	return nil
}

func (h *cronStopLogHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *cronStopLogHandler) WithGroup(string) slog.Handler      { return h }

func requireSignal(t *testing.T, signal <-chan struct{}, description string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", description)
	}
}

func TestSchedule_RunsFnOnTick(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls atomic.Int32
	cron.Schedule(ctx, 20*time.Millisecond, "test-job", func(_ context.Context) error {
		calls.Add(1)
		return nil
	})

	require.Eventually(t, func() bool {
		return calls.Load() >= 1
	}, time.Second, 10*time.Millisecond)
}

func TestSchedule_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const interval = 10 * time.Millisecond

	stopLog := &cronStopLogHandler{name: "stop-job", stopped: make(chan struct{})}
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(stopLog))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	var calls atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	cron.Schedule(ctx, interval, "stop-job", func(tickCtx context.Context) error {
		if calls.Add(1) == 1 {
			close(started)
			<-release
		}
		return tickCtx.Err()
	})

	requireSignal(t, started, "first cron tick")
	queuedTick := time.NewTimer(2 * interval)
	defer queuedTick.Stop()
	select {
	case <-queuedTick.C:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for a queued cron tick")
	}
	cancel()
	close(release)
	requireSignal(t, stopLog.stopped, "cron runner shutdown")
	assert.Equal(t, int32(1), calls.Load())
}

func TestSchedule_LogsErrorWithoutStopping(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls atomic.Int32
	cron.Schedule(ctx, 15*time.Millisecond, "err-job", func(_ context.Context) error {
		n := calls.Add(1)
		if n == 1 {
			return errors.New("transient")
		}
		cancel()
		return nil
	})

	require.Eventually(t, func() bool {
		return calls.Load() >= 2
	}, time.Second, 10*time.Millisecond)
}

// TestRunTick_TimeoutCancelsHang_A3W44002 — hang > budget → ctx cancel (assigned suite mutation).
// Mutation: убрать WithTimeout из runTick → тест краснеет / зависает до Fatal.
func TestRunTick_TimeoutCancelsHang_A3W44002(t *testing.T) {
	require.Equal(t, 10*time.Minute, cron.ExportDefaultTickTimeout())

	var sawDeadline atomic.Bool
	cron.ExportRunTick(context.Background(), 30*time.Millisecond, "hang-job", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			sawDeadline.Store(true)
			return ctx.Err()
		case <-time.After(2 * time.Second):
			t.Fatal("tick should have been cancelled by timeout")
			return nil
		}
	})
	require.True(t, sawDeadline.Load(), "fn must observe context deadline")

	var second atomic.Bool
	cron.ExportRunTick(context.Background(), time.Second, "next-job", func(ctx context.Context) error {
		second.Store(true)
		return nil
	})
	require.True(t, second.Load(), "subsequent tick must still run")
}
