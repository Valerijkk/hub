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

func TestBridge_HandleChannelEvents_IntersectsCurrentMembersWithRoomSubscribers(t *testing.T) {
	fixture := newChannelDeliveryFixture(t)
	resolvedChannelIDs := make([]uuid.UUID, 0, 5)
	bridge := ws.NewBridge(fixture.hub, nil, func(_ context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
		resolvedChannelIDs = append(resolvedChannelIDs, channelID)
		return []uuid.UUID{fixture.currentUserID}, nil
	}, nil)

	// No messenger.user.<former>.channel_removed event reaches this bridge.
	// Both old and current users remain subscribed, so only delivery-time
	// membership filtering can prevent the former member from receiving this.
	ws.ExportBridgeHandleMessengerEvent(bridge, &nats.Msg{
		Subject: "messenger.channel." + fixture.channelID.String() + ".message_sent",
		Data:    []byte(`{"id":"message-123"}`),
	})
	assertChannelEvent(
		t,
		fixture,
		"message_sent",
		map[string]any{"id": "message-123"},
	)

	for _, tt := range []struct {
		suffix    string
		wantEvent string
	}{
		{suffix: "started", wantEvent: "call.started"},
		{suffix: "ended", wantEvent: "call.ended"},
		{suffix: "joined", wantEvent: "call.joined"},
		{suffix: "left", wantEvent: "call.left"},
	} {
		t.Run("calls "+tt.suffix, func(t *testing.T) {
			ws.ExportBridgeHandleCallEvent(bridge, &nats.Msg{
				Subject: "calls.channel." + fixture.channelID.String() + "." + tt.suffix,
				Data:    []byte(`{"call_id":"call-123"}`),
			})
			assertChannelEvent(
				t,
				fixture,
				tt.wantEvent,
				map[string]any{"call_id": "call-123"},
			)
		})
	}

	require.Len(t, resolvedChannelIDs, 5)
	for _, resolvedChannelID := range resolvedChannelIDs {
		assert.Equal(t, fixture.channelID, resolvedChannelID)
	}
}

func TestBridge_HandleChannelEvents_DropWhenMembershipLookupFails(t *testing.T) {
	for _, tt := range validChannelEvents() {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newChannelDeliveryFixture(t)
			lookupErr := errors.New("membership store unavailable")
			resolverCalls := 0
			var resolvedChannelID uuid.UUID
			bridge := ws.NewBridge(fixture.hub, nil, func(_ context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
				resolverCalls++
				resolvedChannelID = channelID
				return nil, lookupErr
			}, nil)

			tt.handle(bridge, &nats.Msg{
				Subject: tt.subject(fixture.channelID),
				Data:    []byte(`{"id":"event-123"}`),
			})

			assert.Equal(t, 1, resolverCalls)
			assert.Equal(t, fixture.channelID, resolvedChannelID)
			assertNoChannelDelivery(t, fixture)
		})
	}
}

func TestBridge_HandleChannelEvents_BoundMembershipLookupAndDropOnTimeout(t *testing.T) {
	for _, tt := range validChannelEvents() {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newChannelDeliveryFixture(t)
			lookupStarted := make(chan uuid.UUID, 1)
			lookupCanceled := make(chan struct{}, 1)
			releaseLookup := make(chan struct{})
			t.Cleanup(func() { close(releaseLookup) })

			bridge := ws.NewBridge(fixture.hub, nil, func(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
				select {
				case lookupStarted <- channelID:
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
			}, nil)

			handlerResult := startChannelHandler(tt.handle, bridge, &nats.Msg{
				Subject: tt.subject(fixture.channelID),
				Data:    []byte(`{"id":"event-123"}`),
			})

			select {
			case channelID := <-lookupStarted:
				assert.Equal(t, fixture.channelID, channelID)
			case <-time.After(time.Second):
				t.Fatal("membership lookup was not started")
			}
			select {
			case panicValue := <-handlerResult:
				assert.Nil(t, panicValue)
			case <-time.After(2 * time.Second):
				t.Fatal("channel event handler did not bound the membership lookup")
			}
			select {
			case <-lookupCanceled:
			case <-time.After(time.Second):
				t.Fatal("membership lookup context was not canceled")
			}
			assertNoChannelDelivery(t, fixture)
		})
	}
}

