//go:build hubtest

package calls_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/calls"
)

func TestNormalizeTrustedProviderParticipantConnectionMapsKnownKinds(t *testing.T) {
	tests := []struct {
		name string
		kind calls.TrustedProviderParticipantConnectionKind
		want calls.ParticipantPresence
	}{
		{
			name: "connected",
			kind: calls.TrustedProviderParticipantConnected,
			want: calls.ParticipantPresenceConnected,
		},
		{
			name: "disconnected",
			kind: calls.TrustedProviderParticipantDisconnected,
			want: calls.ParticipantPresenceLeft,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			observation, err := calls.NormalizeTrustedProviderParticipantConnection(calls.TrustedProviderParticipantConnection{
				CallID: providerPresenceCallID,
				UserID: providerPresenceUserID,
				Kind:   testCase.kind,
			})

			require.NoError(t, err)
			assert.Equal(t, calls.ConfirmedParticipantObservation{
				CallID: providerPresenceCallID,
				UserID: providerPresenceUserID,
				Kind:   testCase.want,
			}, observation)
		})
	}
}

func TestNormalizeTrustedProviderParticipantConnectionRejectsInvalidFacts(t *testing.T) {
	tests := []struct {
		name string
		fact calls.TrustedProviderParticipantConnection
	}{
		{
			name: "zero call ID",
			fact: calls.TrustedProviderParticipantConnection{
				UserID: providerPresenceUserID,
				Kind:   calls.TrustedProviderParticipantConnected,
			},
		},
		{
			name: "zero user ID",
			fact: calls.TrustedProviderParticipantConnection{
				CallID: providerPresenceCallID,
				Kind:   calls.TrustedProviderParticipantConnected,
			},
		},
		{
			name: "zero kind",
			fact: calls.TrustedProviderParticipantConnection{
				CallID: providerPresenceCallID,
				UserID: providerPresenceUserID,
			},
		},
		{
			name: "unknown kind",
			fact: calls.TrustedProviderParticipantConnection{
				CallID: providerPresenceCallID,
				UserID: providerPresenceUserID,
				Kind:   calls.TrustedProviderParticipantConnectionKind(99),
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			observation, err := calls.NormalizeTrustedProviderParticipantConnection(testCase.fact)

			assert.Equal(t, calls.ConfirmedParticipantObservation{}, observation)
			require.ErrorIs(t, err, calls.ErrConfirmedParticipantObservationInvalid)
		})
	}
}

var (
	providerPresenceCallID = uuid.MustParse("c3333333-3333-4333-8333-333333333333")
	providerPresenceUserID = uuid.MustParse("d4444444-4444-4444-8444-444444444444")
)
