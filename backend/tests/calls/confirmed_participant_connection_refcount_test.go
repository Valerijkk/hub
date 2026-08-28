//go:build hubtest

package calls_test

import (
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/calls"
)

var (
	confirmedConnectionRefcountCallID = uuid.MustParse("a1111111-1111-4111-8111-111111111111")
	confirmedConnectionRefcountUserID = uuid.MustParse("b2222222-2222-4222-8222-222222222222")
	confirmedConnectionRefcountEpoch  = uuid.MustParse("c3333333-3333-4333-8333-333333333333")
)

func TestReduceConfirmedParticipantConnectionRejectsInvalidInput(t *testing.T) {
	validSnapshot := confirmedConnectionRefcountSnapshot(
		calls.ParticipantPresenceUnknown,
		0,
		false,
	)
	validObservation := confirmedConnectionRefcountObservation(calls.ParticipantPresenceConnected)

	tests := []struct {
		name        string
		snapshot    calls.ConfirmedParticipantConnectionSnapshot
		observation calls.ConfirmedParticipantObservation
	}{
		{
			name: "zero key",
			snapshot: calls.ConfirmedParticipantConnectionSnapshot{
				State:              calls.ParticipantPresenceUnknown,
				UserConnectedCount: 0,
			},
			observation: validObservation,
		},
		{
			name: "partial key",
			snapshot: calls.ConfirmedParticipantConnectionSnapshot{
				Key: calls.ConfirmedParticipantConnectionKey{
					CallID: confirmedConnectionRefcountCallID,
					UserID: confirmedConnectionRefcountUserID,
				},
				State:              calls.ParticipantPresenceUnknown,
				UserConnectedCount: 0,
			},
			observation: validObservation,
		},
		{
			name:     "zero observation call ID",
			snapshot: validSnapshot,
			observation: calls.ConfirmedParticipantObservation{
				UserID: confirmedConnectionRefcountUserID,
				Kind:   calls.ParticipantPresenceConnected,
			},
		},
		{
			name:     "zero observation user ID",
			snapshot: validSnapshot,
			observation: calls.ConfirmedParticipantObservation{
				CallID: confirmedConnectionRefcountCallID,
				Kind:   calls.ParticipantPresenceConnected,
			},
		},
		{
			name:     "observation call mismatch",
			snapshot: validSnapshot,
			observation: calls.ConfirmedParticipantObservation{
				CallID: uuid.MustParse("d4444444-4444-4444-8444-444444444444"),
				UserID: confirmedConnectionRefcountUserID,
				Kind:   calls.ParticipantPresenceConnected,
			},
		},
		{
			name:     "observation user mismatch",
			snapshot: validSnapshot,
			observation: calls.ConfirmedParticipantObservation{
				CallID: confirmedConnectionRefcountCallID,
				UserID: uuid.MustParse("e5555555-5555-4555-8555-555555555555"),
				Kind:   calls.ParticipantPresenceConnected,
			},
		},
		{
			name:     "zero observation kind",
			snapshot: validSnapshot,
			observation: calls.ConfirmedParticipantObservation{
				CallID: confirmedConnectionRefcountCallID,
				UserID: confirmedConnectionRefcountUserID,
			},
		},
		{
			name:     "unsupported observation kind",
			snapshot: validSnapshot,
			observation: calls.ConfirmedParticipantObservation{
				CallID: confirmedConnectionRefcountCallID,
				UserID: confirmedConnectionRefcountUserID,
				Kind:   calls.ParticipantPresence(99),
			},
		},
		{
			name: "unsupported prior state",
			snapshot: calls.ConfirmedParticipantConnectionSnapshot{
				Key:                validSnapshot.Key,
				State:              calls.ParticipantPresence(99),
				UserConnectedCount: 0,
			},
			observation: validObservation,
		},
		{
			name: "connected state without aggregate count",
			snapshot: confirmedConnectionRefcountSnapshot(
				calls.ParticipantPresenceConnected,
				0,
				false,
			),
			observation: validObservation,
		},
		{
			name: "active unknown connected overflow",
			snapshot: confirmedConnectionRefcountSnapshot(
				calls.ParticipantPresenceUnknown,
				math.MaxUint32,
				false,
			),
			observation: validObservation,
		},
		{
			name: "same epoch left reconnect",
			snapshot: confirmedConnectionRefcountSnapshot(
				calls.ParticipantPresenceLeft,
				1,
				false,
			),
			observation: validObservation,
		},
		{
			name: "call ended does not skip structural validation",
			snapshot: calls.ConfirmedParticipantConnectionSnapshot{
				State:              calls.ParticipantPresenceUnknown,
				UserConnectedCount: 0,
				CallEnded:          true,
			},
			observation: validObservation,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			transition, err := calls.ReduceConfirmedParticipantConnection(
				testCase.snapshot,
				testCase.observation,
			)

			assert.Equal(t, calls.ConfirmedParticipantConnectionTransition{}, transition)
			require.ErrorIs(t, err, calls.ErrConfirmedParticipantConnectionTransitionInvalid)
			assert.Equal(t, calls.ErrConfirmedParticipantConnectionTransitionInvalid, err)
		})
	}
}

