package messenger_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

type phase151EmptyUnreadRepository struct {
	*mockRepository
	coordinatorCalls int
}

func (r *phase151EmptyUnreadRepository) ListUnreadCandidates(
	_ context.Context,
	_ messenger.AuthorizedUnreadScope,
	_ messenger.UnreadCategory,
) ([]messenger.UnreadCandidate, error) {
	r.coordinatorCalls++
	return nil, nil
}

func TestPhase151EmptyAuthorizedUnreadScopeReturnsEmptySnapshotWithoutMongo(t *testing.T) {
	tenantID := uuid.New()
	actorID := uuid.New()
	repo := &phase151EmptyUnreadRepository{mockRepository: newMockRepository()}
	scope := &phase151UnreadScopeFixture{scope: messenger.AuthorizedUnreadScope{
		TenantID: tenantID,
		UserID:   actorID,
		Targets:  []messenger.AuthorizedUnreadTarget{},
	}}
	service := messenger.ExportNewMessengerServiceWithChannelAccessAndUnreadScopeTest(
		repo,
		phase151Allow(messenger.ChannelActionReadTimeline, tenantID, uuid.New()),
		scope,
	)

	response, err := service.GetUnreadResponse(context.Background(), actorID.String())

	require.NoError(t, err)
	assert.Empty(t, response.Channels)
	assert.Empty(t, response.DMs)
	assert.Empty(t, response.Mentions)
	assert.Empty(t, response.Threads)
	assert.False(t, response.AnyUnread)
	assert.Zero(t, repo.coordinatorCalls, "an empty authorized scope must not enter Mongo")
}
