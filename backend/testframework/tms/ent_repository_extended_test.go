package tms_test

// Тесты репозитория TMS: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	entissue "github.com/valerii/hub/backend/ent/issue"
	entproject "github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// TestEntRepository_TestCaseVersions проверяет слой репозитория: ent репозиторий test case versions.
func TestEntRepository_TestCaseVersions(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Versioned", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	snapshot := map[string]interface{}{"title": "Versioned", "priority": "medium"}
	v1, err := repo.CreateTestCaseVersion(ctx, tc.ID, userID, 1, snapshot, "initial")
	require.NoError(t, err)
	assert.Equal(t, 1, v1.VersionNumber)

	versions, err := repo.ListTestCaseVersions(ctx, tc.ID)
	require.NoError(t, err)
	require.Len(t, versions, 1)

	got, err := repo.GetTestCaseVersion(ctx, v1.ID)
	require.NoError(t, err)
	assert.Equal(t, v1.ID, got.ID)
}

// TestEntRepository_TestResultsCRUD проверяет слой репозитория: ent репозиторий test results crud.
func TestEntRepository_TestResultsCRUD(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Results run", RunType: "manual", CreatedByID: userID,
	})
	require.NoError(t, err)

	duration := int64(500)
	created, err := repo.CreateTestResult(ctx, run.ID, tms.CreateTestResultInput{
		Status: "failed", Comment: "assertion", DurationMs: &duration, IsFlaky: true,
	})
	require.NoError(t, err)

	passed := "passed"
	updated, err := repo.UpdateTestResult(ctx, run.ID, created.ID, tms.UpdateTestResultInput{
		Status: &passed, Comment: strPtr("fixed"),
	})
	require.NoError(t, err)
	assert.Equal(t, "passed", string(updated.Status))

	results, err := repo.ListTestResults(ctx, run.ID)
	require.NoError(t, err)
	require.Len(t, results, 1)

	got, err := repo.GetTestResult(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}

// TestEntRepository_AddTestCasesToRunAndAssign проверяет слой репозитория: ent репозиторий add test cases to run and assign.
func TestEntRepository_AddTestCasesToRunAndAssign(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Run case", CaseType: "classic", Priority: "low", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Assign run", CreatedByID: userID,
	})
	require.NoError(t, err)

	require.NoError(t, repo.AddTestCasesToRun(ctx, run.ID, []uuid.UUID{tc.ID}))

	results, err := repo.ListTestRunResults(ctx, run.ID)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.NotNil(t, results[0].Edges.TestCase)
	assert.Equal(t, tc.ID, results[0].Edges.TestCase.ID)

	tester, err := tms.ExportEntRepositoryClient(repo).User.Create().
		SetEmail("tester-" + uuid.New().String()[:8] + "@test.com").
		SetName("Tester").
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, repo.AssignTestersToRun(ctx, run.ID, []uuid.UUID{tester.ID}))

	withResults, err := repo.GetTestRunWithResults(ctx, run.ID)
	require.NoError(t, err)
	require.Len(t, withResults.Edges.Results, 1)
	require.Len(t, withResults.Edges.Assignees, 1)
}

// TestEntRepository_Environments проверяет слой репозитория: ent репозиторий environments.
func TestEntRepository_Environments(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)

	env, err := repo.CreateEnvironment(ctx, projectID, tms.CreateEnvironmentInput{
		Name: "Staging", Label: "staging", Config: map[string]interface{}{"url": "https://stg.example"},
	})
	require.NoError(t, err)

	got, err := repo.GetEnvironment(ctx, env.ID)
	require.NoError(t, err)
	assert.Equal(t, "Staging", got.Name)

	envs, err := repo.ListEnvironmentsByProject(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, envs, 1)

	renamed := "Staging v2"
	updated, err := repo.UpdateEnvironment(ctx, env.ID, tms.UpdateEnvironmentInput{Name: &renamed})
	require.NoError(t, err)
	assert.Equal(t, renamed, updated.Name)

	require.NoError(t, repo.DeleteEnvironment(ctx, env.ID))
	_, err = repo.GetEnvironment(ctx, env.ID)
	require.Error(t, err)
}

