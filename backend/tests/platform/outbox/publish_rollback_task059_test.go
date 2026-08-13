//go:build hubtest

package outbox_test

// TASK-059 / P3-BE-03 / W4-A2-007: PublishInTx + Rollback → 0 outbox rows.
// Mutation: Publish вне tx / на client вместо tx → Count>0 after Rollback → FAIL.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/outbox"
)

func TestPublisher_PublishInTx_RollbackLeavesNoEvents(t *testing.T) {
	client := newTestEntClient(t)
	ctx := context.Background()
	p := outbox.NewPublisher()

	tx, err := client.Tx(ctx)
	require.NoError(t, err)

	require.NoError(t, p.PublishInTx(ctx, tx, "project.rolled_back", map[string]string{"k": "v"}))
	require.NoError(t, tx.Rollback())

	cnt, err := client.OutboxEvent.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, cnt, "outbox row must not survive Rollback")
}
