package ai

import "github.com/google/uuid"

// SaveUserPreferenceInput — частичное обновление ai_user_preference.
type SaveUserPreferenceInput struct {
	DefaultModel      *string
	DefaultProviderID *uuid.UUID
	AllowExternalAI   *bool
	// Version — optimistic lock при update существующей записи (A3W5-3-006).
	// 0 на create path игнорируется; на update обязателен.
	Version int
}
