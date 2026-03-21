package calls

import (
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Recording DTOs (CALL-04, CALL-05)
// ---------------------------------------------------------------------------

// RecordingResponse содержит данные записи звонка для ответа API.
type RecordingResponse struct {
	ID              uuid.UUID  `json:"id"`
	EgressID        string     `json:"egress_id"`
	Status          string     `json:"status"`
	FilePath        string     `json:"file_path,omitempty"`
	FileSize        int64      `json:"file_size"`
	DurationSeconds int        `json:"duration_seconds"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	CallID          uuid.UUID  `json:"call_id"`
	ProjectID       uuid.UUID  `json:"project_id"`
	StartedBy       uuid.UUID  `json:"started_by"`
	CreatedAt       time.Time  `json:"created_at"`
}

// RecordingDownloadResponse содержит presigned URL для скачивания записи.
type RecordingDownloadResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ---------------------------------------------------------------------------
// ScheduledCall DTOs (CALL-02)
// ---------------------------------------------------------------------------

// CreateScheduledCallRequest содержит данные для создания запланированного звонка.
type CreateScheduledCallRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	ScheduledAt time.Time `json:"scheduled_at"`
	RRule       string    `json:"rrule,omitempty"`
	ChannelID   uuid.UUID `json:"channel_id"`
}

// UpdateScheduledCallRequest содержит данные для обновления запланированного звонка.
type UpdateScheduledCallRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	RRule       *string    `json:"rrule,omitempty"`
	IsActive    *bool      `json:"is_active,omitempty"`
}

// ScheduledCallResponse содержит данные запланированного звонка для ответа API.
type ScheduledCallResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	ScheduledAt time.Time `json:"scheduled_at"`
	RRule       string    `json:"rrule,omitempty"`
	ChannelID   uuid.UUID `json:"channel_id"`
	ProjectID   uuid.UUID `json:"project_id"`
	CreatedBy   uuid.UUID `json:"created_by"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
