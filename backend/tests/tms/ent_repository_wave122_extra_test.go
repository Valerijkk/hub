package tms_test

// Тесты репозитория TMS (волна 122): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/schema"
	entissue "github.com/valerii/hub/backend/ent/issue"
)

// Фаза 122 — ent-репозиторий: AddTestCasesToRun, CreateBugIssue..

func TestEntRepository_Wave122_AddTestCasesToRun(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "With cases", CreatedByID: userID,
	})
	require.NoError(t, err)

	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Runnable", CaseType: "classic", Priority: "high", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	require.NoError(t, repo.AddTestCasesToRun(ctx, run.ID, []uuid.UUID{tc.ID}))

	results, err := repo.ListTestResults(ctx, run.ID)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "unknown", string(results[0].Status))
}

// TestEntRepository_Wave122_CreateBugIssue проверяет слой репозитория: ent репозиторий wave 122 создание bug issue.
func TestEntRepository_Wave122_CreateBugIssue(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Bug source", CaseType: "classic", Priority: "high", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	bug, err := repo.CreateBugIssue(ctx, projectID, userID, tc.ID, "Login fails", "steps", "critical")
	require.NoError(t, err)
	assert.Equal(t, entissue.IssueTypeBug, bug.IssueType)
	assert.Equal(t, entissue.PriorityCritical, bug.Priority)

	linked, err := repo.GetTestCaseLinkedIssues(ctx, tc.ID)
	require.NoError(t, err)
	require.Len(t, linked, 1)
	assert.Equal(t, bug.ID, linked[0].ID)
}

// TestEntRepository_Wave122_CreateBugIssue_DefaultPriority проверяет слой репозитория: ent репозиторий wave 122 создание bug issue default priority.
func TestEntRepository_Wave122_CreateBugIssue_DefaultPriority(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Bug default pri", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	bug, err := repo.CreateBugIssue(ctx, projectID, userID, tc.ID, "Minor glitch", "", "")
	require.NoError(t, err)
	assert.Equal(t, entissue.PriorityHigh, bug.Priority)
}

// TestEntRepository_Wave122_CreateTestCase_AllFields проверяет слой репозитория: ent репозиторий wave 122 создание test case all fields.
func TestEntRepository_Wave122_CreateTestCase_AllFields(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	suite, err := repo.CreateSuite(ctx, projectID, tms.CreateSuiteInput{Name: "Suite A"})
	require.NoError(t, err)
	comp, err := repo.CreateComponent(ctx, projectID, "API", "backend")
	require.NoError(t, err)

	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title:         "Full case",
		Description:   "desc",
		Precondition:  "pre",
		Postcondition: "post",
		CaseType:      "bdd",
		Priority:      "critical",
		Status:        "draft",
		FullName:      "com.app.Test",
		ClassName:     "Test",
		MethodName:    "run",
		ExternalID:    "EXT-1",
		AutomationID:  "AUTO-1",
		SuiteID:       &suite.ID,
		ComponentIDs:  []uuid.UUID{comp.ID},
		CreatedByID:   userID,
		Steps: []schema.TestStep{
			{Order: 1, Action: "click", ExpectedResult: "ok"},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "Full case", tc.Title)
	assert.Equal(t, "com.app.Test", tc.FullName)
}
