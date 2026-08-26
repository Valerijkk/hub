package messenger_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/internal/messenger"
)

type phase151ChannelAuthoritySource struct {
	organizations []uuid.UUID
	facts         messenger.ChannelAuthorityFacts
	factsErr      error
	factCalls     int
}

func (s *phase151ChannelAuthoritySource) ListInstanceOrganizationIDs(context.Context) ([]uuid.UUID, error) {
	return append([]uuid.UUID(nil), s.organizations...), nil
}

func (s *phase151ChannelAuthoritySource) ResolveChannelAuthorityFacts(
	_ context.Context,
	_, _, _ uuid.UUID,
) (messenger.ChannelAuthorityFacts, error) {
	s.factCalls++
	return s.facts, s.factsErr
}

func TestPhase151PGChannelAccessResolverBindsRequestedActorChannelAndTenant(t *testing.T) {
	tenantID := uuid.New()
	actorID := uuid.New()
	channelID := uuid.New()
	source := &phase151ChannelAuthoritySource{
		organizations: []uuid.UUID{tenantID},
		facts: messenger.ChannelAuthorityFacts{
			TenantID: tenantID, ActorID: actorID, ActorOrganizationID: tenantID, ActorActive: true,
			Channel: &ent.Channel{ID: channelID, ChannelType: channel.ChannelTypePrivate},
			IsMember: true, Role: "member", HasActiveProjectMembership: true,
		},
	}
	resolver := messenger.NewPGChannelAccessResolver(source)

	access, err := resolver.AuthorizeChannelAccess(context.Background(), actorID, channelID, messenger.ChannelActionReadTimeline)

	require.NoError(t, err)
	require.NotNil(t, access.Channel)
	assert.Equal(t, channelID, access.Channel.ID)
	assert.True(t, access.CanRead)
	assert.Equal(t, 1, source.factCalls)
}

func TestPhase151PGChannelAccessResolverFailsClosedBeforeCapabilities(t *testing.T) {
	tenantID := uuid.New()
	actorID := uuid.New()
	channelID := uuid.New()

	for _, tc := range []struct {
		name  string
		orgs  []uuid.UUID
		facts messenger.ChannelAuthorityFacts
		err   error
	}{
		{name: "ambiguous tenant", orgs: []uuid.UUID{tenantID, uuid.New()}},
		{name: "fact query error", orgs: []uuid.UUID{tenantID}, err: errors.New("postgres unavailable")},
		{name: "actor mismatch", orgs: []uuid.UUID{tenantID}, facts: messenger.ChannelAuthorityFacts{TenantID: tenantID, ActorID: uuid.New(), Channel: &ent.Channel{ID: channelID}}},
		{name: "channel mismatch", orgs: []uuid.UUID{tenantID}, facts: messenger.ChannelAuthorityFacts{TenantID: tenantID, ActorID: actorID, Channel: &ent.Channel{ID: uuid.New()}}},
		{name: "tenant mismatch", orgs: []uuid.UUID{tenantID}, facts: messenger.ChannelAuthorityFacts{TenantID: uuid.New(), ActorID: actorID, Channel: &ent.Channel{ID: channelID}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := &phase151ChannelAuthoritySource{organizations: tc.orgs, facts: tc.facts, factsErr: tc.err}
			resolver := messenger.NewPGChannelAccessResolver(source)

			access, err := resolver.AuthorizeChannelAccess(context.Background(), actorID, channelID, messenger.ChannelActionReadTimeline)

			require.Error(t, err)
			assert.Equal(t, messenger.ChannelAccess{}, access)
		})
	}
}

func TestPhase151PGChannelAccessResolverRejectsNilDependency(t *testing.T) {
	assert.Panics(t, func() { messenger.NewPGChannelAccessResolver(nil) })
}
