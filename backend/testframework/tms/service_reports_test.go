package tms_test

// Тесты сервиса TMS: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/testrun"
)

// TestListTestCases_Success проверяет сценарий: список test cases успех.
func TestListTestCases_Success(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	_, err := svc.CreateTestCase(ctx, projectID, uuid.New(), tms.CreateTestCaseRequest{Title: "A"})
	require.NoError(t, err)
	_, err = svc.CreateTestCase(ctx, projectID, uuid.New(), tms.CreateTestCaseRequest{Title: "B"})
	require.NoError(t, err)

	resp, err := svc.ListTestCases(ctx, projectID, tms.ListTestCasesRequest{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Items, 2)
}

// TestDeleteTestCase проверяет сценарий: удаление test case.
func TestDeleteTestCase(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	created, err := svc.CreateTestCase(ctx, projectID, uuid.New(), tms.CreateTestCaseRequest{Title: "Delete me"})
	require.NoError(t, err)

	require.NoError(t, svc.DeleteTestCase(ctx, projectID, created.CaseKey))
	_, err = svc.GetTestCase(ctx, projectID, created.CaseKey)
	require.Error(t, err)
}

// TestListTestCaseVersions проверяет сценарий: список test case versions.
func TestListTestCaseVersions(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	created, err := svc.CreateTestCase(ctx, projectID, uuid.New(), tms.CreateTestCaseRequest{Title: "Versioned"})
	require.NoError(t, err)

	newTitle := "v2"
	_, err = svc.UpdateTestCase(ctx, projectID, created.CaseKey, uuid.New(), tms.UpdateTestCaseRequest{
		Title:   &newTitle,
		Version: 1,
	})
	require.NoError(t, err)

	versions, err := svc.ListTestCaseVersions(ctx, projectID, created.CaseKey)
	require.NoError(t, err)
	assert.Len(t, versions, 1)
}

// TestGetTestTrends проверяет сценарий: получение test trends.
func TestGetTestTrends(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	runID := uuid.New()
	repo.runs[runID] = &ent.TestRun{
		ID:          runID,
		Title:       "Trend run",
		Status:      testrun.StatusCompleted,
		TotalCount:  10,
		PassedCount: 8,
		FailedCount: 2,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	trends, err := svc.GetTestTrends(ctx, projectID, 5)
	require.NoError(t, err)
	require.Len(t, trends, 1)
	assert.Equal(t, 80.0, trends[0].PassRate)
}

// TestGetTestTrends_DefaultLimit проверяет сценарий: получение test trends default limit.
func TestGetTestTrends_DefaultLimit(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	trends, err := svc.GetTestTrends(context.Background(), uuid.New(), 0)
	require.NoError(t, err)
	assert.Empty(t, trends)
}

// TestParseUUIDs проверяет сценарий: разбор uui ds.
func TestParseUUIDs(t *testing.T) {
	id := uuid.New()
	ids, err := tms.ExportParseUUIDs([]string{id.String()})
	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{id}, ids)

	_, err = tms.ExportParseUUIDs([]string{"bad"})
	require.Error(t, err)
}

// TestGetTraceabilityMatrix проверяет сценарий: получение traceability matrix.
func TestGetTraceabilityMatrix(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	issueID := uuid.New()
	tcID := uuid.New()
	repo.issues[issueID] = &ent.Issue{ID: issueID, IssueKey: "HUB-10", Title: "Story"}
	repo.testCases[tcID] = &ent.TestCase{ID: tcID, CaseKey: "HUB-TC-10", Title: "Check"}
	repo.linkedIssues[tcID] = []uuid.UUID{issueID}
	repo.issues[issueID].Edges.LinkedTestCases = []*ent.TestCase{repo.testCases[tcID]}

	matrix, err := svc.GetTraceabilityMatrix(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, matrix.Rows, 1)
	assert.Equal(t, "HUB-10", matrix.Rows[0].IssueKey)
}

// TestGetDefectCategories проверяет сценарий: получение defect categories.
func TestGetDefectCategories(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	runID := uuid.New()
	resultID := uuid.New()
	tcID := uuid.New()
	repo.runs[runID] = &ent.TestRun{ID: runID, Title: "Run", CreatedAt: time.Now()}
	repo.testCases[tcID] = &ent.TestCase{ID: tcID, CaseKey: "HUB-TC-99", Title: "Fail case"}
	repo.results[resultID] = &ent.TestResult{
		ID: resultID, Status: "failed", ErrorMessage: "NullPointerException at line 1",
		Edges: ent.TestResultEdges{TestCase: repo.testCases[tcID]}, CreatedAt: time.Now(),
	}

	categories, err := svc.GetDefectCategories(ctx, projectID)
	require.NoError(t, err)
	require.NotEmpty(t, categories)
}

// TestGetRunComparison_StatusChanges проверяет сценарий: получение run comparison status changes.
func TestGetRunComparison_StatusChanges(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()

	tcID := uuid.New()
	repo.testCases[tcID] = &ent.TestCase{ID: tcID, CaseKey: "HUB-TC-50", Title: "Compare case"}

	run1ID := uuid.New()
	run2ID := uuid.New()
	now := time.Now()
	repo.runs[run1ID] = &ent.TestRun{
		ID: run1ID, Title: "Run A", TotalCount: 1, PassedCount: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	repo.runs[run2ID] = &ent.TestRun{
		ID: run2ID, Title: "Run B", TotalCount: 1, FailedCount: 1,
		CreatedAt: now.Add(time.Hour), UpdatedAt: now.Add(time.Hour),
	}

	r1 := uuid.New()
	r2 := uuid.New()
	repo.results[r1] = &ent.TestResult{
		ID: r1, Status: "passed",
		Edges: ent.TestResultEdges{TestCase: repo.testCases[tcID]},
	}
	repo.results[r2] = &ent.TestResult{
		ID: r2, Status: "failed",
		Edges: ent.TestResultEdges{TestCase: repo.testCases[tcID]},
	}
	repo.runResults = map[uuid.UUID][]uuid.UUID{
		run1ID: {r1},
		run2ID: {r2},
	}

	cmp, err := svc.GetRunComparison(ctx, run1ID, run2ID)
	require.NoError(t, err)
	require.Len(t, cmp.Changes, 1)
	assert.Equal(t, "HUB-TC-50", cmp.Changes[0].CaseKey)
	assert.Equal(t, "passed", cmp.Changes[0].StatusRun1)
	assert.Equal(t, "failed", cmp.Changes[0].StatusRun2)
}

// TestGetFlakyTests_MixedStatuses проверяет сценарий: получение flaky tests mixed statuses.
func TestGetFlakyTests_MixedStatuses(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	tcID := uuid.New()
	repo.testCases[tcID] = &ent.TestCase{
		ID: tcID, CaseKey: "HUB-TC-77", Title: "Flaky case",
	}

	run1ID := uuid.New()
	run2ID := uuid.New()
	now := time.Now()
	repo.runs[run1ID] = &ent.TestRun{ID: run1ID, Title: "Run 1", CreatedAt: now}
	repo.runs[run2ID] = &ent.TestRun{ID: run2ID, Title: "Run 2", CreatedAt: now.Add(time.Hour)}

	r1 := uuid.New()
	r2 := uuid.New()
	repo.results[r1] = &ent.TestResult{
		ID: r1, Status: "passed",
		Edges: ent.TestResultEdges{TestCase: repo.testCases[tcID]},
	}
	repo.results[r2] = &ent.TestResult{
		ID: r2, Status: "failed", IsFlaky: true,
		Edges: ent.TestResultEdges{TestCase: repo.testCases[tcID]},
	}
	repo.runResults[run1ID] = []uuid.UUID{r1}
	repo.runResults[run2ID] = []uuid.UUID{r2}

	report, err := svc.GetFlakyTests(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, report.Items, 1)
	assert.Equal(t, "HUB-TC-77", report.Items[0].CaseKey)
	assert.True(t, report.Items[0].MarkedFlaky)
}

// TestCompareEnvironments проверяет сценарий: compare environments.
func TestCompareEnvironments(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	env1ID := uuid.New()
	env2ID := uuid.New()
	repo.environments[env1ID] = &ent.Environment{
		ID: env1ID, Name: "Dev", EnvLabel: "dev",
		Edges: ent.EnvironmentEdges{Project: &ent.Project{ID: projectID}},
	}
	repo.environments[env2ID] = &ent.Environment{
		ID: env2ID, Name: "Prod", EnvLabel: "prod",
		Edges: ent.EnvironmentEdges{Project: &ent.Project{ID: projectID}},
	}

	cmp, err := svc.CompareEnvironments(ctx, projectID, env1ID, env2ID)
	require.NoError(t, err)
	assert.Equal(t, "Dev", cmp.Env1.Name)
	assert.Equal(t, "Prod", cmp.Env2.Name)
}
