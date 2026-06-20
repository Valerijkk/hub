package messenger

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channelmember"
)

func TestValidateInvite_AllStates(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)

	t.Run("not found", func(t *testing.T) {
		resp, err := svc.ValidateInvite(ctx, "missing-token")
		require.NoError(t, err)
		assert.False(t, resp.Valid)
		assert.Contains(t, resp.Message, "не найдено")
	})

	t.Run("revoked", func(t *testing.T) {
		inv := &ent.ChannelInvite{
			ID:        uuid.New(),
			ChannelID: channelID,
			Token:     "revoked-token",
			IsRevoked: true,
			CreatedAt: time.Now(),
		}
		repo.invites[inv.ID] = inv
		repo.invitesByToken[inv.Token] = inv

		resp, err := svc.ValidateInvite(ctx, inv.Token)
		require.NoError(t, err)
		assert.False(t, resp.Valid)
		assert.Contains(t, resp.Message, "отозвано")
	})

	t.Run("expired", func(t *testing.T) {
		past := time.Now().UTC().Add(-time.Hour)
		inv := &ent.ChannelInvite{
			ID:        uuid.New(),
			ChannelID: channelID,
			Token:     "expired-token",
			ExpiresAt: &past,
			CreatedAt: time.Now(),
		}
		repo.invites[inv.ID] = inv
		repo.invitesByToken[inv.Token] = inv

		resp, err := svc.ValidateInvite(ctx, inv.Token)
		require.NoError(t, err)
		assert.False(t, resp.Valid)
		assert.Contains(t, resp.Message, "истёк")
	})

	t.Run("max uses exhausted", func(t *testing.T) {
		inv := &ent.ChannelInvite{
			ID:        uuid.New(),
			ChannelID: channelID,
			Token:     "maxed-token",
			MaxUses:   2,
			UseCount:  2,
			CreatedAt: time.Now(),
		}
		repo.invites[inv.ID] = inv
		repo.invitesByToken[inv.Token] = inv

		resp, err := svc.ValidateInvite(ctx, inv.Token)
		require.NoError(t, err)
		assert.False(t, resp.Valid)
		assert.Contains(t, resp.Message, "лимит")
	})

	t.Run("valid invite", func(t *testing.T) {
		inv := &ent.ChannelInvite{
			ID:        uuid.New(),
			ChannelID: channelID,
			Token:     "valid-token",
			MaxUses:   5,
			UseCount:  1,
			CreatedAt: time.Now(),
		}
		repo.invites[inv.ID] = inv
		repo.invitesByToken[inv.Token] = inv
		repo.channels[channelID].Name = "team-chat"

		resp, err := svc.ValidateInvite(ctx, inv.Token)
		require.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Equal(t, "team-chat", resp.ChannelName)
		assert.Equal(t, channelID.String(), resp.ChannelID)
		assert.Equal(t, 1, resp.MemberCount)
	})
}

func TestAcceptInvite_AddsMember(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)

	inv := &ent.ChannelInvite{
		ID:        uuid.New(),
		ChannelID: channelID,
		Token:     "join-token",
		MaxUses:   10,
		UseCount:  0,
		CreatedAt: time.Now(),
	}
	repo.invites[inv.ID] = inv
	repo.invitesByToken[inv.Token] = inv

	ch, err := svc.AcceptInvite(ctx, inv.Token, joinerID)
	require.NoError(t, err)
	assert.Equal(t, channelID.String(), ch.ID)

	isMember, _ := repo.IsMember(ctx, channelID, joinerID)
	assert.True(t, isMember)
	assert.Equal(t, 1, repo.invites[inv.ID].UseCount)
}

func TestAcceptInvite_AlreadyMember(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)

	inv := &ent.ChannelInvite{
		ID:        uuid.New(),
		ChannelID: channelID,
		Token:     "existing-token",
		CreatedAt: time.Now(),
	}
	repo.invites[inv.ID] = inv
	repo.invitesByToken[inv.Token] = inv

	ch, err := svc.AcceptInvite(ctx, inv.Token, ownerID)
	require.NoError(t, err)
	assert.Equal(t, channelID.String(), ch.ID)
	assert.Equal(t, 0, repo.invites[inv.ID].UseCount)
}

