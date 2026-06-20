package tms

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/schema"
)

// ---------------------------------------------------------------------------
// Фикстуры JSON
// ---------------------------------------------------------------------------

const validAllureJSON = `[
  {
    "uuid": "abc-123",
    "historyId": "hist-1",
    "testCaseId": "tc-1",
    "fullName": "com.example.LoginTest.testSuccess",
    "name": "testSuccess",
    "status": "passed",
    "statusDetails": {
      "known": false,
      "message": "",
      "trace": ""
    },
    "stage": "finished",
    "start": 1700000000000,
    "stop": 1700000001000,
    "steps": [
      {
        "name": "Open page",
        "status": "passed",
        "start": 1700000000100,
        "stop": 1700000000500,
        "steps": []
      }
    ],
    "labels": [
      {"name": "suite", "value": "LoginTest"},
      {"name": "severity", "value": "critical"}
    ],
    "links": [
      {"name": "HUB-TC-42", "url": "https://hub.test/tc/42", "type": "tms"}
    ],
    "parameters": [
      {"name": "browser", "value": "chrome"}
    ],
    "attachments": [
      {"name": "screenshot.png", "source": "abc.png", "type": "image/png"}
    ],
    "description": "Проверка успешного логина"
  }
]`

const minimalAllureJSON = `[
  {
    "name": "test_minimal",
    "status": "failed"
  }
]`

const emptyArrayJSON = `[]`

const malformedJSON = `{not valid json`

const nullTimestampsJSON = `[
  {
    "uuid": "null-ts",
    "name": "test_null_timestamps",
    "status": "broken",
    "start": null,
    "stop": null,
    "statusDetails": null
  }
]`

const annotationLinkJSON = `[
  {
    "uuid": "link-1",
    "name": "testWithLink",
    "status": "passed",
    "fullName": "com.example.SomeTest.testWithLink",
    "links": [
      {"name": "HUB-TC-42", "url": "", "type": "tms"}
    ],
    "labels": [
      {"name": "suite", "value": "SomeTest"}
    ]
  }
]`

const asIDLabelJSON = `[
  {
    "uuid": "label-1",
    "name": "testWithASID",
    "status": "passed",
    "fullName": "com.example.SomeTest.testWithASID",
    "labels": [
      {"name": "AS_ID", "value": "HUB-TC-99"},
      {"name": "suite", "value": "SomeTest"}
    ]
  }
]`

const fullNameOnlyJSON = `[
  {
    "uuid": "fn-1",
    "name": "testFullName",
    "status": "passed",
    "fullName": "com.example.SearchTest.testFullName",
    "labels": [
      {"name": "suite", "value": "SearchTest"}
    ]
  }
]`

// ---------------------------------------------------------------------------
// Тесты парсера
// ---------------------------------------------------------------------------

func TestParseAllureResults_ValidJSON(t *testing.T) {
	results, err := ParseAllureResults([]byte(validAllureJSON))
	require.NoError(t, err)
	require.Len(t, results, 1)

	r := results[0]
	assert.Equal(t, "abc-123", r.UUID)
	assert.Equal(t, "hist-1", r.HistoryID)
	assert.Equal(t, "com.example.LoginTest.testSuccess", r.FullName)
	assert.Equal(t, "testSuccess", r.Name)
	assert.Equal(t, "passed", r.Status)
	assert.NotNil(t, r.StatusDetails)
	assert.NotNil(t, r.Start)
	assert.NotNil(t, r.Stop)
	assert.Len(t, r.Steps, 1)
	assert.Len(t, r.Labels, 2)
	assert.Len(t, r.Links, 1)
	assert.Len(t, r.Parameters, 1)
	assert.Len(t, r.Attachments, 1)
	assert.Equal(t, "Проверка успешного логина", r.Description)
}

func TestParseAllureResults_MissingFields(t *testing.T) {
	results, err := ParseAllureResults([]byte(minimalAllureJSON))
	require.NoError(t, err)
	require.Len(t, results, 1)

	r := results[0]
	assert.Equal(t, "test_minimal", r.Name)
	assert.Equal(t, "failed", r.Status)
	assert.Nil(t, r.StatusDetails)
	assert.Nil(t, r.Start)
	assert.Nil(t, r.Stop)
	assert.Empty(t, r.Steps)
	assert.Empty(t, r.Labels)
}

