package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"github.com/valerii/hub/backend/internal/platform/observability"
)

// Bridge — связывает NATS-субъекты с WebSocket-комнатами, рассылая события
// из NATS всем подключённым WebSocket-клиентам в соответствующей комнате.
type Bridge struct {
	hub                     *Hub
	nats                    *natsclient.Client
	channelMembersResolver  ChannelMembersResolver
	callChatAudienceResolver CallChatAudienceResolver
	userEventAuthorizer      UserChannelEventAuthorizer
	userEventAudience        ChannelEventAudienceResolver
	subs                    []*nats.Subscription
}

// ChannelMembersResolver returns the current members of a channel. The bridge
// resolves them at delivery time to prevent stale room subscriptions leaking
// channel events after membership changes.
type ChannelMembersResolver func(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error)

// CallChatAudienceResolver returns the current users who may receive a call
// chat event. The bridge resolves them at delivery time so stale call room
// subscriptions cannot receive messages after the caller loses access.
type CallChatAudienceResolver func(ctx context.Context, callID uuid.UUID) ([]uuid.UUID, error)

type bridgeSubscriptionDefinition struct {
	subject string
	handler func(*Bridge, *nats.Msg)
}

// bridgeSubscriptionDefinitions is the sole source of Bridge subscriptions.
// It is shared by Start and the hubtest export hook to prevent drift.
var bridgeSubscriptionDefinitions = []bridgeSubscriptionDefinition{
	{subject: "project.>", handler: (*Bridge).handleProjectEvent},
	{subject: "auth.session.>", handler: (*Bridge).handleAuthSessionEvent},
	{subject: "messenger.channel.>", handler: (*Bridge).handleMessengerEvent},
	{subject: "messenger.user.>", handler: (*Bridge).handleMessengerUserEvent},
	{subject: "calls.channel.>", handler: (*Bridge).handleCallEvent},
	{subject: "calls.chat.>", handler: (*Bridge).handleCallChatEvent},
	{subject: "notification.user.>", handler: (*Bridge).handleNotificationEvent},
}

// NewBridge creates a bridge between Hub and the NATS client.
func NewBridge(
	hub *Hub,
	nc *natsclient.Client,
	channelMembersResolver ChannelMembersResolver,
	callChatAudienceResolver CallChatAudienceResolver,
) *Bridge {
	return &Bridge{
		hub:                     hub,
		nats:                    nc,
		channelMembersResolver:  channelMembersResolver,
		callChatAudienceResolver: callChatAudienceResolver,
	}
}

// Start подписывается на NATS-субъекты и маршрутизирует события в WebSocket-комнаты.
// Для освобождения подписок вызовите Stop().
//
// W3-A4-008: при ошибке N-й подписки откатываем уже созданные через Stop(),
// иначе утечка NATS subscriptions при переиспользовании Bridge.
func (b *Bridge) Start() error {
	for _, definition := range bridgeSubscriptionDefinitions {
		handler := definition.handler
		sub, err := b.nats.Subscribe(definition.subject, b.traced(func(msg *nats.Msg) {
			handler(b, msg)
		}))
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
	channelID, event, ok := parseChannelEventSubject(msg, "messenger", isMessengerChannelEvent)
	if !ok {
		return
	}
	payload, ok := parseOptionalJSONPayload(msg)
	if !ok {
		return
	}
	b.broadcastAuthorizedChannelEvent(channelID, event, payload)
}

// handleMessengerUserEvent routes recipient-scoped lifecycle and notification
// events only to the exact user-owned room.
func (b *Bridge) handleMessengerUserEvent(msg *nats.Msg) {
	if msg == nil {
		return
	}
	if b.userEventAuthorizer != nil || b.userEventAudience != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		delivery, ok := ExportPhase151AuthorizeAndPlanMessengerUserEventDelivery(
			ctx,
			msg.Subject,
			msg.Data,
			b.userEventAuthorizer,
			b.userEventAudience,
		)
		if !ok {
			return
		}
		b.hub.BroadcastToRoom(delivery.Room, ServerMessage{
			Type:    TypeEvent,
			Event:   delivery.Event,
			Room:    delivery.Room,
			Payload: delivery.Identifiers,
		})
		return
	}

	userID, event, ok := parseMessengerUserEventSubject(msg)
	if !ok {
		return
	}
	payload, ok := parseMessengerUserPayload(msg.Data, event)
	if !ok {
		slog.Warn("ws bridge: malformed messenger user payload", "subject", msg.Subject)
		return
	}
	room := "user:" + userID.String()
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
	channelID, event, ok := parseChannelEventSubject(msg, "calls", isCallChannelEvent)
	if !ok {
		return
	}
	payload, ok := parseOptionalJSONPayload(msg)
	if !ok {
		return
	}
	b.broadcastAuthorizedChannelEvent(channelID, "call."+event, payload)
}

