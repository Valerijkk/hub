package project

// ProjectCreatedEvent публикуется через outbox при создании проекта.
type ProjectCreatedEvent struct {
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	CreatedByID string `json:"created_by_id"`
}

// MemberAddedEvent публикуется при добавлении участника в проект.
type MemberAddedEvent struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
}

// MemberRemovedEvent публикуется при удалении участника из проекта.
type MemberRemovedEvent struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
}
