package tracker

// IssueCreatedEvent is published via outbox when a new issue is created.
type IssueCreatedEvent struct {
	IssueID     string `json:"issue_id"`
	IssueKey    string `json:"issue_key"`
	ProjectID   string `json:"project_id"`
	Title       string `json:"title"`
	IssueType   string `json:"issue_type"`
	CreatedByID string `json:"created_by_id"`
}

// IssueUpdatedEvent is published when an issue is updated.
type IssueUpdatedEvent struct {
	IssueID     string        `json:"issue_id"`
	IssueKey    string        `json:"issue_key"`
	ProjectID   string        `json:"project_id"`
	Changes     []FieldChange `json:"changes"`
	UpdatedByID string        `json:"updated_by_id"`
}

// IssueMovedEvent is published when an issue is moved to a different status.
type IssueMovedEvent struct {
	IssueID     string `json:"issue_id"`
	IssueKey    string `json:"issue_key"`
	ProjectID   string `json:"project_id"`
	OldStatusID string `json:"old_status_id"`
	NewStatusID string `json:"new_status_id"`
	NewRank     string `json:"new_rank"`
	MovedByID   string `json:"moved_by_id"`
}

// IssueDeletedEvent is published when an issue is deleted.
type IssueDeletedEvent struct {
	IssueID     string `json:"issue_id"`
	IssueKey    string `json:"issue_key"`
	ProjectID   string `json:"project_id"`
	DeletedByID string `json:"deleted_by_id"`
}

// IssueAssignedEvent is published when an issue is assigned to a user.
type IssueAssignedEvent struct {
	IssueID      string `json:"issue_id"`
	IssueKey     string `json:"issue_key"`
	ProjectID    string `json:"project_id"`
	AssigneeID   string `json:"assignee_id"`
	AssignedByID string `json:"assigned_by_id"`
}

// CommentCreatedEvent is published when a comment is added to an issue.
type CommentCreatedEvent struct {
	CommentID string   `json:"comment_id"`
	IssueID   string   `json:"issue_id"`
	IssueKey  string   `json:"issue_key"`
	ProjectID string   `json:"project_id"`
	AuthorID  string   `json:"author_id"`
	Mentions  []string `json:"mentions,omitempty"`
}

// TimerStartedEvent is published when a user starts a timer on an issue.
type TimerStartedEvent struct {
	IssueID   string `json:"issue_id"`
	IssueKey  string `json:"issue_key"`
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
}

// TimerStoppedEvent is published when a user stops a timer on an issue.
type TimerStoppedEvent struct {
	IssueID   string `json:"issue_id"`
	IssueKey  string `json:"issue_key"`
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
	Minutes   int    `json:"minutes"`
}
