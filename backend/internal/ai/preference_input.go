package ai

import "github.com/google/uuid"

// SaveUserPreferenceInput — частичное обновление ai_user_preference.
type SaveUserPreferenceInput struct {
	DefaultModel      *string
	DefaultProviderID *uuid.UUID
	AllowExternalAI   *bool
}
