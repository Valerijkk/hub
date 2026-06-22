package auth

import (
	"context"
	"database/sql"
	"sync"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
)

// RegistrationModeService отвечает за системную настройку режима регистрации
// (open / invite_only) и связанные проверки (SC-14-1).
type RegistrationModeService struct {
	repo Repository
	sqlDB *sql.DB

	// firstUserMu сериализует первую регистрацию вместе с advisory lock (SEC-06).
	firstUserMu sync.Mutex
}

// NewRegistrationModeService создаёт RegistrationModeService.
func NewRegistrationModeService(repo Repository, sqlDB *sql.DB) *RegistrationModeService {
	return &RegistrationModeService{repo: repo, sqlDB: sqlDB}
}

// GetRegistrationMode читает системную настройку "registration_mode". По умолчанию "invite_only" (ORG-05).
func (s *RegistrationModeService) GetRegistrationMode(ctx context.Context) (string, error) {
	mode, err := s.repo.GetSystemSettingValue(ctx, "registration_mode")
	if err != nil {
		return "", err
	}
	if mode == "" {
		return "invite_only", nil
	}
	return mode, nil
}

// SetRegistrationMode обновляет режим регистрации (только для администратора).
func (s *RegistrationModeService) SetRegistrationMode(ctx context.Context, mode string) error {
	return s.repo.UpsertSystemSetting(ctx, "registration_mode", mode,
		"Controls whether new users can register: open or invite_only")
}

// CheckRegistrationAllowed проверяет, разрешена ли регистрация в текущем режиме.
func (s *RegistrationModeService) CheckRegistrationAllowed(ctx context.Context) error {
	mode, _ := s.GetRegistrationMode(ctx)
	if mode == "invite_only" {
		return ErrRegistrationInviteOnly
	}
	return nil
}

// IsFirstUser проверяет, является ли текущая регистрация первой.
func (s *RegistrationModeService) IsFirstUser(ctx context.Context) bool {
	count, err := s.repo.CountUsers(ctx)
	if err != nil {
		return false
	}
	return count == 0
}

// CreateFirstOrMemberUser атомарно выбирает SystemRole и создаёт пользователя.
func (s *RegistrationModeService) CreateFirstOrMemberUser(
	ctx context.Context,
	apply func(b *ent.UserCreate, role user.SystemRole) *ent.UserCreate,
) (*ent.User, error) {
	s.firstUserMu.Lock()
	defer s.firstUserMu.Unlock()

	run := func(ctx context.Context, txRepo Repository) (*ent.User, error) {
		return txRepo.CreateFirstOrMemberUser(ctx, apply)
	}
	if s.sqlDB != nil {
		return s.repo.WithPostgresAdvisoryTx(ctx, s.sqlDB, run)
	}
	return s.repo.WithEntTx(ctx, run)
}

// RegisterNewUserByEmail создаёт founder+org или pending/active member (ORG-01..06).
func (s *RegistrationModeService) RegisterNewUserByEmail(
	ctx context.Context,
	email, inviteToken string,
	_ *OrganizationService,
) (*ent.User, error) {
	s.firstUserMu.Lock()
	defer s.firstUserMu.Unlock()

	run := func(ctx context.Context, txRepo Repository) (*ent.User, error) {
		return txRepo.RegisterNewUserByEmail(ctx, email, inviteToken)
	}
	if s.sqlDB != nil {
		return s.repo.WithPostgresAdvisoryTx(ctx, s.sqlDB, run)
	}
	return s.repo.WithEntTx(ctx, run)
}
