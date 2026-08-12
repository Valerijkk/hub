//go:build hubtest

package tms_test

// TASK-140 — A3W2-3 TMS repos residual: list/batch caps + soft-delete gaps.
// Mutation notes — см. комментарии у каждого TestTask140_*.

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	entissue "github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestTask140_ListTestResultsByRunIDs_Cap — A3W2-3-001.
// Mutation: убрать .Limit(maxListTestResults) → len == 3 при cap=2 → FAIL.
func TestTask140_ListTestResultsByRunIDs_Cap(t *testing.T) {
	restore := tms.ExportSetMaxListTestResultsForTest(2)
	defer restore()

	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "ByRunIDs", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)
	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{Title: "R", CreatedByID: userID})
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		_, err = repo.CreateTestResult(ctx, run.ID, tms.CreateTestResultInput{
			TestCaseID: &tc.ID, Status: "passed",
		})
		require.NoError(t, err)
	}

	got, err := repo.ListTestResultsByRunIDs(ctx, []uuid.UUID{run.ID})
	require.NoError(t, err)
	assert.LessOrEqual(t, len(got), tms.ExportMaxListTestResults())
	assert.Equal(t, 2, len(got))
}

// TestTask140_GetTestRunWithResults_Cap — A3W2-3-001 nested WithResults Limit.
// Mutation: убрать Limit внутри WithResults → len == 3 при cap=2 → FAIL.
func TestTask140_GetTestRunWithResults_Cap(t *testing.T) {
	restore := tms.ExportSetMaxListTestResultsForTest(2)
	defer restore()

	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "WithResults", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)
	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{Title: "R", CreatedByID: userID})
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		_, err = repo.CreateTestResult(ctx, run.ID, tms.CreateTestResultInput{
			TestCaseID: &tc.ID, Status: "passed",
		})
		require.NoError(t, err)
	}

	got, err := repo.GetTestRunWithResults(ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(got.Edges.Results))

	gotInProj, err := repo.GetTestRunWithResultsInProject(ctx, projectID, run.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(gotInProj.Edges.Results))
}

// TestTask140_AddTestCasesToRun_BatchErr — A3W2-3-001.
// Mutation: убрать len > maxBatch check → oversized проходит без err → FAIL.
func TestTask140_AddTestCasesToRun_BatchErr(t *testing.T) {
	restore := tms.ExportSetMaxBatchCreateTestResultsForTest(2)
	defer restore()

	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{Title: "R", CreatedByID: userID})
	require.NoError(t, err)

	ids := make([]uuid.UUID, tms.ExportMaxBatchCreateTestResults()+1)
	for i := range ids {
		ids[i] = uuid.New()
	}
	err = repo.AddTestCasesToRun(ctx, run.ID, ids)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "слишком много")
}

// TestTask140_ListEnvironments_Cap — A3W2-3-001.
// Mutation: убрать Limit(maxListEnvironments) → len == 3 при cap=2 → FAIL.
func TestTask140_ListEnvironments_Cap(t *testing.T) {
	restore := tms.ExportSetMaxListEnvironmentsForTest(2)
	defer restore()

	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)
	for i := 0; i < 3; i++ {
		_, err := repo.CreateEnvironment(ctx, projectID, tms.CreateEnvironmentInput{
			Name: "env-" + uuid.New().String()[:8], Label: "staging",
		})
		require.NoError(t, err)
	}
	envs, err := repo.ListEnvironmentsByProject(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(envs))
}

// TestTask140_GetProjectIssuesWithTestCases_Cap — A3W2-3-001.
// Mutation: убрать Limit(maxListProjectIssues) → len == 3 при cap=2 → FAIL.
func TestTask140_GetProjectIssuesWithTestCases_Cap(t *testing.T) {
	restore := tms.ExportSetMaxListProjectIssuesForTest(2)
	defer restore()

	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	statusID := mustDefaultStatusID(t, repo, ctx, projectID)
	client := tms.ExportEntRepositoryClient(repo)
	for i := 0; i < 3; i++ {
		_, err := client.Issue.Create().
			SetIssueNumber(i + 1).
			SetIssueKey("CAP-" + uuid.New().String()[:6]).
			SetTitle("Issue").
			SetIssueType(entissue.IssueTypeTask).
			SetProjectID(projectID).
			SetReporterID(userID).
			SetStatusID(statusID).
			Save(ctx)
		require.NoError(t, err)
	}
	issues, err := repo.GetProjectIssuesWithTestCases(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(issues))
}

