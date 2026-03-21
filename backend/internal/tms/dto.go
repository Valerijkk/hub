package tms

import (
	"encoding/json"
	"time"

	"github.com/valerii/hub/backend/ent/schema"
)

// ---------------------------------------------------------------------------
// DTO запросов
// ---------------------------------------------------------------------------

// CreateTestCaseRequest — тело запроса на создание тест-кейса.
type CreateTestCaseRequest struct {
	Title         string            `json:"title" validate:"required,min=1,max=500"`
	Description   string            `json:"description,omitempty"`
	Precondition  string            `json:"precondition,omitempty"`
	Postcondition string            `json:"postcondition,omitempty"`
	CaseType      string            `json:"case_type,omitempty" validate:"omitempty,oneof=classic bdd"`
	Priority      string            `json:"priority,omitempty" validate:"omitempty,oneof=critical high medium low trivial"`
	Status        string            `json:"status,omitempty" validate:"omitempty,oneof=draft active deprecated"`
	Steps         []schema.TestStep `json:"steps,omitempty"`
	FullName      string            `json:"full_name,omitempty"`
	ClassName     string            `json:"class_name,omitempty"`
	MethodName    string            `json:"method_name,omitempty"`
	ExternalID    string            `json:"external_id,omitempty"`
	SuiteID       string            `json:"suite_id,omitempty"`
	Labels        []string          `json:"labels,omitempty"`
	Components    []string          `json:"components,omitempty"`
}

// UpdateTestCaseRequest — тело запроса на обновление тест-кейса.
type UpdateTestCaseRequest struct {
	Title         *string            `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	Description   *string            `json:"description,omitempty"`
	Precondition  *string            `json:"precondition,omitempty"`
	Postcondition *string            `json:"postcondition,omitempty"`
	CaseType      *string            `json:"case_type,omitempty" validate:"omitempty,oneof=classic bdd"`
	Priority      *string            `json:"priority,omitempty" validate:"omitempty,oneof=critical high medium low trivial"`
	Status        *string            `json:"status,omitempty" validate:"omitempty,oneof=draft active deprecated"`
	Steps         *[]schema.TestStep `json:"steps,omitempty"`
	FullName      *string            `json:"full_name,omitempty"`
	ClassName     *string            `json:"class_name,omitempty"`
	MethodName    *string            `json:"method_name,omitempty"`
	ExternalID    *string            `json:"external_id,omitempty"`
	SuiteID       *string            `json:"suite_id,omitempty"`
	Labels        *[]string          `json:"labels,omitempty"`
	Components    *[]string          `json:"components,omitempty"`
	Version       int                `json:"version" validate:"required"`
}

// RollbackTestCaseRequest — тело запроса на откат тест-кейса к версии.
type RollbackTestCaseRequest struct {
	VersionID string `json:"version_id" validate:"required"`
}

// LinkIssueRequest — тело запроса на привязку тест-кейса к задаче.
type LinkIssueRequest struct {
	IssueID string `json:"issue_id" validate:"required"`
}

// CreateSuiteRequest — тело запроса на создание папки/секции.
type CreateSuiteRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=200"`
	Description string `json:"description,omitempty"`
	SortOrder   int    `json:"sort_order,omitempty"`
	ParentID    string `json:"parent_id,omitempty"`
}

// UpdateSuiteRequest — тело запроса на обновление папки.
type UpdateSuiteRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string `json:"description,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
}

// CreateComponentRequest — тело запроса на создание компонента.
type CreateComponentRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=200"`
	Description string `json:"description,omitempty"`
}

// CreateTestRunRequest — тело запроса на создание прогона.
type CreateTestRunRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=500"`
	Description string `json:"description,omitempty"`
	RunType     string `json:"run_type,omitempty" validate:"omitempty,oneof=manual automatic mixed"`
}

