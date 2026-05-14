// Package auth реализует аутентификацию через Magic Link и Telegram, управление JWT-сессиями и профилями пользователей.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/session"
	"github.com/valerii/hub/backend/ent/systemsetting"
	"github.com/valerii/hub/backend/ent/user"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// AuthService содержит бизнес-логику аутентификации и управления пользователями.
type AuthService struct {
	db        *ent.Client
	redis     *redis.Client
	nats      *natsclient.Client
	jwtSecret []byte
	accessTTL time.Duration
	refreshTTL time.Duration
	magicLink *MagicLinkService
	botToken  string
}

// NewAuthService создаёт новый сервис аутентификации.
func NewAuthService(
	db *ent.Client,
	rdb *redis.Client,
	nats *natsclient.Client,
	jwtSecret string,
	accessTTL, refreshTTL time.Duration,
	magicLink *MagicLinkService,
	botToken string,
) *AuthService {
	return &AuthService{
		db:         db,
		redis:      rdb,
		nats:       nats,
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		magicLink:  magicLink,
		botToken:   botToken,
	}
}

// ---- JWT helpers ----

// JWTClaims — пользовательские claims, встроенные в access-токены.
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID    string `json:"uid"`
	SessionID string `json:"sid"`
}

// generateAccessToken генерирует JWT access-токен для указанного пользователя и сессии.
func (s *AuthService) generateAccessToken(userID, sessionID uuid.UUID) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "hub",
		},
		UserID:    userID.String(),
		SessionID: sessionID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ParseAccessToken валидирует и парсит строку JWT access-токена.
func (s *AuthService) ParseAccessToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse access token: %w", err)
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// ---- Refresh token Redis helpers ----

// refreshPayload — данные refresh-токена, хранящиеся в Redis.
type refreshPayload struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Device    string `json:"device"`
}

// refreshKey возвращает Redis-ключ для refresh-токена.
func refreshKey(token string) string {
	return "refresh:" + token
}

// storeRefreshToken сохраняет refresh-токен в Redis с TTL.
func (s *AuthService) storeRefreshToken(ctx context.Context, token string, payload refreshPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, refreshKey(token), data, s.refreshTTL).Err()
}

// getRefreshPayload извлекает данные refresh-токена из Redis.
func (s *AuthService) getRefreshPayload(ctx context.Context, token string) (*refreshPayload, error) {
	data, err := s.redis.Get(ctx, refreshKey(token)).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("refresh token not found")
	}
	if err != nil {
		return nil, err
	}
	var p refreshPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// deleteRefreshToken удаляет refresh-токен из Redis.
func (s *AuthService) deleteRefreshToken(ctx context.Context, token string) {
	s.redis.Del(ctx, refreshKey(token))
}

// ---- Core session operations ----

// invalidateUserSessions удаляет все активные сессии пользователя (DB + Redis).
// Используется для обеспечения единственной активной сессии на аккаунт.
//
// ME-01: Заменяем N+1 цикл одним bulk DELETE и Redis pipeline. Сначала
// собираем refresh-токены одним запросом, потом удаляем все строки сессий
// одним DELETE, и затем пайплайним DEL по всем ключам Redis.
func (s *AuthService) invalidateUserSessions(ctx context.Context, userID uuid.UUID) {
	// Шаг 1 — собрать refresh-токены, чтобы потом удалить их из Redis.
	tokens, err := s.db.Session.Query().
		Where(session.HasUserWith(user.ID(userID))).
		Select(session.FieldRefreshToken).
		Strings(ctx)
	if err != nil {
		slog.Error("invalidateUserSessions: query refresh tokens", "error", err, "user_id", userID)
		return
	}

	// Шаг 2 — один bulk DELETE по всем сессиям пользователя.
	if _, err := s.db.Session.Delete().
		Where(session.HasUserWith(user.ID(userID))).
		Exec(ctx); err != nil {
		slog.Error("invalidateUserSessions: bulk delete sessions", "error", err, "user_id", userID)
		// Продолжаем — постараемся удалить refresh-токены в Redis, чтобы
		// уменьшить дрейф состояния (JWT перестанет матчиться).
	}

	// Шаг 3 — пайплайним Redis DEL по всем токенам.
	if len(tokens) == 0 {
		return
	}
	pipe := s.redis.Pipeline()
	for _, t := range tokens {
		pipe.Del(ctx, refreshKey(t))
	}
	if _, err := pipe.Exec(ctx); err != nil {
		// Redis Pipeline.Exec возвращает первую ошибку, но остальные
		// команды всё равно отправлены. Логируем и идём дальше.
		slog.Warn("invalidateUserSessions: redis pipeline del", "error", err, "user_id", userID, "tokens", len(tokens))
	}
}

