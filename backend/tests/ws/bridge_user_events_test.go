//go:build hubtest

package ws_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/ws"
)

type phase151UserEventAuthCall struct {
	UserID, ChannelID uuid.UUID
	Event             string
}

type phase151UserEventAuthorizer struct {
	allowed bool
	err     error
	calls   []phase151UserEventAuthCall
}

func (r *phase151UserEventAuthorizer) AuthorizeUserChannelEvent(
	_ context.Context, userID, channelID uuid.UUID, event string,
) (bool, error) {
	r.calls = append(r.calls, phase151UserEventAuthCall{userID, channelID, event})
	return r.allowed, r.err
}

type phase151AudienceCall struct {
	ChannelID uuid.UUID
	Event     string
}

type phase151AudienceResolver struct {
	recipients []uuid.UUID
	err        error
	calls      []phase151AudienceCall
}

func (r *phase151AudienceResolver) ResolveChannelEventAudience(
	_ context.Context, channelID uuid.UUID, event string,
) ([]uuid.UUID, error) {
	r.calls = append(r.calls, phase151AudienceCall{channelID, event})
	return append([]uuid.UUID(nil), r.recipients...), r.err
}

func TestPhase151MessengerUserEvent_FreshGateTargetsOnlyCanonicalUserRoom(t *testing.T) {
	ctx := context.Background()
	userID, foreignID, channelID := uuid.New(), uuid.New(), uuid.New()
	messageID, rootID := uuid.New(), uuid.New()
	tests := []struct{ event, payload string }{
		{"channel_created", `{"channel_id":"` + channelID.String() + `"}`},
		{"channel_updated", `{"channel_id":"` + channelID.String() + `"}`},
		{"new_message_notification", `{"channel_id":"` + channelID.String() + `","message_id":"` + messageID.String() + `"}`},
		{"unread_changed", `{"channel_id":"` + channelID.String() + `","target_kind":"thread","category":"followed_thread","root_message_id":"` + rootID.String() + `"}`},
	}
	for _, tt := range tests {
		t.Run(tt.event, func(t *testing.T) {
			auth := &phase151UserEventAuthorizer{allowed: true}
			audience := &phase151AudienceResolver{recipients: []uuid.UUID{foreignID, userID}}
			delivery, ok := ws.ExportPhase151AuthorizeAndPlanMessengerUserEventDelivery(
				ctx, "messenger.user."+userID.String()+"."+tt.event, []byte(tt.payload), auth, audience,
			)
			require.True(t, ok)
			assert.Equal(t, userID, delivery.RecipientUserID)
			assert.Equal(t, "user:"+userID.String(), delivery.Room)
			assert.NotEqual(t, "user:"+userID.String()+":channels", delivery.Room)
			assert.Equal(t, tt.event, delivery.Event)
			assert.Equal(t, []phase151UserEventAuthCall{{userID, channelID, tt.event}}, auth.calls)
			assert.Equal(t, []phase151AudienceCall{{channelID, tt.event}}, audience.calls)
		})
	}
}

func TestPhase151MessengerUserEvent_ReauthorizesAfterEnqueueAndFailsClosed(t *testing.T) {
	ctx := context.Background()
	userID, channelID := uuid.New(), uuid.New()
	subject := "messenger.user." + userID.String() + ".unread_changed"
	payload := []byte(`{"channel_id":"` + channelID.String() + `","target_kind":"channel","category":"channel"}`)
	auth := &phase151UserEventAuthorizer{allowed: true}
	audience := &phase151AudienceResolver{recipients: []uuid.UUID{userID}}

	_, ok := ws.ExportPhase151AuthorizeAndPlanMessengerUserEventDelivery(ctx, subject, payload, auth, audience)
	require.True(t, ok)
	auth.allowed = false
	_, ok = ws.ExportPhase151AuthorizeAndPlanMessengerUserEventDelivery(ctx, subject, payload, auth, audience)
	assert.False(t, ok)
	assert.Len(t, auth.calls, 2)
	assert.Len(t, audience.calls, 1, "authorization denial stops before audience lookup")
	auth.allowed, audience.recipients = true, nil
	_, ok = ws.ExportPhase151AuthorizeAndPlanMessengerUserEventDelivery(ctx, subject, payload, auth, audience)
	assert.False(t, ok)
	assert.Len(t, auth.calls, 3)
	assert.Len(t, audience.calls, 2, "recipient audience is resolved fresh at delivery")
}

