package messenger_test

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
)

type phase151MentionResolverPort interface {
	ResolveEligibleMentionUserIDs(context.Context, messenger.AuthorizedMessageProducerScope, messenger.PostCreatePipelineRequest) ([]uuid.UUID, error)
}

var _ phase151MentionResolverPort = (*messenger.Phase151EligibleMentionResolver)(nil)
var _ messenger.Phase151MentionAudience = (*messenger.WSChannelAuthority)(nil)
var _ messenger.Phase151MentionPresence = (*messenger.PresenceService)(nil)
var _ messenger.Phase151MentionRoleSource = (*messenger.EntRepository)(nil)

type phase151MentionAuthorityCall struct {
	actorID   uuid.UUID
	channelID uuid.UUID
	action    messenger.ChannelAction
}

type phase151MentionAuthority struct {
	access messenger.ChannelAccess
	err    error
	calls  []phase151MentionAuthorityCall
}

func (a *phase151MentionAuthority) AuthorizeChannelAccess(
	_ context.Context,
	actorID, channelID uuid.UUID,
	action messenger.ChannelAction,
) (messenger.ChannelAccess, error) {
	a.calls = append(a.calls, phase151MentionAuthorityCall{actorID: actorID, channelID: channelID, action: action})
	return a.access, a.err
}

type phase151MentionAudience struct {
	recipients []uuid.UUID
	err        error
	calls      int
	channelIDs []uuid.UUID
	events     []string
}

func (a *phase151MentionAudience) ResolveChannelEventAudience(
	_ context.Context,
	channelID uuid.UUID,
	event string,
) ([]uuid.UUID, error) {
	a.calls++
	a.channelIDs = append(a.channelIDs, channelID)
	a.events = append(a.events, event)
	return append([]uuid.UUID(nil), a.recipients...), a.err
}

type phase151MentionPresence struct {
	userIDs []string
	err     error
	calls   int
}

func (p *phase151MentionPresence) GetOnlineUserIDs(context.Context) ([]string, error) {
	p.calls++
	return append([]string(nil), p.userIDs...), p.err
}

type phase151MentionRoles struct {
	byRole     map[string][]uuid.UUID
	err        error
	calls      []string
	projectIDs []uuid.UUID
}

func (r *phase151MentionRoles) ListProjectMembersByRole(
	_ context.Context,
	projectID uuid.UUID,
	role string,
) ([]uuid.UUID, error) {
	r.calls = append(r.calls, role)
	r.projectIDs = append(r.projectIDs, projectID)
	return append([]uuid.UUID(nil), r.byRole[role]...), r.err
}

func phase151MentionFixture() (
	messenger.AuthorizedMessageProducerScope,
	messenger.PostCreatePipelineRequest,
	*phase151MentionAuthority,
	*phase151MentionAudience,
	*phase151MentionPresence,
	*phase151MentionRoles,
) {
	tenantID, actorID, channelID, projectID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	scope := messenger.AuthorizedMessageProducerScope{
		TenantID: tenantID, ActorID: actorID, ChannelID: channelID,
		Producer: messenger.TrustedMessageProducerUser,
	}
	request := messenger.PostCreatePipelineRequest{
		Producer: messenger.TrustedMessageProducerUser,
		Message: messenger.MessageDocument{
			ChannelID: channelID, SenderID: actorID, ClientNonce: "mentions-0001",
		},
	}
	authority := &phase151MentionAuthority{access: messenger.ChannelAccess{
		TenantID: tenantID,
		Channel:  &ent.Channel{ID: channelID, ProjectID: &projectID},
	}}
	return scope, request, authority, &phase151MentionAudience{}, &phase151MentionPresence{}, &phase151MentionRoles{byRole: map[string][]uuid.UUID{}}
}

func phase151NewMentionResolver(
	authority *phase151MentionAuthority,
	audience *phase151MentionAudience,
	presence *phase151MentionPresence,
	roles *phase151MentionRoles,
) *messenger.Phase151EligibleMentionResolver {
	return messenger.NewPhase151EligibleMentionResolver(authority, audience, presence, roles)
}

func phase151SortedUUIDs(ids ...uuid.UUID) []uuid.UUID {
	result := append([]uuid.UUID(nil), ids...)
	sort.Slice(result, func(i, j int) bool { return result[i].String() < result[j].String() })
	return result
}

