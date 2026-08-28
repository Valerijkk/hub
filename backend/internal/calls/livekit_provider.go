package calls

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/valerii/hub/backend/internal/platform/config"
)

// LiveKitConferenceProvider adapts the LiveKit SDK to the ConferenceProvider boundary.
type LiveKitConferenceProvider struct {
	client *lksdk.RoomServiceClient
	cfg    config.LiveKitConfig
}

var _ ConferenceProvider = (*LiveKitConferenceProvider)(nil)

// NewLiveKitConferenceProvider constructs a network-free LiveKit conference provider.
func NewLiveKitConferenceProvider(cfg config.LiveKitConfig) *LiveKitConferenceProvider {
	return &LiveKitConferenceProvider{
		client: lksdk.NewRoomServiceClient(cfg.URL, cfg.APIKey, cfg.APISecret),
		cfg:    cfg,
	}
}

// Provision creates a LiveKit room with the supplied participant limit.
func (p *LiveKitConferenceProvider) Provision(ctx context.Context, roomName string, maxParticipants uint32) error {
	_, err := p.client.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            roomName,
		MaxParticipants: maxParticipants,
	})
	if err != nil {
		return fmt.Errorf("не удалось создать комнату LiveKit: %w", err)
	}
	return nil
}

// Delete removes a LiveKit room.
func (p *LiveKitConferenceProvider) Delete(ctx context.Context, roomName string) error {
	_, err := p.client.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
		Room: roomName,
	})
	return err
}

// IssueJoinCredential creates a 24-hour LiveKit room-join credential for a user.
func (p *LiveKitConferenceProvider) IssueJoinCredential(roomName, userID, userName string) (JoinCredential, error) {
	at := auth.NewAccessToken(p.cfg.APIKey, p.cfg.APISecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}
	at.SetVideoGrant(grant).
		SetIdentity(userID).
		SetName(userName).
		SetValidFor(24 * time.Hour)

	token, err := at.ToJWT()
	if err != nil {
		return JoinCredential{}, err
	}
	verifier, err := auth.ParseAPIToken(token)
	if err != nil {
		return JoinCredential{}, errors.New("could not verify issued LiveKit credential")
	}
	claims, _, err := verifier.Verify(p.cfg.APISecret)
	if err != nil {
		return JoinCredential{}, errors.New("could not verify issued LiveKit credential")
	}
	if claims == nil || claims.Expiry == nil {
		return JoinCredential{}, errors.New("issued LiveKit credential is missing an expiry")
	}
	expiresAt := claims.Expiry.Time().UTC()
	if expiresAt.IsZero() {
		return JoinCredential{}, errors.New("issued LiveKit credential is missing an expiry")
	}

	return JoinCredential{
		Provider:   ProviderLiveKit,
		RoomName:   roomName,
		ServiceURL: p.cfg.URL,
		Token:      token,
		ExpiresAt:  expiresAt,
	}, nil
}

// IssueGuestCredential creates a fresh guest identity and its LiveKit join credential.
func (p *LiveKitConferenceProvider) IssueGuestCredential(roomName, guestName string) (GuestJoinCredential, error) {
	identity := "guest-" + uuid.New().String()
	credential, err := p.IssueJoinCredential(roomName, identity, guestName)
	if err != nil {
		return GuestJoinCredential{Identity: identity}, err
	}

	return GuestJoinCredential{
		Identity: identity,
		Join:     credential,
	}, nil
}

// UpdateParticipantState stores provider-specific participant state as LiveKit metadata.
func (p *LiveKitConferenceProvider) UpdateParticipantState(ctx context.Context, roomName, identity, stateJSON string) error {
	_, err := p.client.UpdateParticipant(ctx, &livekit.UpdateParticipantRequest{
		Room:     roomName,
		Identity: identity,
		Metadata: stateJSON,
	})
	return err
}
