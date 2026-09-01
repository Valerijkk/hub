package calendar

import (
	"time"

	"github.com/google/uuid"
)

// EventType — тип события в едином календаре (MOD-04).
type EventType string

const (
	EventTypeEvent    EventType = "event"
	EventTypeCall     EventType = "call"
	EventTypeSprint   EventType = "sprint"
	EventTypeDeadline EventType = "deadline"
	EventTypeLeave    EventType = "leave"
)

// EventResponse — одно событие календаря.
type EventResponse struct {
	ID          string     `json:"id"`
	Type        EventType  `json:"type"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	StartsAt    time.Time  `json:"starts_at"`
	EndsAt      *time.Time `json:"ends_at,omitempty"`
	AllDay      bool       `json:"all_day"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
	ProjectKey  string     `json:"project_key,omitempty"`
	URL         string     `json:"url,omitempty"`
}

// EventsResponse — список событий за период.
type EventsResponse struct {
	Events []EventResponse `json:"events"`
	From   time.Time       `json:"from"`
	To     time.Time       `json:"to"`
}
