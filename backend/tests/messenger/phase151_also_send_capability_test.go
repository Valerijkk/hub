package messenger_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

func TestPhase151AlsoSendRequiresPostCapabilityInAdditionToReply(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, senderID)
	root := addLiveThreadRoot(repo, channelID, senderID)
	fixture := phase151Allow(messenger.ChannelActionReply, uuid.New(), channelID)
	access := fixture.accessByAction[messenger.ChannelActionReply]
	access.Role = "owner"
	access.CanReply = true
	access.CanPost = false
	fixture.accessByAction[messenger.ChannelActionReply] = access
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, fixture)

	denied, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{
		ContentText:       "must stay inside the thread",
		ThreadID:          root.ID,
		AlsoSendToChannel: true,
	})
	require.Error(t, err)
	assert.Nil(t, denied)
	requireAppErrorCode(t, err, 403)
	assert.Len(t, repo.messages, 1, "denial must happen before Mongo insert")
	assert.Equal(t, []messenger.ChannelAction{messenger.ChannelActionReply}, fixture.calls)

	allowed, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{
		ContentText: "thread-only reply remains allowed",
		ThreadID:    root.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, allowed)
	assert.False(t, allowed.AlsoSendToChannel)
}
