package outbox

import (
	"context"
	"log/slog"
	"runtime/debug"
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

// processBatch обрабатывает пачку ожидающих событий (до 50 за раз).
// HI-09: одно сбойное событие не должно валить весь loop — обработка каждой
// записи обёрнута в локальный recover, а update'ы статуса больше не используют
// SaveX (паника при transient DB-ошибке). Сбойные апдейты логируются и
// событие попадает в следующую итерацию (Attempts инкрементируются при
// очередном publish-fail; для published-успехов запись остаётся в pending
// до следующего успешного UpdateOne — это безопасно, мы скорее доставим
// событие дважды, чем потеряем его).
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
		w.processEvent(ctx, evt)
	}
}

// processEvent обрабатывает одно outbox-событие с локальным recover'ом,
// чтобы паника на одной записи не уносила за собой обработку остальных.
func (w *Worker) processEvent(ctx context.Context, evt *ent.OutboxEvent) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("outbox: panic while processing event",
				"event_id", evt.ID,
				"event_type", evt.EventType,
				"panic", r,
				"stack", string(debug.Stack()),
			)
		}
	}()

	if err := w.nats.Publish(evt.EventType, evt.Payload); err != nil {
		slog.Warn("outbox: publish failed", "event_type", evt.EventType, "error", err)
		if _, saveErr := w.client.OutboxEvent.UpdateOneID(evt.ID).
			SetAttempts(evt.Attempts + 1).
			SetLastError(err.Error()).
			Save(ctx); saveErr != nil {
			slog.Error("outbox: failed to record publish failure",
				"event_id", evt.ID,
				"event_type", evt.EventType,
				"error", saveErr,
			)
		}
		return
	}

	now := time.Now()
	if _, saveErr := w.client.OutboxEvent.UpdateOneID(evt.ID).
		SetStatus(outboxevent.StatusPublished).
		SetPublishedAt(now).
		Save(ctx); saveErr != nil {
		slog.Error("outbox: failed to mark event published",
			"event_id", evt.ID,
			"event_type", evt.EventType,
			"error", saveErr,
		)
	}
}
