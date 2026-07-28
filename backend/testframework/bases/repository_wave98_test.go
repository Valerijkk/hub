package bases_test
import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

func setupBasesRepo(t *testing.T) (*EntRepository, context.Context, uuid.UUID) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:bases_repo_"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()
	owner, err := client.User.Create().
		SetEmail("owner-repo-"+t.Name()+"@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)
	return NewEntRepository(client), ctx, owner.ID
}

func TestEntRepository_UpdateAndSoftDeleteTable(t *testing.T) {
	repo, ctx, ownerID := setupBasesRepo(t)

	base, err := repo.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "Repo Base"})
	require.NoError(t, err)

	table, err := repo.CreateTable(ctx, base.ID, CreateTableRequest{Name: "Items"})
	require.NoError(t, err)

	sort := 5
	updated, err := repo.UpdateTable(ctx, table.ID, UpdateTableRequest{
		Name:      strPtr("Updated"),
		SortOrder: &sort,
		Version:   table.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
	assert.Equal(t, 5, updated.SortOrder)

	require.NoError(t, repo.SoftDeleteTable(ctx, table.ID))
	_, err = repo.GetTable(ctx, base.ID, table.ID)
	assert.True(t, ent.IsNotFound(err))
}

func TestEntRepository_UpdateAndSoftDeleteColumn(t *testing.T) {
	repo, ctx, ownerID := setupBasesRepo(t)

	base, err := repo.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "Cols"})
	require.NoError(t, err)
	table, err := repo.CreateTable(ctx, base.ID, CreateTableRequest{Name: "T"})
	require.NoError(t, err)

	col, err := repo.CreateColumn(ctx, table.ID, CreateColumnRequest{
		Name: "Status", ColType: ColumnSelect,
	})
	require.NoError(t, err)

	colType := ColumnText
	updated, err := repo.UpdateColumn(ctx, col.ID, UpdateColumnRequest{
		Name:    strPtr("Title"),
		ColType: &colType,
		Version: col.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "Title", updated.Name)

	require.NoError(t, repo.SoftDeleteColumn(ctx, col.ID))
	view, err := repo.GetTable(ctx, base.ID, table.ID)
	require.NoError(t, err)
	// CreateTable атомарно создаёт Name — soft-delete Status оставляет Name.
	require.Len(t, view.Edges.Columns, 1)
	assert.Equal(t, "Name", view.Edges.Columns[0].Name)
}

func strPtr(s string) *string { return &s }
