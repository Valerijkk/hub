package messenger_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

type phase151RecipientProjectionCall struct {
	actorID   uuid.UUID
	channelID uuid.UUID
	action    messenger.ChannelAction
}

type phase151RecipientProjectionResolver struct {
	repo  *mockRepository
	calls []phase151RecipientProjectionCall
}

func (r *phase151RecipientProjectionResolver) AuthorizeChannelAccess(
	_ context.Context,
	actorID uuid.UUID,
	channelID uuid.UUID,
	action messenger.ChannelAction,
) (messenger.ChannelAccess, error) {
	r.calls = append(r.calls, phase151RecipientProjectionCall{actorID: actorID, channelID: channelID, action: action})
	if action != messenger.ChannelActionReadTimeline {
		return messenger.ChannelAccess{}, errors.New("unexpected recipient projection action")
	}
	channelEntity, ok := r.repo.channels[channelID]
	if !ok || channelEntity.DeletedAt != nil {
		return messenger.ChannelAccess{}, errors.New("channel unavailable")
	}
	membership, ok := r.repo.members[channelID][actorID]
	if !ok {
		return messenger.ChannelAccess{}, errors.New("recipient is not a member")
	}
	return messenger.ChannelAccess{
		TenantID:                   uuid.New(),
		Channel:                    channelEntity,
		Role:                       string(membership.Role),
		IsMember:                   true,
		HasActiveProjectMembership: true,
		CanRead:                    true,
		CanPost:                    false,
		CanReply:                   true,
		CanFollow:                  true,
		CanUnfollow:                true,
		CanViewUnread:              true,
	}, nil
}

func requirePhase151RecipientResponse(
	t *testing.T,
	response *messenger.ChannelResponse,
	resolver *phase151RecipientProjectionResolver,
	actorID uuid.UUID,
) {
	t.Helper()
	require.NotNil(t, response)
	require.NotNil(t, response.Capabilities)
	assert.True(t, response.Capabilities.IsMember)
	assert.True(t, response.Capabilities.CanRead)
	assert.False(t, response.Capabilities.CanPost, "response must copy fresh capability rather than infer from role")
	require.NotEmpty(t, resolver.calls)
	assert.Equal(t, phase151RecipientProjectionCall{
		actorID: actorID, channelID: uuid.MustParse(response.ID), action: messenger.ChannelActionReadTimeline,
	}, resolver.calls[len(resolver.calls)-1])
}

func TestPhase151CreateAndUpdateChannelResponsesUseFreshRecipientProjection(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	resolver := &phase151RecipientProjectionResolver{repo: repo}
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)
	ownerID := uuid.New()

	created, err := service.CreateChannel(ctx, ownerID, messenger.CreateChannelRequest{
		ChannelType: "private", Name: "projection-channel",
	})
	require.NoError(t, err)
	requirePhase151RecipientResponse(t, created, resolver, ownerID)

	updated, err := service.UpdateChannel(ctx, uuid.MustParse(created.ID), ownerID, messenger.CreateChannelRequest{
		Name: "projection-channel-renamed", Version: 1,
	})
	require.NoError(t, err)
	requirePhase151RecipientResponse(t, updated, resolver, ownerID)
	assert.Equal(t, "projection-channel-renamed", updated.Name)
}

func TestPhase151NewAndExistingDMResponsesUseFreshRecipientProjection(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	resolver := &phase151RecipientProjectionResolver{repo: repo}
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)
	firstID, secondID := uuid.New(), uuid.New()

	created, err := service.CreateOrGetDM(ctx, firstID, secondID)
	require.NoError(t, err)
	requirePhase151RecipientResponse(t, created, resolver, firstID)

	existing, err := service.CreateOrGetDM(ctx, secondID, firstID)
	require.NoError(t, err)
	requirePhase151RecipientResponse(t, existing, resolver, secondID)
	assert.Equal(t, created.ID, existing.ID)
}

func TestPhase151AcceptInviteResponseUsesFreshRecipientProjection(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	ownerID, invitedID := uuid.New(), uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)
	_, err := repo.CreateChannelInvite(ctx, messenger.CreateChannelInviteInput{
		ChannelID: channelID,
		CreatorID: ownerID,
		Token:     "phase151-recipient-projection-invite",
	})
	require.NoError(t, err)
	resolver := &phase151RecipientProjectionResolver{repo: repo}
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	response, err := service.AcceptInvite(ctx, "phase151-recipient-projection-invite", invitedID)

	require.NoError(t, err)
	requirePhase151RecipientResponse(t, response, resolver, invitedID)
}
