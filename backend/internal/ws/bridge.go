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

	// messenger.channel.{channelID}.{event} -> room "channel:{channelID}"
	messengerSub, err := b.nats.Subscribe("messenger.channel.>", func(msg *nats.Msg) {
		b.handleMessengerEvent(msg)
	})
	if err != nil {
		return err
	}
	b.subs = append(b.subs, messengerSub)

	// calls.channel.{channelID}.{event} -> room "channel:{channelID}"
	// Маршрутизирует события звонков (started/ended/joined/left) в WS-комнаты каналов мессенджера.
	callSub, err := b.nats.Subscribe("calls.channel.>", func(msg *nats.Msg) {
		b.handleCallEvent(msg)
	})
	if err != nil {
		return err
	}
	b.subs = append(b.subs, callSub)

	// calls.chat.{callID} -> room "call:{callID}"
	// Маршрутизирует сообщения эфемерного чата звонка для real-time доставки.
	callChatSub, err := b.nats.Subscribe("calls.chat.>", func(msg *nats.Msg) {
		b.handleCallChatEvent(msg)
	})
	if err != nil {
		return err
	}
	b.subs = append(b.subs, callChatSub)

	// notification.user.{userID} -> send to specific user
	// Маршрутизирует уведомления конкретному пользователю через WS.
	notifSub, err := b.nats.Subscribe("notification.user.>", func(msg *nats.Msg) {
		b.handleNotificationEvent(msg)
	})
	if err != nil {
		return err
	}
	b.subs = append(b.subs, notifSub)

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

// handleMessengerEvent маршрутизирует NATS-события мессенджера в соответствующую WS-комнату.
// Формат subject: messenger.channel.{channelID}.{event}
// Формат room: channel:{channelID}
func (b *Bridge) handleMessengerEvent(msg *nats.Msg) {
	parts := strings.SplitN(msg.Subject, ".", 4)
	if len(parts) < 3 {
		return
	}

	channelID := parts[2]
	room := "channel:" + channelID

	event := msg.Subject
	if len(parts) >= 4 {
		event = parts[3]
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

// handleCallEvent маршрутизирует NATS-события звонков в соответствующую WS-комнату канала.
// Формат subject: calls.channel.{channelID}.{event} (started/ended/joined/left)
// Формат room: channel:{channelID} (те же WS-комнаты что у мессенджера)
func (b *Bridge) handleCallEvent(msg *nats.Msg) {
	parts := strings.SplitN(msg.Subject, ".", 4)
	if len(parts) < 3 {
		return
	}

	channelID := parts[2]
	room := "channel:" + channelID

	event := msg.Subject
	if len(parts) >= 4 {
		event = "call." + parts[3]
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

// handleCallChatEvent маршрутизирует сообщения эфемерного чата звонка в WS-комнату звонка.
// Формат subject: calls.chat.{callID}
// Формат room: call:{callID}
func (b *Bridge) handleCallChatEvent(msg *nats.Msg) {
	parts := strings.SplitN(msg.Subject, ".", 3)
	if len(parts) < 3 {
		return
	}

	callID := parts[2]
	room := "call:" + callID

	var payload any
	if len(msg.Data) > 0 {
		_ = json.Unmarshal(msg.Data, &payload)
	}

	b.hub.BroadcastToRoom(room, ServerMessage{
		Type:    TypeEvent,
		Event:   "chat.message",
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

// handleNotificationEvent маршрутизирует уведомления конкретному пользователю.
// Формат subject: notification.user.{userID}
func (b *Bridge) handleNotificationEvent(msg *nats.Msg) {
	parts := strings.SplitN(msg.Subject, ".", 3)
	if len(parts) < 3 {
		return
	}

	userID := parts[2]

	var payload any
	if len(msg.Data) > 0 {
		_ = json.Unmarshal(msg.Data, &payload)
	}

	b.hub.SendToUser(userID, ServerMessage{
		Type:    TypeEvent,
		Event:   "notification.new",
		Payload: payload,
	})
}
