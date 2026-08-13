package audit_test

// FR-CAL-05: scheduled_call в единой админ-корзине.
import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/scheduledcall"
	"github.com/valerii/hub/backend/internal/audit"
)

func (e auditTrashEnv) createScheduledSeries(t *testing.T, title, rrule string) *ent.ScheduledCall {
	t.Helper()
	ch, err := e.client.Channel.Create().
		SetChannelType(channel.ChannelTypePublic).
		SetName("sched-" + title).
		SetCreatedBy(e.userA).
		Save(e.ctx)
	require.NoError(t, err)

	create := e.client.ScheduledCall.Create().
		SetTitle(title).
		SetScheduledAt(time.Now().UTC().Add(24 * time.Hour)).
		SetChannelID(ch.ID).
		SetProjectID(e.projectID).
		SetCreatedBy(e.userA).
		SetIsActive(true)
	if rrule != "" {
		create = create.SetRrule(rrule)
	}
	sc, err := create.Save(e.ctx)
	require.NoError(t, err)
	return sc
}

func TestRepository_SoftDeleteAndRestoreScheduledCall(t *testing.T) {
	env := newAuditTrashEnv(t)
	sc := env.createScheduledSeries(t, "Daily standup", "FREQ=DAILY;COUNT=5")
	when := time.Now().UTC()

	require.NoError(t, env.repo.SoftDeleteScheduledCall(env.ctx, sc.ID, when))

	items, total, err := env.repo.ListDeletedScheduledCalls(env.ctx, &env.projectID, when.Add(-time.Hour), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, sc.ID, items[0].ID)
	assert.Equal(t, "scheduled_call", items[0].EntityType)

	require.NoError(t, env.repo.RestoreScheduledCall(env.ctx, sc.ID))

	items, total, err = env.repo.ListDeletedScheduledCalls(env.ctx, &env.projectID, when.Add(-time.Hour), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, total)

	restored, err := env.client.ScheduledCall.Get(env.ctx, sc.ID)
	require.NoError(t, err)
	assert.Nil(t, restored.DeletedAt)
	assert.True(t, restored.IsActive)
}

func TestRepository_ScheduledCall_RetentionHidesOlderThan30Days(t *testing.T) {
	env := newAuditTrashEnv(t)
	sc := env.createScheduledSeries(t, "Old series", "FREQ=WEEKLY;COUNT=2")
	old := time.Now().UTC().Add(-40 * 24 * time.Hour)
	cutoff := time.Now().UTC().Add(-30 * 24 * time.Hour)

	require.NoError(t, env.repo.SoftDeleteScheduledCall(env.ctx, sc.ID, old))

	items, total, err := env.repo.ListDeletedScheduledCalls(env.ctx, nil, cutoff, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, items)
}

func TestRepository_PurgeExpiredScheduledCalls(t *testing.T) {
	env := newAuditTrashEnv(t)
	sc := env.createScheduledSeries(t, "Purge series", "")
	old := time.Now().UTC().Add(-40 * 24 * time.Hour)

	require.NoError(t, env.repo.SoftDeleteScheduledCall(env.ctx, sc.ID, old))

	n, err := env.repo.PurgeExpiredScheduledCalls(env.ctx, time.Now().UTC().Add(-30*24*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	_, err = env.client.ScheduledCall.Get(env.ctx, sc.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}

// TestAuditService_RestoreScheduledCall_ViaTrashAPI — сервисный путь RestoreItem("scheduled_call").
func TestAuditService_RestoreScheduledCall_ViaTrashAPI(t *testing.T) {
	env := newAuditTrashEnv(t)
	sc := env.createScheduledSeries(t, "Restore series", "FREQ=DAILY;COUNT=3")
	when := time.Now().UTC()
	require.NoError(t, env.repo.SoftDeleteScheduledCall(env.ctx, sc.ID, when))

	svc := audit.NewAuditService(env.repo, nil)

	list, err := svc.ListTrash(env.ctx, "scheduled_call", &env.projectID, 1, 50)
	require.NoError(t, err)
	require.Equal(t, 1, list.Total)

	require.NoError(t, svc.RestoreItem(env.ctx, "scheduled_call", sc.ID))

	got, err := env.client.ScheduledCall.Query().
		Where(scheduledcall.IDEQ(sc.ID), scheduledcall.DeletedAtIsNil(), scheduledcall.IsActive(true)).
		Only(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, "Restore series", got.Title)
}
