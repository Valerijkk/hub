package calls

import "github.com/google/uuid"

type TrustedProviderParticipantConnectionKind uint8

const (
	TrustedProviderParticipantConnected TrustedProviderParticipantConnectionKind = iota + 1
	TrustedProviderParticipantDisconnected
)

type TrustedProviderParticipantConnection struct {
	CallID uuid.UUID
	UserID uuid.UUID
	Kind   TrustedProviderParticipantConnectionKind
}

func NormalizeTrustedProviderParticipantConnection(
	fact TrustedProviderParticipantConnection,
) (ConfirmedParticipantObservation, error) {
	if fact.CallID == uuid.Nil || fact.UserID == uuid.Nil {
		return ConfirmedParticipantObservation{}, ErrConfirmedParticipantObservationInvalid
	}

	var kind ParticipantPresence
	switch fact.Kind {
	case TrustedProviderParticipantConnected:
		kind = ParticipantPresenceConnected
	case TrustedProviderParticipantDisconnected:
		kind = ParticipantPresenceLeft
	default:
		return ConfirmedParticipantObservation{}, ErrConfirmedParticipantObservationInvalid
	}

	return ConfirmedParticipantObservation{
		CallID: fact.CallID,
		UserID: fact.UserID,
		Kind:   kind,
	}, nil
}
