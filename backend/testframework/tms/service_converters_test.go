package tms_test

// Тесты сервиса TMS: интеграционные и unit-сценарии testframework.
import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/schema"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/ent/testresult"
	"github.com/valerii/hub/backend/ent/testrun"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestToTestCaseResponse_WithEdges проверяет сценарий: to test case response with edges.
func TestToTestCaseResponse_WithEdges(t *testing.T) {
	suiteID := uuid.New()
	userID := uuid.New()
	labelID := uuid.New()
	componentID := uuid.New()
	issueID := uuid.New()
	now := time.Now()

	tc := &ent.TestCase{
		ID:            uuid.New(),
		CaseNumber:    7,
		CaseKey:       "HUB-TC-7",
		Title:         "Login flow",
		Description:   "desc",
		Precondition:  "pre",
		Postcondition: "post",
		CaseType:      testcase.CaseTypeClassic,
		Priority:      testcase.PriorityHigh,
		Status:        testcase.StatusActive,
		Steps:         []schema.TestStep{{Order: 1, Action: "click"}},
		FullName:      "com.example.LoginTest.testLogin",
		ClassName:     "LoginTest",
		MethodName:    "testLogin",
		ExternalID:    "EXT-1",
		AutomationID:  "AUTO-1",
		Version:       3,
		CreatedAt:     now,
		UpdatedAt:     now,
		Edges: ent.TestCaseEdges{
			Suite:     &ent.TestSuite{ID: suiteID, Name: "Auth"},
			CreatedBy: &ent.User{ID: userID, Name: "Alice", Email: "alice@test.com", AvatarURL: "a.png"},
			Labels:    []*ent.Label{{ID: labelID, Name: "smoke", Color: "#fff"}},
			Components: []*ent.Component{
				{ID: componentID, Name: "API", Description: "backend", CreatedAt: now},
			},
			LinkedIssues: []*ent.Issue{{ID: issueID, IssueKey: "HUB-1", Title: "Auth bug"}},
		},
	}

	resp := tms.ExportToTestCaseResponse(tc)
	require.NotNil(t, resp)
	assert.Equal(t, "HUB-TC-7", resp.CaseKey)
	require.NotNil(t, resp.Suite)
	assert.Equal(t, "Auth", resp.Suite.Name)
	require.NotNil(t, resp.CreatedBy)
	assert.Equal(t, "Alice", resp.CreatedBy.Name)
	assert.Len(t, resp.Labels, 1)
	assert.Len(t, resp.Components, 1)
	assert.Len(t, resp.LinkedIssues, 1)
}

// TestToTestRunDetailResponse проверяет сценарий: to test run detail response.
func TestToTestRunDetailResponse(t *testing.T) {
	userID := uuid.New()
	fileID := uuid.New()
	tcID := uuid.New()
	now := time.Now()

	run := &ent.TestRun{
		ID:           uuid.New(),
		Title:        "Nightly",
		RunType:      testrun.RunTypeAutomatic,
		Status:       testrun.StatusInProgress,
		TotalCount:   2,
		PassedCount:  1,
		CreatedAt:    now,
		UpdatedAt:    now,
		Edges: ent.TestRunEdges{
			CreatedBy: &ent.User{ID: userID, Name: "Bob", Email: "bob@test.com"},
			Assignees: []*ent.User{{ID: uuid.New(), Name: "Carol", Email: "carol@test.com"}},
			Results: []*ent.TestResult{
				{
					ID:        uuid.New(),
					Status:    testresult.StatusPassed,
					Comment:   "ok",
					CreatedAt: now,
					Edges: ent.TestResultEdges{
						TestCase: &ent.TestCase{
							ID:       tcID,
							CaseKey:  "HUB-TC-1",
							Title:    "Case 1",
							CaseType: testcase.CaseTypeClassic,
							Priority: testcase.PriorityMedium,
						},
						ExecutedBy: &ent.User{ID: userID, Name: "Bob", Email: "bob@test.com"},
						Files: []*ent.FileMetadata{
							{ID: fileID, Filename: "shot.png", Size: 1024, ContentType: "image/png"},
						},
					},
				},
			},
		},
	}

	resp := tms.ExportToTestRunDetailResponse(run)
	require.NotNil(t, resp)
	assert.Equal(t, "Nightly", resp.Title)
	assert.Len(t, resp.Assignees, 1)
	require.Len(t, resp.Results, 1)
	require.NotNil(t, resp.Results[0].TestCase)
	assert.Equal(t, "HUB-TC-1", resp.Results[0].TestCase.CaseKey)
	assert.Len(t, resp.Results[0].Files, 1)
}