func TestPhase151MessengerUserEvent_DependencyErrorsFailClosed(t *testing.T) {
	ctx := context.Background()
	userID, channelID := uuid.New(), uuid.New()
	subject := "messenger.user." + userID.String() + ".channel_updated"
	payload := []byte(`{"channel_id":"` + channelID.String() + `"}`)
	auth := &phase151UserEventAuthorizer{allowed: true, err: errors.New("authority unavailable")}
	audience := &phase151AudienceResolver{recipients: []uuid.UUID{userID}}
	_, ok := ws.ExportPhase151AuthorizeAndPlanMessengerUserEventDelivery(ctx, subject, payload, auth, audience)
	assert.False(t, ok)
	assert.Empty(t, audience.calls)

	auth = &phase151UserEventAuthorizer{allowed: true}
	audience = &phase151AudienceResolver{recipients: []uuid.UUID{userID}, err: errors.New("audience unavailable")}
	_, ok = ws.ExportPhase151AuthorizeAndPlanMessengerUserEventDelivery(ctx, subject, payload, auth, audience)
	assert.False(t, ok)
}

func TestPhase151MessengerUserEvent_ChannelRemovedIsContentFreePostRevokeException(t *testing.T) {
	ctx := context.Background()
	userID, channelID := uuid.New(), uuid.New()
	auth := &phase151UserEventAuthorizer{err: errors.New("revoked")}
	audience := &phase151AudienceResolver{err: errors.New("removed")}
	delivery, ok := ws.ExportPhase151AuthorizeAndPlanMessengerUserEventDelivery(
		ctx, "messenger.user."+userID.String()+".channel_removed",
		[]byte(`{"channel_id":"`+channelID.String()+`"}`), auth, audience,
	)
	require.True(t, ok)
	assert.Equal(t, userID, delivery.RecipientUserID)
	assert.Equal(t, "user:"+userID.String(), delivery.Room)
	assert.Equal(t, "channel_removed", delivery.Event)
	assert.Equal(t, map[string]string{"channel_id": channelID.String()}, delivery.Identifiers)
	assert.Empty(t, auth.calls, "removal does not imply renewed access")
	assert.Empty(t, audience.calls, "former member is absent from current audience")

	for _, extra := range []string{`"content":"secret"`, `"count":1`, `"message_id":"` + uuid.NewString() + `"`} {
		payload := []byte(`{"channel_id":"` + channelID.String() + `",` + extra + `}`)
		_, accepted := ws.ExportPhase151AuthorizeAndPlanMessengerUserEventDelivery(
			ctx, "messenger.user."+userID.String()+".channel_removed", payload, auth, audience,
		)
		assert.False(t, accepted)
	}
}

