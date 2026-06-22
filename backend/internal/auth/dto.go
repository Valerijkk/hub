package auth

import "time"

// --- DTO запросов ---

// MagicLinkRequest — тело запроса для POST /auth/magic-link.
type MagicLinkRequest struct {
	Email       string `json:"email" validate:"required,email"`
	InviteToken string `json:"invite_token,omitempty"`
}

// VerifyMagicLinkRequest — тело запроса для POST /auth/magic-link/verify.
type VerifyMagicLinkRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Code        string `json:"code" validate:"required,len=6"`
	InviteToken string `json:"invite_token,omitempty"`
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
	Name            *string                 `json:"name"`
	DisplayName     *string                 `json:"display_name"`
	AvatarURL       *string                 `json:"avatar_url"`
	Phone           *string                 `json:"phone"`
	Position        *string                 `json:"position"`
	Department      *string                 `json:"department"`
	Timezone        *string                 `json:"timezone"`
	Language        *string                 `json:"language"`
	Birthday        *time.Time              `json:"birthday"`
	QuietHoursStart *string                 `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd   *string                 `json:"quiet_hours_end,omitempty"`
	CustomFields    map[string]interface{}  `json:"custom_fields,omitempty"`
}

// PendingUserResponse — пользователь в очереди admin Pending (ORG-06).
type PendingUserResponse struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PendingSince time.Time `json:"pending_since"`
}

// OrgInviteResponse — invite-токен и URL org для администраторов.
type OrgInviteResponse struct {
	InviteToken string `json:"invite_token"`
	InviteURL   string `json:"invite_url"`
}

// SetRegistrationModeRequest — тело запроса для PUT /admin/registration-mode.
type SetRegistrationModeRequest struct {
	Mode string `json:"mode" validate:"required,oneof=open invite_only"`
}

// OnboardingCompleteRequest — тело запроса для POST /auth/me/onboarding.
type OnboardingCompleteRequest struct {
	ProfileData UpdateProfileRequest `json:"profile"`
}

// --- DTO ответов ---

// TokenPairResponse — ответ для эндпоинтов login/refresh.
type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// UserResponse — ответ для GET /auth/me и обновления профиля.
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
	Status              string                 `json:"status"`
	OnboardingCompleted bool                   `json:"onboarding_completed"`
	CustomFields        map[string]interface{} `json:"custom_fields,omitempty"`
}

// SessionResponse — одна сессия в списке, возвращаемом GET /auth/sessions.
type SessionResponse struct {
	ID           string    `json:"id"`
	Device       string    `json:"device"`
	IPAddress    string    `json:"ip_address,omitempty"`
	LastActiveAt time.Time `json:"last_active_at"`
	CreatedAt    time.Time `json:"created_at"`
	IsCurrent    bool      `json:"is_current"`
}
