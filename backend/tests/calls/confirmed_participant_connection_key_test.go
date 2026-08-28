//go:build hubtest

package calls_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/calls"
)

func TestConfirmedParticipantConnectionKeyCanonicalIsDeterministicAndOpaque(t *testing.T) {
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	epoch := uuid.MustParse("33333333-3333-4333-8333-333333333333")

	key, err := calls.NewConfirmedParticipantConnectionKey(callID, userID, epoch)
	require.NoError(t, err)
	sameKey, err := calls.NewConfirmedParticipantConnectionKey(callID, userID, epoch)
	require.NoError(t, err)

	assert.True(t, key.IsValid())
	assert.Equal(t, key, sameKey)
	assert.Equal(t, "hub.call-connection.v1/11111111-1111-4111-8111-111111111111/22222222-2222-4222-8222-222222222222/33333333-3333-4333-8333-333333333333", key.Canonical())
	assert.Equal(t, key.Canonical(), sameKey.Canonical())
	_, valueImplementsStringer := any(key).(fmt.Stringer)
	_, pointerImplementsStringer := any(&key).(fmt.Stringer)
	assert.False(t, valueImplementsStringer)
	assert.False(t, pointerImplementsStringer)
}

func TestConfirmedParticipantConnectionKeyChangesWhenAnyComponentChanges(t *testing.T) {
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	epoch := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	base, err := calls.NewConfirmedParticipantConnectionKey(callID, userID, epoch)
	require.NoError(t, err)

	tests := []struct {
		name   string
		callID uuid.UUID
		userID uuid.UUID
		epoch  uuid.UUID
	}{
		{
			name:   "call",
			callID: uuid.MustParse("44444444-4444-4444-8444-444444444444"),
			userID: userID,
			epoch:  epoch,
		},
		{
			name:   "user",
			callID: callID,
			userID: uuid.MustParse("55555555-5555-4555-8555-555555555555"),
			epoch:  epoch,
		},
		{
			name:   "epoch",
			callID: callID,
			userID: userID,
			epoch:  uuid.MustParse("66666666-6666-4666-8666-666666666666"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := calls.NewConfirmedParticipantConnectionKey(tt.callID, tt.userID, tt.epoch)

			require.NoError(t, err)
			assert.NotEqual(t, base, key)
			assert.NotEqual(t, base.Canonical(), key.Canonical())
		})
	}
}

func TestNewConfirmedParticipantConnectionKeyRejectsEachZeroComponent(t *testing.T) {
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	epoch := uuid.MustParse("33333333-3333-4333-8333-333333333333")

	tests := []struct {
		name   string
		callID uuid.UUID
		userID uuid.UUID
		epoch  uuid.UUID
	}{
		{name: "call", userID: userID, epoch: epoch},
		{name: "user", callID: callID, epoch: epoch},
		{name: "epoch", callID: callID, userID: userID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := calls.NewConfirmedParticipantConnectionKey(tt.callID, tt.userID, tt.epoch)

			assert.ErrorIs(t, err, calls.ErrConfirmedParticipantConnectionKeyInvalid)
			assert.Equal(t, calls.ErrConfirmedParticipantConnectionKeyInvalid, err)
			assert.Equal(t, calls.ConfirmedParticipantConnectionKey{}, key)
		})
	}
}

func TestConfirmedParticipantConnectionKeyRejectsZeroAndPartialForgedValues(t *testing.T) {
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	epoch := uuid.MustParse("33333333-3333-4333-8333-333333333333")

	tests := []struct {
		name string
		key  calls.ConfirmedParticipantConnectionKey
	}{
		{name: "zero"},
		{name: "call only", key: calls.ConfirmedParticipantConnectionKey{CallID: callID}},
		{name: "call and user", key: calls.ConfirmedParticipantConnectionKey{CallID: callID, UserID: userID}},
		{name: "call and epoch", key: calls.ConfirmedParticipantConnectionKey{CallID: callID, Epoch: epoch}},
		{name: "user and epoch", key: calls.ConfirmedParticipantConnectionKey{UserID: userID, Epoch: epoch}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.False(t, tt.key.IsValid())
			assert.Empty(t, tt.key.Canonical())
		})
	}
}