func TestPhase151MessengerUserEvent_UnreadChangedParserIsIdentifierOnly(t *testing.T) {
	ctx := context.Background()
	userID, channelID, rootID := uuid.New(), uuid.New(), uuid.New()
	subject := "messenger.user." + userID.String() + ".unread_changed"
	valid := []string{
		`{"channel_id":"` + channelID.String() + `","target_kind":"channel","category":"channel"}`,
		`{"channel_id":"` + channelID.String() + `","target_kind":"thread","category":"followed_thread","root_message_id":"` + rootID.String() + `"}`,
	}
	for _, payload := range valid {
		auth := &phase151UserEventAuthorizer{allowed: true}
		audience := &phase151AudienceResolver{recipients: []uuid.UUID{userID}}
		_, ok := ws.ExportPhase151AuthorizeAndPlanMessengerUserEventDelivery(ctx, subject, []byte(payload), auth, audience)
		assert.True(t, ok, payload)
	}

	invalid := []string{
		``, `null`, `[]`,
		`{"channel_id":"` + channelID.String() + `","channel_id":"` + channelID.String() + `","target_kind":"channel","category":"channel"}`,
		`{"channel_id":"` + channelID.String() + `","target_kind":"channel","category":"channel","unknown":true}`,
		`{"channel_id":"` + channelID.String() + `","target_kind":"channel","category":"channel","content":"secret"}`,
		`{"channel_id":"` + channelID.String() + `","target_kind":"channel","category":"channel","count":1}`,
		`{"channel_id":"` + channelID.String() + `","target_kind":"thread","category":"followed_thread","root_message_id":1}`,
		`{"channel_id":"` + channelID.String() + `","target_kind":"thread","category":"followed_thread","root_message_id":null}`,
		`{"channel_id":"` + channelID.String() + `","target_kind":"thread","category":"followed_thread","root_message_id":"bad"}`,
		`{"channel_id":"` + channelID.String() + `","target_kind":"channel","category":"channel","root_message_id":"` + rootID.String() + `"}`,
		`{"channel_id":"not-a-uuid","target_kind":"channel","category":"channel"}`,
		`{"channel_id":"` + uuid.Nil.String() + `","target_kind":"channel","category":"channel"}`,
		`{"channel_id":"` + channelID.String() + `","target_kind":"unknown","category":"channel"}`,
		`{"channel_id":"` + channelID.String() + `","target_kind":"channel","category":"unknown"}`,
	}
	for _, payload := range invalid {
		auth := &phase151UserEventAuthorizer{allowed: true}
		audience := &phase151AudienceResolver{recipients: []uuid.UUID{userID}}
		_, ok := ws.ExportPhase151AuthorizeAndPlanMessengerUserEventDelivery(ctx, subject, []byte(payload), auth, audience)
		assert.False(t, ok, payload)
		assert.Empty(t, auth.calls)
		assert.Empty(t, audience.calls)
	}
}

func TestPhase151MessengerUserEvent_MalformedSubjectFailsBeforeDependencies(t *testing.T) {
	ctx := context.Background()
	userID, channelID := uuid.New(), uuid.New()
	payload := []byte(`{"channel_id":"` + channelID.String() + `"}`)
	subjects := []string{
		"", "messenger.user." + userID.String(),
		"messenger.user." + userID.String() + ".channel_updated.extra",
		"chat.user." + userID.String() + ".channel_updated",
		"messenger.account." + userID.String() + ".channel_updated",
		"messenger.user.not-a-uuid.channel_updated",
		"messenger.user." + uuid.Nil.String() + ".channel_updated",
		"messenger.user." + userID.String() + ".unknown",
	}
	for _, subject := range subjects {
		auth := &phase151UserEventAuthorizer{allowed: true}
		audience := &phase151AudienceResolver{recipients: []uuid.UUID{userID}}
		_, ok := ws.ExportPhase151AuthorizeAndPlanMessengerUserEventDelivery(ctx, subject, payload, auth, audience)
		assert.False(t, ok, subject)
		assert.Empty(t, auth.calls)
		assert.Empty(t, audience.calls)
	}
}

func TestPhase151BridgeConstructorRequiresUserEventGates(t *testing.T) {
	auth := &phase151UserEventAuthorizer{allowed: true}
	audience := &phase151AudienceResolver{}
	assert.NotPanics(t, func() { ws.ExportNewBridgeWithMessengerUserEventGatesTest(auth, audience) })
	assert.Panics(t, func() { ws.ExportNewBridgeWithMessengerUserEventGatesTest(nil, audience) })
	assert.Panics(t, func() { ws.ExportNewBridgeWithMessengerUserEventGatesTest(auth, nil) })
}

func receiveBridgeUserServerMessage(t *testing.T, send <-chan []byte) ws.ServerMessage {
	t.Helper()
	select {
	case data := <-send:
		var message ws.ServerMessage
		require.NoError(t, json.Unmarshal(data, &message))
		return message
	case <-time.After(time.Second):
		t.Fatal("expected WebSocket message")
		return ws.ServerMessage{}
	}
}

func assertNoBridgeUserServerMessage(t *testing.T, send <-chan []byte) {
	t.Helper()
	select {
	case data := <-send:
		t.Fatalf("unexpected WebSocket message: %s", data)
	default:
	}
}
