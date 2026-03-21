package sprint

import "time"

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateSprintRequest — тело запроса на создание спринта.
type CreateSprintRequest struct {
	Name      string    `json:"name" validate:"required,min=1,max=200"`
	Goal      string    `json:"goal,omitempty" validate:"omitempty,max=1000"`
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required"`
}

// UpdateSprintRequest — тело запроса на обновление спринта.
type UpdateSprintRequest struct {
	Name      *string    `json:"name,omitempty" validate:"omitempty,min=1,max=200"`
	Goal      *string    `json:"goal,omitempty" validate:"omitempty,max=1000"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// RetroNotesRequest — тело запроса на обновление заметок ретроспективы.
type RetroNotesRequest struct {
	WentWell  []string `json:"went_well"`
	ToImprove []string `json:"to_improve"`
	Actions   []string `json:"actions"`
}

// VoteRequest — тело запроса на голосование в Planning Poker.
type VoteRequest struct {
	Value string `json:"value" validate:"required"`
}

// StartPokerRequest — тело запроса на запуск сессии Planning Poker.
type StartPokerRequest struct {
	IssueID string `json:"issue_id" validate:"required"`
}

// MoveIssueToSprintRequest — тело запроса на перемещение задач в спринт.
type MoveIssueToSprintRequest struct {
	IssueIDs []string `json:"issue_ids" validate:"required,min=1"`
}

// ---------------------------------------------------------------------------
// Response DTOs
// ---------------------------------------------------------------------------

// SprintResponse — представление спринта в API-ответах.
type SprintResponse struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Goal      string                 `json:"goal,omitempty"`
	Status    string                 `json:"status"`
	StartDate time.Time              `json:"start_date"`
	EndDate   time.Time              `json:"end_date"`
	RetroNotes map[string]interface{} `json:"retro_notes,omitempty"`
	IssueCount int                   `json:"issue_count"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// BurndownPoint — точка данных для графика burndown.
type BurndownPoint struct {
	Date      string `json:"date"`
	Remaining int    `json:"remaining"`
	Ideal     int    `json:"ideal"`
}

// VelocityPoint — точка данных для графика velocity.
type VelocityPoint struct {
	SprintName string `json:"sprint_name"`
	Committed  int    `json:"committed"`
	Completed  int    `json:"completed"`
}

// PokerSessionResponse — ответ сессии Planning Poker.
type PokerSessionResponse struct {
	SessionID string            `json:"session_id"`
	IssueID   string            `json:"issue_id"`
	Revealed  bool              `json:"revealed"`
	Votes     map[string]string `json:"votes,omitempty"`
	Voters    []string          `json:"voters,omitempty"`
}