// CreateSession генерирует JWT + refresh-токен, сохраняет сессию в БД и публикует NATS-событие.
// Перед созданием новой сессии все существующие сессии пользователя инвалидируются
// (single active session per account).
func (s *AuthService) CreateSession(ctx context.Context, userID uuid.UUID, device, ip string) (*TokenPairResponse, error) {
	// Invalidate all existing sessions for this user (single session policy).
	s.invalidateUserSessions(ctx, userID)

	sessionID := uuid.New()
	refreshToken := uuid.New().String()

	accessToken, err := s.generateAccessToken(userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("create session: generate access token: %w", err)
	}

	// Store refresh token in Redis.
	if err := s.storeRefreshToken(ctx, refreshToken, refreshPayload{
		UserID:    userID.String(),
		SessionID: sessionID.String(),
		Device:    device,
	}); err != nil {
		return nil, fmt.Errorf("create session: store refresh: %w", err)
	}

	// Persist session in Ent DB.
	_, err = s.db.Session.Create().
		SetID(sessionID).
		SetRefreshToken(refreshToken).
		SetDevice(device).
		SetIPAddress(ip).
		SetExpiresAt(time.Now().Add(s.refreshTTL)).
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		s.deleteRefreshToken(ctx, refreshToken)
		return nil, fmt.Errorf("create session: persist: %w", err)
	}

	// Publish NATS event (fire and forget).
	s.publishSessionCreated(userID.String(), sessionID.String(), device, ip)

	return &TokenPairResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

// publishSessionCreated публикует событие создания сессии в NATS (fire-and-forget).
func (s *AuthService) publishSessionCreated(userID, sessionID, device, ip string) {
	evt := SessionCreatedEvent{
		UserID:    userID,
		SessionID: sessionID,
		Device:    device,
		IPAddress: ip,
	}
	data, err := json.Marshal(evt)
	if err != nil {
		slog.Error("marshal session created event", "error", err)
		return
	}
	if err := s.nats.Publish("auth.session.created", data); err != nil {
		slog.Error("publish session created event", "error", err)
	}
}

// RefreshToken валидирует refresh-токен, выполняет ротацию и возвращает новую пару токенов.
func (s *AuthService) RefreshToken(ctx context.Context, oldRefreshToken string) (*TokenPairResponse, error) {
	payload, err := s.getRefreshPayload(ctx, oldRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return nil, fmt.Errorf("refresh token: invalid user id: %w", err)
	}
	sessionID, err := uuid.Parse(payload.SessionID)
	if err != nil {
		return nil, fmt.Errorf("refresh token: invalid session id: %w", err)
	}

	// Delete old refresh token (rotation).
	s.deleteRefreshToken(ctx, oldRefreshToken)

	// Generate new token pair.
	newRefresh := uuid.New().String()
	accessToken, err := s.generateAccessToken(userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("refresh token: generate access: %w", err)
	}
	if err := s.storeRefreshToken(ctx, newRefresh, *payload); err != nil {
		return nil, fmt.Errorf("refresh token: store new: %w", err)
	}

	// Update session in Ent.
	if _, saveErr := s.db.Session.UpdateOneID(sessionID).
		SetRefreshToken(newRefresh).
		SetLastActiveAt(time.Now()).
		Save(ctx); saveErr != nil {
		slog.Error("failed to update session after token refresh", "error", saveErr, "session_id", sessionID)
	}

	return &TokenPairResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefresh,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

// Logout удаляет refresh-токен из Redis и сессию из базы данных.
func (s *AuthService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	sess, err := s.db.Session.Get(ctx, sessionID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // already logged out
		}
		return fmt.Errorf("logout: find session: %w", err)
	}
	s.deleteRefreshToken(ctx, sess.RefreshToken)
	return s.db.Session.DeleteOneID(sessionID).Exec(ctx)
}

// ListSessions возвращает все сессии пользователя, отсортированные по дате создания.
func (s *AuthService) ListSessions(ctx context.Context, userID uuid.UUID) ([]*ent.Session, error) {
	return s.db.Session.Query().
		Where(session.HasUserWith(user.ID(userID))).
		Order(ent.Desc(session.FieldCreatedAt)).
		All(ctx)
}

// RevokeSession отзывает указанную сессию после проверки принадлежности пользователю.
func (s *AuthService) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	sess, err := s.db.Session.Query().
		Where(
			session.ID(sessionID),
			session.HasUserWith(user.ID(userID)),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("session not found or not owned by user")
		}
		return fmt.Errorf("revoke session: %w", err)
	}
	s.deleteRefreshToken(ctx, sess.RefreshToken)
	return s.db.Session.DeleteOneID(sessionID).Exec(ctx)
}

// ---- Login flows ----

// LoginWithMagicLink проверяет код, находит или создаёт пользователя и возвращает пару токенов.
func (s *AuthService) LoginWithMagicLink(ctx context.Context, email, code, device, ip string) (*TokenPairResponse, error) {
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

	return s.CreateSession(ctx, u.ID, device, ip)
}