func TestBridge_HandleChannelEvents_NilResolverFailsClosedWithoutPanic(t *testing.T) {
	for _, tt := range validChannelEvents() {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newChannelDeliveryFixture(t)
			bridge := ws.NewBridge(fixture.hub, nil, nil, nil)

			handlerResult := startChannelHandler(tt.handle, bridge, &nats.Msg{
				Subject: tt.subject(fixture.channelID),
				Data:    []byte(`{"id":"event-123"}`),
			})

			select {
			case panicValue := <-handlerResult:
				assert.Nil(t, panicValue)
			case <-time.After(time.Second):
				t.Fatal("channel event handler blocked with a nil membership resolver")
			}
			assertNoChannelDelivery(t, fixture)
		})
	}
}

func TestBridge_HandleChannelEvents_DropInvalidSubjectOrPayloadBeforeLookup(t *testing.T) {
	fixture := newChannelDeliveryFixture(t)
	resolverCalls := 0
	bridge := ws.NewBridge(fixture.hub, nil, func(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
		resolverCalls++
		return []uuid.UUID{fixture.currentUserID}, nil
	}, nil)

	channelID := fixture.channelID.String()
	nilChannelID := uuid.Nil.String()
	for _, tt := range []struct {
		name    string
		handle  func(*ws.Bridge, *nats.Msg)
		subject string
		data    []byte
	}{
		{
			name:    "messenger non UUID channel id",
			handle:  ws.ExportBridgeHandleMessengerEvent,
			subject: "messenger.channel.not-a-uuid.message_sent",
			data:    []byte(`{"id":"event-123"}`),
		},
		{
			name:    "calls non UUID channel id",
			handle:  ws.ExportBridgeHandleCallEvent,
			subject: "calls.channel.not-a-uuid.started",
			data:    []byte(`{"id":"event-123"}`),
		},
		{
			name:    "calls legacy instant channel id",
			handle:  ws.ExportBridgeHandleCallEvent,
			subject: "calls.channel.instant.started",
			data:    []byte(`{"id":"event-123"}`),
		},
		{
			name:    "messenger nil channel id",
			handle:  ws.ExportBridgeHandleMessengerEvent,
			subject: "messenger.channel." + nilChannelID + ".message_sent",
			data:    []byte(`{"id":"event-123"}`),
		},
		{
			name:    "calls nil channel id",
			handle:  ws.ExportBridgeHandleCallEvent,
			subject: "calls.channel." + nilChannelID + ".started",
			data:    []byte(`{"id":"event-123"}`),
		},
		{
			name:    "messenger fifth segment",
			handle:  ws.ExportBridgeHandleMessengerEvent,
			subject: "messenger.channel." + channelID + ".message_sent.extra",
			data:    []byte(`{"id":"event-123"}`),
		},
		{
			name:    "calls fifth segment",
			handle:  ws.ExportBridgeHandleCallEvent,
			subject: "calls.channel." + channelID + ".started.extra",
			data:    []byte(`{"id":"event-123"}`),
		},
		{
			name:    "messenger empty event",
			handle:  ws.ExportBridgeHandleMessengerEvent,
			subject: "messenger.channel." + channelID + ".",
			data:    []byte(`{"id":"event-123"}`),
		},
		{
			name:    "calls empty event",
			handle:  ws.ExportBridgeHandleCallEvent,
			subject: "calls.channel." + channelID + ".",
			data:    []byte(`{"id":"event-123"}`),
		},
		{
			name:    "messenger wrong namespace",
			handle:  ws.ExportBridgeHandleMessengerEvent,
			subject: "calls.channel." + channelID + ".message_sent",
			data:    []byte(`{"id":"event-123"}`),
		},
		{
			name:    "calls wrong namespace",
			handle:  ws.ExportBridgeHandleCallEvent,
			subject: "messenger.channel." + channelID + ".started",
			data:    []byte(`{"id":"event-123"}`),
		},
		{
			name:    "messenger wrong second segment",
			handle:  ws.ExportBridgeHandleMessengerEvent,
			subject: "messenger.user." + channelID + ".message_sent",
			data:    []byte(`{"id":"event-123"}`),
		},
		{
			name:    "calls wrong second segment",
			handle:  ws.ExportBridgeHandleCallEvent,
			subject: "calls.chat." + channelID + ".started",
			data:    []byte(`{"id":"event-123"}`),
		},
		{
			name:    "calls unknown event",
			handle:  ws.ExportBridgeHandleCallEvent,
			subject: "calls.channel." + channelID + ".unknown",
			data:    []byte(`{"id":"event-123"}`),
		},
		{
			name:    "messenger malformed JSON",
			handle:  ws.ExportBridgeHandleMessengerEvent,
			subject: "messenger.channel." + channelID + ".message_sent",
			data:    []byte(`{"id":`),
		},
		{
			name:    "calls malformed JSON",
			handle:  ws.ExportBridgeHandleCallEvent,
			subject: "calls.channel." + channelID + ".started",
			data:    []byte(`{"id":`),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.handle(bridge, &nats.Msg{Subject: tt.subject, Data: tt.data})
			assertNoChannelDelivery(t, fixture)
		})
	}
	assert.Zero(t, resolverCalls)
}

