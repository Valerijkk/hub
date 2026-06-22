//go:build hubtest

// Package auth — test export hooks for backend/testframework/auth/ (not compiled in release builds).
package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mymmrac/telego"
	"github.com/redis/go-redis/v9"

	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

func ExportTelegramHandleLogin(t *TelegramBotAuth, ctx context.Context, msg *telego.Message) {
	t.handleLogin(ctx, msg)
}

func ExportHandlerService(h *Handler) *AuthService {
	return h.service
}

func ExportMagicLinkVerifyMax() int { return magicLinkVerifyMax }

func ExportMagicLinkRateMax() int { return magicLinkRateMax }

func ExportMagicLinkTTL() time.Duration { return magicLinkTTL }

func ExportMagicLinkFailsPrefix() string { return magicLinkFailsPrefix }

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

func ExportTGCodePrefix() string { return tgCodePrefix }

func ExportTGFailsPrefix() string { return tgFailsPrefix }

func ExportGenerateTGCode() (string, error) { return generateTGCode() }

func ExportNewTelegramBotAuthForTest(rdb *redis.Client) *TelegramBotAuth {
	return &TelegramBotAuth{redis: rdb}
}

func ExportNewTelegramBotAuthForTestWithBot(bot *telego.Bot, rdb *redis.Client) *TelegramBotAuth {
	return &TelegramBotAuth{bot: bot, redis: rdb}
}

func ExportTelegramRecordVerifyFailure(bot *TelegramBotAuth, ctx context.Context, key string) {
	bot.recordVerifyFailure(ctx, key)
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

func ExportGenerateCode() (string, error) { return generateCode() }

func ExportNewSessionServiceForJWTTest(jwtSecret string, accessTTL time.Duration) *SessionService {
	return &SessionService{jwtSecret: []byte(jwtSecret), accessTTL: accessTTL}
}

func ExportSessionGenerateAccessToken(s *SessionService, userID, sessionID uuid.UUID) (string, error) {
	return s.generateAccessToken(userID, sessionID)
}
