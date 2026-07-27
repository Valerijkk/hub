package tms_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/testresult"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestExecuteTestResult_SetsExecutedBy_A3_017 — userID must be persisted as
// ExecutedByID (mutation: omit ExecutedByID in input → FAIL).
func TestExecuteTestResult_SetsExecutedBy_A3_017(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	run, err := svc.CreateTestRun(ctx, projectID, userID, tms.CreateTestRunRequest{Title: "Exec by"})
	require.NoError(t, err)
	runID, _ := uuid.Parse(run.ID)
	_, err = svc.UpdateRunStatus(ctx, projectID, runID, "in_progress")
	require.NoError(t, err)

	resultID := uuid.New()
	repo.results[resultID] = &ent.TestResult{
		ID: resultID, Status: testresult.StatusUnknown, CreatedAt: time.Now(),
	}
	repo.runResults[runID] = []uuid.UUID{resultID}

	passed := "passed"
	resp, err := svc.ExecuteTestResult(ctx, projectID, runID, resultID, userID, tms.UpdateTestResultRequest{
		Status: &passed,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.ExecutedBy)
	assert.Equal(t, userID.String(), resp.ExecutedBy.ID)

	stored := repo.results[resultID]
	require.NotNil(t, stored.Edges.ExecutedBy)
	assert.Equal(t, userID, stored.Edges.ExecutedBy.ID)
}
