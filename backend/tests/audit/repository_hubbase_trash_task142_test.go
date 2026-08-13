//go:build hubtest

package audit_test

// A3W2-A5-001: hubbase в единой админ-корзине (list / restore / purge).
import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/hubbase"
	"github.com/valerii/hub/backend/ent/hubcolumn"
	"github.com/valerii/hub/backend/ent/hubrecord"
	"github.com/valerii/hub/backend/ent/hubtable"
	"github.com/valerii/hub/backend/internal/audit"
	"github.com/valerii/hub/backend/internal/bases"
)

// TestRepository_SoftDeleteAndRestoreHubBase_Task142 — A3W2-A5-001.
// Мутация: убрать case "hubbase" / RestoreHubBase → list/restore краснеют.
func TestRepository_SoftDeleteAndRestoreHubBase_Task142(t *testing.T) {
	env := newAuditTrashEnv(t)
	repo := bases.NewEntRepository(env.client)
	svc := bases.NewService(repo)

	detail, err := svc.CreateBase(env.ctx, env.userA, bases.CreateBaseRequest{Name: "Trash Base"})
	require.NoError(t, err)
	table, err := svc.CreateTable(env.ctx, env.userA, detail.ID, bases.CreateTableRequest{Name: "Extra"})
	require.NoError(t, err)
	_, err = svc.CreateRecord(env.ctx, detail.ID, table.ID, env.userA, bases.CreateRecordRequest{
		Cells: map[string]interface{}{},
	})
	require.NoError(t, err)

	when := time.Now().UTC()
	require.NoError(t, env.repo.SoftDeleteHubBase(env.ctx, detail.ID, when))

	items, total, err := env.repo.ListDeletedHubBases(env.ctx, when.Add(-time.Hour), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, detail.ID, items[0].ID)
	assert.Equal(t, "hubbase", items[0].EntityType)

	require.NoError(t, env.repo.RestoreHubBase(env.ctx, detail.ID))

	items, total, err = env.repo.ListDeletedHubBases(env.ctx, when.Add(-time.Hour), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, total)

	restored, err := env.client.HubBase.Query().
		Where(hubbase.IDEQ(detail.ID), hubbase.DeletedAtIsNil()).
		Only(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, "Trash Base", restored.Name)

	aliveTables, err := env.client.HubTable.Query().
		Where(hubtable.HasBaseWith(hubbase.ID(detail.ID)), hubtable.DeletedAtIsNil()).
		Count(env.ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, aliveTables, 1)
	aliveCols, err := env.client.HubColumn.Query().Where(hubcolumn.DeletedAtIsNil()).Count(env.ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, aliveCols, 1)
	aliveRecs, err := env.client.HubRecord.Query().Where(hubrecord.DeletedAtIsNil()).Count(env.ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, aliveRecs, 1)
}

// TestRepository_HubBase_RetentionHidesOlderThan30Days — A3W2-A5-001 / NFR-11.
func TestRepository_HubBase_RetentionHidesOlderThan30Days(t *testing.T) {
	env := newAuditTrashEnv(t)
	repo := bases.NewEntRepository(env.client)
	svc := bases.NewService(repo)
	detail, err := svc.CreateBase(env.ctx, env.userA, bases.CreateBaseRequest{Name: "Old Base"})
	require.NoError(t, err)

	old := time.Now().UTC().Add(-40 * 24 * time.Hour)
	cutoff := time.Now().UTC().Add(-30 * 24 * time.Hour)
	require.NoError(t, env.repo.SoftDeleteHubBase(env.ctx, detail.ID, old))

	items, total, err := env.repo.ListDeletedHubBases(env.ctx, cutoff, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, items)
}

// TestRepository_PurgeExpiredHubBases_Task142 — hard-delete дерева старше retention.
func TestRepository_PurgeExpiredHubBases_Task142(t *testing.T) {
	env := newAuditTrashEnv(t)
	repo := bases.NewEntRepository(env.client)
	svc := bases.NewService(repo)
	detail, err := svc.CreateBase(env.ctx, env.userA, bases.CreateBaseRequest{Name: "Purge Base"})
	require.NoError(t, err)
	table, err := svc.CreateTable(env.ctx, env.userA, detail.ID, bases.CreateTableRequest{Name: "T"})
	require.NoError(t, err)
	_, err = svc.CreateRecord(env.ctx, detail.ID, table.ID, env.userA, bases.CreateRecordRequest{
		Cells: map[string]interface{}{},
	})
	require.NoError(t, err)

	old := time.Now().UTC().Add(-40 * 24 * time.Hour)
	require.NoError(t, env.repo.SoftDeleteHubBase(env.ctx, detail.ID, old))

	n, err := env.repo.PurgeExpiredHubBases(env.ctx, time.Now().UTC().Add(-30*24*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	_, err = env.client.HubBase.Get(env.ctx, detail.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
	tablesLeft, err := env.client.HubTable.Query().
		Where(hubtable.HasBaseWith(hubbase.ID(detail.ID))).
		Count(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, tablesLeft)
}

// TestAuditService_RestoreHubBase_ViaTrashAPI_Task142 — ListTrash/RestoreItem type=hubbase.
// Мутация: убрать case "hubbase" из ListTrash/RestoreItem → total=0 / unknown type.
func TestAuditService_RestoreHubBase_ViaTrashAPI_Task142(t *testing.T) {
	env := newAuditTrashEnv(t)
	repo := bases.NewEntRepository(env.client)
	svcBases := bases.NewService(repo)
	detail, err := svcBases.CreateBase(env.ctx, env.userA, bases.CreateBaseRequest{Name: "Restore via service"})
	require.NoError(t, err)
	when := time.Now().UTC()
	require.NoError(t, env.repo.SoftDeleteHubBase(env.ctx, detail.ID, when))

	svc := audit.NewAuditService(env.repo, nil)

	list, err := svc.ListTrash(env.ctx, "hubbase", nil, 1, 50)
	require.NoError(t, err)
	require.Equal(t, 1, list.Total)
	require.Len(t, list.Items, 1)
	assert.Equal(t, detail.ID.String(), list.Items[0].ID)
	assert.Equal(t, "hubbase", list.Items[0].EntityType)
	assert.Greater(t, list.Items[0].DaysRemaining, 0)

	require.NoError(t, svc.RestoreItem(env.ctx, "hubbase", detail.ID))

	got, err := env.client.HubBase.Query().
		Where(hubbase.IDEQ(detail.ID), hubbase.DeletedAtIsNil()).
		Only(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, "Restore via service", got.Name)

	listAfter, err := svc.ListTrash(env.ctx, "hubbase", nil, 1, 50)
	require.NoError(t, err)
	assert.Equal(t, 0, listAfter.Total)
}
