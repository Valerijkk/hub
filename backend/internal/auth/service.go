// Package auth реализует аутентификацию через Magic Link и Telegram, управление JWT-сессиями и профилями пользователей.
package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/valerii/hub/backend/ent"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// AuthService — тонкий orchestrator, который объединяет субсервисы auth.
type AuthService struct {
	Sessions  *SessionService
	Login     *LoginService
	Profile   *ProfileService
	RegMode   *RegistrationModeService
	Org       *OrganizationService

	repo      Repository
	redis     *redis.Client
	magicLink *MagicLinkService
}

// NewAuthService создаёт оркестратор и все субсервисы.
func NewAuthService(
	repo Repository,
	rdb *redis.Client,
	nats *natsclient.Client,
	jwtSecret string,
	accessTTL, refreshTTL time.Duration,
	magicLink *MagicLinkService,
	botToken string,
	sqlDB *sql.DB,
) *AuthService {
	sessions := NewSessionService(repo, rdb, nats, jwtSecret, accessTTL, refreshTTL)
	regMode := NewRegistrationModeService(repo, sqlDB)
	org := NewOrganizationService(repo)
	login := NewLoginService(repo, magicLink, botToken, sessions, regMode, org)
	profile := NewProfileService(repo)
	return &AuthService{
		Sessions:  sessions,
		Login:     login,
		Profile:   profile,
		RegMode:   regMode,
		Org:       org,
		repo:      repo,
		redis:     rdb,
		magicLink: magicLink,
	}
}

// ---- Делегирующие методы (back-compat для handler.go) ----

// ParseAccessToken — back-compat для middleware/ws.
func (s *AuthService) ParseAccessToken(t string) (*JWTClaims, error) {
	return s.Sessions.ParseAccessToken(t)
}

// IsUserDeactivated проверяет, деактивирован ли аккаунт (ME-13).
// Используется RequireAuth и WS handler — та же логика Redis-кэш + БД.
func (s *AuthService) IsUserDeactivated(ctx context.Context, uid uuid.UUID) (bool, error) {
	return userIsDeactivated(ctx, s.repo, s.redis, uid)
}

// CreateSession — back-compat для dev.Handler и тестов.
func (s *AuthService) CreateSession(ctx context.Context, uid uuid.UUID, dev, ip string) (*TokenPairResponse, error) {
	return s.Sessions.CreateSession(ctx, uid, dev, ip)
}

// RefreshToken — back-compat для handler.RefreshToken.
func (s *AuthService) RefreshToken(ctx context.Context, t string) (*TokenPairResponse, error) {
	return s.Sessions.RefreshToken(ctx, t)
}

// Logout — back-compat для handler.Logout.
func (s *AuthService) Logout(ctx context.Context, sid uuid.UUID) error {
	return s.Sessions.Logout(ctx, sid)
}

// ListSessions — back-compat для handler.ListSessions.
func (s *AuthService) ListSessions(ctx context.Context, uid uuid.UUID) ([]*ent.Session, error) {
	return s.Sessions.ListSessions(ctx, uid)
}

// RevokeSession — back-compat для handler.RevokeSession.
func (s *AuthService) RevokeSession(ctx context.Context, uid, sid uuid.UUID) error {
	return s.Sessions.RevokeSession(ctx, uid, sid)
}

// LoginWithMagicLink — back-compat для handler.VerifyMagicLink.
func (s *AuthService) LoginWithMagicLink(ctx context.Context, email, code, inviteToken, dev, ip string) (*TokenPairResponse, error) {
	return s.Login.LoginWithMagicLink(ctx, email, code, inviteToken, dev, ip)
}

// UserExistsByEmail — back-compat для handler.SendMagicLink (ARCH-H1).
// Позволяет handler-у проверить существование пользователя без прямого доступа к *ent.Client.
func (s *AuthService) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	return s.Login.UserExistsByEmail(ctx, email)
}

// LoginWithTelegram — back-compat для handler.LoginTelegram.
func (s *AuthService) LoginWithTelegram(ctx context.Context, initData, dev, ip string) (*TokenPairResponse, error) {
	return s.Login.LoginWithTelegram(ctx, initData, dev, ip)
}

// LoginWithTelegramUser — back-compat для handler.LoginTelegramCode.
func (s *AuthService) LoginWithTelegramUser(ctx context.Context, tg *TelegramUser, dev, ip string) (*TokenPairResponse, error) {
	return s.Login.LoginWithTelegramUser(ctx, tg, dev, ip)
}

// GetMe — back-compat для handler.GetMe.
func (s *AuthService) GetMe(ctx context.Context, uid uuid.UUID) (*ent.User, error) {
	return s.Profile.GetMe(ctx, uid)
}

// UpdateProfile — back-compat для handler.UpdateProfile.
func (s *AuthService) UpdateProfile(ctx context.Context, uid uuid.UUID, req UpdateProfileRequest) (*ent.User, error) {
	return s.Profile.UpdateProfile(ctx, uid, req)
}

// CompleteOnboarding — back-compat для handler.CompleteOnboarding.
func (s *AuthService) CompleteOnboarding(ctx context.Context, uid uuid.UUID, req OnboardingCompleteRequest) (*ent.User, error) {
	return s.Profile.CompleteOnboarding(ctx, uid, req)
}

// GetRegistrationMode — back-compat для handler.GetRegistrationMode и handler.SendMagicLink.
func (s *AuthService) GetRegistrationMode(ctx context.Context) (string, error) {
	return s.RegMode.GetRegistrationMode(ctx)
}

// SetRegistrationMode — back-compat для handler.SetRegistrationMode.
func (s *AuthService) SetRegistrationMode(ctx context.Context, mode string) error {
	return s.RegMode.SetRegistrationMode(ctx, mode)
}
