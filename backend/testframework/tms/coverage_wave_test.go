package tms_test

// Тесты модуля TMS: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/filemetadata"
	entissue "github.com/valerii/hub/backend/ent/issue"
)

// TestMapTestRailPriorityAndType проверяет сценарий: map test rail priority and type.
func TestMapTestRailPriorityAndType(t *testing.T) {
	assert.Equal(t, "critical", tms.ExportMapTestRailPriority("1 - Critical"))
	assert.Equal(t, "high", tms.ExportMapTestRailPriority("High"))
	assert.Equal(t, "medium", tms.ExportMapTestRailPriority("unknown"))
	assert.Equal(t, "classic", tms.ExportMapTestRailType("Automation"))
	assert.Equal(t, "classic", tms.ExportMapTestRailType("Manual"))
	assert.Equal(t, "bdd", tms.ExportMapTestRailType("BDD"))
}

// TestStripHTMLAndParseSteps проверяет сценарий: strip html and разбор steps.
func TestStripHTMLAndParseSteps(t *testing.T) {
	assert.Equal(t, "Hello world", tms.ExportStripHTML("<p>Hello <b>world</b></p>"))
	assert.Nil(t, tms.ExportParseTestRailSteps("", ""))

	steps := tms.ExportParseTestRailSteps("1. Open app\n2. Click save", "OK\nSaved")
	require.Len(t, steps, 2)
	assert.Equal(t, 1, steps[0].Order)
	assert.Equal(t, "Open app", steps[0].Action)
}

// TestTrimStepNumbering проверяет сценарий: trim step numbering.
func TestTrimStepNumbering(t *testing.T) {
	assert.Equal(t, "Save file", tms.ExportTrimStepNumbering("Step 1: Save file"))
	assert.Equal(t, "Click", tms.ExportTrimStepNumbering("2) Click"))
}

// TestCSVField проверяет сценарий: csv field.
func TestCSVField(t *testing.T) {
	row := []string{"a", "b", ""}
	colMap := map[string]int{"id": 0, "title": 1}
	assert.Equal(t, "a", tms.ExportCSVField(row, colMap, "id"))
	assert.Equal(t, "", tms.ExportCSVField(row, colMap, "missing"))
}

// TestCombinedRepositoryConstruct проверяет слой репозитория: combined репозиторий construct.
func TestCombinedRepositoryConstruct(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)

	combined := tms.NewCombinedRepository(repo, &tms.MongoRepository{})
	require.NotNil(t, combined)

	prov := tms.NewTMSProjectProvisioner(combined)
	require.NoError(t, prov.CreateDefaultSuite(ctx, projectID))

	suites, err := repo.ListSuites(ctx, projectID)
	require.NoError(t, err)
	require.NotEmpty(t, suites)
	assert.Equal(t, "Общие тесты", suites[0].Name)
}

// TestEntRepository_TraceabilityAndResults проверяет слой репозитория: ent репозиторий traceability and results.
func TestEntRepository_TraceabilityAndResults(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Trace case", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Result run", CreatedByID: userID,
	})
	require.NoError(t, err)

	result, err := repo.CreateTestResult(ctx, run.ID, tms.CreateTestResultInput{
		TestCaseID: &tc.ID, Status: "passed", Comment: "ok",
	})
	require.NoError(t, err)

	last, err := repo.GetLastResultForTestCase(ctx, tc.ID)
	require.NoError(t, err)
	assert.Equal(t, result.ID, last.ID)

	statusID := mustDefaultStatusID(t, repo, ctx, projectID)
	issue, err := tms.ExportEntRepositoryClient(repo).Issue.Create().
		SetIssueNumber(1).
		SetIssueKey("TRC-1").
		SetTitle("Linked").
		SetIssueType(entissue.IssueTypeTask).
		SetProjectID(projectID).
		SetReporterID(userID).
		SetStatusID(statusID).
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.LinkTestCaseToIssue(ctx, tc.ID, issue.ID))

	issues, err := repo.GetProjectIssuesWithTestCases(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.NotEmpty(t, issues[0].Edges.LinkedTestCases)

	fileID := uuid.New()
	_, err = tms.ExportEntRepositoryClient(repo).FileMetadata.Create().
		SetID(fileID).
		SetFilename("shot.png").
		SetContentType("image/png").
		SetSize(100).
		SetS3Key("k").
		SetS3Bucket("b").
		SetStatus(filemetadata.StatusConfirmed).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.AddFilesToResult(ctx, result.ID, []uuid.UUID{fileID}))

	env, err := repo.CreateEnvironment(ctx, projectID, tms.CreateEnvironmentInput{
		Name: "Prod", Label: "prod",
	})
	require.NoError(t, err)

	_, err = tms.ExportEntRepositoryClient(repo).TestRun.Create().
		SetTitle("Env bound").
		SetProjectID(projectID).
		SetCreatedByID(userID).
		SetEnvironmentID(env.ID).
		Save(ctx)
	require.NoError(t, err)

	latest, err := repo.GetLatestTestRunForEnvironment(ctx, projectID, env.ID)
	require.NoError(t, err)
	require.NotNil(t, latest)
}
