//go:build hubtest

package outbox

import (
	"context"

	"github.com/valerii/hub/backend/ent"
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
