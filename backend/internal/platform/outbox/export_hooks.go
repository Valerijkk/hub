//go:build hubtest

package outbox

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/predicate"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// ExportWorkerClient returns Worker.client for hubtest assertions.
func ExportWorkerClient(w *Worker) *ent.Client {
	return w.client
}

// ExportWorkerNats returns Worker.nats for hubtest assertions.
func ExportWorkerNats(w *Worker) *natsclient.Client {
	return w.nats
}

// ExportWorkerProcessEvent wraps Worker.processEvent for hubtest.
func ExportWorkerProcessEvent(w *Worker, ctx context.Context, evt *ent.OutboxEvent) {
	w.processEvent(ctx, evt)
}

// ExportWorkerProcessBatch wraps Worker.processBatch for hubtest.
func ExportWorkerProcessBatch(w *Worker, ctx context.Context) {
	w.processBatch(ctx)
}

// ExportWorkerClaimAndRelease wraps claimAndRelease for dual-worker hubtest (A3W4-2-002).
func ExportWorkerClaimAndRelease(w *Worker, ctx context.Context) ([]*ent.OutboxEvent, error) {
	return w.claimAndRelease(ctx)
}

// ExportStampClaimLease wraps stampClaimLease for hubtest (lost-race / lease).
func ExportStampClaimLease(w *Worker, ctx context.Context, tx *ent.Tx, evt *ent.OutboxEvent, now time.Time) (bool, error) {
	return w.stampClaimLease(ctx, tx, evt, now)
}

// ExportClaimLeaseTTL — A3W4-2-002: TTL reclaim orphan lease (published_at на pending).
func ExportClaimLeaseTTL() time.Duration {
	return claimLeaseTTL
}

// ExportWorkerRecordPublishFailure wraps recordPublishFailure for hubtest.
func ExportWorkerRecordPublishFailure(w *Worker, ctx context.Context, evt *ent.OutboxEvent, err error) {
	w.recordPublishFailure(ctx, evt, err)
}

// ExportSkipLockedPredicate returns the SKIP LOCKED claim predicate for hubtest.
func ExportSkipLockedPredicate() predicate.OutboxEvent {
	return skipLockedPredicate()
}

// ExportApplySkipLocked applies SKIP LOCKED to a selector (mutation-proof assert).
func ExportApplySkipLocked(s *sql.Selector) {
	skipLockedPredicate()(s)
}

// ExportMaxPublishAttempts returns the dead-letter threshold.
func ExportMaxPublishAttempts() int {
	return maxPublishAttempts
}

// ExportSetBetweenOutboxEventsHook — hubtest: между событиями processBatch (cancel mid-batch).
func ExportSetBetweenOutboxEventsHook(fn func()) {
	betweenOutboxEventsHook = fn
}

// ExportSetAfterClaimCommitHook — hubtest: после Commit короткого claim, до Publish
// (A2W4-2-001 / TASK-087). Мутация hold-during-flush → hook не срабатывает до Flush.
func ExportSetAfterClaimCommitHook(fn func()) {
	afterClaimCommitHook = fn
}

// ExportBatchLimit — W4-A2-008 / P3-BE-02: размер claim-batch outbox worker.
func ExportBatchLimit() int {
	return batchLimit
}

// ExportPollInterval — W4-A2-008 / P3-BE-02: интервал ProcessLoop.
func ExportPollInterval() time.Duration {
	return pollInterval
}
