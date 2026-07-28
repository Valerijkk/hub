package outbox

import (
	"context"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/nats-io/nats.go"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/outboxevent"
	"github.com/valerii/hub/backend/ent/predicate"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

const (
	// maxPublishAttempts — после этого события помечаются status=failed (dead-letter).
	maxPublishAttempts = 10
	batchLimit         = 50
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
// W4-A2-002: exhausted pending (attempts≥10) → status=failed.
// W4-A2-003: claim через Tx + FOR UPDATE SKIP LOCKED (Postgres); SQLite —
// fallback без lock (Ent/SQLite отвергает SELECT FOR UPDATE, см. tracker StartTimer).
// HI-09: одно сбойное событие не должно валить весь loop.
func (w *Worker) processBatch(ctx context.Context) {
	w.markExhaustedFailed(ctx)

	tx, err := w.client.Tx(ctx)
	if err != nil {
		slog.Error("outbox: failed to begin claim tx", "error", err)
		return
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	events, err := w.claimPending(ctx, tx)
	if err != nil {
		slog.Error("outbox: failed to claim events", "error", err)
		return
	}

	tw := &Worker{client: tx.Client(), nats: w.nats}
	for _, evt := range events {
		tw.processEvent(ctx, evt)
	}

	if err := tx.Commit(); err != nil {
		slog.Error("outbox: failed to commit claim tx", "error", err)
		return
	}
	committed = true
}

// markExhaustedFailed переводит застрявшие pending с attempts≥10 в failed
// (исторический dead-letter до фикса W4-A2-002).
func (w *Worker) markExhaustedFailed(ctx context.Context) {
	n, err := w.client.OutboxEvent.Update().
		Where(
			outboxevent.StatusEQ(outboxevent.StatusPending),
			outboxevent.AttemptsGTE(maxPublishAttempts),
		).
		SetStatus(outboxevent.StatusFailed).
		Save(ctx)
	if err != nil {
		slog.Error("outbox: failed to mark exhausted events", "error", err)
		return
	}
	if n > 0 {
		slog.Warn("outbox: marked exhausted pending as failed", "count", n)
	}
}

// claimPending выбирает pending-события под row-lock (SKIP LOCKED на Postgres).
func (w *Worker) claimPending(ctx context.Context, tx *ent.Tx) ([]*ent.OutboxEvent, error) {
	base := func(extra ...predicate.OutboxEvent) *ent.OutboxEventQuery {
		preds := []predicate.OutboxEvent{
			outboxevent.StatusEQ(outboxevent.StatusPending),
			outboxevent.AttemptsLT(maxPublishAttempts),
		}
		preds = append(preds, extra...)
		return tx.OutboxEvent.Query().
			Where(preds...).
			Order(ent.Asc(outboxevent.FieldCreatedAt)).
			Limit(batchLimit)
	}

	events, err := base(skipLockedPredicate()).All(ctx)
	if err != nil {
		// SQLite и драйверы без SKIP LOCKED — обычная выборка в той же tx.
		if isUnsupportedLockError(err) {
			return base().All(ctx)
		}
		return nil, err
	}
	return events, nil
}

// skipLockedPredicate добавляет FOR UPDATE SKIP LOCKED к SELECT (W4-A2-003).
func skipLockedPredicate() predicate.OutboxEvent {
	return predicate.OutboxEvent(func(s *sql.Selector) {
		s.ForUpdate(sql.WithLockAction(sql.SkipLocked))
	})
}

func isUnsupportedLockError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "syntax error") ||
		strings.Contains(msg, "for update") ||
		strings.Contains(msg, "skip locked") ||
		strings.Contains(msg, dialect.SQLite)
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

	// Восстанавливаем trace-context исходного запроса из outbox-строки и
	// переносим его в NATS headers: consumer продолжит тот же трейс, а не
	// начнёт новый (async-граница REST -> outbox -> NATS).
	msgCtx := natsclient.ExtractTraceContextFromMap(ctx, evt.TraceContext)
	msg := nats.NewMsg(evt.EventType)
	msg.Data = evt.Payload
	natsclient.InjectTraceContext(msgCtx, msg.Header)
	msg.Header.Set(natsclient.HeaderEventID, evt.ID.String())

	// P1-BE-24: PublishMsg ждёт FlushTimeout (серверный ack буфера) до возврата.
	// status=published выставляем только после успешного Confirm — иначе crash/
	// разрыв между Publish и Flush оставляет «published» в БД без события в NATS.
	if err := w.nats.PublishMsg(msg); err != nil {
		slog.Warn("outbox: publish failed", "event_type", evt.EventType, "error", err)
		w.recordPublishFailure(ctx, evt, err)
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

// recordPublishFailure инкрементирует attempts и сохраняет last_error.
// W4-A2-002: при attempts≥10 статус → failed (dead-letter), иначе pending.
func (w *Worker) recordPublishFailure(ctx context.Context, evt *ent.OutboxEvent, err error) {
	attempts := evt.Attempts + 1
	upd := w.client.OutboxEvent.UpdateOneID(evt.ID).
		SetAttempts(attempts).
		SetLastError(err.Error())
	if attempts >= maxPublishAttempts {
		upd = upd.SetStatus(outboxevent.StatusFailed)
		slog.Error("outbox: event dead-lettered after max attempts",
			"event_id", evt.ID,
			"event_type", evt.EventType,
			"attempts", attempts,
		)
	}
	if _, saveErr := upd.Save(ctx); saveErr != nil {
		slog.Error("outbox: failed to record publish failure",
			"event_id", evt.ID,
			"event_type", evt.EventType,
			"error", saveErr,
		)
	}
}
