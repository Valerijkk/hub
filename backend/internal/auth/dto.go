package auth

import "time"

// --- Request DTOs ---

// MagicLinkRequest — тело запроса для POST /auth/magic-link.
type MagicLinkRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// VerifyMagicLinkRequest — тело запроса для POST /auth/magic-link/verify.
type VerifyMagicLinkRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}

// TelegramAuthRequest — тело запроса для POST /auth/telegram.
type TelegramAuthRequest struct {
	InitData string `json:"init_data" validate:"required"`
}

// RefreshRequest — тело запроса для POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// UpdateProfileRequest — тело запроса для PUT /auth/me.
type UpdateProfileRequest struct {
	Name         *string                 `json:"name"`
	DisplayName  *string                 `json:"display_name"`
	AvatarURL    *string                 `json:"avatar_url"`
	Phone        *string                 `json:"phone"`
	Position     *string                 `json:"position"`
	Department   *string                 `json:"department"`
	Timezone     *string                 `json:"timezone"`
	Language     *string                 `json:"language"`
	Birthday     *time.Time              `json:"birthday"`
	CustomFields map[string]interface{}  `json:"custom_fields,omitempty"`
}

// SetRegistrationModeRequest — тело запроса для PUT /admin/registration-mode.
type SetRegistrationModeRequest struct {
	Mode string `json:"mode" validate:"required,oneof=open invite_only"`
}

// OnboardingCompleteRequest is the request body for POST /auth/me/onboarding.
type OnboardingCompleteRequest struct {
	ProfileData UpdateProfileRequest `json:"profile"`
}

// --- Response DTOs ---

// TokenPairResponse is the response for login/refresh endpoints.
type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// UserResponse is the response for GET /auth/me and profile updates.
type UserResponse struct {
	ID                  string                 `json:"id"`
	Email               string                 `json:"email"`
	Name                string                 `json:"name,omitempty"`
	DisplayName         string                 `json:"display_name,omitempty"`
	AvatarURL           string                 `json:"avatar_url,omitempty"`
	TelegramID          string                 `json:"telegram_id,omitempty"`
	Phone               string                 `json:"phone,omitempty"`
	Position            string                 `json:"position,omitempty"`
	Department          string                 `json:"department,omitempty"`
	Timezone            string                 `json:"timezone"`
	Language            string                 `json:"language"`
	Birthday            *time.Time             `json:"birthday,omitempty"`
	SystemRole          string                 `json:"system_role"`
	OnboardingCompleted bool                   `json:"onboarding_completed"`
	CustomFields        map[string]interface{} `json:"custom_fields,omitempty"`
}

// SessionResponse is a single session in the list returned by GET /auth/sessions.
type SessionResponse struct {
	ID           string    `json:"id"`
	Device       string    `json:"device"`
	IPAddress    string    `json:"ip_address,omitempty"`
	LastActiveAt time.Time `json:"last_active_at"`
	CreatedAt    time.Time `json:"created_at"`
	IsCurrent    bool      `json:"is_current"`
}
