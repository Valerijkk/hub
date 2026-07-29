//go:build hubtest

package cron_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/platform/cron"
)

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

	var calls atomic.Int32
	cron.Schedule(ctx, 10*time.Millisecond, "stop-job", func(_ context.Context) error {
		calls.Add(1)
		return nil
	})

	require.Eventually(t, func() bool {
		return calls.Load() >= 1
	}, time.Second, 10*time.Millisecond)

	before := calls.Load()
	cancel()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, before, calls.Load())
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
