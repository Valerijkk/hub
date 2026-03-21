package tracker

import "time"

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateIssueRequest — тело запроса на создание новой задачи.
type CreateIssueRequest struct {
	Title           string                  `json:"title" validate:"required,min=1,max=500"`
	IssueType       string                  `json:"issue_type" validate:"required,oneof=bug task story epic sub_task improvement tech_debt spike release"`
	DescriptionJSON *map[string]interface{} `json:"description_json,omitempty"`
	Priority        string                  `json:"priority,omitempty" validate:"omitempty,oneof=critical high medium low trivial"`
	AssigneeID      string                  `json:"assignee_id,omitempty" validate:"omitempty"`
	ParentID        string                  `json:"parent_id,omitempty" validate:"omitempty"`
	StatusID        string                  `json:"status_id,omitempty" validate:"omitempty"`
	Labels          []string                `json:"labels,omitempty"`
	CustomFields    map[string]interface{}  `json:"custom_fields,omitempty"`
	EstimateMinutes *int                    `json:"estimate_minutes,omitempty"`
	DueDate         *time.Time              `json:"due_date,omitempty"`
}

// UpdateIssueRequest — тело запроса на обновление существующей задачи.
type UpdateIssueRequest struct {
	Title           *string                 `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	DescriptionJSON *map[string]interface{} `json:"description_json,omitempty"`
	Priority        *string                 `json:"priority,omitempty" validate:"omitempty,oneof=critical high medium low trivial"`
	AssigneeID      *string                 `json:"assignee_id,omitempty"`
	StatusID        *string                 `json:"status_id,omitempty"`
	Labels          *[]string               `json:"labels,omitempty"`
	CustomFields    *map[string]interface{} `json:"custom_fields,omitempty"`
	EstimateMinutes *int                    `json:"estimate_minutes,omitempty"`
	DueDate         *time.Time              `json:"due_date,omitempty"`
	Version         int                     `json:"version" validate:"required"`
}

// MoveIssueRequest — тело запроса на перемещение задачи на доске.
type MoveIssueRequest struct {
	StatusID string `json:"status_id" validate:"required"`
	Rank     string `json:"rank" validate:"required"`
}

// ListIssuesRequest — параметры фильтрации и пагинации для списка задач.
type ListIssuesRequest struct {
	Cursor    string   `json:"cursor,omitempty" query:"cursor"`
	Limit     int      `json:"limit,omitempty" query:"limit" validate:"omitempty,min=1,max=100"`
	Status    []string `json:"status,omitempty" query:"status"`
	Priority  []string `json:"priority,omitempty" query:"priority"`
	Assignee  string   `json:"assignee,omitempty" query:"assignee"`
	IssueType string   `json:"issue_type,omitempty" query:"issue_type"`
	Search    string   `json:"search,omitempty" query:"search"`
	OrderBy   string   `json:"order_by,omitempty" query:"order_by" validate:"omitempty,oneof=created_at priority updated_at"`
}

// CreateIssueLinkRequest — тело запроса на создание связи между двумя задачами.
type CreateIssueLinkRequest struct {
	TargetIssueID string `json:"target_issue_id" validate:"required"`
	LinkType      string `json:"link_type" validate:"required"`
	CustomType    string `json:"custom_type,omitempty"`
}

// CreateCommentRequest — тело запроса на добавление комментария к задаче.
type CreateCommentRequest struct {
	BodyJSON   map[string]interface{} `json:"body_json" validate:"required"`
	BodyText   string                 `json:"body_text" validate:"required"`
	IsInternal bool                   `json:"is_internal"`
	Mentions   []string               `json:"mentions,omitempty"`
}

// UpdateCommentRequest — тело запроса на редактирование комментария.
type UpdateCommentRequest struct {
	BodyJSON map[string]interface{} `json:"body_json" validate:"required"`
	BodyText string                 `json:"body_text" validate:"required"`
}

// LogTimeRequest — тело запроса на списание времени на задачу.
type LogTimeRequest struct {
	Minutes     int    `json:"minutes" validate:"required,min=1"`
	Description string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// CreateLabelRequest — тело запроса на создание метки.
type CreateLabelRequest struct {
	Name    string `json:"name" validate:"required,min=1,max=50"`
	Color   string `json:"color" validate:"required"`
	GroupID string `json:"group_id,omitempty"`
}

// CreateLabelGroupRequest — тело запроса на создание группы меток.
type CreateLabelGroupRequest struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
}

// CreateWorkflowStatusRequest — тело запроса на создание статуса workflow.
type CreateWorkflowStatusRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=100"`
	Category  string `json:"category" validate:"required,oneof=todo in_progress done"`
	Color     string `json:"color,omitempty"`
	SortOrder int    `json:"sort_order,omitempty"`
}

