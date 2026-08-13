package calls_test

// FR-CAL-05: soft-delete серии убирает её из активных выборок calls.
import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/scheduledcall"
)

func TestRepository_DeleteScheduledCall_SoftDeletesSeries(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	proj, err := f.client.Project.Create().
		SetName("Sched Trash").
		SetKey("SCT").
		Save(ctx)
	require.NoError(t, err)

	at := time.Now().UTC().Add(48 * time.Hour)
	sc, err := f.repo.CreateScheduledCall(ctx, "Weekly sync", "desc", at, "FREQ=WEEKLY;COUNT=4", f.channel.ID, proj.ID, f.owner.ID, 0, nil, nil)
	require.NoError(t, err)

	before, err := f.repo.ListScheduledCalls(ctx, f.channel.ID)
	require.NoError(t, err)
	require.Len(t, before, 1)

	require.NoError(t, f.repo.DeleteScheduledCall(ctx, sc.ID))

	after, err := f.repo.ListScheduledCalls(ctx, f.channel.ID)
	require.NoError(t, err)
	assert.Empty(t, after)

	row, err := f.client.ScheduledCall.Get(ctx, sc.ID)
	require.NoError(t, err)
	require.NotNil(t, row.DeletedAt)
	assert.False(t, row.IsActive)

	due, err := f.repo.GetDueScheduledCalls(ctx, time.Now().UTC().Add(72*time.Hour))
	require.NoError(t, err)
	for _, d := range due {
		assert.NotEqual(t, sc.ID, d.ID)
	}

	// Сырой запрос без фильтра видит soft-deleted — защита живёт в List/GetDue.
	raw, err := f.client.ScheduledCall.Query().Where(scheduledcall.IDEQ(sc.ID)).Only(ctx)
	require.NoError(t, err)
	assert.NotNil(t, raw.DeletedAt)
}
