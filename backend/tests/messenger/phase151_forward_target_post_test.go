package messenger_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
)

func TestPhase151ForwardTargetCanForwardWithoutCanPostIsRejectedBeforeMongoOrPublish(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	sourceChannelID := uuid.New()
	targetChannelID := uuid.New()
	sourceMessageID := "phase151-forward-target-without-post"
	repo := newPhase151MutationRepository()
	sourceChannel := &ent.Channel{ID: sourceChannelID, ChannelType: "private"}
	targetChannel := &ent.Channel{ID: targetChannelID, ChannelType: "private"}
	repo.messages[sourceMessageID] = &messenger.MessageDocument{
		ID:        sourceMessageID,
		ChannelID: sourceChannelID,
		SenderID:  uuid.New(),
	}

	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	sourceAccess := phase151ReadableAccess(sourceChannel, true)
	sourceAccess.CanForward = true
	phase151Grant(resolver, actorID, sourceChannel, messenger.ChannelActionForwardSource, sourceAccess)
	targetAccess := phase151ReadableAccess(targetChannel, true)
	targetAccess.CanForward = true
	phase151Grant(resolver, actorID, targetChannel, messenger.ChannelActionForwardTarget, targetAccess)
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	forwarded, err := service.ForwardMessage(ctx, sourceChannelID, actorID, sourceMessageID, targetChannelID.String())

	require.Error(t, err)
	assert.Nil(t, forwarded)
	assert.Equal(t, []phase151ChannelAccessKey{
		{actorID: actorID, channelID: sourceChannelID, action: messenger.ChannelActionForwardSource},
		{actorID: actorID, channelID: targetChannelID, action: messenger.ChannelActionForwardTarget},
	}, resolver.calls)
	assert.Zero(t, repo.findCalls, "a target without CanPost must not disclose the source message")
	assert.Zero(t, repo.insertCalls, "a target without CanPost must not create a forwarded Mongo message or publish it")
	assert.Empty(t, repo.operations)
}
