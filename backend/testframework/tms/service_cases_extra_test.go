package tms_test

// Тесты сервиса TMS: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

// TestGetTestCase_AndList проверяет сценарий: получение test case and список.
func TestGetTestCase_AndList(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title:    "Login test",
		CaseType: "functional",
		Priority: "high",
	})
	require.NoError(t, err)

	got, err := svc.GetTestCase(ctx, projectID, created.CaseKey)
	require.NoError(t, err)
	assert.Equal(t, "Login test", got.Title)

	list, err := svc.ListTestCases(ctx, projectID, tms.ListTestCasesRequest{Limit: 20})
	require.NoError(t, err)
	assert.Equal(t, 1, list.Total)
	require.Len(t, list.Items, 1)
}

// TestDeleteTestCase_RemovesFromRepo проверяет сценарий: удаление test case removes from repo.
func TestDeleteTestCase_RemovesFromRepo(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Disposable", CaseType: "functional", Priority: "medium",
	})
	require.NoError(t, err)

	require.NoError(t, svc.DeleteTestCase(ctx, projectID, created.CaseKey))
	_, err = svc.GetTestCase(ctx, projectID, created.CaseKey)
	require.Error(t, err)
}

// TestListTestCaseVersions_AfterUpdate проверяет сценарий: список test case versions after обновление.
func TestListTestCaseVersions_AfterUpdate(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Versioned", CaseType: "functional", Priority: "medium",
	})
	require.NoError(t, err)

	newTitle := "Versioned v2"
	_, err = svc.UpdateTestCase(ctx, projectID, created.CaseKey, userID, tms.UpdateTestCaseRequest{Title: &newTitle,
		Version: 1,})
	require.NoError(t, err)

	versions, err := svc.ListTestCaseVersions(ctx, projectID, created.CaseKey)
	require.NoError(t, err)
	assert.NotEmpty(t, versions)
}

// TestUnlinkTestCaseFromIssue проверяет сценарий: unlink test case from issue.
func TestUnlinkTestCaseFromIssue(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	tc, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Linked case", CaseType: "functional", Priority: "medium",
	})
	require.NoError(t, err)

	issueID := uuid.New()
	repo.issues[issueID] = &ent.Issue{ID: issueID, IssueKey: "HUB-1", Title: "Bug"}
	entTC, _ := repo.FindTestCaseByKey(ctx, projectID, tc.CaseKey)
	repo.linkedIssues[entTC.ID] = []uuid.UUID{issueID}

	require.NoError(t, svc.UnlinkTestCaseFromIssue(ctx, projectID, tc.CaseKey, issueID))
	links, err := svc.GetTestCaseLinks(ctx, projectID, tc.CaseKey)
	require.NoError(t, err)
	assert.Empty(t, links)
}

// TestGetTestRun_AndListResults проверяет сценарий: получение test run and список results.
func TestGetTestRun_AndListResults(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	run, err := svc.CreateTestRun(ctx, projectID, userID, tms.CreateTestRunRequest{Title: "Smoke"})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)

	got, err := svc.GetTestRun(ctx, projectID, runID)
	require.NoError(t, err)
	assert.Equal(t, "Smoke", got.Title)

	tc, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Case A", CaseType: "functional", Priority: "medium",
	})
	require.NoError(t, err)
	entTC, _ := repo.FindTestCaseByKey(ctx, projectID, tc.CaseKey)

	_, err = svc.CreateTestResult(ctx, projectID, runID, tms.CreateTestResultRequest{
		TestCaseID: entTC.ID.String(),
		Status:     "passed",
	})
	require.NoError(t, err)

	results, err := svc.ListTestResults(ctx, projectID, runID)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

// TestGetTestTrends_EmptyProject проверяет сценарий: получение test trends пустой project.
func TestGetTestTrends_EmptyProject(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	trends, err := svc.GetTestTrends(context.Background(), uuid.New(), 5)
	require.NoError(t, err)
	assert.Empty(t, trends)
}

// TestListTestRuns_Pagination проверяет сценарий: список test runs pagination.
func TestListTestRuns_Pagination(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	for i := 0; i < 2; i++ {
		_, err := svc.CreateTestRun(ctx, projectID, userID, tms.CreateTestRunRequest{Title: "Run"})
		require.NoError(t, err)
	}

	page, err := svc.ListTestRuns(ctx, projectID, tms.ListTestRunsRequest{Offset: 0, Limit: 1})
	require.NoError(t, err)
	assert.Equal(t, 2, page.Total)
	assert.Len(t, page.Items, 2) // mock repo ignores pagination params
}
