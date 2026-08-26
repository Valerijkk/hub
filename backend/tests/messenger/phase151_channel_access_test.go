package messenger_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
)

var _ messenger.ChannelAccessResolver = (*phase151ChannelAccessFixture)(nil)

// phase151ChannelAccessFixture is deliberately explicit: Phase 151 test
// construction must never receive an implicit allow-all or nil resolver.
type phase151ChannelAccessFixture struct {
	accessByAction map[messenger.ChannelAction]messenger.ChannelAccess
	errByAction    map[messenger.ChannelAction]error
	calls          []messenger.ChannelAction
}

func (f *phase151ChannelAccessFixture) AuthorizeChannelAccess(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	action messenger.ChannelAction,
) (messenger.ChannelAccess, error) {
	f.calls = append(f.calls, action)
	if err := f.errByAction[action]; err != nil {
		return messenger.ChannelAccess{}, err
	}
	return f.accessByAction[action], nil
}

func phase151Allow(action messenger.ChannelAction, tenantID, channelID uuid.UUID) *phase151ChannelAccessFixture {
	return &phase151ChannelAccessFixture{
		accessByAction: map[messenger.ChannelAction]messenger.ChannelAccess{
			action: {
				TenantID:                   tenantID,
				Channel:                    &ent.Channel{ID: channelID},
				Role:                       "owner",
				IsMember:                   true,
				HasActiveProjectMembership: true,
				CanRead:                    true,
				CanPost:                    true,
				CanReply:                   true,
				CanJoin:                    true,
				CanFollow:                  true,
				CanUnfollow:                true,
				CanViewUnread:              true,
				CanDeleteMessage:           true,
				CanPin:                     true,
				CanListPins:                true,
			},
		},
		errByAction: make(map[messenger.ChannelAction]error),
	}
}

func phase151Deny(action messenger.ChannelAction) *phase151ChannelAccessFixture {
	return &phase151ChannelAccessFixture{
		accessByAction: make(map[messenger.ChannelAction]messenger.ChannelAccess),
		errByAction: map[messenger.ChannelAction]error{
			action: errors.New("phase151 channel access denied"),
		},
	}
}

func TestPhase151ChannelAccessClosedActionContract(t *testing.T) {
	// This list is intentionally exhaustive for the closed Phase 151 action
	// vocabulary. A new action must be consciously authorised rather than
	// becoming a permissive default branch.
	actions := []messenger.ChannelAction{
		messenger.ChannelActionReadTimeline,
		messenger.ChannelActionReadThread,
		messenger.ChannelActionReadParent,
		messenger.ChannelActionSend,
		messenger.ChannelActionReply,
		messenger.ChannelActionTyping,
		messenger.ChannelActionEditMessage,
		messenger.ChannelActionForwardSource,
		messenger.ChannelActionForwardTarget,
		messenger.ChannelActionDeleteMessage,
		messenger.ChannelActionPin,
		messenger.ChannelActionListPins,
		messenger.ChannelActionReact,
		messenger.ChannelActionBookmark,
		messenger.ChannelActionPromote,
		messenger.ChannelActionReceiveEvent,
	}

	assert.Len(t, actions, 16)
	assert.NotContains(t, actions, messenger.ChannelAction(""))
}

func TestPhase151ChannelAccessCapabilityMatrixIncludesAuthorityFacts(t *testing.T) {
	tenantID := uuid.New()
	channelID := uuid.New()
	fixture := phase151Allow(messenger.ChannelActionSend, tenantID, channelID)

	access, err := fixture.AuthorizeChannelAccess(
		context.Background(),
		uuid.New(),
		uuid.New(),
		messenger.ChannelActionSend,
	)

	require.NoError(t, err)
	assert.Equal(t, tenantID, access.TenantID)
	require.NotNil(t, access.Channel)
	assert.Equal(t, channelID, access.Channel.ID)
	assert.Equal(t, "owner", access.Role)
	assert.True(t, access.IsMember)
	assert.True(t, access.HasActiveProjectMembership)
	assert.True(t, access.CanRead)
	assert.True(t, access.CanPost)
	assert.True(t, access.CanReply)
	assert.True(t, access.CanJoin)
	assert.True(t, access.CanFollow)
	assert.True(t, access.CanUnfollow)
	assert.True(t, access.CanViewUnread)
	assert.True(t, access.CanDeleteMessage)
	assert.True(t, access.CanPin)
	assert.True(t, access.CanListPins)
}

