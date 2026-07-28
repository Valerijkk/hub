//go:build hubtest

package cron_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/platform/cron"
)

// TestRunTick_TimeoutCancelsHang_W4A4008 — fn sleep > budget → ctx cancel, следующий tick идёт.
func TestRunTick_TimeoutCancelsHang_W4A4008(t *testing.T) {
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

// TestScheduleWithTimeout_BudgetHonored — короткий budget на ScheduleWithTimeout.
func TestScheduleWithTimeout_BudgetHonored(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var timedOut atomic.Int32
	cron.ScheduleWithTimeout(ctx, 20*time.Millisecond, 25*time.Millisecond, "budget-job", func(c context.Context) error {
		select {
		case <-c.Done():
			timedOut.Add(1)
			cancel()
			return c.Err()
		case <-time.After(500 * time.Millisecond):
			return nil
		}
	})

	require.Eventually(t, func() bool {
		return timedOut.Load() >= 1
	}, time.Second, 10*time.Millisecond)
}
