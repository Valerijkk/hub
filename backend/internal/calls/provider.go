package calls

import (
	"context"
	"time"
)

// ProviderLiveKit identifies the LiveKit conference provider.
const ProviderLiveKit = "livekit"

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
