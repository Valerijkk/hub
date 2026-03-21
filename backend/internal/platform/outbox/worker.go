package outbox

import (
	"context"
	"log/slog"
	"time"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/outboxevent"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// Worker опрашивает outbox-таблицу на наличие ожидающих событий и публикует их в NATS.
// События, не обработанные после 10 попыток, помечаются как failed.
type Worker struct {
	client *ent.Client
	nats   *natsclient.Client
}

// NewWorker создаёт новый outbox-воркер.
func NewWorker(client *ent.Client, nats *natsclient.Client) *Worker {
	return &Worker{client: client, nats: nats}
}

// ProcessLoop опрашивает outbox-таблицу каждые 100мс и публикует ожидающие события в NATS.
// Работает до отмены контекста.
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
