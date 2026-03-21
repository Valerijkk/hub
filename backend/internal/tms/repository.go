package tms

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/schema"
)

// Repository определяет интерфейс доступа к данным модуля TMS.
// Позволяет тестировать сервисный слой с помощью mock-реализаций.
type Repository interface {
	// --- Тест-кейсы ---

	// CreateTestCase создаёт новый тест-кейс в проекте.
	CreateTestCase(ctx context.Context, projectID uuid.UUID, input CreateTestCaseInput) (*ent.TestCase, error)
	// GetTestCase возвращает тест-кейс по ID с подгруженными связями.
	GetTestCase(ctx context.Context, id uuid.UUID) (*ent.TestCase, error)
	// FindTestCaseByKey возвращает тест-кейс по ключу (например "HUB-TC-42").
	FindTestCaseByKey(ctx context.Context, caseKey string) (*ent.TestCase, error)
	// FindTestCaseByExternalID возвращает тест-кейс по внешнему ID аннотации.
	FindTestCaseByExternalID(ctx context.Context, projectID uuid.UUID, externalID string) (*ent.TestCase, error)
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

	// --- Версионирование ---

	// CreateTestCaseVersion создаёт снимок версии тест-кейса.
	CreateTestCaseVersion(ctx context.Context, testCaseID, userID uuid.UUID, versionNumber int, snapshot map[string]interface{}, description string) (*ent.TestCaseVersion, error)
	// ListTestCaseVersions возвращает все версии тест-кейса.
	ListTestCaseVersions(ctx context.Context, testCaseID uuid.UUID) ([]*ent.TestCaseVersion, error)
	// GetTestCaseVersion возвращает конкретную версию по ID.
	GetTestCaseVersion(ctx context.Context, versionID uuid.UUID) (*ent.TestCaseVersion, error)

	// --- Папки/сьюты ---

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

	// --- Компоненты ---

	// CreateComponent создаёт новый компонент в проекте.
	CreateComponent(ctx context.Context, projectID uuid.UUID, name, description string) (*ent.Component, error)
	// ListComponents возвращает все компоненты проекта.
	ListComponents(ctx context.Context, projectID uuid.UUID) ([]*ent.Component, error)
	// DeleteComponent удаляет компонент по ID.
	DeleteComponent(ctx context.Context, id uuid.UUID) error

	// --- Прогоны ---

	// CreateTestRun создаёт новый прогон тестов.
	CreateTestRun(ctx context.Context, projectID uuid.UUID, input CreateTestRunInput) (*ent.TestRun, error)
	// GetTestRun возвращает прогон по ID с подгруженными связями.
	GetTestRun(ctx context.Context, id uuid.UUID) (*ent.TestRun, error)
	// ListTestRuns возвращает пагинированный список прогонов проекта.
	ListTestRuns(ctx context.Context, projectID uuid.UUID, params ListTestRunsParams) ([]*ent.TestRun, int, error)
	// UpdateTestRunStatus обновляет статус прогона.
	UpdateTestRunStatus(ctx context.Context, id uuid.UUID, status string, timestamp *time.Time) (*ent.TestRun, error)
	// ReopenTestRun переоткрывает завершённый прогон.
	ReopenTestRun(ctx context.Context, id uuid.UUID) (*ent.TestRun, error)
	// UpdateTestRunStats обновляет счётчики результатов прогона.
	UpdateTestRunStats(ctx context.Context, id uuid.UUID, stats TestRunStats) (*ent.TestRun, error)

	// --- Результаты ---

	// CreateTestResult создаёт результат выполнения теста.
	CreateTestResult(ctx context.Context, runID uuid.UUID, input CreateTestResultInput) (*ent.TestResult, error)
	// UpdateTestResult обновляет результат теста.
	UpdateTestResult(ctx context.Context, id uuid.UUID, input UpdateTestResultInput) (*ent.TestResult, error)
	// ListTestResults возвращает все результаты прогона.
	ListTestResults(ctx context.Context, runID uuid.UUID) ([]*ent.TestResult, error)
	// BatchCreateTestResults массово создаёт результаты (для импорта Allure).
	BatchCreateTestResults(ctx context.Context, runID uuid.UUID, inputs []CreateTestResultInput) ([]*ent.TestResult, error)

	// --- Счётчики ---

	// IncrementTestCaseCounter атомарно увеличивает счётчик тест-кейсов проекта.
	IncrementTestCaseCounter(ctx context.Context, projectID uuid.UUID) (int, error)

	// --- MongoDB (Allure raw storage) ---

	// StoreAllureResults сохраняет сырые результаты Allure в MongoDB.
	StoreAllureResults(ctx context.Context, runID uuid.UUID, results []AllureResult) error
	// GetAllureResultsByRunID возвращает сырые результаты Allure для прогона.
	GetAllureResultsByRunID(ctx context.Context, runID uuid.UUID) ([]AllureResult, error)

	// --- Привязка к задачам ---

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
	// GetLastResultForTestCase возвращает последний результат для тест-кейса.
	GetLastResultForTestCase(ctx context.Context, testCaseID uuid.UUID) (*ent.TestResult, error)

	// --- Расширенные операции прогонов ---

	// AddTestCasesToRun массово создаёт TestResult записи для каждого тест-кейса.
	AddTestCasesToRun(ctx context.Context, runID uuid.UUID, caseIDs []uuid.UUID) error
	// GetTestResult возвращает результат по ID с подгруженными связями.
	GetTestResult(ctx context.Context, resultID uuid.UUID) (*ent.TestResult, error)
	// ListTestRunResults возвращает результаты прогона с подгруженными тест-кейсами и файлами.
	ListTestRunResults(ctx context.Context, runID uuid.UUID) ([]*ent.TestResult, error)
	// AssignTestersToRun устанавливает M2M связь прогона с назначенными тестировщиками.
	AssignTestersToRun(ctx context.Context, runID uuid.UUID, userIDs []uuid.UUID) error
	// GetTestRunWithResults возвращает прогон с результатами, тест-кейсами, файлами и назначенными.
	GetTestRunWithResults(ctx context.Context, runID uuid.UUID) (*ent.TestRun, error)
	// AddFilesToResult привязывает файлы к результату теста.
	AddFilesToResult(ctx context.Context, resultID uuid.UUID, fileIDs []uuid.UUID) error
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
