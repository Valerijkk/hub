package messenger_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

func TestPhase151AuthorizedUnreadScopeIncludesOnlyFreshEligibleMembershipTargets(t *testing.T) {
	tenantID := uuid.New()
	actorID := uuid.New()
	publicMemberID := uuid.New()
	privateMemberID := uuid.New()
	directMemberID := uuid.New()
	groupMemberID := uuid.New()

	scope, err := messenger.ExportPhase151ResolveAuthorizedUnreadScopeFromFacts(
		messenger.AuthorizedUnreadScopeFacts{
			TenantID:            tenantID,
			ActorID:             actorID,
			ActorOrganizationID: tenantID,
			ActorActive:         true,
			MembershipTargetFacts: []messenger.AuthorizedUnreadMembershipTargetFacts{
				{ChannelID: publicMemberID, ChannelType: "public", IsCurrentMember: true, IsChannelActive: true},
				{ChannelID: privateMemberID, ChannelType: "private", IsCurrentMember: true, IsChannelActive: true},
				{ChannelID: directMemberID, ChannelType: "dm", IsCurrentMember: true, IsChannelActive: true},
				{ChannelID: groupMemberID, ChannelType: "group", IsCurrentMember: true, IsChannelActive: true},
				{ChannelID: uuid.New(), ChannelType: "public", IsChannelActive: true},  // public pre-join
				{ChannelID: uuid.New(), ChannelType: "private", IsChannelActive: true}, // revoked membership
				{ChannelID: uuid.New(), ChannelType: "public", IsCurrentMember: true},  // archived/deleted channel
				{
					ChannelID:                  uuid.New(),
					ChannelType:                "public",
					IsCurrentMember:            true,
					IsChannelActive:            true,
					IsProjectChannel:           true,
					HasActiveProjectMembership: true,
					IsProjectActive:            false,
				},
				{
					ChannelID:        uuid.New(),
					ChannelType:      "private",
					IsCurrentMember:  true,
					IsChannelActive:  true,
					IsProjectChannel: true,
					IsProjectActive:  true,
				},
			},
		},
	)

	require.NoError(t, err)
	assert.Equal(t, tenantID, scope.TenantID)
	assert.Equal(t, actorID, scope.UserID)
	assert.ElementsMatch(t, []messenger.AuthorizedUnreadTarget{
		{ChannelID: publicMemberID, Kind: "channel"},
		{ChannelID: privateMemberID, Kind: "channel"},
		{ChannelID: directMemberID, Kind: "dm"},
		{ChannelID: groupMemberID, Kind: "dm"},
	}, scope.Targets)
}

func TestPhase151AuthorizedUnreadScopeFailsClosedForInvalidActorAuthority(t *testing.T) {
	tenantID := uuid.New()
	actorID := uuid.New()

	for name, facts := range map[string]messenger.AuthorizedUnreadScopeFacts{
		"disabled actor": {
			TenantID:            tenantID,
			ActorID:             actorID,
			ActorOrganizationID: tenantID,
		},
		"cross tenant actor": {
			TenantID:            tenantID,
			ActorID:             actorID,
			ActorOrganizationID: uuid.New(),
			ActorActive:         true,
		},
		"missing tenant": {
			ActorID:             actorID,
			ActorOrganizationID: tenantID,
			ActorActive:         true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			scope, err := messenger.ExportPhase151ResolveAuthorizedUnreadScopeFromFacts(facts)

			require.ErrorIs(t, err, messenger.ErrAuthorizedUnreadScopeDenied)
			assert.Equal(t, messenger.AuthorizedUnreadScope{}, scope)
		})
	}
}

func TestPhase151AuthorizedUnreadScopeAllowsActiveUserWithNoEligibleTargets(t *testing.T) {
	tenantID := uuid.New()
	actorID := uuid.New()

	scope, err := messenger.ExportPhase151ResolveAuthorizedUnreadScopeFromFacts(
		messenger.AuthorizedUnreadScopeFacts{
			TenantID:            tenantID,
			ActorID:             actorID,
			ActorOrganizationID: tenantID,
			ActorActive:         true,
			MembershipTargetFacts: []messenger.AuthorizedUnreadMembershipTargetFacts{
				{ChannelID: uuid.New(), ChannelType: "public", IsChannelActive: true},
			},
		},
	)

	require.NoError(t, err)
	assert.Equal(t, tenantID, scope.TenantID)
	assert.Equal(t, actorID, scope.UserID)
	assert.Empty(t, scope.Targets)
}
