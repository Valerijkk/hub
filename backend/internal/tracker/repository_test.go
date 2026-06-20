package tracker

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

type trackerTestEnv struct {
	repo      Repository
	ctx       context.Context
	projectID uuid.UUID
	userID    uuid.UUID
	statusID  uuid.UUID
}

func newTrackerRepoTestEnv(t *testing.T) trackerTestEnv {
	t.Helper()
	dsn := "file:trackertest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("tracker-" + uuid.New().String()[:8] + "@test.com").
		SetName("Reporter").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Tracker Test").
		SetKey("TST").
		Save(ctx)
	require.NoError(t, err)

	status, err := client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetColor("#6B7280").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	return trackerTestEnv{
		repo:      NewEntRepository(client, nil),
		ctx:       ctx,
		projectID: proj.ID,
		userID:    user.ID,
		statusID:  status.ID,
	}
}

func TestEntRepository_CreateIssue_IncrementsCounter(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	created, err := env.repo.CreateIssue(env.ctx, env.projectID, CreateIssueInput{
		Title:      "First issue",
		IssueType:  "task",
		Priority:   "medium",
		ReporterID: env.userID,
	})
	require.NoError(t, err)
	assert.Equal(t, "TST-1", created.IssueKey)
	assert.Equal(t, 1, created.IssueNumber)
	assert.Equal(t, 1, created.Version)

	second, err := env.repo.CreateIssue(env.ctx, env.projectID, CreateIssueInput{
		Title:      "Second issue",
		IssueType:  "bug",
		Priority:   "high",
		ReporterID: env.userID,
	})
	require.NoError(t, err)
	assert.Equal(t, "TST-2", second.IssueKey)
}

func TestEntRepository_FindIssueByKey_NotFound(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	_, err := env.repo.FindIssueByKey(env.ctx, "MISSING-1")
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}

func TestEntRepository_UpdateIssue_OptimisticLock(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	created, err := env.repo.CreateIssue(env.ctx, env.projectID, CreateIssueInput{
		Title:      "Lock test",
		IssueType:  "task",
		ReporterID: env.userID,
	})
	require.NoError(t, err)

	newTitle := "Updated"
	updated, err := env.repo.UpdateIssue(env.ctx, created.ID, UpdateIssueInput{
		Title:   &newTitle,
		Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Title)
	assert.Equal(t, 2, updated.Version)

	_, err = env.repo.UpdateIssue(env.ctx, created.ID, UpdateIssueInput{
		Title:   &newTitle,
		Version: created.Version,
	})
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	assert.Equal(t, 409, appErr.Code)
}

func TestEntRepository_ListIssues_FiltersByType(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	for _, it := range []string{"task", "bug", "task"} {
		_, err := env.repo.CreateIssue(env.ctx, env.projectID, CreateIssueInput{
			Title: "x", IssueType: it, ReporterID: env.userID,
		})
		require.NoError(t, err)
	}

	items, total, err := env.repo.ListIssues(env.ctx, env.projectID, ListIssuesParams{IssueType: "task", Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
}

func TestEntRepository_WorkflowTransition(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	inProgress, err := env.repo.CreateWorkflowStatus(env.ctx, env.projectID, "In Progress", "in_progress", "#3B82F6", 1)
	require.NoError(t, err)

	_, err = env.repo.CreateWorkflowTransition(env.ctx, env.statusID, inProgress.ID)
	require.NoError(t, err)

	found, err := env.repo.FindTransition(env.ctx, env.statusID, inProgress.ID)
	require.NoError(t, err)
	assert.NotNil(t, found)

	_, err = env.repo.FindTransition(env.ctx, inProgress.ID, env.statusID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}
