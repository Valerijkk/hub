package tms

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/schema"
)

// ---------------------------------------------------------------------------
// Узкие доменные интерфейсы (ISP — Interface Segregation Principle)
// ---------------------------------------------------------------------------

// TestCaseRepository управляет CRUD тест-кейсов, поиском и версионированием.
type TestCaseRepository interface {
	// CreateTestCase создаёт новый тест-кейс в проекте.
	CreateTestCase(ctx context.Context, projectID uuid.UUID, input CreateTestCaseInput) (*ent.TestCase, error)
	// GetTestCase возвращает тест-кейс по ID с подгруженными связями.
	GetTestCase(ctx context.Context, id uuid.UUID) (*ent.TestCase, error)
	// FindTestCaseByKey возвращает тест-кейс по ключу только внутри projectID (A3-003).
	FindTestCaseByKey(ctx context.Context, projectID uuid.UUID, caseKey string) (*ent.TestCase, error)
	// FindTestCaseByExternalID возвращает тест-кейс по внешнему ID аннотации.
	FindTestCaseByExternalID(ctx context.Context, projectID uuid.UUID, externalID string) (*ent.TestCase, error)
	// FindTestCaseByAutomationID возвращает тест-кейс по automation_id (явная привязка автотестов).
	FindTestCaseByAutomationID(ctx context.Context, projectID uuid.UUID, automationID string) (*ent.TestCase, error)
	// FindTestCaseByFullName возвращает тест-кейс по полному имени (Allure matching).
	FindTestCaseByFullName(ctx context.Context, projectID uuid.UUID, fullName string) (*ent.TestCase, error)
	// FindTestCaseByClassMethod возвращает тест-кейс по имени класса и метода.
	FindTestCaseByClassMethod(ctx context.Context, projectID uuid.UUID, className, methodName string) (*ent.TestCase, error)
	// ListTestCases возвращает пагинированный список тест-кейсов с фильтрацией.
	ListTestCases(ctx context.Context, projectID uuid.UUID, params ListTestCasesParams) ([]*ent.TestCase, int, error)
	// UpdateTestCase обновляет тест-кейс.
	UpdateTestCase(ctx context.Context, id uuid.UUID, input UpdateTestCaseInput) (*ent.TestCase, error)
	// DeleteTestCase удаляет тест-кейс по ID.
	DeleteTestCase(ctx context.Context, id uuid.UUID) error
	// IncrementTestCaseCounter атомарно увеличивает счётчик тест-кейсов проекта.
	IncrementTestCaseCounter(ctx context.Context, projectID uuid.UUID) (int, error)
}

// TestCaseVersionRepository управляет версиями (снимками) тест-кейсов.
type TestCaseVersionRepository interface {
	// CreateTestCaseVersion создаёт снимок версии тест-кейса.
	CreateTestCaseVersion(ctx context.Context, testCaseID, userID uuid.UUID, versionNumber int, snapshot map[string]interface{}, description string) (*ent.TestCaseVersion, error)
	// ListTestCaseVersions возвращает все версии тест-кейса.
	ListTestCaseVersions(ctx context.Context, testCaseID uuid.UUID) ([]*ent.TestCaseVersion, error)
	// GetTestCaseVersion возвращает версию по ID только если она принадлежит testCaseID (A3-007).
	GetTestCaseVersion(ctx context.Context, testCaseID, versionID uuid.UUID) (*ent.TestCaseVersion, error)
}

// SuiteRepository управляет папками/секциями тест-кейсов.
type SuiteRepository interface {
	// CreateSuite создаёт новую папку/секцию тест-кейсов.
	CreateSuite(ctx context.Context, projectID uuid.UUID, input CreateSuiteInput) (*ent.TestSuite, error)
	// GetSuite возвращает папку по ID.
	GetSuite(ctx context.Context, id uuid.UUID) (*ent.TestSuite, error)
	// ListSuites возвращает все папки проекта (для построения дерева).
	ListSuites(ctx context.Context, projectID uuid.UUID) ([]*ent.TestSuite, error)
	// UpdateSuite обновляет папку.
	UpdateSuite(ctx context.Context, id uuid.UUID, input UpdateSuiteInput) (*ent.TestSuite, error)
	// DeleteSuite удаляет папку по ID.
	DeleteSuite(ctx context.Context, id uuid.UUID) error
	// MoveSuite перемещает папку к другому родителю.
	MoveSuite(ctx context.Context, id uuid.UUID, parentID *uuid.UUID) (*ent.TestSuite, error)
}

