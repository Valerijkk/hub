package messenger_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

var _ messenger.Phase151PendingRecoveryIntervalGate = (*messenger.Phase151ProcessLocalPendingRecoveryIntervalGate)(nil)

type phase151PendingRecoveryIntervalGateClock struct {
	mu      sync.Mutex
	now     time.Time
	calls   int
	entered chan<- struct{}
	release <-chan struct{}
}

func (c *phase151PendingRecoveryIntervalGateClock) Now() time.Time {
	c.mu.Lock()
	c.calls++
	now := c.now
	entered := c.entered
	release := c.release
	c.mu.Unlock()

	if entered != nil {
		entered <- struct{}{}
	}
	if release != nil {
		<-release
	}
	return now
}

func (c *phase151PendingRecoveryIntervalGateClock) Calls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

func (c *phase151PendingRecoveryIntervalGateClock) SetNow(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = now
}

type phase151PendingRecoveryIntervalGateTypedNilContext struct{}

func (*phase151PendingRecoveryIntervalGateTypedNilContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (*phase151PendingRecoveryIntervalGateTypedNilContext) Done() <-chan struct{} {
	return nil
}

func (*phase151PendingRecoveryIntervalGateTypedNilContext) Err() error {
	return nil
}

func (*phase151PendingRecoveryIntervalGateTypedNilContext) Value(any) any {
	return nil
}

type phase151PendingRecoveryIntervalGateObservedContext struct {
	context.Context
	checked chan<- struct{}
}

func (c phase151PendingRecoveryIntervalGateObservedContext) Err() error {
	if c.checked != nil {
		select {
		case c.checked <- struct{}{}:
		default:
		}
	}
	return c.Context.Err()
}

func TestPhase151ProcessLocalPendingRecoveryIntervalGateConstructorAndNilReceiver(t *testing.T) {
	clock := &phase151PendingRecoveryIntervalGateClock{now: time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)}

	require.NotNil(t, messenger.NewPhase151ProcessLocalPendingRecoveryIntervalGate(clock))
	require.Nil(t, messenger.NewPhase151ProcessLocalPendingRecoveryIntervalGate(nil))

	var typedNilClock *phase151PendingRecoveryIntervalGateClock
	require.Nil(t, messenger.NewPhase151ProcessLocalPendingRecoveryIntervalGate(typedNilClock))

	var gate *messenger.Phase151ProcessLocalPendingRecoveryIntervalGate
	admitted, err := gate.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)

	require.False(t, admitted)
	require.Equal(t, messenger.ErrPostCreatePipelineInvalid, err)
	require.Zero(t, clock.Calls())
}

func TestPhase151ProcessLocalPendingRecoveryIntervalGateRejectsInvalidContextBeforeClockAccess(t *testing.T) {
	clock := &phase151PendingRecoveryIntervalGateClock{now: time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)}
	gate := messenger.NewPhase151ProcessLocalPendingRecoveryIntervalGate(clock)
	require.NotNil(t, gate)

	admitted, err := gate.TryAcquirePendingRecoveryInterval(nil, time.Minute)
	require.False(t, admitted)
	require.Equal(t, messenger.ErrPostCreatePipelineInvalid, err)

	var typedNilContext *phase151PendingRecoveryIntervalGateTypedNilContext
	admitted, err = gate.TryAcquirePendingRecoveryInterval(typedNilContext, time.Minute)
	require.False(t, admitted)
	require.Equal(t, messenger.ErrPostCreatePipelineInvalid, err)

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	admitted, err = gate.TryAcquirePendingRecoveryInterval(cancelled, 0)
	require.False(t, admitted)
	require.Equal(t, context.Canceled, err)
	require.Zero(t, clock.Calls())
}

func TestPhase151ProcessLocalPendingRecoveryIntervalGateRejectsNonPositiveMinimumWithoutClockAccess(t *testing.T) {
	clock := &phase151PendingRecoveryIntervalGateClock{now: time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)}
	gate := messenger.NewPhase151ProcessLocalPendingRecoveryIntervalGate(clock)
	require.NotNil(t, gate)

	for _, minimum := range []time.Duration{0, -time.Nanosecond} {
		admitted, err := gate.TryAcquirePendingRecoveryInterval(context.Background(), minimum)
		require.False(t, admitted)
		require.Equal(t, messenger.ErrPostCreatePipelineInvalid, err)
	}

	require.Zero(t, clock.Calls())
}

func TestPhase151ProcessLocalPendingRecoveryIntervalGateAdmitsAtFirstCallAndExactBoundary(t *testing.T) {
	base := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	clock := &phase151PendingRecoveryIntervalGateClock{now: base}
	gate := messenger.NewPhase151ProcessLocalPendingRecoveryIntervalGate(clock)
	require.NotNil(t, gate)

	admitted, err := gate.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
	require.True(t, admitted)
	require.NoError(t, err)

	clock.SetNow(base.Add(time.Minute - time.Nanosecond))
	admitted, err = gate.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
	require.False(t, admitted)
	require.NoError(t, err)

	clock.SetNow(base.Add(time.Minute))
	admitted, err = gate.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
	require.True(t, admitted)
	require.NoError(t, err)
}

