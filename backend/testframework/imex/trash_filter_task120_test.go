package imex_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/internal/imex"
)

// TestListProjectIssuesPlain_ExcludesTrash_W7A3008 — soft-deleted issues не попадают в ZIP path.
// Mutation: убрать issue.DeletedAtIsNil() из ListProjectIssuesPlain → Len==2 FAIL.
func TestListProjectIssuesPlain_ExcludesTrash_W7A3008(t *testing.T) {
	dsn := "file:imex_trash_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := t.Context()
	user, err := client.User.Create().
		SetEmail("trash-" + uuid.New().String()[:8] + "@test.com").
		SetName("Trash User").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Trash Project").
		SetKey("TR").
		Save(ctx)
	require.NoError(t, err)

	status, err := client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	repo := imex.NewEntRepository(client)

	active, err := repo.CreateIssue(ctx, imex.CreateIssueInput{
		Title:       "Active",
		IssueType:   issue.IssueTypeTask,
		Priority:    issue.PriorityMedium,
		IssueNumber: 1,
		IssueKey:    "TR-1",
		ProjectID:   proj.ID,
		StatusID:    status.ID,
		ReporterID:  &user.ID,
	})
	require.NoError(t, err)

	trashed, err := repo.CreateIssue(ctx, imex.CreateIssueInput{
		Title:       "In trash",
		IssueType:   issue.IssueTypeTask,
		Priority:    issue.PriorityLow,
		IssueNumber: 2,
		IssueKey:    "TR-2",
		ProjectID:   proj.ID,
		StatusID:    status.ID,
		ReporterID:  &user.ID,
	})
	require.NoError(t, err)

	_, err = client.Issue.UpdateOneID(trashed.ID).
		SetDeletedAt(time.Now().UTC()).
		Save(ctx)
	require.NoError(t, err)

	plain, err := repo.ListProjectIssuesPlain(ctx, proj.ID)
	require.NoError(t, err)
	require.Len(t, plain, 1)
	assert.Equal(t, active.ID, plain[0].ID)
	assert.Equal(t, "Active", plain[0].Title)

	csvPath, err := repo.ListProjectIssuesForExport(ctx, proj.ID)
	require.NoError(t, err)
	require.Len(t, csvPath, 1)
	assert.Equal(t, active.ID, csvPath[0].ID)
}
