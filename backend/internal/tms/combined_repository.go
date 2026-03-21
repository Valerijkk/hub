package tms

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/schema"
)

// CombinedRepository объединяет EntRepository (PostgreSQL) и MongoRepository (MongoDB)
// в единую реализацию интерфейса Repository.
type CombinedRepository struct {
	ent   *EntRepository
	mongo *MongoRepository
}

// NewCombinedRepository создаёт комбинированный репозиторий из Ent и MongoDB частей.
func NewCombinedRepository(entRepo *EntRepository, mongoRepo *MongoRepository) *CombinedRepository {
	return &CombinedRepository{ent: entRepo, mongo: mongoRepo}
}

// EntClient возвращает Ent клиент для прямых операций с БД.
func (r *CombinedRepository) EntClient() *ent.Client {
	return r.ent.EntClient()
}

// --- Тест-кейсы (делегируем в EntRepository) ---

func (r *CombinedRepository) CreateTestCase(ctx context.Context, projectID uuid.UUID, input CreateTestCaseInput) (*ent.TestCase, error) {
	return r.ent.CreateTestCase(ctx, projectID, input)
}
func (r *CombinedRepository) GetTestCase(ctx context.Context, id uuid.UUID) (*ent.TestCase, error) {
	return r.ent.GetTestCase(ctx, id)
}
func (r *CombinedRepository) FindTestCaseByKey(ctx context.Context, caseKey string) (*ent.TestCase, error) {
	return r.ent.FindTestCaseByKey(ctx, caseKey)
}
func (r *CombinedRepository) FindTestCaseByExternalID(ctx context.Context, projectID uuid.UUID, externalID string) (*ent.TestCase, error) {
	return r.ent.FindTestCaseByExternalID(ctx, projectID, externalID)
}
func (r *CombinedRepository) FindTestCaseByFullName(ctx context.Context, projectID uuid.UUID, fullName string) (*ent.TestCase, error) {
	return r.ent.FindTestCaseByFullName(ctx, projectID, fullName)
}
func (r *CombinedRepository) FindTestCaseByClassMethod(ctx context.Context, projectID uuid.UUID, className, methodName string) (*ent.TestCase, error) {
	return r.ent.FindTestCaseByClassMethod(ctx, projectID, className, methodName)
}
func (r *CombinedRepository) ListTestCases(ctx context.Context, projectID uuid.UUID, params ListTestCasesParams) ([]*ent.TestCase, int, error) {
	return r.ent.ListTestCases(ctx, projectID, params)
}
func (r *CombinedRepository) UpdateTestCase(ctx context.Context, id uuid.UUID, input UpdateTestCaseInput) (*ent.TestCase, error) {
	return r.ent.UpdateTestCase(ctx, id, input)
}
func (r *CombinedRepository) DeleteTestCase(ctx context.Context, id uuid.UUID) error {
	return r.ent.DeleteTestCase(ctx, id)
}

// --- Версионирование (делегируем в EntRepository) ---

func (r *CombinedRepository) CreateTestCaseVersion(ctx context.Context, testCaseID, userID uuid.UUID, versionNumber int, snapshot map[string]interface{}, description string) (*ent.TestCaseVersion, error) {
	return r.ent.CreateTestCaseVersion(ctx, testCaseID, userID, versionNumber, snapshot, description)
}
func (r *CombinedRepository) ListTestCaseVersions(ctx context.Context, testCaseID uuid.UUID) ([]*ent.TestCaseVersion, error) {
	return r.ent.ListTestCaseVersions(ctx, testCaseID)
}
func (r *CombinedRepository) GetTestCaseVersion(ctx context.Context, versionID uuid.UUID) (*ent.TestCaseVersion, error) {
	return r.ent.GetTestCaseVersion(ctx, versionID)
}

// --- Папки/сьюты (делегируем в EntRepository) ---

