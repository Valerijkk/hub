package calls_test

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/valerii/hub/backend/internal/calls"
	"github.com/valerii/hub/backend/internal/ws"
)

// prodCallWSAuthorizer mirrors wiring_ws.go: NewAuthorizer with calls repo checker.
func prodCallWSAuthorizer(repo *calls.EntRepository) ws.Authorizer {
	return ws.NewAuthorizer(ws.AuthorizerConfig{
		IsCallParticipantOrChannelMember: repo.IsCallParticipantOrChannelMember,
	})
}

func startCallsWSHub(t *testing.T) *ws.Hub {
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
	return hub
}

func waitCallsWS(timeout time.Duration, condition func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return condition()
}

func startCallsWSUpgrade(t *testing.T, onConn func(*websocket.Conn)) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			if string(ctx.Path()) != "/ws" {
				ctx.SetStatusCode(fasthttp.StatusNotFound)
				return
			}
			if upErr := ws.ExportWSUpgrader().Upgrade(ctx, func(conn *websocket.Conn) {
				onConn(conn)
			}); upErr != nil {
				t.Log("ws upgrade failed", upErr)
			}
		},
	}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() {
		_ = srv.Shutdown()
		_ = ln.Close()
	})
	return "ws://" + ln.Addr().String() + "/ws"
}

func dialCallsWS(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// subscribeViaReadPump sends subscribe then ping; pong proves readPump processed subscribe.
func subscribeViaReadPump(t *testing.T, conn *websocket.Conn, client *ws.Client, room string) {
	t.Helper()
	sub, err := json.Marshal(ws.ClientMessage{Type: ws.TypeSubscribe, Room: room})
	require.NoError(t, err)
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, sub))

	ping, err := json.Marshal(ws.ClientMessage{Type: ws.TypePing})
	require.NoError(t, err)
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, ping))

	select {
	case raw := <-ws.ExportClientSendCh(client):
		var pong ws.ServerMessage
		require.NoError(t, json.Unmarshal(raw, &pong))
		assert.Equal(t, ws.TypePong, pong.Type)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("ping did not enqueue pong — readPump stalled")
	}
}

// TestWSAuthorizer_CallRoom_CanSubscribe_A2_007 — prod Authorizer + calls repo:
// stranger deny / participant allow on call:{id}. Repo-only tests miss this path.
// Mutation: nil IsCallParticipantOrChannelMember → participant allow FAIL;
// always-deny instant (ChannelID==nil) → participant allow FAIL.
func TestWSAuthorizer_CallRoom_CanSubscribe_A2_007(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createInstantCall(t, f.owner, "wsauthz01")
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.owner.ID))
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.member.ID))

	authz := prodCallWSAuthorizer(f.repo)
	room := "call:" + callEnt.ID.String()

	ok, err := authz.CanSubscribe(ctx, f.stranger.ID.String(), room)
	require.NoError(t, err)
	assert.False(t, ok, "non-participant must be denied call:{id}")

	ok, err = authz.CanSubscribe(ctx, f.member.ID.String(), room)
	require.NoError(t, err)
	assert.True(t, ok, "active participant must be allowed call:{id}")
}

// TestReadPump_CallAuthorizer_StrangerDenied_A2_007 — readPump + prod Authorizer:
// stranger subscribe to call:{id} must not join room (no broadcast delivery).
// Mutation: skip CanSubscribe in readPump → stranger receives event → FAIL.
func TestReadPump_CallAuthorizer_StrangerDenied_A2_007(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createInstantCall(t, f.owner, "wsauthz02")
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.owner.ID))

	authz := prodCallWSAuthorizer(f.repo)
	room := "call:" + callEnt.ID.String()

	hub := startCallsWSHub(t)
	client := ws.NewClient(f.stranger.ID.String(), "sess-stranger")
	hub.Register(client)
	if !waitCallsWS(500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, f.stranger.ID.String()) == 1
	}) {
		t.Fatal("client not registered")
	}

	done := make(chan struct{})
	url := startCallsWSUpgrade(t, func(serverConn *websocket.Conn) {
		defer close(done)
		ws.ExportReadPump(serverConn, client, hub, authz)
	})
	conn := dialCallsWS(t, url)
	subscribeViaReadPump(t, conn, client, room)

	hub.BroadcastToRoom(room, ws.ServerMessage{Type: ws.TypeEvent, Event: "chat.message"})
	select {
	case raw := <-ws.ExportClientSendCh(client):
		t.Fatalf("stranger must not receive call room events after denied subscribe: %s", raw)
	case <-time.After(150 * time.Millisecond):
		// expected: not in room
	}

	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("readPump did not exit after close")
	}
}

// TestReadPump_CallAuthorizer_ParticipantAllowed_A2_007 — participant subscribe
// succeeds and receives room broadcast (instant-call WS path).
// Mutation: ChannelID==nil → false in repo checker → no delivery → FAIL.
func TestReadPump_CallAuthorizer_ParticipantAllowed_A2_007(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createInstantCall(t, f.owner, "wsauthz03")
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.member.ID))

	authz := prodCallWSAuthorizer(f.repo)
	room := "call:" + callEnt.ID.String()

	hub := startCallsWSHub(t)
	client := ws.NewClient(f.member.ID.String(), "sess-member")
	hub.Register(client)
	if !waitCallsWS(500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, f.member.ID.String()) == 1
	}) {
		t.Fatal("client not registered")
	}
	done := make(chan struct{})
	url := startCallsWSUpgrade(t, func(serverConn *websocket.Conn) {
		defer close(done)
		ws.ExportReadPump(serverConn, client, hub, authz)
	})
	conn := dialCallsWS(t, url)
	subscribeViaReadPump(t, conn, client, room)

	hub.BroadcastToRoom(room, ws.ServerMessage{Type: ws.TypeEvent, Event: "chat.message"})
	select {
	case raw := <-ws.ExportClientSendCh(client):
		var msg ws.ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, ws.TypeEvent, msg.Type)
		assert.Equal(t, "chat.message", msg.Event)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("participant must receive call room events after allowed subscribe")
	}

	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("readPump did not exit after close")
	}
}
