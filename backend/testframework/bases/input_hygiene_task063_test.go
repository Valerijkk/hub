//go:build hubtest

package bases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/bases"
)

// Mutation: убрать allowlist ColType → invalid type уходит в Ent / не ErrInvalidInput.
func TestCreateColumn_InvalidColType_P3BE09(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:bases_coltype_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("col-" + uuid.NewString()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	svc := bases.NewService(bases.NewEntRepository(client))
	created, err := svc.CreateBase(ctx, owner.ID, bases.CreateBaseRequest{Name: "Cols"})
	require.NoError(t, err)
	tableID := created.Tables[0].ID

	_, err = svc.CreateColumn(ctx, owner.ID, created.ID, tableID, bases.CreateColumnRequest{
		Name: "Bad", ColType: bases.ColumnType("blob"),
	})
	require.ErrorIs(t, err, bases.ErrInvalidInput)
}

// Mutation: version=0 → ErrVersionConflict / last-write-wins → ErrorIs ErrInvalidInput краснеет.
func TestUpdateRecord_VersionZero_P3BE09(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:bases_ver0_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("ver0-" + uuid.NewString()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	svc := bases.NewService(bases.NewEntRepository(client))
	detail, err := svc.CreateBase(ctx, owner.ID, bases.CreateBaseRequest{Name: "V0"})
	require.NoError(t, err)
	tableID := detail.Tables[0].ID

	rec, err := svc.CreateRecord(ctx, detail.ID, tableID, owner.ID, bases.CreateRecordRequest{
		Cells: map[string]interface{}{},
	})
	require.NoError(t, err)

	_, err = svc.UpdateRecord(ctx, owner.ID, detail.ID, tableID, rec.ID, bases.UpdateRecordRequest{
		Cells: map[string]interface{}{}, Version: 0,
	})
	require.ErrorIs(t, err, bases.ErrInvalidInput)
}