// LoginWithTelegram проверяет initData Telegram, находит или создаёт пользователя, возвращает пару токенов.
func (s *AuthService) LoginWithTelegram(ctx context.Context, initData, device, ip string) (*TokenPairResponse, error) {
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

	return s.CreateSession(ctx, u.ID, device, ip)
}

// LoginWithTelegramUser находит или создаёт пользователя по TelegramUser и возвращает пару токенов.
func (s *AuthService) LoginWithTelegramUser(ctx context.Context, tgUser *TelegramUser, device, ip string) (*TokenPairResponse, error) {
	telegramID := strconv.FormatInt(tgUser.ID, 10)
	u, isNew, err := s.findOrCreateUserByTelegram(ctx, telegramID, tgUser)
	if err != nil {
		return nil, err
	}
	if isNew {
		slog.Info("new user registered via telegram bot code", "user_id", u.ID.String(), "telegram_id", telegramID)
	}
	return s.CreateSession(ctx, u.ID, device, ip)
}

// findOrCreateUserByEmail ищет пользователя по email. Если не найден — создаёт нового.
// Самый первый пользователь в системе получает роль admin.
func (s *AuthService) findOrCreateUserByEmail(ctx context.Context, email string) (*ent.User, bool, error) {
	u, err := s.db.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err == nil {
		return u, false, nil
	}
	if !ent.IsNotFound(err) {
		return nil, false, fmt.Errorf("find user by email: %w", err)
	}

	// Check registration mode.
	if err := s.checkRegistrationAllowed(ctx); err != nil {
		return nil, false, err
	}

	role := user.SystemRoleMember
	if s.isFirstUser(ctx) {
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
func (s *AuthService) findOrCreateUserByTelegram(ctx context.Context, telegramID string, tg *TelegramUser) (*ent.User, bool, error) {
	u, err := s.db.User.Query().Where(user.TelegramIDEQ(telegramID)).Only(ctx)
	if err == nil {
		return u, false, nil
	}
	if !ent.IsNotFound(err) {
		return nil, false, fmt.Errorf("find user by telegram: %w", err)
	}

	if err := s.checkRegistrationAllowed(ctx); err != nil {
		return nil, false, err
	}

	role := user.SystemRoleMember
	if s.isFirstUser(ctx) {
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

// isFirstUser проверяет, является ли текущая регистрация первой (для назначения admin).
func (s *AuthService) isFirstUser(ctx context.Context) bool {
	count, err := s.db.User.Query().Count(ctx)
	if err != nil {
		return false
	}
	return count == 0
}

// checkRegistrationAllowed проверяет, разрешена ли регистрация в текущем режиме.
func (s *AuthService) checkRegistrationAllowed(ctx context.Context) error {
	mode, _ := s.GetRegistrationMode(ctx)
	if mode == "invite_only" {
		return ErrRegistrationInviteOnly
	}
	return nil
}

// ---- Profile ----

// UpdateProfile обновляет поля профиля пользователя. system_role никогда не изменяется.
func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*ent.User, error) {
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
func (s *AuthService) CompleteOnboarding(ctx context.Context, userID uuid.UUID, req OnboardingCompleteRequest) (*ent.User, error) {
	if _, err := s.UpdateProfile(ctx, userID, req.ProfileData); err != nil {
		return nil, err
	}
	return s.db.User.UpdateOneID(userID).
		SetOnboardingCompleted(true).
		Save(ctx)
}

// GetMe возвращает пользователя по ID.
func (s *AuthService) GetMe(ctx context.Context, userID uuid.UUID) (*ent.User, error) {
	return s.db.User.Get(ctx, userID)
}

// ---- Registration mode ----

// GetRegistrationMode читает системную настройку "registration_mode". По умолчанию "open".
func (s *AuthService) GetRegistrationMode(ctx context.Context) (string, error) {
	setting, err := s.db.SystemSetting.Query().
		Where(systemsetting.KeyEQ("registration_mode")).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "open", nil
		}
		return "", fmt.Errorf("get registration mode: %w", err)
	}
	return setting.Value, nil
}

// SetRegistrationMode обновляет режим регистрации (только для администратора).
func (s *AuthService) SetRegistrationMode(ctx context.Context, mode string) error {
	// Upsert: try update first, create if not found.
	n, err := s.db.SystemSetting.Update().
		Where(systemsetting.KeyEQ("registration_mode")).
		SetValue(mode).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("set registration mode: %w", err)
	}
	if n == 0 {
		_, err = s.db.SystemSetting.Create().
			SetKey("registration_mode").
			SetValue(mode).
			SetDescription("Controls whether new users can register: open or invite_only").
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create registration mode setting: %w", err)
		}
	}
	return nil
}

// ---- UserResponse builder ----

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
		OnboardingCompleted: u.OnboardingCompleted,
		CustomFields:        u.CustomFields,
	}
	if u.TelegramID != nil {
		resp.TelegramID = *u.TelegramID
	}
	return resp
}

// sentinel errors
var (
	ErrRegistrationInviteOnly = fmt.Errorf("registration is by invite only")
)