func TestPhase151ProcessLocalPendingRecoveryIntervalGateRejectsInvalidClockWithoutMutatingAdmission(t *testing.T) {
	base := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)

	for name, invalidNow := range map[string]time.Time{
		"zero timestamp": time.Time{},
		"backward clock": base.Add(-time.Nanosecond),
	} {
		t.Run(name, func(t *testing.T) {
			clock := &phase151PendingRecoveryIntervalGateClock{now: base}
			gate := messenger.NewPhase151ProcessLocalPendingRecoveryIntervalGate(clock)
			require.NotNil(t, gate)

			admitted, err := gate.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
			require.True(t, admitted)
			require.NoError(t, err)

			clock.SetNow(invalidNow)
			admitted, err = gate.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
			require.False(t, admitted)
			require.Equal(t, messenger.ErrPostCreatePipelineInvalid, err)

			clock.SetNow(base.Add(time.Minute - time.Nanosecond))
			admitted, err = gate.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
			require.False(t, admitted)
			require.NoError(t, err)

			clock.SetNow(base.Add(time.Minute))
			admitted, err = gate.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
			require.True(t, admitted)
			require.NoError(t, err)
		})
	}
}

func TestPhase151ProcessLocalPendingRecoveryIntervalGateRechecksCancellationWhileQueued(t *testing.T) {
	base := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	clockEntered := make(chan struct{}, 1)
	clockRelease := make(chan struct{})
	clock := &phase151PendingRecoveryIntervalGateClock{
		now: base, entered: clockEntered, release: clockRelease,
	}
	gate := messenger.NewPhase151ProcessLocalPendingRecoveryIntervalGate(clock)
	require.NotNil(t, gate)

	type acquisition struct {
		admitted bool
		err      error
	}
	firstResult := make(chan acquisition, 1)
	go func() {
		admitted, err := gate.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
		firstResult <- acquisition{admitted: admitted, err: err}
	}()
	<-clockEntered

	queuedContext, cancel := context.WithCancel(context.Background())
	queuedChecks := make(chan struct{}, 2)
	queuedResult := make(chan acquisition, 1)
	go func() {
		admitted, err := gate.TryAcquirePendingRecoveryInterval(
			phase151PendingRecoveryIntervalGateObservedContext{Context: queuedContext, checked: queuedChecks},
			time.Minute,
		)
		queuedResult <- acquisition{admitted: admitted, err: err}
	}()
	<-queuedChecks
	cancel()
	close(clockRelease)

	first := <-firstResult
	require.True(t, first.admitted)
	require.NoError(t, first.err)

	queued := <-queuedResult
	require.False(t, queued.admitted)
	require.Equal(t, context.Canceled, queued.err)
	require.Equal(t, 1, clock.Calls())
}

func TestPhase151ProcessLocalPendingRecoveryIntervalGateRechecksCancellationAfterClockRead(t *testing.T) {
	base := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	clockEntered := make(chan struct{}, 1)
	clockRelease := make(chan struct{})
	clock := &phase151PendingRecoveryIntervalGateClock{
		now: base, entered: clockEntered, release: clockRelease,
	}
	gate := messenger.NewPhase151ProcessLocalPendingRecoveryIntervalGate(clock)
	require.NotNil(t, gate)

	type acquisition struct {
		admitted bool
		err      error
	}
	cancelledContext, cancel := context.WithCancel(context.Background())
	result := make(chan acquisition, 1)
	go func() {
		admitted, err := gate.TryAcquirePendingRecoveryInterval(cancelledContext, time.Minute)
		result <- acquisition{admitted: admitted, err: err}
	}()
	<-clockEntered
	cancel()
	close(clockRelease)

	cancelled := <-result
	require.False(t, cancelled.admitted)
	require.Equal(t, context.Canceled, cancelled.err)
	require.Equal(t, 1, clock.Calls())

	admitted, err := gate.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
	require.True(t, admitted)
	require.NoError(t, err)
	require.Equal(t, 2, clock.Calls())
}
func TestPhase151ProcessLocalPendingRecoveryIntervalGateAdmitsExactlyOneConcurrentCaller(t *testing.T) {
	base := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	clock := &phase151PendingRecoveryIntervalGateClock{now: base}
	gate := messenger.NewPhase151ProcessLocalPendingRecoveryIntervalGate(clock)
	require.NotNil(t, gate)

	const callers = 16
	type acquisition struct {
		admitted bool
		err      error
	}
	start := make(chan struct{})
	results := make(chan acquisition, callers)
	var ready sync.WaitGroup
	var done sync.WaitGroup
	ready.Add(callers)
	done.Add(callers)
	for range callers {
		go func() {
			ready.Done()
			<-start
			admitted, err := gate.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
			results <- acquisition{admitted: admitted, err: err}
			done.Done()
		}()
	}
	ready.Wait()
	close(start)
	done.Wait()
	close(results)

	admissions := 0
	for result := range results {
		require.NoError(t, result.err)
		if result.admitted {
			admissions++
		}
	}
	require.Equal(t, 1, admissions)
	require.Equal(t, callers, clock.Calls())
}

func TestPhase151ProcessLocalPendingRecoveryIntervalGateIsNotGlobalAcrossInstances(t *testing.T) {
	base := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	clock := &phase151PendingRecoveryIntervalGateClock{now: base}
	firstGate := messenger.NewPhase151ProcessLocalPendingRecoveryIntervalGate(clock)
	secondGate := messenger.NewPhase151ProcessLocalPendingRecoveryIntervalGate(clock)
	require.NotNil(t, firstGate)
	require.NotNil(t, secondGate)

	admitted, err := firstGate.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
	require.True(t, admitted)
	require.NoError(t, err)

	admitted, err = secondGate.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
	require.True(t, admitted)
	require.NoError(t, err)
}

func TestPhase151ProcessLocalPendingRecoveryIntervalGateConsumesIntervalOnAdmission(t *testing.T) {
	base := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	clock := &phase151PendingRecoveryIntervalGateClock{now: base}
	gate := messenger.NewPhase151ProcessLocalPendingRecoveryIntervalGate(clock)
	require.NotNil(t, gate)

	admitted, err := gate.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
	require.True(t, admitted)
	require.NoError(t, err)

	admitted, err = gate.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
	require.False(t, admitted)
	require.NoError(t, err)
}
