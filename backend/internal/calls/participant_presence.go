package calls

import (
	"errors"

	"github.com/google/uuid"
)

// ErrConfirmedParticipantObservationInvalid marks malformed trusted observations.
var ErrConfirmedParticipantObservationInvalid = errors.New("invalid confirmed participant observation")

// ParticipantPresence is the canonical confirmed participant state.
type ParticipantPresence uint8

const (
	ParticipantPresenceUnknown ParticipantPresence = iota
	ParticipantPresenceConnected
	ParticipantPresenceLeft
)

// ConfirmedParticipantObservation is a provider-normalized presence fact.
type ConfirmedParticipantObservation struct {
	CallID uuid.UUID
	UserID uuid.UUID
	Kind   ParticipantPresence
}

// ParticipantPresenceTransition describes the canonical effect of an observation.
type ParticipantPresenceTransition struct {
	Previous      ParticipantPresence
	Next          ParticipantPresence
	Persist       bool
	PublishJoined bool
	PublishLeft   bool
}

// ReduceConfirmedParticipantObservation calculates the idempotent transition for a trusted observation.
func ReduceConfirmedParticipantObservation(
	previous ParticipantPresence,
	observation ConfirmedParticipantObservation,
) (ParticipantPresenceTransition, error) {
	if !isParticipantPresence(previous) ||
		observation.CallID == uuid.Nil ||
		observation.UserID == uuid.Nil ||
		(observation.Kind != ParticipantPresenceConnected && observation.Kind != ParticipantPresenceLeft) {
		return ParticipantPresenceTransition{}, ErrConfirmedParticipantObservationInvalid
	}

	switch observation.Kind {
	case ParticipantPresenceConnected:
		if previous == ParticipantPresenceConnected {
			return ParticipantPresenceTransition{
				Previous: previous,
				Next:     ParticipantPresenceConnected,
			}, nil
		}

		return ParticipantPresenceTransition{
			Previous:      previous,
			Next:          ParticipantPresenceConnected,
			Persist:       true,
			PublishJoined: true,
		}, nil
	case ParticipantPresenceLeft:
		if previous == ParticipantPresenceConnected {
			return ParticipantPresenceTransition{
				Previous:    previous,
				Next:        ParticipantPresenceLeft,
				Persist:     true,
				PublishLeft: true,
			}, nil
		}

		return ParticipantPresenceTransition{
			Previous: previous,
			Next:     previous,
		}, nil
	}

	return ParticipantPresenceTransition{}, ErrConfirmedParticipantObservationInvalid
}

func isParticipantPresence(presence ParticipantPresence) bool {
	return presence == ParticipantPresenceUnknown ||
		presence == ParticipantPresenceConnected ||
		presence == ParticipantPresenceLeft
}