type validChannelEvent struct {
	name    string
	handle  func(*ws.Bridge, *nats.Msg)
	subject func(uuid.UUID) string
}

func validChannelEvents() []validChannelEvent {
	return []validChannelEvent{
		{
			name:   "messenger",
			handle: ws.ExportBridgeHandleMessengerEvent,
			subject: func(channelID uuid.UUID) string {
				return "messenger.channel." + channelID.String() + ".message_sent"
			},
		},
		{
			name:   "calls",
			handle: ws.ExportBridgeHandleCallEvent,
			subject: func(channelID uuid.UUID) string {
				return "calls.channel." + channelID.String() + ".started"
			},
		},
	}
}

type channelDeliveryFixture struct {
	hub                      *ws.Hub
	channelID                uuid.UUID
	currentUserID            uuid.UUID
	currentClient            *ws.Client
	currentSecondClient      *ws.Client
	currentOffRoomClient     *ws.Client
	formerSubscribedClient   *ws.Client
}

func newChannelDeliveryFixture(t *testing.T) channelDeliveryFixture {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	hub := ws.NewHub()
	runDone := make(chan struct{})
	go func() {
		hub.Run(ctx)
		close(runDone)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-runDone:
		case <-time.After(time.Second):
			t.Error("hub did not stop")
		}
	})

	channelID := uuid.New()
	currentUserID := uuid.New()
	formerUserID := uuid.New()
	fixture := channelDeliveryFixture{
		hub:                    hub,
		channelID:              channelID,
		currentUserID:          currentUserID,
		currentClient:          ws.NewClient(currentUserID.String(), "current-session-1"),
		currentSecondClient:    ws.NewClient(currentUserID.String(), "current-session-2"),
		currentOffRoomClient:   ws.NewClient(currentUserID.String(), "current-session-off-room"),
		formerSubscribedClient: ws.NewClient(formerUserID.String(), "former-session"),
	}

	hub.Register(fixture.currentClient)
	hub.Register(fixture.currentSecondClient)
	hub.Register(fixture.currentOffRoomClient)
	hub.Register(fixture.formerSubscribedClient)
	require.Eventually(t, func() bool {
		return ws.ExportConnectionsForUser(hub, currentUserID.String()) == 3 &&
			ws.ExportConnectionsForUser(hub, formerUserID.String()) == 1
	}, time.Second, time.Millisecond)

	room := "channel:" + channelID.String()
	hub.Subscribe(fixture.currentClient, room)
	hub.Subscribe(fixture.currentSecondClient, room)
	hub.Subscribe(fixture.formerSubscribedClient, room)
	require.Eventually(t, func() bool {
		return ws.ExportClientInRoom(fixture.currentClient, room) &&
			ws.ExportClientInRoom(fixture.currentSecondClient, room) &&
			ws.ExportClientInRoom(fixture.formerSubscribedClient, room)
	}, time.Second, time.Millisecond)
	assert.False(t, ws.ExportClientInRoom(fixture.currentOffRoomClient, room))

	return fixture
}

func assertChannelEvent(
	t *testing.T,
	fixture channelDeliveryFixture,
	wantEvent string,
	wantPayload map[string]any,
) {
	t.Helper()
	room := "channel:" + fixture.channelID.String()
	for _, client := range []*ws.Client{fixture.currentClient, fixture.currentSecondClient} {
		message := receiveBridgeUserServerMessage(t, ws.ExportClientSendCh(client))
		assert.Equal(t, ws.TypeEvent, message.Type)
		assert.Equal(t, wantEvent, message.Event)
		assert.Equal(t, room, message.Room)
		assert.Equal(t, wantPayload, message.Payload)
	}
	assertNoBridgeUserServerMessage(t, ws.ExportClientSendCh(fixture.currentOffRoomClient))
	assertNoBridgeUserServerMessage(t, ws.ExportClientSendCh(fixture.formerSubscribedClient))
}

func assertNoChannelDelivery(t *testing.T, fixture channelDeliveryFixture) {
	t.Helper()
	for _, client := range []*ws.Client{
		fixture.currentClient,
		fixture.currentSecondClient,
		fixture.currentOffRoomClient,
		fixture.formerSubscribedClient,
	} {
		assertNoBridgeUserServerMessage(t, ws.ExportClientSendCh(client))
	}
}