func (r *CombinedRepository) CreateSuite(ctx context.Context, projectID uuid.UUID, input CreateSuiteInput) (*ent.TestSuite, error) {
	return r.ent.CreateSuite(ctx, projectID, input)
}
func (r *CombinedRepository) GetSuite(ctx context.Context, id uuid.UUID) (*ent.TestSuite, error) {
	return r.ent.GetSuite(ctx, id)
}
func (r *CombinedRepository) ListSuites(ctx context.Context, projectID uuid.UUID) ([]*ent.TestSuite, error) {
	return r.ent.ListSuites(ctx, projectID)
}
func (r *CombinedRepository) UpdateSuite(ctx context.Context, id uuid.UUID, input UpdateSuiteInput) (*ent.TestSuite, error) {
	return r.ent.UpdateSuite(ctx, id, input)
}
func (r *CombinedRepository) DeleteSuite(ctx context.Context, id uuid.UUID) error {
	return r.ent.DeleteSuite(ctx, id)
}
func (r *CombinedRepository) MoveSuite(ctx context.Context, id uuid.UUID, parentID *uuid.UUID) (*ent.TestSuite, error) {
	return r.ent.MoveSuite(ctx, id, parentID)
}

// --- Компоненты (делегируем в EntRepository) ---

func (r *CombinedRepository) CreateComponent(ctx context.Context, projectID uuid.UUID, name, description string) (*ent.Component, error) {
	return r.ent.CreateComponent(ctx, projectID, name, description)
}
func (r *CombinedRepository) ListComponents(ctx context.Context, projectID uuid.UUID) ([]*ent.Component, error) {
	return r.ent.ListComponents(ctx, projectID)
}
func (r *CombinedRepository) DeleteComponent(ctx context.Context, id uuid.UUID) error {
	return r.ent.DeleteComponent(ctx, id)
}

// --- Прогоны (делегируем в EntRepository) ---

func (r *CombinedRepository) CreateTestRun(ctx context.Context, projectID uuid.UUID, input CreateTestRunInput) (*ent.TestRun, error) {
	return r.ent.CreateTestRun(ctx, projectID, input)
}
func (r *CombinedRepository) GetTestRun(ctx context.Context, id uuid.UUID) (*ent.TestRun, error) {
	return r.ent.GetTestRun(ctx, id)
}
func (r *CombinedRepository) ListTestRuns(ctx context.Context, projectID uuid.UUID, params ListTestRunsParams) ([]*ent.TestRun, int, error) {
	return r.ent.ListTestRuns(ctx, projectID, params)
}
func (r *CombinedRepository) UpdateTestRunStatus(ctx context.Context, id uuid.UUID, status string, timestamp *time.Time) (*ent.TestRun, error) {
	return r.ent.UpdateTestRunStatus(ctx, id, status, timestamp)
}
func (r *CombinedRepository) ReopenTestRun(ctx context.Context, id uuid.UUID) (*ent.TestRun, error) {
	return r.ent.ReopenTestRun(ctx, id)
}
func (r *CombinedRepository) UpdateTestRunStats(ctx context.Context, id uuid.UUID, stats TestRunStats) (*ent.TestRun, error) {
	return r.ent.UpdateTestRunStats(ctx, id, stats)
}

// --- Результаты (делегируем в EntRepository) ---

func (r *CombinedRepository) CreateTestResult(ctx context.Context, runID uuid.UUID, input CreateTestResultInput) (*ent.TestResult, error) {
	return r.ent.CreateTestResult(ctx, runID, input)
}
func (r *CombinedRepository) UpdateTestResult(ctx context.Context, id uuid.UUID, input UpdateTestResultInput) (*ent.TestResult, error) {
	return r.ent.UpdateTestResult(ctx, id, input)
}
func (r *CombinedRepository) ListTestResults(ctx context.Context, runID uuid.UUID) ([]*ent.TestResult, error) {
	return r.ent.ListTestResults(ctx, runID)
}
func (r *CombinedRepository) BatchCreateTestResults(ctx context.Context, runID uuid.UUID, inputs []CreateTestResultInput) ([]*ent.TestResult, error) {
	return r.ent.BatchCreateTestResults(ctx, runID, inputs)
}

// --- Счётчики (делегируем в EntRepository) ---

func (r *CombinedRepository) IncrementTestCaseCounter(ctx context.Context, projectID uuid.UUID) (int, error) {
	return r.ent.IncrementTestCaseCounter(ctx, projectID)
}

