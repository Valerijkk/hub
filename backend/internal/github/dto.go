package github

import "time"

// ---------------------------------------------------------------------------
// DTO запросов
// ---------------------------------------------------------------------------

// CreateIntegrationRequest -- запрос на создание интеграции с GitHub-репозиторием.
type CreateIntegrationRequest struct {
	RepoOwner     string `json:"repo_owner" validate:"required"`
	RepoName      string `json:"repo_name" validate:"required"`
	AccessToken   string `json:"access_token" validate:"required"`
	WebhookSecret string `json:"webhook_secret" validate:"required"`
	MergeStatusID string `json:"merge_status_id,omitempty"`
}

// CreateBranchRequest -- запрос на создание ветки из задачи.
type CreateBranchRequest struct {
	IntegrationID string `json:"integration_id" validate:"required"`
	BranchName    string `json:"branch_name,omitempty"`
}

// ---------------------------------------------------------------------------
// DTO ответов
// ---------------------------------------------------------------------------

// IntegrationResponse -- ответ с данными интеграции.
type IntegrationResponse struct {
	ID            string    `json:"id"`
	RepoOwner     string    `json:"repo_owner"`
	RepoName      string    `json:"repo_name"`
	MergeStatusID string    `json:"merge_status_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// GitHubLinkResponse -- ответ с данными связи PR/коммита с задачей.
type GitHubLinkResponse struct {
	ID         string    `json:"id"`
	LinkType   string    `json:"link_type"`
	ExternalID string    `json:"external_id"`
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	Status     string    `json:"status"`
	Author     string    `json:"author,omitempty"`
	Branch     string    `json:"branch,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateBranchResponse -- ответ при создании ветки.
type CreateBranchResponse struct {
	BranchName string `json:"branch_name"`
}

// ---------------------------------------------------------------------------
// DTO CI/CD-пайплайна
// ---------------------------------------------------------------------------

// WorkflowRunResponse -- ответ с данными запуска workflow (CI/CD пайплайна).
type WorkflowRunResponse struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`     // queued, in_progress, completed
	Conclusion string    `json:"conclusion"` // success, failure, cancelled, skipped, timed_out, action_required, ""
	Branch     string    `json:"branch"`
	Event      string    `json:"event"`      // push, pull_request, schedule, etc.
	HTMLURL    string    `json:"html_url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	HeadSHA    string    `json:"head_sha"`
	RunNumber  int       `json:"run_number"`
}

// WorkflowRunDetailResponse -- детальный ответ запуска workflow с джобами.
type WorkflowRunDetailResponse struct {
	WorkflowRunResponse
	Jobs []WorkflowJobResponse `json:"jobs"`
}

// WorkflowJobResponse -- джоб внутри workflow run.
type WorkflowJobResponse struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Conclusion string    `json:"conclusion"`
	StartedAt  time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// LinkPreviewResponse -- ответ с данными превью ссылки.
type LinkPreviewResponse struct {
	URL         string `json:"url"`
	Type        string `json:"type"` // issue, test_case, github, external
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	Domain      string `json:"domain,omitempty"`
	// Поля для внутренних сущностей
	IssueKey   string `json:"issue_key,omitempty"`
	Status     string `json:"status,omitempty"`
	IssueType  string `json:"issue_type,omitempty"`
	AssigneeName string `json:"assignee_name,omitempty"`
	AvatarURL  string `json:"avatar_url,omitempty"`
}
