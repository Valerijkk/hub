package auth_test

// Тесты HTTP-обработчика аутентификации: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
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

// TestHandler_ListSessions_MarksCurrent проверяет HTTP-слой: ровно одна сессия
// с is_current=true и её id совпадает с sid из JWT access token (A2W6-A1-004).
func TestHandler_ListSessions_MarksCurrent(t *testing.T) {
	h, svc, _, magic := newHandlerTestEnv(t)
	ctx := t.Context()
	email := "sess-mark-" + uuid.New().String()[:6] + "@example.com"
	code, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)
	pair, err := svc.LoginWithMagicLink(ctx, email, code, "", "browser", "127.0.0.1")
	require.NoError(t, err)

	claims, err := svc.Sessions.ParseAccessToken(pair.AccessToken)
	require.NoError(t, err)
	require.NotEmpty(t, claims.SessionID)

	app := fiber.New()
	app.Get("/auth/sessions", h.ListSessions, auth.RequireAuth(svc))

	req := httptest.NewRequest(http.MethodGet, "/auth/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var sessions []auth.SessionResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&sessions))
	require.NotEmpty(t, sessions)

	var currentCount int
	var currentID string
	for _, s := range sessions {
		if s.IsCurrent {
			currentCount++
			currentID = s.ID
		}
	}
	require.Equal(t, 1, currentCount, "ровно одна сессия должна быть is_current")
	assert.Equal(t, claims.SessionID, currentID)
}

// TestHandler_GetRegistrationMode_Unauthenticated проверяет HTTP-слой: HTTP-обработчик получение регистрация mode unauthenticated.
func TestHandler_GetRegistrationMode_Unauthenticated(t *testing.T) {
	h, svc, _, _ := newHandlerTestEnv(t)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, auth.RequireAuth(svc))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/registration-mode", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestHandler_RefreshToken_InvalidBody проверяет отклонение при HTTP-обработчик refresh token некорректный body.
func TestHandler_RefreshToken_InvalidBody(t *testing.T) {
	h, _, _, _ := newHandlerTestEnv(t)
	app := testApp(http.MethodPost, "/auth/refresh", h.RefreshToken)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestAuthService_GetMe проверяет бизнес-логику: auth сервис получение me.
func TestAuthService_GetMe(t *testing.T) {
	_, svc, _, magic := newHandlerTestEnv(t)
	ctx := t.Context()
	email := "getme-" + uuid.New().String()[:6] + "@example.com"
	code, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)
	pair, err := svc.LoginWithMagicLink(ctx, email, code, "", "browser", "127.0.0.1")
	require.NoError(t, err)

	claims, err := svc.Sessions.ParseAccessToken(pair.AccessToken)
	require.NoError(t, err)
	userID, err := uuid.Parse(claims.UserID)
	require.NoError(t, err)

	u, err := svc.GetMe(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, email, u.Email)
	// Первый пользователь в пустой БД становится system admin (политика гонки регистрации).
	assert.Equal(t, user.SystemRoleFounder, u.SystemRole)
}

// TestHandler_SendMagicLink_ExistingUserInviteOnly проверяет HTTP-слой: HTTP-обработчик отправка magic link existing user приглашение only.
func TestHandler_SendMagicLink_ExistingUserInviteOnly(t *testing.T) {
	h, svc, _, magic := newHandlerTestEnv(t)
	ctx := t.Context()

	email := "existing-inv-" + uuid.New().String()[:6] + "@example.com"
	code, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)
	_, err = svc.LoginWithMagicLink(ctx, email, code, "", "browser", "127.0.0.1")
	require.NoError(t, err)

	require.NoError(t, svc.SetRegistrationMode(ctx, "invite_only"))

	app := testApp(http.MethodPost, "/auth/magic-link", h.SendMagicLink)
	body, _ := json.Marshal(auth.MagicLinkRequest{Email: email})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_VerifyMagicLink_RateLimited проверяет HTTP-слой: HTTP-обработчик verify magic link rate limited.
func TestHandler_VerifyMagicLink_RateLimited(t *testing.T) {
	h, _, _, magic := newHandlerTestEnv(t)
	ctx := t.Context()
	email := "rate-" + uuid.New().String()[:6] + "@example.com"
	_, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)

	for i := 0; i < auth.ExportMagicLinkVerifyMax(); i++ {
		_ = magic.VerifyCode(ctx, email, "000000")
	}

	app := testApp(http.MethodPost, "/auth/magic-link/verify", h.VerifyMagicLink)
	body, _ := json.Marshal(auth.VerifyMagicLinkRequest{Email: email, Code: "000000"})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}