// --- Привязка к задачам (делегируем в EntRepository) ---

func (r *CombinedRepository) LinkTestCaseToIssue(ctx context.Context, testCaseID, issueID uuid.UUID) error {
	return r.ent.LinkTestCaseToIssue(ctx, testCaseID, issueID)
}
func (r *CombinedRepository) UnlinkTestCaseFromIssue(ctx context.Context, testCaseID, issueID uuid.UUID) error {
	return r.ent.UnlinkTestCaseFromIssue(ctx, testCaseID, issueID)
}

// --- Расширенные операции прогонов (делегируем в EntRepository) ---

func (r *CombinedRepository) AddTestCasesToRun(ctx context.Context, runID uuid.UUID, caseIDs []uuid.UUID) error {
	return r.ent.AddTestCasesToRun(ctx, runID, caseIDs)
}
func (r *CombinedRepository) GetTestResult(ctx context.Context, resultID uuid.UUID) (*ent.TestResult, error) {
	return r.ent.GetTestResult(ctx, resultID)
}
func (r *CombinedRepository) ListTestRunResults(ctx context.Context, runID uuid.UUID) ([]*ent.TestResult, error) {
	return r.ent.ListTestRunResults(ctx, runID)
}
func (r *CombinedRepository) AssignTestersToRun(ctx context.Context, runID uuid.UUID, userIDs []uuid.UUID) error {
	return r.ent.AssignTestersToRun(ctx, runID, userIDs)
}
func (r *CombinedRepository) GetTestRunWithResults(ctx context.Context, runID uuid.UUID) (*ent.TestRun, error) {
	return r.ent.GetTestRunWithResults(ctx, runID)
}
func (r *CombinedRepository) AddFilesToResult(ctx context.Context, resultID uuid.UUID, fileIDs []uuid.UUID) error {
	return r.ent.AddFilesToResult(ctx, resultID, fileIDs)
}

// --- Кросс-модульные связи (делегируем в EntRepository) ---

// GetTestCaseLinkedIssues возвращает задачи, привязанные к тест-кейсу.
func (r *CombinedRepository) GetTestCaseLinkedIssues(ctx context.Context, testCaseID uuid.UUID) ([]*ent.Issue, error) {
	return r.ent.GetTestCaseLinkedIssues(ctx, testCaseID)
}

// GetIssueLinkedTestCases возвращает тест-кейсы, привязанные к задаче.
func (r *CombinedRepository) GetIssueLinkedTestCases(ctx context.Context, issueID uuid.UUID) ([]*ent.TestCase, error) {
	return r.ent.GetIssueLinkedTestCases(ctx, issueID)
}

// GetProjectIssuesWithTestCases возвращает задачи проекта с привязанными тест-кейсами.
func (r *CombinedRepository) GetProjectIssuesWithTestCases(ctx context.Context, projectID uuid.UUID) ([]*ent.Issue, error) {
	return r.ent.GetProjectIssuesWithTestCases(ctx, projectID)
}

// CreateBugIssue создаёт задачу типа bug и привязывает к тест-кейсу.
func (r *CombinedRepository) CreateBugIssue(ctx context.Context, projectID, userID, testCaseID uuid.UUID, title, description, priority string) (*ent.Issue, error) {
	return r.ent.CreateBugIssue(ctx, projectID, userID, testCaseID, title, description, priority)
}

// GetLastResultForTestCase возвращает последний результат для тест-кейса.
func (r *CombinedRepository) GetLastResultForTestCase(ctx context.Context, testCaseID uuid.UUID) (*ent.TestResult, error) {
	return r.ent.GetLastResultForTestCase(ctx, testCaseID)
}

// --- MongoDB (делегируем в MongoRepository) ---

func (r *CombinedRepository) StoreAllureResults(ctx context.Context, runID uuid.UUID, results []AllureResult) error {
	return r.mongo.StoreAllureResults(ctx, runID, results)
}
func (r *CombinedRepository) GetAllureResultsByRunID(ctx context.Context, runID uuid.UUID) ([]AllureResult, error) {
	return r.mongo.GetAllureResultsByRunID(ctx, runID)
}

// Suppress unused import warning for schema.
var _ = schema.TestStep{}