// CreateWorkflowTransitionRequest — тело запроса на создание перехода workflow.
type CreateWorkflowTransitionRequest struct {
	FromStatusID string `json:"from_status_id" validate:"required"`
	ToStatusID   string `json:"to_status_id" validate:"required"`
}

// CreateBoardRequest — тело запроса на создание доски.
type CreateBoardRequest struct {
	Name    string                 `json:"name" validate:"required,min=1,max=100"`
	Filters map[string]interface{} `json:"filters,omitempty"`
}

// CreateCustomFieldRequest — тело запроса на создание определения кастомного поля.
type CreateCustomFieldRequest struct {
	Name      string   `json:"name" validate:"required,min=1,max=100"`
	FieldType string   `json:"field_type" validate:"required,oneof=text number date checkbox select multi_select user url"`
	Options   []string `json:"options,omitempty"`
	Required  bool     `json:"required,omitempty"`
}

// StartTimerRequest — тело запроса на запуск таймера на задаче.
type StartTimerRequest struct{}

// StopTimerRequest — тело запроса на остановку запущенного таймера на задаче.
type StopTimerRequest struct{}

// CreateReactionRequest — тело запроса на переключение реакции на комментарий.
type CreateReactionRequest struct {
	Emoji string `json:"emoji" validate:"required,max=32"`
}

// ---------------------------------------------------------------------------
// Response DTOs
// ---------------------------------------------------------------------------

