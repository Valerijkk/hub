package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/platform/observability"
)

// ProfileService отвечает за чтение и обновление профиля пользователя.
type ProfileService struct {
	repo Repository
}

// NewProfileService создаёт ProfileService.
func NewProfileService(repo Repository) *ProfileService {
	return &ProfileService{repo: repo}
}

// GetMe возвращает пользователя по ID.
func (s *ProfileService) GetMe(ctx context.Context, userID uuid.UUID) (*ent.User, error) {
	return observability.Timed(ctx, "auth.GetMe", func(ctx context.Context) (*ent.User, error) {
		return s.repo.GetUser(ctx, userID)
	})
}

// UpdateProfile обновляет поля профиля пользователя. system_role никогда не изменяется.
func (s *ProfileService) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*ent.User, error) {
	return s.repo.UpdateUserProfile(ctx, userID, req)
}

// CompleteOnboarding обновляет профиль и помечает онбординг как завершённый в одной транзакции.
// W6-A1-010: сбой MarkOnboardingComplete откатывает UpdateProfile.
func (s *ProfileService) CompleteOnboarding(ctx context.Context, userID uuid.UUID, req OnboardingCompleteRequest) (*ent.User, error) {
	return s.repo.WithEntTx(ctx, func(ctx context.Context, txRepo Repository) (*ent.User, error) {
		if _, err := txRepo.UpdateUserProfile(ctx, userID, req.ProfileData); err != nil {
			return nil, err
		}
		return txRepo.MarkOnboardingComplete(ctx, userID)
	})
}

// UserToResponse конвертирует Ent User в DTO UserResponse.
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
		Status:              string(u.Status),
		OnboardingCompleted: u.OnboardingCompleted,
		CustomFields:        u.CustomFields,
	}
	if u.TelegramID != nil {
		resp.TelegramID = *u.TelegramID
	}
	return resp
}
