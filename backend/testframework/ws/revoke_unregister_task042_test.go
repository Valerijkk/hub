package ws_test

// TASK-042 / P2-BE-14: revoke → Unregister (W3-A4-004).
import (
	"encoding/json"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/ws"
)

// Mutation: убрать UnregisterBySession в handleAuthSessionEvent → client остаётся в room.
func TestBridge_AuthRevoked_UnregistersClient_P2_BE_14(t *testing.T) {
	hub, cancel := startTestHub(t)
	defer cancel()

	client := registerAndSubscribe(t, hub, "user-revoke", "room-secret")
	require.True(t, ws.ExportClientInRoom(client, "room-secret"))

	bridge := ws.NewBridge(hub, nil)
	ws.ExportBridgeHandleAuthSessionEvent(bridge, &nats.Msg{
		Subject: "auth.session.user-revoke.revoked",
		Data:    []byte(`{"reason":"logout"}`),
	})

	require.Eventually(t, func() bool {
		select {
		case <-ws.ExportClientDone(client):
			return true
		default:
			return false
		}
	}, 500*time.Millisecond, 10*time.Millisecond, "client.done must close after revoke")

	require.False(t, ws.ExportClientInRoom(client, "room-secret"))
	require.Equal(t, 0, ws.ExportConnectionsForUser(hub, "user-revoke"))

	// Drain optional revoked notification from send buffer.
	select {
	case raw := <-ws.ExportClientSendCh(client):
		var msg ws.ServerMessage
		_ = json.Unmarshal(raw, &msg)
	default:
	}
}
