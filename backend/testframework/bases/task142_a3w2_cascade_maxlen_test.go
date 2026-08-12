//go:build hubtest

package bases_test

// A3W2-A5-002 / A3W2-A5-003 — SoftDeleteTable cascade + MaxLen→400 (TASK-142).
import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/hubcolumn"
	"github.com/valerii/hub/backend/ent/hubrecord"
	"github.com/valerii/hub/backend/ent/hubtable"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/bases"
)

func openBasesTask142(t *testing.T) (*ent.Client, *bases.EntRepository, *bases.Service, context.Context, uuid.UUID) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:bases_task142_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	owner, err := client.User.Create().
		SetEmail("t142-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)
	repo := bases.NewEntRepository(client)
	return client, repo, bases.NewService(repo), ctx, owner.ID
}

// TestTask142_SoftDeleteTable_CascadesChildren — A3W2-A5-002.
// Мутация: SoftDeleteTable только table deleted_at → aliveCols/aliveRecs ≠ 0.
func TestTask142_SoftDeleteTable_CascadesChildren(t *testing.T) {
	client, repo, svc, ctx, ownerID := openBasesTask142(t)

	base, err := svc.CreateBase(ctx, ownerID, bases.CreateBaseRequest{Name: "CascadeBase"})
	require.NoError(t, err)
	table, err := svc.CreateTable(ctx, ownerID, base.ID, bases.CreateTableRequest{Name: "Items"})
	require.NoError(t, err)
	_, err = svc.CreateColumn(ctx, ownerID, base.ID, table.ID, bases.CreateColumnRequest{
		Name: "Extra", ColType: bases.ColumnText,
	})
	require.NoError(t, err)
	_, err = svc.CreateRecord(ctx, base.ID, table.ID, ownerID, bases.CreateRecordRequest{
		Cells: map[string]interface{}{},
	})
	require.NoError(t, err)

	require.NoError(t, repo.SoftDeleteTable(ctx, table.ID))

	aliveCols, err := client.HubColumn.Query().
		Where(
			hubcolumn.DeletedAtIsNil(),
			hubcolumn.HasTableWith(hubtable.ID(table.ID)),
		).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, aliveCols, "SoftDeleteTable must cascade soft-delete columns")

	aliveRecs, err := client.HubRecord.Query().
		Where(
			hubrecord.DeletedAtIsNil(),
			hubrecord.HasTableWith(hubtable.ID(table.ID)),
		).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, aliveRecs, "SoftDeleteTable must cascade soft-delete records")
}

// TestTask142_SoftDeleteTableThenBase_NoOrphans — A3W2-A5-002.
// SoftDeleteTable first, затем SoftDeleteBase — дети уже deleted table не orphan.
func TestTask142_SoftDeleteTableThenBase_NoOrphans(t *testing.T) {
	client, repo, svc, ctx, ownerID := openBasesTask142(t)

	base, err := svc.CreateBase(ctx, ownerID, bases.CreateBaseRequest{Name: "OrphanBase"})
	require.NoError(t, err)
	table, err := svc.CreateTable(ctx, ownerID, base.ID, bases.CreateTableRequest{Name: "T"})
	require.NoError(t, err)
	_, err = svc.CreateColumn(ctx, ownerID, base.ID, table.ID, bases.CreateColumnRequest{
		Name: "X", ColType: bases.ColumnText,
	})
	require.NoError(t, err)
	_, err = svc.CreateRecord(ctx, base.ID, table.ID, ownerID, bases.CreateRecordRequest{
		Cells: map[string]interface{}{},
	})
	require.NoError(t, err)

	require.NoError(t, repo.SoftDeleteTable(ctx, table.ID))
	require.NoError(t, svc.DeleteBase(ctx, ownerID, base.ID))

	aliveTables, err := client.HubTable.Query().Where(hubtable.DeletedAtIsNil()).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, aliveTables)
	aliveCols, err := client.HubColumn.Query().Where(hubcolumn.DeletedAtIsNil()).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, aliveCols)
	aliveRecs, err := client.HubRecord.Query().Where(hubrecord.DeletedAtIsNil()).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, aliveRecs)
}

// TestTask142_NameMaxLen_InvalidInput400 — A3W2-A5-003.
// Мутация: убрать validateNameLen → CreateBase с 201 руной → не ErrInvalidInput.
func TestTask142_NameMaxLen_InvalidInput400(t *testing.T) {
	svc, ownerID := setupBasesService(t)
	ctx := context.Background()
	long := strings.Repeat("я", bases.MaxNameRunes+1)

	_, err := svc.CreateBase(ctx, ownerID, bases.CreateBaseRequest{Name: long})
	require.ErrorIs(t, err, bases.ErrInvalidInput)
	assert.Equal(t, 400, bases.ExportFiberStatusFromError(bases.ExportMapServiceError(err)))

	ok, err := svc.CreateBase(ctx, ownerID, bases.CreateBaseRequest{Name: "OK"})
	require.NoError(t, err)

	_, err = svc.CreateTable(ctx, ownerID, ok.ID, bases.CreateTableRequest{Name: long})
	require.ErrorIs(t, err, bases.ErrInvalidInput)

	table, err := svc.CreateTable(ctx, ownerID, ok.ID, bases.CreateTableRequest{Name: "T"})
	require.NoError(t, err)

	_, err = svc.CreateColumn(ctx, ownerID, ok.ID, table.ID, bases.CreateColumnRequest{
		Name: long, ColType: bases.ColumnText,
	})
	require.ErrorIs(t, err, bases.ErrInvalidInput)

	tooLong := long
	_, err = svc.UpdateBase(ctx, ownerID, ok.ID, bases.UpdateBaseRequest{
		Name: &tooLong, Version: ok.Version,
	})
	require.ErrorIs(t, err, bases.ErrInvalidInput)
}
