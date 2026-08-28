package calls

import "errors"

// ErrConfirmedParticipantConnectionTransitionInvalid marks malformed or
// impossible confirmed participant connection transitions.
var ErrConfirmedParticipantConnectionTransitionInvalid = errors.New(
	"invalid confirmed participant connection transition",
)

// ConfirmedParticipantConnectionSnapshot is the trusted, already-resolved
// state of one Hub-owned participant connection and its user's aggregate count.
type ConfirmedParticipantConnectionSnapshot struct {
	Key                ConfirmedParticipantConnectionKey
	State              ParticipantPresence
	UserConnectedCount uint32
	CallEnded          bool
}

// ConfirmedParticipantConnectionTransition describes the persistence and
// publication effects calculated from one trusted presence observation.
type ConfirmedParticipantConnectionTransition struct {
	Key                        ConfirmedParticipantConnectionKey
	PreviousState              ParticipantPresence
	NextState                  ParticipantPresence
	PreviousUserConnectedCount uint32
	NextUserConnectedCount     uint32
	PersistConnection          bool
	PersistUserRefCount        bool
	PublishJoined              bool
	PublishLeft                bool
	RequestReconcile           bool
}

// ReduceConfirmedParticipantConnection calculates the connection-aware,
// idempotent transition for a trusted confirmed participant observation.
func ReduceConfirmedParticipantConnection(
	snapshot ConfirmedParticipantConnectionSnapshot,
	observation ConfirmedParticipantObservation,
) (ConfirmedParticipantConnectionTransition, error) {
	if !snapshot.Key.IsValid() ||
		observation.CallID != snapshot.Key.CallID ||
		observation.UserID != snapshot.Key.UserID ||
		(observation.Kind != ParticipantPresenceConnected && observation.Kind != ParticipantPresenceLeft) ||
		!isParticipantPresence(snapshot.State) ||
		(snapshot.State == ParticipantPresenceConnected && snapshot.UserConnectedCount == 0) {
		return ConfirmedParticipantConnectionTransition{}, ErrConfirmedParticipantConnectionTransitionInvalid
	}

	transition := ConfirmedParticipantConnectionTransition{
		Key:                        snapshot.Key,
		PreviousState:              snapshot.State,
		NextState:                  snapshot.State,
		PreviousUserConnectedCount: snapshot.UserConnectedCount,
		NextUserConnectedCount:     snapshot.UserConnectedCount,
	}
	if snapshot.CallEnded {
		return transition, nil
	}

	if (snapshot.State == ParticipantPresenceUnknown &&
		observation.Kind == ParticipantPresenceConnected &&
		snapshot.UserConnectedCount == ^uint32(0)) ||
		(snapshot.State == ParticipantPresenceLeft && observation.Kind == ParticipantPresenceConnected) {
		return ConfirmedParticipantConnectionTransition{}, ErrConfirmedParticipantConnectionTransitionInvalid
	}

	switch {
	case snapshot.State == ParticipantPresenceUnknown && observation.Kind == ParticipantPresenceConnected:
		transition.NextState = ParticipantPresenceConnected
		transition.NextUserConnectedCount++
		transition.PersistConnection = true
		transition.PersistUserRefCount = true
		transition.PublishJoined = snapshot.UserConnectedCount == 0
		return transition, nil
	case snapshot.State == ParticipantPresenceConnected && observation.Kind == ParticipantPresenceConnected:
		return transition, nil
	case snapshot.State == ParticipantPresenceConnected && observation.Kind == ParticipantPresenceLeft:
		transition.NextState = ParticipantPresenceLeft
		transition.NextUserConnectedCount--
		transition.PersistConnection = true
		transition.PersistUserRefCount = true
		transition.PublishLeft = snapshot.UserConnectedCount == 1
		return transition, nil
	case snapshot.State == ParticipantPresenceUnknown && observation.Kind == ParticipantPresenceLeft:
		transition.RequestReconcile = true
		return transition, nil
	case snapshot.State == ParticipantPresenceLeft && observation.Kind == ParticipantPresenceLeft:
		return transition, nil
	default:
		return ConfirmedParticipantConnectionTransition{}, ErrConfirmedParticipantConnectionTransitionInvalid
	}
}
