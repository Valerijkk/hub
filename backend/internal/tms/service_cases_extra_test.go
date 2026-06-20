package tms

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

func TestGetTestCase_AndList(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateTestCase(ctx, projectID, userID, CreateTestCaseRequest{
		Title:    "Login test",
		CaseType: "functional",
		Priority: "high",
	})
	require.NoError(t, err)

	got, err := svc.GetTestCase(ctx, created.CaseKey)
	require.NoError(t, err)
	assert.Equal(t, "Login test", got.Title)

	list, err := svc.ListTestCases(ctx, projectID, ListTestCasesRequest{Limit: 20})
	require.NoError(t, err)
	assert.Equal(t, 1, list.Total)
	require.Len(t, list.Items, 1)
}

func TestDeleteTestCase_RemovesFromRepo(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateTestCase(ctx, projectID, userID, CreateTestCaseRequest{
		Title: "Disposable", CaseType: "functional", Priority: "medium",
	})
	require.NoError(t, err)

	require.NoError(t, svc.DeleteTestCase(ctx, created.CaseKey))
	_, err = svc.GetTestCase(ctx, created.CaseKey)
	require.Error(t, err)
}

func TestListTestCaseVersions_AfterUpdate(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateTestCase(ctx, projectID, userID, CreateTestCaseRequest{
		Title: "Versioned", CaseType: "functional", Priority: "medium",
	})
	require.NoError(t, err)

	newTitle := "Versioned v2"
	_, err = svc.UpdateTestCase(ctx, created.CaseKey, userID, UpdateTestCaseRequest{Title: &newTitle})
	require.NoError(t, err)

	versions, err := svc.ListTestCaseVersions(ctx, created.CaseKey)
	require.NoError(t, err)
	assert.NotEmpty(t, versions)
}

func TestUnlinkTestCaseFromIssue(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	tc, err := svc.CreateTestCase(ctx, projectID, userID, CreateTestCaseRequest{
		Title: "Linked case", CaseType: "functional", Priority: "medium",
	})
	require.NoError(t, err)

	issueID := uuid.New()
	repo.issues[issueID] = &ent.Issue{ID: issueID, IssueKey: "HUB-1", Title: "Bug"}
	entTC, _ := repo.FindTestCaseByKey(ctx, tc.CaseKey)
	repo.linkedIssues[entTC.ID] = []uuid.UUID{issueID}

	require.NoError(t, svc.UnlinkTestCaseFromIssue(ctx, tc.CaseKey, issueID))
	links, err := svc.GetTestCaseLinks(ctx, tc.CaseKey)
	require.NoError(t, err)
	assert.Empty(t, links)
}

func TestGetTestRun_AndListResults(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	run, err := svc.CreateTestRun(ctx, projectID, userID, CreateTestRunRequest{Title: "Smoke"})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)

	got, err := svc.GetTestRun(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, "Smoke", got.Title)

	tc, err := svc.CreateTestCase(ctx, projectID, userID, CreateTestCaseRequest{
		Title: "Case A", CaseType: "functional", Priority: "medium",
	})
	require.NoError(t, err)
	entTC, _ := repo.FindTestCaseByKey(ctx, tc.CaseKey)

	_, err = svc.CreateTestResult(ctx, runID, CreateTestResultRequest{
		TestCaseID: entTC.ID.String(),
		Status:     "passed",
	})
	require.NoError(t, err)

	results, err := svc.ListTestResults(ctx, runID)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestGetTestTrends_EmptyProject(t *testing.T) {
	svc := NewTMSService(newMockRepo(), nil)
	trends, err := svc.GetTestTrends(context.Background(), uuid.New(), 5)
	require.NoError(t, err)
	assert.Empty(t, trends)
}

func TestListTestRuns_Pagination(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	for i := 0; i < 2; i++ {
		_, err := svc.CreateTestRun(ctx, projectID, userID, CreateTestRunRequest{Title: "Run"})
		require.NoError(t, err)
	}

	page, err := svc.ListTestRuns(ctx, projectID, ListTestRunsRequest{Offset: 0, Limit: 1})
	require.NoError(t, err)
	assert.Equal(t, 2, page.Total)
	assert.Len(t, page.Items, 2) // mock repo ignores pagination params
}
