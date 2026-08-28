//go:build hubtest

package calls_test

import (
	"encoding"
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/calls"
)

var (
	confirmedConnectionStateChangeCallID = uuid.MustParse("91111111-1111-4111-8111-111111111111")
	confirmedConnectionStateChangeUserID = uuid.MustParse("92222222-2222-4222-8222-222222222222")
	confirmedConnectionStateChangeEpoch  = uuid.MustParse("93333333-3333-4333-8333-333333333333")
)

func TestBuildConfirmedParticipantConnectionStateChangeBuildsEachPersistClass(t *testing.T) {
	key := confirmedConnectionStateChangeKey()
	tests := []struct {
		name       string
		transition calls.ConfirmedParticipantConnectionTransition
		previous   calls.ParticipantPresence
		next       calls.ParticipantPresence
		before     uint32
		after      uint32
		joined     bool
		left       bool
	}{
		{
			name: "unknown to connected publishes joined at zero to one",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceConnected,
				PreviousUserConnectedCount: 0,
				NextUserConnectedCount:     1,
				PersistConnection:          true,
				PersistUserRefCount:        true,
				PublishJoined:              true,
			},
			previous: calls.ParticipantPresenceUnknown,
			next:     calls.ParticipantPresenceConnected,
			before:   0,
			after:    1,
			joined:   true,
		},
		{
			name: "unknown to connected retains joined boundary when another connection exists",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceConnected,
				PreviousUserConnectedCount: 3,
				NextUserConnectedCount:     4,
				PersistConnection:          true,
				PersistUserRefCount:        true,
			},
			previous: calls.ParticipantPresenceUnknown,
			next:     calls.ParticipantPresenceConnected,
			before:   3,
			after:    4,
		},
		{
			name: "connected to left publishes left at one to zero",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceConnected,
				NextState:                  calls.ParticipantPresenceLeft,
				PreviousUserConnectedCount: 1,
				NextUserConnectedCount:     0,
				PersistConnection:          true,
				PersistUserRefCount:        true,
				PublishLeft:                true,
			},
			previous: calls.ParticipantPresenceConnected,
			next:     calls.ParticipantPresenceLeft,
			before:   1,
			after:    0,
			left:     true,
		},
		{
			name: "connected to left retains left boundary while another connection remains",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceConnected,
				NextState:                  calls.ParticipantPresenceLeft,
				PreviousUserConnectedCount: 4,
				NextUserConnectedCount:     3,
				PersistConnection:          true,
				PersistUserRefCount:        true,
			},
			previous: calls.ParticipantPresenceConnected,
			next:     calls.ParticipantPresenceLeft,
			before:   4,
			after:    3,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			change, shouldApply, err := calls.BuildConfirmedParticipantConnectionStateChange(testCase.transition)

			require.NoError(t, err)
			assert.True(t, shouldApply)
			assert.Equal(t, key, change.Key())
			assert.Equal(t, testCase.previous, change.PreviousState())
			assert.Equal(t, testCase.next, change.NextState())
			assert.Equal(t, testCase.before, change.PreviousUserConnectedCount())
			assert.Equal(t, testCase.after, change.NextUserConnectedCount())
			assert.Equal(t, testCase.joined, change.PublishJoined())
			assert.Equal(t, testCase.left, change.PublishLeft())
		})
	}
}

func TestBuildConfirmedParticipantConnectionStateChangeAcceptsDescriptiveNoOpsAndReconciliation(t *testing.T) {
	key := confirmedConnectionStateChangeKey()
	tests := []struct {
		name       string
		transition calls.ConfirmedParticipantConnectionTransition
	}{
		{
			name: "unknown zero no op",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceUnknown,
				PreviousUserConnectedCount: 0,
				NextUserConnectedCount:     0,
			},
		},
		{
			name: "unknown nonzero ended-call-compatible no op",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceUnknown,
				PreviousUserConnectedCount: 6,
				NextUserConnectedCount:     6,
			},
		},
		{
			name: "connected positive no op",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceConnected,
				NextState:                  calls.ParticipantPresenceConnected,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     2,
			},
		},
		{
			name: "left zero no op",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:           key,
				PreviousState: calls.ParticipantPresenceLeft,
				NextState:     calls.ParticipantPresenceLeft,
			},
		},
		{
			name: "left nonzero ended-call-compatible no op",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceLeft,
				NextState:                  calls.ParticipantPresenceLeft,
				PreviousUserConnectedCount: 5,
				NextUserConnectedCount:     5,
			},
		},
		{
			name: "unknown reconciliation",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceUnknown,
				PreviousUserConnectedCount: 3,
				NextUserConnectedCount:     3,
				RequestReconcile:           true,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			change, shouldApply, err := calls.BuildConfirmedParticipantConnectionStateChange(testCase.transition)

			require.NoError(t, err)
			assert.False(t, shouldApply)
			assert.Zero(t, change)
			assert.Equal(t, calls.ConfirmedParticipantConnectionKey{}, change.Key())
			assert.Equal(t, calls.ParticipantPresenceUnknown, change.PreviousState())
			assert.Equal(t, calls.ParticipantPresenceUnknown, change.NextState())
			assert.Zero(t, change.PreviousUserConnectedCount())
			assert.Zero(t, change.NextUserConnectedCount())
			assert.False(t, change.PublishJoined())
			assert.False(t, change.PublishLeft())
		})
	}
}

