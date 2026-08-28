package messenger

import (
	"context"
	"sync"
	"time"
)

type Phase151PendingRecoveryIntervalClock interface {
	Now() time.Time
}

type Phase151PendingRecoveryIntervalGate interface {
	TryAcquirePendingRecoveryInterval(context.Context, time.Duration) (bool, error)
}

// Phase151ProcessLocalPendingRecoveryIntervalGate admits at most one pending
// recovery attempt per interval within its owning process.
type Phase151ProcessLocalPendingRecoveryIntervalGate struct {
	mu             sync.Mutex
	clock          Phase151PendingRecoveryIntervalClock
	lastAdmittedAt time.Time
}

var _ Phase151PendingRecoveryIntervalGate = (*Phase151ProcessLocalPendingRecoveryIntervalGate)(nil)

func NewPhase151ProcessLocalPendingRecoveryIntervalGate(
	clock Phase151PendingRecoveryIntervalClock,
) *Phase151ProcessLocalPendingRecoveryIntervalGate {
	if phase151PostCreateDependencyNil(clock) {
		return nil
	}
	return &Phase151ProcessLocalPendingRecoveryIntervalGate{clock: clock}
}

func (g *Phase151ProcessLocalPendingRecoveryIntervalGate) TryAcquirePendingRecoveryInterval(
	ctx context.Context,
	minimum time.Duration,
) (bool, error) {
	if g == nil || phase151PostCreateDependencyNil(g.clock) || phase151PostCreateDependencyNil(ctx) {
		return false, ErrPostCreatePipelineInvalid
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if minimum <= 0 {
		return false, ErrPostCreatePipelineInvalid
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return false, err
	}
	now := g.clock.Now()
	if now.IsZero() {
		return false, ErrPostCreatePipelineInvalid
	}
	if g.lastAdmittedAt.IsZero() {
		g.lastAdmittedAt = now
		return true, nil
	}
	if now.Before(g.lastAdmittedAt) {
		return false, ErrPostCreatePipelineInvalid
	}
	if now.Sub(g.lastAdmittedAt) < minimum {
		return false, nil
	}

	g.lastAdmittedAt = now
	return true, nil
}
