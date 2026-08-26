package messenger_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/internal/messenger"
)

func TestPhase151DeleteStageOneAllowsFreshMemberForOwnershipCheck(t *testing.T) {
	tenantID := uuid.New()
	actorID := uuid.New()
	channelID := uuid.New()
	facts := messenger.ChannelAuthorityFacts{
		TenantID:                   tenantID,
		ActorID:                    actorID,
		ActorOrganizationID:        tenantID,
		ActorActive:                true,
		Channel:                    &ent.Channel{ID: channelID, ChannelType: "private"},
		IsMember:                   true,
		Role:                       "member",
		HasActiveProjectMembership: true,
	}

	access, err := messenger.ExportPhase151ResolveChannelAccessFromFacts(
		facts,
		messenger.ChannelActionDeleteMessage,
	)

	require.NoError(t, err)
	assert.True(t, access.CanRead)
	assert.False(t, access.CanDeleteMessage, "ordinary membership must not grant delete-any authority")
}

func TestPhase151OrdinaryAuthorCanDeleteOwnMessageAfterFreshStageOneProof(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	ownerID := uuid.New()
	authorID := uuid.New()
	repo := newMockRepository()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID, authorID)
	messageID := uuid.NewString()
	repo.messages[messageID] = &messenger.MessageDocument{
		ID: messageID, ChannelID: channelID, SenderID: authorID,
	}
	resolver := phase151Allow(messenger.ChannelActionDeleteMessage, tenantID, channelID)
	access := resolver.accessByAction[messenger.ChannelActionDeleteMessage]
	access.Role = "member"
	access.IsMember = true
	access.CanRead = true
	access.CanDeleteMessage = false
	resolver.accessByAction[messenger.ChannelActionDeleteMessage] = access
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	err := service.DeleteMessage(ctx, messageID, authorID, channelID.String(), false)

	require.NoError(t, err)
	assert.NotNil(t, repo.messages[messageID].DeletedAt)
}

func TestPhase151OrdinaryMemberCannotDeleteAnotherUsersMessage(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	ownerID := uuid.New()
	actorID := uuid.New()
	authorID := uuid.New()
	repo := newMockRepository()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID, actorID, authorID)
	messageID := uuid.NewString()
	repo.messages[messageID] = &messenger.MessageDocument{
		ID: messageID, ChannelID: channelID, SenderID: authorID,
	}
	resolver := phase151Allow(messenger.ChannelActionDeleteMessage, tenantID, channelID)
	access := resolver.accessByAction[messenger.ChannelActionDeleteMessage]
	access.Role = "member"
	access.IsMember = true
	access.CanRead = true
	access.CanDeleteMessage = false
	resolver.accessByAction[messenger.ChannelActionDeleteMessage] = access
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	err := service.DeleteMessage(ctx, messageID, actorID, channelID.String(), false)

	require.Error(t, err)
	assert.Nil(t, repo.messages[messageID].DeletedAt)
}

func TestPhase151ModeratorCanDeleteAnotherUsersMessage(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	ownerID := uuid.New()
	moderatorID := uuid.New()
	authorID := uuid.New()
	repo := newMockRepository()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID, moderatorID, authorID)
	repo.members[channelID][moderatorID].Role = channelmember.Role("moderator")
	messageID := uuid.NewString()
	repo.messages[messageID] = &messenger.MessageDocument{
		ID: messageID, ChannelID: channelID, SenderID: authorID,
	}
	resolver := phase151Allow(messenger.ChannelActionDeleteMessage, tenantID, channelID)
	access := resolver.accessByAction[messenger.ChannelActionDeleteMessage]
	access.Role = "moderator"
	access.IsMember = true
	access.CanRead = true
	access.CanDeleteMessage = true
	resolver.accessByAction[messenger.ChannelActionDeleteMessage] = access
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	err := service.DeleteMessage(ctx, messageID, moderatorID, channelID.String(), false)

	require.NoError(t, err)
	assert.NotNil(t, repo.messages[messageID].DeletedAt)
}