// TestEntRepository_LinkTestCaseAndCreateBug проверяет слой репозитория: ent репозиторий link test case and создание bug.
func TestEntRepository_LinkTestCaseAndCreateBug(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Linked case", CaseType: "classic", Priority: "high", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	statusID := mustDefaultStatusID(t, repo, ctx, projectID)
	issue, err := tms.ExportEntRepositoryClient(repo).Issue.Create().
		SetIssueNumber(100).
		SetIssueKey("TMS-100").
		SetTitle("Existing issue").
		SetIssueType(entissue.IssueTypeTask).
		SetPriority(entissue.PriorityMedium).
		SetProjectID(projectID).
		SetReporterID(userID).
		SetStatusID(statusID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, repo.LinkTestCaseToIssue(ctx, tc.ID, issue.ID))
	linked, err := repo.GetTestCaseLinkedIssues(ctx, tc.ID)
	require.NoError(t, err)
	require.Len(t, linked, 1)

	byIssue, err := repo.GetIssueLinkedTestCases(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, byIssue, 1)

	require.NoError(t, repo.UnlinkTestCaseFromIssue(ctx, tc.ID, issue.ID))
	linked, err = repo.GetTestCaseLinkedIssues(ctx, tc.ID)
	require.NoError(t, err)
	assert.Empty(t, linked)

	bug, err := repo.CreateBugIssue(ctx, projectID, userID, tc.ID, "Bug from test", "steps to repro", "critical")
	require.NoError(t, err)
	assert.Equal(t, entissue.IssueTypeBug, bug.IssueType)

	linked, err = repo.GetTestCaseLinkedIssues(ctx, tc.ID)
	require.NoError(t, err)
	require.Len(t, linked, 1)
}

// TestEntRepository_ReopenTestRun проверяет слой репозитория: ent репозиторий reopen test run.
func TestEntRepository_ReopenTestRun(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Reopen run", CreatedByID: userID,
	})
	require.NoError(t, err)

	now := time.Now()
	_, err = repo.UpdateTestRunStatus(ctx, run.ID, "completed", &now)
	require.NoError(t, err)

	reopened, err := repo.ReopenTestRun(ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, "in_progress", string(reopened.Status))
}

func strPtr(s string) *string { return &s }

func mustDefaultStatusID(t *testing.T, repo *tms.EntRepository, ctx context.Context, projectID uuid.UUID) uuid.UUID {
	t.Helper()
	st, err := tms.ExportEntRepositoryClient(repo).WorkflowStatus.Query().
		Where(
			workflowstatus.HasProjectWith(entproject.ID(projectID)),
			workflowstatus.IsDefault(true),
		).
		Only(ctx)
	require.NoError(t, err)
	return st.ID
}

// TestEntRepository_SuiteUpdateDeleteMove проверяет слой репозитория: ent репозиторий suite обновление удаление move.
func TestEntRepository_SuiteUpdateDeleteMove(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)

	parent, err := repo.CreateSuite(ctx, projectID, tms.CreateSuiteInput{Name: "Parent"})
	require.NoError(t, err)
	child, err := repo.CreateSuite(ctx, projectID, tms.CreateSuiteInput{Name: "Child", ParentID: &parent.ID})
	require.NoError(t, err)

	renamed := "Child renamed"
	updated, err := repo.UpdateSuite(ctx, child.ID, tms.UpdateSuiteInput{Name: &renamed})
	require.NoError(t, err)
	assert.Equal(t, renamed, updated.Name)

	moved, err := repo.MoveSuite(ctx, child.ID, nil)
	require.NoError(t, err)
	assert.Nil(t, moved.Edges.Parent)

	require.NoError(t, repo.DeleteSuite(ctx, child.ID))
	_, err = repo.GetSuite(ctx, child.ID)
	require.Error(t, err)
}

// TestEntRepository_BatchCreateTestResults проверяет слой репозитория: ent репозиторий batch создание test results.
func TestEntRepository_BatchCreateTestResults(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Batch run", CreatedByID: userID,
	})
	require.NoError(t, err)

	results, err := repo.BatchCreateTestResults(ctx, run.ID, []tms.CreateTestResultInput{
		{Status: "passed", Comment: "ok"},
		{Status: "failed", Comment: "nope"},
	})
	require.NoError(t, err)
	require.Len(t, results, 2)

	listed, err := repo.ListTestResults(ctx, run.ID)
	require.NoError(t, err)
	assert.Len(t, listed, 2)
}

// TestEntRepository_DeleteComponent проверяет слой репозитория: ent репозиторий удаление component.
func TestEntRepository_DeleteComponent(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)

	comp, err := repo.CreateComponent(ctx, projectID, "UI", "frontend")
	require.NoError(t, err)
	require.NoError(t, repo.DeleteComponent(ctx, comp.ID))
}
