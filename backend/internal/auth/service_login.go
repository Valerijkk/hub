package auth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/platform/observability"
)

// LoginService инкапсулирует Magic Link login-флоу (AUTH-01: Telegram не login).
type LoginService struct {
	repo      Repository
	magicLink *MagicLinkService
	sessions  *SessionService
	regMode   *RegistrationModeService
	org       *OrganizationService
}

// NewLoginService создаёт LoginService с явно переданными зависимостями.
func NewLoginService(
	repo Repository,
	magicLink *MagicLinkService,
	sessions *SessionService,
	regMode *RegistrationModeService,
	org *OrganizationService,
) *LoginService {
	return &LoginService{
		repo:      repo,
		magicLink: magicLink,
		sessions:  sessions,
		regMode:   regMode,
		org:       org,
	}
}

// UserExistsByEmail возвращает true, если в БД есть пользователь с указанным email.
func (s *LoginService) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	return s.repo.UserExistsByEmail(ctx, email)
}

// LoginWithMagicLink проверяет код, находит или создаёт пользователя и возвращает пару токенов.
func (s *LoginService) LoginWithMagicLink(ctx context.Context, email, code, inviteToken, device, ip string) (*TokenPairResponse, error) {
	return observability.Timed(ctx, "auth.Login", func(ctx context.Context) (*TokenPairResponse, error) {
		if err := s.magicLink.VerifyCode(ctx, email, code); err != nil {
			return nil, err
		}

		u, isNew, err := s.findOrCreateUserByEmail(ctx, email, inviteToken)
		if err != nil {
			return nil, err
		}

		if isNew {
			// A14-006: не логировать полный email.
			slog.Info("new user registered via magic link", "user_id", u.ID.String(), "email", maskEmailMinimized(email), "status", u.Status)
		}

		if u.Status != user.StatusActive {
			return nil, ErrAccountNotActive
		}

		return s.sessions.CreateSession(ctx, u.ID, device, ip)
	})
}

// CanSendMagicLink проверяет, можно ли отправить magic link для нового email (ORG-05).
func (s *LoginService) CanSendMagicLink(ctx context.Context, email, inviteToken string) error {
	exists, err := s.UserExistsByEmail(ctx, email)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	hasOrg, err := s.org.HasOrganization(ctx)
	if err != nil {
		return err
	}
	if !hasOrg {
		return nil
	}
	mode, err := s.regMode.GetRegistrationMode(ctx)
	if err != nil {
		return err
	}
	if mode != "invite_only" {
		return nil
	}
	return s.org.ValidateInviteToken(ctx, inviteToken)
}

func (s *LoginService) findOrCreateUserByEmail(ctx context.Context, email, inviteToken string) (*ent.User, bool, error) {
	u, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil {
		return u, false, nil
	}
	if !ent.IsNotFound(err) {
		return nil, false, fmt.Errorf("find user by email: %w", err)
	}

	u, err = s.regMode.RegisterNewUserByEmail(ctx, email, inviteToken, s.org)
	if err != nil {
		return nil, false, err
	}
	return u, true, nil
}
