package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
)

// LoginService инкапсулирует Magic Link / Telegram login-флоу и создание/поиск
// пользователей. Делегирует выпуск токенов в SessionService, а проверку режима
// регистрации — в RegistrationModeService (SC-14-1).
type LoginService struct {
	db        *ent.Client
	magicLink *MagicLinkService
	botToken  string
	sessions  *SessionService
	regMode   *RegistrationModeService
}

// NewLoginService создаёт LoginService с явно переданными зависимостями.
func NewLoginService(
	db *ent.Client,
	magicLink *MagicLinkService,
	botToken string,
	sessions *SessionService,
	regMode *RegistrationModeService,
) *LoginService {
	return &LoginService{
		db:        db,
		magicLink: magicLink,
		botToken:  botToken,
		sessions:  sessions,
		regMode:   regMode,
	}
}

// LoginWithMagicLink проверяет код, находит или создаёт пользователя и возвращает пару токенов.
func (s *LoginService) LoginWithMagicLink(ctx context.Context, email, code, device, ip string) (*TokenPairResponse, error) {
	if err := s.magicLink.VerifyCode(ctx, email, code); err != nil {
		return nil, err
	}

	u, isNew, err := s.findOrCreateUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if isNew {
		slog.Info("new user registered via magic link", "user_id", u.ID.String(), "email", email)
	}

	return s.sessions.CreateSession(ctx, u.ID, device, ip)
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
	return s.sessions.CreateSession(ctx, u.ID, device, ip)
}

// findOrCreateUserByEmail ищет пользователя по email. Если не найден — создаёт нового.
// Самый первый пользователь в системе получает роль admin.
func (s *LoginService) findOrCreateUserByEmail(ctx context.Context, email string) (*ent.User, bool, error) {
	u, err := s.db.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err == nil {
		return u, false, nil
	}
	if !ent.IsNotFound(err) {
		return nil, false, fmt.Errorf("find user by email: %w", err)
	}

	if err := s.regMode.CheckRegistrationAllowed(ctx); err != nil {
		return nil, false, err
	}

	role := user.SystemRoleMember
	if s.regMode.IsFirstUser(ctx) {
		role = user.SystemRoleAdmin
	}

	u, err = s.db.User.Create().
		SetEmail(email).
		SetSystemRole(role).
		Save(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("create user: %w", err)
	}
	return u, true, nil
}

// findOrCreateUserByTelegram ищет пользователя по telegram_id. Если не найден — создаёт нового.
func (s *LoginService) findOrCreateUserByTelegram(ctx context.Context, telegramID string, tg *TelegramUser) (*ent.User, bool, error) {
	u, err := s.db.User.Query().Where(user.TelegramIDEQ(telegramID)).Only(ctx)
	if err == nil {
		return u, false, nil
	}
	if !ent.IsNotFound(err) {
		return nil, false, fmt.Errorf("find user by telegram: %w", err)
	}

	if err := s.regMode.CheckRegistrationAllowed(ctx); err != nil {
		return nil, false, err
	}

	role := user.SystemRoleMember
	if s.regMode.IsFirstUser(ctx) {
		role = user.SystemRoleAdmin
	}

	name := tg.FirstName
	if tg.LastName != "" {
		name += " " + tg.LastName
	}

	// Telegram users get a placeholder email (telegram_<id>@hub.local) until they
	// provide a real one during onboarding.
	placeholderEmail := fmt.Sprintf("telegram_%s@hub.local", telegramID)

	u, err = s.db.User.Create().
		SetEmail(placeholderEmail).
		SetNillableTelegramID(&telegramID).
		SetName(name).
		SetDisplayName(tg.FirstName).
		SetSystemRole(role).
		Save(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("create telegram user: %w", err)
	}
	return u, true, nil
}