// ComponentRepository управляет компонентами проекта.
type ComponentRepository interface {
	// CreateComponent создаёт новый компонент в проекте.
	CreateComponent(ctx context.Context, projectID uuid.UUID, name, description string) (*ent.Component, error)
	// ListComponents возвращает все компоненты проекта.
	ListComponents(ctx context.Context, projectID uuid.UUID) ([]*ent.Component, error)
	// DeleteComponent удаляет компонент по ID.
	DeleteComponent(ctx context.Context, id uuid.UUID) error
}

// TestRunRepository управляет прогонами тестов и назначением тестировщиков.
type TestRunRepository interface {
	// CreateTestRun создаёт новый прогон тестов.
	CreateTestRun(ctx context.Context, projectID uuid.UUID, input CreateTestRunInput) (*ent.TestRun, error)
	// GetTestRun возвращает прогон по ID с подгруженными связями (внутренний, без project-scope).
	GetTestRun(ctx context.Context, id uuid.UUID) (*ent.TestRun, error)
	// GetTestRunInProject возвращает прогон только если он принадлежит projectID (A3-012).
	GetTestRunInProject(ctx context.Context, projectID, id uuid.UUID) (*ent.TestRun, error)
	// ListTestRuns возвращает пагинированный список прогонов проекта.
	ListTestRuns(ctx context.Context, projectID uuid.UUID, params ListTestRunsParams) ([]*ent.TestRun, int, error)
	// UpdateTestRunStatus обновляет статус прогона.
	UpdateTestRunStatus(ctx context.Context, id uuid.UUID, status string, timestamp *time.Time) (*ent.TestRun, error)
	// ReopenTestRun переоткрывает завершённый прогон.
	ReopenTestRun(ctx context.Context, id uuid.UUID) (*ent.TestRun, error)
	// UpdateTestRunStats обновляет счётчики результатов прогона.
	UpdateTestRunStats(ctx context.Context, id uuid.UUID, stats TestRunStats) (*ent.TestRun, error)
	// AddTestCasesToRun массово создаёт TestResult записи для каждого тест-кейса.
	AddTestCasesToRun(ctx context.Context, runID uuid.UUID, caseIDs []uuid.UUID) error
	// AssignTestersToRun устанавливает M2M связь прогона с назначенными тестировщиками.
	AssignTestersToRun(ctx context.Context, runID uuid.UUID, userIDs []uuid.UUID) error
	// GetTestRunWithResults возвращает прогон с результатами, тест-кейсами, файлами и назначенными.
	GetTestRunWithResults(ctx context.Context, runID uuid.UUID) (*ent.TestRun, error)
	// GetTestRunWithResultsInProject — то же, но только внутри projectID (A3-012).
	GetTestRunWithResultsInProject(ctx context.Context, projectID, runID uuid.UUID) (*ent.TestRun, error)
}

