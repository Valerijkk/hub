package tms

// ---------------------------------------------------------------------------
// NATS-субъекты событий TMS.
// ---------------------------------------------------------------------------

const (
	// EventTestRunCreated публикуется при создании нового прогона.
	EventTestRunCreated = "tms.testrun.created"
	// EventTestRunCompleted публикуется при завершении прогона (A2W2-2-007 канон).
	EventTestRunCompleted = "tms.testrun.completed"
	// EventCoverageUpdated публикуется при изменении покрытия требований.
	EventCoverageUpdated = "tms.coverage.updated"
	// EventTestCaseLinked публикуется при привязке тест-кейса к задаче.
	EventTestCaseLinked = "tms.testcase.linked"
)

// TestRunCreatedEvent — событие создания прогона.
type TestRunCreatedEvent struct {
	RunID     string `json:"run_id"`
	ProjectID string `json:"project_id"`
	Title     string `json:"title"`
	RunType   string `json:"run_type"`
	CreatedBy string `json:"created_by"`
}

// TestRunCompletedEvent — событие завершения прогона с итоговой статистикой.
type TestRunCompletedEvent struct {
	RunID       string `json:"run_id"`
	ProjectID   string `json:"project_id"`
	Title       string `json:"title"`
	RunType     string `json:"run_type"`
	Total       int    `json:"total"`
	Passed      int    `json:"passed"`
	Failed      int    `json:"failed"`
	Skipped     int    `json:"skipped"`
	Broken      int    `json:"broken"`
	CompletedBy string `json:"completed_by"`
}

// CoverageUpdatedEvent — событие обновления покрытия.
type CoverageUpdatedEvent struct {
	ProjectID   string `json:"project_id"`
	TestCaseID  string `json:"test_case_id"`
	IssueID     string `json:"issue_id"`
	Action      string `json:"action"` // "linked" или "unlinked"
}

// TestCaseLinkedEvent — событие привязки тест-кейса к задаче.
type TestCaseLinkedEvent struct {
	TestCaseID  string `json:"test_case_id"`
	TestCaseKey string `json:"test_case_key"`
	IssueID     string `json:"issue_id"`
	IssueKey    string `json:"issue_key"`
	ProjectID   string `json:"project_id"`
}