func TestBookmarks_AddListRemove(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()
	channelID := uuid.New()
	messageID := "msg-bookmark"

	require.NoError(t, svc.AddBookmark(ctx, userID, messageID, channelID))

	list, err := svc.ListBookmarks(ctx, userID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, messageID, list[0].MessageID)
	assert.Equal(t, channelID.String(), list[0].ChannelID)

	bookmarkID := uuid.MustParse(list[0].ID)
	require.NoError(t, svc.RemoveBookmark(ctx, bookmarkID))

	list, err = svc.ListBookmarks(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestTogglePin_AddAndRemove(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()
	channelID := uuid.New()
	messageID := "pin-msg"

	pinned, err := svc.TogglePin(ctx, channelID, messageID, userID)
	require.NoError(t, err)
	assert.True(t, pinned)

	pinned, err = svc.TogglePin(ctx, channelID, messageID, userID)
	require.NoError(t, err)
	assert.False(t, pinned)
}

func TestListPins_RequiresMembership(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	outsiderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	_, err := svc.ListPins(ctx, channelID, outsiderID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "участником")
}

func TestGetChannel_ForbiddenForNonMember(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	_, err := svc.GetChannel(ctx, channelID, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "участником")
}

func TestCreatePollAndVote(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	poll, err := svc.CreatePoll(ctx, channelID, ownerID, CreatePollRequest{
		Question: "Pick one",
		Options:  []string{"A", "B"},
	})
	require.NoError(t, err)
	assert.Equal(t, "Pick one", poll.Question)
	assert.Len(t, poll.Options, 2)

	pollID := uuid.MustParse(poll.ID)
	optionID := uuid.MustParse(poll.Options[0].ID)

	voted, err := svc.VotePoll(ctx, pollID, ownerID, VotePollRequest{OptionIDs: []string{optionID.String()}})
	require.NoError(t, err)
	assert.True(t, voted.Options[0].Voted)
	assert.Equal(t, 1, voted.TotalVotes)
}

func TestResolveGroupMentions_AllAndRole(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	devID := uuid.New()
	outsiderID := uuid.New()
	projectID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, devID)

	repo.projectMembersByRole[projectID] = map[string][]uuid.UUID{
		"developer": {devID, outsiderID},
	}

	mentions := svc.resolveGroupMentions(ctx, channelID, &projectID, []string{"all", "developer"}, nil, ownerID)
	assert.Contains(t, mentions, devID.String())
	assert.NotContains(t, mentions, ownerID.String())
	assert.NotContains(t, mentions, outsiderID.String())
}

func TestListMyChannels_EnrichesMembership(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	mutedUntil := time.Now().UTC().Add(time.Hour)
	repo.members[channelID][userID].IsStarred = true
	repo.members[channelID][userID].MutedUntil = &mutedUntil

	channels, err := svc.ListMyChannels(ctx, userID)
	require.NoError(t, err)
	require.Len(t, channels, 1)
	assert.True(t, channels[0].IsStarred)
	require.NotNil(t, channels[0].MutedUntil)
}

func TestAddMembers_InvalidUUIDRejected(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	validID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	err := svc.AddMembers(ctx, channelID, ownerID, []string{"bad-uuid", validID.String()})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный member ID")

	isMember, _ := repo.IsMember(ctx, channelID, validID)
	assert.False(t, isMember)
}

func TestUpdateChannel_RequiresAdmin(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)

	_, err := svc.UpdateChannel(ctx, channelID, memberID, CreateChannelRequest{Name: "new-name"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "owner или admin")

	resp, err := svc.UpdateChannel(ctx, channelID, ownerID, CreateChannelRequest{Name: "renamed"})
	require.NoError(t, err)
	assert.Equal(t, "renamed", resp.Name)
}

func TestDeleteChannel_OnlyOwner(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	adminID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)
	repo.members[channelID][adminID] = &ent.ChannelMember{
		ID: uuid.New(), ChannelID: channelID, UserID: adminID,
		Role: channelmember.RoleAdmin, JoinedAt: time.Now(),
	}

	err := svc.DeleteChannel(ctx, channelID, adminID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "только owner")

	require.NoError(t, svc.DeleteChannel(ctx, channelID, ownerID))
	_, err = repo.GetChannel(ctx, channelID)
	require.Error(t, err)
}