// TestResultRepository управляет результатами выполнения тестов.
type TestResultRepository interface {
	// CreateTestResult создаёт результат выполнения теста.
	CreateTestResult(ctx context.Context, runID uuid.UUID, input CreateTestResultInput) (*ent.TestResult, error)
	// UpdateTestResult обновляет результат теста только внутри runID (A3-008).
	UpdateTestResult(ctx context.Context, runID, id uuid.UUID, input UpdateTestResultInput) (*ent.TestResult, error)
	// ListTestResults возвращает все результаты прогона.
	ListTestResults(ctx context.Context, runID uuid.UUID) ([]*ent.TestResult, error)
	// BatchCreateTestResults массово создаёт результаты (для импорта Allure).
	BatchCreateTestResults(ctx context.Context, runID uuid.UUID, inputs []CreateTestResultInput) ([]*ent.TestResult, error)
	// GetTestResult возвращает результат по ID с подгруженными связями.
	GetTestResult(ctx context.Context, resultID uuid.UUID) (*ent.TestResult, error)
	// ListTestRunResults возвращает результаты прогона с подгруженными тест-кейсами и файлами.
	ListTestRunResults(ctx context.Context, runID uuid.UUID) ([]*ent.TestResult, error)
	// AddFilesToResult привязывает файлы к результату теста.
	AddFilesToResult(ctx context.Context, resultID uuid.UUID, fileIDs []uuid.UUID) error
	// GetLastResultForTestCase возвращает последний результат для тест-кейса.
	GetLastResultForTestCase(ctx context.Context, testCaseID uuid.UUID) (*ent.TestResult, error)
}

// AllureRepository управляет сырыми результатами Allure в MongoDB.
type AllureRepository interface {
	// StoreAllureResults сохраняет сырые результаты Allure в MongoDB.
	StoreAllureResults(ctx context.Context, runID uuid.UUID, results []AllureResult) error
	// GetAllureResultsByRunID возвращает сырые результаты Allure для прогона.
	GetAllureResultsByRunID(ctx context.Context, runID uuid.UUID) ([]AllureResult, error)
}

// EnvironmentRepository управляет тестовыми окружениями проекта.
type EnvironmentRepository interface {
	// CreateEnvironment создаёт тестовое окружение в проекте.
	CreateEnvironment(ctx context.Context, projectID uuid.UUID, input CreateEnvironmentInput) (*ent.Environment, error)
	// GetEnvironment возвращает окружение по ID.
	GetEnvironment(ctx context.Context, envID uuid.UUID) (*ent.Environment, error)
	// ListEnvironmentsByProject возвращает все окружения проекта.
	ListEnvironmentsByProject(ctx context.Context, projectID uuid.UUID) ([]*ent.Environment, error)
	// UpdateEnvironment обновляет тестовое окружение.
	UpdateEnvironment(ctx context.Context, envID uuid.UUID, input UpdateEnvironmentInput) (*ent.Environment, error)
	// DeleteEnvironment удаляет тестовое окружение.
	DeleteEnvironment(ctx context.Context, envID uuid.UUID) error
	// GetLatestTestRunForEnvironment возвращает последний прогон проекта в указанном окружении.
	// Если прогонов нет — возвращает (nil, nil) без ошибки.
	GetLatestTestRunForEnvironment(ctx context.Context, projectID, envID uuid.UUID) (*ent.TestRun, error)
}

// IssueLinkRepository управляет привязкой тест-кейсов к задачам трекера.
type IssueLinkRepository interface {
	// LinkTestCaseToIssue привязывает тест-кейс к задаче.
	LinkTestCaseToIssue(ctx context.Context, testCaseID, issueID uuid.UUID) error
	// UnlinkTestCaseFromIssue отвязывает тест-кейс от задачи.
	UnlinkTestCaseFromIssue(ctx context.Context, testCaseID, issueID uuid.UUID) error
	// GetTestCaseLinkedIssues возвращает задачи, привязанные к тест-кейсу.
	GetTestCaseLinkedIssues(ctx context.Context, testCaseID uuid.UUID) ([]*ent.Issue, error)
	// GetIssueLinkedTestCases возвращает тест-кейсы, привязанные к задаче (обратная связь).
	GetIssueLinkedTestCases(ctx context.Context, issueID uuid.UUID) ([]*ent.TestCase, error)
	// GetProjectIssuesWithTestCases возвращает задачи проекта с привязанными тест-кейсами.
	GetProjectIssuesWithTestCases(ctx context.Context, projectID uuid.UUID) ([]*ent.Issue, error)
	// CreateBugIssue создаёт задачу типа bug и привязывает к тест-кейсу.
	CreateBugIssue(ctx context.Context, projectID, userID, testCaseID uuid.UUID, title, description, priority string) (*ent.Issue, error)
}

