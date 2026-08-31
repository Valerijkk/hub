package calls

import (
	"errors"

	"github.com/google/uuid"
)

const invalidCallParticipantConnectionKey = "invalid call participant connection key"

// ErrCallParticipantConnectionKeyInvalid reports an invalid neutral call
// participant connection key.
var ErrCallParticipantConnectionKeyInvalid = errors.New(invalidCallParticipantConnectionKey)

// CallParticipantConnectionKey identifies a user's connection to a call for a
// specific connection epoch. It is provider-neutral correlation data, not
// authentication or participant-confirmation evidence. Raw values must not be
// logged.
type CallParticipantConnectionKey struct {
	callID uuid.UUID
	userID uuid.UUID
	epoch  uuid.UUID
}

// NewCallParticipantConnectionKey constructs a provider-neutral participant
// connection key. Any zero component returns the zero value and the exact
// invalid-key sentinel.
func NewCallParticipantConnectionKey(callID, userID, epoch uuid.UUID) (CallParticipantConnectionKey, error) {
	key := CallParticipantConnectionKey{
		callID: callID,
		userID: userID,
		epoch:  epoch,
	}
	if !key.IsValid() {
		return CallParticipantConnectionKey{}, ErrCallParticipantConnectionKeyInvalid
	}

	return key, nil
}

// IsValid reports whether every identifying component is non-zero.
func (k CallParticipantConnectionKey) IsValid() bool {
	return k.callID != uuid.Nil && k.userID != uuid.Nil && k.epoch != uuid.Nil
}

// Canonical returns the stable provider-neutral wire representation.
func (k CallParticipantConnectionKey) Canonical() string {
	if !k.IsValid() {
		return ""
	}

	return "hub.call-connection.v1/" + k.callID.String() + "/" + k.userID.String() + "/" + k.epoch.String()
}

// CallID returns the call identifier.
func (k CallParticipantConnectionKey) CallID() uuid.UUID {
	return k.callID
}

// UserID returns the user identifier.
func (k CallParticipantConnectionKey) UserID() uuid.UUID {
	return k.userID
}

// Epoch returns the connection epoch identifier.
func (k CallParticipantConnectionKey) Epoch() uuid.UUID {
	return k.epoch
}
