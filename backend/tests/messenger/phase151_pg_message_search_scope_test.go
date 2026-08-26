package messenger_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/internal/messenger"
)

type phase151MessageSearchAuthoritySource struct {
	phase151ChannelAuthoritySource
	candidateIDs []uuid.UUID
	factsByID    map[uuid.UUID]messenger.ChannelAuthorityFacts
}

func (s *phase151MessageSearchAuthoritySource) ListMessageSearchCandidateChannelIDs(
	context.Context, uuid.UUID, uuid.UUID, int,
) ([]uuid.UUID, error) {
	return append([]uuid.UUID(nil), s.candidateIDs...), nil
}

func (s *phase151MessageSearchAuthoritySource) ResolveChannelAuthorityFacts(
	_ context.Context, _, _, channelID uuid.UUID,
) (messenger.ChannelAuthorityFacts, error) {
	return s.factsByID[channelID], nil
}

func TestPhase151PGMessageSearchScopeIncludesPublicPrejoinAndMemberPrivateOnly(t *testing.T) {
	tenantID, actorID := uuid.New(), uuid.New()
	publicID, privateID, deniedID := uuid.New(), uuid.New(), uuid.New()
	source := &phase151MessageSearchAuthoritySource{
		phase151ChannelAuthoritySource: phase151ChannelAuthoritySource{organizations: []uuid.UUID{tenantID}},
		candidateIDs: []uuid.UUID{publicID, privateID, deniedID},
		factsByID: map[uuid.UUID]messenger.ChannelAuthorityFacts{
			publicID: {
				TenantID: tenantID, ActorID: actorID, ActorOrganizationID: tenantID, ActorActive: true,
				Channel: &ent.Channel{ID: publicID, ChannelType: channel.ChannelTypePublic},
			},
			privateID: {
				TenantID: tenantID, ActorID: actorID, ActorOrganizationID: tenantID, ActorActive: true,
				Channel: &ent.Channel{ID: privateID, ChannelType: channel.ChannelTypePrivate}, IsMember: true,
				HasActiveProjectMembership: true,
			},
			deniedID: {
				TenantID: tenantID, ActorID: actorID, ActorOrganizationID: tenantID, ActorActive: true,
				Channel: &ent.Channel{ID: deniedID, ChannelType: channel.ChannelTypePrivate},
			},
		},
	}
	resolver := messenger.NewPGMessageSearchScopeResolver(source)

	scope, err := resolver.ResolveAuthorizedMessageSearchScope(context.Background(), actorID)

	require.NoError(t, err)
	assert.Equal(t, tenantID, scope.TenantID)
	assert.Equal(t, actorID, scope.UserID)
	require.Len(t, scope.Targets, 2)
	assert.Equal(t, publicID, scope.Targets[0].ChannelID)
	assert.Equal(t, privateID, scope.Targets[1].ChannelID)
}
