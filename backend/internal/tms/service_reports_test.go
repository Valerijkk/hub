package tms

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/testrun"
)

func TestListTestCases_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	_, err := svc.CreateTestCase(ctx, projectID, uuid.New(), CreateTestCaseRequest{Title: "A"})
	require.NoError(t, err)
	_, err = svc.CreateTestCase(ctx, projectID, uuid.New(), CreateTestCaseRequest{Title: "B"})
	require.NoError(t, err)

	resp, err := svc.ListTestCases(ctx, projectID, ListTestCasesRequest{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Items, 2)
}

func TestDeleteTestCase(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()

	created, err := svc.CreateTestCase(ctx, uuid.New(), uuid.New(), CreateTestCaseRequest{Title: "Delete me"})
	require.NoError(t, err)

	require.NoError(t, svc.DeleteTestCase(ctx, created.CaseKey))
	_, err = svc.GetTestCase(ctx, created.CaseKey)
	require.Error(t, err)
}

func TestListTestCaseVersions(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()

	created, err := svc.CreateTestCase(ctx, uuid.New(), uuid.New(), CreateTestCaseRequest{Title: "Versioned"})
	require.NoError(t, err)

	newTitle := "v2"
	_, err = svc.UpdateTestCase(ctx, created.CaseKey, uuid.New(), UpdateTestCaseRequest{
		Title:   &newTitle,
		Version: 1,
	})
	require.NoError(t, err)

	versions, err := svc.ListTestCaseVersions(ctx, created.CaseKey)
	require.NoError(t, err)
	assert.Len(t, versions, 1)
}

func TestGetTestTrends(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
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

func TestGetTestTrends_DefaultLimit(t *testing.T) {
	svc := NewTMSService(newMockRepo(), nil)
	trends, err := svc.GetTestTrends(context.Background(), uuid.New(), 0)
	require.NoError(t, err)
	assert.Empty(t, trends)
}

func TestParseUUIDs(t *testing.T) {
	id := uuid.New()
	ids, err := parseUUIDs([]string{id.String()})
	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{id}, ids)

	_, err = parseUUIDs([]string{"bad"})
	require.Error(t, err)
}

func TestGetTraceabilityMatrix(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
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

func TestGetDefectCategories(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
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