func startChannelHandler(
	handle func(*ws.Bridge, *nats.Msg),
	bridge *ws.Bridge,
	message *nats.Msg,
) <-chan any {
	result := make(chan any, 1)
	go func() {
		var panicValue any
		func() {
			defer func() {
				panicValue = recover()
			}()
			handle(bridge, message)
		}()
		result <- panicValue
	}()
	return result
}
type phase151ChannelAudienceResolver struct {
	recipients []uuid.UUID
	err        error
	calls      []struct {
		channelID uuid.UUID
		event     string
	}
}

func (r *phase151ChannelAudienceResolver) ResolveChannelEventAudience(
	_ context.Context, channelID uuid.UUID, event string,
) ([]uuid.UUID, error) {
	r.calls = append(r.calls, struct {
		channelID uuid.UUID
		event     string
	}{channelID, event})
	return append([]uuid.UUID(nil), r.recipients...), r.err
}

func TestPhase151ChannelDeliveryAudience_IntersectsFreshAuthorizedRecipientsWithSubscribers(t *testing.T) {
	channelID := uuid.New()
	active, former, authorizedOffRoom := uuid.New(), uuid.New(), uuid.New()
	resolver := &phase151ChannelAudienceResolver{recipients: []uuid.UUID{active, authorizedOffRoom, active}}

	recipients := ws.ExportPhase151ResolveChannelDeliveryRecipients(
		context.Background(), channelID, "message_sent", []uuid.UUID{former, active}, resolver,
	)

	assert.Equal(t, []uuid.UUID{active}, recipients)
	assert.Equal(t, []struct {
		channelID uuid.UUID
		event     string
	}{{channelID, "message_sent"}}, resolver.calls)

	resolver.recipients = nil
	recipients = ws.ExportPhase151ResolveChannelDeliveryRecipients(
		context.Background(), channelID, "message_sent", []uuid.UUID{former, active}, resolver,
	)
	assert.Empty(t, recipients, "access revoked after enqueue must suppress delivery")
	assert.Len(t, resolver.calls, 2)
}

func TestPhase151ChannelDeliveryAudience_FailsClosedOnResolverErrorOrNilDependency(t *testing.T) {
	channelID, subscriber := uuid.New(), uuid.New()
	resolver := &phase151ChannelAudienceResolver{recipients: []uuid.UUID{subscriber}, err: errors.New("authority unavailable")}
	assert.Empty(t, ws.ExportPhase151ResolveChannelDeliveryRecipients(
		context.Background(), channelID, "message_sent", []uuid.UUID{subscriber}, resolver,
	))
	assert.Empty(t, ws.ExportPhase151ResolveChannelDeliveryRecipients(
		context.Background(), channelID, "message_sent", []uuid.UUID{subscriber}, nil,
	))
}

func TestPhase151TypingDelivery_ReauthorizesSenderAndRecipientAudience(t *testing.T) {
	channelID := uuid.New()
	senderID := uuid.New()
	active, disabled, removed, projectIneligible := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	subscribers := []uuid.UUID{active, disabled, removed, projectIneligible}
	senderAuthority := &phase151UserEventAuthorizer{allowed: true}
	audience := &phase151ChannelAudienceResolver{recipients: []uuid.UUID{active}}

	recipients := ws.ExportPhase151ResolveTypingDeliveryRecipients(
		context.Background(), channelID, senderID, subscribers, senderAuthority, audience,
	)
	assert.Equal(t, []uuid.UUID{active}, recipients)
	assert.Equal(t, []phase151UserEventAuthCall{{senderID, channelID, "typing"}}, senderAuthority.calls)
	assert.Len(t, audience.calls, 1)

	senderAuthority.allowed = false
	recipients = ws.ExportPhase151ResolveTypingDeliveryRecipients(
		context.Background(), channelID, senderID, subscribers, senderAuthority, audience,
	)
	assert.Empty(t, recipients, "sender access change after publish must suppress typing delivery")
	assert.Len(t, senderAuthority.calls, 2)
	assert.Len(t, audience.calls, 1, "denied sender stops before recipient audience resolution")

	senderAuthority.allowed = true
	audience.recipients = nil
	recipients = ws.ExportPhase151ResolveTypingDeliveryRecipients(
		context.Background(), channelID, senderID, subscribers, senderAuthority, audience,
	)
	assert.Empty(t, recipients, "recipient access change after publish must suppress typing delivery")
	assert.Len(t, senderAuthority.calls, 3)
	assert.Len(t, audience.calls, 2)
}
