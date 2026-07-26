package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
)

const pendingUserTTL = 7 * 24 * time.Hour

// OrganizationService управляет единственной организацией на инстанс (ORG-01).
type OrganizationService struct {
	repo Repository
}

// NewOrganizationService создаёт OrganizationService.
func NewOrganizationService(repo Repository) *OrganizationService {
	return &OrganizationService{repo: repo}
}

// IsSystemAdmin возвращает true для ролей founder и admin (ORG-04).
func IsSystemAdmin(role user.SystemRole) bool {
	return role == user.SystemRoleFounder || role == user.SystemRoleAdmin
}

// HasOrganization сообщает, существует ли запись организации в БД.
func (s *OrganizationService) HasOrganization(ctx context.Context) (bool, error) {
	n, err := s.repo.CountOrganizations(ctx)
	if err != nil {
		return false, fmt.Errorf("count organizations: %w", err)
	}
	return n > 0, nil
}

// GetOrganization возвращает единственную организацию инстанса.
func (s *OrganizationService) GetOrganization(ctx context.Context) (*ent.Organization, error) {
	org, err := s.repo.GetOrganization(ctx)
	if err != nil {
		return nil, fmt.Errorf("get organization: %w", err)
	}
	return org, nil
}

// ValidateInviteToken проверяет invite-токен организации (ORG-05).
func (s *OrganizationService) ValidateInviteToken(ctx context.Context, token string) error {
	if token == "" {
		return ErrInvalidInviteToken
	}
	ok, err := s.repo.OrganizationExistsByInviteToken(ctx, token)
	if err != nil {
		return fmt.Errorf("validate invite token: %w", err)
	}
	if !ok {
		return ErrInvalidInviteToken
	}
	return nil
}

// ListPendingUsers возвращает пользователей в статусе pending.
func (s *OrganizationService) ListPendingUsers(ctx context.Context) ([]*ent.User, error) {
	users, err := s.repo.ListPendingUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list pending users: %w", err)
	}
	return users, nil
}

// ApprovePendingUser активирует pending-пользователя (ORG-06).
func (s *OrganizationService) ApprovePendingUser(ctx context.Context, userID uuid.UUID) error {
	return s.repo.ApprovePendingUser(ctx, userID)
}

// RejectPendingUser удаляет pending-регистрацию.
func (s *OrganizationService) RejectPendingUser(ctx context.Context, userID uuid.UUID) error {
	return s.repo.RejectPendingUser(ctx, userID)
}

// PurgeExpiredPending удаляет pending-пользователей старше 7 дней (ORG-06).
func (s *OrganizationService) PurgeExpiredPending(ctx context.Context) (int, error) {
	cutoff := time.Now().UTC().Add(-pendingUserTTL)
	n, err := s.repo.PurgeExpiredPending(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("purge expired pending: %w", err)
	}
	return n, nil
}

// ErrInvalidInviteToken — invite-токен org отсутствует или неверен.
var ErrInvalidInviteToken = fmt.Errorf("invalid invite token")

// sentinel-ошибки регистрации.
var (
	ErrRegistrationInviteOnly = fmt.Errorf("registration is by invite only")
	ErrUserNotFound           = fmt.Errorf("user not found")
	// ErrAccountNotActive — login/refresh запрещены для pending/deactivated (ORG-06 / ECC-AUTH-224).
	ErrAccountNotActive = fmt.Errorf("account is not active")
)

// generateInviteToken генерирует криптостойкий invite-токен org (32 hex-символа).
func generateInviteToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate invite token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
