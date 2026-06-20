package auth

import (
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

	"github.com/valerii/hub/backend/ent/enttest"
)

func newHandlerTestEnv(t *testing.T) (*Handler, *AuthService, *MagicLinkService) {
	t.Helper()
	dsn := "file:handlertest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	magic := NewMagicLinkService(rdb, nil, "test@example.com")
	svc := NewAuthService(client, rdb, nil, "handler-secret", 15*time.Minute, 30*24*time.Hour, magic, "bot", nil)
	return &Handler{service: svc}, svc, magic
}

func TestHandler_RefreshToken_Success(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	u, err := svc.db.User.Create().SetEmail("refresh-h-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/auth/refresh", h.RefreshToken)

	body, _ := json.Marshal(RefreshRequest{RefreshToken: pair.RefreshToken})
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_VerifyMagicLink_Success(t *testing.T) {
	h, _, magic := newHandlerTestEnv(t)
	ctx := t.Context()
	email := "verify-h-" + uuid.New().String()[:6] + "@example.com"

	code, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/auth/magic-link/verify", h.VerifyMagicLink)

	body, _ := json.Marshal(VerifyMagicLinkRequest{Email: email, Code: code})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_GetMe_Authenticated(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	u, err := svc.db.User.Create().SetEmail("me-h-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	terminal := h.GetMe
	app.Get("/auth/me", terminal, RequireAuth(svc))

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(respBody), u.Email)
}


func TestHandler_SendMagicLink_DevMode(t *testing.T) {
	h, _, _ := newHandlerTestEnv(t)
	app := fiber.New()
	app.Post("/auth/magic-link", h.SendMagicLink)

	body, _ := json.Marshal(MagicLinkRequest{Email: "send-h-" + uuid.New().String()[:6] + "@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_Logout_Authenticated(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	u, err := svc.db.User.Create().SetEmail("logout-h-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	terminal := h.Logout
	app.Post("/auth/logout", terminal, RequireAuth(svc))

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_ListSessions_Authenticated(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	u, err := svc.db.User.Create().SetEmail("sess-h-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	terminal := h.ListSessions
	app.Get("/auth/sessions", terminal, RequireAuth(svc))

	req := httptest.NewRequest(http.MethodGet, "/auth/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
