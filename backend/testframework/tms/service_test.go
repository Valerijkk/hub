package tms_test

// Тесты сервиса TMS: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"bytes"
	"context"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/environment"
	entissue "github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/schema"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/ent/testresult"
	"github.com/valerii/hub/backend/ent/testrun"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Мок tms.Repository
// ---------------------------------------------------------------------------

type mockRepository struct {
	testCases    map[uuid.UUID]*ent.TestCase
	versions     map[uuid.UUID]*ent.TestCaseVersion
	runs         map[uuid.UUID]*ent.TestRun
	results      map[uuid.UUID]*ent.TestResult
	runResults   map[uuid.UUID][]uuid.UUID // runID -> ID результатов
	suites       map[uuid.UUID]*ent.TestSuite
	components   map[uuid.UUID]*ent.Component
	issues       map[uuid.UUID]*ent.Issue
	linkedIssues  map[uuid.UUID][]uuid.UUID // testCaseID -> ID задач
	environments  map[uuid.UUID]*ent.Environment

	caseCounter  int
	issueCounter int
	lastVersion  *ent.TestCaseVersion

	// A2W2-2-001: счётчики batch-вызовов (mutation: N+1 снова → FAIL).
	getLastResultsBatchCalls  int
	listResultsByRunIDsCalls  int
	getLastResultSingleCalls  int
}