// ---------------------------------------------------------------------------
// Композитный интерфейс (обратная совместимость)
// ---------------------------------------------------------------------------

// Repository — композитный интерфейс для обратной совместимости.
// Позволяет тестировать сервисный слой с помощью mock-реализаций.
type Repository interface {
	TestCaseRepository
	TestCaseVersionRepository
	SuiteRepository
	ComponentRepository
	TestRunRepository
	TestResultRepository
	AllureRepository
	IssueLinkRepository
	EnvironmentRepository
}

// ---------------------------------------------------------------------------
// Входные структуры репозитория
// ---------------------------------------------------------------------------

// CreateTestCaseInput содержит данные для создания тест-кейса.
type CreateTestCaseInput struct {
	Title         string
	Description   string
	Precondition  string
	Postcondition string
	CaseType      string
	Priority      string
	Status        string
	Steps         []schema.TestStep
	FullName      string
	ClassName     string
	MethodName    string
	ExternalID    string
	AutomationID  string
	SuiteID       *uuid.UUID
	CreatedByID   uuid.UUID
	LabelIDs      []uuid.UUID
	ComponentIDs  []uuid.UUID
}

// UpdateTestCaseInput содержит данные для обновления тест-кейса.
type UpdateTestCaseInput struct {
	Title         *string
	Description   *string
	Precondition  *string
	Postcondition *string
	CaseType      *string
	Priority      *string
	Status        *string
	Steps         *[]schema.TestStep
	FullName      *string
	ClassName     *string
	MethodName    *string
	ExternalID    *string
	AutomationID  *string
	SuiteID       *uuid.UUID
	ClearSuite    bool
	LabelIDs      *[]uuid.UUID
	ComponentIDs  *[]uuid.UUID
	Version       int
}

// ListTestCasesParams содержит параметры фильтрации и пагинации для списка тест-кейсов.
type ListTestCasesParams struct {
	Offset   int
	Limit    int
	SuiteID  *uuid.UUID
	Status   string
	Priority string
	Search   string
}

// CreateSuiteInput содержит данные для создания папки.
type CreateSuiteInput struct {
	Name        string
	Description string
	SortOrder   int
	ParentID    *uuid.UUID
}

// UpdateSuiteInput содержит данные для обновления папки.
type UpdateSuiteInput struct {
	Name        *string
	Description *string
	SortOrder   *int
}

// CreateTestRunInput содержит данные для создания прогона.
type CreateTestRunInput struct {
	Title       string
	Description string
	RunType     string
	CreatedByID uuid.UUID
	AllureRunID string
}

// ListTestRunsParams содержит параметры пагинации для списка прогонов.
type ListTestRunsParams struct {
	Offset int
	Limit  int
	Status string
}

// TestRunStats содержит счётчики результатов прогона.
type TestRunStats struct {
	TotalCount   int
	PassedCount  int
	FailedCount  int
	SkippedCount int
	BrokenCount  int
	UnknownCount int
}

// CreateTestResultInput содержит данные для создания результата теста.
type CreateTestResultInput struct {
	Status        string
	Comment       string
	StepResults   []schema.StepResult
	DurationMs    *int64
	AllureUUID    string
	ErrorMessage  string
	ErrorTrace    string
	IsFlaky       bool
	MatchStrategy string
	TestCaseID    *uuid.UUID
	ExecutedByID  *uuid.UUID
	StartedAt     *time.Time
	CompletedAt   *time.Time
}

// UpdateTestResultInput содержит данные для обновления результата теста.
type UpdateTestResultInput struct {
	Status      *string
	Comment     *string
	StepResults *[]schema.StepResult
	IsFlaky     *bool
}

// CreateEnvironmentInput содержит данные для создания тестового окружения.
type CreateEnvironmentInput struct {
	Name   string
	Label  string
	Config map[string]interface{}
}

// UpdateEnvironmentInput содержит данные для обновления тестового окружения.
type UpdateEnvironmentInput struct {
	Name   *string
	Label  *string
	Config *map[string]interface{}
}
