//go:build hubtest

package ws_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/ws"
)

func TestBridge_HandleMessengerUserEvent_RoutesOnlyToExactUserRoom(t *testing.T) {
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

	userID := uuid.NewString()
	foreignUserID := uuid.NewString()
	client := ws.NewClient(userID, "session-owner")
	foreignClient := ws.NewClient(foreignUserID, "session-foreign")
	hub.Register(client)
	hub.Register(foreignClient)

	require.Eventually(t, func() bool {
		return ws.ExportConnectionsForUser(hub, userID) == 1 &&
			ws.ExportConnectionsForUser(hub, foreignUserID) == 1
	}, time.Second, time.Millisecond)

	ownerRoom := "user:" + userID
	foreignRoom := "user:" + foreignUserID
	hub.Subscribe(client, ownerRoom)
	hub.Subscribe(foreignClient, foreignRoom)
	require.Eventually(t, func() bool {
		return ws.ExportClientInRoom(client, ownerRoom) &&
			ws.ExportClientInRoom(foreignClient, foreignRoom)
	}, time.Second, time.Millisecond)

	channelID := uuid.NewString()
	messageID := uuid.NewString()
	bridge := ws.NewBridge(hub, nil, nil, nil)
	for _, tt := range []struct {
		name    string
		event   string
		data    []byte
		payload map[string]any
	}{
		{
			name:    "channel created",
			event:   "channel_created",
			data:    []byte(`{"channel_id":"` + channelID + `"}`),
			payload: map[string]any{"channel_id": channelID},
		},
		{
			name:    "channel updated",
			event:   "channel_updated",
			data:    []byte(`{"channel_id":"` + channelID + `"}`),
			payload: map[string]any{"channel_id": channelID},
		},
		{
			name:    "channel removed",
			event:   "channel_removed",
			data:    []byte(`{"channel_id":"` + channelID + `"}`),
			payload: map[string]any{"channel_id": channelID},
		},
		{
			name:    "new message notification",
			event:   "new_message_notification",
			data:    []byte(`{"channel_id":"` + channelID + `","message_id":"` + messageID + `"}`),
			payload: map[string]any{"channel_id": channelID, "message_id": messageID},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ws.ExportBridgeHandleMessengerUserEvent(bridge, &nats.Msg{
				Subject: "messenger.user." + userID + "." + tt.event,
				Data:    tt.data,
			})

			ownerMessage := receiveBridgeUserServerMessage(t, ws.ExportClientSendCh(client))
			assert.Equal(t, ws.TypeEvent, ownerMessage.Type)
			assert.Equal(t, tt.event, ownerMessage.Event)
			assert.Equal(t, ownerRoom, ownerMessage.Room)
			assert.Equal(t, tt.payload, ownerMessage.Payload)
			assertNoBridgeUserServerMessage(t, ws.ExportClientSendCh(foreignClient))
		})
	}
}

