package calls

import (
	"context"
	"errors"
)

// ErrConfirmedParticipantConnectionStateChangeInvalid marks a malformed
// confirmed participant connection state change.
var ErrConfirmedParticipantConnectionStateChangeInvalid = errors.New(
	"invalid confirmed participant connection state change",
)

// confirmedParticipantConnectionStateChange is the immutable compare-and-swap
// payload derived from one valid persistence transition.
type confirmedParticipantConnectionStateChange struct {
	key                        ConfirmedParticipantConnectionKey
	previousState              ParticipantPresence
	nextState                  ParticipantPresence
	previousUserConnectedCount uint32
	nextUserConnectedCount     uint32
	publishJoined              bool
	publishLeft                bool
}

// BuildConfirmedParticipantConnectionStateChange converts one trusted,
// reducer-derived persistence transition into an immutable CAS payload.
func BuildConfirmedParticipantConnectionStateChange(
	transition ConfirmedParticipantConnectionTransition,
) (change confirmedParticipantConnectionStateChange, shouldApply bool, err error) {
	if !transition.Key.IsValid() ||
		!isParticipantPresence(transition.PreviousState) ||
		!isParticipantPresence(transition.NextState) ||
		(transition.PreviousState == ParticipantPresenceConnected &&
			transition.PreviousUserConnectedCount == 0) ||
		(transition.NextState == ParticipantPresenceConnected &&
			transition.NextUserConnectedCount == 0) ||
		(transition.PublishJoined && transition.PublishLeft) {
		return confirmedParticipantConnectionStateChange{}, false, ErrConfirmedParticipantConnectionStateChangeInvalid
	}

	if transition.PersistConnection &&
		transition.PersistUserRefCount &&
		transition.PreviousState == ParticipantPresenceUnknown &&
		transition.NextState == ParticipantPresenceConnected &&
		!transition.RequestReconcile &&
		!transition.PublishLeft &&
		transition.PreviousUserConnectedCount != ^uint32(0) &&
		transition.NextUserConnectedCount == transition.PreviousUserConnectedCount+1 &&
		transition.PublishJoined == (transition.PreviousUserConnectedCount == 0) {
		return newConfirmedParticipantConnectionStateChange(transition), true, nil
	}

	if transition.PersistConnection &&
		transition.PersistUserRefCount &&
		transition.PreviousState == ParticipantPresenceConnected &&
		transition.NextState == ParticipantPresenceLeft &&
		!transition.RequestReconcile &&
		!transition.PublishJoined &&
		transition.PreviousUserConnectedCount > 0 &&
		transition.NextUserConnectedCount == transition.PreviousUserConnectedCount-1 &&
		transition.PublishLeft == (transition.PreviousUserConnectedCount == 1) {
		return newConfirmedParticipantConnectionStateChange(transition), true, nil
	}

	if !transition.PersistConnection &&
		!transition.PersistUserRefCount &&
		!transition.PublishJoined &&
		!transition.PublishLeft {
		if !transition.RequestReconcile &&
			transition.PreviousState == transition.NextState &&
			transition.PreviousUserConnectedCount == transition.NextUserConnectedCount {
			return confirmedParticipantConnectionStateChange{}, false, nil
		}
		if transition.RequestReconcile &&
			transition.PreviousState == ParticipantPresenceUnknown &&
			transition.NextState == ParticipantPresenceUnknown &&
			transition.PreviousUserConnectedCount == transition.NextUserConnectedCount {
			return confirmedParticipantConnectionStateChange{}, false, nil
		}
	}

	return confirmedParticipantConnectionStateChange{}, false, ErrConfirmedParticipantConnectionStateChangeInvalid
}

func newConfirmedParticipantConnectionStateChange(
	transition ConfirmedParticipantConnectionTransition,
) confirmedParticipantConnectionStateChange {
	return confirmedParticipantConnectionStateChange{
		key:                        transition.Key,
		previousState:              transition.PreviousState,
		nextState:                  transition.NextState,
		previousUserConnectedCount: transition.PreviousUserConnectedCount,
		nextUserConnectedCount:     transition.NextUserConnectedCount,
		publishJoined:              transition.PublishJoined,
		publishLeft:                transition.PublishLeft,
	}
}

// Key returns the immutable Hub connection key for this CAS payload.
func (change confirmedParticipantConnectionStateChange) Key() ConfirmedParticipantConnectionKey {
	return change.key
}

// PreviousState returns the state that a future CAS must compare.
func (change confirmedParticipantConnectionStateChange) PreviousState() ParticipantPresence {
	return change.previousState
}

// NextState returns the state that a future CAS must update on a match.
func (change confirmedParticipantConnectionStateChange) NextState() ParticipantPresence {
	return change.nextState
}

// PreviousUserConnectedCount returns the aggregate count a future CAS must compare.
func (change confirmedParticipantConnectionStateChange) PreviousUserConnectedCount() uint32 {
	return change.previousUserConnectedCount
}

// NextUserConnectedCount returns the aggregate count a future CAS must update on a match.
func (change confirmedParticipantConnectionStateChange) NextUserConnectedCount() uint32 {
	return change.nextUserConnectedCount
}

// PublishJoined reports whether a future caller may publish only after a successful CAS.
func (change confirmedParticipantConnectionStateChange) PublishJoined() bool {
	return change.publishJoined
}

// PublishLeft reports whether a future caller may publish only after a successful CAS.
func (change confirmedParticipantConnectionStateChange) PublishLeft() bool {
	return change.publishLeft
}

// ConfirmedParticipantConnectionStateChangeCASResult reports the result of a
// future atomic connection-and-refcount compare-and-swap.
type ConfirmedParticipantConnectionStateChangeCASResult uint8

const (
	ConfirmedParticipantConnectionStateChangeCASApplied ConfirmedParticipantConnectionStateChangeCASResult = iota + 1
	ConfirmedParticipantConnectionStateChangeCASConflict
)

// confirmedParticipantConnectionStateChangeStore atomically compares the
// connection state and aggregate connected count, then changes both together.
// An applied result is legal only after both updates commit. A conflict or
// non-applied outcome performs neither mutation nor publication.
type confirmedParticipantConnectionStateChangeStore interface {
	// CompareAndSwapConfirmedParticipantConnectionState compares both previous
	// values and commits both next values together without publication.
	CompareAndSwapConfirmedParticipantConnectionState(
		ctx context.Context,
		change confirmedParticipantConnectionStateChange,
	) (ConfirmedParticipantConnectionStateChangeCASResult, error)
}
