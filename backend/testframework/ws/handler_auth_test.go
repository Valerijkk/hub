package ws_test

// Тесты HTTP-обработчика WebSocket: интеграционные и unit-сценарии testframework.
import (
	"context"
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
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/auth"
	"github.com/valerii/hub/backend/internal/ws"
)

func newWSAuthService(t *testing.T) (*auth.AuthService, *auth.MagicLinkService, *ent.Client) {
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
	svc := auth.NewAuthService(auth.NewEntRepository(client), rdb, nil, "ws-test-secret", 15*time.Minute, 30*24*time.Hour, magic, nil)
	return svc, magic, client
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

func wsIssueTicket(t *testing.T, svc *auth.AuthService, userID, sessionID string) string {
	t.Helper()
	ticket, _, err := svc.IssueWSTicket(t.Context(), userID, sessionID)
	require.NoError(t, err)
	return ticket
}

func wsTicketForAccess(t *testing.T, svc *auth.AuthService, accessToken string) string {
	t.Helper()
	claims, err := svc.ParseAccessToken(accessToken)
	require.NoError(t, err)
	return wsIssueTicket(t, svc, claims.UserID, claims.SessionID)
}

func wsReqWithTicket(ticket string) *http.Request {
	req := httptest.NewRequest(fiber.MethodGet, "/ws", nil)
	req.Header.Set("Sec-WebSocket-Protocol", ws.WSTicketProtocolPrefix+ticket)
	return req
}

// TestWSHandler_MissingTicket_Unauthorized — нет subprotocol ticket → 401.
func TestWSHandler_MissingTicket_Unauthorized(t *testing.T) {
	hub := ws.NewHub()
	svc, _, _ := newWSAuthService(t)
	app := fiber.New()
	app.Get("/ws", ws.Handler(hub, svc, ws.ExportAllowAllAuthorizer()))

	req := httptest.NewRequest(fiber.MethodGet, "/ws", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestWSHandler_QueryToken_Denied — A2W3-4-001: ?token= всегда 401.
func TestWSHandler_QueryToken_Denied(t *testing.T) {
	hub := ws.NewHub()
	svc, magic, _ := newWSAuthService(t)
	token := wsAccessToken(t, svc, magic)

	app := fiber.New()
	app.Get("/ws", ws.Handler(hub, svc, ws.ExportAllowAllAuthorizer()))

	req := httptest.NewRequest(fiber.MethodGet, "/ws?token="+token, nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestWSHandler_InvalidTicket_Unauthorized — битый/чужой ticket → 401.
func TestWSHandler_InvalidTicket_Unauthorized(t *testing.T) {
	hub := ws.NewHub()
	svc, _, _ := newWSAuthService(t)
	app := fiber.New()
	app.Get("/ws", ws.Handler(hub, svc, ws.ExportAllowAllAuthorizer()))

	req := wsReqWithTicket("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestWSHandler_DeactivatedUser_Forbidden проверяет отклонение WS-апгрейда для деактивированного пользователя (ME-13).
func TestWSHandler_DeactivatedUser_Forbidden(t *testing.T) {
	hub := ws.NewHub()
	svc, _, client := newWSAuthService(t)
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("ws-deact-" + uuid.New().String()[:6] + "@example.com").
		SetStatus(user.StatusDeactivated).
		Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	claims, err := svc.ParseAccessToken(pair.AccessToken)
	require.NoError(t, err)
	ticket := wsIssueTicket(t, svc, claims.UserID, claims.SessionID)

	app := fiber.New()
	app.Get("/ws", ws.Handler(hub, svc, ws.ExportAllowAllAuthorizer()))

	req := wsReqWithTicket(ticket)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// TestWSHandler_TicketOneShot — повторный redeem того же ticket → 401.
func TestWSHandler_TicketOneShot(t *testing.T) {
	hub := ws.NewHub()
	svc, magic, _ := newWSAuthService(t)
	access := wsAccessToken(t, svc, magic)
	ticket := wsTicketForAccess(t, svc, access)

	app := fiber.New()
	app.Get("/ws", ws.Handler(hub, svc, ws.ExportAllowAllAuthorizer()))

	// Первый redeem съедает ticket на HTTP-слое (до upgrade fail без WS headers —
	// достаточно проверить, что после RedeemWSTicket повторный handler даёт 401).
	_, err := svc.RedeemWSTicket(t.Context(), ticket)
	require.NoError(t, err)

	req := wsReqWithTicket(ticket)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
