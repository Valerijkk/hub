package ws_test

// C-W19-001 / TASK-177: Audit Deactivate kill-path end-to-end —
// SessionService.RevokeAllUserSessions → NATS revoked → Bridge → UnregisterBySession.
// Mutation: убрать publishSessionEnded в RevokeAllUserSessions → client остаётся в room (RED).

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/audit"
	"github.com/valerii/hub/backend/internal/auth"
	"github.com/valerii/hub/backend/internal/ws"
)

// TestDeactivateUser_LiveNATS_UnregistersWS_C_W19 —
// открытый WS жертвы рвётся после admin Deactivate (не только REST/upgrade deny).
func TestDeactivateUser_LiveNATS_UnregistersWS_C_W19(t *testing.T) {
	dsn := "file:c_w19_ws_kill_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	nc := startEmbeddedNATS(t)
	sessionSvc := auth.NewSessionService(auth.NewEntRepository(client), rdb, nc, "c-w19-jwt", 15*time.Minute, 30*24*time.Hour)

	auditRepo := audit.NewRepository(client)
	auditSvc := audit.NewAuditService(auditRepo, rdb)
	auditSvc.SetSessionRevoker(sessionSvc)

	ctx := context.Background()
	victim, err := client.User.Create().
		SetEmail("victim-c-w19-" + uuid.New().String()[:8] + "@example.com").
		SetStatus(user.StatusActive).
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	pair, err := sessionSvc.CreateSession(ctx, victim.ID, "browser", "127.0.0.1")
	require.NoError(t, err)
	claims, err := sessionSvc.ParseAccessToken(pair.AccessToken)
	require.NoError(t, err)

	hub, _ := startTestHub(t)
	room := "room-c-w19-deactivate"
	wsClient := ws.NewClient(victim.ID.String(), claims.SessionID)
	hub.Register(wsClient)
	hub.Subscribe(wsClient, room)
	require.Eventually(t, func() bool {
		return ws.ExportConnectionsForUser(hub, victim.ID.String()) == 1 && ws.ExportClientInRoom(wsClient, room)
	}, 500*time.Millisecond, 10*time.Millisecond)

	bridge := ws.NewBridge(hub, nc)
	require.NoError(t, bridge.Start())
	t.Cleanup(bridge.Stop)

	require.NoError(t, auditSvc.DeactivateUser(ctx, victim.ID))

	require.Eventually(t, func() bool {
		select {
		case <-ws.ExportClientDone(wsClient):
			return true
		default:
			return false
		}
	}, 2*time.Second, 10*time.Millisecond, "WS must close after Deactivate (mutation: remove publishSessionEnded)")

	require.False(t, ws.ExportClientInRoom(wsClient, room))
	require.Equal(t, 0, ws.ExportConnectionsForUser(hub, victim.ID.String()))
}
