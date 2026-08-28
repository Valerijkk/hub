package calls

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ErrConfirmedParticipantConnectionKeyInvalid reports an invalid confirmed
// participant connection key.
var ErrConfirmedParticipantConnectionKeyInvalid = errors.New("invalid confirmed participant connection key")

// ConfirmedParticipantConnectionKey is an opaque correlation and state key for
// a confirmed Hub participant connection. It is not authentication evidence:
// a trusted upstream establishes the call, user, and epoch before construction.
// Raw key values must not be logged.
type ConfirmedParticipantConnectionKey struct {
	CallID uuid.UUID
	UserID uuid.UUID
	Epoch  uuid.UUID
}

// NewConfirmedParticipantConnectionKey constructs a valid opaque confirmed
// participant connection key from trusted upstream call, user, and epoch IDs.
func NewConfirmedParticipantConnectionKey(
	callID, userID, epoch uuid.UUID,
) (ConfirmedParticipantConnectionKey, error) {
	key := ConfirmedParticipantConnectionKey{
		CallID: callID,
		UserID: userID,
		Epoch:  epoch,
	}
	if !key.IsValid() {
		return ConfirmedParticipantConnectionKey{}, ErrConfirmedParticipantConnectionKeyInvalid
	}

	return key, nil
}

// IsValid reports whether every component of the connection key is present.
func (key ConfirmedParticipantConnectionKey) IsValid() bool {
	return key.CallID != uuid.Nil && key.UserID != uuid.Nil && key.Epoch != uuid.Nil
}

// Canonical returns the lowercase opaque connection-key representation, or an
// empty string when the key is invalid.
func (key ConfirmedParticipantConnectionKey) Canonical() string {
	if !key.IsValid() {
		return ""
	}

	return fmt.Sprintf("hub.call-connection.v1/%s/%s/%s", key.CallID, key.UserID, key.Epoch)
}
