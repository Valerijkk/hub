//go:build hubtest

package ws_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/ws"
)

func TestBridge_HandleCallChatEvent_IntersectsCurrentAudienceWithCallRoomSubscribers(t *testing.T) {
	fixture := newChannelDeliveryFixture(t)
	callID := uuid.New()
	room := subscribeCallChatRoom(t, fixture, callID)
	resolvedCallIDs := make([]uuid.UUID, 0, 1)

	var resolver ws.CallChatAudienceResolver = func(_ context.Context, resolvedCallID uuid.UUID) ([]uuid.UUID, error) {
		resolvedCallIDs = append(resolvedCallIDs, resolvedCallID)
		return []uuid.UUID{fixture.currentUserID}, nil
	}
	bridge := ws.NewBridge(fixture.hub, nil, nil, resolver)
	ws.ExportBridgeHandleCallChatEvent(bridge, &nats.Msg{
		Subject: "calls.chat." + callID.String(),
		Data:    []byte(`{"message":"hello","kind":"text"}`),
	})

	assert.Equal(t, []uuid.UUID{callID}, resolvedCallIDs)
	assertCallChatEvent(t, fixture, room, map[string]any{"message": "hello", "kind": "text"})
}

func TestBridge_HandleCallChatEvent_FailsClosedWhenAudienceCannotBeResolved(t *testing.T) {
	for _, tt := range []struct {
		name     string
		resolver ws.CallChatAudienceResolver
	}{
		{
			name: "nil resolver",
		},
		{
			name: "lookup error",
			resolver: func(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
				return nil, errors.New("call audience lookup unavailable")
			},
		},
		{
			name: "empty audience",
			resolver: func(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
				return []uuid.UUID{}, nil
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newChannelDeliveryFixture(t)
			callID := uuid.New()
			subscribeCallChatRoom(t, fixture, callID)
			bridge := ws.NewBridge(fixture.hub, nil, nil, tt.resolver)

			handlerResult := startChannelHandler(ws.ExportBridgeHandleCallChatEvent, bridge, &nats.Msg{
				Subject: "calls.chat." + callID.String(),
				Data:    []byte(`{"message":"should not leak"}`),
			})
			select {
			case panicValue := <-handlerResult:
				assert.Nil(t, panicValue)
			case <-time.After(time.Second):
				t.Fatal("call chat handler blocked while failing closed")
			}
			assertNoChannelDelivery(t, fixture)
		})
	}
}

func TestBridge_HandleCallChatEvent_BoundsAudienceLookupAndDropsOnTimeout(t *testing.T) {
	fixture := newChannelDeliveryFixture(t)
	callID := uuid.New()
	subscribeCallChatRoom(t, fixture, callID)
	lookupStarted := make(chan uuid.UUID, 1)
	lookupDeadline := make(chan time.Time, 1)
	lookupCanceled := make(chan struct{}, 1)
	releaseLookup := make(chan struct{})
	t.Cleanup(func() { close(releaseLookup) })

	bridge := ws.NewBridge(fixture.hub, nil, nil, func(ctx context.Context, resolvedCallID uuid.UUID) ([]uuid.UUID, error) {
		deadline, hasDeadline := ctx.Deadline()
		if hasDeadline {
			select {
			case lookupDeadline <- deadline:
			default:
			}
		}
		select {
		case lookupStarted <- resolvedCallID:
		default:
		}
		select {
		case <-ctx.Done():
			select {
			case lookupCanceled <- struct{}{}:
			default:
			}
			return nil, ctx.Err()
		case <-releaseLookup:
			return nil, errors.New("test lookup released before context cancellation")
		}
	})

	handlerResult := startChannelHandler(ws.ExportBridgeHandleCallChatEvent, bridge, &nats.Msg{
		Subject: "calls.chat." + callID.String(),
		Data:    []byte(`{"message":"should not leak"}`),
	})
	select {
	case resolvedCallID := <-lookupStarted:
		assert.Equal(t, callID, resolvedCallID)
	case <-time.After(time.Second):
		t.Fatal("call audience lookup was not started")
	}
	select {
	case deadline := <-lookupDeadline:
		assert.WithinDuration(t, time.Now().Add(time.Second), deadline, 150*time.Millisecond)
	case <-time.After(time.Second):
		t.Fatal("call audience lookup was not given a deadline")
	}
	select {
	case panicValue := <-handlerResult:
		assert.Nil(t, panicValue)
	case <-time.After(1250 * time.Millisecond):
		t.Fatal("call chat handler did not bound the audience lookup")
	}
	select {
	case <-lookupCanceled:
	case <-time.After(time.Second):
		t.Fatal("call audience lookup context was not canceled")
	}
	assertNoChannelDelivery(t, fixture)
}

