package outbox

import (
	"context"
	"fmt"
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
	pollInterval       = 100 * time.Millisecond
	// claimLeaseTTL — сколько published_at на pending держит claim без schema/enum
	// processing (A3W4-2-002). Crash mid-publish: at-least-once retry после TTL.
	claimLeaseTTL = 30 * time.Second
)

// betweenOutboxEventsHook — только hubtest: между событиями батча (cancel mid-batch).
var betweenOutboxEventsHook func()

// afterClaimCommitHook — только hubtest: сразу после Commit короткого claim,
// до первого PublishMsg (A2W4-2-001 / TASK-087).
var afterClaimCommitHook func()

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
	ticker := time.NewTicker(pollInterval)
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
// A2W4-2-001 / TASK-087: claim Commit ДО NATS PublishMsg/Flush — row-lock не
// держится на время flush (до 50×5s). Publish и mark published — вне claim-tx.
// A3W4-2-002: после SELECT — atomic stamp published_at (lease) WHERE ещё
// claim-available; второй worker не видит те же строки как свободный pending.
// HI-09: одно сбойное событие не должно валить весь loop.
func (w *Worker) processBatch(ctx context.Context) {
	w.markExhaustedFailed(ctx)

	events, err := w.claimAndRelease(ctx)
	if err != nil {
		slog.Error("outbox: failed to claim events", "error", err)
		return
	}
	if afterClaimCommitHook != nil {
		afterClaimCommitHook()
	}

	for _, evt := range events {
		// W4-A2-009 / P3-BE-12: cancel mid-batch — не дожимаем до 50 при shutdown.
		select {
		case <-ctx.Done():
			return
		default:
		}
		w.processEvent(ctx, evt)
		if betweenOutboxEventsHook != nil {
			betweenOutboxEventsHook()
		}
	}
}

// claimAndRelease выбирает pending под SKIP LOCKED, ставит claim-lease
// (published_at) и сразу Commit'ит claim-tx.
// Дальнейший NATS publish идёт уже без открытой транзакции/row-lock
// (A2W4-2-001: publish after commit claim).
//
// WHY lease без enum processing: schema сегодня pending|published|failed;
// migrate/enum — только с owner approve. published_at на pending = lease
// start; reclaim после claimLeaseTTL (crash mid-publish → at-least-once OK).
func (w *Worker) claimAndRelease(ctx context.Context) ([]*ent.OutboxEvent, error) {
	tx, err := w.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().UTC()
	events, err := w.claimPending(ctx, tx, now)
	if err != nil {
		return nil, err
	}

	claimed := make([]*ent.OutboxEvent, 0, len(events))
	for _, evt := range events {
		ok, stampErr := w.stampClaimLease(ctx, tx, evt, now)
		if stampErr != nil {
			return nil, stampErr
		}
		if ok {
			claimed = append(claimed, evt)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	return claimed, nil
}

// stampClaimLease атомарно помечает строку claim-lease через published_at.
// Where claim-available: второй concurrent worker получает NotFound → skip.
func (w *Worker) stampClaimLease(ctx context.Context, tx *ent.Tx, evt *ent.OutboxEvent, now time.Time) (bool, error) {
	_, err := tx.OutboxEvent.UpdateOneID(evt.ID).
		Where(claimAvailablePreds(now)...).
		SetPublishedAt(now).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("stamp claim lease: %w", err)
	}
	evt.PublishedAt = &now
	return true, nil
}

// claimAvailablePreds — pending + attempts + (нет lease / lease протух).
func claimAvailablePreds(now time.Time) []predicate.OutboxEvent {
	return []predicate.OutboxEvent{
		outboxevent.StatusEQ(outboxevent.StatusPending),
		outboxevent.AttemptsLT(maxPublishAttempts),
		outboxevent.Or(
			outboxevent.PublishedAtIsNil(),
			outboxevent.PublishedAtLT(now.Add(-claimLeaseTTL)),
		),
	}
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

// claimPending выбирает claim-available pending под row-lock (SKIP LOCKED на Postgres).
func (w *Worker) claimPending(ctx context.Context, tx *ent.Tx, now time.Time) ([]*ent.OutboxEvent, error) {
	base := func(extra ...predicate.OutboxEvent) *ent.OutboxEventQuery {
		preds := claimAvailablePreds(now)
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
			// A3W4-2-001: panic = publish-failure → attempts↑ / dead-letter at max.
			w.recordPublishFailure(ctx, evt, fmt.Errorf("panic: %v", r))
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

	// A2W4-2-002 / P1-BE-24: JetStream PubAck гарантирует запись в stream EVENTS
	// (Flush буфера core NATS этого не даёт при InterestPolicy). status=published
	// только после успешного PubAck — иначе crash/разрыв оставляет «published»
	// в БД без durable сообщения.
	if err := w.nats.PublishMsgAck(ctx, msg); err != nil {
		slog.Warn("outbox: publish failed", "event_type", evt.EventType, "error", err)
		w.recordPublishFailure(ctx, evt, err)
		return
	}

	now := time.Now().UTC()
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
// ClearPublishedAt снимает claim-lease, чтобы retry/другой worker могли взять
// строку сразу (не ждать claimLeaseTTL).
func (w *Worker) recordPublishFailure(ctx context.Context, evt *ent.OutboxEvent, err error) {
	attempts := evt.Attempts + 1
	upd := w.client.OutboxEvent.UpdateOneID(evt.ID).
		SetAttempts(attempts).
		SetLastError(err.Error()).
		ClearPublishedAt()
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
	// Синхронизируем in-memory для повторных panic/fail в том же батче.
	evt.Attempts = attempts
}