func TestPhase151EligibleMentionsIntersectIndividualsAndAllWithFreshAudience(t *testing.T) {
	scope, request, authority, audience, presence, roles := phase151MentionFixture()
	individual, allOnly, foreign := uuid.New(), uuid.New(), uuid.New()
	request.Message.Mentions = []string{foreign.String(), individual.String(), individual.String(), scope.ActorID.String(), "bad"}
	request.Message.MentionGroups = []string{"all", "all", "unknown"}
	audience.recipients = []uuid.UUID{scope.ActorID, allOnly, uuid.Nil, individual, allOnly}

	resolved, err := phase151NewMentionResolver(authority, audience, presence, roles).
		ResolveEligibleMentionUserIDs(context.Background(), scope, request)

	require.NoError(t, err)
	assert.Equal(t, phase151SortedUUIDs(individual, allOnly), resolved)
	assert.Equal(t, []phase151MentionAuthorityCall{{
		actorID: scope.ActorID, channelID: scope.ChannelID, action: messenger.ChannelActionReceiveEvent,
	}}, authority.calls)
	assert.Equal(t, 1, audience.calls)
	assert.Equal(t, []uuid.UUID{scope.ChannelID}, audience.channelIDs)
	assert.Equal(t, []string{messenger.UserNewMessageNotificationEvent}, audience.events)
	assert.Zero(t, presence.calls)
	assert.Empty(t, roles.calls)
}

func TestPhase151EligibleMentionsIntersectOnlineAndProjectRolesWithAudience(t *testing.T) {
	scope, request, authority, audience, presence, roles := phase151MentionFixture()
	online, developer, foreign := uuid.New(), uuid.New(), uuid.New()
	request.Message.MentionGroups = []string{"online", "developer", "online", "developer", "viewer"}
	audience.recipients = []uuid.UUID{scope.ActorID, developer, online}
	presence.userIDs = []string{foreign.String(), online.String(), "bad", online.String()}
	roles.byRole["developer"] = []uuid.UUID{developer, foreign, developer}
	roles.byRole["viewer"] = []uuid.UUID{foreign}

	resolved, err := phase151NewMentionResolver(authority, audience, presence, roles).
		ResolveEligibleMentionUserIDs(context.Background(), scope, request)

	require.NoError(t, err)
	assert.Equal(t, phase151SortedUUIDs(online, developer), resolved)
	assert.Equal(t, 1, presence.calls, "duplicate @online reads presence once")
	assert.Equal(t, []string{"developer", "viewer"}, roles.calls)
	require.NotNil(t, authority.access.Channel.ProjectID)
	assert.Equal(t, []uuid.UUID{*authority.access.Channel.ProjectID, *authority.access.Channel.ProjectID}, roles.projectIDs)
}

func TestPhase151EligibleMentionsSkipProjectRolesOutsideProjectChannels(t *testing.T) {
	scope, request, authority, audience, presence, roles := phase151MentionFixture()
	direct := uuid.New()
	authority.access.Channel.ProjectID = nil
	request.Message.Mentions = []string{direct.String()}
	request.Message.MentionGroups = []string{"developer"}
	audience.recipients = []uuid.UUID{scope.ActorID, direct}

	resolved, err := phase151NewMentionResolver(authority, audience, presence, roles).
		ResolveEligibleMentionUserIDs(context.Background(), scope, request)

	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{direct}, resolved)
	assert.Empty(t, roles.calls)
}

func TestPhase151EligibleMentionsFailClosedOnAuthorityAndProjectionErrors(t *testing.T) {
	dependencyErr := errors.New("dependency unavailable")
	for _, scenario := range []struct {
		name   string
		mutate func(*phase151MentionAuthority, *phase151MentionAudience, *phase151MentionPresence, *phase151MentionRoles)
		group  string
	}{
		{name: "authority", mutate: func(a *phase151MentionAuthority, _ *phase151MentionAudience, _ *phase151MentionPresence, _ *phase151MentionRoles) {
			a.err = dependencyErr
		}},
		{name: "audience", mutate: func(_ *phase151MentionAuthority, a *phase151MentionAudience, _ *phase151MentionPresence, _ *phase151MentionRoles) {
			a.err = dependencyErr
		}},
		{name: "presence", group: "online", mutate: func(_ *phase151MentionAuthority, _ *phase151MentionAudience, p *phase151MentionPresence, _ *phase151MentionRoles) {
			p.err = dependencyErr
		}},
		{name: "role", group: "developer", mutate: func(_ *phase151MentionAuthority, _ *phase151MentionAudience, _ *phase151MentionPresence, r *phase151MentionRoles) {
			r.err = dependencyErr
		}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			scope, request, authority, audience, presence, roles := phase151MentionFixture()
			candidate := uuid.New()
			request.Message.Mentions = []string{candidate.String()}
			if scenario.group != "" {
				request.Message.MentionGroups = []string{scenario.group}
			}
			audience.recipients = []uuid.UUID{scope.ActorID, candidate}
			scenario.mutate(authority, audience, presence, roles)

			resolved, err := phase151NewMentionResolver(authority, audience, presence, roles).
				ResolveEligibleMentionUserIDs(context.Background(), scope, request)

			require.Error(t, err)
			assert.Nil(t, resolved)
		})
	}
}