func TestParseAllureResults_EmptyInput(t *testing.T) {
	results, err := ParseAllureResults([]byte(emptyArrayJSON))
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestParseAllureResults_MalformedJSON(t *testing.T) {
	results, err := ParseAllureResults([]byte(malformedJSON))
	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestParseAllureResults_NilPointers(t *testing.T) {
	results, err := ParseAllureResults([]byte(nullTimestampsJSON))
	require.NoError(t, err)
	require.Len(t, results, 1)

	r := results[0]
	assert.Nil(t, r.Start)
	assert.Nil(t, r.Stop)
	assert.Nil(t, r.StatusDetails)
}

func TestParseAllureResult_Single(t *testing.T) {
	data := []byte(`{"uuid":"single-1","name":"solo","status":"passed","fullName":"com.example.SoloTest.solo"}`)
	result, err := ParseAllureResult(data)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "single-1", result.UUID)
	assert.Equal(t, "solo", result.Name)
	assert.Equal(t, "passed", result.Status)
}

func TestParseAllureResult_MalformedJSON(t *testing.T) {
	result, err := ParseAllureResult([]byte(`{broken`))
	assert.Error(t, err)
	assert.Nil(t, result)
}

// ---------------------------------------------------------------------------
// Тесты вспомогательных функций парсера
// ---------------------------------------------------------------------------

func TestExtractLabelValue(t *testing.T) {
	labels := []AllureLabel{
		{Name: "suite", Value: "LoginTest"},
		{Name: "severity", Value: "critical"},
	}
	assert.Equal(t, "LoginTest", extractLabelValue(labels, "suite"))
	assert.Equal(t, "critical", extractLabelValue(labels, "severity"))
	assert.Equal(t, "", extractLabelValue(labels, "nonexistent"))
	assert.Equal(t, "", extractLabelValue(nil, "suite"))
}

func TestAllureStatusToTMS(t *testing.T) {
	assert.Equal(t, "passed", allureStatusToTMS("passed"))
	assert.Equal(t, "failed", allureStatusToTMS("failed"))
	assert.Equal(t, "broken", allureStatusToTMS("broken"))
	assert.Equal(t, "skipped", allureStatusToTMS("skipped"))
	assert.Equal(t, "unknown", allureStatusToTMS("unknown"))
	assert.Equal(t, "unknown", allureStatusToTMS(""))
	assert.Equal(t, "unknown", allureStatusToTMS("weird_status"))
}

// ---------------------------------------------------------------------------
// Мок-репозиторий для тестов матчера
// ---------------------------------------------------------------------------

type mockMatcherRepo struct {
	Repository
	externalIDMap  map[string]*ent.TestCase
	fullNameMap    map[string]*ent.TestCase
	classMethodMap map[string]*ent.TestCase
}

func newMockMatcherRepo() *mockMatcherRepo {
	return &mockMatcherRepo{
		externalIDMap:  make(map[string]*ent.TestCase),
		fullNameMap:    make(map[string]*ent.TestCase),
		classMethodMap: make(map[string]*ent.TestCase),
	}
}

func (m *mockMatcherRepo) FindTestCaseByExternalID(_ context.Context, _ uuid.UUID, externalID string) (*ent.TestCase, error) {
	if tc, ok := m.externalIDMap[externalID]; ok {
		return tc, nil
	}
	return nil, &ent.NotFoundError{}
}

func (m *mockMatcherRepo) FindTestCaseByFullName(_ context.Context, _ uuid.UUID, fullName string) (*ent.TestCase, error) {
	if tc, ok := m.fullNameMap[fullName]; ok {
		return tc, nil
	}
	return nil, &ent.NotFoundError{}
}

func (m *mockMatcherRepo) FindTestCaseByClassMethod(_ context.Context, _ uuid.UUID, className, methodName string) (*ent.TestCase, error) {
	key := className + "::" + methodName
	if tc, ok := m.classMethodMap[key]; ok {
		return tc, nil
	}
	return nil, &ent.NotFoundError{}
}

// Остальные методы репозитория не нужны для тестов матчера, embedded interface обеспечивает заглушки.

// Стаб-реализации для embedded Repository interface (предотвращают nil panic).
func (m *mockMatcherRepo) CreateTestCase(_ context.Context, _ uuid.UUID, _ CreateTestCaseInput) (*ent.TestCase, error) {
	return nil, nil
}
func (m *mockMatcherRepo) GetTestCase(_ context.Context, _ uuid.UUID) (*ent.TestCase, error) {
	return nil, nil
}
func (m *mockMatcherRepo) FindTestCaseByKey(_ context.Context, _ string) (*ent.TestCase, error) {
	return nil, nil
}
func (m *mockMatcherRepo) ListTestCases(_ context.Context, _ uuid.UUID, _ ListTestCasesParams) ([]*ent.TestCase, int, error) {
	return nil, 0, nil
}
func (m *mockMatcherRepo) UpdateTestCase(_ context.Context, _ uuid.UUID, _ UpdateTestCaseInput) (*ent.TestCase, error) {
	return nil, nil
}
func (m *mockMatcherRepo) DeleteTestCase(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockMatcherRepo) CreateTestCaseVersion(_ context.Context, _, _ uuid.UUID, _ int, _ map[string]interface{}, _ string) (*ent.TestCaseVersion, error) {
	return nil, nil
}
func (m *mockMatcherRepo) ListTestCaseVersions(_ context.Context, _ uuid.UUID) ([]*ent.TestCaseVersion, error) {
	return nil, nil
}
func (m *mockMatcherRepo) GetTestCaseVersion(_ context.Context, _ uuid.UUID) (*ent.TestCaseVersion, error) {
	return nil, nil
}
func (m *mockMatcherRepo) CreateSuite(_ context.Context, _ uuid.UUID, _ CreateSuiteInput) (*ent.TestSuite, error) {
	return nil, nil
}
func (m *mockMatcherRepo) GetSuite(_ context.Context, _ uuid.UUID) (*ent.TestSuite, error) {
	return nil, nil
}
func (m *mockMatcherRepo) ListSuites(_ context.Context, _ uuid.UUID) ([]*ent.TestSuite, error) {
	return nil, nil
}
func (m *mockMatcherRepo) UpdateSuite(_ context.Context, _ uuid.UUID, _ UpdateSuiteInput) (*ent.TestSuite, error) {
	return nil, nil
}
func (m *mockMatcherRepo) DeleteSuite(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockMatcherRepo) MoveSuite(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (*ent.TestSuite, error) {
	return nil, nil
}
func (m *mockMatcherRepo) CreateComponent(_ context.Context, _ uuid.UUID, _, _ string) (*ent.Component, error) {
	return nil, nil
}
func (m *mockMatcherRepo) ListComponents(_ context.Context, _ uuid.UUID) ([]*ent.Component, error) {
	return nil, nil
}
func (m *mockMatcherRepo) DeleteComponent(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockMatcherRepo) CreateTestRun(_ context.Context, _ uuid.UUID, _ CreateTestRunInput) (*ent.TestRun, error) {
	return nil, nil
}
func (m *mockMatcherRepo) GetTestRun(_ context.Context, _ uuid.UUID) (*ent.TestRun, error) {
	return nil, nil
}
func (m *mockMatcherRepo) ListTestRuns(_ context.Context, _ uuid.UUID, _ ListTestRunsParams) ([]*ent.TestRun, int, error) {
	return nil, 0, nil
}
func (m *mockMatcherRepo) UpdateTestRunStatus(_ context.Context, _ uuid.UUID, _ string, _ *time.Time) (*ent.TestRun, error) {
	return nil, nil
}
func (m *mockMatcherRepo) ReopenTestRun(_ context.Context, _ uuid.UUID) (*ent.TestRun, error) {
	return nil, nil
}
func (m *mockMatcherRepo) UpdateTestRunStats(_ context.Context, _ uuid.UUID, _ TestRunStats) (*ent.TestRun, error) {
	return nil, nil
}
func (m *mockMatcherRepo) CreateTestResult(_ context.Context, _ uuid.UUID, _ CreateTestResultInput) (*ent.TestResult, error) {
	return nil, nil
}
func (m *mockMatcherRepo) UpdateTestResult(_ context.Context, _ uuid.UUID, _ UpdateTestResultInput) (*ent.TestResult, error) {
	return nil, nil
}
func (m *mockMatcherRepo) ListTestResults(_ context.Context, _ uuid.UUID) ([]*ent.TestResult, error) {
	return nil, nil
}
func (m *mockMatcherRepo) BatchCreateTestResults(_ context.Context, _ uuid.UUID, _ []CreateTestResultInput) ([]*ent.TestResult, error) {
	return nil, nil
}
func (m *mockMatcherRepo) IncrementTestCaseCounter(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockMatcherRepo) LinkTestCaseToIssue(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockMatcherRepo) UnlinkTestCaseFromIssue(_ context.Context, _, _ uuid.UUID) error {
	return nil
}
func (m *mockMatcherRepo) AddTestCasesToRun(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error {
	return nil
}
func (m *mockMatcherRepo) GetTestResult(_ context.Context, _ uuid.UUID) (*ent.TestResult, error) {
	return nil, nil
}
func (m *mockMatcherRepo) ListTestRunResults(_ context.Context, _ uuid.UUID) ([]*ent.TestResult, error) {
	return nil, nil
}
func (m *mockMatcherRepo) AssignTestersToRun(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error {
	return nil
}
func (m *mockMatcherRepo) GetTestRunWithResults(_ context.Context, _ uuid.UUID) (*ent.TestRun, error) {
	return nil, nil
}
func (m *mockMatcherRepo) AddFilesToResult(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error {
	return nil
}
func (m *mockMatcherRepo) StoreAllureResults(_ context.Context, _ uuid.UUID, _ []AllureResult) error {
	return nil
}
func (m *mockMatcherRepo) GetAllureResultsByRunID(_ context.Context, _ uuid.UUID) ([]AllureResult, error) {
	return nil, nil
}

// Тест-кейс для использования в моках.
func makeMockTestCase(id uuid.UUID, key string) *ent.TestCase {
	return &ent.TestCase{
		ID:      id,
		CaseKey: key,
		Title:   "Mock " + key,
	}
}

// Suppress unused import warning for schema
var _ = schema.TestStep{}

// ---------------------------------------------------------------------------
// Тесты матчера
// ---------------------------------------------------------------------------

func TestMatchTestCase_ByAnnotation(t *testing.T) {
	repo := newMockMatcherRepo()
	tcID := uuid.New()
	tc := makeMockTestCase(tcID, "HUB-TC-42")
	repo.externalIDMap["HUB-TC-42"] = tc

	matcher := NewMatcher(repo)
	result := AllureResult{
		UUID:     "link-1",
		Name:     "testWithLink",
		Status:   "passed",
		FullName: "com.example.SomeTest.testWithLink",
		Links: []AllureLink{
			{Name: "HUB-TC-42", URL: "", Type: "tms"},
		},
	}

	projectID := uuid.New()
	matched, strategy, err := matcher.MatchTestCase(context.Background(), projectID, &result)
	require.NoError(t, err)
	require.NotNil(t, matched)
	assert.Equal(t, tcID, matched.ID)
	assert.Equal(t, MatchByAnnotation, strategy)
}

func TestMatchTestCase_ByLink(t *testing.T) {
	repo := newMockMatcherRepo()
	tcID := uuid.New()
	tc := makeMockTestCase(tcID, "HUB-TC-42")
	repo.externalIDMap["HUB-TC-42"] = tc

	matcher := NewMatcher(repo)
	result := AllureResult{
		UUID:   "link-2",
		Name:   "testWithTmsLink",
		Status: "passed",
		Links: []AllureLink{
			{Name: "HUB-TC-42", URL: "", Type: "tms"},
		},
	}

	projectID := uuid.New()
	matched, strategy, err := matcher.MatchTestCase(context.Background(), projectID, &result)
	require.NoError(t, err)
	require.NotNil(t, matched)
	assert.Equal(t, tcID, matched.ID)
	assert.Equal(t, MatchByAnnotation, strategy)
}

func TestMatchTestCase_ByASIDLabel(t *testing.T) {
	repo := newMockMatcherRepo()
	tcID := uuid.New()
	tc := makeMockTestCase(tcID, "HUB-TC-99")
	repo.externalIDMap["HUB-TC-99"] = tc

	matcher := NewMatcher(repo)
	result := AllureResult{
		UUID:     "label-1",
		Name:     "testWithASID",
		Status:   "passed",
		FullName: "com.example.SomeTest.testWithASID",
		Labels: []AllureLabel{
			{Name: "AS_ID", Value: "HUB-TC-99"},
			{Name: "suite", Value: "SomeTest"},
		},
	}

	projectID := uuid.New()
	matched, strategy, err := matcher.MatchTestCase(context.Background(), projectID, &result)
	require.NoError(t, err)
	require.NotNil(t, matched)
	assert.Equal(t, tcID, matched.ID)
	assert.Equal(t, MatchByAnnotation, strategy)
}

func TestMatchTestCase_ByFullName(t *testing.T) {
	repo := newMockMatcherRepo()
	tcID := uuid.New()
	tc := makeMockTestCase(tcID, "HUB-TC-55")
	repo.fullNameMap["com.example.SearchTest.testFullName"] = tc

	matcher := NewMatcher(repo)
	result := AllureResult{
		UUID:     "fn-1",
		Name:     "testFullName",
		Status:   "passed",
		FullName: "com.example.SearchTest.testFullName",
		Labels: []AllureLabel{
			{Name: "suite", Value: "SearchTest"},
		},
	}

	projectID := uuid.New()
	matched, strategy, err := matcher.MatchTestCase(context.Background(), projectID, &result)
	require.NoError(t, err)
	require.NotNil(t, matched)
	assert.Equal(t, tcID, matched.ID)
	assert.Equal(t, MatchByFullName, strategy)
}

func TestMatchTestCase_ByClassName(t *testing.T) {
	repo := newMockMatcherRepo()
	tcID := uuid.New()
	tc := makeMockTestCase(tcID, "HUB-TC-77")
	repo.classMethodMap["LoginTest::testLogin"] = tc

	matcher := NewMatcher(repo)
	result := AllureResult{
		UUID:   "cm-1",
		Name:   "testLogin",
		Status: "passed",
		Labels: []AllureLabel{
			{Name: "suite", Value: "LoginTest"},
		},
	}

	projectID := uuid.New()
	matched, strategy, err := matcher.MatchTestCase(context.Background(), projectID, &result)
	require.NoError(t, err)
	require.NotNil(t, matched)
	assert.Equal(t, tcID, matched.ID)
	assert.Equal(t, MatchByClassName, strategy)
}

func TestMatchTestCase_NoMatch(t *testing.T) {
	repo := newMockMatcherRepo()
	matcher := NewMatcher(repo)
	result := AllureResult{
		UUID:   "no-1",
		Name:   "testUnknown",
		Status: "passed",
	}

	projectID := uuid.New()
	matched, strategy, err := matcher.MatchTestCase(context.Background(), projectID, &result)
	require.NoError(t, err)
	assert.Nil(t, matched)
	assert.Equal(t, MatchStrategy(-1), strategy)
}

func TestMatchTestCase_PriorityOrder(t *testing.T) {
	// Результат имеет и аннотацию (link type=tms), и fullName.
	// Аннотация должна победить.
	repo := newMockMatcherRepo()
	annotationTC := makeMockTestCase(uuid.New(), "HUB-TC-10")
	fullNameTC := makeMockTestCase(uuid.New(), "HUB-TC-20")
	repo.externalIDMap["HUB-TC-10"] = annotationTC
	repo.fullNameMap["com.example.PrioTest.testPrio"] = fullNameTC

	matcher := NewMatcher(repo)
	result := AllureResult{
		UUID:     "prio-1",
		Name:     "testPrio",
		Status:   "passed",
		FullName: "com.example.PrioTest.testPrio",
		Links: []AllureLink{
			{Name: "HUB-TC-10", URL: "", Type: "tms"},
		},
	}

	projectID := uuid.New()
	matched, strategy, err := matcher.MatchTestCase(context.Background(), projectID, &result)
	require.NoError(t, err)
	require.NotNil(t, matched)
	assert.Equal(t, annotationTC.ID, matched.ID)
	assert.Equal(t, MatchByAnnotation, strategy)
}
