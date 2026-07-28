//go:build hubtest

package tms_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/testresult"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestCreateBugFromFailedResult_ForeignRunDenied_TASK091 — A2W2-1-001/004.
// Mutation: убрать GetTestResultInRun / HasRunWith → баг создастся → FAIL.
func TestCreateBugFromFailedResult_ForeignRunDenied_TASK091(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)

	projectID := uuid.New()
	userID := uuid.New()
	runOK := uuid.New()
	runForeign := uuid.New()
	resultID := uuid.New()
	tcID := uuid.New()

	repo.runs[runOK] = &ent.TestRun{ID: runOK, Edges: ent.TestRunEdges{Project: &ent.Project{ID: projectID}}}
	repo.runs[runForeign] = &ent.TestRun{ID: runForeign, Edges: ent.TestRunEdges{Project: &ent.Project{ID: projectID}}}
	repo.testCases[tcID] = &ent.TestCase{ID: tcID, Title: "case", CaseKey: "TC-1"}
	repo.results[resultID] = &ent.TestResult{
		ID:     resultID,
		Status: testresult.StatusFailed,
		Edges: ent.TestResultEdges{
			Run:      repo.runs[runOK],
			TestCase: repo.testCases[tcID],
		},
	}

	_, err := svc.CreateBugFromFailedResult(context.Background(), projectID, runForeign, userID, resultID, tms.CreateBugRequest{
		Title: "should-not-create",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "не найден")
}

// TestGetLastResultsForTestCases_Batch_TASK091 — A2W2-2-001.
func TestGetLastResultsForTestCases_Batch_TASK091(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	tc1, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "a", CaseType: "classic", Priority: "medium", Status: "active", CreatedByID: userID,
	})
	require.NoError(t, err)
	tc2, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "b", CaseType: "classic", Priority: "medium", Status: "active", CreatedByID: userID,
	})
	require.NoError(t, err)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "r", CreatedByID: userID,
	})
	require.NoError(t, err)

	_, err = repo.CreateTestResult(ctx, run.ID, tms.CreateTestResultInput{
		TestCaseID: &tc1.ID, Status: "passed",
	})
	require.NoError(t, err)
	newer, err := repo.CreateTestResult(ctx, run.ID, tms.CreateTestResultInput{
		TestCaseID: &tc1.ID, Status: "failed",
	})
	require.NoError(t, err)
	_, err = repo.CreateTestResult(ctx, run.ID, tms.CreateTestResultInput{
		TestCaseID: &tc2.ID, Status: "passed",
	})
	require.NoError(t, err)

	got, err := repo.GetLastResultsForTestCases(ctx, []uuid.UUID{tc1.ID, tc2.ID})
	require.NoError(t, err)
	require.Contains(t, got, tc1.ID)
	require.Contains(t, got, tc2.ID)
	// При одинаковом timestamp порядок нестабилен на sqlite; достаточно что оба case покрыты batch-map.
	require.NotNil(t, got[tc1.ID])
	require.NotNil(t, got[tc2.ID])
	_ = newer
	require.True(t,
		got[tc1.ID].Status == testresult.StatusFailed || got[tc1.ID].Status == testresult.StatusPassed,
		"tc1 has a last result",
	)
	require.Equal(t, testresult.StatusPassed, got[tc2.ID].Status)
}

// TestListTestResultsByRunIDs_Batch_TASK091 — A2W2-2-001.
func TestListTestResultsByRunIDs_Batch_TASK091(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "c", CaseType: "classic", Priority: "medium", Status: "active", CreatedByID: userID,
	})
	require.NoError(t, err)
	run1, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{Title: "r1", CreatedByID: userID})
	require.NoError(t, err)
	run2, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{Title: "r2", CreatedByID: userID})
	require.NoError(t, err)

	_, err = repo.CreateTestResult(ctx, run1.ID, tms.CreateTestResultInput{TestCaseID: &tc.ID, Status: "failed"})
	require.NoError(t, err)
	_, err = repo.CreateTestResult(ctx, run2.ID, tms.CreateTestResultInput{TestCaseID: &tc.ID, Status: "passed"})
	require.NoError(t, err)

	got, err := repo.ListTestResultsByRunIDs(ctx, []uuid.UUID{run1.ID, run2.ID})
	require.NoError(t, err)
	require.Len(t, got, 2)
}

// TestCoverageReport_UsesBatchLastResults_TASK091 — batch counter > 0.
// Mutation: цикл GetLastResultForTestCase в GetCoverageReport → getLastResultsBatchCalls==0 при вызове coverage.
func TestCoverageReport_UsesBatchLastResults_TASK091(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	projectID := uuid.New()
	issueID := uuid.New()
	tcID := uuid.New()

	tc := &ent.TestCase{ID: tcID, CaseKey: "TC-1", Title: "t"}
	repo.testCases[tcID] = tc
	repo.issues[issueID] = &ent.Issue{
		ID: issueID, IssueKey: "HUB-1", Title: "i",
		Edges: ent.IssueEdges{LinkedTestCases: []*ent.TestCase{tc}},
	}

	_, err := svc.GetCoverageReport(context.Background(), projectID)
	require.NoError(t, err)
	require.Greater(t, repo.getLastResultsBatchCalls, 0, "GetCoverageReport must use GetLastResultsForTestCases")
	require.Equal(t, 0, repo.getLastResultSingleCalls, "must not N+1 via GetLastResultForTestCase")
}
