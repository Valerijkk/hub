package ws

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startTestHub(t *testing.T) (*Hub, context.CancelFunc) {
	t.Helper()
	hub := NewHub()
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

func registerAndSubscribe(t *testing.T, hub *Hub, userID, room string) *Client {
	t.Helper()
	client := NewClient(userID, "session-"+userID)
	hub.Register(client)
	hub.Subscribe(client, room)
	if !waitFor(t, 500*time.Millisecond, func() bool {
		return hub.connectionsForUser(userID) == 1
	}) {
		t.Fatalf("client %s did not register", userID)
	}
	return client
}

func TestBridge_HandleProjectEvent_RoutesToRoom(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "user-proj", "project:abc-123")

	bridge := NewBridge(hub, nil)
	bridge.handleProjectEvent(&nats.Msg{
		Subject: "project.abc-123.issue.updated",
		Data:    []byte(`{"issue_key":"DEMO-1"}`),
	})

	select {
	case raw := <-client.send:
		var msg ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, TypeEvent, msg.Type)
		assert.Equal(t, "issue.updated", msg.Event)
		assert.Equal(t, "project:abc-123", msg.Room)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("client did not receive project event")
	}
}

func TestBridge_HandleMessengerEvent_RoutesToChannelRoom(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "user-msg", "channel:ch-42")

	bridge := NewBridge(hub, nil)
	bridge.handleMessengerEvent(&nats.Msg{
		Subject: "messenger.channel.ch-42.message.sent",
		Data:    []byte(`{"text":"hello"}`),
	})

	select {
	case raw := <-client.send:
		var msg ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, "message.sent", msg.Event)
		assert.Equal(t, "channel:ch-42", msg.Room)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("client did not receive messenger event")
	}
}

func TestBridge_HandleCallEvent_PrefixesEvent(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "user-call", "channel:ch-call")

	bridge := NewBridge(hub, nil)
	bridge.handleCallEvent(&nats.Msg{
		Subject: "calls.channel.ch-call.started",
		Data:    []byte(`{"call_id":"c-1"}`),
	})

	select {
	case raw := <-client.send:
		var msg ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, "call.started", msg.Event)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("client did not receive call event")
	}
}

func TestBridge_HandleCallChatEvent_RoutesToCallRoom(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "user-chat", "call:call-99")

	bridge := NewBridge(hub, nil)
	bridge.handleCallChatEvent(&nats.Msg{
		Subject: "calls.chat.call-99",
		Data:    []byte(`{"body":"ping"}`),
	})

	select {
	case raw := <-client.send:
		var msg ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, "chat.message", msg.Event)
		assert.Equal(t, "call:call-99", msg.Room)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("client did not receive call chat event")
	}
}

func TestBridge_HandleAuthSessionEvent_SendsToUser(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "user-auth", "ignored")

	bridge := NewBridge(hub, nil)
	bridge.handleAuthSessionEvent(&nats.Msg{
		Subject: "auth.session.user-auth.revoked",
		Data:    []byte(`{"reason":"logout"}`),
	})

	select {
	case raw := <-client.send:
		var msg ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, "revoked", msg.Event)
		assert.Empty(t, msg.Room)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("client did not receive auth session event")
	}
}

func TestBridge_HandleNotificationEvent_SendsToUser(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "user-notif", "ignored")

	bridge := NewBridge(hub, nil)
	bridge.handleNotificationEvent(&nats.Msg{
		Subject: "notification.user.user-notif",
		Data:    []byte(`{"title":"Ping"}`),
	})

	select {
	case raw := <-client.send:
		var msg ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, "notification.new", msg.Event)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("client did not receive notification event")
	}
}

func TestBridge_SkipsMalformedPayload(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "user-bad", "project:bad")

	bridge := NewBridge(hub, nil)
	bridge.handleProjectEvent(&nats.Msg{
		Subject: "project.bad.updated",
		Data:    []byte(`{broken`),
	})

	select {
	case msg := <-client.send:
		t.Fatalf("unexpected message on malformed payload: %q", msg)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestBridge_SkipsShortSubject(t *testing.T) {
	hub := NewHub()
	bridge := NewBridge(hub, nil)
	assert.NotPanics(t, func() {
		bridge.handleProjectEvent(&nats.Msg{Subject: "project", Data: []byte(`{}`)})
		bridge.handleMessengerEvent(&nats.Msg{Subject: "messenger", Data: []byte(`{}`)})
		bridge.handleCallEvent(&nats.Msg{Subject: "calls", Data: []byte(`{}`)})
	})
}
