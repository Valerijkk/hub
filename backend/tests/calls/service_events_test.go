//go:build hubtest

package calls_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/valerii/hub/backend/internal/calls"
)

type capturedPublication struct {
	subject string
	payload any
}

// capturePublisher is the in-memory event boundary for CallEventPublisher
// units. It deliberately has no NATS client or network behaviour.
type capturePublisher struct {
	calls []capturedPublication
}

func (p *capturePublisher) PublishJSON(subject string, payload any) {
	p.calls = append(p.calls, capturedPublication{subject: subject, payload: payload})
}

func TestCallEventPublisherCaptureConstructorRejectsTypedNilPublisher(t *testing.T) {
	var typedNil *capturePublisher

	assert.PanicsWithValue(t, "calls hubtest event publisher must not be nil or typed-nil", func() {
		calls.ExportNewCallEventPublisherWithEventPublisher(typedNil)
	})
}

func TestCallEventPublisherProductionNilNATSRemainsNoOp(t *testing.T) {
	events := calls.NewCallEventPublisher(nil)
	channelID := uuid.New()
	callID := uuid.New()
	userID := uuid.New()
	message := callChatMessageFixture()

	require.NotPanics(t, func() {
		events.Started(&channelID, callID, "team-standup", userID)
		events.ChatMessage(callID.String(), message)
	})
}

func TestCallChannelEventsSuppressNilAndZeroChannelIDs(t *testing.T) {
	callID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name string
		call func(*calls.CallEventPublisher, *uuid.UUID)
	}{
		{
			name: "started",
			call: func(publisher *calls.CallEventPublisher, channelID *uuid.UUID) {
				publisher.Started(channelID, callID, "team-standup", userID)
			},
		},
		{
			name: "joined",
			call: func(publisher *calls.CallEventPublisher, channelID *uuid.UUID) {
				publisher.Joined(channelID, callID, userID, "Ada Lovelace")
			},
		},
		{
			name: "left",
			call: func(publisher *calls.CallEventPublisher, channelID *uuid.UUID) {
				publisher.Left(channelID, callID, userID)
			},
		},
		{
			name: "ended",
			call: func(publisher *calls.CallEventPublisher, channelID *uuid.UUID) {
				publisher.Ended(channelID, callID, 3)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			for _, channelID := range []*uuid.UUID{nil, uuidPtr(uuid.Nil)} {
				publisher := &capturePublisher{}
				events := calls.ExportNewCallEventPublisherWithEventPublisher(publisher)

				testCase.call(events, channelID)

				require.Empty(t, publisher.calls, "nil and uuid.Nil channel IDs must not publish channel events")
			}
		})
	}
}

func TestCallChannelEventsPublishCanonicalSubjectsAndPayloads(t *testing.T) {
	channelID := uuid.New()
	callID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name            string
		suffix          string
		call            func(*calls.CallEventPublisher)
		expectedPayload map[string]interface{}
	}{
		{
			name:   "started",
			suffix: "started",
			call: func(publisher *calls.CallEventPublisher) {
				publisher.Started(&channelID, callID, "team-standup", userID)
			},
			expectedPayload: map[string]interface{}{
				"call_id":    callID,
				"room_name":  "team-standup",
				"created_by": userID,
				"started_by": userID,
			},
		},
		{
			name:   "joined",
			suffix: "joined",
			call: func(publisher *calls.CallEventPublisher) {
				publisher.Joined(&channelID, callID, userID, "Ada Lovelace")
			},
			expectedPayload: map[string]interface{}{
				"call_id":   callID,
				"user_id":   userID,
				"user_name": "Ada Lovelace",
			},
		},
		{
			name:   "left",
			suffix: "left",
			call: func(publisher *calls.CallEventPublisher) {
				publisher.Left(&channelID, callID, userID)
			},
			expectedPayload: map[string]interface{}{
				"call_id": callID,
				"user_id": userID,
			},
		},
		{
			name:   "ended",
			suffix: "ended",
			call: func(publisher *calls.CallEventPublisher) {
				publisher.Ended(&channelID, callID, 3)
			},
			expectedPayload: map[string]interface{}{
				"call_id":           callID,
				"participant_count": 3,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			publisher := &capturePublisher{}
			events := calls.ExportNewCallEventPublisherWithEventPublisher(publisher)

			testCase.call(events)

			require.Len(t, publisher.calls, 1)
			assert.Equal(t, fmt.Sprintf("calls.channel.%s.%s", channelID, testCase.suffix), publisher.calls[0].subject)
			assert.IsType(t, map[string]interface{}{}, publisher.calls[0].payload)
			assert.Equal(t, testCase.expectedPayload, publisher.calls[0].payload)
		})
	}
}

func TestCallChatMessagePublishesCanonicalSubjectAndPayload(t *testing.T) {
	publisher := &capturePublisher{}
	events := calls.ExportNewCallEventPublisherWithEventPublisher(publisher)
	message := callChatMessageFixture()

	events.ChatMessage(message.CallID, message)

	require.Len(t, publisher.calls, 1)
	assert.Equal(t, "calls.chat."+message.CallID, publisher.calls[0].subject)
	assert.IsType(t, map[string]interface{}{}, publisher.calls[0].payload)
	assert.Equal(t, map[string]interface{}{
		"id":         message.ID.Hex(),
		"call_id":    message.CallID,
		"user_id":    message.UserID,
		"user_name":  message.UserName,
		"text":       message.Text,
		"created_at": message.CreatedAt,
	}, publisher.calls[0].payload)
}

func callChatMessageFixture() *calls.CallChatMessage {
	return &calls.CallChatMessage{
		ID:        bson.NewObjectID(),
		CallID:    uuid.NewString(),
		UserID:    uuid.NewString(),
		UserName:  "Ada Lovelace",
		Text:      "Status update",
		CreatedAt: time.Date(2026, time.August, 25, 12, 34, 56, 0, time.UTC),
	}
}

func uuidPtr(value uuid.UUID) *uuid.UUID {
	return &value
}
