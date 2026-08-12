//go:build hubtest

package outbox_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/outboxevent"
	"github.com/valerii/hub/backend/internal/platform/outbox"
)

// TestClaimAndRelease_DualWorker_DisjointIDs — A3W4-2-002:
// после claim первого worker'а второй не видит те же строки как свободный pending
// (lease via published_at). На SQLite concurrent tx → «table is locked», поэтому
// проверяем post-commit семантику двух worker'ов последовательно + lost-race stamp.
//
// Мутация «Commit claim без stamp published_at»: второй claim забирает те же ID.
func TestClaimAndRelease_DualWorker_DisjointIDs(t *testing.T) {
	client := newTestEntClient(t)
	nc := startTestNATSServer(t)
	w1 := outbox.NewWorker(client, nc)
	w2 := outbox.NewWorker(client, nc)
	ctx := context.Background()

	const n = 12
	for i := 0; i < n; i++ {
		insertOutboxEvent(t, client, "project.outbox.dual.claim", []byte(`{"i":1}`))
	}

	first, err := outbox.ExportWorkerClaimAndRelease(w1, ctx)
	require.NoError(t, err)
	require.Len(t, first, n)

	second, err := outbox.ExportWorkerClaimAndRelease(w2, ctx)
	require.NoError(t, err)
	assert.Empty(t, second,
		"second worker must not re-claim leased pending (published_at lease)")

	seen := make(map[string]struct{}, n)
	for _, e := range first {
		seen[e.ID.String()] = struct{}{}
		assert.NotNil(t, e.PublishedAt, "claimed row must carry lease published_at")
	}
	assert.Len(t, seen, n)
}

// TestStampClaimLease_LostRace_Skips — A3W4-2-002 mutation:
// повторный stamp той же строки → ok=false (Where claim-available не проходит).
func TestStampClaimLease_LostRace_Skips(t *testing.T) {
	client := newTestEntClient(t)
	w := outbox.NewWorker(client, startTestNATSServer(t))
	ctx := context.Background()

	evt := insertOutboxEvent(t, client, "project.outbox.stamp.race", []byte(`{}`))
	now := time.Now().UTC()

	tx1, err := client.Tx(ctx)
	require.NoError(t, err)
	ok1, err := outbox.ExportStampClaimLease(w, ctx, tx1, evt, now)
	require.NoError(t, err)
	require.True(t, ok1)
	require.NoError(t, tx1.Commit())

	tx2, err := client.Tx(ctx)
	require.NoError(t, err)
	ok2, err := outbox.ExportStampClaimLease(w, ctx, tx2, evt, now)
	require.NoError(t, err)
	assert.False(t, ok2, "second stamp must lose race when lease already set")
	require.NoError(t, tx2.Commit())

	reloaded, err := client.OutboxEvent.Get(ctx, evt.ID)
	require.NoError(t, err)
	assert.Equal(t, outboxevent.StatusPending, reloaded.Status)
	assert.NotNil(t, reloaded.PublishedAt)
}
