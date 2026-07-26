package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/platform/observability"
)

// LoginService инкапсулирует Magic Link / Telegram login-флоу.
type LoginService struct {
	repo      Repository
	magicLink *MagicLinkService
	botToken  string
	sessions  *SessionService
	regMode   *RegistrationModeService
	org       *OrganizationService
}

// NewLoginService создаёт LoginService с явно переданными зависимостями.
func NewLoginService(
	repo Repository,
	magicLink *MagicLinkService,
	botToken string,
	sessions *SessionService,
	regMode *RegistrationModeService,
	org *OrganizationService,
) *LoginService {
	return &LoginService{
		repo:      repo,
		magicLink: magicLink,
		botToken:  botToken,
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
			slog.Info("new user registered via magic link", "user_id", u.ID.String(), "email", email, "status", u.Status)
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

// LoginWithTelegram проверяет initData Telegram, находит или создаёт пользователя, возвращает пару токенов.
func (s *LoginService) LoginWithTelegram(ctx context.Context, initData, device, ip string) (*TokenPairResponse, error) {
	tgUser, err := VerifyTelegramAuth(s.botToken, initData)
	if err != nil {
		return nil, err
	}

	telegramID := strconv.FormatInt(tgUser.ID, 10)
	u, isNew, err := s.findOrCreateUserByTelegram(ctx, telegramID, tgUser)
	if err != nil {
		return nil, err
	}

	if isNew {
		slog.Info("new user registered via telegram", "user_id", u.ID.String(), "telegram_id", telegramID)
	}

	if u.Status != user.StatusActive {
		return nil, ErrAccountNotActive
	}

	return s.sessions.CreateSession(ctx, u.ID, device, ip)
}

// LoginWithTelegramUser находит или создаёт пользователя по TelegramUser и возвращает пару токенов.
func (s *LoginService) LoginWithTelegramUser(ctx context.Context, tgUser *TelegramUser, device, ip string) (*TokenPairResponse, error) {
	telegramID := strconv.FormatInt(tgUser.ID, 10)
	u, isNew, err := s.findOrCreateUserByTelegram(ctx, telegramID, tgUser)
	if err != nil {
		return nil, err
	}
	if isNew {
		slog.Info("new user registered via telegram bot code", "user_id", u.ID.String(), "telegram_id", telegramID)
	}
	if u.Status != user.StatusActive {
		return nil, ErrAccountNotActive
	}
	return s.sessions.CreateSession(ctx, u.ID, device, ip)
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

func (s *LoginService) findOrCreateUserByTelegram(ctx context.Context, telegramID string, tg *TelegramUser) (*ent.User, bool, error) {
	u, err := s.repo.GetUserByTelegramID(ctx, telegramID)
	if err == nil {
		return u, false, nil
	}
	if !ent.IsNotFound(err) {
		return nil, false, fmt.Errorf("find user by telegram: %w", err)
	}

	if err := s.regMode.CheckRegistrationAllowed(ctx); err != nil {
		return nil, false, err
	}

	name := tg.FirstName
	if tg.LastName != "" {
		name += " " + tg.LastName
	}

	placeholderEmail := fmt.Sprintf("telegram_%s@hub.local", telegramID)
	tgID := telegramID

	u, err = s.regMode.CreateFirstOrMemberUser(ctx, func(b *ent.UserCreate, role user.SystemRole) *ent.UserCreate {
		return b.
			SetEmail(placeholderEmail).
			SetNillableTelegramID(&tgID).
			SetName(name).
			SetDisplayName(tg.FirstName).
			SetSystemRole(role)
	})
	if err != nil {
		return nil, false, err
	}
	return u, true, nil
}
