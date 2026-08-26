package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/google/uuid"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// MessengerUserEventDelivery is a validated, recipient-scoped WS delivery plan.
type MessengerUserEventDelivery struct {
	RecipientUserID uuid.UUID
	Room            string
	Event           string
	Identifiers     map[string]string
}

type phase151UnreadChangedPayload struct {
	ChannelID    uuid.UUID  `json:"channel_id"`
	TargetKind   string     `json:"target_kind"`
	Category     string     `json:"category"`
	RootMessageID *uuid.UUID `json:"root_message_id,omitempty"`
}

// ExportPhase151AuthorizeAndPlanMessengerUserEventDelivery parses the strict
// subject/payload contract and re-authorizes ordinary events at delivery time.
// channel_removed is the sole content-free post-revocation exception.
func ExportPhase151AuthorizeAndPlanMessengerUserEventDelivery(
	ctx context.Context,
	subject string,
	data []byte,
	authorizer UserChannelEventAuthorizer,
	audience ChannelEventAudienceResolver,
) (MessengerUserEventDelivery, bool) {
	userID, event, ok := parsePhase151MessengerUserSubject(subject)
	if !ok || ctx == nil || ctx.Err() != nil {
		return MessengerUserEventDelivery{}, false
	}
	identifiers, channelID, ok := parsePhase151MessengerUserIdentifiers(data, event)
	if !ok {
		return MessengerUserEventDelivery{}, false
	}
	delivery := MessengerUserEventDelivery{
		RecipientUserID: userID,
		Room:            "user:" + userID.String(),
		Event:           event,
		Identifiers:     identifiers,
	}
	if event == "channel_removed" {
		return delivery, true
	}
	if authorizer == nil || audience == nil {
		return MessengerUserEventDelivery{}, false
	}
	allowed, err := authorizer.AuthorizeUserChannelEvent(ctx, userID, channelID, event)
	if err != nil || !allowed {
		return MessengerUserEventDelivery{}, false
	}
	recipients, err := audience.ResolveChannelEventAudience(ctx, channelID, event)
	if err != nil {
		return MessengerUserEventDelivery{}, false
	}
	for _, recipientID := range recipients {
		if recipientID == userID {
			return delivery, true
		}
	}
	return MessengerUserEventDelivery{}, false
}

func parsePhase151MessengerUserSubject(subject string) (uuid.UUID, string, bool) {
	parts := strings.Split(subject, ".")
	if len(parts) != 4 || parts[0] != "messenger" || parts[1] != "user" {
		return uuid.Nil, "", false
	}
	userID, err := uuid.Parse(parts[2])
	if err != nil || userID == uuid.Nil || !isPhase151MessengerUserEvent(parts[3]) {
		return uuid.Nil, "", false
	}
	return userID, parts[3], true
}

func isPhase151MessengerUserEvent(event string) bool {
	switch event {
	case "channel_created", "channel_updated", "channel_removed", "new_message_notification", "unread_changed":
		return true
	default:
		return false
	}
}

func parsePhase151MessengerUserIdentifiers(data []byte, event string) (map[string]string, uuid.UUID, bool) {
	switch event {
	case "new_message_notification":
		var payload messengerUserNotificationPayload
		if !decodePhase151StrictJSONObject(data, &payload) || payload.ChannelID == uuid.Nil || payload.MessageID == uuid.Nil {
			return nil, uuid.Nil, false
		}
		return map[string]string{
			"channel_id": payload.ChannelID.String(),
			"message_id": payload.MessageID.String(),
		}, payload.ChannelID, true
	case "unread_changed":
		var payload phase151UnreadChangedPayload
		if !decodePhase151StrictJSONObject(data, &payload) || payload.ChannelID == uuid.Nil {
			return nil, uuid.Nil, false
		}
		identifiers := map[string]string{
			"channel_id": payload.ChannelID.String(),
			"target_kind": payload.TargetKind,
			"category": payload.Category,
		}
		switch {
		case payload.TargetKind == "channel" && payload.Category == "channel" && payload.RootMessageID == nil:
			return identifiers, payload.ChannelID, true
		case payload.TargetKind == "thread" && payload.Category == "followed_thread" && payload.RootMessageID != nil && *payload.RootMessageID != uuid.Nil:
			identifiers["root_message_id"] = payload.RootMessageID.String()
			return identifiers, payload.ChannelID, true
		default:
			return nil, uuid.Nil, false
		}
	default:
		var payload messengerUserLifecyclePayload
		if !decodePhase151StrictJSONObject(data, &payload) || payload.ChannelID == uuid.Nil {
			return nil, uuid.Nil, false
		}
		return map[string]string{"channel_id": payload.ChannelID.String()}, payload.ChannelID, true
	}
}

func decodePhase151StrictJSONObject(data []byte, target any) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return false
	}
	seen := make(map[string]struct{})
	for decoder.More() {
		token, err := decoder.Token()
		key, isString := token.(string)
		if err != nil || !isString {
			return false
		}
		if _, duplicate := seen[key]; duplicate {
			return false
		}
		seen[key] = struct{}{}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return false
		}
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') {
		return false
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return false
	}
	return decodeStrictJSONObject(data, target)
}

type UserChannelEventAuthorizerFunc func(context.Context, uuid.UUID, uuid.UUID, string) (bool, error)

func (fn UserChannelEventAuthorizerFunc) AuthorizeUserChannelEvent(
	ctx context.Context,
	userID, channelID uuid.UUID,
	event string,
) (bool, error) {
	return fn(ctx, userID, channelID, event)
}

type ChannelEventAudienceResolverFunc func(context.Context, uuid.UUID, string) ([]uuid.UUID, error)

func (fn ChannelEventAudienceResolverFunc) ResolveChannelEventAudience(
	ctx context.Context,
	channelID uuid.UUID,
	event string,
) ([]uuid.UUID, error) {
	return fn(ctx, channelID, event)
}

// NewBridgeWithMessengerUserEventGates preserves NewBridge while binding the
// mandatory delivery-time gates for messenger.user subjects.
func NewBridgeWithMessengerUserEventGates(
	hub *Hub,
	nc *natsclient.Client,
	channelMembersResolver ChannelMembersResolver,
	callChatAudienceResolver CallChatAudienceResolver,
	authorizer UserChannelEventAuthorizer,
	audience ChannelEventAudienceResolver,
) *Bridge {
	if authorizer == nil || audience == nil {
		panic("messenger user-event authorization gates are required")
	}
	bridge := NewBridge(hub, nc, channelMembersResolver, callChatAudienceResolver)
	bridge.userEventAuthorizer = authorizer
	bridge.userEventAudience = audience
	return bridge
}

func ExportNewBridgeWithMessengerUserEventGatesTest(
	authorizer UserChannelEventAuthorizer,
	audience ChannelEventAudienceResolver,
) *Bridge {
	return NewBridgeWithMessengerUserEventGates(nil, nil, nil, nil, authorizer, audience)
}