func TestLeaveChannel(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)

	require.NoError(t, svc.LeaveChannel(ctx, channelID, memberID))
	isMember, _ := repo.IsMember(ctx, channelID, memberID)
	assert.False(t, isMember)
}

func TestToggleStarAndMute(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()
	channelID := uuid.New()

	starred, err := svc.ToggleStar(ctx, channelID, userID)
	require.NoError(t, err)
	assert.True(t, starred)

	mutedUntil, err := svc.MuteChannel(ctx, channelID, userID, "1h")
	require.NoError(t, err)
	assert.NotEmpty(t, mutedUntil)

	muted, err := svc.IsChannelMuted(ctx, channelID, userID)
	require.NoError(t, err)
	assert.True(t, muted)

	require.NoError(t, svc.UnmuteChannel(ctx, channelID, userID))
	muted, err = svc.IsChannelMuted(ctx, channelID, userID)
	require.NoError(t, err)
	assert.False(t, muted)
}

func TestListMembers(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)

	members, err := svc.ListMembers(ctx, channelID, ownerID)
	require.NoError(t, err)
	assert.Len(t, members, 2)
}

func TestListMessagesAndThreadReplies(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	main := &MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		SenderID:    ownerID,
		ContentText: "main",
		CreatedAt:   time.Now(),
	}
	repo.messages[main.ID] = main
	reply := &MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		SenderID:    ownerID,
		ContentText: "reply",
		ThreadID:    main.ID,
		CreatedAt:   time.Now(),
	}
	repo.messages[reply.ID] = reply

	messages, err := svc.ListMessages(ctx, channelID, ownerID, nil, 10)
	require.NoError(t, err)
	assert.Len(t, messages, 1)

	replies, err := svc.ListThreadReplies(ctx, main.ID, ownerID, nil, 10)
	require.NoError(t, err)
	assert.Len(t, replies, 1)
}

func TestInviteLinkLifecycle(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)

	created, err := svc.CreateInviteLink(ctx, channelID, ownerID, CreateInviteLinkRequest{
		MaxUses:   5,
		ExpiresIn: "1d",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.Token)
	assert.Contains(t, created.InviteURL, created.Token)

	invites, err := svc.ListChannelInvites(ctx, channelID, ownerID)
	require.NoError(t, err)
	assert.Len(t, invites, 1)

	require.NoError(t, svc.RevokeInvite(ctx, channelID, created.ID, ownerID))
}

func TestListPins_ReturnsPinnedMessages(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)
	messageID := "pinned-msg"

	_, err := svc.TogglePin(ctx, channelID, messageID, ownerID)
	require.NoError(t, err)

	pins, err := svc.ListPins(ctx, channelID, ownerID)
	require.NoError(t, err)
	assert.Len(t, pins, 1)
	assert.Equal(t, messageID, pins[0].MessageID)
}

func TestClosePoll(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	poll, err := svc.CreatePoll(ctx, channelID, ownerID, CreatePollRequest{
		Question: "Close me",
		Options:  []string{"A", "B"},
	})
	require.NoError(t, err)

	closed, err := svc.ClosePoll(ctx, uuid.MustParse(poll.ID), ownerID)
	require.NoError(t, err)
	assert.True(t, closed.IsClosed)
}

func TestForwardMessage_SourceChannelMembershipRequired(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	outsiderID := uuid.New()
	sourceChannelID := setupChannelWithMembers(repo, "public", false, ownerID)
	targetChannelID := setupChannelWithMembers(repo, "public", false, outsiderID)

	doc := &MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   sourceChannelID,
		SenderID:    ownerID,
		ContentText: "secret",
		CreatedAt:   time.Now(),
	}
	repo.messages[doc.ID] = doc

	_, err := svc.ForwardMessage(ctx, sourceChannelID, outsiderID, doc.ID, targetChannelID.String())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "исходного канала")
}
