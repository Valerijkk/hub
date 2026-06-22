package tms_test

// Тесты сервиса TMS (волна 115): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/testrun"
)

// TestReopenTestRun_NotFound проверяет сценарий: reopen test run not found.
func TestReopenTestRun_NotFound(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	_, err := svc.ReopenTestRun(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

// TestReopenTestRun_InvalidStatus проверяет отклонение при reopen test run некорректный status.
func TestReopenTestRun_InvalidStatus(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	runID := uuid.New()
	repo.runs[runID] = &ent.TestRun{
		ID:     runID,
		Status: testrun.StatusInProgress,
	}

	_, err := svc.ReopenTestRun(ctx, runID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "переоткрыть")
}

// TestCompleteTestRun_NotFound проверяет сценарий: complete test run not found.
func TestCompleteTestRun_NotFound(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	_, err := svc.CompleteTestRun(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

// TestCreateTestResult_InvalidTestCaseID проверяет отклонение при создание test result некорректный test case id.
func TestCreateTestResult_InvalidTestCaseID(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	runID := uuid.New()
	repo.runs[runID] = &ent.TestRun{ID: runID, Status: testrun.StatusInProgress}

	_, err := svc.CreateTestResult(context.Background(), runID, tms.CreateTestResultRequest{
		TestCaseID: "not-uuid",
		Status:     "passed",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "test_case_id")
}