func TestBridge_HandleMessengerUserEvent_DropsMalformedOrUnknownSubjectsAndPayloads(t *testing.T) {
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

	userID := uuid.NewString()
	client := ws.NewClient(userID, "session-owner")
	hub.Register(client)
	require.Eventually(t, func() bool {
		return ws.ExportConnectionsForUser(hub, userID) == 1
	}, time.Second, time.Millisecond)

	room := "user:" + userID
	hub.Subscribe(client, room)
	require.Eventually(t, func() bool {
		return ws.ExportClientInRoom(client, room)
	}, time.Second, time.Millisecond)

	channelID := uuid.NewString()
	messageID := uuid.NewString()
	validLifecyclePayload := []byte(`{"channel_id":"` + channelID + `"}`)
	validNotificationPayload := []byte(`{"channel_id":"` + channelID + `","message_id":"` + messageID + `"}`)
	bridge := ws.NewBridge(hub, nil, nil, nil)
	for _, tt := range []struct {
		name    string
		subject string
		data    []byte
	}{
		{
			name:    "wrong segment count",
			subject: "messenger.user." + userID,
			data:    validLifecyclePayload,
		},
		{
			name:    "overlong subject",
			subject: "messenger.user." + userID + ".channel_created.extra",
			data:    validLifecyclePayload,
		},
		{
			name:    "wrong namespace",
			subject: "chat.user." + userID + ".channel_created",
			data:    validLifecyclePayload,
		},
		{
			name:    "wrong second segment",
			subject: "messenger.account." + userID + ".channel_created",
			data:    validLifecyclePayload,
		},
		{
			name:    "malformed user id",
			subject: "messenger.user.not-a-uuid.channel_created",
			data:    validLifecyclePayload,
		},
		{
			name:    "nil user id",
			subject: "messenger.user." + uuid.Nil.String() + ".channel_created",
			data:    validLifecyclePayload,
		},
		{
			name:    "unknown suffix",
			subject: "messenger.user." + userID + ".unknown",
			data:    validLifecyclePayload,
		},
		{
			name:    "malformed JSON",
			subject: "messenger.user." + userID + ".channel_created",
			data:    []byte(`{"channel_id":`),
		},
		{
			name:    "empty payload",
			subject: "messenger.user." + userID + ".channel_created",
		},
		{
			name:    "null payload",
			subject: "messenger.user." + userID + ".channel_created",
			data:    []byte(`null`),
		},
		{
			name:    "array payload",
			subject: "messenger.user." + userID + ".channel_created",
			data:    []byte(`[]`),
		},
		{
			name:    "lifecycle extra field",
			subject: "messenger.user." + userID + ".channel_created",
			data:    []byte(`{"channel_id":"` + channelID + `","extra":true}`),
		},
		{
			name:    "malformed channel id",
			subject: "messenger.user." + userID + ".channel_created",
			data:    []byte(`{"channel_id":"not-a-uuid"}`),
		},
		{
			name:    "nil channel id",
			subject: "messenger.user." + userID + ".channel_created",
			data:    []byte(`{"channel_id":"` + uuid.Nil.String() + `"}`),
		},
		{
			name:    "null channel id",
			subject: "messenger.user." + userID + ".channel_created",
			data:    []byte(`{"channel_id":null}`),
		},
		{
			name:    "missing channel id",
			subject: "messenger.user." + userID + ".channel_created",
			data:    []byte(`{}`),
		},
		{
			name:    "notification extra field",
			subject: "messenger.user." + userID + ".new_message_notification",
			data:    []byte(`{"channel_id":"` + channelID + `","message_id":"` + messageID + `","extra":true}`),
		},
		{
			name:    "notification missing channel id",
			subject: "messenger.user." + userID + ".new_message_notification",
			data:    []byte(`{"message_id":"` + messageID + `"}`),
		},
		{
			name:    "notification malformed channel id",
			subject: "messenger.user." + userID + ".new_message_notification",
			data:    []byte(`{"channel_id":"not-a-uuid","message_id":"` + messageID + `"}`),
		},
		{
			name:    "notification nil channel id",
			subject: "messenger.user." + userID + ".new_message_notification",
			data:    []byte(`{"channel_id":"` + uuid.Nil.String() + `","message_id":"` + messageID + `"}`),
		},
		{
			name:    "notification null channel id",
			subject: "messenger.user." + userID + ".new_message_notification",
			data:    []byte(`{"channel_id":null,"message_id":"` + messageID + `"}`),
		},
		{
			name:    "notification missing message id",
			subject: "messenger.user." + userID + ".new_message_notification",
			data:    validLifecyclePayload,
		},
		{
			name:    "notification malformed message id",
			subject: "messenger.user." + userID + ".new_message_notification",
			data:    []byte(`{"channel_id":"` + channelID + `","message_id":"not-a-uuid"}`),
		},
		{
			name:    "notification nil message id",
			subject: "messenger.user." + userID + ".new_message_notification",
			data:    []byte(`{"channel_id":"` + channelID + `","message_id":"` + uuid.Nil.String() + `"}`),
		},
		{
			name:    "notification null message id",
			subject: "messenger.user." + userID + ".new_message_notification",
			data:    []byte(`{"channel_id":"` + channelID + `","message_id":null}`),
		},
		{
			name:    "overlong notification subject",
			subject: "messenger.user." + userID + ".new_message_notification.extra",
			data:    validNotificationPayload,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ws.ExportBridgeHandleMessengerUserEvent(bridge, &nats.Msg{
				Subject: tt.subject,
				Data:    tt.data,
			})
			assertNoBridgeUserServerMessage(t, ws.ExportClientSendCh(client))
		})
	}
}

func TestBridge_Subjects_IncludeExactMessengerUserSubscriptionOnce(t *testing.T) {
	bridge := ws.NewBridge(ws.NewHub(), nil, nil, nil)
	subjects := ws.ExportBridgeSubjects(bridge)

	assert.ElementsMatch(t, []string{
		"project.>",
		"auth.session.>",
		"messenger.channel.>",
		"messenger.user.>",
		"calls.channel.>",
		"calls.chat.>",
		"notification.user.>",
	}, subjects)

	messengerUserSubscriptions := 0
	for _, subject := range subjects {
		if subject == "messenger.user.>" {
			messengerUserSubscriptions++
		}
	}
	assert.Equal(t, 1, messengerUserSubscriptions)
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
