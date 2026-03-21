package tracker

import "time"

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateIssueRequest is the request body for creating a new issue.
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

// UpdateIssueRequest is the request body for updating an existing issue.
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

// MoveIssueRequest is the request body for moving an issue on a board.
type MoveIssueRequest struct {
	StatusID string `json:"status_id" validate:"required"`
	Rank     string `json:"rank" validate:"required"`
}

// ListIssuesRequest contains filter and pagination parameters for listing issues.
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

// CreateIssueLinkRequest is the request body for linking two issues.
type CreateIssueLinkRequest struct {
	TargetIssueID string `json:"target_issue_id" validate:"required"`
	LinkType      string `json:"link_type" validate:"required"`
	CustomType    string `json:"custom_type,omitempty"`
}

// CreateCommentRequest is the request body for adding a comment to an issue.
type CreateCommentRequest struct {
	BodyJSON   map[string]interface{} `json:"body_json" validate:"required"`
	BodyText   string                 `json:"body_text" validate:"required"`
	IsInternal bool                   `json:"is_internal"`
	Mentions   []string               `json:"mentions,omitempty"`
}

// UpdateCommentRequest is the request body for editing a comment.
type UpdateCommentRequest struct {
	BodyJSON map[string]interface{} `json:"body_json" validate:"required"`
	BodyText string                 `json:"body_text" validate:"required"`
}

// LogTimeRequest is the request body for logging time on an issue.
type LogTimeRequest struct {
	Minutes     int    `json:"minutes" validate:"required,min=1"`
	Description string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// CreateLabelRequest is the request body for creating a label.
type CreateLabelRequest struct {
	Name    string `json:"name" validate:"required,min=1,max=50"`
	Color   string `json:"color" validate:"required"`
	GroupID string `json:"group_id,omitempty"`
}

// CreateLabelGroupRequest is the request body for creating a label group.
type CreateLabelGroupRequest struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
}

// CreateWorkflowStatusRequest is the request body for creating a workflow status.
type CreateWorkflowStatusRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=100"`
	Category  string `json:"category" validate:"required,oneof=todo in_progress done"`
	Color     string `json:"color,omitempty"`
	SortOrder int    `json:"sort_order,omitempty"`
}

// CreateWorkflowTransitionRequest is the request body for creating a workflow transition.
type CreateWorkflowTransitionRequest struct {
	FromStatusID string `json:"from_status_id" validate:"required"`
	ToStatusID   string `json:"to_status_id" validate:"required"`
}

// CreateBoardRequest is the request body for creating a board.
type CreateBoardRequest struct {
	Name    string                 `json:"name" validate:"required,min=1,max=100"`
	Filters map[string]interface{} `json:"filters,omitempty"`
}

// CreateCustomFieldRequest is the request body for creating a custom field definition.
type CreateCustomFieldRequest struct {
	Name      string   `json:"name" validate:"required,min=1,max=100"`
	FieldType string   `json:"field_type" validate:"required,oneof=text number date checkbox select multi_select user url"`
	Options   []string `json:"options,omitempty"`
	Required  bool     `json:"required,omitempty"`
}

// StartTimerRequest is the request body for starting a timer on an issue.
type StartTimerRequest struct{}

// StopTimerRequest is the request body for stopping a running timer on an issue.
type StopTimerRequest struct{}

// CreateReactionRequest is the request body for toggling a reaction on a comment.
type CreateReactionRequest struct {
	Emoji string `json:"emoji" validate:"required,max=32"`
}

// ---------------------------------------------------------------------------
// Response DTOs
// ---------------------------------------------------------------------------

// IssueResponse represents an issue in API responses.
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

// IssueBriefResponse is a compact issue representation used in references and lists.
type IssueBriefResponse struct {
	ID        string `json:"id"`
	IssueKey  string `json:"issue_key"`
	Title     string `json:"title"`
	IssueType string `json:"issue_type"`
	Priority  string `json:"priority"`
	Status    string `json:"status"`
}

// UserBriefResponse is a compact user representation for embedded references.
type UserBriefResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// WorkflowStatusResponse represents a workflow status in API responses.
type WorkflowStatusResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	IsDefault bool   `json:"is_default"`
}

// WorkflowTransitionResponse represents a workflow transition in API responses.
type WorkflowTransitionResponse struct {
	ID           string `json:"id"`
	FromStatusID string `json:"from_status_id"`
	ToStatusID   string `json:"to_status_id"`
}

// CommentResponse represents a comment in API responses.
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

// ReactionResponse represents an aggregated reaction in API responses.
type ReactionResponse struct {
	Emoji      string              `json:"emoji"`
	Count      int                 `json:"count"`
	Users      []UserBriefResponse `json:"users"`
	HasReacted bool                `json:"has_reacted"`
}

// ActivityLogResponse represents an activity log entry in API responses.
type ActivityLogResponse struct {
	ID        string            `json:"id"`
	Action    string            `json:"action"`
	Changes   []FieldChange     `json:"changes,omitempty"`
	User      UserBriefResponse `json:"user"`
	CreatedAt time.Time         `json:"created_at"`
}

// FieldChange represents a single field-level diff in an activity log entry.
type FieldChange struct {
	Field    string `json:"field"`
	OldValue any    `json:"old_value"`
	NewValue any    `json:"new_value"`
}

// TimeLogResponse represents a time log entry in API responses.
type TimeLogResponse struct {
	ID          string            `json:"id"`
	Minutes     int               `json:"minutes"`
	Description string            `json:"description,omitempty"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	EndedAt     *time.Time        `json:"ended_at,omitempty"`
	User        UserBriefResponse `json:"user"`
	CreatedAt   time.Time         `json:"created_at"`
}

// LabelResponse represents a label in API responses.
type LabelResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	GroupName string `json:"group_name,omitempty"`
}

// LabelGroupResponse represents a label group with its labels in API responses.
type LabelGroupResponse struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Labels []LabelResponse `json:"labels,omitempty"`
}

// BoardResponse represents a board in API responses.
type BoardResponse struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	IsDefault bool                   `json:"is_default"`
	CreatedAt time.Time              `json:"created_at"`
}

// BoardDataResponse represents full board data including columns and issues.
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

// BoardColumnResponse represents a single column on a board.
type BoardColumnResponse struct {
	Status WorkflowStatusResponse `json:"status"`
	Issues []IssueBriefResponse   `json:"issues"`
}

// CustomFieldDefinitionResponse represents a custom field definition in API responses.
type CustomFieldDefinitionResponse struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	FieldType string   `json:"field_type"`
	Options   []string `json:"options,omitempty"`
	Required  bool     `json:"required"`
	SortOrder int      `json:"sort_order"`
}

// IssueLinkResponse represents an issue link in API responses.
type IssueLinkResponse struct {
	ID         string             `json:"id"`
	LinkType   string             `json:"link_type"`
	CustomType string             `json:"custom_type,omitempty"`
	Issue      IssueBriefResponse `json:"issue"`
}

// PaginatedResponse is a generic paginated response wrapper.
type PaginatedResponse[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Total      int    `json:"total"`
}
