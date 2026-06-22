package dashboard_test
import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/testrun"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func TestEntRepository_ProjectScopedQueries(t *testing.T) {
	repo, client, ctx, userID, projectID := newDashboardRepoEnv(t)

	todo, err := client.WorkflowStatus.Query().
		Where(workflowstatus.HasProjectWith(project.ID(projectID))).
		Only(ctx)
	require.NoError(t, err)

	done, err := client.WorkflowStatus.Create().
		SetName("Done").
		SetCategory(workflowstatus.CategoryDone).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	now := time.Now().UTC()
	openIssue, err := client.Issue.Create().
		SetProjectID(projectID).
		SetStatusID(todo.ID).
		SetReporterID(userID).
		SetIssueNumber(1).
		SetIssueKey("DASH-1").
		SetTitle("Open task").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium).
		Save(ctx)
	require.NoError(t, err)

	completedIssue, err := client.Issue.Create().
		SetProjectID(projectID).
		SetStatusID(done.ID).
		SetReporterID(userID).
		SetIssueNumber(2).
		SetIssueKey("DASH-2").
		SetTitle("Done task").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityLow).
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, client.Issue.UpdateOneID(completedIssue.ID).SetUpdatedAt(now).Exec(ctx))

	issues, err := repo.ListProjectIssuesWithStatusAssignee(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, issues, 2)
	assert.Equal(t, openIssue.ID, issues[0].ID)

	completed, err := repo.ListProjectCompletedIssuesSince(ctx, projectID, now.Add(-time.Hour))
	require.NoError(t, err)
	assert.Len(t, completed, 1)
	assert.Equal(t, completedIssue.ID, completed[0].ID)

	count, err := repo.CountProjectCompletedIssuesInRange(ctx, projectID, now.Add(-time.Hour), now.Add(time.Hour))
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	run, err := client.TestRun.Create().
		SetProjectID(projectID).
		SetTitle("Nightly").
		SetStatus(testrun.StatusCompleted).
		SetCreatedByID(userID).
		Save(ctx)
	require.NoError(t, err)

	runs, err := repo.ListProjectRecentTestRuns(ctx, projectID, 5)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	assert.Equal(t, run.ID, runs[0].ID)

	_, err = client.ActivityLog.Create().
		SetIssueID(openIssue.ID).
		SetUserID(userID).
		SetAction("updated").
		Save(ctx)
	require.NoError(t, err)

	logs, err := repo.ListProjectActivityLogs(ctx, projectID, 10)
	require.NoError(t, err)
	assert.Len(t, logs, 1)
}
