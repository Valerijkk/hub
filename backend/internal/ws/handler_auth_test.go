package ws

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/auth"
)

func newWSAuthService(t *testing.T) (*auth.AuthService, *auth.MagicLinkService) {
	t.Helper()
	dsn := "file:wstest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	magic := auth.NewMagicLinkService(rdb, nil, "test@example.com")
	svc := auth.NewAuthService(client, rdb, nil, "ws-test-secret", 15*time.Minute, 30*24*time.Hour, magic, "bot", nil)
	return svc, magic
}

func wsAccessToken(t *testing.T, svc *auth.AuthService, magic *auth.MagicLinkService) string {
	t.Helper()
	ctx := t.Context()
	email := "ws-" + uuid.New().String()[:6] + "@example.com"
	code, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)
	pair, err := svc.LoginWithMagicLink(ctx, email, code, "", "browser", "127.0.0.1")
	require.NoError(t, err)
	return pair.AccessToken
}

func TestWSHandler_MissingToken_Unauthorized(t *testing.T) {
	hub := NewHub()
	svc, _ := newWSAuthService(t)
	app := fiber.New()
	app.Get("/ws", Handler(hub, svc, allowAllAuthorizer{}))

	req := httptest.NewRequest(fiber.MethodGet, "/ws", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestWSHandler_InvalidToken_Unauthorized(t *testing.T) {
	hub := NewHub()
	svc, _ := newWSAuthService(t)
	app := fiber.New()
	app.Get("/ws", Handler(hub, svc, allowAllAuthorizer{}))

	req := httptest.NewRequest(fiber.MethodGet, "/ws?token=not-a-jwt", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
