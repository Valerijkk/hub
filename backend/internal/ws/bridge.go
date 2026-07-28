package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/nats-io/nats.go"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"github.com/valerii/hub/backend/internal/platform/observability"
)

// Bridge — связывает NATS-субъекты с WebSocket-комнатами, рассылая события
// из NATS всем подключённым WebSocket-клиентам в соответствующей комнате.
type Bridge struct {
	hub  *Hub
	nats *natsclient.Client
	subs []*nats.Subscription
}

// NewBridge создаёт новый Bridge, связывающий Hub и NATS-клиент.
func NewBridge(hub *Hub, nc *natsclient.Client) *Bridge {
	return &Bridge{
		hub:  hub,
		nats: nc,
	}
}

// Start подписывается на NATS-субъекты и маршрутизирует события в WebSocket-комнаты.
// Для освобождения подписок вызовите Stop().
//
// W3-A4-008: при ошибке N-й подписки откатываем уже созданные через Stop(),
// иначе утечка NATS subscriptions при переиспользовании Bridge.
func (b *Bridge) Start() error {
	type subDef struct {
		subject string
		handler nats.MsgHandler
	}
	defs := []subDef{
		{"project.>", b.traced(b.handleProjectEvent)},
		{"auth.session.>", b.traced(b.handleAuthSessionEvent)},
		{"messenger.channel.>", b.traced(b.handleMessengerEvent)},
		// calls.channel — события звонков в WS-комнаты каналов мессенджера.
		{"calls.channel.>", b.traced(b.handleCallEvent)},
		// calls.chat — эфемерный чат звонка.
		{"calls.chat.>", b.traced(b.handleCallChatEvent)},
		{"notification.user.>", b.traced(b.handleNotificationEvent)},
	}

	for _, d := range defs {
		sub, err := b.nats.Subscribe(d.subject, d.handler)
		if err != nil {
			b.Stop()
			return err
		}
		b.subs = append(b.subs, sub)
	}

	slog.Info("ws bridge started", "subscriptions", len(b.subs))
	return nil
}

// traced оборачивает NATS-handler consumer-span'ом «consume <subject>»,
// продолжающим трейс продюсера (traceparent из headers). Handlers bridge не
// принимают ctx (WS-broadcast синхронный), поэтому span только фиксирует
// факт и длительность доставки в трейсе — без изменения их сигнатур.
func (b *Bridge) traced(handler func(*nats.Msg)) nats.MsgHandler {
	return func(msg *nats.Msg) {
		ctx := natsclient.ExtractTraceContext(context.Background(), msg.Header)
		_, span := observability.StartConsumerSpan(ctx, msg.Subject, msg.Header.Get(natsclient.HeaderEventID))
		defer observability.EndConsumerSpan(span, nil)
		handler(msg)
	}
}

// Stop отписывается от всех NATS-подписок.
func (b *Bridge) Stop() {
	for _, sub := range b.subs {
		_ = sub.Unsubscribe()
	}
	b.subs = nil
}

// handleProjectEvent маршрутизирует NATS-событие проекта в соответствующую WS-комнату.
// Формат subject: project.{projectID}.{event}
// Формат room: project:{projectID}
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
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			slog.Warn("ws bridge: malformed payload", "subject", msg.Subject, "error", err)
			return
		}
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
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			slog.Warn("ws bridge: malformed payload", "subject", msg.Subject, "error", err)
			return
		}
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
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			slog.Warn("ws bridge: malformed payload", "subject", msg.Subject, "error", err)
			return
		}
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
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			slog.Warn("ws bridge: malformed payload", "subject", msg.Subject, "error", err)
			return
		}
	}

	b.hub.BroadcastToRoom(room, ServerMessage{
		Type:    TypeEvent,
		Event:   "chat.message",
		Room:    room,
		Payload: payload,
	})
}

// handleAuthSessionEvent маршрутизирует NATS-событие авторизационной сессии конкретному пользователю.
// Формат subject: auth.session.{userID}.{event}
// W3-A4-004 / P2-BE-14: revoked/logout → Unregister (рвёт комнаты), затем уведомление клиенту.
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
	var sessionID string
	if len(msg.Data) > 0 {
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			slog.Warn("ws bridge: malformed payload", "subject", msg.Subject, "error", err)
			return
		}
		if m, ok := payload.(map[string]any); ok {
			if sid, ok := m["session_id"].(string); ok {
				sessionID = sid
			}
		}
	}

	// Сначала уведомляем (пока conn ещё в maps), затем рвём комнаты.
	b.hub.SendToUser(userID, ServerMessage{
		Type:    TypeEvent,
		Event:   event,
		Payload: payload,
	})

	switch event {
	case "revoked", "logout":
		b.hub.UnregisterBySession(userID, sessionID)
	}
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
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			slog.Warn("ws bridge: malformed payload", "subject", msg.Subject, "error", err)
			return
		}
	}

	b.hub.SendToUser(userID, ServerMessage{
		Type:    TypeEvent,
		Event:   "notification.new",
		Payload: payload,
	})
}