// IssueResponse — полное представление задачи в API-ответах.
type IssueResponse struct {
	ID               string                  `json:"id"`
	IssueNumber      int                     `json:"issue_number"`
	IssueKey         string                  `json:"issue_key"`
	Title            string                  `json:"title"`
	DescriptionJSON  *map[string]interface{} `json:"description_json,omitempty"`
	DescriptionText  string                  `json:"description_text,omitempty"`
	IssueType        string                  `json:"issue_type"`
	Priority         string                  `json:"priority"`
	Rank             string                  `json:"rank,omitempty"`
	Version          int                     `json:"version"`
	EstimateMinutes  *int                    `json:"estimate_minutes,omitempty"`
	TimeSpentMinutes int                     `json:"time_spent_minutes"`
	CustomFields     map[string]interface{}  `json:"custom_fields,omitempty"`
	DueDate          *time.Time              `json:"due_date,omitempty"`
	Status           WorkflowStatusResponse  `json:"status"`
	Assignee         *UserBriefResponse      `json:"assignee,omitempty"`
	Reporter         *UserBriefResponse      `json:"reporter,omitempty"`
	Parent           *IssueBriefResponse     `json:"parent,omitempty"`
	ChildrenCount    int                     `json:"children_count"`
	Labels           []LabelResponse         `json:"labels,omitempty"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
}

// IssueBriefResponse — краткое представление задачи для ссылок и списков.
type IssueBriefResponse struct {
	ID        string `json:"id"`
	IssueKey  string `json:"issue_key"`
	Title     string `json:"title"`
	IssueType string `json:"issue_type"`
	Priority  string `json:"priority"`
	Status    string `json:"status"`
}

// UserBriefResponse — краткое представление пользователя для встроенных ссылок.
type UserBriefResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// WorkflowStatusResponse — представление статуса workflow в API-ответах.
type WorkflowStatusResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	IsDefault bool   `json:"is_default"`
}

// WorkflowTransitionResponse — представление перехода workflow в API-ответах.
type WorkflowTransitionResponse struct {
	ID           string `json:"id"`
	FromStatusID string `json:"from_status_id"`
	ToStatusID   string `json:"to_status_id"`
}

// CommentResponse — представление комментария в API-ответах.
type CommentResponse struct {
	ID         string                 `json:"id"`
	BodyJSON   map[string]interface{} `json:"body_json"`
	BodyText   string                 `json:"body_text"`
	IsInternal bool                   `json:"is_internal"`
	Mentions   []string               `json:"mentions,omitempty"`
	Author     UserBriefResponse      `json:"author"`
	Reactions  []ReactionResponse     `json:"reactions,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// ReactionResponse — агрегированная реакция в API-ответах (emoji + количество + пользователи).
type ReactionResponse struct {
	Emoji      string              `json:"emoji"`
	Count      int                 `json:"count"`
	Users      []UserBriefResponse `json:"users"`
	HasReacted bool                `json:"has_reacted"`
}

// ActivityLogResponse — запись лога активности в API-ответах.
type ActivityLogResponse struct {
	ID        string            `json:"id"`
	Action    string            `json:"action"`
	Changes   []FieldChange     `json:"changes,omitempty"`
	User      UserBriefResponse `json:"user"`
	CreatedAt time.Time         `json:"created_at"`
}

// FieldChange — изменение одного поля задачи (старое и новое значения) для лога активности.
type FieldChange struct {
	Field    string `json:"field"`
	OldValue any    `json:"old_value"`
	NewValue any    `json:"new_value"`
}

// TimeLogResponse — запись учёта времени в API-ответах.
type TimeLogResponse struct {
	ID          string            `json:"id"`
	Minutes     int               `json:"minutes"`
	Description string            `json:"description,omitempty"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	EndedAt     *time.Time        `json:"ended_at,omitempty"`
	User        UserBriefResponse `json:"user"`
	CreatedAt   time.Time         `json:"created_at"`
}

// LabelResponse — представление метки в API-ответах.
type LabelResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	GroupName string `json:"group_name,omitempty"`
}

// LabelGroupResponse — представление группы меток с вложенными метками в API-ответах.
type LabelGroupResponse struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Labels []LabelResponse `json:"labels,omitempty"`
}

// BoardResponse — представление доски в API-ответах.
type BoardResponse struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	IsDefault bool                   `json:"is_default"`
	CreatedAt time.Time              `json:"created_at"`
}

// BoardDataResponse — полные данные доски включая колонки и задачи.
type BoardDataResponse struct {
	Board     BoardResponse          `json:"board"`
	Columns   []BoardColumnResponse  `json:"columns"`
	Swimlanes []SwimlaneGroup        `json:"swimlanes,omitempty"`
}

// SwimlaneGroup — группа задач в swimlane (по assignee/priority/type).
type SwimlaneGroup struct {
	Key     string                `json:"key"`
	Label   string                `json:"label"`
	Columns []BoardColumnResponse `json:"columns"`
}

// BoardColumnResponse — одна колонка доски (статус + задачи в этом статусе).
type BoardColumnResponse struct {
	Status WorkflowStatusResponse `json:"status"`
	Issues []IssueBriefResponse   `json:"issues"`
}

// CustomFieldDefinitionResponse — определение кастомного поля в API-ответах.
type CustomFieldDefinitionResponse struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	FieldType string   `json:"field_type"`
	Options   []string `json:"options,omitempty"`
	Required  bool     `json:"required"`
	SortOrder int      `json:"sort_order"`
}

// IssueLinkResponse — представление связи между задачами в API-ответах.
type IssueLinkResponse struct {
	ID         string             `json:"id"`
	LinkType   string             `json:"link_type"`
	CustomType string             `json:"custom_type,omitempty"`
	Issue      IssueBriefResponse `json:"issue"`
}

// PaginatedResponse — обёртка для пагинированного ответа с курсором.
type PaginatedResponse[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Total      int    `json:"total"`
}
