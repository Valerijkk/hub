//go:build hubtest

package ws_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/ws"
)

// TASK-037 / P2-BE-02 / W3-A4-005 — deny/error subscribe must not join room
// and must not receive BroadcastToRoom (mutation: ignore !allowed → FAIL).

func TestReadPump_DeniedSubscribe_NotInRoom_TASK037(t *testing.T) {
	hub, _ := startTestHub(t)
	client := ws.NewClient("user-deny-037", "session-deny-037")
	hub.Register(client)
	require.True(t, waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, "user-deny-037") == 1
	}))

	const room = "channel:secret-037"
	done := make(chan struct{})
	url := startWSUpgradeServer(t, func(serverConn *websocket.Conn) {
		defer close(done)
		ws.ExportReadPump(serverConn, client, hub, denyAuthorizer{})
	})

	conn := dialTestWS(t, url)
	sub, _ := json.Marshal(ws.ClientMessage{Type: ws.TypeSubscribe, Room: room})
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, sub))

	// Дать readPump обработать deny до закрытия.
	require.Eventually(t, func() bool {
		return !ws.ExportClientInRoom(client, room)
	}, time.Second, 20*time.Millisecond)

	hub.BroadcastToRoom(room, ws.ServerMessage{Type: ws.TypeEvent, Event: "leak"})
	select {
	case msg := <-ws.ExportClientSendCh(client):
		t.Fatalf("denied client received room broadcast: %q", msg)
	case <-time.After(100 * time.Millisecond):
	}

	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("readPump did not exit after close")
	}
}

func TestReadPump_AuthorizerError_NotInRoom_TASK037(t *testing.T) {
	hub, _ := startTestHub(t)
	client := ws.NewClient("user-authz-err-037", "session-authz-err-037")
	hub.Register(client)
	require.True(t, waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, "user-authz-err-037") == 1
	}))

	const room = "project:secret-037"
	done := make(chan struct{})
	url := startWSUpgradeServer(t, func(serverConn *websocket.Conn) {
		defer close(done)
		ws.ExportReadPump(serverConn, client, hub, errAuthorizer{})
	})

	conn := dialTestWS(t, url)
	sub, _ := json.Marshal(ws.ClientMessage{Type: ws.TypeSubscribe, Room: room})
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, sub))

	require.Eventually(t, func() bool {
		return !ws.ExportClientInRoom(client, room)
	}, time.Second, 20*time.Millisecond)

	hub.BroadcastToRoom(room, ws.ServerMessage{Type: ws.TypeEvent, Event: "leak-err"})
	select {
	case msg := <-ws.ExportClientSendCh(client):
		t.Fatalf("authz-error client received room broadcast: %q", msg)
	case <-time.After(100 * time.Millisecond):
	}

	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("readPump did not exit after close")
	}
}
