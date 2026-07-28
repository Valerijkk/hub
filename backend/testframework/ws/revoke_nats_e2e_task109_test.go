package ws_test

// TASK-109 / A2W3-4-002: live NATS auth.session.*.revoked → Bridge → UnregisterBySession.
// Mutation: убрать case "revoked"/"logout" в handleAuthSessionEvent → client остаётся в room.

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/ws"
)

// TestBridge_LiveNATS_AuthRevoked_UnregistersBySession_TASK109 —
// публикует тот же subject/payload, что auth.RevokeSession, через реальный NATS Bridge.
func TestBridge_LiveNATS_AuthRevoked_UnregistersBySession_TASK109(t *testing.T) {
	hub, _ := startTestHub(t)
	nc := startEmbeddedNATS(t)

	userID := uuid.New().String()
	sessionID := uuid.New().String()
	room := "room-task109"

	client := ws.NewClient(userID, sessionID)
	hub.Register(client)
	hub.Subscribe(client, room)
	require.Eventually(t, func() bool {
		return ws.ExportConnectionsForUser(hub, userID) == 1 && ws.ExportClientInRoom(client, room)
	}, 500*time.Millisecond, 10*time.Millisecond)

	// Чужая сессия того же пользователя — не должна отключиться.
	other := ws.NewClient(userID, uuid.New().String())
	hub.Register(other)
	hub.Subscribe(other, room)
	require.Eventually(t, func() bool {
		return ws.ExportConnectionsForUser(hub, userID) == 2
	}, 500*time.Millisecond, 10*time.Millisecond)

	bridge := ws.NewBridge(hub, nc)
	require.NoError(t, bridge.Start())
	t.Cleanup(bridge.Stop)

	payload, err := json.Marshal(map[string]string{
		"user_id":    userID,
		"session_id": sessionID,
		"reason":     "revoked",
	})
	require.NoError(t, err)

	subject := fmt.Sprintf("auth.session.%s.revoked", userID)
	require.NoError(t, nc.Publish(subject, payload))

	require.Eventually(t, func() bool {
		select {
		case <-ws.ExportClientDone(client):
			return true
		default:
			return false
		}
	}, 2*time.Second, 10*time.Millisecond, "target session client.done must close after revoke")

	require.False(t, ws.ExportClientInRoom(client, room))
	require.Equal(t, 1, ws.ExportConnectionsForUser(hub, userID), "other session must stay connected")
	require.True(t, ws.ExportClientInRoom(other, room))
}

// TestBridge_LiveNATS_AuthLogout_UnregistersBySession_TASK109 —
// auth.session.{uid}.logout с session_id рвёт только целевую сессию.
func TestBridge_LiveNATS_AuthLogout_UnregistersBySession_TASK109(t *testing.T) {
	hub, _ := startTestHub(t)
	nc := startEmbeddedNATS(t)

	userID := uuid.New().String()
	sessionID := uuid.New().String()
	room := "room-logout-109"

	client := ws.NewClient(userID, sessionID)
	hub.Register(client)
	hub.Subscribe(client, room)
	require.Eventually(t, func() bool {
		return ws.ExportClientInRoom(client, room)
	}, 500*time.Millisecond, 10*time.Millisecond)

	bridge := ws.NewBridge(hub, nc)
	require.NoError(t, bridge.Start())
	t.Cleanup(bridge.Stop)

	payload, err := json.Marshal(map[string]string{
		"user_id":    userID,
		"session_id": sessionID,
		"reason":     "logout",
	})
	require.NoError(t, err)

	subject := fmt.Sprintf("auth.session.%s.logout", userID)
	require.NoError(t, nc.Publish(subject, payload))

	require.Eventually(t, func() bool {
		select {
		case <-ws.ExportClientDone(client):
			return true
		default:
			return false
		}
	}, 2*time.Second, 10*time.Millisecond, "client.done must close after logout event")

	require.False(t, ws.ExportClientInRoom(client, room))
	require.Equal(t, 0, ws.ExportConnectionsForUser(hub, userID))
}
