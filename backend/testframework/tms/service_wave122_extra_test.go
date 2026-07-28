package tms_test

// Тесты сервиса TMS (волна 122): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/testresult"
)

// Фаза 122 — пробелы покрытия сервиса: UpdateTestResult, Rollback, recomputeRunStats, validation paths..

func TestService_Wave122_UpdateTestResult_NotFound(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	status := "passed"
	_, err := svc.UpdateTestResult(context.Background(), uuid.New(), tms.UpdateTestResultRequest{Status: &status})
	require.Error(t, err)
}

// TestService_Wave122_RollbackTestCase_VersionNotFound проверяет бизнес-логику: сервис wave 122 rollback test case version not found.
func TestService_Wave122_RollbackTestCase_VersionNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Rollback me", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)

	_, err = svc.RollbackTestCase(ctx, projectID, created.CaseKey, uuid.New(), userID)
	require.Error(t, err)
}

// TestService_Wave122_UpdateTestCase_InvalidSuiteID проверяет отклонение при сервис wave 122 обновление test case некорректный suite id.
func TestService_Wave122_UpdateTestCase_InvalidSuiteID(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Suite bind", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)

	badSuite := "not-a-uuid"
	_, err = svc.UpdateTestCase(ctx, projectID, created.CaseKey, userID, tms.UpdateTestCaseRequest{SuiteID: &badSuite,
		Version: 1,})
	require.Error(t, err)
}

// TestService_Wave122_UpdateTestCase_ClearSuite проверяет бизнес-логику: сервис wave 122 обновление test case clear suite.
func TestService_Wave122_UpdateTestCase_ClearSuite(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	suiteID := uuid.New()
	repo.suites[suiteID] = &ent.TestSuite{
		ID: suiteID, Name: "S1",
		Edges: ent.TestSuiteEdges{Project: &ent.Project{ID: projectID}},
	}

	created, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "In suite", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)

	suiteStr := suiteID.String()
	_, err = svc.UpdateTestCase(ctx, projectID, created.CaseKey, userID, tms.UpdateTestCaseRequest{SuiteID: &suiteStr, Version: 1})
	require.NoError(t, err)

	empty := ""
	updated, err := svc.UpdateTestCase(ctx, projectID, created.CaseKey, userID, tms.UpdateTestCaseRequest{SuiteID: &empty, Version: 2})
	require.NoError(t, err)
	assert.NotNil(t, updated)
}

// TestService_Wave122_ExecuteTestResult_WithFiles проверяет бизнес-логику: сервис wave 122 execute test result with files.
func TestService_Wave122_ExecuteTestResult_WithFiles(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	run, err := svc.CreateTestRun(ctx, projectID, uuid.New(), tms.CreateTestRunRequest{Title: "Files run"})
	require.NoError(t, err)
	runID, _ := uuid.Parse(run.ID)
	_, err = svc.UpdateRunStatus(ctx, projectID, runID, "in_progress")
	require.NoError(t, err)

	resultID := uuid.New()
	repo.results[resultID] = &ent.TestResult{ID: resultID, Status: testresult.StatusUnknown, CreatedAt: time.Now()}

	fileID := uuid.New()
	passed := "passed"
	resp, err := svc.ExecuteTestResult(ctx, projectID, runID, resultID, uuid.New(), tms.UpdateTestResultRequest{
		Status:  &passed,
		FileIDs: []string{fileID.String()},
	})
	require.NoError(t, err)
	assert.Equal(t, "passed", resp.Status)
}

// TestService_Wave122_RecomputeRunStats_AllStatuses проверяет бизнес-логику: сервис wave 122 recompute run stats all statuses.
func TestService_Wave122_RecomputeRunStats_AllStatuses(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	statuses := []string{"passed", "failed", "broken", "skipped", "unknown"}
	caseIDs := make([]string, 0, len(statuses))
	for _, st := range statuses {
		tc, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
			Title: "Case " + st, CaseType: "classic", Priority: "medium",
		})
		require.NoError(t, err)
		caseIDs = append(caseIDs, tc.ID)
	}

	run, err := svc.CreateTestRun(ctx, projectID, userID, tms.CreateTestRunRequest{Title: "Stats mix"})
	require.NoError(t, err)
	runID, _ := uuid.Parse(run.ID)

	detail, err := svc.AddTestCasesToRun(ctx, projectID, runID, tms.AddTestCasesToRunRequest{CaseIDs: caseIDs})
	require.NoError(t, err)
	require.Len(t, detail.Results, 5)

	_, err = svc.StartTestRun(ctx, projectID, runID)
	require.NoError(t, err)

	for i, st := range statuses {
		resultID, err := uuid.Parse(detail.Results[i].ID)
		require.NoError(t, err)
		status := st
		_, err = svc.ExecuteTestResult(ctx, projectID, runID, resultID, userID, tms.UpdateTestResultRequest{Status: &status})
		require.NoError(t, err)
	}

	got, err := svc.GetTestRunDetail(ctx, projectID, runID)
	require.NoError(t, err)
	assert.Equal(t, 5, got.TotalCount)
	assert.Equal(t, 1, got.PassedCount)
	assert.Equal(t, 1, got.FailedCount)
	assert.Equal(t, 1, got.BrokenCount)
	assert.Equal(t, 1, got.SkippedCount)
}

