package tms_test

// Тесты репозитория TMS: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	entissue "github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/schema"
	"github.com/valerii/hub/backend/ent/testresult"
)

// TestEntRepository_BatchCreateTestResults_AllOptionalFields проверяет слой репозитория: ent репозиторий batch создание test results all optional fields.
func TestEntRepository_BatchCreateTestResults_AllOptionalFields(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Batch run", CreatedByID: userID,
	})
	require.NoError(t, err)

	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Batch case", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	now := time.Now().UTC()
	duration := int64(1200)
	steps := []schema.StepResult{{StepOrder: 1, Status: "passed", Comment: "ok"}}

	results, err := repo.BatchCreateTestResults(ctx, run.ID, []tms.CreateTestResultInput{
		{
			TestCaseID:    &tc.ID,
			Status:        "failed",
			Comment:       "assertion",
			StepResults:   steps,
			DurationMs:    &duration,
			AllureUUID:    "allure-1",
			ErrorMessage:  "boom",
			ErrorTrace:    "trace",
			IsFlaky:       true,
			MatchStrategy: string(testresult.MatchStrategyFullname),
			ExecutedByID:  &userID,
			StartedAt:     &now,
			CompletedAt:   &now,
		},
		{Status: "skipped", Comment: "ignored"},
	})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "failed", string(results[0].Status))
	assert.True(t, results[0].IsFlaky)

	listed, err := repo.ListTestResults(ctx, run.ID)
	require.NoError(t, err)
	assert.Len(t, listed, 2)
}

// TestEntRepository_ListTestRuns_StatusFilter проверяет слой репозитория: ent репозиторий список test runs status filter.
func TestEntRepository_ListTestRuns_StatusFilter(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	_, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Open run", CreatedByID: userID,
	})
	require.NoError(t, err)

	completed, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Done run", CreatedByID: userID,
	})
	require.NoError(t, err)
	_, err = repo.UpdateTestRunStatus(ctx, completed.ID, "completed", nil)
	require.NoError(t, err)

	items, total, err := repo.ListTestRuns(ctx, projectID, tms.ListTestRunsParams{
		Status: "completed", Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, "Done run", items[0].Title)
}

// TestTMSService_GetCoverageReport_Ent проверяет слой репозитория: tms сервис получение покрытие report ent.
func TestTMSService_GetCoverageReport_Ent(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	svc := tms.NewTMSService(tms.NewCombinedRepository(repo, &tms.MongoRepository{}), nil)

	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Cov case", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{Title: "Cov run", CreatedByID: userID})
	require.NoError(t, err)
	_, err = repo.CreateTestResult(ctx, run.ID, tms.CreateTestResultInput{
		TestCaseID: &tc.ID, Status: "passed",
	})
	require.NoError(t, err)

	statusID := mustDefaultStatusID(t, repo, ctx, projectID)
	issue, err := tms.ExportEntRepositoryClient(repo).Issue.Create().
		SetIssueNumber(10).
		SetIssueKey("COV-10").
		SetTitle("Coverage story").
		SetIssueType(entissue.IssueTypeTask).
		SetProjectID(projectID).
		SetReporterID(userID).
		SetStatusID(statusID).
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.LinkTestCaseToIssue(ctx, tc.ID, issue.ID))

	report, err := svc.GetCoverageReport(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, report.Issues, 1)
	assert.Equal(t, "COV-10", report.Issues[0].IssueKey)
	assert.Equal(t, "passed", report.Issues[0].TestCases[0].LastStatus)
}
