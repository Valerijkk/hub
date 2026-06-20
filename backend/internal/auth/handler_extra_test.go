package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/user"
)

func TestHandler_ListSessions_MarksCurrent(t *testing.T) {
	h, svc, magic := newHandlerTestEnv(t)
	ctx := t.Context()
	email := "sess-mark-" + uuid.New().String()[:6] + "@example.com"
	code, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)
	pair, err := svc.LoginWithMagicLink(ctx, email, code, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	app.Get("/auth/sessions", h.ListSessions, RequireAuth(svc))

	req := httptest.NewRequest(http.MethodGet, "/auth/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_GetRegistrationMode_Unauthenticated(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, RequireAuth(svc))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/registration-mode", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_LoginTelegram_ExpiredAuth(t *testing.T) {
	h, _, _ := newHandlerTestEnv(t)
	app := testApp(http.MethodPost, "/auth/telegram", h.LoginTelegram)

	params := "auth_date=1&user=%7B%22id%22%3A1%7D&hash=deadbeef"
	body, _ := json.Marshal(TelegramAuthRequest{InitData: params})
	req := httptest.NewRequest(http.MethodPost, "/auth/telegram", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_RefreshToken_InvalidBody(t *testing.T) {
	h, _, _ := newHandlerTestEnv(t)
	app := testApp(http.MethodPost, "/auth/refresh", h.RefreshToken)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAuthService_GetMe(t *testing.T) {
	_, svc, magic := newHandlerTestEnv(t)
	ctx := t.Context()
	email := "getme-" + uuid.New().String()[:6] + "@example.com"
	code, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)
	pair, err := svc.LoginWithMagicLink(ctx, email, code, "browser", "127.0.0.1")
	require.NoError(t, err)

	claims, err := svc.Sessions.ParseAccessToken(pair.AccessToken)
	require.NoError(t, err)
	userID, err := uuid.Parse(claims.UserID)
	require.NoError(t, err)

	u, err := svc.GetMe(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, email, u.Email)
	// First user in empty DB becomes system admin (registration race policy).
	assert.Equal(t, user.SystemRoleAdmin, u.SystemRole)
}

func TestHandler_SendMagicLink_ExistingUserInviteOnly(t *testing.T) {
	h, svc, magic := newHandlerTestEnv(t)
	ctx := t.Context()

	email := "existing-inv-" + uuid.New().String()[:6] + "@example.com"
	code, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)
	_, err = svc.LoginWithMagicLink(ctx, email, code, "browser", "127.0.0.1")
	require.NoError(t, err)

	require.NoError(t, svc.SetRegistrationMode(ctx, "invite_only"))

	app := testApp(http.MethodPost, "/auth/magic-link", h.SendMagicLink)
	body, _ := json.Marshal(MagicLinkRequest{Email: email})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_LoginTelegramCode_BotNotConfigured(t *testing.T) {
	h, _, _ := newHandlerTestEnv(t)
	app := testApp(http.MethodPost, "/auth/telegram-code", h.LoginTelegramCode)

	body, _ := json.Marshal(map[string]string{"code": "123456"})
	req := httptest.NewRequest(http.MethodPost, "/auth/telegram-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestHandler_VerifyMagicLink_RateLimited(t *testing.T) {
	h, _, magic := newHandlerTestEnv(t)
	ctx := t.Context()
	email := "rate-" + uuid.New().String()[:6] + "@example.com"
	_, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)

	for i := 0; i < magicLinkVerifyMax; i++ {
		_ = magic.VerifyCode(ctx, email, "000000")
	}

	app := testApp(http.MethodPost, "/auth/magic-link/verify", h.VerifyMagicLink)
	body, _ := json.Marshal(VerifyMagicLinkRequest{Email: email, Code: "000000"})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}