func TestReduceConfirmedParticipantConnectionActiveTransitions(t *testing.T) {
	key := confirmedConnectionRefcountKey()
	tests := []struct {
		name        string
		snapshot    calls.ConfirmedParticipantConnectionSnapshot
		observation calls.ConfirmedParticipantObservation
		want        calls.ConfirmedParticipantConnectionTransition
	}{
		{
			name: "unknown connected publishes joined at zero to one",
			snapshot: confirmedConnectionRefcountSnapshot(
				calls.ParticipantPresenceUnknown,
				0,
				false,
			),
			observation: confirmedConnectionRefcountObservation(calls.ParticipantPresenceConnected),
			want: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceConnected,
				PreviousUserConnectedCount: 0,
				NextUserConnectedCount:     1,
				PersistConnection:          true,
				PersistUserRefCount:        true,
				PublishJoined:              true,
			},
		},
		{
			name: "unknown connected increments without duplicate joined",
			snapshot: confirmedConnectionRefcountSnapshot(
				calls.ParticipantPresenceUnknown,
				2,
				false,
			),
			observation: confirmedConnectionRefcountObservation(calls.ParticipantPresenceConnected),
			want: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceConnected,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     3,
				PersistConnection:          true,
				PersistUserRefCount:        true,
			},
		},
		{
			name: "duplicate connected is a no op",
			snapshot: confirmedConnectionRefcountSnapshot(
				calls.ParticipantPresenceConnected,
				2,
				false,
			),
			observation: confirmedConnectionRefcountObservation(calls.ParticipantPresenceConnected),
			want: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceConnected,
				NextState:                  calls.ParticipantPresenceConnected,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     2,
			},
		},
		{
			name: "connected left decrements without left while another connection remains",
			snapshot: confirmedConnectionRefcountSnapshot(
				calls.ParticipantPresenceConnected,
				2,
				false,
			),
			observation: confirmedConnectionRefcountObservation(calls.ParticipantPresenceLeft),
			want: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceConnected,
				NextState:                  calls.ParticipantPresenceLeft,
				PreviousUserConnectedCount: 2,
				NextUserConnectedCount:     1,
				PersistConnection:          true,
				PersistUserRefCount:        true,
			},
		},
		{
			name: "connected left publishes left at one to zero",
			snapshot: confirmedConnectionRefcountSnapshot(
				calls.ParticipantPresenceConnected,
				1,
				false,
			),
			observation: confirmedConnectionRefcountObservation(calls.ParticipantPresenceLeft),
			want: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceConnected,
				NextState:                  calls.ParticipantPresenceLeft,
				PreviousUserConnectedCount: 1,
				NextUserConnectedCount:     0,
				PersistConnection:          true,
				PersistUserRefCount:        true,
				PublishLeft:                true,
			},
		},
		{
			name: "unknown left requests reconcile without persistence",
			snapshot: confirmedConnectionRefcountSnapshot(
				calls.ParticipantPresenceUnknown,
				3,
				false,
			),
			observation: confirmedConnectionRefcountObservation(calls.ParticipantPresenceLeft),
			want: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceUnknown,
				NextState:                  calls.ParticipantPresenceUnknown,
				PreviousUserConnectedCount: 3,
				NextUserConnectedCount:     3,
				RequestReconcile:           true,
			},
		},
		{
			name: "duplicate left is a no op",
			snapshot: confirmedConnectionRefcountSnapshot(
				calls.ParticipantPresenceLeft,
				0,
				false,
			),
			observation: confirmedConnectionRefcountObservation(calls.ParticipantPresenceLeft),
			want: calls.ConfirmedParticipantConnectionTransition{
				Key:                        key,
				PreviousState:              calls.ParticipantPresenceLeft,
				NextState:                  calls.ParticipantPresenceLeft,
				PreviousUserConnectedCount: 0,
				NextUserConnectedCount:     0,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			transition, err := calls.ReduceConfirmedParticipantConnection(
				testCase.snapshot,
				testCase.observation,
			)

			require.NoError(t, err)
			assert.Equal(t, testCase.want, transition)
		})
	}
}

