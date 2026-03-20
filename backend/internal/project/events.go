package project

// ProjectCreatedEvent is published via outbox when a new project is created.
type ProjectCreatedEvent struct {
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	CreatedByID string `json:"created_by_id"`
}

// MemberAddedEvent is published when a member is added to a project.
type MemberAddedEvent struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
}

// MemberRemovedEvent is published when a member is removed from a project.
type MemberRemovedEvent struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
}
