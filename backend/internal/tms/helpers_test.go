package tms

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/ent/testresult"
	"github.com/valerii/hub/backend/ent/testrun"
)

func TestToSuiteResponse_WithParent(t *testing.T) {
	parentID := uuid.New()
	suite := &ent.TestSuite{
		ID:          uuid.New(),
		Name:        "Regression",
		Description: "full suite",
		SortOrder:   1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Edges: ent.TestSuiteEdges{
			Parent: &ent.TestSuite{ID: parentID},
		},
	}

	resp := toSuiteResponse(suite)
	require.NotNil(t, resp)
	assert.Equal(t, parentID.String(), resp.ParentID)
}

func TestToComponentResponse(t *testing.T) {
	now := time.Now()
	c := &ent.Component{
		ID:          uuid.New(),
		Name:        "API",
		Description: "backend tests",
		CreatedAt:   now,
	}
	resp := toComponentResponse(c)
	assert.Equal(t, "API", resp.Name)
	assert.Equal(t, "backend tests", resp.Description)
}

func TestToTestRunResponse_WithCreator(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	run := &ent.TestRun{
		ID:           uuid.New(),
		Title:        "Smoke",
		RunType:      testrun.RunTypeManual,
		Status:       testrun.StatusCompleted,
		TotalCount:   10,
		PassedCount:  9,
		FailedCount:  1,
		CreatedAt:    now,
		UpdatedAt:    now,
		Edges: ent.TestRunEdges{
			CreatedBy: &ent.User{ID: userID, Name: "QA", Email: "qa@test.com", AvatarURL: "q.png"},
		},
	}

	resp := toTestRunResponse(run)
	require.NotNil(t, resp)
	assert.Equal(t, "Smoke", resp.Title)
	require.NotNil(t, resp.CreatedBy)
	assert.Equal(t, "QA", resp.CreatedBy.Name)
}

func TestToTestResultResponse(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	duration := int64(1200)
	r := &ent.TestResult{
		ID:         uuid.New(),
		Status:     testresult.StatusFailed,
		Comment:    "assertion",
		DurationMs: &duration,
		IsFlaky:    true,
		CreatedAt:  now,
		Edges: ent.TestResultEdges{
			TestCase:   &ent.TestCase{CaseKey: "HUB-TC-3"},
			ExecutedBy: &ent.User{ID: userID, Name: "Tester", Email: "t@test.com"},
		},
	}

	resp := toTestResultResponse(r)
	require.NotNil(t, resp)
	assert.Equal(t, "failed", resp.Status)
	assert.Equal(t, "HUB-TC-3", resp.TestCaseKey)
	require.NotNil(t, resp.ExecutedBy)
	assert.True(t, resp.IsFlaky)
}

func TestBuildSuiteTree_NestedChildren(t *testing.T) {
	rootID := uuid.New()
	childID := uuid.New()
	grandID := uuid.New()

	suites := []*ent.TestSuite{
		{ID: rootID, Name: "Root"},
		{ID: childID, Name: "Child", Edges: ent.TestSuiteEdges{Parent: &ent.TestSuite{ID: rootID}}},
		{ID: grandID, Name: "Grand", Edges: ent.TestSuiteEdges{Parent: &ent.TestSuite{ID: childID}}},
	}

	tree := buildSuiteTree(suites)
	require.Len(t, tree, 1)
	assert.Equal(t, "Root", tree[0].Name)
}

func TestSnapshotToUpdateInput_PartialFields(t *testing.T) {
	input := snapshotToUpdateInput(map[string]interface{}{
		"title":    "Only title",
		"priority": string(testcase.PriorityHigh),
	})
	require.NotNil(t, input.Title)
	assert.Equal(t, "Only title", *input.Title)
	require.NotNil(t, input.Priority)
	assert.Equal(t, string(testcase.PriorityHigh), *input.Priority)
	assert.Nil(t, input.Description)
}
