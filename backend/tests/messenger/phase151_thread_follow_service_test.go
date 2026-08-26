package messenger_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/internal/messenger"
)

type phase151ThreadFollowMutationRepository struct {
	*phase151ThreadUnreadRepository
	headCalls       int
	transitionCalls int
	transition      messenger.ThreadFollowTransition
	head            messenger.MessagePosition
}

func (r *phase151ThreadFollowMutationRepository) CurrentAuthorizedFollowedRootHead(
	context.Context,
	messenger.AuthorizedUnreadScope,
	messenger.AuthorizedUnreadTarget,
	string,
) (messenger.MessagePosition, error) {
	r.headCalls++
	return r.head, nil
}

func (r *phase151ThreadFollowMutationRepository) ApplyAuthorizedThreadFollowTransition(
	_ context.Context,
	_ messenger.AuthorizedUnreadScope,
	_ uuid.UUID,
	_ string,
	transition messenger.ThreadFollowTransition,
) (bool, error) {
	r.transitionCalls++
	r.transition = transition
	return true, nil
}

func TestPhase151ManualFollowAndUnfollowUseFreshMemberScopeAndCurrentHead(t *testing.T) {
	tenantID, actorID, channelID := uuid.New(), uuid.New(), uuid.New()
	rootID := "phase151-follow-root"
	repo := &phase151ThreadFollowMutationRepository{
		phase151ThreadUnreadRepository: newPhase151ThreadUnreadRepository(),
		head: messenger.MessagePosition{CreatedAt: time.Unix(10, 0).UTC(), MessageID: rootID},
	}
	resolver := &phase151ThreadAccessResolver{grants: make(map[phase151ThreadAccessKey]messenger.ChannelAccess)}
	key := phase151ThreadAccessKey{actorID: actorID, channelID: channelID, action: messenger.ChannelActionReadThread}
	resolver.grants[key] = messenger.ChannelAccess{
		TenantID: tenantID,
		Channel: &ent.Channel{ID: channelID, ChannelType: channel.ChannelTypePrivate},
		IsMember: true, CanRead: true, CanFollow: true, CanUnfollow: true,
	}
	scope := &phase151UnreadScopeFixture{scope: messenger.AuthorizedUnreadScope{
		TenantID: tenantID,
		UserID:   actorID,
		Targets:  []messenger.AuthorizedUnreadTarget{{ChannelID: channelID, Kind: "channel"}},
	}}
	service := messenger.ExportNewMessengerServiceWithChannelAccessAndUnreadScopeTest(repo, resolver, scope)

	changed, err := service.FollowThread(context.Background(), channelID, rootID, actorID)
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, messenger.ThreadFollowModeManualFollow, repo.transition.Mode)
	assert.Equal(t, repo.head, repo.transition.Position)

	changed, err = service.UnfollowThread(context.Background(), channelID, rootID, actorID)
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, messenger.ThreadFollowModeManualUnfollow, repo.transition.Mode)
	assert.Equal(t, 2, repo.headCalls)
	assert.Equal(t, 2, repo.transitionCalls)
	assert.Equal(t, []uuid.UUID{actorID, actorID}, scope.calls)
}

func TestPhase151PublicPrejoinCannotReadOrMutateFollowState(t *testing.T) {
	tenantID, actorID, channelID := uuid.New(), uuid.New(), uuid.New()
	repo := &phase151ThreadFollowMutationRepository{
		phase151ThreadUnreadRepository: newPhase151ThreadUnreadRepository(),
		head: messenger.MessagePosition{CreatedAt: time.Unix(10, 0).UTC(), MessageID: "root"},
	}
	resolver := &phase151ThreadAccessResolver{grants: map[phase151ThreadAccessKey]messenger.ChannelAccess{
		{actorID: actorID, channelID: channelID, action: messenger.ChannelActionReadThread}: {
			TenantID: tenantID,
			Channel:  &ent.Channel{ID: channelID, ChannelType: channel.ChannelTypePublic},
			CanRead:  true,
		},
	}}
	scope := &phase151UnreadScopeFixture{scope: messenger.AuthorizedUnreadScope{TenantID: tenantID, UserID: actorID}}
	service := messenger.ExportNewMessengerServiceWithChannelAccessAndUnreadScopeTest(repo, resolver, scope)

	changed, err := service.FollowThread(context.Background(), channelID, "root", actorID)

	require.Error(t, err)
	assert.False(t, changed)
	assert.Empty(t, scope.calls)
	assert.Zero(t, repo.headCalls)
	assert.Zero(t, repo.transitionCalls)
}