// TestTask140_GetLastResultsForTestCases_PerCaseNoFullHistory — A3W2-3-001.
// N исторических строк на case → 1 актуальный; batch oversized → err.
// Mutation: вернуть .All(ctx) без per-case First → (корректность latest может остаться,
// но) убрать batch cap на caseIDs → TestTask140_GetLastResults_BatchErr краснеет;
// убрать First → при All без Limit тест batch-err всё равно защищает вход.
func TestTask140_GetLastResultsForTestCases_PerCaseNoFullHistory(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	tcA, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "A", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)
	tcB, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "B", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)
	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{Title: "R", CreatedByID: userID})
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		status := "passed"
		if i == 4 {
			status = "failed"
		}
		_, err = repo.CreateTestResult(ctx, run.ID, tms.CreateTestResultInput{
			TestCaseID: &tcA.ID, Status: status,
		})
		require.NoError(t, err)
		_, err = repo.CreateTestResult(ctx, run.ID, tms.CreateTestResultInput{
			TestCaseID: &tcB.ID, Status: status,
		})
		require.NoError(t, err)
		time.Sleep(2 * time.Millisecond)
	}

	got, err := repo.GetLastResultsForTestCases(ctx, []uuid.UUID{tcA.ID, tcB.ID})
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "failed", string(got[tcA.ID].Status))
	assert.Equal(t, "failed", string(got[tcB.ID].Status))

	all, err := repo.ListTestResults(ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, 5, countResultsForCaseID(all, tcA.ID), "history в БД = 5, map = 1")
	assert.Equal(t, 5, countResultsForCaseID(all, tcB.ID))
}

// TestTask140_GetLastResults_BatchErr — A3W2-3-001 вход capped.
// Mutation: убрать len(caseIDs) > maxBatch → oversized без err → FAIL.
func TestTask140_GetLastResults_BatchErr(t *testing.T) {
	restore := tms.ExportSetMaxBatchCreateTestResultsForTest(2)
	defer restore()

	repo, ctx, _, _ := newEntRepositoryTestEnv(t)
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	_, err := repo.GetLastResultsForTestCases(ctx, ids)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "слишком много")
}

func countResultsForCaseID(results []*ent.TestResult, caseID uuid.UUID) int {
	n := 0
	for _, r := range results {
		if r.Edges.TestCase != nil && r.Edges.TestCase.ID == caseID {
			n++
		}
	}
	return n
}

// TestTask140_CountLabelsInProject_SkipsSoftDeleted — A3W2-3-002.
// Mutation: убрать label.DeletedAtIsNil() → Count=1 для trash → FAIL.
func TestTask140_CountLabelsInProject_SkipsSoftDeleted(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)
	client := tms.ExportEntRepositoryClient(repo)

	alive, err := client.Label.Create().
		SetName("alive").
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)
	trash, err := client.Label.Create().
		SetName("trash").
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)
	now := time.Now().UTC()
	_, err = client.Label.UpdateOneID(trash.ID).SetDeletedAt(now).Save(ctx)
	require.NoError(t, err)

	n, err := repo.CountLabelsInProject(ctx, projectID, []uuid.UUID{alive.ID, trash.ID})
	require.NoError(t, err)
	assert.Equal(t, 1, n, "soft-deleted label must not count")

	nAlive, err := repo.CountLabelsInProject(ctx, projectID, []uuid.UUID{alive.ID})
	require.NoError(t, err)
	assert.Equal(t, 1, nAlive)
}

// TestTask140_LinkedEdges_SkipSoftDeleted — A3W2-3-003.
// Mutation: убрать edge DeletedAtIsNil с WithLinked* → trash в ответе → FAIL.
func TestTask140_LinkedEdges_SkipSoftDeleted(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	client := tms.ExportEntRepositoryClient(repo)
	statusID := mustDefaultStatusID(t, repo, ctx, projectID)

	tcAlive, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Alive TC", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)
	tcTrash, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Trash TC", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	issueAlive, err := client.Issue.Create().
		SetIssueNumber(1).
		SetIssueKey("LNK-1").
		SetTitle("Alive issue").
		SetIssueType(entissue.IssueTypeTask).
		SetProjectID(projectID).
		SetReporterID(userID).
		SetStatusID(statusID).
		Save(ctx)
	require.NoError(t, err)
	issueTrash, err := client.Issue.Create().
		SetIssueNumber(2).
		SetIssueKey("LNK-2").
		SetTitle("Trash issue").
		SetIssueType(entissue.IssueTypeTask).
		SetProjectID(projectID).
		SetReporterID(userID).
		SetStatusID(statusID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, repo.LinkTestCaseToIssue(ctx, tcAlive.ID, issueAlive.ID))
	require.NoError(t, repo.LinkTestCaseToIssue(ctx, tcAlive.ID, issueTrash.ID))
	require.NoError(t, repo.LinkTestCaseToIssue(ctx, tcTrash.ID, issueAlive.ID))

	now := time.Now().UTC()
	_, err = client.Issue.UpdateOneID(issueTrash.ID).SetDeletedAt(now).Save(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.DeleteTestCase(ctx, projectID, tcTrash.ID))

	issues, err := repo.GetTestCaseLinkedIssues(ctx, tcAlive.ID)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Equal(t, issueAlive.ID, issues[0].ID)

	tcs, err := repo.GetIssueLinkedTestCases(ctx, issueAlive.ID)
	require.NoError(t, err)
	require.Len(t, tcs, 1)
	assert.Equal(t, tcAlive.ID, tcs[0].ID)
}