// UpdateRunStatusRequest — тело запроса на обновление статуса прогона.
type UpdateRunStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=in_progress completed aborted"`
}

// CreateTestResultRequest — тело запроса на создание результата теста.
type CreateTestResultRequest struct {
	Status        string              `json:"status" validate:"required,oneof=passed failed broken skipped unknown"`
	Comment       string              `json:"comment,omitempty"`
	StepResults   []schema.StepResult `json:"step_results,omitempty"`
	DurationMs    *int64              `json:"duration_ms,omitempty"`
	TestCaseID    string              `json:"test_case_id,omitempty"`
	MatchStrategy string              `json:"match_strategy,omitempty" validate:"omitempty,oneof=annotation fullname classmethod unmapped"`
	IsFlaky       bool                `json:"is_flaky,omitempty"`
}

// UpdateTestResultRequest — тело запроса на обновление результата теста.
type UpdateTestResultRequest struct {
	Status      *string              `json:"status,omitempty" validate:"omitempty,oneof=passed failed broken skipped unknown"`
	Comment     *string              `json:"comment,omitempty"`
	StepResults *[]schema.StepResult `json:"step_results,omitempty"`
	IsFlaky     *bool                `json:"is_flaky,omitempty"`
	FileIDs     []string             `json:"file_ids,omitempty"`
}

// AddTestCasesToRunRequest — тело запроса на добавление тест-кейсов в прогон.
type AddTestCasesToRunRequest struct {
	CaseIDs []string `json:"case_ids" validate:"required,min=1"`
}

// AssignTestersRequest — тело запроса на назначение тестировщиков прогону.
type AssignTestersRequest struct {
	UserIDs []string `json:"user_ids" validate:"required,min=1"`
}

// ImportAllureRequest — тело запроса на импорт результатов Allure.
type ImportAllureRequest struct {
	Results     json.RawMessage `json:"results" validate:"required"`
	Title       string          `json:"title,omitempty"`
	AllureRunID string          `json:"allure_run_id,omitempty"`
}

// TestRunSummary — краткая информация о прогоне для истории.
type TestRunSummary struct {
	RunID     string    `json:"run_id"`
	Title     string    `json:"title"`
	Total     int       `json:"total"`
	Passed    int       `json:"passed"`
	Failed    int       `json:"failed"`
	PassRate  float64   `json:"pass_rate"`
	CreatedAt time.Time `json:"created_at"`
}

// TestRunDetailResponse — полная информация о прогоне с результатами и назначенными.
type TestRunDetailResponse struct {
	TestRunResponse
	Results   []TestResultDetailResponse `json:"results"`
	Assignees []UserBriefResponse        `json:"assignees"`
}

