package ws_test

// Тесты модуля WebSocket: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/ws"
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

type denyAuthorizer struct{}

func (denyAuthorizer) CanSubscribe(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

type allowRoomAuthorizer struct {
	room string
}

func (a allowRoomAuthorizer) CanSubscribe(_ context.Context, _, room string) (bool, error) {
	return room == a.room, nil
}

func startWSUpgradeServer(t *testing.T, onConn func(*websocket.Conn)) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			if string(ctx.Path()) != "/ws" {
				ctx.SetStatusCode(fasthttp.StatusNotFound)
				return
			}
			err := ws.ExportWSUpgrader().Upgrade(ctx, func(conn *websocket.Conn) {
				onConn(conn)
			})
			if err != nil {
				t.Log("ws upgrade failed", err)
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

func dialTestWS(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// TestClient_Send_ReturnsChannel проверяет сценарий: client отправка returns channel.
func TestClient_Send_ReturnsChannel(t *testing.T) {
	c := ws.NewClient("u1", "s1")
	ch := c.Send()
	require.NotNil(t, ch)
	ch <- []byte("ping")
	select {
	case msg := <-ch:
		assert.Equal(t, []byte("ping"), msg)
	default:
		t.Fatal("expected message on send channel")
	}
}

// TestReadPump_HandlesMessages проверяет сценарий: read pump handles сообщения.
func TestReadPump_HandlesMessages(t *testing.T) {
	hub, _ := startTestHub(t)
	client := ws.NewClient("user-pump", "session-pump")
	hub.Register(client)
	if !waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, "user-pump") == 1
	}) {
		t.Fatal("client not registered")
	}

	done := make(chan struct{})
	url := startWSUpgradeServer(t, func(serverConn *websocket.Conn) {
		defer close(done)
		ws.ExportReadPump(serverConn, client, hub, allowRoomAuthorizer{room: "project:abc"})
	})

	conn := dialTestWS(t, url)

	sub, _ := json.Marshal(ws.ClientMessage{Type: ws.TypeSubscribe, Room: "project:abc"})
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, sub))

	ping, _ := json.Marshal(ws.ClientMessage{Type: ws.TypePing})
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, ping))
	select {
	case raw := <-ws.ExportClientSendCh(client):
		var pong ws.ServerMessage
		require.NoError(t, json.Unmarshal(raw, &pong))
		assert.Equal(t, ws.TypePong, pong.Type)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("ping did not enqueue pong on client channel")
	}

	require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`not-json`)))
	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("readPump did not exit after close")
	}
}

// TestReadPump_DeniedSubscribe проверяет сценарий: read pump denied subscribe.
func TestReadPump_DeniedSubscribe(t *testing.T) {
	hub, _ := startTestHub(t)
	client := ws.NewClient("user-deny", "session-deny")
	hub.Register(client)

	done := make(chan struct{})
	url := startWSUpgradeServer(t, func(serverConn *websocket.Conn) {
		defer close(done)
		ws.ExportReadPump(serverConn, client, hub, denyAuthorizer{})
	})

	conn := dialTestWS(t, url)
	sub, _ := json.Marshal(ws.ClientMessage{Type: ws.TypeSubscribe, Room: "channel:secret"})
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, sub))
	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("readPump did not exit after close")
	}
}

// TestWritePump_ForwardsSendChannel проверяет сценарий: write pump forwards отправка channel.
func TestWritePump_ForwardsSendChannel(t *testing.T) {
	client := ws.NewClient("user-write", "session-write")

	done := make(chan struct{})
	url := startWSUpgradeServer(t, func(serverConn *websocket.Conn) {
		defer close(done)
		ws.ExportWritePump(serverConn, client)
	})

	conn := dialTestWS(t, url)
	ws.ExportClientSendCh(client) <- []byte("hello")

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, "hello", string(msg))

	ws.ExportClientCloseDone(client)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("writePump did not exit after done signal")
	}
}

// TestWSHandler_ValidTicket_Upgrades проверяет HTTP-слой: ws handshake через ticket subprotocol.
func TestWSHandler_ValidTicket_Upgrades(t *testing.T) {
	hub, _ := startTestHub(t)
	svc, magic, _ := newWSAuthService(t)
	token := wsAccessToken(t, svc, magic)
	ticket := wsTicketForAccess(t, svc, token)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	app := fiber.New()
	app.Get("/ws", ws.Handler(hub, svc, ws.ExportAllowAllAuthorizer()))
	go func() { _ = app.Listener(ln) }()
	t.Cleanup(func() { _ = ln.Close() })

	wsURL := "ws://" + ln.Addr().String() + "/ws"
	dialer := websocket.Dialer{
		Subprotocols: []string{ws.WSTicketProtocolPrefix + ticket},
	}
	conn, resp, err := dialer.Dial(wsURL, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	require.NoError(t, err)

	ping, _ := json.Marshal(ws.ClientMessage{Type: ws.TypePing})
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, ping))

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := conn.ReadMessage()
	require.NoError(t, err)

	var msg ws.ServerMessage
	require.NoError(t, json.Unmarshal(raw, &msg))
	assert.Equal(t, ws.TypePong, msg.Type)

	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	_ = conn.Close()
}

// TestAllowAllAuthorizer_CanSubscribe проверяет сценарий: allow all авторизатор can subscribe.
func TestAllowAllAuthorizer_CanSubscribe(t *testing.T) {
	ok, err := ws.ExportAllowAllAuthorizer().CanSubscribe(context.Background(), "u1", "room:x")
	require.NoError(t, err)
	assert.True(t, ok)
}

type errAuthorizer struct{}

func (errAuthorizer) CanSubscribe(_ context.Context, _, _ string) (bool, error) {
	return false, assert.AnError
}

// TestReadPump_AuthorizerError проверяет сценарий: read pump авторизатор ошибка.
func TestReadPump_AuthorizerError(t *testing.T) {
	hub, _ := startTestHub(t)
	client := ws.NewClient("user-authz-err", "session-authz-err")
	hub.Register(client)

	done := make(chan struct{})
	url := startWSUpgradeServer(t, func(serverConn *websocket.Conn) {
		defer close(done)
		ws.ExportReadPump(serverConn, client, hub, errAuthorizer{})
	})

	conn := dialTestWS(t, url)
	sub, _ := json.Marshal(ws.ClientMessage{Type: ws.TypeSubscribe, Room: "project:secret"})
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, sub))
	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("readPump did not exit after close")
	}
}

// TestReadPump_Unsubscribe проверяет сценарий: read pump unsubscribe.
func TestReadPump_Unsubscribe(t *testing.T) {
	hub, _ := startTestHub(t)
	client := ws.NewClient("user-unsub", "session-unsub")
	hub.Register(client)
	if !waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, "user-unsub") == 1
	}) {
		t.Fatal("client not registered")
	}

	done := make(chan struct{})
	url := startWSUpgradeServer(t, func(serverConn *websocket.Conn) {
		defer close(done)
		ws.ExportReadPump(serverConn, client, hub, ws.ExportAllowAllAuthorizer())
	})

	conn := dialTestWS(t, url)
	sub, _ := json.Marshal(ws.ClientMessage{Type: ws.TypeSubscribe, Room: "channel:abc"})
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, sub))
	unsub, _ := json.Marshal(ws.ClientMessage{Type: ws.TypeUnsubscribe, Room: "channel:abc"})
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, unsub))
	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("readPump did not exit after close")
	}
}
