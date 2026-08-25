//go:build hubtest

package calls_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/valerii/hub/backend/internal/calls"
)

func TestCallChatAudienceIsStableDeduplicatedAndCallScoped(t *testing.T) {
	participantA := uuid.New()
	participantB := uuid.New()
	channelMemberC := uuid.New()
	creatorD := uuid.New()
	channelID := uuid.New()

	tests := []struct {
		name                 string
		activeParticipantIDs []uuid.UUID
		channelMemberIDs     []uuid.UUID
		creatorID            uuid.UUID
		channelID            *uuid.UUID
		want                 []uuid.UUID
	}{
		{
			name:                 "channel call includes creator once when active and current member",
			activeParticipantIDs: []uuid.UUID{participantA, participantB, participantA, uuid.Nil},
			channelMemberIDs:     []uuid.UUID{participantB, channelMemberC, uuid.Nil},
			creatorID:            participantB,
			channelID:            &channelID,
			want:                 []uuid.UUID{participantA, participantB, channelMemberC},
		},
		{
			name:                 "channel call does not auto add creator outside stable union",
			activeParticipantIDs: []uuid.UUID{participantA, uuid.Nil, participantA},
			channelMemberIDs:     []uuid.UUID{participantB, channelMemberC, participantB},
			creatorID:            creatorD,
			channelID:            &channelID,
			want:                 []uuid.UUID{participantA, participantB, channelMemberC},
		},
		{
			name:                 "instant call keeps active participants and adds creator once",
			activeParticipantIDs: []uuid.UUID{participantA, uuid.Nil, participantA},
			creatorID:            participantB,
			want:                 []uuid.UUID{participantA, participantB},
		},
		{
			name:                 "instant call does not duplicate creator already active",
			activeParticipantIDs: []uuid.UUID{participantA, participantB, uuid.Nil, participantB, participantA},
			creatorID:            participantB,
			want:                 []uuid.UUID{participantA, participantB},
		},
		{
			name:                 "instant call skips nil creator",
			activeParticipantIDs: []uuid.UUID{participantA, uuid.Nil},
			creatorID:            uuid.Nil,
			want:                 []uuid.UUID{participantA},
		},
		{
			name:                 "zero channel id fails closed to active participants only",
			activeParticipantIDs: []uuid.UUID{participantA, uuid.Nil, participantA},
			channelMemberIDs:     []uuid.UUID{participantB, channelMemberC},
			creatorID:            creatorD,
			channelID:            uuidPtr(uuid.Nil),
			want:                 []uuid.UUID{participantA},
		},
		{
			name: "empty inputs remain empty",
			want: []uuid.UUID{},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.want, calls.ExportBuildCallChatAudience(
				testCase.activeParticipantIDs,
				testCase.channelMemberIDs,
				testCase.creatorID,
				testCase.channelID,
			))
		})
	}
}