func (b *Bridge) broadcastAuthorizedChannelEvent(channelID uuid.UUID, event string, payload any) {
	if b.channelMembersResolver == nil {
		slog.Warn("ws bridge: channel members resolver is unavailable", "channel_id", channelID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	members, err := b.channelMembersResolver(ctx, channelID)
	if err != nil {
		slog.Warn("ws bridge: channel membership lookup failed", "channel_id", channelID, "error", err)
		return
	}

	allowedUserIDs := make(map[uuid.UUID]struct{}, len(members))
	for _, memberID := range members {
		if memberID != uuid.Nil {
			allowedUserIDs[memberID] = struct{}{}
		}
	}
	if len(allowedUserIDs) == 0 {
		return
	}

	room := "channel:" + channelID.String()
	b.hub.BroadcastToRoomForUsers(room, ServerMessage{
		Type:    TypeEvent,
		Event:   event,
		Room:    room,
		Payload: payload,
	}, allowedUserIDs)
}

func parseChannelEventSubject(msg *nats.Msg, namespace string, allowedEvent func(string) bool) (uuid.UUID, string, bool) {
	if msg == nil {
		return uuid.Nil, "", false
	}
	parts := strings.Split(msg.Subject, ".")
	if len(parts) != 4 || parts[0] != namespace || parts[1] != "channel" || parts[3] == "" || !allowedEvent(parts[3]) {
		return uuid.Nil, "", false
	}
	channelID, err := uuid.Parse(parts[2])
	if err != nil || channelID == uuid.Nil {
		return uuid.Nil, "", false
	}
	return channelID, parts[3], true
}

func isMessengerChannelEvent(event string) bool {
	switch event {
	case "message_sent", "message_edited", "message_deleted", "reaction_toggled", "typing", "member_joined", "member_left", "poll_created", "poll_voted", "poll_closed":
		return true
	default:
		return false
	}
}

func isCallChannelEvent(event string) bool {
	switch event {
	case "started", "ended", "joined", "left":
		return true
	default:
		return false
	}
}

func parseOptionalJSONPayload(msg *nats.Msg) (any, bool) {
	if msg == nil {
		return nil, false
	}
	if len(msg.Data) == 0 {
		return nil, true
	}
	var payload any
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		slog.Warn("ws bridge: malformed payload", "subject", msg.Subject, "error", err)
		return nil, false
	}
	return payload, true
}

func parseMessengerUserEventSubject(msg *nats.Msg) (uuid.UUID, string, bool) {
	if msg == nil {
		return uuid.Nil, "", false
	}
	parts := strings.Split(msg.Subject, ".")
	if len(parts) != 4 || parts[0] != "messenger" || parts[1] != "user" || !isMessengerUserEvent(parts[3]) {
		return uuid.Nil, "", false
	}
	userID, err := uuid.Parse(parts[2])
	if err != nil || userID == uuid.Nil {
		return uuid.Nil, "", false
	}
	return userID, parts[3], true
}

func isMessengerUserEvent(event string) bool {
	switch event {
	case "channel_created", "channel_updated", "channel_removed", "new_message_notification":
		return true
	default:
		return false
	}
}

type messengerUserLifecyclePayload struct {
	ChannelID uuid.UUID `json:"channel_id"`
}

type messengerUserNotificationPayload struct {
	ChannelID uuid.UUID `json:"channel_id"`
	MessageID uuid.UUID `json:"message_id"`
}

func parseMessengerUserPayload(data []byte, event string) (map[string]any, bool) {
	if event == "new_message_notification" {
		var payload messengerUserNotificationPayload
		if !decodeStrictJSONObject(data, &payload) || payload.ChannelID == uuid.Nil || payload.MessageID == uuid.Nil {
			return nil, false
		}
		return map[string]any{
			"channel_id": payload.ChannelID.String(),
			"message_id": payload.MessageID.String(),
		}, true
	}

	var payload messengerUserLifecyclePayload
	if !decodeStrictJSONObject(data, &payload) || payload.ChannelID == uuid.Nil {
		return nil, false
	}
	return map[string]any{"channel_id": payload.ChannelID.String()}, true
}

func decodeStrictJSONObject(data []byte, target any) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return false
	}
	var trailing any
	return decoder.Decode(&trailing) == io.EOF
}
// handleCallChatEvent маршрутизирует сообщения эфемерного чата звонка в WS-комнату звонка.
// Формат subject: calls.chat.{callID}
// Формат room: call:{callID}
func (b *Bridge) handleCallChatEvent(msg *nats.Msg) {
	callID, ok := parseCallChatSubject(msg)
	if !ok {
		return
	}
	payload, ok := parseRequiredJSONPayload(msg)
	if !ok {
		return
	}
	if b.callChatAudienceResolver == nil {
		slog.Warn("ws bridge: call chat audience resolver is unavailable", "call_id", callID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	audience, err := b.callChatAudienceResolver(ctx, callID)
	if err != nil {
		slog.Warn("ws bridge: call chat audience lookup failed", "call_id", callID, "error", err)
		return
	}

	allowedUserIDs := make(map[uuid.UUID]struct{}, len(audience))
	for _, userID := range audience {
		if userID != uuid.Nil {
			allowedUserIDs[userID] = struct{}{}
		}
	}
	if len(allowedUserIDs) == 0 {
		return
	}

	room := "call:" + callID.String()
	b.hub.BroadcastToRoomForUsers(room, ServerMessage{
		Type:    TypeEvent,
		Event:   "chat.message",
		Room:    room,
		Payload: payload,
	}, allowedUserIDs)
}

func parseCallChatSubject(msg *nats.Msg) (uuid.UUID, bool) {
	if msg == nil {
		return uuid.Nil, false
	}
	parts := strings.Split(msg.Subject, ".")
	if len(parts) != 3 || parts[0] != "calls" || parts[1] != "chat" {
		return uuid.Nil, false
	}
	callID, err := uuid.Parse(parts[2])
	if err != nil || callID == uuid.Nil {
		return uuid.Nil, false
	}
	return callID, true
}

func parseRequiredJSONPayload(msg *nats.Msg) (any, bool) {
	if msg == nil || len(msg.Data) == 0 {
		return nil, false
	}
	var payload any
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		slog.Warn("ws bridge: malformed payload", "subject", msg.Subject, "error", err)
		return nil, false
	}
	return payload, true
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
