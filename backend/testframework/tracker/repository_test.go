package tracker_test

// Тесты репозитория трекера задач: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/testframework/env"
)

type trackerTestEnv struct {
	client    *ent.Client
	repo      tracker.Repository
	ctx       context.Context
	projectID uuid.UUID
	userID    uuid.UUID
	statusID  uuid.UUID
}

func newTrackerRepoTestEnv(t *testing.T) trackerTestEnv {
	t.Helper()
	e := env.New(t, "tracker").Build()
	return trackerTestEnv{
		client:    e.Client,
		repo:      tracker.NewEntRepository(e.Client, nil),
		ctx:       e.Ctx,
		projectID: e.ProjectID,
		userID:    e.UserID,
		statusID:  e.StatusID,
	}
}

// TestEntRepository_CreateIssue_IncrementsCounter проверяет слой репозитория: ent репозиторий создание issue increments counter.
func TestEntRepository_CreateIssue_IncrementsCounter(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	created, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title:      "First issue",
		IssueType:  "task",
		Priority:   "medium",
		ReporterID: env.userID,
	})
	require.NoError(t, err)
	assert.Equal(t, "TST-1", created.IssueKey)
	assert.Equal(t, 1, created.IssueNumber)
	assert.Equal(t, 1, created.Version)

	second, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title:      "Second issue",
		IssueType:  "bug",
		Priority:   "high",
		ReporterID: env.userID,
	})
	require.NoError(t, err)
	assert.Equal(t, "TST-2", second.IssueKey)
}

// TestEntRepository_FindIssueByKey_NotFound проверяет слой репозитория: ent репозиторий find issue by key not found.
func TestEntRepository_FindIssueByKey_NotFound(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	_, err := env.repo.FindIssueByKey(env.ctx, "MISSING-1")
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}

// TestEntRepository_UpdateIssue_OptimisticLock проверяет слой репозитория: ent репозиторий обновление issue optimistic lock.
func TestEntRepository_UpdateIssue_OptimisticLock(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	created, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title:      "Lock test",
		IssueType:  "task",
		ReporterID: env.userID,
	})
	require.NoError(t, err)

	newTitle := "Updated"
	updated, err := env.repo.UpdateIssue(env.ctx, created.ID, tracker.UpdateIssueInput{
		Title:   &newTitle,
		Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Title)
	assert.Equal(t, 2, updated.Version)

	_, err = env.repo.UpdateIssue(env.ctx, created.ID, tracker.UpdateIssueInput{
		Title:   &newTitle,
		Version: created.Version,
	})
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	assert.Equal(t, 409, appErr.Code)
}

// TestEntRepository_ListIssues_FiltersByType проверяет слой репозитория: ent репозиторий список issues filters by type.
func TestEntRepository_ListIssues_FiltersByType(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	for _, it := range []string{"task", "bug", "task"} {
		_, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
			Title: "x", IssueType: it, ReporterID: env.userID,
		})
		require.NoError(t, err)
	}

	items, total, err := env.repo.ListIssues(env.ctx, env.projectID, tracker.ListIssuesParams{IssueType: "task", Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
}

// TestEntRepository_WorkflowTransition проверяет слой репозитория: ent репозиторий workflow transition.
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
