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

type phase151ReactionRepository struct {
	*phase151MutationRepository
	isMemberCalls int
}

func newPhase151ReactionRepository() *phase151ReactionRepository {
	return &phase151ReactionRepository{phase151MutationRepository: newPhase151MutationRepository()}
}

func (r *phase151ReactionRepository) IsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	r.isMemberCalls++
	return r.mockRepository.IsMember(ctx, channelID, userID)
}

func TestPhase151ReactionRequiresFreshCanReactBeforeMongo(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	channelID := uuid.New()
	messageID := "phase151-reaction-denied"
	repo := newPhase151ReactionRepository()
	channel := &ent.Channel{ID: channelID, ChannelType: "private"}
	repo.channels[channelID] = channel
	repo.messages[messageID] = &messenger.MessageDocument{ID: messageID, ChannelID: channelID, SenderID: uuid.New()}
	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	denied := phase151ReadableAccess(channel, true)
	phase151Grant(resolver, actorID, channel, messenger.ChannelActionReact, denied)
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	added, err := service.ToggleReaction(ctx, channelID, messageID, actorID, "thumbsup")

	require.Error(t, err)
	assert.False(t, added)
	assert.Equal(t, []phase151ChannelAccessKey{{actorID: actorID, channelID: channelID, action: messenger.ChannelActionReact}}, resolver.calls)
	assert.Zero(t, repo.findCalls, "denied reaction must not discover a Mongo message")
	assert.Zero(t, repo.isMemberCalls, "scoped authority must not fall back to repository membership")
	assert.Empty(t, repo.messages[messageID].Reactions)
}

func TestPhase151ReactionUsesFreshCanReactBeforeMongo(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	channelID := uuid.New()
	messageID := "phase151-reaction-authorised"
	repo := newPhase151ReactionRepository()
	channel := &ent.Channel{ID: channelID, ChannelType: "private"}
	repo.channels[channelID] = channel
	repo.messages[messageID] = &messenger.MessageDocument{ID: messageID, ChannelID: channelID, SenderID: uuid.New()}
	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	access := phase151ReadableAccess(channel, true)
	access.CanReact = true
	phase151Grant(resolver, actorID, channel, messenger.ChannelActionReact, access)
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	added, err := service.ToggleReaction(ctx, channelID, messageID, actorID, "thumbsup")

	require.NoError(t, err)
	assert.True(t, added)
	assert.Equal(t, 1, repo.findCalls)
	assert.Zero(t, repo.isMemberCalls, "fresh scoped proof replaces legacy membership")
	require.Len(t, repo.messages[messageID].Reactions, 1)
	assert.Equal(t, "thumbsup", repo.messages[messageID].Reactions[0].Emoji)
}

func TestPhase151LegacyReactionChecksMembershipBeforeMongo(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	channelID := uuid.New()
	messageID := "phase151-legacy-reaction-oracle"
	repo := newPhase151ReactionRepository()
	repo.messages[messageID] = &messenger.MessageDocument{ID: messageID, ChannelID: channelID, SenderID: uuid.New()}
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

	added, err := service.ToggleReaction(ctx, channelID, messageID, actorID, "thumbsup")

	require.Error(t, err)
	assert.False(t, added)
	assert.Equal(t, 1, repo.isMemberCalls)
	assert.Zero(t, repo.findCalls, "legacy denial must not disclose whether a Mongo message exists")
	assert.Empty(t, publisher.snapshot())
}