func TestBridge_HandleCallChatEvent_DropsInvalidSubjectOrPayloadBeforeAudienceLookup(t *testing.T) {
	fixture := newChannelDeliveryFixture(t)
	callID := uuid.New()
	subscribeCallChatRoom(t, fixture, callID)
	resolverCalls := 0
	bridge := ws.NewBridge(fixture.hub, nil, nil, func(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
		resolverCalls++
		return []uuid.UUID{fixture.currentUserID}, nil
	})

	for _, tt := range []struct {
		name    string
		subject string
		data    []byte
	}{
		{
			name:    "subject has too few segments",
			subject: "calls.chat",
			data:    []byte(`{"message":"invalid"}`),
		},
		{
			name:    "subject has trailing segment",
			subject: "calls.chat." + callID.String() + ".extra",
			data:    []byte(`{"message":"invalid"}`),
		},
		{
			name:    "wrong namespace",
			subject: "calls.channel." + callID.String(),
			data:    []byte(`{"message":"invalid"}`),
		},
		{
			name:    "malformed call id",
			subject: "calls.chat.not-a-uuid",
			data:    []byte(`{"message":"invalid"}`),
		},
		{
			name:    "nil call id",
			subject: "calls.chat." + uuid.Nil.String(),
			data:    []byte(`{"message":"invalid"}`),
		},
		{
			name:    "empty payload",
			subject: "calls.chat." + callID.String(),
		},
		{
			name:    "malformed JSON",
			subject: "calls.chat." + callID.String(),
			data:    []byte(`{"message":`),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ws.ExportBridgeHandleCallChatEvent(bridge, &nats.Msg{Subject: tt.subject, Data: tt.data})
			assertNoChannelDelivery(t, fixture)
		})
	}
	assert.Zero(t, resolverCalls)
}

func subscribeCallChatRoom(t *testing.T, fixture channelDeliveryFixture, callID uuid.UUID) string {
	t.Helper()
	room := "call:" + callID.String()
	for _, client := range []*ws.Client{
		fixture.currentClient,
		fixture.currentSecondClient,
		fixture.formerSubscribedClient,
	} {
		fixture.hub.Subscribe(client, room)
	}
	require.Eventually(t, func() bool {
		return ws.ExportClientInRoom(fixture.currentClient, room) &&
			ws.ExportClientInRoom(fixture.currentSecondClient, room) &&
			ws.ExportClientInRoom(fixture.formerSubscribedClient, room)
	}, time.Second, time.Millisecond)
	assert.False(t, ws.ExportClientInRoom(fixture.currentOffRoomClient, room))
	return room
}

func assertCallChatEvent(t *testing.T, fixture channelDeliveryFixture, room string, wantPayload map[string]any) {
	t.Helper()
	for _, client := range []*ws.Client{fixture.currentClient, fixture.currentSecondClient} {
		message := receiveBridgeUserServerMessage(t, ws.ExportClientSendCh(client))
		assert.Equal(t, ws.TypeEvent, message.Type)
		assert.Equal(t, "chat.message", message.Event)
		assert.Equal(t, room, message.Room)
		assert.Equal(t, wantPayload, message.Payload)
	}
	assertNoBridgeUserServerMessage(t, ws.ExportClientSendCh(fixture.currentOffRoomClient))
	assertNoBridgeUserServerMessage(t, ws.ExportClientSendCh(fixture.formerSubscribedClient))
}