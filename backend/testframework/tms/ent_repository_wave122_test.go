package tms_test

// Тесты репозитория TMS (волна 122): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Фаза 122 — ent-репозиторий suite/run coverage..

func TestEntRepository_Wave122_UpdateAndMoveSuite(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)

	parent, err := repo.CreateSuite(ctx, projectID, tms.CreateSuiteInput{Name: "Parent"})
	require.NoError(t, err)

	child, err := repo.CreateSuite(ctx, projectID, tms.CreateSuiteInput{
		Name: "Child", ParentID: &parent.ID,
	})
	require.NoError(t, err)

	newName := "Child renamed"
	updated, err := repo.UpdateSuite(ctx, projectID, child.ID, tms.UpdateSuiteInput{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)

	_, err = repo.MoveSuite(ctx, child.ID, &parent.ID)
	require.NoError(t, err)
}

// TestEntRepository_Wave122_ReopenTestRun проверяет слой репозитория: ent репозиторий wave 122 reopen test run.
func TestEntRepository_Wave122_ReopenTestRun(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Reopen", CreatedByID: userID,
	})
	require.NoError(t, err)

	now := time.Now()
	_, err = repo.UpdateTestRunStatus(ctx, run.ID, "in_progress", nil)
	require.NoError(t, err)
	_, err = repo.UpdateTestRunStatus(ctx, run.ID, "completed", &now)
	require.NoError(t, err)

	reopened, err := repo.ReopenTestRun(ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, "in_progress", string(reopened.Status))
}

// TestEntRepository_Wave122_UpdateTestRunStats проверяет слой репозитория: ent репозиторий wave 122 обновление test run stats.
func TestEntRepository_Wave122_UpdateTestRunStats(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Stats", CreatedByID: userID,
	})
	require.NoError(t, err)

	_, err = repo.UpdateTestRunStats(ctx, run.ID, tms.TestRunStats{
		PassedCount: 10, FailedCount: 2, SkippedCount: 1,
	})
	require.NoError(t, err)
	updated, err := repo.GetTestRun(ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, 10, updated.PassedCount)
	assert.Equal(t, 2, updated.FailedCount)
}
