package calls

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

const (
	invalidProviderParticipantIdentity = "invalid provider participant identity"
	providerParticipantIdentityPrefix = "hub.call-participant.v1/"
	providerParticipantIdentityLength = 134
)

// ProviderParticipantIdentity is a provider-facing identity derived only from
// a provider-neutral participant connection key. It is not authentication or
// authorization evidence. Raw values must not be logged.
type ProviderParticipantIdentity struct {
	connectionKey CallParticipantConnectionKey
}

// NewProviderParticipantIdentity constructs an identity from a valid neutral
// connection key.
func NewProviderParticipantIdentity(connectionKey CallParticipantConnectionKey) (ProviderParticipantIdentity, error) {
	if !connectionKey.IsValid() {
		return ProviderParticipantIdentity{}, ErrProviderParticipantIdentityInvalid
	}

	return ProviderParticipantIdentity{connectionKey: connectionKey}, nil
}

// ParseProviderParticipantIdentity parses only the canonical provider-facing
// representation. All other inputs return the zero value and invalid sentinel.
func ParseProviderParticipantIdentity(canonical string) (ProviderParticipantIdentity, error) {
	if len(canonical) != providerParticipantIdentityLength || !strings.HasPrefix(canonical, providerParticipantIdentityPrefix) {
		return ProviderParticipantIdentity{}, ErrProviderParticipantIdentityInvalid
	}

	segments := strings.Split(strings.TrimPrefix(canonical, providerParticipantIdentityPrefix), "/")
	if len(segments) != 3 {
		return ProviderParticipantIdentity{}, ErrProviderParticipantIdentityInvalid
	}

	callID, ok := parseCanonicalNonZeroUUID(segments[0])
	if !ok {
		return ProviderParticipantIdentity{}, ErrProviderParticipantIdentityInvalid
	}
	userID, ok := parseCanonicalNonZeroUUID(segments[1])
	if !ok {
		return ProviderParticipantIdentity{}, ErrProviderParticipantIdentityInvalid
	}
	epoch, ok := parseCanonicalNonZeroUUID(segments[2])
	if !ok {
		return ProviderParticipantIdentity{}, ErrProviderParticipantIdentityInvalid
	}

	connectionKey, err := NewCallParticipantConnectionKey(callID, userID, epoch)
	if err != nil {
		return ProviderParticipantIdentity{}, ErrProviderParticipantIdentityInvalid
	}

	return NewProviderParticipantIdentity(connectionKey)
}

// IsValid reports whether the underlying neutral connection key is valid.
func (i ProviderParticipantIdentity) IsValid() bool {
	return i.connectionKey.IsValid()
}

// Canonical returns the stable provider-facing wire representation.
func (i ProviderParticipantIdentity) Canonical() string {
	if !i.IsValid() {
		return ""
	}

	return providerParticipantIdentityPrefix + i.connectionKey.CallID().String() + "/" + i.connectionKey.UserID().String() + "/" + i.connectionKey.Epoch().String()
}

// ConnectionKey returns the neutral connection key from which this identity is derived.
func (i ProviderParticipantIdentity) ConnectionKey() CallParticipantConnectionKey {
	if !i.IsValid() {
		return CallParticipantConnectionKey{}
	}

	return i.connectionKey
}

func parseCanonicalNonZeroUUID(value string) (uuid.UUID, bool) {
	if len(value) != 36 || value != strings.ToLower(value) {
		return uuid.Nil, false
	}

	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.String() != value {
		return uuid.Nil, false
	}

	return parsed, true
}

// ErrProviderParticipantIdentityInvalid reports an invalid canonical provider
// participant identity.
var ErrProviderParticipantIdentityInvalid = errors.New(invalidProviderParticipantIdentity)
