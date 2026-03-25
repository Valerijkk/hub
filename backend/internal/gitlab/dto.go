package gitlab

import "time"

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateIntegrationRequest -- запрос на подключение GitLab-репозитория.
type CreateIntegrationRequest struct {
	GitLabURL   string `json:"gitlab_url,omitempty"`   // URL инстанса GitLab (по умолчанию https://gitlab.com)
	ProjectPath string `json:"project_path" validate:"required"` // Путь проекта (group/repo)
	AccessToken string `json:"access_token" validate:"required"` // Personal Access Token
	WebhookSecret string `json:"webhook_secret,omitempty"`        // Секрет для валидации webhook
}

// CreateBranchRequest -- запрос на создание ветки в GitLab.
type CreateBranchRequest struct {
	IntegrationID string `json:"integration_id" validate:"required"`
	BranchName    string `json:"branch_name,omitempty"`
	Ref           string `json:"ref,omitempty"` // Исходная ветка (по умолчанию default branch)
}

// ---------------------------------------------------------------------------
// Response DTOs
// ---------------------------------------------------------------------------

// IntegrationResponse -- ответ с данными интеграции GitLab.
type IntegrationResponse struct {
	ID          string    `json:"id"`
	GitLabURL   string    `json:"gitlab_url"`
	ProjectPath string    `json:"project_path"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateBranchResponse -- ответ при создании ветки.
type CreateBranchResponse struct {
	BranchName string `json:"branch_name"`
}

// ---------------------------------------------------------------------------
// Pipeline DTOs
// ---------------------------------------------------------------------------

// PipelineResponse -- ответ с данными пайплайна GitLab CI/CD.
type PipelineResponse struct {
	ID        int64     `json:"id"`
	Status    string    `json:"status"`    // created, waiting_for_resource, preparing, pending, running, success, failed, canceled, skipped, manual, scheduled
	Ref       string    `json:"ref"`       // Ветка или тег
	SHA       string    `json:"sha"`       // Коммит SHA
	WebURL    string    `json:"web_url"`   // Ссылка на пайплайн в GitLab
	Source    string    `json:"source"`    // push, web, trigger, schedule, api, external, pipeline, chat, merge_request_event
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ---------------------------------------------------------------------------
// Webhook Payloads (входящие от GitLab)
// ---------------------------------------------------------------------------

// WebhookPushEvent -- payload push-события GitLab webhook.
type WebhookPushEvent struct {
	ObjectKind string           `json:"object_kind"` // "push"
	Ref        string           `json:"ref"`
	Project    WebhookProject   `json:"project"`
	Commits    []WebhookCommit  `json:"commits"`
}

// WebhookMergeRequestEvent -- payload merge_request-события GitLab webhook.
type WebhookMergeRequestEvent struct {
	ObjectKind      string              `json:"object_kind"` // "merge_request"
	Project         WebhookProject      `json:"project"`
	ObjectAttributes WebhookMergeRequest `json:"object_attributes"`
}

// WebhookPipelineEvent -- payload pipeline-события GitLab webhook.
type WebhookPipelineEvent struct {
	ObjectKind       string              `json:"object_kind"` // "pipeline"
	Project          WebhookProject      `json:"project"`
	ObjectAttributes WebhookPipeline     `json:"object_attributes"`
}

// WebhookProject -- данные проекта из GitLab webhook.
type WebhookProject struct {
	ID                int64  `json:"id"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
}

// WebhookCommit -- данные коммита из push webhook.
type WebhookCommit struct {
	ID      string `json:"id"`      // SHA коммита
	Message string `json:"message"`
	URL     string `json:"url"`
	Author  struct {
		Name string `json:"name"`
	} `json:"author"`
}

// WebhookMergeRequest -- данные MR из merge_request webhook.
type WebhookMergeRequest struct {
	IID         int64  `json:"iid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`    // opened, closed, merged
	Action      string `json:"action"`   // open, close, reopen, merge, update
	URL         string `json:"url"`
	SourceBranch string `json:"source_branch"`
	AuthorID    int64  `json:"author_id"`
}

// WebhookPipeline -- данные пайплайна из pipeline webhook.
type WebhookPipeline struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
	Ref    string `json:"ref"`
	SHA    string `json:"sha"`
}