func (m *mockRepository) CreateEnvironment(_ context.Context, projectID uuid.UUID, input tms.CreateEnvironmentInput) (*ent.Environment, error) {
	id := uuid.New()
	label := environment.EnvLabel(orDefault(input.Label, "custom"))
	env := &ent.Environment{
		ID:        id,
		Name:      input.Name,
		EnvLabel:  label,
		Config:    input.Config,
		CreatedAt: time.Now(),
		Edges: ent.EnvironmentEdges{
			Project: &ent.Project{ID: projectID},
		},
	}
	m.environments[id] = env
	return env, nil
}
func (m *mockRepository) GetEnvironment(_ context.Context, envID uuid.UUID) (*ent.Environment, error) {
	env, ok := m.environments[envID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return env, nil
}
func (m *mockRepository) GetEnvironmentInProject(_ context.Context, projectID, envID uuid.UUID) (*ent.Environment, error) {
	env, ok := m.environments[envID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if env.Edges.Project == nil || env.Edges.Project.ID != projectID {
		return nil, &ent.NotFoundError{}
	}
	return env, nil
}
func (m *mockRepository) ListEnvironmentsByProject(_ context.Context, projectID uuid.UUID) ([]*ent.Environment, error) {
	items := make([]*ent.Environment, 0)
	for _, env := range m.environments {
		if env.Edges.Project != nil && env.Edges.Project.ID == projectID {
			items = append(items, env)
		}
	}
	return items, nil
}
func (m *mockRepository) UpdateEnvironment(_ context.Context, projectID, envID uuid.UUID, input tms.UpdateEnvironmentInput) (*ent.Environment, error) {
	env, ok := m.environments[envID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if env.Edges.Project == nil || env.Edges.Project.ID != projectID {
		return nil, &ent.NotFoundError{}
	}
	if input.Name != nil {
		env.Name = *input.Name
	}
	if input.Label != nil {
		env.EnvLabel = environment.EnvLabel(*input.Label)
	}
	if input.Config != nil {
		env.Config = *input.Config
	}
	return env, nil
}
func (m *mockRepository) DeleteEnvironment(_ context.Context, projectID, envID uuid.UUID) error {
	env, ok := m.environments[envID]
	if !ok {
		return &ent.NotFoundError{}
	}
	if env.Edges.Project == nil || env.Edges.Project.ID != projectID {
		return &ent.NotFoundError{}
	}
	delete(m.environments, envID)
	return nil
}
func (m *mockRepository) GetLatestTestRunForEnvironment(ctx context.Context, projectID, envID uuid.UUID) (*ent.TestRun, error) {
	return nil, nil
}

func newMockRepo() *mockRepository {
	return &mockRepository{
		testCases:    make(map[uuid.UUID]*ent.TestCase),
		versions:     make(map[uuid.UUID]*ent.TestCaseVersion),
		runs:         make(map[uuid.UUID]*ent.TestRun),
		results:      make(map[uuid.UUID]*ent.TestResult),
		runResults:   make(map[uuid.UUID][]uuid.UUID),
		suites:       make(map[uuid.UUID]*ent.TestSuite),
		components:   make(map[uuid.UUID]*ent.Component),
		issues:       make(map[uuid.UUID]*ent.Issue),
		linkedIssues: make(map[uuid.UUID][]uuid.UUID),
		environments: make(map[uuid.UUID]*ent.Environment),
	}
}

func (m *mockRepository) CreateTestCase(_ context.Context, projectID uuid.UUID, input tms.CreateTestCaseInput) (*ent.TestCase, error) {
	m.caseCounter++
	id := uuid.New()
	tc := &ent.TestCase{
		ID:         id,
		CaseNumber: m.caseCounter,
		CaseKey:    "HUB-TC-" + itoa(m.caseCounter),
		Title:      input.Title,
		Description:   input.Description,
		Precondition:  input.Precondition,
		Postcondition: input.Postcondition,
		CaseType:   testcase.CaseType(orDefault(input.CaseType, "classic")),
		Priority:   testcase.Priority(orDefault(input.Priority, "medium")),
		Status:     testcase.Status(orDefault(input.Status, "draft")),
		Steps:      input.Steps,
		FullName:   input.FullName,
		ClassName:  input.ClassName,
		MethodName: input.MethodName,
		ExternalID: input.ExternalID,
		Version:    1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Edges:      ent.TestCaseEdges{Project: &ent.Project{ID: projectID}},
	}
	m.testCases[id] = tc
	return tc, nil
}

func (m *mockRepository) GetTestCase(_ context.Context, id uuid.UUID) (*ent.TestCase, error) {
	tc, ok := m.testCases[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return tc, nil
}

func (m *mockRepository) FindTestCaseByKey(_ context.Context, projectID uuid.UUID, caseKey string) (*ent.TestCase, error) {
	for _, tc := range m.testCases {
		if tc.CaseKey == caseKey {
			return tc, nil
		}
	}
	return nil, &ent.NotFoundError{}
}

func (m *mockRepository) FindTestCaseByExternalID(_ context.Context, _ uuid.UUID, _ string) (*ent.TestCase, error) {
	return nil, &ent.NotFoundError{}
}

func (m *mockRepository) FindTestCaseByAutomationID(_ context.Context, _ uuid.UUID, _ string) (*ent.TestCase, error) {
	return nil, &ent.NotFoundError{}
}

func (m *mockRepository) FindTestCaseByFullName(_ context.Context, _ uuid.UUID, _ string) (*ent.TestCase, error) {
	return nil, &ent.NotFoundError{}
}

func (m *mockRepository) FindTestCaseByClassMethod(_ context.Context, _ uuid.UUID, _, _ string) (*ent.TestCase, error) {
	return nil, &ent.NotFoundError{}
}

func (m *mockRepository) ListTestCases(_ context.Context, _ uuid.UUID, _ tms.ListTestCasesParams) ([]*ent.TestCase, int, error) {
	items := make([]*ent.TestCase, 0, len(m.testCases))
	for _, tc := range m.testCases {
		items = append(items, tc)
	}
	return items, len(items), nil
}

func (m *mockRepository) UpdateTestCase(_ context.Context, id uuid.UUID, input tms.UpdateTestCaseInput) (*ent.TestCase, error) {
	tc, ok := m.testCases[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if tc.Version != input.Version {
		return nil, apperrors.NewConflict("version mismatch: тест-кейс был изменён другим пользователем")
	}
	if input.Title != nil {
		tc.Title = *input.Title
	}
	if input.Description != nil {
		tc.Description = *input.Description
	}
	if input.Steps != nil {
		tc.Steps = *input.Steps
	}
	if input.Priority != nil {
		tc.Priority = testcase.Priority(*input.Priority)
	}
	if input.CaseType != nil {
		tc.CaseType = testcase.CaseType(*input.CaseType)
	}
	tc.Version++
	tc.UpdatedAt = time.Now()
	return tc, nil
}

func (m *mockRepository) DeleteTestCase(_ context.Context, id uuid.UUID) error {
	delete(m.testCases, id)
	return nil
}

func (m *mockRepository) CountTestCasesInProject(_ context.Context, projectID uuid.UUID, ids []uuid.UUID) (int, error) {
	n := 0
	for _, id := range ids {
		tc, ok := m.testCases[id]
		if !ok {
			continue
		}
		if tc.Edges.Project != nil && tc.Edges.Project.ID == projectID {
			n++
		}
	}
	return n, nil
}

func (m *mockRepository) CountLabelsInProject(_ context.Context, _ uuid.UUID, ids []uuid.UUID) (int, error) {
	// Мок не хранит метки отдельно — для unit-тестов FK считаем валидными.
	return len(ids), nil
}

func (m *mockRepository) CreateTestCaseVersion(_ context.Context, testCaseID, userID uuid.UUID, versionNumber int, snapshot map[string]interface{}, description string) (*ent.TestCaseVersion, error) {
	id := uuid.New()
	v := &ent.TestCaseVersion{
		ID:                id,
		VersionNumber:     versionNumber,
		Snapshot:          snapshot,
		ChangeDescription: description,
		CreatedAt:         time.Now(),
		Edges: ent.TestCaseVersionEdges{
			TestCase: &ent.TestCase{ID: testCaseID},
		},
	}
	m.versions[id] = v
	m.lastVersion = v
	return v, nil
}

func (m *mockRepository) ListTestCaseVersions(_ context.Context, _ uuid.UUID) ([]*ent.TestCaseVersion, error) {
	items := make([]*ent.TestCaseVersion, 0, len(m.versions))
	for _, v := range m.versions {
		items = append(items, v)
	}
	return items, nil
}

func (m *mockRepository) GetTestCaseVersion(_ context.Context, testCaseID, versionID uuid.UUID) (*ent.TestCaseVersion, error) {
	v, ok := m.versions[versionID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if v.Edges.TestCase == nil || v.Edges.TestCase.ID != testCaseID {
		return nil, &ent.NotFoundError{}
	}
	return v, nil
}

func (m *mockRepository) CreateSuite(_ context.Context, projectID uuid.UUID, input tms.CreateSuiteInput) (*ent.TestSuite, error) {
	id := uuid.New()
	s := &ent.TestSuite{
		ID: id, Name: input.Name, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		Edges: ent.TestSuiteEdges{Project: &ent.Project{ID: projectID}},
	}
	if input.ParentID != nil {
		s.Edges.Parent = &ent.TestSuite{ID: *input.ParentID}
	}
	m.suites[id] = s
	return s, nil
}

func (m *mockRepository) GetSuite(_ context.Context, id uuid.UUID) (*ent.TestSuite, error) {
	s, ok := m.suites[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return s, nil
}

func (m *mockRepository) GetSuiteInProject(_ context.Context, projectID, id uuid.UUID) (*ent.TestSuite, error) {
	s, ok := m.suites[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if s.Edges.Project == nil || s.Edges.Project.ID != projectID {
		return nil, &ent.NotFoundError{}
	}
	return s, nil
}

func (m *mockRepository) ListSuites(_ context.Context, projectID uuid.UUID) ([]*ent.TestSuite, error) {
	items := make([]*ent.TestSuite, 0, len(m.suites))
	for _, s := range m.suites {
		if s.Edges.Project != nil && s.Edges.Project.ID == projectID {
			items = append(items, s)
		}
	}
	return items, nil
}

func (m *mockRepository) UpdateSuite(_ context.Context, projectID, id uuid.UUID, input tms.UpdateSuiteInput) (*ent.TestSuite, error) {
	s, ok := m.suites[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if s.Edges.Project == nil || s.Edges.Project.ID != projectID {
		return nil, &ent.NotFoundError{}
	}
	if input.Name != nil {
		s.Name = *input.Name
	}
	return s, nil
}

func (m *mockRepository) DeleteSuite(_ context.Context, projectID, id uuid.UUID) error {
	s, ok := m.suites[id]
	if !ok {
		return &ent.NotFoundError{}
	}
	if s.Edges.Project == nil || s.Edges.Project.ID != projectID {
		return &ent.NotFoundError{}
	}
	delete(m.suites, id)
	return nil
}

func (m *mockRepository) MoveSuite(_ context.Context, _ uuid.UUID, id uuid.UUID, _ *uuid.UUID) (*ent.TestSuite, error) {
	s, ok := m.suites[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return s, nil
}

func (m *mockRepository) CreateComponent(_ context.Context, projectID uuid.UUID, name, description string) (*ent.Component, error) {
	id := uuid.New()
	c := &ent.Component{
		ID: id, Name: name, Description: description, CreatedAt: time.Now(),
		Edges: ent.ComponentEdges{Project: &ent.Project{ID: projectID}},
	}
	m.components[id] = c
	return c, nil
}

func (m *mockRepository) ListComponents(_ context.Context, projectID uuid.UUID) ([]*ent.Component, error) {
	items := make([]*ent.Component, 0, len(m.components))
	for _, c := range m.components {
		if c.Edges.Project != nil && c.Edges.Project.ID == projectID {
			items = append(items, c)
		}
	}
	return items, nil
}

func (m *mockRepository) CountComponentsInProject(_ context.Context, projectID uuid.UUID, ids []uuid.UUID) (int, error) {
	n := 0
	for _, id := range ids {
		c, ok := m.components[id]
		if !ok {
			continue
		}
		if c.Edges.Project != nil && c.Edges.Project.ID == projectID {
			n++
		}
	}
	return n, nil
}

func (m *mockRepository) DeleteComponent(_ context.Context, projectID, id uuid.UUID) error {
	c, ok := m.components[id]
	if !ok {
		return &ent.NotFoundError{}
	}
	if c.Edges.Project == nil || c.Edges.Project.ID != projectID {
		return &ent.NotFoundError{}
	}
	delete(m.components, id)
	return nil
}

func (m *mockRepository) CreateTestRun(_ context.Context, projectID uuid.UUID, input tms.CreateTestRunInput) (*ent.TestRun, error) {
	id := uuid.New()
	runType := testrun.RunType(orDefault(input.RunType, "manual"))
	run := &ent.TestRun{
		ID:        id,
		Title:     input.Title,
		RunType:   runType,
		Status:    testrun.StatusCreated,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Edges: ent.TestRunEdges{
			Project: &ent.Project{ID: projectID},
		},
	}
	m.runs[id] = run
	return run, nil
}

func (m *mockRepository) GetTestRun(_ context.Context, id uuid.UUID) (*ent.TestRun, error) {
	run, ok := m.runs[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return run, nil
}

func (m *mockRepository) GetTestRunInProject(_ context.Context, projectID, id uuid.UUID) (*ent.TestRun, error) {
	run, ok := m.runs[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if run.Edges.Project == nil || run.Edges.Project.ID != projectID {
		return nil, &ent.NotFoundError{}
	}
	return run, nil
}

func (m *mockRepository) ListTestRuns(_ context.Context, _ uuid.UUID, _ tms.ListTestRunsParams) ([]*ent.TestRun, int, error) {
	items := make([]*ent.TestRun, 0, len(m.runs))
	for _, r := range m.runs {
		items = append(items, r)
	}
	return items, len(items), nil
}

func (m *mockRepository) UpdateTestRunStatus(_ context.Context, id uuid.UUID, status string, _ *time.Time) (*ent.TestRun, error) {
	run, ok := m.runs[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	run.Status = testrun.Status(status)
	return run, nil
}

func (m *mockRepository) ReopenTestRun(_ context.Context, id uuid.UUID) (*ent.TestRun, error) {
	run, ok := m.runs[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	run.Status = testrun.StatusInProgress
	return run, nil
}

func (m *mockRepository) UpdateTestRunStats(_ context.Context, id uuid.UUID, stats tms.TestRunStats) (*ent.TestRun, error) {
	run, ok := m.runs[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	run.TotalCount = stats.TotalCount
	run.PassedCount = stats.PassedCount
	run.FailedCount = stats.FailedCount
	run.SkippedCount = stats.SkippedCount
	run.BrokenCount = stats.BrokenCount
	run.UnknownCount = stats.UnknownCount
	return run, nil
}

func (m *mockRepository) CreateTestResult(_ context.Context, _ uuid.UUID, input tms.CreateTestResultInput) (*ent.TestResult, error) {
	id := uuid.New()
	r := &ent.TestResult{ID: id, Status: "passed", CreatedAt: time.Now()}
	m.results[id] = r
	return r, nil
}

func (m *mockRepository) UpdateTestResult(_ context.Context, runID, id uuid.UUID, input tms.UpdateTestResultInput) (*ent.TestResult, error) {
	r, ok := m.results[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if r.Edges.Run != nil && r.Edges.Run.ID != runID {
		return nil, &ent.NotFoundError{}
	}
	if ids, ok := m.runResults[runID]; ok {
		found := false
		for _, rid := range ids {
			if rid == id {
				found = true
				break
			}
		}
		if !found {
			return nil, &ent.NotFoundError{}
		}
	}
	if input.Status != nil {
		r.Status = testresult.Status(*input.Status)
	}
	if input.Comment != nil {
		r.Comment = *input.Comment
	}
	if input.ExecutedByID != nil {
		r.Edges.ExecutedBy = &ent.User{ID: *input.ExecutedByID}
	}
	return r, nil
}

func (m *mockRepository) ListTestResults(_ context.Context, runID uuid.UUID) ([]*ent.TestResult, error) {
	if ids, ok := m.runResults[runID]; ok {
		items := make([]*ent.TestResult, 0, len(ids))
		for _, id := range ids {
			if r, ok := m.results[id]; ok {
				items = append(items, r)
			}
		}
		return items, nil
	}
	items := make([]*ent.TestResult, 0, len(m.results))
	for _, r := range m.results {
		items = append(items, r)
	}
	return items, nil
}

func (m *mockRepository) BatchCreateTestResults(_ context.Context, _ uuid.UUID, inputs []tms.CreateTestResultInput) ([]*ent.TestResult, error) {
	var results []*ent.TestResult
	for range inputs {
		id := uuid.New()
		r := &ent.TestResult{ID: id, CreatedAt: time.Now()}
		m.results[id] = r
		results = append(results, r)
	}
	return results, nil
}

func (m *mockRepository) IncrementTestCaseCounter(_ context.Context, _ uuid.UUID) (int, error) {
	m.caseCounter++
	return m.caseCounter, nil
}

func (m *mockRepository) FindIssueInProject(_ context.Context, projectID, issueID uuid.UUID) (*ent.Issue, error) {
	issue, ok := m.issues[issueID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if issue.Edges.Project == nil || issue.Edges.Project.ID != projectID {
		return nil, &ent.NotFoundError{}
	}
	return issue, nil
}

func (m *mockRepository) LinkTestCaseToIssue(_ context.Context, testCaseID, issueID uuid.UUID) error {
	m.linkedIssues[testCaseID] = append(m.linkedIssues[testCaseID], issueID)
	return nil
}

func (m *mockRepository) UnlinkTestCaseFromIssue(_ context.Context, testCaseID, issueID uuid.UUID) error {
	ids := m.linkedIssues[testCaseID]
	for i, id := range ids {
		if id == issueID {
			m.linkedIssues[testCaseID] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockRepository) GetTestCaseLinkedIssues(_ context.Context, testCaseID uuid.UUID) ([]*ent.Issue, error) {
	issueIDs := m.linkedIssues[testCaseID]
	result := make([]*ent.Issue, 0, len(issueIDs))
	for _, id := range issueIDs {
		if issue, ok := m.issues[id]; ok {
			result = append(result, issue)
		}
	}
	return result, nil
}

func (m *mockRepository) GetIssueLinkedTestCases(_ context.Context, issueID uuid.UUID) ([]*ent.TestCase, error) {
	result := make([]*ent.TestCase, 0)
	for tcID, issueIDs := range m.linkedIssues {
		for _, id := range issueIDs {
			if id == issueID {
				if tc, ok := m.testCases[tcID]; ok {
					result = append(result, tc)
				}
			}
		}
	}
	return result, nil
}

func (m *mockRepository) GetProjectIssuesWithTestCases(_ context.Context, _ uuid.UUID) ([]*ent.Issue, error) {
	result := make([]*ent.Issue, 0, len(m.issues))
	for _, issue := range m.issues {
		// загружаем связанные тест-кейсы
		linkedTCs := make([]*ent.TestCase, 0)
		for tcID, issueIDs := range m.linkedIssues {
			for _, iid := range issueIDs {
				if iid == issue.ID {
					if tc, ok := m.testCases[tcID]; ok {
						linkedTCs = append(linkedTCs, tc)
					}
				}
			}
		}
		issue.Edges.LinkedTestCases = linkedTCs
		result = append(result, issue)
	}
	return result, nil
}

func (m *mockRepository) CreateBugIssue(_ context.Context, projectID, _, testCaseID uuid.UUID, title, _, _ string) (*ent.Issue, error) {
	m.issueCounter++
	id := uuid.New()
	issue := &ent.Issue{
		ID:          id,
		IssueNumber: m.issueCounter,
		IssueKey:    "HUB-" + itoa(m.issueCounter),
		Title:       title,
		IssueType:   entissue.IssueType("bug"),
		Priority:    entissue.Priority("high"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Edges:       ent.IssueEdges{Project: &ent.Project{ID: projectID}},
	}
	m.issues[id] = issue
	m.linkedIssues[testCaseID] = append(m.linkedIssues[testCaseID], id)
	return issue, nil
}

func (m *mockRepository) GetLastResultForTestCase(_ context.Context, testCaseID uuid.UUID) (*ent.TestResult, error) {
	m.getLastResultSingleCalls++
	var latest *ent.TestResult
	for _, r := range m.results {
		if r.Edges.TestCase != nil && r.Edges.TestCase.ID == testCaseID {
			if latest == nil || r.CreatedAt.After(latest.CreatedAt) {
				latest = r
			}
		}
	}
	if latest == nil {
		return nil, &ent.NotFoundError{}
	}
	return latest, nil
}

func (m *mockRepository) GetLastResultsForTestCases(_ context.Context, testCaseIDs []uuid.UUID) (map[uuid.UUID]*ent.TestResult, error) {
	m.getLastResultsBatchCalls++
	out := make(map[uuid.UUID]*ent.TestResult, len(testCaseIDs))
	want := make(map[uuid.UUID]struct{}, len(testCaseIDs))
	for _, id := range testCaseIDs {
		want[id] = struct{}{}
	}
	for _, r := range m.results {
		if r.Edges.TestCase == nil {
			continue
		}
		tcID := r.Edges.TestCase.ID
		if _, ok := want[tcID]; !ok {
			continue
		}
		prev, exists := out[tcID]
		if !exists || r.CreatedAt.After(prev.CreatedAt) {
			out[tcID] = r
		}
	}
	return out, nil
}

func (m *mockRepository) ListTestResultsByRunIDs(_ context.Context, runIDs []uuid.UUID) ([]*ent.TestResult, error) {
	m.listResultsByRunIDsCalls++
	if len(runIDs) == 0 {
		return nil, nil
	}
	seen := make(map[uuid.UUID]struct{})
	var out []*ent.TestResult
	anyMapped := false
	for _, runID := range runIDs {
		ids, ok := m.runResults[runID]
		if !ok {
			continue
		}
		anyMapped = true
		run := m.runs[runID]
		for _, id := range ids {
			if _, dup := seen[id]; dup {
				continue
			}
			r, ok := m.results[id]
			if !ok {
				continue
			}
			seen[id] = struct{}{}
			if r.Edges.Run == nil && run != nil {
				r.Edges.Run = run
			}
			out = append(out, r)
		}
	}
	if anyMapped {
		return out, nil
	}
	want := make(map[uuid.UUID]struct{}, len(runIDs))
	for _, id := range runIDs {
		want[id] = struct{}{}
	}
	for _, r := range m.results {
		if r.Edges.Run == nil {
			continue
		}
		if _, ok := want[r.Edges.Run.ID]; ok {
			out = append(out, r)
		}
	}
	if len(out) > 0 {
		return out, nil
	}
	// Legacy mock: results без Edges.Run / runResults (GetDefectCategories).
	for _, r := range m.results {
		out = append(out, r)
	}
	return out, nil
}

func (m *mockRepository) AddTestCasesToRun(_ context.Context, runID uuid.UUID, caseIDs []uuid.UUID) error {
	for _, caseID := range caseIDs {
		id := uuid.New()
		r := &ent.TestResult{
			ID:     id,
			Status: "unknown",
			Edges: ent.TestResultEdges{
				TestCase: m.testCases[caseID],
			},
			CreatedAt: time.Now(),
		}
		m.results[id] = r
	}
	return nil
}

func (m *mockRepository) GetTestResultInRun(_ context.Context, runID, resultID uuid.UUID) (*ent.TestResult, error) {
	r, err := m.GetTestResult(context.Background(), resultID)
	if err != nil {
		return nil, err
	}
	if r.Edges.Run == nil || r.Edges.Run.ID != runID {
		return nil, &ent.NotFoundError{}
	}
	return r, nil
}

func (m *mockRepository) GetTestResult(_ context.Context, resultID uuid.UUID) (*ent.TestResult, error) {
	r, ok := m.results[resultID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return r, nil
}

func (m *mockRepository) ListTestRunResults(_ context.Context, _ uuid.UUID) ([]*ent.TestResult, error) {
	return m.listResultsSlice(), nil
}

func (m *mockRepository) AssignTestersToRun(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error {
	return nil
}

func (m *mockRepository) GetTestRunWithResults(_ context.Context, runID uuid.UUID) (*ent.TestRun, error) {
	run, ok := m.runs[runID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	results := m.listResultsSlice()
	run.Edges.Results = results
	return run, nil
}

func (m *mockRepository) GetTestRunWithResultsInProject(_ context.Context, projectID, runID uuid.UUID) (*ent.TestRun, error) {
	run, ok := m.runs[runID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if run.Edges.Project == nil || run.Edges.Project.ID != projectID {
		return nil, &ent.NotFoundError{}
	}
	results := m.listResultsSlice()
	run.Edges.Results = results
	return run, nil
}

func (m *mockRepository) AddFilesToResult(_ context.Context, _, _ uuid.UUID, _ []uuid.UUID) error {
	return nil
}

func (m *mockRepository) StoreAllureResults(_ context.Context, _ uuid.UUID, _ []tms.AllureResult) error {
	return nil
}

func (m *mockRepository) GetAllureResultsByRunID(_ context.Context, _ uuid.UUID) ([]tms.AllureResult, error) {
	return nil, nil
}

func (m *mockRepository) listResultsSlice() []*ent.TestResult {
	items := make([]*ent.TestResult, 0, len(m.results))
	for _, r := range m.results {
		items = append(items, r)
	}
	return items
}

// ---------------------------------------------------------------------------
// Вспомогательная функция: создаёт сервис с mock-зависимостями.
// ---------------------------------------------------------------------------

func itoa(n int) string {
	return string(rune('0' + n))
}

func orDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

// ---------------------------------------------------------------------------
// Вспомогательная функция: создаёт сервис с mock-зависимостями.
// ---------------------------------------------------------------------------

// TestCreateTestCase_Classic проверяет сценарий: создание test case classic.
func TestCreateTestCase_Classic(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()

	resp, err := svc.CreateTestCase(ctx, uuid.New(), uuid.New(), tms.CreateTestCaseRequest{
		Title:    "Авторизация по email",
		CaseType: "classic",
		Priority: "high",
		Steps: []schema.TestStep{
			{Order: 1, Action: "Ввести email", ExpectedResult: "Поле заполнено"},
			{Order: 2, Action: "Нажать Войти", ExpectedResult: "Magic link отправлен"},
		},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Contains(t, resp.CaseKey, "TC-")
	assert.Equal(t, "classic", resp.CaseType)
	assert.Equal(t, "high", resp.Priority)
	assert.Len(t, resp.Steps, 2)
	assert.Equal(t, 1, resp.Version)
}

// TestCreateTestCase_BDD проверяет сценарий: создание test case bdd.
func TestCreateTestCase_BDD(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()

	resp, err := svc.CreateTestCase(ctx, uuid.New(), uuid.New(), tms.CreateTestCaseRequest{
		Title:    "Пользователь входит через Telegram",
		CaseType: "bdd",
		Steps: []schema.TestStep{
			{Order: 1, Keyword: "Given", Action: "Пользователь на странице входа"},
			{Order: 2, Keyword: "When", Action: "Нажимает кнопку Telegram"},
			{Order: 3, Keyword: "Then", Action: "Открывается Telegram бот", ExpectedResult: "Бот отправляет код"},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "bdd", resp.CaseType)
	assert.Len(t, resp.Steps, 3)
	assert.Equal(t, "Given", resp.Steps[0].Keyword)
	assert.Equal(t, "When", resp.Steps[1].Keyword)
	assert.Equal(t, "Then", resp.Steps[2].Keyword)
}

// TestUpdateTestCase_CreatesVersion проверяет сценарий: обновление test case creates version.
func TestUpdateTestCase_CreatesVersion(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	created, err := svc.CreateTestCase(ctx, projectID, uuid.New(), tms.CreateTestCaseRequest{
		Title: "Исходный заголовок",
	})
	require.NoError(t, err)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	newTitle := "Обновлённый заголовок"
	updated, err := svc.UpdateTestCase(ctx, projectID, created.CaseKey, uuid.New(), tms.UpdateTestCaseRequest{
		Title:   &newTitle,
		Version: 1,
	})

	require.NoError(t, err)
	assert.Equal(t, "Обновлённый заголовок", updated.Title)
	assert.Equal(t, 2, updated.Version)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	assert.NotNil(t, repo.lastVersion)
	assert.Equal(t, 1, repo.lastVersion.VersionNumber)
	assert.Equal(t, "Исходный заголовок", repo.lastVersion.Snapshot["title"])
}

// TestRollbackTestCase проверяет сценарий: rollback test case.
func TestRollbackTestCase(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	created, err := svc.CreateTestCase(ctx, projectID, uuid.New(), tms.CreateTestCaseRequest{
		Title:    "Версия 1",
		Priority: "high",
	})
	require.NoError(t, err)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	newTitle := "Версия 2"
	_, err = svc.UpdateTestCase(ctx, projectID, created.CaseKey, uuid.New(), tms.UpdateTestCaseRequest{
		Title:   &newTitle,
		Version: 1,
	})
	require.NoError(t, err)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	versionID := repo.lastVersion.ID

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	rolledBack, err := svc.RollbackTestCase(ctx, projectID, created.CaseKey, versionID, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "Версия 1", rolledBack.Title)
}

// TestUpdateRunStatus_ValidTransitions проверяет сценарий: обновление run status корректный transitions.
func TestUpdateRunStatus_ValidTransitions(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	run, err := svc.CreateTestRun(ctx, projectID, uuid.New(), tms.CreateTestRunRequest{
		Title: "Прогон 1",
	})
	require.NoError(t, err)
	assert.Equal(t, "created", run.Status)

	runID, _ := uuid.Parse(run.ID)

	// created -> in_progress (создан -> в работе)
	updated, err := svc.UpdateRunStatus(ctx, projectID, runID, "in_progress")
	require.NoError(t, err)
	assert.Equal(t, "in_progress", updated.Status)

	// in_progress -> completed (в работе -> завершён)
	completed, err := svc.UpdateRunStatus(ctx, projectID, runID, "completed")
	require.NoError(t, err)
	assert.Equal(t, "completed", completed.Status)
}

// TestUpdateRunStatus_InvalidTransition проверяет отклонение при обновление run status некорректный transition.
func TestUpdateRunStatus_InvalidTransition(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	run, err := svc.CreateTestRun(ctx, projectID, uuid.New(), tms.CreateTestRunRequest{
		Title: "Прогон 2",
	})
	require.NoError(t, err)

	runID, _ := uuid.Parse(run.ID)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	_, err = svc.UpdateRunStatus(ctx, projectID, runID, "completed")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "недопустимый переход статуса")
}

// TestReopenTestRun проверяет сценарий: reopen test run.
func TestReopenTestRun(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	run, err := svc.CreateTestRun(ctx, projectID, uuid.New(), tms.CreateTestRunRequest{
		Title: "Прогон для переоткрытия",
	})
	require.NoError(t, err)
	runID, _ := uuid.Parse(run.ID)

	// created -> in_progress -> completed (создан -> в работе -> завершён)
	_, err = svc.UpdateRunStatus(ctx, projectID, runID, "in_progress")
	require.NoError(t, err)
	_, err = svc.UpdateRunStatus(ctx, projectID, runID, "completed")
	require.NoError(t, err)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	reopened, err := svc.ReopenTestRun(ctx, projectID, runID)
	require.NoError(t, err)
	assert.Equal(t, "in_progress", reopened.Status)
}

// TestAddTestCasesToRun проверяет сценарий: add test cases to run.
func TestAddTestCasesToRun(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	tc1, err := svc.CreateTestCase(ctx, projectID, uuid.New(), tms.CreateTestCaseRequest{
		Title: "Кейс 1",
	})
	require.NoError(t, err)
	tc2, err := svc.CreateTestCase(ctx, projectID, uuid.New(), tms.CreateTestCaseRequest{
		Title: "Кейс 2",
	})
	require.NoError(t, err)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	run, err := svc.CreateTestRun(ctx, projectID, uuid.New(), tms.CreateTestRunRequest{
		Title: "Прогон с кейсами",
	})
	require.NoError(t, err)
	runID, _ := uuid.Parse(run.ID)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	detail, err := svc.AddTestCasesToRun(ctx, projectID, runID, tms.AddTestCasesToRunRequest{
		CaseIDs: []string{tc1.ID, tc2.ID},
	})
	require.NoError(t, err)
	assert.Len(t, detail.Results, 2)
}

// TestExecuteTestResult_UpdatesStats проверяет сценарий: execute test result updates stats.
func TestExecuteTestResult_UpdatesStats(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	run, err := svc.CreateTestRun(ctx, projectID, uuid.New(), tms.CreateTestRunRequest{
		Title: "Прогон для выполнения",
	})
	require.NoError(t, err)
	runID, _ := uuid.Parse(run.ID)

	_, err = svc.UpdateRunStatus(ctx, projectID, runID, "in_progress")
	require.NoError(t, err)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	resultID := uuid.New()
	repo.results[resultID] = &ent.TestResult{
		ID:        resultID,
		Status:    "unknown",
		CreatedAt: time.Now(),
	}

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	passedStatus := "passed"
	comment := "Тест пройден"
	resp, err := svc.ExecuteTestResult(ctx, projectID, runID, resultID, uuid.New(), tms.UpdateTestResultRequest{
		Status:  &passedStatus,
		Comment: &comment,
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	updatedRun := repo.runs[runID]
	assert.Equal(t, 1, updatedRun.TotalCount)
	assert.Equal(t, 1, updatedRun.PassedCount)
}

// TestCompleteTestRun проверяет сценарий: complete test run.
func TestCompleteTestRun(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	// Создаём и стартуем прогон
	run, err := svc.CreateTestRun(ctx, projectID, uuid.New(), tms.CreateTestRunRequest{
		Title: "Прогон для завершения",
	})
	require.NoError(t, err)
	runID, _ := uuid.Parse(run.ID)

	_, err = svc.StartTestRun(ctx, projectID, runID)
	require.NoError(t, err)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	completed, err := svc.CompleteTestRun(ctx, projectID, runID)
	require.NoError(t, err)
	assert.Equal(t, "completed", completed.Status)
}

// TestLinkTestCaseToIssue проверяет сценарий: link test case to issue.
func TestLinkTestCaseToIssue(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	tc, err := svc.CreateTestCase(ctx, projectID, uuid.New(), tms.CreateTestCaseRequest{
		Title: "Кейс для привязки",
	})
	require.NoError(t, err)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	issueID := uuid.New()
	repo.issues[issueID] = &ent.Issue{
		ID:       issueID,
		IssueKey: "HUB-100",
		Title:    "Задача для привязки",
		Edges:    ent.IssueEdges{Project: &ent.Project{ID: projectID}},
	}

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	err = svc.LinkTestCaseToIssue(ctx, projectID, tc.CaseKey, issueID)
	require.NoError(t, err)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	links, err := svc.GetTestCaseLinks(ctx, projectID, tc.CaseKey)
	require.NoError(t, err)
	assert.Len(t, links, 1)
	assert.Equal(t, "HUB-100", links[0].IssueKey)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	err = svc.UnlinkTestCaseFromIssue(ctx, projectID, tc.CaseKey, issueID)
	require.NoError(t, err)

	links, err = svc.GetTestCaseLinks(ctx, projectID, tc.CaseKey)
	require.NoError(t, err)
	assert.Len(t, links, 0)
}

// TestCreateBugFromFailedResult проверяет сценарий: создание bug from failed result.
func TestCreateBugFromFailedResult(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	tc, err := svc.CreateTestCase(ctx, projectID, uuid.New(), tms.CreateTestCaseRequest{
		Title: "Кейс с провалом",
	})
	require.NoError(t, err)
	tcID, _ := uuid.Parse(tc.ID)

	runID := uuid.New()
	repo.runs[runID] = &ent.TestRun{
		ID: runID,
		Edges: ent.TestRunEdges{
			Project: &ent.Project{ID: projectID},
		},
	}

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	resultID := uuid.New()
	repo.results[resultID] = &ent.TestResult{
		ID:           resultID,
		Status:       "failed",
		ErrorMessage: "NullPointerException",
		CreatedAt:    time.Now(),
		Edges: ent.TestResultEdges{
			TestCase: repo.testCases[tcID],
			Run:      repo.runs[runID],
		},
	}

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	bug, err := svc.CreateBugFromFailedResult(ctx, projectID, runID, uuid.New(), resultID, tms.CreateBugRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, bug.ID)
	assert.Contains(t, bug.Title, "FAILED:")
	assert.Contains(t, bug.Title, "Кейс с провалом")
}

// TestGetCoverageReport проверяет сценарий: получение покрытие report.
func TestGetCoverageReport(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	tc, err := svc.CreateTestCase(ctx, projectID, uuid.New(), tms.CreateTestCaseRequest{
		Title: "Кейс покрытия",
	})
	require.NoError(t, err)
	tcID, _ := uuid.Parse(tc.ID)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	issueID := uuid.New()
	repo.issues[issueID] = &ent.Issue{
		ID:       issueID,
		IssueKey: "HUB-200",
		Title:    "Задача покрытия",
	}
	repo.linkedIssues[tcID] = []uuid.UUID{issueID}

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	resultID := uuid.New()
	repo.results[resultID] = &ent.TestResult{
		ID:     resultID,
		Status: "passed",
		Edges: ent.TestResultEdges{
			TestCase: repo.testCases[tcID],
		},
		CreatedAt: time.Now(),
	}

	report, err := svc.GetCoverageReport(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, report.Issues, 1)
	assert.Equal(t, "HUB-200", report.Issues[0].IssueKey)
	assert.Len(t, report.Issues[0].TestCases, 1)
	assert.Equal(t, "passed", report.Issues[0].TestCases[0].LastStatus)
}

// TestImportFromCSV проверяет сценарий: импорт from csv.
func TestImportFromCSV(t *testing.T) {
	csvData := "Title,Priority,Status\nКейс 1,high,active\nКейс 2,medium,draft\nКейс 3,low,active\n"
	reader := strings.NewReader(csvData)

	inputs, err := tms.ImportFromCSV(reader, nil)
	require.NoError(t, err)
	assert.Len(t, inputs, 3)
	assert.Equal(t, "Кейс 1", inputs[0].Title)
	assert.Equal(t, "high", inputs[0].Priority)
	assert.Equal(t, "active", inputs[0].Status)
	assert.Equal(t, "Кейс 3", inputs[2].Title)
}

// TestAutoDetectMapping проверяет сценарий: auto detect mapping.
func TestAutoDetectMapping(t *testing.T) {
	headers := []string{"Title", "Описание", "Priority", "Шаги", "Тип"}
	mapping := tms.AutoDetectMapping(headers)

	assert.Equal(t, "title", mapping["Title"])
	assert.Equal(t, "description", mapping["Описание"])
	assert.Equal(t, "priority", mapping["Priority"])
	assert.Equal(t, "steps", mapping["Шаги"])
	assert.Equal(t, "case_type", mapping["Тип"])
}

// TestExportToCSV проверяет сценарий: экспорт to csv.
func TestExportToCSV(t *testing.T) {
	cases := []tms.TestCaseExportData{
		{Key: "HUB-TC-1", Title: "Тест 1", Type: "classic", Priority: "high", Status: "active"},
		{Key: "HUB-TC-2", Title: "Тест 2", Type: "bdd", Priority: "medium", Status: "draft"},
	}

	var buf bytes.Buffer
	err := tms.ExportTestCasesToCSV(&buf, cases)
	require.NoError(t, err)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	require.NoError(t, err)
	assert.Len(t, records, 3) // заголовок + 2 строки
	assert.Equal(t, "HUB-TC-1", records[1][0])
	assert.Equal(t, "HUB-TC-2", records[2][0])
}

// TestValidateImportData проверяет сценарий: validate импорт data.
func TestValidateImportData(t *testing.T) {
	inputs := []tms.CreateTestCaseInput{
		{Title: "Валидный кейс", Priority: "high"},
		{Title: "", Priority: "medium"}, // нет названия
		{Title: "Кейс с неверным приоритетом", Priority: "invalid"},
	}

	errors := tms.ValidateImportData(inputs)
	assert.Len(t, errors, 2)
	assert.Equal(t, 3, errors[0].Row) // строка 3 (индекс 1 + 2)
	assert.Equal(t, "title", errors[0].Field)
	assert.Equal(t, 4, errors[1].Row) // строка 4 (индекс 2 + 2)
	assert.Equal(t, "priority", errors[1].Field)
}

// TestStartTestRun_InvalidTransition проверяет отклонение при start test run некорректный transition.
func TestStartTestRun_InvalidTransition(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	run, err := svc.CreateTestRun(ctx, projectID, uuid.New(), tms.CreateTestRunRequest{
		Title: "Прогон",
	})
	require.NoError(t, err)
	runID, _ := uuid.Parse(run.ID)

	_, err = svc.StartTestRun(ctx, projectID, runID)
	require.NoError(t, err)
	_, err = svc.CompleteTestRun(ctx, projectID, runID)
	require.NoError(t, err)

	// Комментарий к тесту (восстановлен после повреждения кодировки).
	_, err = svc.StartTestRun(ctx, projectID, runID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "недопустимый переход статуса")
}
