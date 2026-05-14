package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
)

// ProfileService отвечает за чтение и обновление профиля пользователя
// (GetMe, UpdateProfile, CompleteOnboarding) — SC-14-1.
type ProfileService struct {
	db *ent.Client
}

// NewProfileService создаёт ProfileService.
func NewProfileService(db *ent.Client) *ProfileService {
	return &ProfileService{db: db}
}

// GetMe возвращает пользователя по ID.
func (s *ProfileService) GetMe(ctx context.Context, userID uuid.UUID) (*ent.User, error) {
	return s.db.User.Get(ctx, userID)
}

// UpdateProfile обновляет поля профиля пользователя. system_role никогда не изменяется.
func (s *ProfileService) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*ent.User, error) {
	update := s.db.User.UpdateOneID(userID)

	if req.Name != nil {
		update.SetName(*req.Name)
	}
	if req.DisplayName != nil {
		update.SetDisplayName(*req.DisplayName)
	}
	if req.AvatarURL != nil {
		update.SetAvatarURL(*req.AvatarURL)
	}
	if req.Phone != nil {
		update.SetPhone(*req.Phone)
	}
	if req.Position != nil {
		update.SetPosition(*req.Position)
	}
	if req.Department != nil {
		update.SetDepartment(*req.Department)
	}
	if req.Timezone != nil {
		update.SetTimezone(*req.Timezone)
	}
	if req.Language != nil {
		update.SetLanguage(*req.Language)
	}
	if req.Birthday != nil {
		update.SetBirthday(*req.Birthday)
	}
	if req.QuietHoursStart != nil {
		update.SetQuietHoursStart(*req.QuietHoursStart)
	}
	if req.QuietHoursEnd != nil {
		update.SetQuietHoursEnd(*req.QuietHoursEnd)
	}
	if req.CustomFields != nil {
		update.SetCustomFields(req.CustomFields)
	}

	return update.Save(ctx)
}

// CompleteOnboarding обновляет профиль и помечает онбординг как завершённый.
func (s *ProfileService) CompleteOnboarding(ctx context.Context, userID uuid.UUID, req OnboardingCompleteRequest) (*ent.User, error) {
	if _, err := s.UpdateProfile(ctx, userID, req.ProfileData); err != nil {
		return nil, err
	}
	return s.db.User.UpdateOneID(userID).
		SetOnboardingCompleted(true).
		Save(ctx)
}

// UserToResponse конвертирует Ent User в DTO UserResponse.
// Stateless — оставлена package-level функцией (вызывается из handler.go без сервиса).
func UserToResponse(u *ent.User) *UserResponse {
	resp := &UserResponse{
		ID:                  u.ID.String(),
		Email:               u.Email,
		Name:                u.Name,
		DisplayName:         u.DisplayName,
		AvatarURL:           u.AvatarURL,
		Phone:               u.Phone,
		Position:            u.Position,
		Department:          u.Department,
		Timezone:            u.Timezone,
		Language:            u.Language,
		Birthday:            u.Birthday,
		SystemRole:          string(u.SystemRole),
		OnboardingCompleted: u.OnboardingCompleted,
		CustomFields:        u.CustomFields,
	}
	if u.TelegramID != nil {
		resp.TelegramID = *u.TelegramID
	}
	return resp
}
