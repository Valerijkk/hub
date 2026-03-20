package outbox

import (
	"context"
	"log/slog"
	"time"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/outboxevent"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// Worker polls the outbox table for pending events and publishes them to NATS.
// Events that fail after 10 attempts are marked as failed.
type Worker struct {
	client *ent.Client
	nats   *natsclient.Client
}

// NewWorker creates a new outbox worker.
func NewWorker(client *ent.Client, nats *natsclient.Client) *Worker {
	return &Worker{client: client, nats: nats}
}

// ProcessLoop polls the outbox table every 100ms and publishes pending events to NATS.
// It runs until the context is cancelled.
func (w *Worker) ProcessLoop(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) {
	events, err := w.client.OutboxEvent.Query().
		Where(
			outboxevent.StatusEQ(outboxevent.StatusPending),
			outboxevent.AttemptsLT(10),
		).
		Order(ent.Asc(outboxevent.FieldCreatedAt)).
		Limit(50).
		All(ctx)
	if err != nil {
		slog.Error("outbox: failed to query events", "error", err)
		return
	}

	for _, evt := range events {
		if err := w.nats.Publish(evt.EventType, evt.Payload); err != nil {
			slog.Warn("outbox: publish failed", "event_type", evt.EventType, "error", err)
			w.client.OutboxEvent.UpdateOneID(evt.ID).
				SetAttempts(evt.Attempts + 1).
				SetLastError(err.Error()).
				SaveX(ctx)
			continue
		}
		now := time.Now()
		w.client.OutboxEvent.UpdateOneID(evt.ID).
			SetStatus(outboxevent.StatusPublished).
			SetPublishedAt(now).
			SaveX(ctx)
	}
}