func TestPhase151ReadonlyMemberCannotPostOrReplyButOwnerCan(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	memberID := uuid.New()
	ownerID := uuid.New()
	tenantID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", true, ownerID, memberID)
	rootID := uuid.NewString()
	repo.messages[rootID] = &messenger.MessageDocument{
		ID:        rootID,
		ChannelID: channelID,
		SenderID:  ownerID,
	}

	memberFixture := phase151Allow(messenger.ChannelActionSend, tenantID, channelID)
	memberAccess := memberFixture.accessByAction[messenger.ChannelActionSend]
	memberAccess.Channel.IsReadonly = true
	memberAccess.Role = "member"
	memberAccess.CanPost = false
	memberAccess.CanReply = false
	memberFixture.accessByAction[messenger.ChannelActionSend] = memberAccess
	memberFixture.accessByAction[messenger.ChannelActionReply] = memberAccess
	memberService := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, memberFixture)

	post, err := memberService.SendMessage(ctx, channelID, memberID, messenger.SendMessageRequest{ContentText: "blocked"})
	require.Error(t, err)
	assert.Nil(t, post)
	reply, err := memberService.SendMessage(ctx, channelID, memberID, messenger.SendMessageRequest{ContentText: "blocked", ThreadID: rootID})
	require.Error(t, err)
	assert.Nil(t, reply)
	assert.Len(t, repo.messages, 1, "read-only denial must happen before Mongo insert")
	assert.Equal(t, []messenger.ChannelAction{
		messenger.ChannelActionSend,
		messenger.ChannelActionReply,
	}, memberFixture.calls)

	ownerFixture := phase151Allow(messenger.ChannelActionReply, tenantID, channelID)
	ownerAccess := ownerFixture.accessByAction[messenger.ChannelActionReply]
	ownerAccess.Channel.IsReadonly = true
	ownerAccess.Role = "owner"
	ownerFixture.accessByAction[messenger.ChannelActionSend] = ownerAccess
	ownerFixture.accessByAction[messenger.ChannelActionReply] = ownerAccess
	ownerService := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, ownerFixture)

	post, err = ownerService.SendMessage(ctx, channelID, ownerID, messenger.SendMessageRequest{ContentText: "allowed"})
	require.NoError(t, err)
	require.NotNil(t, post)
	reply, err = ownerService.SendMessage(ctx, channelID, ownerID, messenger.SendMessageRequest{ContentText: "allowed", ThreadID: rootID})
	require.NoError(t, err)
	require.NotNil(t, reply)
	assert.Equal(t, []messenger.ChannelAction{
		messenger.ChannelActionSend,
		messenger.ChannelActionReply,
	}, ownerFixture.calls)
}

func TestPhase151ExplicitTestConstructorRejectsNilChannelAccessResolver(t *testing.T) {
	repo := newMockRepository()

	assert.Panics(t, func() {
		messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, nil)
	})
}

func TestPhase151DeniedDeleteStopsBeforeMessageMutation(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	actorID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, actorID)
	messageID := uuid.NewString()
	repo.messages[messageID] = &messenger.MessageDocument{
		ID:        messageID,
		ChannelID: channelID,
		SenderID:  actorID,
	}
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(
		repo,
		phase151Deny(messenger.ChannelActionDeleteMessage),
	)

	err := service.DeleteMessage(ctx, messageID, actorID, channelID.String(), false)

	require.Error(t, err)
	assert.Nil(t, repo.messages[messageID].DeletedAt, "denied Stage 1 must not mutate a message")
}

