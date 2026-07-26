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
	"github.com/valerii/hub/backend/ent/testrun"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestExecuteTestResult_RejectsWrongRun_A3_008 — result чужого run не обновляется.
// Mutation: UpdateTestResult без HasRunWith / без runID → обновление чужого result.
func TestExecuteTestResult_RejectsWrongRun_A3_008(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	userID := uuid.New()

	runA := uuid.New()
	runB := uuid.New()
	repo.runs[runA] = &ent.TestRun{ID: runA, Status: testrun.StatusInProgress}
	repo.runs[runB] = &ent.TestRun{ID: runB, Status: testrun.StatusInProgress}

	resultB := uuid.New()
	repo.results[resultB] = &ent.TestResult{
		ID: resultB, Status: testresult.StatusUnknown, CreatedAt: time.Now(),
		Edges: ent.TestResultEdges{Run: repo.runs[runB]},
	}
	repo.runResults[runB] = []uuid.UUID{resultB}

	status := "passed"
	_, err := svc.ExecuteTestResult(ctx, runA, resultB, userID, tms.UpdateTestResultRequest{Status: &status})
	require.Error(t, err)
	assert.Equal(t, testresult.StatusUnknown, repo.results[resultB].Status)
}

// TestCreateBugFromFailedResult_RejectsForeignProject_A3_008 — result чужого project → 404.
func TestCreateBugFromFailedResult_RejectsForeignProject_A3_008(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectA := uuid.New()
	projectB := uuid.New()

	runID := uuid.New()
	repo.runs[runID] = &ent.TestRun{
		ID: runID,
		Edges: ent.TestRunEdges{
			Project: &ent.Project{ID: projectB},
		},
	}
	resultID := uuid.New()
	repo.results[resultID] = &ent.TestResult{
		ID: resultID, Status: testresult.StatusFailed, CreatedAt: time.Now(),
		Edges: ent.TestResultEdges{Run: repo.runs[runID]},
	}

	_, err := svc.CreateBugFromFailedResult(ctx, projectA, uuid.New(), resultID, tms.CreateBugRequest{Title: "x"})
	require.Error(t, err)
}
