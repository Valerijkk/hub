package messenger_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
)

func TestPhase151GetChannelReturnsFreshPublicPrejoinCapabilities(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	channelID := uuid.New()
	actorID := uuid.New()
	fixture := phase151Allow(messenger.ChannelActionReadTimeline, uuid.New(), channelID)
	access := fixture.accessByAction[messenger.ChannelActionReadTimeline]
	access.Channel = &ent.Channel{
		ID:          channelID,
		ChannelType: "public",
		Name:        "discoverable",
		CreatedBy:   uuid.New(),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	access.Role = ""
	access.IsMember = false
	access.CanRead = true
	access.CanJoin = true
	access.CanPost = false
	access.CanReply = false
	fixture.accessByAction[messenger.ChannelActionReadTimeline] = access
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, fixture)

	response, err := service.GetChannel(ctx, channelID, actorID)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.Capabilities)
	assert.Equal(t, channelID.String(), response.ID)
	assert.False(t, response.Capabilities.IsMember)
	assert.True(t, response.Capabilities.CanRead)
	assert.True(t, response.Capabilities.CanJoin)
	assert.False(t, response.Capabilities.CanPost)
	assert.False(t, response.Capabilities.CanReply)
	assert.Equal(t, []messenger.ChannelAction{messenger.ChannelActionReadTimeline}, fixture.calls)
}

func TestPhase151GetChannelDeniesBeforeRepositoryDisclosure(t *testing.T) {
	repo := newMockRepository()
	channelID := uuid.New()
	fixture := phase151Deny(messenger.ChannelActionReadTimeline)
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, fixture)

	response, err := service.GetChannel(context.Background(), channelID, uuid.New())

	require.Error(t, err)
	assert.Nil(t, response)
	requireAppErrorCode(t, err, 403)
	assert.Equal(t, []messenger.ChannelAction{messenger.ChannelActionReadTimeline}, fixture.calls)
}

type phase151ListAccessResolver struct {
	accessByChannel map[uuid.UUID]messenger.ChannelAccess
	calls           []uuid.UUID
}

func (r *phase151ListAccessResolver) AuthorizeChannelAccess(
	_ context.Context,
	_ uuid.UUID,
	channelID uuid.UUID,
	action messenger.ChannelAction,
) (messenger.ChannelAccess, error) {
	if action != messenger.ChannelActionReadTimeline {
		return messenger.ChannelAccess{}, messenger.ErrChannelAccessDenied
	}
	r.calls = append(r.calls, channelID)
	access, ok := r.accessByChannel[channelID]
	if !ok {
		return messenger.ChannelAccess{}, messenger.ErrChannelAccessDenied
	}
	return access, nil
}

func TestPhase151ListMyChannelsUsesFreshAccessAndDropsRevokedCandidates(t *testing.T) {
	repo := newMockRepository()
	actorID := uuid.New()
	allowedID := setupChannelWithMembers(repo, "private", false, actorID)
	revokedID := setupChannelWithMembers(repo, "private", false, actorID)
	repo.channels[allowedID].Name = "stale candidate"
	resolver := &phase151ListAccessResolver{
		accessByChannel: map[uuid.UUID]messenger.ChannelAccess{
			allowedID: {
				TenantID: uuid.New(),
				Channel: &ent.Channel{
					ID: allowedID, ChannelType: "private", Name: "fresh authority",
					CreatedBy: uuid.New(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
				},
				IsMember: true, CanRead: true, CanPost: true, CanReply: true,
			},
		},
	}
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	channels, err := service.ListMyChannels(context.Background(), actorID)

	require.NoError(t, err)
	require.Len(t, channels, 1)
	assert.Equal(t, allowedID.String(), channels[0].ID)
	assert.Equal(t, "fresh authority", channels[0].Name)
	require.NotNil(t, channels[0].Capabilities)
	assert.True(t, channels[0].Capabilities.CanRead)
	assert.ElementsMatch(t, []uuid.UUID{allowedID, revokedID}, resolver.calls)
}
