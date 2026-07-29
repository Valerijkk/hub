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

// TestListSprintIssues_ExcludesTrash_A3W73_001 — soft-deleted issue не в sprint PDF/list.
// Mutation: убрать issue.DeletedAtIsNil() из ListSprintIssues → Len==2 FAIL.
func TestListSprintIssues_ExcludesTrash_A3W73_001(t *testing.T) {
	dsn := "file:imex_sprint_trash_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := t.Context()
	user, err := client.User.Create().
		SetEmail("sprint-trash-" + uuid.New().String()[:8] + "@test.com").
		SetName("Sprint Trash").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Sprint Trash Project").
		SetKey("ST").
		Save(ctx)
	require.NoError(t, err)

	status, err := client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	now := time.Now().UTC()
	sp, err := client.Sprint.Create().
		SetName("Trash Sprint").
		SetStartDate(now).
		SetEndDate(now.AddDate(0, 0, 14)).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	active, err := client.Issue.Create().
		SetIssueNumber(1).
		SetIssueKey("ST-1").
		SetTitle("Active in sprint").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium).
		SetProjectID(proj.ID).
		SetStatusID(status.ID).
		SetReporterID(user.ID).
		SetSprintID(sp.ID).
		Save(ctx)
	require.NoError(t, err)

	trashed, err := client.Issue.Create().
		SetIssueNumber(2).
		SetIssueKey("ST-2").
		SetTitle("Trashed in sprint").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityLow).
		SetProjectID(proj.ID).
		SetStatusID(status.ID).
		SetReporterID(user.ID).
		SetSprintID(sp.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Issue.UpdateOneID(trashed.ID).
		SetDeletedAt(time.Now().UTC()).
		Save(ctx)
	require.NoError(t, err)

	repo := imex.NewEntRepository(client)
	gotSprint, err := repo.GetSprintInProject(ctx, proj.ID, sp.ID)
	require.NoError(t, err)

	issues, err := repo.ListSprintIssues(ctx, gotSprint)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Equal(t, active.ID, issues[0].ID)
	assert.Equal(t, "Active in sprint", issues[0].Title)
}
