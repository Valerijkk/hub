//go:build hubtest

package dashboard_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// TestEntRepository_AggregateProjectIssueStats_A13_003 — агрегаты совпадают с
// ручным подсчётом без материализации полного списка в hot path service.
// Mutation: вернуть service на ListProjectIssuesWithStatusAssignee.All →
// regression на больших проектах (здесь проверяем корректность counts).
func TestEntRepository_AggregateProjectIssueStats_A13_003(t *testing.T) {
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

	past := time.Now().UTC().Add(-48 * time.Hour)

	_, err = client.Issue.Create().
		SetProjectID(projectID).
		SetStatusID(todo.ID).
		SetReporterID(userID).
		SetIssueNumber(10).
		SetIssueKey("AGG-1").
		SetTitle("Open unassigned overdue bug").
		SetIssueType(issue.IssueTypeBug).
		SetPriority(issue.PriorityHigh).
		SetDueDate(past).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Issue.Create().
		SetProjectID(projectID).
		SetStatusID(todo.ID).
		SetReporterID(userID).
		SetAssigneeID(userID).
		SetIssueNumber(11).
		SetIssueKey("AGG-2").
		SetTitle("Open assigned task").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Issue.Create().
		SetProjectID(projectID).
		SetStatusID(done.ID).
		SetReporterID(userID).
		SetIssueNumber(12).
		SetIssueKey("AGG-3").
		SetTitle("Done bug").
		SetIssueType(issue.IssueTypeBug).
		SetPriority(issue.PriorityLow).
		Save(ctx)
	require.NoError(t, err)

	agg, err := repo.AggregateProjectIssueStats(ctx, projectID)
	require.NoError(t, err)

	assert.Equal(t, 2, agg.TasksByStatus[todo.Name])
	assert.Equal(t, 1, agg.TasksByStatus["Done"])
	assert.Equal(t, 1, agg.BugCount, "only open bugs")
	assert.Equal(t, 2, agg.Unassigned, "open bug + done bug without assignee")
	assert.Equal(t, 1, agg.Overdue, "only open overdue")
}
