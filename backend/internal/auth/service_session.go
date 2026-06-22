package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/valerii/hub/backend/ent"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"github.com/valerii/hub/backend/internal/platform/observability"
)

// SessionService отвечает за JWT, refresh-токены и жизненный цикл сессий (SC-14-1).
type SessionService struct {
	repo       Repository
	redis      *redis.Client
	nats       *natsclient.Client
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewSessionService создаёт SessionService с явно переданными зависимостями.
func NewSessionService(repo Repository, rdb *redis.Client, nc *natsclient.Client, jwtSecret string, accessTTL, refreshTTL time.Duration) *SessionService {
	return &SessionService{repo: repo, redis: rdb, nats: nc, jwtSecret: []byte(jwtSecret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

// JWTClaims — пользовательские claims, встроенные в access-токены.
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID    string `json:"uid"`
	SessionID string `json:"sid"`
}

// refreshPayload — данные refresh-токена, хранящиеся в Redis.
type refreshPayload struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Device    string `json:"device"`
}

func refreshRedisKey(plaintext string) string {
	return "refresh:" + hashRefreshToken(plaintext)
}

func refreshRedisKeyStored(storedHash string) string {
	return "refresh:" + storedHash
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// generateAccessToken генерирует JWT access-токен.
func (s *SessionService) generateAccessToken(userID, sessionID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "hub",
		},
		UserID:    userID.String(),
		SessionID: sessionID.String(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}

// ParseAccessToken валидирует и парсит строку JWT access-токена.
func (s *SessionService) ParseAccessToken(tokenStr string) (*JWTClaims, error) {
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

// storeRefreshToken сохраняет refresh-токен в Redis с TTL.
func (s *SessionService) storeRefreshToken(ctx context.Context, token string, payload refreshPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, refreshRedisKey(token), data, s.refreshTTL).Err()
}

// getRefreshPayload извлекает данные refresh-токена из Redis.
func (s *SessionService) getRefreshPayload(ctx context.Context, token string) (*refreshPayload, error) {
	data, err := s.redis.Get(ctx, refreshRedisKey(token)).Bytes()
	if errors.Is(err, redis.Nil) {
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
func (s *SessionService) deleteRefreshToken(ctx context.Context, token string) {
	s.redis.Del(ctx, refreshRedisKey(token))
}

// invalidateUserSessions удаляет все активные сессии пользователя (DB + Redis).
// ME-01: bulk DELETE + Redis pipeline вместо N+1.
// Возвращает ошибку, чтобы single-session policy не нарушалась на сбое БД:
// если предыдущие сессии не удалены, новая создаваться не должна.
// Redis-сбой при удалении refresh-токенов трактуется как warning — TTL всё равно
// сработает, а сама запись о сессии в БД уже отсутствует, так что доступ закрыт.
func (s *SessionService) invalidateUserSessions(ctx context.Context, userID uuid.UUID) error {
	tokens, err := s.repo.ListUserSessionRefreshTokens(ctx, userID)
	if err != nil {
		slog.Error("invalidateUserSessions: query refresh tokens", "error", err, "user_id", userID)
		return fmt.Errorf("invalidate user sessions: query: %w", err)
	}
	if err := s.repo.DeleteUserSessions(ctx, userID); err != nil {
		slog.Error("invalidateUserSessions: bulk delete sessions", "error", err, "user_id", userID)
		return fmt.Errorf("invalidate user sessions: delete: %w", err)
	}
	if len(tokens) == 0 {
		return nil
	}
	pipe := s.redis.Pipeline()
	for _, t := range tokens {
		pipe.Del(ctx, refreshRedisKeyStored(t))
	}
	if _, err := pipe.Exec(ctx); err != nil {
		// Не блокируем создание новой сессии: запись в БД уже удалена,
		// refresh-токены протухнут по TTL.
		slog.Warn("invalidateUserSessions: redis pipeline del", "error", err, "user_id", userID, "tokens", len(tokens))
	}
	return nil
}

// CreateSession генерирует JWT + refresh-токен, сохраняет сессию и публикует NATS-событие.
// Перед созданием все существующие сессии пользователя инвалидируются (single session policy).
func (s *SessionService) CreateSession(ctx context.Context, userID uuid.UUID, device, ip string) (*TokenPairResponse, error) {
	if err := s.invalidateUserSessions(ctx, userID); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	sessionID := uuid.New()
	refreshToken := uuid.New().String()
	accessToken, err := s.generateAccessToken(userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("create session: generate access token: %w", err)
	}
	tokenHash := hashRefreshToken(refreshToken)
	if err := s.storeRefreshToken(ctx, refreshToken, refreshPayload{
		UserID: userID.String(), SessionID: sessionID.String(), Device: device,
	}); err != nil {
		return nil, fmt.Errorf("create session: store refresh: %w", err)
	}
	if _, err = s.repo.CreateSessionRecord(ctx, SessionRecordInput{
		ID: sessionID, RefreshTokenHash: tokenHash, Device: device, IPAddress: ip,
		ExpiresAt: time.Now().UTC().Add(s.refreshTTL), UserID: userID,
	}); err != nil {
		s.deleteRefreshToken(ctx, refreshToken)
		return nil, fmt.Errorf("create session: persist: %w", err)
	}
	s.publishSessionCreated(userID.String(), sessionID.String(), device, ip)
	return &TokenPairResponse{AccessToken: accessToken, RefreshToken: refreshToken, ExpiresIn: int64(s.accessTTL.Seconds())}, nil
}

// publishSessionCreated публикует событие создания сессии в NATS (fire-and-forget).
func (s *SessionService) publishSessionCreated(userID, sessionID, device, ip string) {
	if s.nats == nil {
		return
	}
	data, err := json.Marshal(SessionCreatedEvent{UserID: userID, SessionID: sessionID, Device: device, IPAddress: ip})
	if err != nil {
		slog.Error("marshal session created event", "error", err)
		return
	}
	if err := s.nats.Publish("auth.session.created", data); err != nil {
		slog.Error("publish session created event", "error", err)
	}
}

// validateSessionForRefresh проверяет, что сессия в БД активна, hash совпадает и пользователь не деактивирован.
func (s *SessionService) validateSessionForRefresh(ctx context.Context, userID, sessionID uuid.UUID, oldRefreshToken string) error {
	sess, err := s.repo.GetUserSession(ctx, userID, sessionID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("refresh token: session not found")
		}
		return fmt.Errorf("refresh token: get session: %w", err)
	}
	if time.Now().UTC().After(sess.ExpiresAt) {
		return fmt.Errorf("refresh token: session expired")
	}
	oldHash := hashRefreshToken(oldRefreshToken)
	if subtle.ConstantTimeCompare([]byte(sess.RefreshToken), []byte(oldHash)) != 1 {
		return fmt.Errorf("refresh token: hash mismatch")
	}
	deactivated, err := userIsDeactivated(ctx, s.repo, s.redis, userID)
	if err != nil {
		return fmt.Errorf("refresh token: verify user status: %w", err)
	}
	if deactivated {
		return fmt.Errorf("refresh token: user deactivated")
	}
	return nil
}

// RefreshToken валидирует refresh-токен, выполняет ротацию и возвращает новую пару токенов.
// ME-02: порядок (1) gen → (2) store new → (3) update session → (4) delete old, для устранения TOCTOU.
func (s *SessionService) RefreshToken(ctx context.Context, oldRefreshToken string) (*TokenPairResponse, error) {
	return observability.Timed(ctx, "auth.Refresh", func(ctx context.Context) (*TokenPairResponse, error) {
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
		if err := s.validateSessionForRefresh(ctx, userID, sessionID, oldRefreshToken); err != nil {
			return nil, err
		}
		newRefresh := uuid.New().String()
		newHash := hashRefreshToken(newRefresh)
		accessToken, err := s.generateAccessToken(userID, sessionID)
		if err != nil {
			return nil, fmt.Errorf("refresh token: generate access: %w", err)
		}
		if err := s.storeRefreshToken(ctx, newRefresh, *payload); err != nil {
			return nil, fmt.Errorf("refresh token: store new: %w", err)
		}
		if err := s.repo.UpdateSessionRefreshHash(ctx, sessionID, newHash, time.Now().UTC()); err != nil {
			return nil, fmt.Errorf("refresh token: update session: %w", err)
		}
		s.deleteRefreshToken(ctx, oldRefreshToken)
		return &TokenPairResponse{AccessToken: accessToken, RefreshToken: newRefresh, ExpiresIn: int64(s.accessTTL.Seconds())}, nil
	})
}

// Logout удаляет refresh-токен из Redis и сессию из базы данных.
func (s *SessionService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	sess, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("logout: find session: %w", err)
	}
	s.redis.Del(ctx, refreshRedisKeyStored(sess.RefreshToken))
	return s.repo.DeleteSession(ctx, sessionID)
}

// ListSessions возвращает все сессии пользователя, отсортированные по дате создания.
func (s *SessionService) ListSessions(ctx context.Context, userID uuid.UUID) ([]*ent.Session, error) {
	return s.repo.ListUserSessions(ctx, userID)
}

// RevokeSession отзывает указанную сессию после проверки принадлежности пользователю.
func (s *SessionService) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	sess, err := s.repo.GetUserSession(ctx, userID, sessionID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("session not found or not owned by user")
		}
		return fmt.Errorf("revoke session: %w", err)
	}
	s.redis.Del(ctx, refreshRedisKeyStored(sess.RefreshToken))
	return s.repo.DeleteSession(ctx, sessionID)
}
