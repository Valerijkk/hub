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

type phase151ForwardLegacyRepository struct {
	*phase151MutationRepository
	membershipCalls []uuid.UUID
}

func newPhase151ForwardLegacyRepository() *phase151ForwardLegacyRepository {
	return &phase151ForwardLegacyRepository{phase151MutationRepository: newPhase151MutationRepository()}
}

func (r *phase151ForwardLegacyRepository) IsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	r.membershipCalls = append(r.membershipCalls, channelID)
	return r.mockRepository.IsMember(ctx, channelID, userID)
}

func TestPhase151LegacyForwardChecksSourceMembershipBeforeMongo(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	sourceChannelID := uuid.New()
	targetChannelID := uuid.New()
	sourceMessageID := "phase151-legacy-source-oracle"
	repo := newPhase151ForwardLegacyRepository()
	repo.messages[sourceMessageID] = &messenger.MessageDocument{ID: sourceMessageID, ChannelID: sourceChannelID, SenderID: uuid.New()}
	repo.members[targetChannelID] = map[uuid.UUID]*ent.ChannelMember{actorID: {UserID: actorID, ChannelID: targetChannelID}}
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

	forwarded, err := service.ForwardMessage(ctx, sourceChannelID, actorID, sourceMessageID, targetChannelID.String())

	require.Error(t, err)
	assert.Nil(t, forwarded)
	assert.Equal(t, []uuid.UUID{sourceChannelID}, repo.membershipCalls)
	assert.Zero(t, repo.findCalls, "legacy source denial must not disclose whether the source message exists")
	assert.Zero(t, repo.insertCalls)
	assert.Empty(t, publisher.snapshot())
}

func TestPhase151LegacyForwardChecksTargetMembershipBeforeMongo(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	sourceChannelID := uuid.New()
	targetChannelID := uuid.New()
	sourceMessageID := "phase151-legacy-target-oracle"
	repo := newPhase151ForwardLegacyRepository()
	repo.messages[sourceMessageID] = &messenger.MessageDocument{ID: sourceMessageID, ChannelID: sourceChannelID, SenderID: uuid.New()}
	repo.channels[sourceChannelID] = &ent.Channel{ID: sourceChannelID, ChannelType: "private"}
	repo.members[sourceChannelID] = map[uuid.UUID]*ent.ChannelMember{actorID: {UserID: actorID, ChannelID: sourceChannelID}}
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

	forwarded, err := service.ForwardMessage(ctx, sourceChannelID, actorID, sourceMessageID, targetChannelID.String())

	require.Error(t, err)
	assert.Nil(t, forwarded)
	assert.Equal(t, []uuid.UUID{sourceChannelID, targetChannelID}, repo.membershipCalls)
	assert.Zero(t, repo.findCalls, "legacy target denial must not disclose whether the source message exists")
	assert.Zero(t, repo.insertCalls)
	assert.Empty(t, publisher.snapshot())
}
