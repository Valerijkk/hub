//go:build hubtest

// Package auth — test export hooks for backend/testframework/auth/ (not compiled in release builds).
package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

func ExportHandlerService(h *Handler) *AuthService {
	return h.service
}

func ExportMagicLinkVerifyMax() int { return magicLinkVerifyMax }

func ExportMagicLinkRateMax() int { return magicLinkRateMax }

func ExportMagicLinkTTL() time.Duration { return magicLinkTTL }

func ExportMagicLinkFailsPrefix() string { return magicLinkFailsPrefix }

func ExportMaskEmailMinimized(email string) string { return maskEmailMinimized(email) }

func ExportAuthServiceMagicLink(s *AuthService) *MagicLinkService {
	return s.magicLink
}

func ExportAuthServiceRedis(s *AuthService) *redis.Client {
	return s.redis
}

func ExportUserStatusCacheDeactivated() string { return userStatusCacheDeactivated }

func ExportUserStatusCacheTTL() time.Duration { return userStatusCacheTTL }

func ExportResendEmailClientAPIKey(c *ResendEmailClient) string { return c.apiKey }

func ExportResendEmailClientFromEmail(c *ResendEmailClient) string { return c.fromEmail }

func ExportResendEmailClientInjectHTTP(c *ResendEmailClient, hc *http.Client) {
	c.client = hc
}

func ExportHashRefreshToken(token string) string { return hashRefreshToken(token) }

func ExportRefreshRedisKey(plaintext string) string { return refreshRedisKey(plaintext) }

func ExportRefreshRedisKeyStored(storedHash string) string { return refreshRedisKeyStored(storedHash) }

// ExportRefreshPayload mirrors refreshPayload for external session tests.
type ExportRefreshPayload struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Device    string `json:"device"`
}

func ExportSessionStoreRefreshToken(s *SessionService, ctx context.Context, token string, p ExportRefreshPayload) error {
	return s.storeRefreshToken(ctx, token, refreshPayload(p))
}

func ExportSessionInvalidateUserSessions(s *SessionService, ctx context.Context, userID uuid.UUID) error {
	return s.invalidateUserSessions(ctx, userID)
}

func ExportSessionSetNATS(s *SessionService, nc *natsclient.Client) {
	s.nats = nc
}

func ExportSessionPublishSessionCreated(s *SessionService, userID, sessionID, device, ip string) {
	s.publishSessionCreated(userID, sessionID, device, ip)
}

func ExportSessionPublishSessionEnded(s *SessionService, userID, sessionID, event string) {
	s.publishSessionEnded(userID, sessionID, event)
}

func ExportGenerateCode() (string, error) { return generateCode() }

func ExportNewSessionServiceForJWTTest(jwtSecret string, accessTTL time.Duration) *SessionService {
	return &SessionService{jwtSecret: []byte(jwtSecret), accessTTL: accessTTL}
}

func ExportSessionGenerateAccessToken(s *SessionService, userID, sessionID uuid.UUID) (string, error) {
	return s.generateAccessToken(userID, sessionID)
}
