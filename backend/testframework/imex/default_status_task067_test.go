//go:build hubtest

package imex_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/internal/imex"
)

// TestAllocateIssueKey_PrefersDefaultStatus_Task067 — W7-A3-013 / P3-BE-15:
// default workflow status = is_default (не произвольный First).
// Мутация: вернуть First() без IsDefault → может вернуть Done.
func TestAllocateIssueKey_PrefersDefaultStatus_Task067(t *testing.T) {
	dsn := "file:imex_defstat_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := t.Context()
	_, err := client.User.Create().
		SetEmail("imex-" + uuid.New().String()[:8] + "@test.com").
		SetName("Imex").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().SetName("Imex Def").SetKey("IDF").Save(ctx)
	require.NoError(t, err)

	done, err := client.WorkflowStatus.Create().
		SetName("Done").
		SetCategory(workflowstatus.CategoryDone).
		SetSortOrder(0).
		SetIsDefault(false).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	todo, err := client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetSortOrder(99).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)
	_ = done

	repo := imex.NewEntRepository(client)
	_, _, statusID, err := repo.AllocateIssueKey(ctx, proj.ID)
	require.NoError(t, err)
	assert.Equal(t, todo.ID, statusID, "must pick is_default status, not First() by sort_order")
}
