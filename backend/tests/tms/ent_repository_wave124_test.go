package tms_test

// Тесты репозитория TMS (волна 124): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Фаза 124 — пробелы покрытия ent-репозитория for runs and bug creation..

func TestEntRepository_Wave124_CreateBugIssue(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Fail case", CaseType: "classic", Priority: "high", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	bug, err := repo.CreateBugIssue(ctx, projectID, userID, tc.ID, "Bug title", "steps", "critical")
	require.NoError(t, err)
	assert.Equal(t, "Bug title", bug.Title)

	linked, err := repo.GetTestCaseLinkedIssues(ctx, tc.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, linked)
}

// TestEntRepository_Wave124_ReopenTestRun проверяет слой репозитория: ent репозиторий wave 124 reopen test run.
func TestEntRepository_Wave124_ReopenTestRun(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Reopen", CreatedByID: userID,
	})
	require.NoError(t, err)

	now := time.Now().UTC()
	_, err = repo.UpdateTestRunStatus(ctx, run.ID, "completed", &now)
	require.NoError(t, err)

	reopened, err := repo.ReopenTestRun(ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, "in_progress", string(reopened.Status))
}
