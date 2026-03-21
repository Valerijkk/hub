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
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/ent/testrun"
)

// ---------------------------------------------------------------------------
// Mock Repository
// ---------------------------------------------------------------------------

type mockRepository struct {
	testCases map[uuid.UUID]*ent.TestCase
	versions  map[uuid.UUID]*ent.TestCaseVersion
	runs      map[uuid.UUID]*ent.TestRun
	results   map[uuid.UUID]*ent.TestResult
	suites    map[uuid.UUID]*ent.TestSuite
	components map[uuid.UUID]*ent.Component

	caseCounter int
	lastVersion *ent.TestCaseVersion
}

func newMockRepo() *mockRepository {
	return &mockRepository{
		testCases:  make(map[uuid.UUID]*ent.TestCase),
		versions:   make(map[uuid.UUID]*ent.TestCaseVersion),
		runs:       make(map[uuid.UUID]*ent.TestRun),
		results:    make(map[uuid.UUID]*ent.TestResult),
		suites:     make(map[uuid.UUID]*ent.TestSuite),
		components: make(map[uuid.UUID]*ent.Component),
	}
}

func (m *mockRepository) CreateTestCase(_ context.Context, _ uuid.UUID, input CreateTestCaseInput) (*ent.TestCase, error) {
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

func (m *mockRepository) FindTestCaseByKey(_ context.Context, caseKey string) (*ent.TestCase, error) {
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

func (m *mockRepository) FindTestCaseByFullName(_ context.Context, _ uuid.UUID, _ string) (*ent.TestCase, error) {
	return nil, &ent.NotFoundError{}
}

func (m *mockRepository) FindTestCaseByClassMethod(_ context.Context, _ uuid.UUID, _, _ string) (*ent.TestCase, error) {
	return nil, &ent.NotFoundError{}
}

func (m *mockRepository) ListTestCases(_ context.Context, _ uuid.UUID, _ ListTestCasesParams) ([]*ent.TestCase, int, error) {
	items := make([]*ent.TestCase, 0, len(m.testCases))
	for _, tc := range m.testCases {
		items = append(items, tc)
	}
	return items, len(items), nil
}

func (m *mockRepository) UpdateTestCase(_ context.Context, id uuid.UUID, input UpdateTestCaseInput) (*ent.TestCase, error) {
	tc, ok := m.testCases[id]
	if !ok {
		return nil, &ent.NotFoundError{}
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

func (m *mockRepository) CreateTestCaseVersion(_ context.Context, testCaseID, userID uuid.UUID, versionNumber int, snapshot map[string]interface{}, description string) (*ent.TestCaseVersion, error) {
	id := uuid.New()
	v := &ent.TestCaseVersion{
		ID:                id,
		VersionNumber:     versionNumber,
		Snapshot:          snapshot,
		ChangeDescription: description,
		CreatedAt:         time.Now(),
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

func (m *mockRepository) GetTestCaseVersion(_ context.Context, versionID uuid.UUID) (*ent.TestCaseVersion, error) {
	v, ok := m.versions[versionID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return v, nil
}

func (m *mockRepository) CreateSuite(_ context.Context, _ uuid.UUID, input CreateSuiteInput) (*ent.TestSuite, error) {
	id := uuid.New()
	s := &ent.TestSuite{ID: id, Name: input.Name, CreatedAt: time.Now(), UpdatedAt: time.Now()}
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

func (m *mockRepository) ListSuites(_ context.Context, _ uuid.UUID) ([]*ent.TestSuite, error) {
	items := make([]*ent.TestSuite, 0, len(m.suites))
	for _, s := range m.suites {
		items = append(items, s)
	}
	return items, nil
}

func (m *mockRepository) UpdateSuite(_ context.Context, id uuid.UUID, input UpdateSuiteInput) (*ent.TestSuite, error) {
	s, ok := m.suites[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if input.Name != nil {
		s.Name = *input.Name
	}
	return s, nil
}

func (m *mockRepository) DeleteSuite(_ context.Context, id uuid.UUID) error {
	delete(m.suites, id)
	return nil
}

func (m *mockRepository) MoveSuite(_ context.Context, id uuid.UUID, _ *uuid.UUID) (*ent.TestSuite, error) {
	s, ok := m.suites[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return s, nil
}

func (m *mockRepository) CreateComponent(_ context.Context, _ uuid.UUID, name, description string) (*ent.Component, error) {
	id := uuid.New()
	c := &ent.Component{ID: id, Name: name, Description: description, CreatedAt: time.Now()}
	m.components[id] = c
	return c, nil
}

func (m *mockRepository) ListComponents(_ context.Context, _ uuid.UUID) ([]*ent.Component, error) {
	items := make([]*ent.Component, 0, len(m.components))
	for _, c := range m.components {
		items = append(items, c)
	}
	return items, nil
}

func (m *mockRepository) DeleteComponent(_ context.Context, id uuid.UUID) error {
	delete(m.components, id)
	return nil
}

func (m *mockRepository) CreateTestRun(_ context.Context, _ uuid.UUID, input CreateTestRunInput) (*ent.TestRun, error) {
	id := uuid.New()
	runType := testrun.RunType(orDefault(input.RunType, "manual"))
	run := &ent.TestRun{
		ID:        id,
		Title:     input.Title,
		RunType:   runType,
		Status:    testrun.StatusCreated,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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

func (m *mockRepository) ListTestRuns(_ context.Context, _ uuid.UUID, _ ListTestRunsParams) ([]*ent.TestRun, int, error) {
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

func (m *mockRepository) UpdateTestRunStats(_ context.Context, id uuid.UUID, stats TestRunStats) (*ent.TestRun, error) {
	run, ok := m.runs[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	run.TotalCount = stats.TotalCount
	run.PassedCount = stats.PassedCount
	return run, nil
}

func (m *mockRepository) CreateTestResult(_ context.Context, _ uuid.UUID, input CreateTestResultInput) (*ent.TestResult, error) {
	id := uuid.New()
	r := &ent.TestResult{ID: id, Status: "passed", CreatedAt: time.Now()}
	m.results[id] = r
	return r, nil
}

func (m *mockRepository) UpdateTestResult(_ context.Context, id uuid.UUID, _ UpdateTestResultInput) (*ent.TestResult, error) {
	r, ok := m.results[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return r, nil
}

func (m *mockRepository) ListTestResults(_ context.Context, _ uuid.UUID) ([]*ent.TestResult, error) {
	items := make([]*ent.TestResult, 0, len(m.results))
	for _, r := range m.results {
		items = append(items, r)
	}
	return items, nil
}

func (m *mockRepository) BatchCreateTestResults(_ context.Context, _ uuid.UUID, inputs []CreateTestResultInput) ([]*ent.TestResult, error) {
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

func (m *mockRepository) LinkTestCaseToIssue(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (m *mockRepository) UnlinkTestCaseFromIssue(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

// ---------------------------------------------------------------------------
// Хелперы
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
// Тесты
// ---------------------------------------------------------------------------

func TestCreateTestCase_Classic(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo)
	ctx := context.Background()

	resp, err := svc.CreateTestCase(ctx, uuid.New(), uuid.New(), CreateTestCaseRequest{
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

func TestCreateTestCase_BDD(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo)
	ctx := context.Background()

	resp, err := svc.CreateTestCase(ctx, uuid.New(), uuid.New(), CreateTestCaseRequest{
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

func TestUpdateTestCase_CreatesVersion(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo)
	ctx := context.Background()

	// Создаём тест-кейс
	created, err := svc.CreateTestCase(ctx, uuid.New(), uuid.New(), CreateTestCaseRequest{
		Title: "Исходный заголовок",
	})
	require.NoError(t, err)

	// Обновляем
	newTitle := "Обновлённый заголовок"
	updated, err := svc.UpdateTestCase(ctx, created.CaseKey, uuid.New(), UpdateTestCaseRequest{
		Title:   &newTitle,
		Version: 1,
	})

	require.NoError(t, err)
	assert.Equal(t, "Обновлённый заголовок", updated.Title)
	assert.Equal(t, 2, updated.Version)

	// Проверяем, что версия была создана
	assert.NotNil(t, repo.lastVersion)
	assert.Equal(t, 1, repo.lastVersion.VersionNumber)
	assert.Equal(t, "Исходный заголовок", repo.lastVersion.Snapshot["title"])
}

func TestRollbackTestCase(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo)
	ctx := context.Background()

	// Создаём тест-кейс
	created, err := svc.CreateTestCase(ctx, uuid.New(), uuid.New(), CreateTestCaseRequest{
		Title:    "Версия 1",
		Priority: "high",
	})
	require.NoError(t, err)

	// Обновляем (создаётся version snapshot)
	newTitle := "Версия 2"
	_, err = svc.UpdateTestCase(ctx, created.CaseKey, uuid.New(), UpdateTestCaseRequest{
		Title:   &newTitle,
		Version: 1,
	})
	require.NoError(t, err)

	// Берём ID версии
	versionID := repo.lastVersion.ID

	// Откатываем
	rolledBack, err := svc.RollbackTestCase(ctx, created.CaseKey, versionID, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "Версия 1", rolledBack.Title)
}

func TestUpdateRunStatus_ValidTransitions(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo)
	ctx := context.Background()

	// Создаём прогон (status = created)
	run, err := svc.CreateTestRun(ctx, uuid.New(), uuid.New(), CreateTestRunRequest{
		Title: "Прогон 1",
	})
	require.NoError(t, err)
	assert.Equal(t, "created", run.Status)

	runID, _ := uuid.Parse(run.ID)

	// created -> in_progress
	updated, err := svc.UpdateRunStatus(ctx, runID, "in_progress")
	require.NoError(t, err)
	assert.Equal(t, "in_progress", updated.Status)

	// in_progress -> completed
	completed, err := svc.UpdateRunStatus(ctx, runID, "completed")
	require.NoError(t, err)
	assert.Equal(t, "completed", completed.Status)
}

func TestUpdateRunStatus_InvalidTransition(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo)
	ctx := context.Background()

	// Создаём прогон (status = created)
	run, err := svc.CreateTestRun(ctx, uuid.New(), uuid.New(), CreateTestRunRequest{
		Title: "Прогон 2",
	})
	require.NoError(t, err)

	runID, _ := uuid.Parse(run.ID)

	// created -> completed (недопустимо, минуя in_progress)
	_, err = svc.UpdateRunStatus(ctx, runID, "completed")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "недопустимый переход статуса")
}

func TestReopenTestRun(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo)
	ctx := context.Background()

	// Создаём и завершаем прогон
	run, err := svc.CreateTestRun(ctx, uuid.New(), uuid.New(), CreateTestRunRequest{
		Title: "Прогон для переоткрытия",
	})
	require.NoError(t, err)
	runID, _ := uuid.Parse(run.ID)

	// created -> in_progress -> completed
	_, err = svc.UpdateRunStatus(ctx, runID, "in_progress")
	require.NoError(t, err)
	_, err = svc.UpdateRunStatus(ctx, runID, "completed")
	require.NoError(t, err)

	// Переоткрываем
	reopened, err := svc.ReopenTestRun(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, "in_progress", reopened.Status)
}
