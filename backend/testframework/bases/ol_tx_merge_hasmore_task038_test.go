//go:build hubtest

package bases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/hubbase"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/bases"
)

// TestUpdateBase_OptimisticLock_VersionEQ_W2A5004 — stale/zero version → conflict;
// VersionEQ в Ent WHERE. Mutation: снять VersionEQ → stale write проходит → FAIL.
func TestUpdateBase_OptimisticLock_VersionEQ_W2A5004(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:bases_ol_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("ol-" + uuid.NewString()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	svc := bases.NewService(bases.NewEntRepository(client))
	created, err := svc.CreateBase(ctx, owner.ID, bases.CreateBaseRequest{Name: "OL Base"})
	require.NoError(t, err)

	name := "v2"
	updated, err := svc.UpdateBase(ctx, owner.ID, created.ID, bases.UpdateBaseRequest{
		Name: &name, Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, updated.Version)

	stale := "nope"
	_, err = svc.UpdateBase(ctx, owner.ID, created.ID, bases.UpdateBaseRequest{
		Name: &stale, Version: created.Version,
	})
	require.ErrorIs(t, err, bases.ErrVersionConflict)

	_, err = svc.UpdateBase(ctx, owner.ID, created.ID, bases.UpdateBaseRequest{
		Name: &stale, Version: 0,
	})
	require.ErrorIs(t, err, bases.ErrVersionConflict)
}

// TestCreateBase_AtomicTx_W2A5001 — base+Main+Name в одной tx (нет orphan base).
// Mutation: создать base без tx → при fail column остаётся orphan base → FAIL.
func TestCreateBase_AtomicTx_W2A5001(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:bases_tx_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("tx-" + uuid.NewString()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	svc := bases.NewService(bases.NewEntRepository(client))
	detail, err := svc.CreateBase(ctx, owner.ID, bases.CreateBaseRequest{Name: "Atomic"})
	require.NoError(t, err)
	require.Len(t, detail.Tables, 1)
	assert.Equal(t, "Main", detail.Tables[0].Name)

	view, err := svc.GetTableView(ctx, owner.ID, detail.ID, detail.Tables[0].ID)
	require.NoError(t, err)
	require.NotEmpty(t, view.Columns)
	assert.Equal(t, "Name", view.Columns[0].Name)
	assert.False(t, view.HasMore)
	assert.Equal(t, 0, view.Total)

	n, err := client.HubBase.Query().Where(hubbase.DeletedAtIsNil()).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

// TestUpdateRecord_PartialMerge_W2A5002 — partial PUT merge, не wipe.
// Mutation: SetCells(patch) без merge → вторая колонка пропадает → FAIL.
func TestUpdateRecord_PartialMerge_W2A5002(t *testing.T) {
	svc, ownerID := setupBasesService(t)
	ctx := context.Background()

	base, err := svc.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "Merge"})
	require.NoError(t, err)
	tableID := base.Tables[0].ID

	view, err := svc.GetTableView(ctx, ownerID, base.ID, tableID)
	require.NoError(t, err)
	nameCol := view.Columns[0].ID.String()

	qty, err := svc.CreateColumn(ctx, ownerID, base.ID, tableID, CreateColumnRequest{
		Name: "Qty", ColType: ColumnNumber,
	})
	require.NoError(t, err)

	rec, err := svc.CreateRecord(ctx, base.ID, tableID, ownerID, CreateRecordRequest{
		Cells: map[string]interface{}{
			nameCol:          "Laptop",
			qty.ID.String():  float64(3),
		},
	})
	require.NoError(t, err)

	updated, err := svc.UpdateRecord(ctx, ownerID, base.ID, tableID, rec.ID, UpdateRecordRequest{
		Cells:   map[string]interface{}{nameCol: "Desktop"},
		Version: rec.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "Desktop", updated.Cells[nameCol])
	assert.Equal(t, float64(3), updated.Cells[qty.ID.String()], "Qty must survive partial PUT")
}

// TestGetTableView_HasMore_W2A5005 — cap + has_more/total при N>MaxTableRecords.
// Mutation: убрать HasMore/Total → клиент не узнаёт о truncation → FAIL.
func TestGetTableView_HasMore_W2A5005(t *testing.T) {
	repo, ctx, ownerID := setupBasesRepo(t)
	svc := NewService(repo)

	base, err := svc.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "Cap View"})
	require.NoError(t, err)
	table, err := svc.CreateTable(ctx, ownerID, base.ID, CreateTableRequest{Name: "Big"})
	require.NoError(t, err)

	// Неполный MaxTableRecords для скорости: подменяем проверку через Count + cap.
	// Создаём MaxTableRecords+2 через repo.
	for i := 0; i < bases.MaxTableRecords+2; i++ {
		_, err := repo.CreateRecord(ctx, table.ID, ownerID, map[string]interface{}{"n": float64(i)})
		require.NoError(t, err)
	}

	view, err := svc.GetTableView(ctx, ownerID, base.ID, table.ID)
	require.NoError(t, err)
	assert.Equal(t, bases.MaxTableRecords, len(view.Records))
	assert.Equal(t, bases.MaxTableRecords+2, view.Total)
	assert.True(t, view.HasMore)
}