func TestPhase151DeniedPinAndListPinsStopBeforePinDisclosure(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	actorID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, actorID)
	messageID := uuid.NewString()
	repo.messages[messageID] = &messenger.MessageDocument{
		ID:        messageID,
		ChannelID: channelID,
		SenderID:  actorID,
	}

	deniedPin := messenger.ExportNewMessengerServiceWithChannelAccessTest(
		repo,
		phase151Deny(messenger.ChannelActionPin),
	)
	pinned, err := deniedPin.TogglePin(ctx, channelID, messageID, actorID)
	require.Error(t, err)
	assert.False(t, pinned)
	assert.Empty(t, repo.pins, "denied pin must not write a pin record")

	deniedList := messenger.ExportNewMessengerServiceWithChannelAccessTest(
		repo,
		phase151Deny(messenger.ChannelActionListPins),
	)
	pins, err := deniedList.ListPins(ctx, channelID, actorID)
	require.Error(t, err)
	assert.Nil(t, pins, "denied list-pin must not disclose pinned metadata")
}

func TestPhase151AllowedDeletePinAndListPinsRequireDistinctCapabilities(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	actorID := uuid.New()
	tenantID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, actorID)
	messageID := uuid.NewString()
	repo.messages[messageID] = &messenger.MessageDocument{
		ID:        messageID,
		ChannelID: channelID,
		SenderID:  actorID,
	}

	deleteFixture := phase151Allow(messenger.ChannelActionDeleteMessage, tenantID, channelID)
	deleteService := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, deleteFixture)
	require.NoError(t, deleteService.DeleteMessage(ctx, messageID, actorID, channelID.String(), false))
	assert.NotNil(t, repo.messages[messageID].DeletedAt)
	assert.Equal(t, []messenger.ChannelAction{messenger.ChannelActionDeleteMessage}, deleteFixture.calls)

	visibleMessageID := uuid.NewString()
	repo.messages[visibleMessageID] = &messenger.MessageDocument{
		ID:        visibleMessageID,
		ChannelID: channelID,
		SenderID:  actorID,
	}
	pinFixture := phase151Allow(messenger.ChannelActionPin, tenantID, channelID)
	pinService := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, pinFixture)
	pinned, err := pinService.TogglePin(ctx, channelID, visibleMessageID, actorID)
	require.NoError(t, err)
	assert.True(t, pinned)
	assert.Equal(t, []messenger.ChannelAction{messenger.ChannelActionPin}, pinFixture.calls)

	listFixture := phase151Allow(messenger.ChannelActionListPins, tenantID, channelID)
	listService := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, listFixture)
	pins, err := listService.ListPins(ctx, channelID, actorID)
	require.NoError(t, err)
	assert.Len(t, pins, 1)
	assert.Equal(t, []messenger.ChannelAction{messenger.ChannelActionListPins}, listFixture.calls)
}

type phase151InstanceOrganizationSource struct {
	ids   []uuid.UUID
	err   error
	calls int
}

var _ messenger.InstanceOrganizationIDSource = (*phase151InstanceOrganizationSource)(nil)

func (s *phase151InstanceOrganizationSource) ListInstanceOrganizationIDs(
	_ context.Context,
) ([]uuid.UUID, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return append([]uuid.UUID(nil), s.ids...), nil
}

func TestPhase151ResolveInstanceOrganizationIDUsesExactlyOneServerSideOrganization(t *testing.T) {
	organizationID := uuid.New()
	source := &phase151InstanceOrganizationSource{ids: []uuid.UUID{organizationID}}

	resolved, err := messenger.ExportPhase151ResolveInstanceOrganizationID(context.Background(), source)

	require.NoError(t, err)
	assert.Equal(t, organizationID, resolved)
	assert.Equal(t, 1, source.calls)
}

func TestPhase151ResolveInstanceOrganizationIDFailsClosedForZeroOrMultiple(t *testing.T) {
	for _, tt := range []struct {
		name string
		ids  []uuid.UUID
	}{
		{name: "zero organizations"},
		{name: "multiple organizations", ids: []uuid.UUID{uuid.New(), uuid.New()}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := &phase151InstanceOrganizationSource{ids: tt.ids}
			resolved, err := messenger.ExportPhase151ResolveInstanceOrganizationID(context.Background(), source)
			require.Error(t, err)
			assert.Equal(t, uuid.Nil, resolved)
			assert.Equal(t, 1, source.calls)
		})
	}
}

