//go:build hubtest

package calls_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/calls"
)

func TestReduceConfirmedParticipantObservationRejectsInvalidInput(t *testing.T) {
	validObservation := calls.ConfirmedParticipantObservation{
		CallID: confirmedPresenceCallID,
		UserID: confirmedPresenceUserID,
		Kind:   calls.ParticipantPresenceConnected,
	}

	tests := []struct {
		name        string
		previous    calls.ParticipantPresence
		observation calls.ConfirmedParticipantObservation
	}{
		{
			name:        "unknown previous presence",
			previous:    calls.ParticipantPresence(99),
			observation: validObservation,
		},
		{
			name:     "zero call ID",
			previous: calls.ParticipantPresenceUnknown,
			observation: calls.ConfirmedParticipantObservation{
				UserID: confirmedPresenceUserID,
				Kind:   calls.ParticipantPresenceConnected,
			},
		},
		{
			name:     "zero user ID",
			previous: calls.ParticipantPresenceUnknown,
			observation: calls.ConfirmedParticipantObservation{
				CallID: confirmedPresenceCallID,
				Kind:   calls.ParticipantPresenceConnected,
			},
		},
		{
			name:     "unknown observation kind",
			previous: calls.ParticipantPresenceUnknown,
			observation: calls.ConfirmedParticipantObservation{
				CallID: confirmedPresenceCallID,
				UserID: confirmedPresenceUserID,
				Kind:   calls.ParticipantPresence(99),
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			transition, err := calls.ReduceConfirmedParticipantObservation(testCase.previous, testCase.observation)

			assert.Equal(t, calls.ParticipantPresenceTransition{}, transition)
			require.ErrorIs(t, err, calls.ErrConfirmedParticipantObservationInvalid)
		})
	}
}

func TestReduceConfirmedParticipantObservationTransitions(t *testing.T) {
	tests := []struct {
		name     string
		previous calls.ParticipantPresence
		kind     calls.ParticipantPresence
		want     calls.ParticipantPresenceTransition
	}{
		{
			name:     "connect from unknown",
			previous: calls.ParticipantPresenceUnknown,
			kind:     calls.ParticipantPresenceConnected,
			want: calls.ParticipantPresenceTransition{
				Previous:      calls.ParticipantPresenceUnknown,
				Next:          calls.ParticipantPresenceConnected,
				Persist:       true,
				PublishJoined: true,
			},
		},
		{
			name:     "reconnect from left",
			previous: calls.ParticipantPresenceLeft,
			kind:     calls.ParticipantPresenceConnected,
			want: calls.ParticipantPresenceTransition{
				Previous:      calls.ParticipantPresenceLeft,
				Next:          calls.ParticipantPresenceConnected,
				Persist:       true,
				PublishJoined: true,
			},
		},
		{
			name:     "duplicate connect",
			previous: calls.ParticipantPresenceConnected,
			kind:     calls.ParticipantPresenceConnected,
			want: calls.ParticipantPresenceTransition{
				Previous: calls.ParticipantPresenceConnected,
				Next:     calls.ParticipantPresenceConnected,
			},
		},
		{
			name:     "connected to left",
			previous: calls.ParticipantPresenceConnected,
			kind:     calls.ParticipantPresenceLeft,
			want: calls.ParticipantPresenceTransition{
				Previous:    calls.ParticipantPresenceConnected,
				Next:        calls.ParticipantPresenceLeft,
				Persist:     true,
				PublishLeft: true,
			},
		},
		{
			name:     "left from unknown",
			previous: calls.ParticipantPresenceUnknown,
			kind:     calls.ParticipantPresenceLeft,
			want: calls.ParticipantPresenceTransition{
				Previous: calls.ParticipantPresenceUnknown,
				Next:     calls.ParticipantPresenceUnknown,
			},
		},
		{
			name:     "duplicate left",
			previous: calls.ParticipantPresenceLeft,
			kind:     calls.ParticipantPresenceLeft,
			want: calls.ParticipantPresenceTransition{
				Previous: calls.ParticipantPresenceLeft,
				Next:     calls.ParticipantPresenceLeft,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			transition, err := calls.ReduceConfirmedParticipantObservation(testCase.previous, calls.ConfirmedParticipantObservation{
				CallID: confirmedPresenceCallID,
				UserID: confirmedPresenceUserID,
				Kind:   testCase.kind,
			})

			require.NoError(t, err)
			assert.Equal(t, testCase.want, transition)
		})
	}
}

var (
	confirmedPresenceCallID = uuid.MustParse("a1111111-1111-4111-8111-111111111111")
	confirmedPresenceUserID = uuid.MustParse("b2222222-2222-4222-8222-222222222222")
)
