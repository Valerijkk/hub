package tracker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// Repository defines the data access interface for the tracker module.
// This allows the service layer to be tested with mock implementations.
type Repository interface {
	// Issue CRUD

	// CreateIssue creates a new issue in the given project.
	CreateIssue(ctx context.Context, projectID uuid.UUID, input CreateIssueInput) (*ent.Issue, error)
	// FindIssueByKey returns an issue by its project-prefixed key (e.g. "HUB-1").
	FindIssueByKey(ctx context.Context, issueKey string) (*ent.Issue, error)
	// FindIssueByID returns an issue by its UUID.
	FindIssueByID(ctx context.Context, id uuid.UUID) (*ent.Issue, error)
	// ListIssues returns a paginated list of issues for a project, with total count.
	ListIssues(ctx context.Context, projectID uuid.UUID, params ListIssuesParams) ([]*ent.Issue, int, error)
	// UpdateIssue applies partial updates to an issue.
	UpdateIssue(ctx context.Context, issueID uuid.UUID, input UpdateIssueInput) (*ent.Issue, error)
	// DeleteIssue permanently removes an issue by ID.
	DeleteIssue(ctx context.Context, issueID uuid.UUID) error
	// MoveIssue changes the status and rank of an issue (Kanban board move).
	MoveIssue(ctx context.Context, issueID uuid.UUID, statusID uuid.UUID, rank string) (*ent.Issue, error)

	// Issue links

	// CreateIssueLink creates a directional link between two issues.
	CreateIssueLink(ctx context.Context, sourceID, targetID uuid.UUID, linkType string, customType string) (*ent.IssueLink, error)
	// ListIssueLinks returns all links (incoming and outgoing) for an issue.
	ListIssueLinks(ctx context.Context, issueID uuid.UUID) ([]*ent.IssueLink, error)
	// DeleteIssueLink removes a link between two issues.
	DeleteIssueLink(ctx context.Context, linkID uuid.UUID) error

	// Workflow

	// ListWorkflowStatuses returns all workflow statuses for a project, ordered by sort_order.
	ListWorkflowStatuses(ctx context.Context, projectID uuid.UUID) ([]*ent.WorkflowStatus, error)
	// CreateWorkflowStatus creates a new workflow status in a project.
	CreateWorkflowStatus(ctx context.Context, projectID uuid.UUID, name, category, color string, sortOrder int) (*ent.WorkflowStatus, error)
	// UpdateWorkflowStatus applies partial updates to a workflow status.
	UpdateWorkflowStatus(ctx context.Context, statusID uuid.UUID, name, color *string, sortOrder *int) (*ent.WorkflowStatus, error)
	// DeleteWorkflowStatus removes a workflow status by ID.
	DeleteWorkflowStatus(ctx context.Context, statusID uuid.UUID) error
	// FindDefaultStatus returns the default workflow status for a project.
	FindDefaultStatus(ctx context.Context, projectID uuid.UUID) (*ent.WorkflowStatus, error)
	// ListWorkflowTransitions returns all allowed transitions for a project.
	ListWorkflowTransitions(ctx context.Context, projectID uuid.UUID) ([]*ent.WorkflowTransition, error)
	// CreateWorkflowTransition creates a new allowed transition between two statuses.
	CreateWorkflowTransition(ctx context.Context, fromStatusID, toStatusID uuid.UUID) (*ent.WorkflowTransition, error)
	// DeleteWorkflowTransition removes a workflow transition by ID.
	DeleteWorkflowTransition(ctx context.Context, transitionID uuid.UUID) error
	// FindTransition checks whether a transition between two statuses is allowed.
	FindTransition(ctx context.Context, fromStatusID, toStatusID uuid.UUID) (*ent.WorkflowTransition, error)
	// SeedDefaultWorkflow creates the default workflow statuses and transitions for a new project.
	SeedDefaultWorkflow(ctx context.Context, projectID uuid.UUID) error

	// Comments

	// CreateComment adds a new comment to an issue.
	CreateComment(ctx context.Context, issueID, authorID uuid.UUID, input CreateCommentInput) (*ent.Comment, error)
	// ListComments returns all comments for an issue.
	ListComments(ctx context.Context, issueID uuid.UUID) ([]*ent.Comment, error)
	// UpdateComment edits a comment's content.
	UpdateComment(ctx context.Context, commentID uuid.UUID, input UpdateCommentInput) (*ent.Comment, error)
	// DeleteComment removes a comment by ID.
	DeleteComment(ctx context.Context, commentID uuid.UUID) error

	// Reactions

	// ToggleReaction adds or removes a reaction. Returns true if added, false if removed.
	ToggleReaction(ctx context.Context, commentID, userID uuid.UUID, emoji string) (bool, error)

	// Activity log

	// LogActivity records an activity entry for an issue.
	LogActivity(ctx context.Context, issueID, userID uuid.UUID, action string, changes []FieldChange) (*ent.ActivityLog, error)
	// ListActivities returns all activity log entries for an issue.
	ListActivities(ctx context.Context, issueID uuid.UUID) ([]*ent.ActivityLog, error)

	// Time tracking

	// LogTime records manually entered time on an issue.
	LogTime(ctx context.Context, issueID, userID uuid.UUID, minutes int, description string) (*ent.TimeLog, error)
	// StartTimer starts a running timer on an issue for a user.
	StartTimer(ctx context.Context, issueID, userID uuid.UUID) (*ent.TimeLog, error)
	// StopTimer stops a running timer and records elapsed minutes.
	StopTimer(ctx context.Context, issueID, userID uuid.UUID) (*ent.TimeLog, error)
	// ListTimeLogs returns all time log entries for an issue.
	ListTimeLogs(ctx context.Context, issueID uuid.UUID) ([]*ent.TimeLog, error)

	// Labels

	// CreateLabel creates a new label in a project, optionally in a group.
	CreateLabel(ctx context.Context, projectID uuid.UUID, name, color string, groupID *uuid.UUID) (*ent.Label, error)
	// ListLabels returns all labels for a project.
	ListLabels(ctx context.Context, projectID uuid.UUID) ([]*ent.Label, error)
	// DeleteLabel removes a label by ID.
	DeleteLabel(ctx context.Context, labelID uuid.UUID) error
	// CreateLabelGroup creates a new label group in a project.
	CreateLabelGroup(ctx context.Context, projectID uuid.UUID, name string) (*ent.LabelGroup, error)
	// ListLabelGroups returns all label groups for a project, with their labels.
	ListLabelGroups(ctx context.Context, projectID uuid.UUID) ([]*ent.LabelGroup, error)

	// Boards

	// CreateBoard creates a new board in a project with optional saved filters.
	CreateBoard(ctx context.Context, projectID uuid.UUID, name string, filters map[string]interface{}) (*ent.Board, error)
	// ListBoards returns all boards for a project.
	ListBoards(ctx context.Context, projectID uuid.UUID) ([]*ent.Board, error)
	// GetBoardData returns a board by ID with its data for rendering.
	GetBoardData(ctx context.Context, boardID uuid.UUID) (*ent.Board, error)

	// Custom fields

	// CreateCustomField creates a new custom field definition in a project.
	CreateCustomField(ctx context.Context, projectID uuid.UUID, input CreateCustomFieldInput) (*ent.CustomFieldDefinition, error)
	// ListCustomFields returns all custom field definitions for a project.
	ListCustomFields(ctx context.Context, projectID uuid.UUID) ([]*ent.CustomFieldDefinition, error)
	// DeleteCustomField removes a custom field definition by ID.
	DeleteCustomField(ctx context.Context, fieldID uuid.UUID) error
}

