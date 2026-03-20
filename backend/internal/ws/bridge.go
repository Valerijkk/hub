package ws

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/nats-io/nats.go"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// Bridge connects NATS subjects to WebSocket rooms, fanning out events
// from NATS to all connected WebSocket clients in the corresponding room.
type Bridge struct {
	hub  *Hub
	nats *natsclient.Client
	subs []*nats.Subscription
}

// NewBridge creates a new Bridge linking the given Hub and NATS client.
func NewBridge(hub *Hub, nc *natsclient.Client) *Bridge {
	return &Bridge{
		hub:  hub,
		nats: nc,
	}
}

// Start subscribes to NATS subjects and routes events to WebSocket rooms.
// Call Stop() to clean up subscriptions.
func (b *Bridge) Start() error {
	// project.{projectID}.{event} -> room "project:{projectID}"
	projectSub, err := b.nats.Subscribe("project.>", func(msg *nats.Msg) {
		b.handleProjectEvent(msg)
	})
	if err != nil {
		return err
	}
	b.subs = append(b.subs, projectSub)

	// auth.session.{userID}.{event} -> send to specific user
	authSub, err := b.nats.Subscribe("auth.session.>", func(msg *nats.Msg) {
		b.handleAuthSessionEvent(msg)
	})
	if err != nil {
		return err
	}
	b.subs = append(b.subs, authSub)

	slog.Info("ws bridge started", "subscriptions", len(b.subs))
	return nil
}

// Stop unsubscribes all NATS subscriptions.
func (b *Bridge) Stop() {
	for _, sub := range b.subs {
		_ = sub.Unsubscribe()
	}
	b.subs = nil
}

// handleProjectEvent routes a NATS project event to the corresponding WS room.
// Subject format: project.{projectID}.{event}
// Room format: project:{projectID}
func (b *Bridge) handleProjectEvent(msg *nats.Msg) {
	parts := strings.SplitN(msg.Subject, ".", 3)
	if len(parts) < 2 {
		return
	}

	projectID := parts[1]
	room := "project:" + projectID

	event := msg.Subject
	if len(parts) >= 3 {
		event = parts[2]
	}

	var payload any
	if len(msg.Data) > 0 {
		_ = json.Unmarshal(msg.Data, &payload)
	}

	b.hub.BroadcastToRoom(room, ServerMessage{
		Type:    TypeEvent,
		Event:   event,
		Room:    room,
		Payload: payload,
	})
}

// handleAuthSessionEvent routes a NATS auth session event to a specific user.
// Subject format: auth.session.{userID}.{event}
func (b *Bridge) handleAuthSessionEvent(msg *nats.Msg) {
	parts := strings.SplitN(msg.Subject, ".", 4)
	if len(parts) < 3 {
		return
	}

	userID := parts[2]

	event := msg.Subject
	if len(parts) >= 4 {
		event = parts[3]
	}

	var payload any
	if len(msg.Data) > 0 {
		_ = json.Unmarshal(msg.Data, &payload)
	}

	b.hub.SendToUser(userID, ServerMessage{
		Type:    TypeEvent,
		Event:   event,
		Payload: payload,
	})
}