func TestPhase151EligibleMentionsRejectInvalidAuthorityFactsAndRequestBeforeProjection(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		mutate func(*messenger.AuthorizedMessageProducerScope, *messenger.PostCreatePipelineRequest, *phase151MentionAuthority)
	}{
		{name: "unknown producer", mutate: func(scope *messenger.AuthorizedMessageProducerScope, _ *messenger.PostCreatePipelineRequest, _ *phase151MentionAuthority) {
			scope.Producer = messenger.TrustedMessageProducer("unknown")
		}},
		{name: "foreign message channel", mutate: func(_ *messenger.AuthorizedMessageProducerScope, request *messenger.PostCreatePipelineRequest, _ *phase151MentionAuthority) {
			request.Message.ChannelID = uuid.New()
		}},
		{name: "foreign tenant proof", mutate: func(_ *messenger.AuthorizedMessageProducerScope, _ *messenger.PostCreatePipelineRequest, authority *phase151MentionAuthority) {
			authority.access.TenantID = uuid.New()
		}},
		{name: "foreign channel proof", mutate: func(_ *messenger.AuthorizedMessageProducerScope, _ *messenger.PostCreatePipelineRequest, authority *phase151MentionAuthority) {
			authority.access.Channel.ID = uuid.New()
		}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			scope, request, authority, audience, presence, roles := phase151MentionFixture()
			request.Message.MentionGroups = []string{"all"}
			audience.recipients = []uuid.UUID{scope.ActorID, uuid.New()}
			scenario.mutate(&scope, &request, authority)

			resolved, err := phase151NewMentionResolver(authority, audience, presence, roles).
				ResolveEligibleMentionUserIDs(context.Background(), scope, request)

			require.Error(t, err)
			assert.Nil(t, resolved)
			if scenario.name == "unknown producer" || scenario.name == "foreign message channel" {
				assert.Empty(t, authority.calls)
			}
			assert.Zero(t, audience.calls)
		})
	}

	t.Run("sender absent from fresh audience", func(t *testing.T) {
		scope, request, authority, audience, presence, roles := phase151MentionFixture()
		request.Message.MentionGroups = []string{"all"}
		audience.recipients = []uuid.UUID{uuid.New()}

		resolved, err := phase151NewMentionResolver(authority, audience, presence, roles).
			ResolveEligibleMentionUserIDs(context.Background(), scope, request)

		require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
		assert.Nil(t, resolved)
		assert.Equal(t, 1, audience.calls)
		assert.Zero(t, presence.calls)
		assert.Empty(t, roles.calls)
	})
}

func TestPhase151EligibleMentionsRequireDependenciesAndSkipReadsWhenEmpty(t *testing.T) {
	scope, request, authority, audience, presence, roles := phase151MentionFixture()
	assert.Panics(t, func() { messenger.NewPhase151EligibleMentionResolver(nil, audience, presence, roles) })
	assert.Panics(t, func() { messenger.NewPhase151EligibleMentionResolver(authority, nil, presence, roles) })
	assert.Panics(t, func() { messenger.NewPhase151EligibleMentionResolver(authority, audience, nil, roles) })
	assert.Panics(t, func() { messenger.NewPhase151EligibleMentionResolver(authority, audience, presence, nil) })

	resolver := phase151NewMentionResolver(authority, audience, presence, roles)
	resolved, err := resolver.ResolveEligibleMentionUserIDs(context.Background(), scope, request)
	require.NoError(t, err)
	assert.Empty(t, resolved)
	assert.Empty(t, authority.calls)
	assert.Zero(t, audience.calls)
	assert.Zero(t, presence.calls)
	assert.Empty(t, roles.calls)

	request.Message.Mentions = []string{uuid.NewString()}
	resolved, err = resolver.ResolveEligibleMentionUserIDs(nil, scope, request)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	assert.Nil(t, resolved)

	cancelledContext, cancel := context.WithCancel(context.Background())
	cancel()
	resolved, err = resolver.ResolveEligibleMentionUserIDs(cancelledContext, scope, request)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	assert.Nil(t, resolved)
}
