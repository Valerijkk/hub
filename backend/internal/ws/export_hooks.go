//go:build hubtest

package ws

import (
	"sync"

	"github.com/fasthttp/websocket"
	"github.com/nats-io/nats.go"
	"github.com/valyala/fasthttp"
)

func ExportRegisterClient(h *Hub, client *Client) {
	h.registerClient(client)
}

func ExportRemoveClient(h *Hub, client *Client) {
	h.removeClient(client)
}

func ExportConnectionsForUser(h *Hub, userID string) int {
	val, ok := h.userClients.Load(userID)
	if !ok {
		return 0
	}
	count := 0
	val.(*sync.Map).Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// ExportHasConnection сообщает, есть ли connectionID в h.clients (P1-BE-21).
func ExportHasConnection(h *Hub, connectionID string) bool {
	_, ok := h.clients.Load(connectionID)
	return ok
}

func ExportBridgeHandleProjectEvent(b *Bridge, msg *nats.Msg) {
	b.handleProjectEvent(msg)
}

func ExportBridgeHandleMessengerEvent(b *Bridge, msg *nats.Msg) {
	b.handleMessengerEvent(msg)
}

func ExportBridgeHandleCallEvent(b *Bridge, msg *nats.Msg) {
	b.handleCallEvent(msg)
}

func ExportBridgeHandleCallChatEvent(b *Bridge, msg *nats.Msg) {
	b.handleCallChatEvent(msg)
}

func ExportBridgeHandleAuthSessionEvent(b *Bridge, msg *nats.Msg) {
	b.handleAuthSessionEvent(msg)
}

func ExportBridgeHandleNotificationEvent(b *Bridge, msg *nats.Msg) {
	b.handleNotificationEvent(msg)
}

func ExportClientSendCh(c *Client) chan []byte {
	return c.send
}

func ExportSplitRoom(room string) (prefix, id string, ok bool) {
	return splitRoom(room)
}

func ExportAllowAllAuthorizer() Authorizer { return allowAllAuthorizer{} }

func ExportClientDone(c *Client) <-chan struct{} { return c.done }

func ExportClientCloseDone(c *Client) {
	c.closeOnce.Do(func() { close(c.done) })
}

func ExportBridgeSubsLen(b *Bridge) int { return len(b.subs) }

func ExportBridgeSubs(b *Bridge) []*nats.Subscription { return b.subs }

func ExportReadPump(conn *websocket.Conn, client *Client, hub *Hub, authz Authorizer) {
	readPump(conn, client, hub, authz)
}

func ExportClientRooms(c *Client) map[string]bool {
	// W3-A4-010: копируем под mu — иначе race с Subscribe/Unregister.
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]bool, len(c.rooms))
	for k, v := range c.rooms {
		out[k] = v
	}
	return out
}

// ExportRoomExists сообщает, есть ли ключ комнаты в h.rooms (в т.ч. пустая map).
func ExportRoomExists(h *Hub, room string) bool {
	_, ok := h.rooms.Load(room)
	return ok
}

// ExportClientInRoom reports whether client is subscribed to room (locked; race-safe).
func ExportClientInRoom(c *Client, room string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.rooms[room]
}

func ExportResetOriginCheck() { resetOriginCheck() }

func ExportCheckOrigin(ctx *fasthttp.RequestCtx) bool { return checkOrigin(ctx) }

func ExportWritePump(conn *websocket.Conn, client *Client) { writePump(conn, client) }

func ExportWSUpgrader() *websocket.FastHTTPUpgrader { return &upgrader }
