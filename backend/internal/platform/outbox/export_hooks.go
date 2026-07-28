//go:build hubtest

package outbox

import (
	"context"

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
