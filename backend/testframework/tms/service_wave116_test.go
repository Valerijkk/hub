package tms_test

// Тесты сервиса TMS (волна 116): интеграционные и unit-сценарии testframework.
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

// TestUpdateRunStatus_NotFound проверяет сценарий: обновление run status not found.
func TestUpdateRunStatus_NotFound(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	_, err := svc.UpdateRunStatus(context.Background(), uuid.New(), "completed")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

// TestAddTestCasesToRun_InvalidTestCaseID проверяет отклонение при add test cases to run некорректный test case id.
func TestAddTestCasesToRun_InvalidTestCaseID(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	runID := uuid.New()
	repo.runs[runID] = &ent.TestRun{ID: runID, Status: testrun.StatusInProgress}

	_, err := svc.AddTestCasesToRun(context.Background(), runID, tms.AddTestCasesToRunRequest{
		CaseIDs: []string{"not-uuid"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "тест-кейса")
}

// TestStartTestRun_NotFound проверяет сценарий: start test run not found.
func TestStartTestRun_NotFound(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	_, err := svc.StartTestRun(context.Background(), uuid.New())
	require.Error(t, err)
}
