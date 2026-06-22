package ws_test

// Тесты модуля WebSocket: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/ws"
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startTestHub(t *testing.T) (*ws.Hub, context.CancelFunc) {
	t.Helper()
	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		hub.Run(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})
	return hub, cancel
}

func registerAndSubscribe(t *testing.T, hub *ws.Hub, userID, room string) *ws.Client {
	t.Helper()
	client := ws.NewClient(userID, "session-"+userID)
	hub.Register(client)
	hub.Subscribe(client, room)
	if !waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, userID) == 1
	}) {
		t.Fatalf("client %s did not register", userID)
	}
	return client
}

// TestBridge_HandleProjectEvent_RoutesToRoom проверяет сценарий: bridge handle project event маршруты to room.
func TestBridge_HandleProjectEvent_RoutesToRoom(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "user-proj", "project:abc-123")

	bridge := ws.NewBridge(hub, nil)
	ws.ExportBridgeHandleProjectEvent(bridge, &nats.Msg{
		Subject: "project.abc-123.issue.updated",
		Data:    []byte(`{"issue_key":"DEMO-1"}`),
	})

	select {
	case raw := <-ws.ExportClientSendCh(client):
		var msg ws.ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, ws.TypeEvent, msg.Type)
		assert.Equal(t, "issue.updated", msg.Event)
		assert.Equal(t, "project:abc-123", msg.Room)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("client did not receive project event")
	}
}

// TestBridge_HandleMessengerEvent_RoutesToChannelRoom проверяет сценарий: bridge handle мессенджер event маршруты to channel room.
func TestBridge_HandleMessengerEvent_RoutesToChannelRoom(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "user-msg", "channel:ch-42")

	bridge := ws.NewBridge(hub, nil)
	ws.ExportBridgeHandleMessengerEvent(bridge, &nats.Msg{
		Subject: "messenger.channel.ch-42.message.sent",
		Data:    []byte(`{"text":"hello"}`),
	})

	select {
	case raw := <-ws.ExportClientSendCh(client):
		var msg ws.ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, "message.sent", msg.Event)
		assert.Equal(t, "channel:ch-42", msg.Room)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("client did not receive messenger event")
	}
}

// TestBridge_HandleCallEvent_PrefixesEvent проверяет сценарий: bridge handle call event prefixes event.
func TestBridge_HandleCallEvent_PrefixesEvent(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "user-call", "channel:ch-call")

	bridge := ws.NewBridge(hub, nil)
	ws.ExportBridgeHandleCallEvent(bridge, &nats.Msg{
		Subject: "calls.channel.ch-call.started",
		Data:    []byte(`{"call_id":"c-1"}`),
	})

	select {
	case raw := <-ws.ExportClientSendCh(client):
		var msg ws.ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, "call.started", msg.Event)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("client did not receive call event")
	}
}

// TestBridge_HandleCallChatEvent_RoutesToCallRoom проверяет сценарий: bridge handle call chat event маршруты to call room.
func TestBridge_HandleCallChatEvent_RoutesToCallRoom(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "user-chat", "call:call-99")

	bridge := ws.NewBridge(hub, nil)
	ws.ExportBridgeHandleCallChatEvent(bridge, &nats.Msg{
		Subject: "calls.chat.call-99",
		Data:    []byte(`{"body":"ping"}`),
	})

	select {
	case raw := <-ws.ExportClientSendCh(client):
		var msg ws.ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, "chat.message", msg.Event)
		assert.Equal(t, "call:call-99", msg.Room)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("client did not receive call chat event")
	}
}

// TestBridge_HandleAuthSessionEvent_SendsToUser проверяет сценарий: bridge handle auth сессия event sends to user.
func TestBridge_HandleAuthSessionEvent_SendsToUser(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "user-auth", "ignored")

	bridge := ws.NewBridge(hub, nil)
	ws.ExportBridgeHandleAuthSessionEvent(bridge, &nats.Msg{
		Subject: "auth.session.user-auth.revoked",
		Data:    []byte(`{"reason":"logout"}`),
	})

	select {
	case raw := <-ws.ExportClientSendCh(client):
		var msg ws.ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, "revoked", msg.Event)
		assert.Empty(t, msg.Room)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("client did not receive auth session event")
	}
}

// TestBridge_HandleNotificationEvent_SendsToUser проверяет сценарий: bridge handle notification event sends to user.
func TestBridge_HandleNotificationEvent_SendsToUser(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "user-notif", "ignored")

	bridge := ws.NewBridge(hub, nil)
	ws.ExportBridgeHandleNotificationEvent(bridge, &nats.Msg{
		Subject: "notification.user.user-notif",
		Data:    []byte(`{"title":"Ping"}`),
	})

	select {
	case raw := <-ws.ExportClientSendCh(client):
		var msg ws.ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, "notification.new", msg.Event)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("client did not receive notification event")
	}
}

// TestBridge_SkipsMalformedPayload проверяет сценарий: bridge skips malformed payload.
func TestBridge_SkipsMalformedPayload(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "user-bad", "project:bad")

	bridge := ws.NewBridge(hub, nil)
	ws.ExportBridgeHandleProjectEvent(bridge, &nats.Msg{
		Subject: "project.bad.updated",
		Data:    []byte(`{broken`),
	})

	select {
	case msg := <-ws.ExportClientSendCh(client):
		t.Fatalf("unexpected message on malformed payload: %q", msg)
	case <-time.After(100 * time.Millisecond):
	}
}

// TestBridge_SkipsShortSubject проверяет сценарий: bridge skips short subject.
func TestBridge_SkipsShortSubject(t *testing.T) {
	hub := ws.NewHub()
	bridge := ws.NewBridge(hub, nil)
	assert.NotPanics(t, func() {
		ws.ExportBridgeHandleProjectEvent(bridge, &nats.Msg{Subject: "project", Data: []byte(`{}`)})
		ws.ExportBridgeHandleMessengerEvent(bridge, &nats.Msg{Subject: "messenger", Data: []byte(`{}`)})
		ws.ExportBridgeHandleCallEvent(bridge, &nats.Msg{Subject: "calls", Data: []byte(`{}`)})
	})
}
