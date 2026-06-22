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
