package messenger_test

// TASK-008 coverage: exercise internal/messenger mockRepository via HookTestRepo.
// The in-package mock (~330 stmts) was nearly uncovered because testframework
// uses its own copy; HookTestRepo is the only production-path entry to it.

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/messenger"
)

// TestHookMockRepository_FullSurface_TASK008 walks Repository methods on the
// package-local mock (via HookTestRepo) including happy-path and NotFound edges.
func TestHookMockRepository_FullSurface_TASK008(t *testing.T) {
	ctx := context.Background()
	fixture := messenger.NewHookTestRepo()
	repo := fixture.Repository()

	owner := uuid.New()
	member := uuid.New()
	stranger := uuid.New()
	projectID := uuid.New()

	dmKey := "dm:" + owner.String() + ":" + member.String()
	ch, err := repo.CreateChannel(ctx, messenger.CreateChannelInput{
		ChannelType: "public",
		Name:        "general",
		Description: "desc",
		CreatedBy:   owner,
		ProjectID:   &projectID,
		DMKey:       &dmKey,
	})
	require.NoError(t, err)
	require.NotNil(t, fixture.Channel(ch.ID))

	require.NoError(t, repo.AddMembers(ctx, ch.ID, []uuid.UUID{owner}, "owner"))
	require.NoError(t, repo.AddMembers(ctx, ch.ID, []uuid.UUID{member}, "member"))

	got, err := repo.GetChannel(ctx, ch.ID)
	require.NoError(t, err)
	assert.Equal(t, ch.ID, got.ID)

	_, err = repo.GetChannel(ctx, uuid.New())
	require.Error(t, err)

	list, err := repo.ListUserChannels(ctx, owner)
	require.NoError(t, err)
	assert.NotEmpty(t, list)

	name := "renamed"
	desc := "new-desc"
	ro := true
	updated, err := repo.UpdateChannel(ctx, ch.ID, messenger.UpdateChannelInput{
		Name: &name, Description: &desc, IsReadonly: &ro, Version: ch.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "renamed", updated.Name)

	_, err = repo.UpdateChannel(ctx, ch.ID, messenger.UpdateChannelInput{Version: 999})
	require.Error(t, err)
	_, err = repo.UpdateChannel(ctx, uuid.New(), messenger.UpdateChannelInput{Version: 1})
	require.Error(t, err)

	ok, err := repo.IsMember(ctx, ch.ID, owner)
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = repo.IsMember(ctx, ch.ID, stranger)
	require.NoError(t, err)
	assert.False(t, ok)
	ok, err = repo.IsMember(ctx, uuid.New(), owner)
	require.NoError(t, err)
	assert.False(t, ok)

	role, err := repo.GetMemberRole(ctx, ch.ID, owner)
	require.NoError(t, err)
	assert.Equal(t, "owner", role)
	_, err = repo.GetMemberRole(ctx, ch.ID, stranger)
	require.Error(t, err)
	_, err = repo.GetMemberRole(ctx, uuid.New(), owner)
	require.Error(t, err)

	ids, err := repo.GetChannelMemberIDs(ctx, ch.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 2)
	ids, err = repo.GetChannelMemberIDs(ctx, uuid.New())
	require.NoError(t, err)
	assert.Nil(t, ids)

	members, err := repo.ListMembers(ctx, ch.ID)
	require.NoError(t, err)
	assert.Len(t, members, 2)
	members, err = repo.ListMembers(ctx, uuid.New())
	require.NoError(t, err)
	assert.Nil(t, members)

	shared, err := repo.FilterUsersSharingChannel(ctx, owner, []uuid.UUID{member, stranger})
	require.NoError(t, err)
	assert.Contains(t, shared, member)
	shared, err = repo.FilterUsersSharingChannel(ctx, owner, nil)
	require.NoError(t, err)
	assert.Nil(t, shared)

	dm, err := repo.FindDMChannel(ctx, dmKey)
	require.NoError(t, err)
	assert.Equal(t, ch.ID, dm.ID)
	_, err = repo.FindDMChannel(ctx, "missing")
	require.Error(t, err)

	msgID, err := repo.InsertMessage(ctx, &messenger.MessageDocument{
		ChannelID:   ch.ID,
		SenderID:    owner,
		ContentText: "hello world",
		CreatedAt:   time.Now().UTC(),
	})
	require.NoError(t, err)

	doc, err := repo.FindMessageByID(ctx, msgID)
	require.NoError(t, err)
	assert.Equal(t, "hello world", doc.ContentText)
	_, err = repo.FindMessageByID(ctx, "missing")
	require.Error(t, err)

	msgs, err := repo.ListMessages(ctx, ch.ID, nil, 1)
	require.NoError(t, err)
	assert.NotEmpty(t, msgs)

	_, err = repo.InsertMessage(ctx, &messenger.MessageDocument{
		ChannelID: ch.ID, SenderID: member, ContentText: "second", CreatedAt: time.Now().UTC(),
	})
	require.NoError(t, err)
	msgs, err = repo.ListMessages(ctx, ch.ID, nil, 1)
	require.NoError(t, err)
	assert.Len(t, msgs, 1)

	edited, err := repo.UpdateMessage(ctx, msgID, map[string]interface{}{"t": 1}, "edited")
	require.NoError(t, err)
	assert.Equal(t, "edited", edited.ContentText)
	_, err = repo.UpdateMessage(ctx, "missing", nil, "x")
	require.Error(t, err)

	require.NoError(t, repo.IncrementReplyCount(ctx, msgID))
	require.NoError(t, repo.IncrementReplyCount(ctx, "ghost"))
	require.NoError(t, repo.DecrementReplyCount(ctx, msgID))
	require.NoError(t, repo.DecrementReplyCount(ctx, "ghost"))

	replyID, err := repo.InsertMessage(ctx, &messenger.MessageDocument{
		ChannelID: ch.ID, SenderID: member, ContentText: "reply", ThreadID: msgID, CreatedAt: time.Now().UTC(),
	})
	require.NoError(t, err)
	replies, err := repo.ListThreadReplies(ctx, msgID, nil, 10)
	require.NoError(t, err)
	assert.NotEmpty(t, replies)
	_, _ = replyID, replies
	replies, err = repo.ListThreadReplies(ctx, msgID, nil, 1)
	require.NoError(t, err)
	assert.Len(t, replies, 1)

	uid := owner.String()
	require.NoError(t, repo.AddReaction(ctx, msgID, uid, "👍"))
	require.NoError(t, repo.AddReaction(ctx, msgID, uid, "👍")) // duplicate
	require.NoError(t, repo.AddReaction(ctx, msgID, member.String(), "👍"))
	require.NoError(t, repo.AddReaction(ctx, msgID, uid, "🔥"))
	require.Error(t, repo.AddReaction(ctx, "missing", uid, "👍"))

	found, err := repo.FindReaction(ctx, msgID, uid, "👍")
	require.NoError(t, err)
	assert.True(t, found)
	found, err = repo.FindReaction(ctx, msgID, uid, "nope")
	require.NoError(t, err)
	assert.False(t, found)
	found, err = repo.FindReaction(ctx, "missing", uid, "👍")
	require.NoError(t, err)
	assert.False(t, found)

	require.NoError(t, repo.RemoveReaction(ctx, msgID, uid, "👍"))
	require.NoError(t, repo.RemoveReaction(ctx, msgID, uid, "ghost"))
	require.Error(t, repo.RemoveReaction(ctx, "missing", uid, "👍"))

	hits, err := repo.SearchMessages(ctx, "edited", []uuid.UUID{ch.ID}, 10)
	require.NoError(t, err)
	assert.NotEmpty(t, hits)
	hits, err = repo.SearchMessages(ctx, "edited", nil, 10)
	require.NoError(t, err)
	assert.Empty(t, hits)
	hits, err = repo.SearchMessages(ctx, "edited", []uuid.UUID{ch.ID}, 1)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(hits), 1)
	_, _ = repo.SearchMessages(ctx, "", []uuid.UUID{ch.ID}, 10) // empty sub via contains

	require.NoError(t, repo.AddPin(ctx, ch.ID, msgID, owner))
	pin, err := repo.FindPin(ctx, ch.ID, msgID)
	require.NoError(t, err)
	assert.NotNil(t, pin)
	_, err = repo.FindPin(ctx, ch.ID, "missing")
	require.Error(t, err)
	pins, err := repo.ListPins(ctx, ch.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, pins)
	require.NoError(t, repo.RemovePin(ctx, ch.ID, msgID))

	bm, err := repo.AddBookmark(ctx, owner, msgID, ch.ID)
	require.NoError(t, err)
	bms, err := repo.ListBookmarks(ctx, owner)
	require.NoError(t, err)
	assert.NotEmpty(t, bms)
	require.NoError(t, repo.RemoveBookmark(ctx, owner, bm.ID))
	require.Error(t, repo.RemoveBookmark(ctx, owner, uuid.New()))

	starred, err := repo.ToggleStar(ctx, ch.ID, owner)
	require.NoError(t, err)
	assert.True(t, starred)
	starred, err = repo.ToggleStar(ctx, ch.ID, owner)
	require.NoError(t, err)
	assert.False(t, starred)

	until, err := repo.MuteChannel(ctx, ch.ID, owner, "1h")
	require.NoError(t, err)
	assert.True(t, until.After(time.Now()))
	_, err = repo.MuteChannel(ctx, ch.ID, member, "forever")
	require.NoError(t, err)
	muted, err := repo.IsChannelMuted(ctx, ch.ID, owner)
	require.NoError(t, err)
	assert.True(t, muted)
	muted, err = repo.IsChannelMuted(ctx, ch.ID, stranger)
	require.NoError(t, err)
	assert.False(t, muted)
	require.NoError(t, repo.UnmuteChannel(ctx, ch.ID, owner))

	memships, err := repo.GetUserChannelMemberships(ctx, owner)
	require.NoError(t, err)
	assert.NotEmpty(t, memships)
	require.NoError(t, repo.EnsureIndexes(ctx))

	byRole, err := repo.ListProjectMembersByRole(ctx, projectID, "developer")
	require.NoError(t, err)
	assert.Nil(t, byRole)
	isPM, err := repo.IsProjectMember(ctx, projectID, owner)
	require.NoError(t, err)
	assert.False(t, isPM)
	dn, err := repo.GetUserDisplayName(ctx, owner)
	require.NoError(t, err)
	assert.Empty(t, dn)

	inv, err := repo.CreateChannelInvite(ctx, messenger.CreateChannelInviteInput{
		ChannelID: ch.ID, CreatorID: owner, Token: "tok-abc", MaxUses: 1,
	})
	require.NoError(t, err)
	foundInv, err := repo.FindChannelInviteByToken(ctx, "tok-abc")
	require.NoError(t, err)
	assert.Equal(t, inv.ID, foundInv.ID)
	_, err = repo.FindChannelInviteByToken(ctx, "nope")
	require.Error(t, err)

	require.NoError(t, repo.IncrementInviteUses(ctx, inv.ID, 1))
	require.ErrorIs(t, repo.IncrementInviteUses(ctx, inv.ID, 1), messenger.ErrInviteUsesExhausted)
	require.Error(t, repo.IncrementInviteUses(ctx, uuid.New(), 1))
	require.NoError(t, repo.DecrementInviteUses(ctx, inv.ID))
	require.Error(t, repo.DecrementInviteUses(ctx, inv.ID)) // use_count 0
	require.Error(t, repo.DecrementInviteUses(ctx, uuid.New()))

	invites, err := repo.ListChannelInvites(ctx, ch.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, invites)
	require.NoError(t, repo.RevokeChannelInvite(ctx, ch.ID, inv.ID))
	require.Error(t, repo.RevokeChannelInvite(ctx, ch.ID, uuid.New()))
	invites, err = repo.ListChannelInvites(ctx, ch.ID)
	require.NoError(t, err)
	assert.Empty(t, invites)

	poll, opts, err := repo.CreatePollWithOptions(ctx, messenger.CreatePollInput{
		ChannelID: ch.ID, CreatorID: owner, Question: "Q?", Options: []string{"A", "B"},
	})
	require.NoError(t, err)
	require.Len(t, opts, 2)
	p2, opts2, err := repo.GetPollWithOptions(ctx, poll.ID)
	require.NoError(t, err)
	assert.Equal(t, poll.ID, p2.ID)
	assert.Len(t, opts2, 2)
	_, _, err = repo.GetPollWithOptions(ctx, uuid.New())
	require.Error(t, err)
	p3, err := repo.GetPoll(ctx, poll.ID)
	require.NoError(t, err)
	assert.Equal(t, poll.ID, p3.ID)
	_, err = repo.GetPoll(ctx, uuid.New())
	require.Error(t, err)

	require.NoError(t, repo.VotePoll(ctx, poll.ID, owner, []uuid.UUID{opts[0].ID}))
	require.NoError(t, repo.VotePoll(ctx, poll.ID, owner, []uuid.UUID{opts[1].ID})) // replace
	votes, err := repo.ListVotes(ctx, poll.ID)
	require.NoError(t, err)
	assert.Len(t, votes, 1)
	require.NoError(t, repo.ClosePoll(ctx, poll.ID))
	require.Error(t, repo.ClosePoll(ctx, uuid.New()))

	require.NoError(t, repo.RemoveMember(ctx, ch.ID, member))
	require.NoError(t, repo.SoftDeleteMessage(ctx, msgID))
	require.Error(t, repo.SoftDeleteMessage(ctx, "missing"))
	_, err = repo.FindMessageByID(ctx, msgID)
	require.Error(t, err)

	require.NoError(t, repo.DeleteChannel(ctx, ch.ID))
	_, err = repo.GetChannel(ctx, ch.ID)
	require.Error(t, err)
	require.Error(t, repo.DeleteChannel(ctx, uuid.New()))
	ok, err = repo.IsMember(ctx, ch.ID, owner)
	require.NoError(t, err)
	assert.False(t, ok)

	pubID := fixture.SetupPublicChannel(owner, member)
	assert.NotEqual(t, uuid.Nil, pubID)
	dmID := fixture.SetupDM(owner, member)
	assert.NotEqual(t, uuid.Nil, dmID)
}
