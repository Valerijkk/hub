package dashboard_test
import (
	"github.com/valerii/hub/backend/internal/dashboard"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/testrun"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func TestGetPersonalDashboardData_WithTasksDeadlinesAndTime(t *testing.T) {
	repo, client, ctx, userID, projectID := newDashboardRepoEnv(t)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	svc := dashboard.NewDashboardService(repo, rdb, nil)

	openStatus, err := client.WorkflowStatus.Query().
		Where(workflowstatus.HasProjectWith(project.IDEQ(projectID))).
		Only(ctx)
	require.NoError(t, err)

	dueSoon := time.Now().UTC().Add(48 * time.Hour)
	iss, err := client.Issue.Create().
		SetProjectID(projectID).
		SetStatusID(openStatus.ID).
		SetReporterID(userID).
		SetAssigneeID(userID).
		SetIssueNumber(1).
		SetIssueKey("DASH-1").
		SetTitle("Personal task").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityHigh).
		SetDueDate(dueSoon).
		Save(ctx)
	require.NoError(t, err)

	now := time.Now().UTC()
	_, err = client.TimeLog.Create().
		SetIssueID(iss.ID).
		SetUserID(userID).
		SetMinutes(45).
		SetDescription("dashboard work").
		SetCreatedAt(now).
		Save(ctx)
	require.NoError(t, err)

	data, err := svc.GetPersonalDashboardData(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, data)
	require.Len(t, data.MyTasks, 1)
	assert.Equal(t, "DASH-1", data.MyTasks[0].Key)
	require.Len(t, data.Deadlines, 1)
	assert.Equal(t, "Personal task", data.Deadlines[0].Title)
	assert.Equal(t, 45, data.TimeToday)
}

func TestGetProjectStats_OverdueVelocityAndActivity(t *testing.T) {
	repo, client, ctx, userID, projectID := newDashboardRepoEnv(t)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	svc := dashboard.NewDashboardService(repo, rdb, nil)

	openStatus, err := client.WorkflowStatus.Query().
		Where(workflowstatus.HasProjectWith(project.IDEQ(projectID))).
		Only(ctx)
	require.NoError(t, err)

	doneStatus, err := client.WorkflowStatus.Create().
		SetName("Done").
		SetCategory(workflowstatus.CategoryDone).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	overdue := time.Now().Add(-24 * time.Hour)
	_, err = client.Issue.Create().
		SetProjectID(projectID).
		SetStatusID(openStatus.ID).
		SetReporterID(userID).
		SetIssueNumber(2).
		SetIssueKey("DASH-2").
		SetTitle("Overdue bug").
		SetIssueType(issue.IssueTypeBug).
		SetPriority(issue.PriorityMedium).
		SetDueDate(overdue).
		Save(ctx)
	require.NoError(t, err)

	completedIssue, err := client.Issue.Create().
		SetProjectID(projectID).
		SetStatusID(doneStatus.ID).
		SetReporterID(userID).
		SetIssueNumber(3).
		SetIssueKey("DASH-3").
		SetTitle("Completed").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityLow).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.TestRun.Create().
		SetProjectID(projectID).
		SetCreatedByID(userID).
		SetTitle("Nightly").
		SetStatus(testrun.StatusCompleted).
		SetPassedCount(8).
		SetFailedCount(2).
		SetSkippedCount(0).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ActivityLog.Create().
		SetIssueID(completedIssue.ID).
		SetUserID(userID).
		SetAction("issue.updated").
		Save(ctx)
	require.NoError(t, err)

	stats, err := svc.GetProjectStats(ctx, userID, projectID)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.BugCount)
	assert.Equal(t, 1, stats.Overdue)
	assert.NotEmpty(t, stats.BurndownData)
	assert.NotEmpty(t, stats.VelocityData)
	assert.NotEmpty(t, stats.TestTrends)
	assert.NotEmpty(t, stats.TeamActivity)
}

func TestGetPersonalDashboardData_WithRecentTestRun(t *testing.T) {
	repo, client, ctx, userID, projectID := newDashboardRepoEnv(t)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	svc := dashboard.NewDashboardService(repo, rdb, nil)

	_, err = client.TestRun.Create().
		SetProjectID(projectID).
		SetCreatedByID(userID).
		SetTitle("Smoke").
		SetStatus(testrun.StatusCompleted).
		SetPassedCount(5).
		SetFailedCount(0).
		SetSkippedCount(1).
		Save(ctx)
	require.NoError(t, err)

	data, err := svc.GetPersonalDashboardData(ctx, userID)
	require.NoError(t, err)
	require.Len(t, data.RecentRuns, 1)
	assert.Equal(t, "Smoke", data.RecentRuns[0].Name)
	assert.Equal(t, 5, data.RecentRuns[0].Passed)
}
