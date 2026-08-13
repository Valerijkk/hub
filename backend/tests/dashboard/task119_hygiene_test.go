//go:build hubtest

package dashboard_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// TestListTopProjectsByIssueCount_Aggregate_W7A4008 — топ без WithIssues().All().
// Mutation: вернуть Project.Query().WithIssues().All → unbounded memory; count asserts держат семантику.
func TestListTopProjectsByIssueCount_Aggregate_W7A4008(t *testing.T) {
	repo, client, ctx, userID, projectID := newDashboardRepoEnv(t)

	todo, err := client.WorkflowStatus.Query().
		Where(workflowstatus.HasProjectWith(project.ID(projectID))).
		Only(ctx)
	require.NoError(t, err)

	p2, err := client.Project.Create().
		SetName("Second").
		SetKey("SEC").
		Save(ctx)
	require.NoError(t, err)
	st2, err := client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetProjectID(p2.ID).
		Save(ctx)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		_, err = client.Issue.Create().
			SetProjectID(projectID).
			SetStatusID(todo.ID).
			SetReporterID(userID).
			SetIssueNumber(100 + i).
			SetIssueKey(fmt.Sprintf("TOPA-%d", i)).
			SetTitle("a").
			SetIssueType(issue.IssueTypeTask).
			SetPriority(issue.PriorityMedium).
			Save(ctx)
		require.NoError(t, err)
	}
	_, err = client.Issue.Create().
		SetProjectID(p2.ID).
		SetStatusID(st2.ID).
		SetReporterID(userID).
		SetIssueNumber(1).
		SetIssueKey("TOPB-1").
		SetTitle("b").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityLow).
		Save(ctx)
	require.NoError(t, err)

	top, err := repo.ListTopProjectsByIssueCount(ctx, 5)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(top), 2)
	assert.Equal(t, projectID, top[0].ID)
	assert.Equal(t, 3, top[0].IssueCount)
	assert.Equal(t, p2.ID, top[1].ID)
	assert.Equal(t, 1, top[1].IssueCount)
}

// TestGetPersonalDashboard_FailVisible_W7A4009 — PG widget err → error, не silent 200.
// Mutation: return data, nil после slog.Error → require.Error FAIL.
func TestGetPersonalDashboard_FailVisible_W7A4009(t *testing.T) {
	svc, client, ctx, userID := newDashboardTestEnv(t)
	require.NoError(t, client.Close())

	_, err := svc.GetPersonalDashboardData(ctx, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load my tasks")
}