func TestReduceConfirmedParticipantConnectionCallEndedIsDescriptiveNoOp(t *testing.T) {
	tests := []struct {
		name  string
		state calls.ParticipantPresence
		count uint32
	}{
		{name: "unknown zero", state: calls.ParticipantPresenceUnknown, count: 0},
		{name: "unknown positive", state: calls.ParticipantPresenceUnknown, count: 3},
		{name: "unknown max", state: calls.ParticipantPresenceUnknown, count: math.MaxUint32},
		{name: "connected one", state: calls.ParticipantPresenceConnected, count: 1},
		{name: "connected positive", state: calls.ParticipantPresenceConnected, count: 3},
		{name: "connected max", state: calls.ParticipantPresenceConnected, count: math.MaxUint32},
		{name: "left zero", state: calls.ParticipantPresenceLeft, count: 0},
		{name: "left positive", state: calls.ParticipantPresenceLeft, count: 3},
		{name: "left max", state: calls.ParticipantPresenceLeft, count: math.MaxUint32},
	}

	for _, testCase := range tests {
		for _, kind := range []calls.ParticipantPresence{
			calls.ParticipantPresenceConnected,
			calls.ParticipantPresenceLeft,
		} {
			t.Run(testCase.name+" "+confirmedConnectionRefcountPresenceName(kind), func(t *testing.T) {
				snapshot := confirmedConnectionRefcountSnapshot(testCase.state, testCase.count, true)
				transition, err := calls.ReduceConfirmedParticipantConnection(
					snapshot,
					confirmedConnectionRefcountObservation(kind),
				)

				require.NoError(t, err)
				assert.Equal(t, calls.ConfirmedParticipantConnectionTransition{
					Key:                        snapshot.Key,
					PreviousState:              snapshot.State,
					NextState:                  snapshot.State,
					PreviousUserConnectedCount: snapshot.UserConnectedCount,
					NextUserConnectedCount:     snapshot.UserConnectedCount,
				}, transition)
			})
		}
	}
}

func confirmedConnectionRefcountKey() calls.ConfirmedParticipantConnectionKey {
	return calls.ConfirmedParticipantConnectionKey{
		CallID: confirmedConnectionRefcountCallID,
		UserID: confirmedConnectionRefcountUserID,
		Epoch:  confirmedConnectionRefcountEpoch,
	}
}

func confirmedConnectionRefcountSnapshot(
	state calls.ParticipantPresence,
	count uint32,
	callEnded bool,
) calls.ConfirmedParticipantConnectionSnapshot {
	return calls.ConfirmedParticipantConnectionSnapshot{
		Key:                confirmedConnectionRefcountKey(),
		State:              state,
		UserConnectedCount: count,
		CallEnded:          callEnded,
	}
}

func confirmedConnectionRefcountObservation(kind calls.ParticipantPresence) calls.ConfirmedParticipantObservation {
	return calls.ConfirmedParticipantObservation{
		CallID: confirmedConnectionRefcountCallID,
		UserID: confirmedConnectionRefcountUserID,
		Kind:   kind,
	}
}

func confirmedConnectionRefcountPresenceName(presence calls.ParticipantPresence) string {
	switch presence {
	case calls.ParticipantPresenceConnected:
		return "connected observation"
	case calls.ParticipantPresenceLeft:
		return "left observation"
	default:
		return "unsupported observation"
	}
}