// TestMapTestRailPriorityAndType_AllBranches проверяет сценарий: map test rail priority and type all branches.
func TestMapTestRailPriorityAndType_AllBranches(t *testing.T) {
	assert.Equal(t, "critical", tms.ExportMapTestRailPriority("p1"))
	assert.Equal(t, "high", tms.ExportMapTestRailPriority("2 - High"))
	assert.Equal(t, "medium", tms.ExportMapTestRailPriority("p3"))
	assert.Equal(t, "low", tms.ExportMapTestRailPriority("4 - Low"))
	assert.Equal(t, "trivial", tms.ExportMapTestRailPriority("5 - Trivial"))

	assert.Equal(t, "classic", tms.ExportMapTestRailType("Automated"))
	assert.Equal(t, "classic", tms.ExportMapTestRailType("Functional"))
	assert.Equal(t, "classic", tms.ExportMapTestRailType("Regression"))
	assert.Equal(t, "classic", tms.ExportMapTestRailType("Smoke"))
	assert.Equal(t, "bdd", tms.ExportMapTestRailType("Scenario"))
}

// TestService_Wave122_CreateBugFromFailedResult_NoTestCase проверяет бизнес-логику: сервис wave 122 создание bug from failed result no test case.
func TestService_Wave122_CreateBugFromFailedResult_NoTestCase(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	runID := uuid.New()
	repo.runs[runID] = &ent.TestRun{
		ID: runID,
		Edges: ent.TestRunEdges{
			Project: &ent.Project{ID: projectID},
		},
	}

	resultID := uuid.New()
	repo.results[resultID] = &ent.TestResult{
		ID: resultID, Status: testresult.StatusFailed, ErrorMessage: "assertion failed", CreatedAt: time.Now(),
		Edges: ent.TestResultEdges{Run: repo.runs[runID]},
	}

	bug, err := svc.CreateBugFromFailedResult(ctx, projectID, runID, userID, resultID, tms.CreateBugRequest{})
	require.NoError(t, err)
	assert.Contains(t, bug.Title, "FAILED")
}

// TestService_Wave122_UnlinkTestCaseFromIssue_NotFound проверяет бизнес-логику: сервис wave 122 unlink test case from issue not found.
func TestService_Wave122_UnlinkTestCaseFromIssue_NotFound(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	err := svc.UnlinkTestCaseFromIssue(context.Background(), uuid.New(), "TC-MISSING", uuid.New())
	require.Error(t, err)
}

// TestService_Wave122_ImportAllureResults_Success проверяет бизнес-логику: сервис wave 122 импорт allure results успех.
func TestService_Wave122_ImportAllureResults_Success(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	detail, err := svc.ImportAllureResults(ctx, projectID, userID, tms.ImportAllureRequest{
		Title:   "Allure batch",
		Results: json.RawMessage(validAllureJSON),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, detail.ID)
	assert.NotEmpty(t, detail.Results)
}

// TestService_Wave122_GetTestRunHistory_WithRuns проверяет бизнес-логику: сервис wave 122 получение test run history with runs.
func TestService_Wave122_GetTestRunHistory_WithRuns(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	svc := tms.NewTMSService(tms.NewCombinedRepository(repo, &tms.MongoRepository{}), nil)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "History run", CreatedByID: userID,
	})
	require.NoError(t, err)
	_, err = repo.UpdateTestRunStats(ctx, run.ID, tms.TestRunStats{
		TotalCount: 10, PassedCount: 8, FailedCount: 2,
	})
	require.NoError(t, err)

	history, err := svc.GetTestRunHistory(ctx, projectID, 5)
	require.NoError(t, err)
	require.NotEmpty(t, history)
	assert.Equal(t, run.ID.String(), history[0].RunID)
	assert.InDelta(t, 80.0, history[0].PassRate, 0.1)
}

// TestService_Wave122_UpdateTestCase_InvalidLabels проверяет отклонение при сервис wave 122 обновление test case некорректный метки.
func TestService_Wave122_UpdateTestCase_InvalidLabels(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Labels", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)

	bad := []string{"not-uuid"}
	_, err = svc.UpdateTestCase(ctx, projectID, created.CaseKey, userID, tms.UpdateTestCaseRequest{Labels: &bad,
		Version: 1,})
	require.Error(t, err)
}

// TestService_Wave122_GetFlakyTests_Ent проверяет слой репозитория: сервис wave 122 получение flaky tests ent.
func TestService_Wave122_GetFlakyTests_Ent(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	svc := tms.NewTMSService(tms.NewCombinedRepository(repo, &tms.MongoRepository{}), nil)

	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Flaky case", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{Title: "Flaky run", CreatedByID: userID})
	require.NoError(t, err)

	_, err = repo.CreateTestResult(ctx, run.ID, tms.CreateTestResultInput{
		TestCaseID: &tc.ID, Status: "passed", IsFlaky: true,
	})
	require.NoError(t, err)

	report, err := svc.GetFlakyTests(ctx, projectID)
	require.NoError(t, err)
	assert.NotNil(t, report)
}
