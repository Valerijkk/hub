package auth_test

// Тесты HTTP-обработчика аутентификации: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
	"bytes"
	"encoding/json"
	"io"
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
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
)

func newHandlerTestEnv(t *testing.T) (*auth.Handler, *auth.AuthService, *ent.Client, *auth.MagicLinkService) {
	t.Helper()
	dsn := "file:handlertest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	magic := auth.NewMagicLinkService(rdb, nil, "test@example.com")
	svc := auth.NewAuthService(auth.NewEntRepository(client), rdb, nil, "handler-secret", 15*time.Minute, 30*24*time.Hour, magic, nil)
	return auth.NewHandler(svc), svc, client, magic
}

// TestHandler_RefreshToken_Success проверяет HTTP-слой: HTTP-обработчик refresh token успех.
func TestHandler_RefreshToken_Success(t *testing.T) {
	h, svc, client, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	u, err := client.User.Create().SetEmail("refresh-h-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/auth/refresh", h.RefreshToken)

	body, _ := json.Marshal(auth.RefreshRequest{RefreshToken: pair.RefreshToken})
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_VerifyMagicLink_Success проверяет HTTP-слой: HTTP-обработчик verify magic link успех.
func TestHandler_VerifyMagicLink_Success(t *testing.T) {
	h, _, _, magic := newHandlerTestEnv(t)
	ctx := t.Context()
	email := "verify-h-" + uuid.New().String()[:6] + "@example.com"

	code, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/auth/magic-link/verify", h.VerifyMagicLink)

	body, _ := json.Marshal(auth.VerifyMagicLinkRequest{Email: email, Code: code})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_GetMe_Authenticated проверяет HTTP-слой: HTTP-обработчик получение me authenticated.
func TestHandler_GetMe_Authenticated(t *testing.T) {
	h, svc, client, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	u, err := client.User.Create().SetEmail("me-h-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	terminal := h.GetMe
	app.Get("/auth/me", terminal, auth.RequireAuth(svc))

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(respBody), u.Email)
}


// TestHandler_SendMagicLink_DevMode проверяет HTTP-слой: HTTP-обработчик отправка magic link dev mode.
func TestHandler_SendMagicLink_DevMode(t *testing.T) {
	h, _, _, _ := newHandlerTestEnv(t)
	app := fiber.New()
	app.Post("/auth/magic-link", h.SendMagicLink)

	body, _ := json.Marshal(auth.MagicLinkRequest{Email: "send-h-" + uuid.New().String()[:6] + "@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Logout_Authenticated проверяет HTTP-слой: HTTP-обработчик выход authenticated.
func TestHandler_Logout_Authenticated(t *testing.T) {
	h, svc, client, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	u, err := client.User.Create().SetEmail("logout-h-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	terminal := h.Logout
	app.Post("/auth/logout", terminal, auth.RequireAuth(svc))

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ListSessions_Authenticated проверяет HTTP-слой: HTTP-обработчик список sessions authenticated.
func TestHandler_ListSessions_Authenticated(t *testing.T) {
	h, svc, client, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	u, err := client.User.Create().SetEmail("sess-h-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	terminal := h.ListSessions
	app.Get("/auth/sessions", terminal, auth.RequireAuth(svc))

	req := httptest.NewRequest(http.MethodGet, "/auth/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
