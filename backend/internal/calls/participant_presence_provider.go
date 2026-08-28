package calls

import "github.com/google/uuid"

// TrustedProviderParticipantConnectionKind identifies a provider-normalized connection transition.
type TrustedProviderParticipantConnectionKind uint8

const (
	// TrustedProviderParticipantConnected reports that the resolved participant connected.
	TrustedProviderParticipantConnected TrustedProviderParticipantConnectionKind = iota + 1
	// TrustedProviderParticipantDisconnected reports that the resolved participant disconnected.
	TrustedProviderParticipantDisconnected
)

// TrustedProviderParticipantConnection is a trusted, provider-normalized presence fact.
// The source must already authenticate the provider event, resolve the room to a call,
// and resolve the participant identity to a Hub user before constructing this value.
type TrustedProviderParticipantConnection struct {
	CallID uuid.UUID
	UserID uuid.UUID
	Kind   TrustedProviderParticipantConnectionKind
}

// NormalizeTrustedProviderParticipantConnection maps a trusted provider fact to a confirmed observation.
// It does not verify raw events, apply or publish presence, or perform side effects.
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
