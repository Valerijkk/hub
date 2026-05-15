package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/session"
	"github.com/valerii/hub/backend/ent/user"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// SessionService отвечает за JWT, refresh-токены и жизненный цикл сессий (SC-14-1).
type SessionService struct {
	db         *ent.Client
	redis      *redis.Client
	nats       *natsclient.Client
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewSessionService создаёт SessionService с явно переданными зависимостями.
func NewSessionService(db *ent.Client, rdb *redis.Client, nc *natsclient.Client, jwtSecret string, accessTTL, refreshTTL time.Duration) *SessionService {
	return &SessionService{db: db, redis: rdb, nats: nc, jwtSecret: []byte(jwtSecret), accessTTL: accessTTL, refreshTTL: refreshTTL}
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

func refreshKey(token string) string { return "refresh:" + token }

// generateAccessToken генерирует JWT access-токен.
func (s *SessionService) generateAccessToken(userID, sessionID uuid.UUID) (string, error) {
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
	return s.redis.Set(ctx, refreshKey(token), data, s.refreshTTL).Err()
}

// getRefreshPayload извлекает данные refresh-токена из Redis.
func (s *SessionService) getRefreshPayload(ctx context.Context, token string) (*refreshPayload, error) {
	data, err := s.redis.Get(ctx, refreshKey(token)).Bytes()
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
	s.redis.Del(ctx, refreshKey(token))
}

// invalidateUserSessions удаляет все активные сессии пользователя (DB + Redis).
// ME-01: bulk DELETE + Redis pipeline вместо N+1.
func (s *SessionService) invalidateUserSessions(ctx context.Context, userID uuid.UUID) {
	tokens, err := s.db.Session.Query().
		Where(session.HasUserWith(user.ID(userID))).
		Select(session.FieldRefreshToken).
		Strings(ctx)
	if err != nil {
		slog.Error("invalidateUserSessions: query refresh tokens", "error", err, "user_id", userID)
		return
	}
	if _, err := s.db.Session.Delete().
		Where(session.HasUserWith(user.ID(userID))).
		Exec(ctx); err != nil {
		slog.Error("invalidateUserSessions: bulk delete sessions", "error", err, "user_id", userID)
	}
	if len(tokens) == 0 {
		return
	}
	pipe := s.redis.Pipeline()
	for _, t := range tokens {
		pipe.Del(ctx, refreshKey(t))
	}
	if _, err := pipe.Exec(ctx); err != nil {
		slog.Warn("invalidateUserSessions: redis pipeline del", "error", err, "user_id", userID, "tokens", len(tokens))
	}
}

// CreateSession генерирует JWT + refresh-токен, сохраняет сессию и публикует NATS-событие.
// Перед созданием все существующие сессии пользователя инвалидируются (single session policy).
func (s *SessionService) CreateSession(ctx context.Context, userID uuid.UUID, device, ip string) (*TokenPairResponse, error) {
	s.invalidateUserSessions(ctx, userID)
	sessionID := uuid.New()
	refreshToken := uuid.New().String()
	accessToken, err := s.generateAccessToken(userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("create session: generate access token: %w", err)
	}
	if err := s.storeRefreshToken(ctx, refreshToken, refreshPayload{
		UserID: userID.String(), SessionID: sessionID.String(), Device: device,
	}); err != nil {
		return nil, fmt.Errorf("create session: store refresh: %w", err)
	}
	if _, err = s.db.Session.Create().
		SetID(sessionID).SetRefreshToken(refreshToken).SetDevice(device).SetIPAddress(ip).
		SetExpiresAt(time.Now().Add(s.refreshTTL)).SetUserID(userID).Save(ctx); err != nil {
		s.deleteRefreshToken(ctx, refreshToken)
		return nil, fmt.Errorf("create session: persist: %w", err)
	}
	s.publishSessionCreated(userID.String(), sessionID.String(), device, ip)
	return &TokenPairResponse{AccessToken: accessToken, RefreshToken: refreshToken, ExpiresIn: int64(s.accessTTL.Seconds())}, nil
}

// publishSessionCreated публикует событие создания сессии в NATS (fire-and-forget).
func (s *SessionService) publishSessionCreated(userID, sessionID, device, ip string) {
	data, err := json.Marshal(SessionCreatedEvent{UserID: userID, SessionID: sessionID, Device: device, IPAddress: ip})
	if err != nil {
		slog.Error("marshal session created event", "error", err)
		return
	}
	if err := s.nats.Publish("auth.session.created", data); err != nil {
		slog.Error("publish session created event", "error", err)
	}
}

// RefreshToken валидирует refresh-токен, выполняет ротацию и возвращает новую пару токенов.
// ME-02: порядок (1) gen → (2) store new → (3) update session → (4) delete old, для устранения TOCTOU.
func (s *SessionService) RefreshToken(ctx context.Context, oldRefreshToken string) (*TokenPairResponse, error) {
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
	newRefresh := uuid.New().String()
	accessToken, err := s.generateAccessToken(userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("refresh token: generate access: %w", err)
	}
	if err := s.storeRefreshToken(ctx, newRefresh, *payload); err != nil {
		return nil, fmt.Errorf("refresh token: store new: %w", err)
	}
	if _, saveErr := s.db.Session.UpdateOneID(sessionID).
		SetRefreshToken(newRefresh).SetLastActiveAt(time.Now()).Save(ctx); saveErr != nil {
		return nil, fmt.Errorf("refresh token: update session: %w", saveErr)
	}
	s.deleteRefreshToken(ctx, oldRefreshToken)
	return &TokenPairResponse{AccessToken: accessToken, RefreshToken: newRefresh, ExpiresIn: int64(s.accessTTL.Seconds())}, nil
}

// Logout удаляет refresh-токен из Redis и сессию из базы данных.
func (s *SessionService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	sess, err := s.db.Session.Get(ctx, sessionID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("logout: find session: %w", err)
	}
	s.deleteRefreshToken(ctx, sess.RefreshToken)
	return s.db.Session.DeleteOneID(sessionID).Exec(ctx)
}

// ListSessions возвращает все сессии пользователя, отсортированные по дате создания.
func (s *SessionService) ListSessions(ctx context.Context, userID uuid.UUID) ([]*ent.Session, error) {
	return s.db.Session.Query().
		Where(session.HasUserWith(user.ID(userID))).
		Order(ent.Desc(session.FieldCreatedAt)).
		All(ctx)
}

// RevokeSession отзывает указанную сессию после проверки принадлежности пользователю.
func (s *SessionService) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	sess, err := s.db.Session.Query().
		Where(session.ID(sessionID), session.HasUserWith(user.ID(userID))).
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
