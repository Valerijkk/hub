package calls

import (
	"context"
	"errors"
	"time"
)

const (
	// ProviderLiveKit identifies the LiveKit conference provider.
	ProviderLiveKit = "livekit"
	// ProviderJitsi identifies the dormant self-hosted Jitsi provider.
	ProviderJitsi = "jitsi"
)

// ConferenceProviderID is the closed identity used by the dormant provider registry.
type ConferenceProviderID string

// ConferenceProviderCapability identifies one independently verifiable provider operation.
type ConferenceProviderCapability uint8

// ConferenceProviderCapabilities is an immutable provider capability bit set.
type ConferenceProviderCapabilities uint8

const (
	CapabilityIssueCallJoinCredential ConferenceProviderCapability = 1 << iota
	CapabilityProvisionRoom
	CapabilityDeleteRoom
	CapabilityIssueGuestCredential
	CapabilityUpdateParticipantMetadata
)

const allConferenceProviderCapabilities = ConferenceProviderCapabilities(
	CapabilityIssueCallJoinCredential |
		CapabilityProvisionRoom |
		CapabilityDeleteRoom |
		CapabilityIssueGuestCredential |
		CapabilityUpdateParticipantMetadata,
)

// Has reports whether the capability set contains the supplied bit.
func (c ConferenceProviderCapabilities) Has(capability ConferenceProviderCapability) bool {
	bit := ConferenceProviderCapabilities(capability)
	return bit != 0 && c&bit == bit
}

// ConferenceRoomProvisioner exposes only room provisioning capability.
type ConferenceRoomProvisioner interface {
	Provision(context.Context, string, uint32) error
}

// ConferenceRoomDeleter exposes only room deletion capability.
type ConferenceRoomDeleter interface {
	Delete(context.Context, string) error
}

// ConferenceGuestCredentialIssuer exposes only guest credential capability.
type ConferenceGuestCredentialIssuer interface {
	IssueGuestCredential(string, string) (GuestJoinCredential, error)
}

// ConferenceParticipantMetadataUpdater exposes only participant metadata capability.
type ConferenceParticipantMetadataUpdater interface {
	UpdateParticipantState(context.Context, string, string, string) error
}

// RegisteredConferenceProvider is the immutable registry boundary.
type RegisteredConferenceProvider interface {
	CallJoinCredentialIssuer
	ProviderID() ConferenceProviderID
	Capabilities() ConferenceProviderCapabilities
}

var (
	ErrConferenceProviderRegistryInvalid       = errors.New("invalid conference provider registry")
	ErrConferenceProviderRegistryDuplicate     = errors.New("duplicate conference provider registration")
	ErrConferenceProviderUnknown               = errors.New("unknown conference provider")
	ErrConferenceProviderNotRegistered         = errors.New("conference provider is not registered")
	ErrConferenceProviderCapabilityUnavailable = errors.New("conference provider capability is unavailable")
)

// JoinCredential contains provider-neutral data required to join a conference.
type JoinCredential struct {
	Provider   string
	RoomName   string
	ServiceURL string
	Token      string
	ExpiresAt  time.Time
}

// GuestJoinCredential binds a generated guest identity to its join credential.
type GuestJoinCredential struct {
	Identity string
	Join     JoinCredential
}

// ConferenceProvider defines the provider-neutral conference lifecycle boundary.
type ConferenceProvider interface {
	Provision(ctx context.Context, roomName string, maxParticipants uint32) error
	Delete(ctx context.Context, roomName string) error
	IssueJoinCredential(roomName, userID, userName string) (JoinCredential, error)
	IssueGuestCredential(roomName, guestName string) (GuestJoinCredential, error)
	UpdateParticipantState(ctx context.Context, roomName, identity, stateJSON string) error
}
