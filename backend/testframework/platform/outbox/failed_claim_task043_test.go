//go:build hubtest

package outbox_test

import (
	"context"
	"errors"
	"testing"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/outboxevent"
	"github.com/valerii/hub/backend/internal/platform/outbox"
)

// TestSkipLockedPredicate_AddsFORUpdateSkipLocked — мутация W4-A2-003:
// claim predicate обязан суффиксировать SELECT фразой SKIP LOCKED.
func TestSkipLockedPredicate_AddsFORUpdateSkipLocked(t *testing.T) {
	s := sql.Dialect(dialect.Postgres).
		Select("*").
		From(sql.Table(outboxevent.Table))
	outbox.ExportApplySkipLocked(s)
	query, _ := s.Query()
	assert.Contains(t, query, "FOR UPDATE")
	assert.Contains(t, query, "SKIP LOCKED")
}

// TestRecordPublishFailure_AttemptsGTE10_SetsFailed — мутация W4-A2-002:
// attempts≥10 → status=failed (dead-letter), не вечный pending.
func TestRecordPublishFailure_AttemptsGTE10_SetsFailed(t *testing.T) {
	client := newTestEntClient(t)
	nats := startTestNATSServer(t)
	nats.Close()

	w := outbox.NewWorker(client, nats)
	ctx := context.Background()
	evt := insertOutboxEvent(t, client, "dead.letter", []byte(`{"x":1}`))

	// Доводим attempts до max-1, затем ещё один fail → failed.
	max := outbox.ExportMaxPublishAttempts()
	_, err := client.OutboxEvent.UpdateOneID(evt.ID).
		SetAttempts(max - 1).
		Save(ctx)
	require.NoError(t, err)
	evt.Attempts = max - 1

	outbox.ExportWorkerRecordPublishFailure(w, ctx, evt, errors.New("nats down"))

	reloaded, err := client.OutboxEvent.Get(ctx, evt.ID)
	require.NoError(t, err)
	assert.Equal(t, max, reloaded.Attempts)
	assert.Equal(t, outboxevent.StatusFailed, reloaded.Status)
	assert.NotEmpty(t, reloaded.LastError)
}

// TestProcessBatch_MarksExhaustedPendingAsFailed — sweep W4-A2-002 для
// уже застрявших pending с attempts≥10.
func TestProcessBatch_MarksExhaustedPendingAsFailed(t *testing.T) {
	client := newTestEntClient(t)
	nats := startTestNATSServer(t)
	w := outbox.NewWorker(client, nats)
	ctx := context.Background()

	evt := insertOutboxEvent(t, client, "stuck.event", []byte(`{}`))
	_, err := client.OutboxEvent.UpdateOneID(evt.ID).
		SetAttempts(outbox.ExportMaxPublishAttempts()).
		Save(ctx)
	require.NoError(t, err)

	require.NotPanics(t, func() {
		outbox.ExportWorkerProcessBatch(w, ctx)
	})

	reloaded, err := client.OutboxEvent.Get(ctx, evt.ID)
	require.NoError(t, err)
	assert.Equal(t, outboxevent.StatusFailed, reloaded.Status)
}

// TestProcessEvent_PublishFail_BelowMax_StaysPending — до порога статус pending.
func TestProcessEvent_PublishFail_BelowMax_StaysPending(t *testing.T) {
	client := newTestEntClient(t)
	nats := startTestNATSServer(t)
	nats.Close()

	w := outbox.NewWorker(client, nats)
	ctx := context.Background()
	evt := insertOutboxEvent(t, client, "retry.event", []byte(`{"x":1}`))

	outbox.ExportWorkerProcessEvent(w, ctx, evt)

	reloaded, err := client.OutboxEvent.Get(ctx, evt.ID)
	require.NoError(t, err)
	assert.Equal(t, outboxevent.StatusPending, reloaded.Status)
	assert.Equal(t, 1, reloaded.Attempts)
}
