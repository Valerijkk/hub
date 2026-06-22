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

func ExportClientRooms(c *Client) map[string]bool { return c.rooms }

func ExportResetOriginCheck() { resetOriginCheck() }

func ExportCheckOrigin(ctx *fasthttp.RequestCtx) bool { return checkOrigin(ctx) }

func ExportWritePump(conn *websocket.Conn, client *Client) { writePump(conn, client) }

func ExportWSUpgrader() *websocket.FastHTTPUpgrader { return &upgrader }
