package tracker_test

// Тесты репозитория трекера задач: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEntRepository_LabelsAndGroups проверяет слой репозитория: ent репозиторий метки and groups.
func TestEntRepository_LabelsAndGroups(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	group, err := env.repo.CreateLabelGroup(env.ctx, env.projectID, "Priority")
	require.NoError(t, err)

	label, err := env.repo.CreateLabel(env.ctx, env.projectID, "urgent", "#FF0000", &group.ID)
	require.NoError(t, err)
	assert.Equal(t, "urgent", label.Name)

	labels, err := env.repo.ListLabels(env.ctx, env.projectID)
	require.NoError(t, err)
	require.Len(t, labels, 1)

	groups, err := env.repo.ListLabelGroups(env.ctx, env.projectID)
	require.NoError(t, err)
	require.Len(t, groups, 1)

	require.NoError(t, env.repo.DeleteLabel(env.ctx, label.ID))
	labels, err = env.repo.ListLabels(env.ctx, env.projectID)
	require.NoError(t, err)
	assert.Empty(t, labels)
}

// TestEntRepository_IssueLinks проверяет слой репозитория: ent репозиторий issue links.
func TestEntRepository_IssueLinks(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	source, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Source", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	target, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Target", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	link, err := env.repo.CreateIssueLink(env.ctx, source.ID, target.ID, "blocks", "")
	require.NoError(t, err)

	links, err := env.repo.ListIssueLinks(env.ctx, source.ID)
	require.NoError(t, err)
	require.Len(t, links, 1)

	_, err = env.repo.CreateIssueLink(env.ctx, source.ID, target.ID, "blocks", "")
	require.Error(t, err)

	require.NoError(t, env.repo.DeleteIssueLink(env.ctx, source.ID, link.ID))
	links, err = env.repo.ListIssueLinks(env.ctx, source.ID)
	require.NoError(t, err)
	assert.Empty(t, links)
}

// TestEntRepository_TimeLogsAndActivity проверяет слой репозитория: ent репозиторий time logs and activity.
func TestEntRepository_TimeLogsAndActivity(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Timed", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	tl, err := env.repo.LogTime(env.ctx, issue.ID, env.userID, 60, "coding")
	require.NoError(t, err)
	assert.Equal(t, 60, tl.Minutes)

	logs, err := env.repo.ListTimeLogs(env.ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, logs, 1)

	_, err = env.repo.LogActivity(env.ctx, issue.ID, env.userID, "commented", []tracker.FieldChange{
		{Field: "status", OldValue: "todo", NewValue: "done"},
	})
	require.NoError(t, err)

	activity, err := env.repo.ListActivities(env.ctx, issue.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, activity)
}

// TestEntRepository_SeedWorkflowAndBoard проверяет слой репозитория: ent репозиторий seed workflow and board.
func TestEntRepository_SeedWorkflowAndBoard(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	proj2, err := env.client.Project.Create().
		SetName("Fresh").
		SetKey("FRH").
		Save(env.ctx)
	require.NoError(t, err)

	require.NoError(t, env.repo.SeedDefaultWorkflow(env.ctx, proj2.ID))

	statuses, err := env.repo.ListWorkflowStatuses(env.ctx, proj2.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(statuses), 2)

	board, err := env.repo.CreateBoard(env.ctx, proj2.ID, "Main Board", nil)
	require.NoError(t, err)
	assert.Equal(t, "Main Board", board.Name)

	boards, err := env.repo.ListBoards(env.ctx, proj2.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, boards)
}

// TestEntRepository_DeleteWorkflowStatus проверяет слой репозитория: ent репозиторий удаление workflow status.
func TestEntRepository_DeleteWorkflowStatus(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	status, err := env.repo.CreateWorkflowStatus(env.ctx, env.projectID, "Done", "done", "#00FF00", 99)
	require.NoError(t, err)

	require.NoError(t, env.repo.DeleteWorkflowStatus(env.ctx, status.ID))
}