// ---------------------------------------------------------------------------
// Repository input structs (use uuid.UUID, not string)
// ---------------------------------------------------------------------------

// CreateIssueInput holds parsed and validated data for creating an issue.
type CreateIssueInput struct {
	Title           string
	IssueType       string
	DescriptionJSON map[string]interface{}
	DescriptionText string
	Priority        string
	AssigneeID      *uuid.UUID
	ParentID        *uuid.UUID
	StatusID        *uuid.UUID
	ReporterID      uuid.UUID
	Labels          []uuid.UUID
	CustomFields    map[string]interface{}
	EstimateMinutes *int
	DueDate         *time.Time
}

// UpdateIssueInput holds parsed and validated data for updating an issue.
type UpdateIssueInput struct {
	Title           *string
	DescriptionJSON *map[string]interface{}
	DescriptionText *string
	Priority        *string
	AssigneeID      *uuid.UUID
	StatusID        *uuid.UUID
	Labels          *[]uuid.UUID
	CustomFields    *map[string]interface{}
	EstimateMinutes *int
	DueDate         *time.Time
	Version         int
}

// ListIssuesParams holds parsed filter and pagination parameters for listing issues.
type ListIssuesParams struct {
	Cursor    string
	Limit     int
	Status    []string
	Priority  []string
	Assignee  *uuid.UUID
	IssueType string
	Search    string
	OrderBy   string
}

// CreateCommentInput holds parsed data for creating a comment.
type CreateCommentInput struct {
	BodyJSON   map[string]interface{}
	BodyText   string
	IsInternal bool
	Mentions   []string
}

// UpdateCommentInput holds parsed data for updating a comment.
type UpdateCommentInput struct {
	BodyJSON map[string]interface{}
	BodyText string
}

// CreateCustomFieldInput holds parsed data for creating a custom field definition.
type CreateCustomFieldInput struct {
	Name      string
	FieldType string
	Options   []string
	Required  bool
}

// ---------------------------------------------------------------------------
// OutboxPublisher + EntRepository skeleton
// ---------------------------------------------------------------------------

// OutboxPublisher abstracts outbox event publishing for the repository.
type OutboxPublisher interface {
	PublishInTx(ctx context.Context, tx *ent.Tx, eventType string, payload any) error
}

// EntRepository implements Repository using Ent ORM with outbox event publishing.
type EntRepository struct {
	client    *ent.Client
	publisher OutboxPublisher
}

// NewEntRepository creates a new Ent-backed tracker repository.
func NewEntRepository(client *ent.Client, publisher OutboxPublisher) *EntRepository {
	return &EntRepository{client: client, publisher: publisher}
}
