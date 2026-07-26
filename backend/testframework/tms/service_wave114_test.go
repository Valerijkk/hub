package tms_test

// Тесты сервиса TMS (волна 114): интеграционные и unit-сценарии testframework.
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

// TestExecuteTestResult_RunNotFound проверяет сценарий: execute test result run not found.
func TestExecuteTestResult_RunNotFound(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	status := "passed"
	_, err := svc.ExecuteTestResult(context.Background(), uuid.New(), uuid.New(), uuid.New(), tms.UpdateTestResultRequest{
		Status: &status,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

// TestExecuteTestResult_RunNotInProgress проверяет сценарий: execute test result run not in progress.
func TestExecuteTestResult_RunNotInProgress(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()

	run, err := svc.CreateTestRun(ctx, uuid.New(), uuid.New(), tms.CreateTestRunRequest{Title: "Draft run"})
	require.NoError(t, err)
	runID, _ := uuid.Parse(run.ID)

	resultID := uuid.New()
	repo.results[resultID] = &ent.TestResult{ID: resultID, Status: "unknown"}

	status := "passed"
	_, err = svc.ExecuteTestResult(ctx, runID, resultID, uuid.New(), tms.UpdateTestResultRequest{Status: &status})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "in_progress")
}

// TestExecuteTestResult_InvalidFileID проверяет отклонение при execute test result некорректный file id.
func TestExecuteTestResult_InvalidFileID(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()

	runID := uuid.New()
	repo.runs[runID] = &ent.TestRun{ID: runID, Status: testrun.StatusInProgress}
	resultID := uuid.New()
	repo.results[resultID] = &ent.TestResult{ID: resultID, Status: "unknown"}

	status := "passed"
	_, err := svc.ExecuteTestResult(ctx, runID, resultID, uuid.New(), tms.UpdateTestResultRequest{
		Status:  &status,
		FileIDs: []string{"bad-file-id"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "файла")
}

// TestExecuteTestResult_ResultNotFound проверяет сценарий: execute test result result not found.
func TestExecuteTestResult_ResultNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(&notFoundResultUpdateRepo{mockRepository: repo}, nil)
	ctx := context.Background()

	runID := uuid.New()
	repo.runs[runID] = &ent.TestRun{ID: runID, Status: testrun.StatusInProgress}

	status := "passed"
	_, err := svc.ExecuteTestResult(ctx, runID, uuid.New(), uuid.New(), tms.UpdateTestResultRequest{Status: &status})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

type notFoundResultUpdateRepo struct {
	*mockRepository
}

func (r *notFoundResultUpdateRepo) UpdateTestResult(_ context.Context, _, _ uuid.UUID, _ tms.UpdateTestResultInput) (*ent.TestResult, error) {
	return nil, &ent.NotFoundError{}
}
