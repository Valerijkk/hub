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
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func newEntRepositoryTestEnv(t *testing.T) (*tms.EntRepository, context.Context, uuid.UUID, uuid.UUID) {
	t.Helper()
	dsn := "file:tmstest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("tms-" + uuid.New().String()[:8] + "@test.com").
		SetName("TMS Tester").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("TMS Project").
		SetKey("TMS").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	return tms.NewEntRepository(client), ctx, proj.ID, user.ID
}

// TestEntRepository_CreateAndGetTestCase проверяет слой репозитория: ent репозиторий создание and получение test case.
func TestEntRepository_CreateAndGetTestCase(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	created, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title:        "Ent repo case",
		CaseType:     "classic",
		Priority:     "high",
		Status:       "draft",
		ExternalID:   "ext-1",
		AutomationID: "auto-1",
		FullName:     "com.example.LoginTest",
		ClassName:    "LoginTest",
		MethodName:   "shouldLogin",
		CreatedByID:  userID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.CaseKey)
	assert.Equal(t, "Ent repo case", created.Title)

	got, err := repo.GetTestCase(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	byKey, err := repo.FindTestCaseByKey(ctx, projectID, created.CaseKey)
	require.NoError(t, err)
	assert.Equal(t, created.ID, byKey.ID)

	byExt, err := repo.FindTestCaseByExternalID(ctx, projectID, "ext-1")
	require.NoError(t, err)
	assert.Equal(t, created.ID, byExt.ID)

	byAuto, err := repo.FindTestCaseByAutomationID(ctx, projectID, "auto-1")
	require.NoError(t, err)
	assert.Equal(t, created.ID, byAuto.ID)

	byFull, err := repo.FindTestCaseByFullName(ctx, projectID, "com.example.LoginTest")
	require.NoError(t, err)
	assert.Equal(t, created.ID, byFull.ID)

	byClass, err := repo.FindTestCaseByClassMethod(ctx, projectID, "LoginTest", "shouldLogin")
	require.NoError(t, err)
	assert.Equal(t, created.ID, byClass.ID)
}

// TestEntRepository_ListUpdateDeleteTestCase проверяет слой репозитория: ent репозиторий список обновление удаление test case.
func TestEntRepository_ListUpdateDeleteTestCase(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	created, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title:       "Listed case",
		CaseType:    "classic",
		Priority:    "medium",
		Status:      "draft",
		CreatedByID: userID,
	})
	require.NoError(t, err)

	items, total, err := repo.ListTestCases(ctx, projectID, tms.ListTestCasesParams{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)

	newTitle := "Updated title"
	updated, err := repo.UpdateTestCase(ctx, created.ID, tms.UpdateTestCaseInput{Title: &newTitle, Version: 1})
	require.NoError(t, err)
	assert.Equal(t, newTitle, updated.Title)

	require.NoError(t, repo.DeleteTestCase(ctx, projectID, created.ID))
	_, err = repo.GetTestCase(ctx, created.ID)
	require.Error(t, err)
}

// TestEntRepository_IncrementTestCaseCounter проверяет слой репозитория: ent репозиторий increment test case counter.
func TestEntRepository_IncrementTestCaseCounter(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)

	n1, err := repo.IncrementTestCaseCounter(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, 1, n1)

	n2, err := repo.IncrementTestCaseCounter(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, 2, n2)
}

// TestEntRepository_CreateSuiteAndComponent проверяет слой репозитория: ent репозиторий создание suite and component.
func TestEntRepository_CreateSuiteAndComponent(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)

	suite, err := repo.CreateSuite(ctx, projectID, tms.CreateSuiteInput{Name: "Regression"})
	require.NoError(t, err)
	assert.Equal(t, "Regression", suite.Name)

	suites, err := repo.ListSuites(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, suites, 1)

	comp, err := repo.CreateComponent(ctx, projectID, "Backend", "api layer")
	require.NoError(t, err)
	assert.Equal(t, "Backend", comp.Name)

	comps, err := repo.ListComponents(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, comps, 1)
}

// TestEntRepository_TestRunLifecycle проверяет слой репозитория: ent репозиторий test run жизненный цикл.
func TestEntRepository_TestRunLifecycle(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title:       "Ent run",
		RunType:     "manual",
		CreatedByID: userID,
	})
	require.NoError(t, err)
	assert.Equal(t, "Ent run", run.Title)

	got, err := repo.GetTestRun(ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, run.ID, got.ID)

	items, total, err := repo.ListTestRuns(ctx, projectID, tms.ListTestRunsParams{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)

	now := time.Now()
	updated, err := repo.UpdateTestRunStatus(ctx, run.ID, "in_progress", &now)
	require.NoError(t, err)
	assert.Equal(t, "in_progress", string(updated.Status))

	stats, err := repo.UpdateTestRunStats(ctx, run.ID, tms.TestRunStats{TotalCount: 5, PassedCount: 4, FailedCount: 1})
	require.NoError(t, err)
	assert.Equal(t, 5, stats.TotalCount)
}