func TestConfirmedParticipantConnectionStateChangeAvoidsAccidentalOutputInterfaces(t *testing.T) {
	change, shouldApply, err := calls.BuildConfirmedParticipantConnectionStateChange(
		calls.ConfirmedParticipantConnectionTransition{
			Key:                        confirmedConnectionStateChangeKey(),
			PreviousState:              calls.ParticipantPresenceUnknown,
			NextState:                  calls.ParticipantPresenceConnected,
			PreviousUserConnectedCount: 0,
			NextUserConnectedCount:     1,
			PersistConnection:          true,
			PersistUserRefCount:        true,
			PublishJoined:              true,
		},
	)

	require.NoError(t, err)
	require.True(t, shouldApply)

	// This avoids accidental output interfaces; callers must still not log
	// exposed key values because it is not a secrecy or redaction guarantee.
	_, valueStringer := any(change).(fmt.Stringer)
	_, pointerStringer := any(&change).(fmt.Stringer)
	_, valueTextMarshaler := any(change).(encoding.TextMarshaler)
	_, pointerTextMarshaler := any(&change).(encoding.TextMarshaler)
	_, valueTextUnmarshaler := any(change).(encoding.TextUnmarshaler)
	_, pointerTextUnmarshaler := any(&change).(encoding.TextUnmarshaler)
	_, valueJSONMarshaler := any(change).(json.Marshaler)
	_, pointerJSONMarshaler := any(&change).(json.Marshaler)
	_, valueJSONUnmarshaler := any(change).(json.Unmarshaler)
	_, pointerJSONUnmarshaler := any(&change).(json.Unmarshaler)

	assert.False(t, valueStringer)
	assert.False(t, pointerStringer)
	assert.False(t, valueTextMarshaler)
	assert.False(t, pointerTextMarshaler)
	assert.False(t, valueTextUnmarshaler)
	assert.False(t, pointerTextUnmarshaler)
	assert.False(t, valueJSONMarshaler)
	assert.False(t, pointerJSONMarshaler)
	assert.False(t, valueJSONUnmarshaler)
	assert.False(t, pointerJSONUnmarshaler)
}

