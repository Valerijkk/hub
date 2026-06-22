package tracker

// IssueCreatedEvent публикуется через outbox при создании новой задачи.
type IssueCreatedEvent struct {
	IssueID     string `json:"issue_id"`
	IssueKey    string `json:"issue_key"`
	ProjectID   string `json:"project_id"`
	Title       string `json:"title"`
	IssueType   string `json:"issue_type"`
	CreatedByID string `json:"created_by_id"`
}

// IssueUpdatedEvent публикуется при обновлении задачи.
type IssueUpdatedEvent struct {
	IssueID     string        `json:"issue_id"`
	IssueKey    string        `json:"issue_key"`
	ProjectID   string        `json:"project_id"`
	Changes     []FieldChange `json:"changes"`
	UpdatedByID string        `json:"updated_by_id"`
}

// IssueMovedEvent публикуется при перемещении задачи в другой статус.
type IssueMovedEvent struct {
	IssueID     string `json:"issue_id"`
	IssueKey    string `json:"issue_key"`
	ProjectID   string `json:"project_id"`
	OldStatusID string `json:"old_status_id"`
	NewStatusID string `json:"new_status_id"`
	NewRank     string `json:"new_rank"`
	MovedByID   string `json:"moved_by_id"`
}

// IssueDeletedEvent публикуется при удалении задачи.
type IssueDeletedEvent struct {
	IssueID     string `json:"issue_id"`
	IssueKey    string `json:"issue_key"`
	ProjectID   string `json:"project_id"`
	DeletedByID string `json:"deleted_by_id"`
}

// IssueAssignedEvent публикуется при назначении исполнителя задачи.
type IssueAssignedEvent struct {
	IssueID      string `json:"issue_id"`
	IssueKey     string `json:"issue_key"`
	ProjectID    string `json:"project_id"`
	AssigneeID   string `json:"assignee_id"`
	AssignedByID string `json:"assigned_by_id"`
}

// CommentCreatedEvent публикуется при добавлении комментария к задаче.
type CommentCreatedEvent struct {
	CommentID string   `json:"comment_id"`
	IssueID   string   `json:"issue_id"`
	IssueKey  string   `json:"issue_key"`
	ProjectID string   `json:"project_id"`
	AuthorID  string   `json:"author_id"`
	Mentions  []string `json:"mentions,omitempty"`
}

// IssueLinkCreatedEvent публикуется при создании связи между задачами.
type IssueLinkCreatedEvent struct {
	SourceID  string `json:"source_id"`
	TargetID  string `json:"target_id"`
	LinkType  string `json:"link_type"`
	ProjectID string `json:"project_id"`
	CreatedBy string `json:"created_by"`
}

// TimerStartedEvent публикуется при запуске таймера на задаче.
type TimerStartedEvent struct {
	IssueID   string `json:"issue_id"`
	IssueKey  string `json:"issue_key"`
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
}

// TimerStoppedEvent публикуется при остановке таймера на задаче.
type TimerStoppedEvent struct {
	IssueID   string `json:"issue_id"`
	IssueKey  string `json:"issue_key"`
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
	Minutes   int    `json:"minutes"`
}
