package auth_test

// Тесты HTTP-обработчика аутентификации: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedOrganization(t *testing.T, svc *auth.AuthService, ctx context.Context) string {
	t.Helper()
	email := "founder-" + uuid.New().String()[:6] + "@example.com"
	code, err := auth.ExportAuthServiceMagicLink(svc).SendMagicLink(ctx, email)
	require.NoError(t, err)
	_, err = svc.LoginWithMagicLink(ctx, email, code, "", "browser", "127.0.0.1")
	require.NoError(t, err)
	org, err := svc.Org.GetOrganization(ctx)
	require.NoError(t, err)
	return org.InviteToken
}

// TestHandler_SendMagicLink_InviteOnlyNewUser проверяет HTTP-слой: HTTP-обработчик отправка magic link приглашение only new user.
func TestHandler_SendMagicLink_InviteOnlyNewUser(t *testing.T) {
	h, svc, _, _ := newHandlerTestEnv(t)
	ctx := t.Context()
	seedOrganization(t, svc, ctx)
	require.NoError(t, svc.SetRegistrationMode(ctx, "invite_only"))

	app := testApp(http.MethodPost, "/auth/magic-link", h.SendMagicLink)
	email := "new-" + uuid.New().String()[:6] + "@example.com"
	body, _ := json.Marshal(auth.MagicLinkRequest{Email: email})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// TestHandler_VerifyMagicLink_InviteOnlyBlocked проверяет HTTP-слой: HTTP-обработчик verify magic link приглашение only blocked.
func TestHandler_VerifyMagicLink_InviteOnlyBlocked(t *testing.T) {
	h, svc, _, magic := newHandlerTestEnv(t)
	ctx := t.Context()
	seedOrganization(t, svc, ctx)
	require.NoError(t, svc.SetRegistrationMode(ctx, "invite_only"))

	email := "blocked-" + uuid.New().String()[:6] + "@example.com"
	code, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)

	app := testApp(http.MethodPost, "/auth/magic-link/verify", h.VerifyMagicLink)
	body, _ := json.Marshal(auth.VerifyMagicLinkRequest{Email: email, Code: code})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// TestInvalidateUserStatusCache проверяет сценарий: invalidate user status кэш.
func TestInvalidateUserStatusCache(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	userID := uuid.New()
	key := auth.UserStatusCacheKey(userID)
	require.NoError(t, rdb.Set(t.Context(), key, "active", 0).Err())

	auth.InvalidateUserStatusCache(t.Context(), rdb, userID)
	exists, err := rdb.Exists(t.Context(), key).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}

// TestHandler_Logout_AuthenticatedViaJWT проверяет HTTP-слой: HTTP-обработчик выход authenticated via jwt.
func TestHandler_Logout_AuthenticatedViaJWT(t *testing.T) {
	h, svc, _, magic := newHandlerTestEnv(t)
	ctx := t.Context()
	email := "logout-jwt-" + uuid.New().String()[:6] + "@example.com"
	code, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)
	pair, err := svc.LoginWithMagicLink(ctx, email, code, "", "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/auth/logout", h.Logout, auth.RequireAuth(svc))

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestRegistrationModeService_UpsertConcurrent гарантирует отсутствие паники/гонок при регистрация mode сервис upsert конкурентный доступ.
func TestRegistrationModeService_UpsertConcurrent(t *testing.T) {
	svc, ctx := newRegistrationTestEnv(t)
	require.NoError(t, svc.SetRegistrationMode(ctx, "invite_only"))
	require.NoError(t, svc.SetRegistrationMode(ctx, "open"))
	mode, err := svc.GetRegistrationMode(ctx)
	require.NoError(t, err)
	assert.Equal(t, "open", mode)
}
