//go:build hubtest

package outbox_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/outboxevent"
	"github.com/valerii/hub/backend/internal/platform/outbox"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// TestProcessBatch_ClaimReleasedBeforePublish_TASK087 — A2W4-2-001:
// claim-tx Commit'ится ДО PublishMsg/Flush.
//
// Мутация hold-during-flush (publish внутри claim-tx до Commit):
// afterClaimCommitHook сработает только после Flush → claimCommittedDuringPublish=false → красный.
func TestProcessBatch_ClaimReleasedBeforePublish_TASK087(t *testing.T) {
	client := newTestEntClient(t)
	nc := startTestNATSServer(t)
	w := outbox.NewWorker(client, nc)
	ctx := context.Background()

	insertOutboxEvent(t, client, "project.outbox.claim.lease", []byte(`{"ok":true}`))

	var claimCommitted atomic.Bool
	var publishSawClaimCommitted atomic.Bool

	outbox.ExportSetAfterClaimCommitHook(func() {
		claimCommitted.Store(true)
	})
	t.Cleanup(func() { outbox.ExportSetAfterClaimCommitHook(nil) })

	natsclient.ExportSetAfterCorePublishHook(func() {
		publishSawClaimCommitted.Store(claimCommitted.Load())
	})
	t.Cleanup(func() { natsclient.ExportSetAfterCorePublishHook(nil) })

	outbox.ExportWorkerProcessBatch(w, ctx)

	assert.True(t, claimCommitted.Load(),
		"claim must Commit before publish loop")
	assert.True(t, publishSawClaimCommitted.Load(),
		"PublishMsg/Flush must observe claim already committed (not hold-during-flush)")

	pending, err := client.OutboxEvent.Query().
		Where(outboxevent.StatusEQ(outboxevent.StatusPending)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, pending)
}

// TestProcessBatch_ConcurrentWorkerDuringFlush_TASK087 — пока первый worker
// в Flush, второй processBatch на новом pending не блокируется открытым claim-tx.
//
// Мутация hold-during-flush: SQLite write-lock открытой claim-tx блокирует
// concurrent batch → timeout → тест краснеет.
func TestProcessBatch_ConcurrentWorkerDuringFlush_TASK087(t *testing.T) {
	client := newTestEntClient(t)
	nc := startTestNATSServer(t)
	w := outbox.NewWorker(client, nc)
	ctx := context.Background()

	insertOutboxEvent(t, client, "project.outbox.claim.a", []byte(`{"i":1}`))

	var concurrentOK atomic.Bool
	var hookFired atomic.Bool

	natsclient.ExportSetAfterCorePublishHook(func() {
		if !hookFired.CompareAndSwap(false, true) {
			return // только первый Flush батча A
		}
		evtB := insertOutboxEvent(t, client, "project.outbox.claim.b", []byte(`{"i":2}`))

		done := make(chan struct{})
		go func() {
			defer close(done)
			outbox.ExportWorkerProcessBatch(w, ctx)
		}()

		select {
		case <-done:
			reloaded, err := client.OutboxEvent.Get(ctx, evtB.ID)
			if err == nil && reloaded.Status == outboxevent.StatusPublished {
				concurrentOK.Store(true)
			}
		case <-time.After(3 * time.Second):
			// claim-tx всё ещё держит lock — concurrent worker заблокирован
		}
	})
	t.Cleanup(func() { natsclient.ExportSetAfterCorePublishHook(nil) })

	outbox.ExportWorkerProcessBatch(w, ctx)

	require.True(t, hookFired.Load(), "expected Flush hook on first publish")
	assert.True(t, concurrentOK.Load(),
		"concurrent processBatch must finish during Flush (claim already released)")
}