// TestBuildSuiteTree проверяет сценарий: build suite tree.
func TestBuildSuiteTree(t *testing.T) {
	rootID := uuid.New()
	childID := uuid.New()
	orphanParentID := uuid.New()

	suites := []*ent.TestSuite{
		{ID: rootID, Name: "Root"},
		{ID: childID, Name: "Child", Edges: ent.TestSuiteEdges{Parent: &ent.TestSuite{ID: rootID}}},
		{ID: uuid.New(), Name: "Orphan", Edges: ent.TestSuiteEdges{Parent: &ent.TestSuite{ID: orphanParentID}}},
	}

	tree := tms.ExportBuildSuiteTree(suites)
	require.Len(t, tree, 1)
	assert.Equal(t, "Root", tree[0].Name)
}

// TestBuildSuiteTree_EmptyReturnsSlice проверяет сценарий: build suite tree пустой returns slice.
func TestBuildSuiteTree_EmptyReturnsSlice(t *testing.T) {
	tree := tms.ExportBuildSuiteTree(nil)
	assert.NotNil(t, tree)
	assert.Empty(t, tree)
}

// TestBuildSnapshotAndSnapshotToUpdateInput проверяет сценарий: build snapshot and snapshot to обновление input.
func TestBuildSnapshotAndSnapshotToUpdateInput(t *testing.T) {
	tc := &ent.TestCase{
		Title:        "Title v1",
		Description:  "desc",
		Precondition: "pre",
		CaseType:     testcase.CaseTypeBdd,
		Priority:     testcase.PriorityLow,
		Status:       testcase.StatusDraft,
		Version:      2,
		Steps:        []schema.TestStep{{Order: 1, Keyword: "Given", Action: "setup"}},
		FullName:     "com.example.Test.test",
		ClassName:    "Test",
		MethodName:   "test",
		ExternalID:   "EXT",
	}

	snapshot := tms.ExportBuildSnapshot(tc)
	assert.Equal(t, "Title v1", snapshot["title"])
	assert.Equal(t, "bdd", snapshot["case_type"])
	assert.NotNil(t, snapshot["steps"])

	input := tms.ExportSnapshotToUpdateInput(snapshot)
	require.NotNil(t, input.Title)
	assert.Equal(t, "Title v1", *input.Title)
	require.NotNil(t, input.Steps)
	assert.Len(t, *input.Steps, 1)
}

// TestNormalizeErrorPattern проверяет сценарий: normalize ошибка pattern.
func TestNormalizeErrorPattern(t *testing.T) {
	short := "NullPointerException"
	assert.Equal(t, short, tms.ExportNormalizeErrorPattern(short))

	longMsg := string(make([]byte, 150))
	for i := range longMsg {
		longMsg = longMsg[:i] + "x" + longMsg[i+1:]
	}
	assert.Len(t, tms.ExportNormalizeErrorPattern(longMsg), 100)
}

// TestToTestCaseVersionResponse проверяет сценарий: to test case version response.
func TestToTestCaseVersionResponse(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	v := &ent.TestCaseVersion{
		ID:                uuid.New(),
		VersionNumber:     2,
		Snapshot:          map[string]interface{}{"title": "old"},
		ChangeDescription: "edit",
		CreatedAt:         now,
		Edges: ent.TestCaseVersionEdges{
			CreatedBy: &ent.User{ID: userID, Name: "Dev", Email: "dev@test.com"},
		},
	}

	resp := tms.ExportToTestCaseVersionResponse(v)
	assert.Equal(t, 2, resp.VersionNumber)
	require.NotNil(t, resp.CreatedBy)
	assert.Equal(t, "Dev", resp.CreatedBy.Name)
}