func TestPhase151ChannelAccessAuthorityMatrix(t *testing.T) {
	tenantID := uuid.New()
	actorID := uuid.New()
	projectID := uuid.New()

	base := messenger.ChannelAuthorityFacts{
		TenantID:                     tenantID,
		ActorID:                      actorID,
		ActorOrganizationID:          tenantID,
		ActorActive:                  true,
		IsMember:                    true,
		Role:                         "member",
		HasActiveProjectMembership:   true,
	}

	t.Run("public same-tenant prejoin exposes read and join only", func(t *testing.T) {
		facts := base
		facts.Channel = &ent.Channel{ID: uuid.New(), ChannelType: "public"}
		facts.IsMember = false
		facts.Role = ""

		access, err := messenger.ExportPhase151ResolveChannelAccessFromFacts(facts, messenger.ChannelActionReadTimeline)
		require.NoError(t, err)
		assert.True(t, access.CanRead)
		assert.True(t, access.CanJoin)
		assert.False(t, access.CanPost)
		assert.False(t, access.CanReply)
		assert.False(t, access.CanFollow)
		assert.False(t, access.CanUnfollow)
		assert.False(t, access.CanViewUnread)
	})

	t.Run("private nonmember is non-disclosive", func(t *testing.T) {
		facts := base
		facts.Channel = &ent.Channel{ID: uuid.New(), ChannelType: "private"}
		facts.IsMember = false
		facts.Role = ""

		access, err := messenger.ExportPhase151ResolveChannelAccessFromFacts(facts, messenger.ChannelActionReadTimeline)
		require.Error(t, err)
		assert.False(t, access.CanRead)
		assert.Nil(t, access.Channel)
	})

	t.Run("disabled account fails before capability projection", func(t *testing.T) {
		facts := base
		facts.Channel = &ent.Channel{ID: uuid.New(), ChannelType: "public"}
		facts.ActorActive = false

		access, err := messenger.ExportPhase151ResolveChannelAccessFromFacts(facts, messenger.ChannelActionReadTimeline)
		require.Error(t, err)
		assert.False(t, access.CanRead)
		assert.Nil(t, access.Channel)
	})

	t.Run("cross-tenant actor cannot select tenant from caller data", func(t *testing.T) {
		facts := base
		facts.Channel = &ent.Channel{ID: uuid.New(), ChannelType: "public"}
		facts.ActorOrganizationID = uuid.New()

		access, err := messenger.ExportPhase151ResolveChannelAccessFromFacts(facts, messenger.ChannelActionReadTimeline)
		require.Error(t, err)
		assert.Equal(t, uuid.Nil, access.TenantID)
		assert.Nil(t, access.Channel)
	})

	t.Run("project channel requires current active project membership", func(t *testing.T) {
		facts := base
		facts.Channel = &ent.Channel{ID: uuid.New(), ChannelType: "private", ProjectID: &projectID}
		facts.HasActiveProjectMembership = false

		access, err := messenger.ExportPhase151ResolveChannelAccessFromFacts(facts, messenger.ChannelActionReadTimeline)
		require.Error(t, err)
		assert.False(t, access.CanRead)
		assert.Nil(t, access.Channel)
	})

	t.Run("active project member retains channel capabilities", func(t *testing.T) {
		facts := base
		facts.Channel = &ent.Channel{ID: uuid.New(), ChannelType: "private", ProjectID: &projectID}

		access, err := messenger.ExportPhase151ResolveChannelAccessFromFacts(facts, messenger.ChannelActionReadTimeline)
		require.NoError(t, err)
		assert.True(t, access.CanRead)
		assert.True(t, access.HasActiveProjectMembership)
		assert.Equal(t, facts.Channel.ID, access.Channel.ID)
	})
}

func TestPhase151RecipientProjectionCopiesFreshCapabilitiesWithoutRoleInference(t *testing.T) {
	channel := &ent.Channel{ID: uuid.New(), ChannelType: "private"}
	access := messenger.ChannelAccess{
		TenantID: uuid.New(),
		Channel: channel,
		IsMember: true,
		Role: "owner",
		CanRead: true,
		CanPost: false,
		CanReply: false,
	}

	projection := messenger.ExportPhase151RecipientChannelProjection(access)

	assert.Equal(t, channel.ID.String(), projection.ID)
	assert.Equal(t, "owner", projection.Role)
	assert.True(t, projection.CanRead)
	assert.False(t, projection.CanPost, "role labels must not manufacture a capability")
	assert.False(t, projection.CanReply, "recipient projection copies fresh ChannelAccess only")
}