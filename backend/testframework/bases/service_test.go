package bases_test
import (
	"github.com/valerii/hub/backend/internal/bases"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

func setupBasesService(t *testing.T) (*Service, uuid.UUID) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:bases_svc_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	owner, err := client.User.Create().
		SetEmail("owner@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(context.Background())
	require.NoError(t, err)

	return NewService(NewEntRepository(client)), owner.ID
}

func TestService_CreateBaseInvalidName(t *testing.T) {
	svc, ownerID := setupBasesService(t)
	_, err := svc.CreateBase(context.Background(), ownerID, CreateBaseRequest{Name: "   "})
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestService_UpdateBaseAndVersionConflict(t *testing.T) {
	svc, ownerID := setupBasesService(t)
	ctx := context.Background()

	created, err := svc.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "CRM DB"})
	require.NoError(t, err)

	newName := "Sales CRM"
	updated, err := svc.UpdateBase(ctx, ownerID, created.ID, UpdateBaseRequest{
		Name:    &newName,
		Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "Sales CRM", updated.Name)

	_, err = svc.UpdateBase(ctx, ownerID, created.ID, UpdateBaseRequest{
		Name:    &newName,
		Version: created.Version,
	})
	assert.ErrorIs(t, err, ErrVersionConflict)
}

func TestService_DeleteBase(t *testing.T) {
	svc, ownerID := setupBasesService(t)
	ctx := context.Background()

	created, err := svc.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "Temp"})
	require.NoError(t, err)

	err = svc.DeleteBase(ctx, ownerID, created.ID)
	require.NoError(t, err)

	_, err = svc.GetBase(ctx, ownerID, created.ID)
	assert.ErrorIs(t, err, bases.ErrNotFound)

	err = svc.DeleteBase(ctx, ownerID, uuid.New())
	assert.ErrorIs(t, err, bases.ErrNotFound)
}

func TestService_TableColumnRecordLifecycle(t *testing.T) {
	svc, ownerID := setupBasesService(t)
	ctx := context.Background()

	base, err := svc.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "Inventory"})
	require.NoError(t, err)

	table, err := svc.CreateTable(ctx, ownerID, base.ID, CreateTableRequest{Name: "Assets"})
	require.NoError(t, err)
	assert.Equal(t, "Assets", table.Name)

	view, err := svc.GetTableView(ctx, ownerID, base.ID, table.ID)
	require.NoError(t, err)
	require.NotEmpty(t, view.Columns)

	col, err := svc.CreateColumn(ctx, ownerID, base.ID, table.ID, CreateColumnRequest{
		Name:    "Qty",
		ColType: ColumnNumber,
	})
	require.NoError(t, err)

	view, err = svc.GetTableView(ctx, ownerID, base.ID, table.ID)
	require.NoError(t, err)
	nameColID := view.Columns[0].ID.String()

	rec, err := svc.CreateRecord(ctx, base.ID, table.ID, ownerID, CreateRecordRequest{
		Cells: map[string]interface{}{
			nameColID:  "Laptop",
			col.ID.String(): float64(3),
			"unknown-col":   "ignored",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "Laptop", rec.Cells[nameColID])
	assert.NotContains(t, rec.Cells, "unknown-col")

	updated, err := svc.UpdateRecord(ctx, ownerID, base.ID, table.ID, rec.ID, UpdateRecordRequest{
		Cells:   map[string]interface{}{nameColID: "Desktop"},
		Version: rec.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "Desktop", updated.Cells[nameColID])
	// W2-A5-002: partial PUT merge — Qty не затирается.
	assert.Equal(t, float64(3), updated.Cells[col.ID.String()])

	err = svc.DeleteRecord(ctx, ownerID, base.ID, table.ID, rec.ID)
	require.NoError(t, err)

	view, err = svc.GetTableView(ctx, ownerID, base.ID, table.ID)
	require.NoError(t, err)
	assert.Empty(t, view.Records)
}

func TestService_ListBases(t *testing.T) {
	svc, ownerID := setupBasesService(t)
	ctx := context.Background()

	_, err := svc.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "A"})
	require.NoError(t, err)
	_, err = svc.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "B"})
	require.NoError(t, err)

	list, err := svc.ListBases(ctx, ownerID)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}