func TestBuildConfirmedParticipantConnectionStateChangeRejectsMalformedTransitions(t *testing.T) {
	key := confirmedConnectionStateChangeKey()

	tests := []struct {
		name       string
		transition calls.ConfirmedParticipantConnectionTransition
	}{
		{
			name: "zero key even when transition looks like a no op",
			transition: calls.ConfirmedParticipantConnectionTransition{
				PreviousState: calls.ParticipantPresenceUnknown,
				NextState:     calls.ParticipantPresenceUnknown,
			},
		},
		{
			name: "partial key",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key: calls.ConfirmedParticipantConnectionKey{
					CallID: confirmedConnectionStateChangeCallID,
					UserID: confirmedConnectionStateChangeUserID,
				},
				PreviousState: calls.ParticipantPresenceUnknown,
				NextState:     calls.ParticipantPresenceUnknown,
			},
		},
		{
			name: "unsupported previous state",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:           key,
				PreviousState: calls.ParticipantPresence(99),
				NextState:     calls.ParticipantPresenceUnknown,
			},
		},
		{
			name: "unsupported next state",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:           key,
				PreviousState: calls.ParticipantPresenceUnknown,
				NextState:     calls.ParticipantPresence(99),
			},
		},
		{
			name: "previous connected without count",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:           key,
				PreviousState: calls.ParticipantPresenceConnected,
				NextState:     calls.ParticipantPresenceConnected,
			},
		},
		{
			name: "next connected without count",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceConnected,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     0,
			},
		},
		{
			name: "unknown connected counter overflow",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceConnected,
				PreviousUserConnectedCount: math.MaxUint32,
				NextUserConnectedCount:     0,
				PersistConnection:          true,
				PersistUserRefCount:        true,
			},
		},
		{
			name: "partial persistence connection only",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceConnected,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     3,
				PersistConnection:          true,
			},
		},
		{
			name: "partial persistence refcount only",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceConnected,
				NextState:                  calls.ParticipantPresenceLeft,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     1,
				PersistUserRefCount:        true,
			},
		},
		{
			name: "changed count does not match unknown connected transition",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceConnected,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     4,
				PersistConnection:          true,
				PersistUserRefCount:        true,
			},
		},
		{
			name: "changed count does not match connected left transition",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceConnected,
				NextState:                  calls.ParticipantPresenceLeft,
				PreviousUserConnectedCount: 3,
				NextUserConnectedCount:     3,
				PersistConnection:          true,
				PersistUserRefCount:        true,
			},
		},
		{
			name: "unknown connected has wrong joined boundary",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceConnected,
				PreviousUserConnectedCount: 0,
				NextUserConnectedCount:     1,
				PersistConnection:          true,
				PersistUserRefCount:        true,
			},
		},
		{
			name: "unknown connected publishes joined beyond boundary",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceConnected,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     3,
				PersistConnection:          true,
				PersistUserRefCount:        true,
				PublishJoined:              true,
			},
		},
		{
			name: "connected left has wrong left boundary",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceConnected,
				NextState:                  calls.ParticipantPresenceLeft,
				PreviousUserConnectedCount: 1,
				NextUserConnectedCount:     0,
				PersistConnection:          true,
				PersistUserRefCount:        true,
			},
		},
		{
			name: "connected left publishes left beyond boundary",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceConnected,
				NextState:                  calls.ParticipantPresenceLeft,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     1,
				PersistConnection:          true,
				PersistUserRefCount:        true,
				PublishLeft:                true,
			},
		},
		{
			name: "both publication flags",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceConnected,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     3,
				PersistConnection:          true,
				PersistUserRefCount:        true,
				PublishJoined:              true,
				PublishLeft:                true,
			},
		},
		{
			name: "persistence on descriptive no op",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceUnknown,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     2,
				PersistConnection:          true,
				PersistUserRefCount:        true,
			},
		},
		{
			name: "publication on descriptive no op",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceLeft,
				NextState:                  calls.ParticipantPresenceLeft,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     2,
				PublishJoined:              true,
			},
		},
		{
			name: "persistence on reconciliation",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceUnknown,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     2,
				PersistConnection:          true,
				PersistUserRefCount:        true,
				RequestReconcile:           true,
			},
		},
		{
			name: "publication on reconciliation",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceUnknown,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     2,
				PublishLeft:                true,
				RequestReconcile:           true,
			},
		},
		{
			name: "malformed reconciliation state",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceLeft,
				NextState:                  calls.ParticipantPresenceLeft,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     2,
				RequestReconcile:           true,
			},
		},
		{
			name: "malformed reconciliation count change",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceUnknown,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     1,
				RequestReconcile:           true,
			},
		},
		{
			name: "unknown left marked persistent",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceLeft,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     1,
				PersistConnection:          true,
				PersistUserRefCount:        true,
			},
		},
		{
			name: "left reconnect is terminal for one epoch",
			transition: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceLeft,
				NextState:                  calls.ParticipantPresenceConnected,
				PreviousUserConnectedCount: 1,
				NextUserConnectedCount:     2,
				PersistConnection:          true,
				PersistUserRefCount:        true,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			change, shouldApply, err := calls.BuildConfirmedParticipantConnectionStateChange(testCase.transition)

			assert.False(t, shouldApply)
			assert.Zero(t, change)
			assert.Equal(t, calls.ConfirmedParticipantConnectionKey{}, change.Key())
			assert.Equal(t, calls.ParticipantPresenceUnknown, change.PreviousState())
			assert.Equal(t, calls.ParticipantPresenceUnknown, change.NextState())
			assert.Zero(t, change.PreviousUserConnectedCount())
			assert.Zero(t, change.NextUserConnectedCount())
			assert.False(t, change.PublishJoined())
			assert.False(t, change.PublishLeft())
			require.ErrorIs(t, err, calls.ErrConfirmedParticipantConnectionStateChangeInvalid)
			assert.Equal(t, calls.ErrConfirmedParticipantConnectionStateChangeInvalid, err)
		})
	}
}

func confirmedConnectionStateChangeKey() calls.ConfirmedParticipantConnectionKey {
	return calls.ConfirmedParticipantConnectionKey{
		CallID: confirmedConnectionStateChangeCallID,
		UserID: confirmedConnectionStateChangeUserID,
		Epoch:  confirmedConnectionStateChangeEpoch,
	}
}