// TestResultDetailResponse — результат с полной информацией о тест-кейсе и файлах.
type TestResultDetailResponse struct {
	ID            string              `json:"id"`
	Status        string              `json:"status"`
	Comment       string              `json:"comment,omitempty"`
	StepResults   []schema.StepResult `json:"step_results,omitempty"`
	DurationMs    *int64              `json:"duration_ms,omitempty"`
	IsFlaky       bool                `json:"is_flaky"`
	TestCase      *TestCaseBriefDTO   `json:"test_case,omitempty"`
	ExecutedBy    *UserBriefResponse  `json:"executed_by,omitempty"`
	Files         []FileBriefResponse `json:"files,omitempty"`
	StartedAt     *time.Time          `json:"started_at,omitempty"`
	CompletedAt   *time.Time          `json:"completed_at,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
}

// TestCaseBriefDTO — краткое представление тест-кейса для результата.
type TestCaseBriefDTO struct {
	ID       string `json:"id"`
	CaseKey  string `json:"case_key"`
	Title    string `json:"title"`
	CaseType string `json:"case_type"`
	Priority string `json:"priority"`
}

// FileBriefResponse — краткое представление файла-вложения.
type FileBriefResponse struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
}

// ListTestCasesRequest — параметры запроса списка тест-кейсов.
type ListTestCasesRequest struct {
	Offset   int    `query:"offset" validate:"omitempty,min=0"`
	Limit    int    `query:"limit" validate:"omitempty,min=1,max=100"`
	SuiteID  string `query:"suite_id"`
	Status   string `query:"status" validate:"omitempty,oneof=draft active deprecated"`
	Priority string `query:"priority" validate:"omitempty,oneof=critical high medium low trivial"`
	Search   string `query:"search"`
}

// ListTestRunsRequest — параметры запроса списка прогонов.
type ListTestRunsRequest struct {
	Offset int    `query:"offset" validate:"omitempty,min=0"`
	Limit  int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Status string `query:"status" validate:"omitempty,oneof=created in_progress completed aborted"`
}

// ---------------------------------------------------------------------------
// DTO ответов
// ---------------------------------------------------------------------------

// TestCaseResponse представляет тест-кейс в ответе API.
type TestCaseResponse struct {
	ID            string              `json:"id"`
	CaseNumber    int                 `json:"case_number"`
	CaseKey       string              `json:"case_key"`
	Title         string              `json:"title"`
	Description   string              `json:"description,omitempty"`
	Precondition  string              `json:"precondition,omitempty"`
	Postcondition string              `json:"postcondition,omitempty"`
	CaseType      string              `json:"case_type"`
	Priority      string              `json:"priority"`
	Status        string              `json:"status"`
	Steps         []schema.TestStep   `json:"steps,omitempty"`
	FullName      string              `json:"full_name,omitempty"`
	ClassName     string              `json:"class_name,omitempty"`
	MethodName    string              `json:"method_name,omitempty"`
	ExternalID    string              `json:"external_id,omitempty"`
	Version       int                 `json:"version"`
	Suite         *SuiteBriefResponse `json:"suite,omitempty"`
	CreatedBy     *UserBriefResponse  `json:"created_by,omitempty"`
	Labels        []LabelResponse     `json:"labels,omitempty"`
	Components    []ComponentResponse `json:"components,omitempty"`
	LinkedIssues  []IssueBriefResponse `json:"linked_issues,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// TestCaseVersionResponse представляет версию тест-кейса.
type TestCaseVersionResponse struct {
	ID                string                 `json:"id"`
	VersionNumber     int                    `json:"version_number"`
	Snapshot          map[string]interface{} `json:"snapshot"`
	ChangeDescription string                `json:"change_description,omitempty"`
	CreatedBy         *UserBriefResponse     `json:"created_by,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

// SuiteBriefResponse — краткое представление папки.
type SuiteBriefResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SuiteResponse представляет папку/секцию в ответе API.
type SuiteResponse struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	SortOrder   int             `json:"sort_order"`
	ParentID    string          `json:"parent_id,omitempty"`
	Children    []SuiteResponse `json:"children,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ComponentResponse представляет компонент в ответе API.
type ComponentResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// TestRunResponse представляет прогон в ответе API.
type TestRunResponse struct {
	ID           string             `json:"id"`
	Title        string             `json:"title"`
	Description  string             `json:"description,omitempty"`
	RunType      string             `json:"run_type"`
	Status       string             `json:"status"`
	TotalCount   int                `json:"total_count"`
	PassedCount  int                `json:"passed_count"`
	FailedCount  int                `json:"failed_count"`
	SkippedCount int                `json:"skipped_count"`
	BrokenCount  int                `json:"broken_count"`
	UnknownCount int                `json:"unknown_count"`
	AllureRunID  string             `json:"allure_run_id,omitempty"`
	CreatedBy    *UserBriefResponse `json:"created_by,omitempty"`
	StartedAt    *time.Time         `json:"started_at,omitempty"`
	CompletedAt  *time.Time         `json:"completed_at,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// TestResultResponse представляет результат теста в ответе API.
type TestResultResponse struct {
	ID            string              `json:"id"`
	Status        string              `json:"status"`
	Comment       string              `json:"comment,omitempty"`
	StepResults   []schema.StepResult `json:"step_results,omitempty"`
	DurationMs    *int64              `json:"duration_ms,omitempty"`
	AllureUUID    string              `json:"allure_uuid,omitempty"`
	ErrorMessage  string              `json:"error_message,omitempty"`
	ErrorTrace    string              `json:"error_trace,omitempty"`
	IsFlaky       bool                `json:"is_flaky"`
	MatchStrategy string              `json:"match_strategy"`
	TestCaseKey   string              `json:"test_case_key,omitempty"`
	ExecutedBy    *UserBriefResponse  `json:"executed_by,omitempty"`
	StartedAt     *time.Time          `json:"started_at,omitempty"`
	CompletedAt   *time.Time          `json:"completed_at,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
}

// UserBriefResponse — краткое представление пользователя.
type UserBriefResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// LabelResponse — представление метки.
type LabelResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// IssueBriefResponse — краткое представление задачи.
type IssueBriefResponse struct {
	ID       string `json:"id"`
	IssueKey string `json:"issue_key"`
	Title    string `json:"title"`
}

// CreateBugRequest — тело запроса на создание бага из упавшего теста.
type CreateBugRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Priority    string `json:"priority,omitempty" validate:"omitempty,oneof=critical high medium low trivial"`
}

// CoverageReport — отчёт о покрытии задач тестами.
type CoverageReport struct {
	Issues []IssueCoverage `json:"issues"`
}

// IssueCoverage — покрытие одной задачи тест-кейсами.
type IssueCoverage struct {
	IssueID    string             `json:"issue_id"`
	IssueKey   string             `json:"issue_key"`
	IssueTitle string             `json:"issue_title"`
	TestCases  []TestCaseCoverage `json:"test_cases"`
}

// TestCaseCoverage — покрытие тест-кейса с последним статусом результата.
type TestCaseCoverage struct {
	CaseKey    string `json:"case_key"`
	CaseTitle  string `json:"case_title"`
	LastStatus string `json:"last_status"`
}

// TrendPoint — точка данных для графика трендов прогонов.
type TrendPoint struct {
	RunID     string    `json:"run_id"`
	Title     string    `json:"title"`
	Total     int       `json:"total"`
	Passed    int       `json:"passed"`
	Failed    int       `json:"failed"`
	Skipped   int       `json:"skipped"`
	Broken    int       `json:"broken"`
	PassRate  float64   `json:"pass_rate"`
	CreatedAt time.Time `json:"created_at"`
}

// RunComparison — сравнение двух прогонов.
type RunComparison struct {
	Run1    TestRunSummary `json:"run1"`
	Run2    TestRunSummary `json:"run2"`
	Changes []ResultChange `json:"changes"`
}

// ResultChange — изменение статуса тест-кейса между прогонами.
type ResultChange struct {
	CaseKey    string `json:"case_key"`
	CaseTitle  string `json:"case_title"`
	StatusRun1 string `json:"status_run1"`
	StatusRun2 string `json:"status_run2"`
}

// ImportRequest — тело запроса на импорт тест-кейсов.
type ImportRequest struct {
	Format  string            `json:"format" validate:"required,oneof=csv json xml excel"`
	Mapping map[string]string `json:"mapping,omitempty"`
}

// ImportResult — результат импорта тест-кейсов.
type ImportResult struct {
	Created          int                     `json:"created"`
	Skipped          int                     `json:"skipped"`
	Errors           int                     `json:"errors"`
	ValidationErrors []ImportValidationError `json:"validation_errors,omitempty"`
}

// ImportValidationError — ошибка валидации строки при импорте.
type ImportValidationError struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

// PaginatedResponse — обёртка для пагинированного ответа.
type PaginatedResponse[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}
