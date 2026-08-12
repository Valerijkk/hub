//go:build hubtest

package ws_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/ws"
)

// TASK-064 / P3-BE-10 — empty rooms cleanup, bridge Start→Stop, deny frame, ExportRooms под mu.

func TestHub_EmptyRoomDeletedFromMap_TASK064(t *testing.T) {
	hub, _ := startTestHub(t)
	const room = "channel:cleanup-064"
	client := registerAndSubscribe(t, hub, "user-cleanup-064", room)
	require.True(t, ws.ExportRoomExists(hub, room))

	hub.Unsubscribe(client, room)
	require.True(t, waitFor(t, 500*time.Millisecond, func() bool {
		return !ws.ExportClientInRoom(client, room) && !ws.ExportRoomExists(hub, room)
	}), "empty room key must be deleted from hub.rooms")
}

func TestHub_EmptyRoomDeletedOnUnregister_TASK064(t *testing.T) {
	hub, _ := startTestHub(t)
	const room = "project:unreg-064"
	client := registerAndSubscribe(t, hub, "user-unreg-064", room)
	require.True(t, ws.ExportRoomExists(hub, room))

	hub.Unregister(client)
	require.True(t, waitFor(t, 500*time.Millisecond, func() bool {
		return !ws.ExportRoomExists(hub, room)
	}), "unregister of last member must delete empty room")
}

func TestBridge_StartFail_CallsStop_TASK064(t *testing.T) {
	hub, _ := startTestHub(t)
	nc := startEmbeddedNATS(t)
	bridge := ws.NewBridge(hub, nc)
	require.NoError(t, bridge.Start())
	assert.Equal(t, 6, ws.ExportBridgeSubsLen(bridge))

	nc.Close()
	err := bridge.Start()
	require.Error(t, err)
	assert.Equal(t, 0, ws.ExportBridgeSubsLen(bridge), "Start fail must Stop and clear prior subs")
}

func TestReadPump_DeniedSubscribe_SendsErrorFrame_TASK064(t *testing.T) {
	hub, _ := startTestHub(t)
	client := ws.NewClient("user-deny-064", "session-deny-064")
	hub.Register(client)
	require.True(t, waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, "user-deny-064") == 1
	}))

	const room = "channel:deny-frame-064"
	done := make(chan struct{})
	url := startWSUpgradeServer(t, func(serverConn *websocket.Conn) {
		defer close(done)
		ws.ExportReadPump(serverConn, client, hub, denyAuthorizer{})
	})

	conn := dialTestWS(t, url)
	sub, _ := json.Marshal(ws.ClientMessage{Type: ws.TypeSubscribe, Room: room})
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, sub))

	select {
	case raw := <-ws.ExportClientSendCh(client):
		var msg ws.ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, ws.TypeError, msg.Type)
		assert.Equal(t, ws.EventSubscribeDenied, msg.Event)
		assert.Equal(t, room, msg.Room)
	case <-time.After(time.Second):
		t.Fatal("expected subscribe_denied error frame")
	}

	assert.False(t, ws.ExportClientInRoom(client, room))

	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("readPump did not exit after close")
	}
}

func TestReadPump_AuthorizerError_SendsErrorFrame_TASK064(t *testing.T) {
	hub, _ := startTestHub(t)
	client := ws.NewClient("user-err-064", "session-err-064")
	hub.Register(client)
	require.True(t, waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, "user-err-064") == 1
	}))

	const room = "project:err-frame-064"
	done := make(chan struct{})
	url := startWSUpgradeServer(t, func(serverConn *websocket.Conn) {
		defer close(done)
		ws.ExportReadPump(serverConn, client, hub, errAuthorizer{})
	})

	conn := dialTestWS(t, url)
	sub, _ := json.Marshal(ws.ClientMessage{Type: ws.TypeSubscribe, Room: room})
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, sub))

	select {
	case raw := <-ws.ExportClientSendCh(client):
		var msg ws.ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, ws.TypeError, msg.Type)
		assert.Equal(t, ws.EventSubscribeError, msg.Event)
		assert.Equal(t, room, msg.Room)
	case <-time.After(time.Second):
		t.Fatal("expected subscribe_error error frame")
	}

	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("readPump did not exit after close")
	}
}

func TestExportClientRooms_CopyUnderMu_TASK064(t *testing.T) {
	c := ws.NewClient("user-rooms-064", "session-rooms-064")
	hub, _ := startTestHub(t)
	hub.Register(c)
	require.True(t, waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, "user-rooms-064") == 1
	}))
	hub.Subscribe(c, "room:a")
	require.True(t, waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportClientInRoom(c, "room:a")
	}))

	snap := ws.ExportClientRooms(c)
	require.True(t, snap["room:a"])
	snap["room:mutated"] = true
	assert.False(t, ws.ExportClientInRoom(c, "room:mutated"), "ExportClientRooms must return a copy")
}
