package ws

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/valyala/fasthttp"
	"github.com/valerii/hub/backend/internal/auth"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

var upgrader = websocket.FastHTTPUpgrader{
	CheckOrigin: func(_ *fasthttp.RequestCtx) bool { return true },
}

// Handler returns a Fiber handler that upgrades HTTP connections to WebSocket.
// JWT token is passed via query parameter ?token= because WebSocket does not
// support custom headers during the handshake.
//
// CRITICAL (Pitfall 1): All values are extracted from fiber.Ctx BEFORE the
// upgrade, using utils.CopyString to prevent fasthttp RequestCtx reuse issues.
func Handler(hub *Hub, authSvc *auth.AuthService) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Extract and validate JWT from query param BEFORE upgrade (Pitfall 1).
		tokenStr := utils.CopyString(c.Query("token"))
		if tokenStr == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing token query parameter")
		}

		claims, err := authSvc.ParseAccessToken(tokenStr)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		// Copy values before upgrade (Pitfall 1).
		userID := utils.CopyString(claims.UserID)
		sessionID := utils.CopyString(claims.SessionID)

		return upgrader.Upgrade(c.Context(), func(conn *websocket.Conn) {
			client := NewClient(userID, sessionID)
			hub.Register(client)

			defer func() {
				hub.Unregister(client)
				conn.Close()
			}()

			// Start write pump in a separate goroutine.
			go writePump(conn, client)

			// Read pump runs on the current goroutine.
			readPump(conn, client, hub)
		})
	}
}

// readPump reads messages from the WebSocket connection and dispatches them.
func readPump(conn *websocket.Conn, client *Client, hub *Hub) {
	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("ws read error", "userID", client.UserID, "error", err)
			}
			return
		}

		var msg ClientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			slog.Warn("ws invalid message", "userID", client.UserID, "error", err)
			continue
		}

		switch msg.Type {
		case TypeSubscribe:
			if msg.Room != "" {
				hub.Subscribe(client, msg.Room)
			}
		case TypeUnsubscribe:
			if msg.Room != "" {
				hub.Unsubscribe(client, msg.Room)
			}
		case TypePing:
			pong := ServerMessage{Type: TypePong}
			data, _ := json.Marshal(pong)
			select {
			case client.send <- data:
			default:
			}
		}
	}
}

// writePump writes messages from the client send channel to the WebSocket connection.
func writePump(conn *websocket.Conn, client *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-client.send:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel.
				_ = conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
