//go:build hubtest

package messenger_test

// TASK-144 / A3W3-2: Join IsMember err, Create extra-members rollback,
// bookmark channel-scope, star/mute NotFound→403, ClosePoll pre-membership,
// AddMembers/DeleteChannel AppError wrap.
// Mutation: см. Progress в TASK-144.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/internal/messenger"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// TestJoinChannel_IsMemberError_A3W3_2_001 — DB err IsMember → 500, AddMembers не вызывается.
// Mutation: isMember, _ := … → assert 500 / addMembersCalls краснеет.
func TestJoinChannel_IsMemberError_A3W3_2_001(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)
	repo.isMemberErr = errors.New("db down")
	before := repo.addMembersCalls

	err := svc.JoinChannel(ctx, channelID, joinerID)
	requireMessengerAppErrorCode(t, err, 500)
	assert.Equal(t, before, repo.addMembersCalls, "AddMembers не должен вызываться при IsMember err")
	_, inMembers := repo.members[channelID][joinerID]
	assert.False(t, inMembers)
}

// TestCreateChannel_ExtraMembersFailRollsBack_A3W3_2_002 — fail AddMembers(extra) → soft-delete.
// Mutation: убрать DeleteChannel при fail extra → DeletedAt == nil, тест краснеет.
func TestCreateChannel_ExtraMembersFailRollsBack_A3W3_2_002(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	extraID := uuid.New()
	repo.addMembersErrExceptOwner = errors.New("extra members failed")

	_, err := svc.CreateChannel(ctx, ownerID, messenger.CreateChannelRequest{
		ChannelType: "public",
		Name:        "partial-members",
		MemberIDs:   []string{extraID.String()},
	})
	require.Error(t, err)
	requireMessengerAppErrorCode(t, err, 500)
	require.Len(t, repo.channels, 1)
	for _, ch := range repo.channels {
		assert.NotNil(t, ch.DeletedAt, "канал после fail доп. участников должен быть soft-deleted")
	}
}

// TestAddBookmark_RejectsCrossChannel_A3W3_2_003 — member B + msg из A + channelID=B → 404.
// Mutation: убрать msg.ChannelID == channelID → bookmark создаётся, тест краснеет.
func TestAddBookmark_RejectsCrossChannel_A3W3_2_003(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	userB := uuid.New()
	ownerA := uuid.New()
	channelA := setupChannelWithMembers(repo, "public", false, ownerA)
	channelB := setupChannelWithMembers(repo, "public", false, userB)
	msgFromA := "msg-from-a"
	repo.messages[msgFromA] = &messenger.MessageDocument{
		ID: msgFromA, ChannelID: channelA, SenderID: ownerA, CreatedAt: time.Now(),
	}

	err := svc.AddBookmark(ctx, userB, msgFromA, channelB)
	requireMessengerAppErrorCode(t, err, 404)
	assert.Empty(t, repo.bookmarks[userB], "закладка не должна создаваться cross-channel")
}

// TestToggleStarMuteUnmute_NonMemberForbidden_A3W3_2_004 — non-member → 403, не 500.
// Mutation: NotFound→Internal в ToggleStar → Code==500, тест краснеет.
func TestToggleStarMuteUnmute_NonMemberForbidden_A3W3_2_004(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	outsiderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	_, err := svc.ToggleStar(ctx, channelID, outsiderID)
	requireMessengerAppErrorCode(t, err, 403)

	_, err = svc.MuteChannel(ctx, channelID, outsiderID, "1h")
	requireMessengerAppErrorCode(t, err, 403)

	err = svc.UnmuteChannel(ctx, channelID, outsiderID)
	requireMessengerAppErrorCode(t, err, 403)
}

// TestClosePoll_CreatorLeftForbiddenBeforeClose_A3W3_2_005 — 403 до close, опрос открыт.
// Mutation: Close до IsMember → IsClosed=true при 403, тест краснеет.
func TestClosePoll_CreatorLeftForbiddenBeforeClose_A3W3_2_005(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	coOwnerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, coOwnerID)

	// Второй owner, чтобы LeaveChannel не блокировал единственного owner.
	repo.members[channelID][coOwnerID].Role = channelmember.RoleOwner

	poll, err := svc.CreatePoll(ctx, channelID, ownerID, messenger.CreatePollRequest{
		Question: "Stay open?",
		Options:  []string{"Yes", "No"},
	})
	require.NoError(t, err)
	pollID := uuid.MustParse(poll.ID)

	require.NoError(t, svc.LeaveChannel(ctx, channelID, ownerID))

	_, err = svc.ClosePoll(ctx, pollID, ownerID)
	requireMessengerAppErrorCode(t, err, 403)
	assert.False(t, repo.polls[pollID].IsClosed, "опрос не должен закрываться при 403 creator-left")
}

// TestAddMembers_DeleteChannel_WrapAppError_A3W3_2_006 — repo err → AppError, не сырой Ent.
// Mutation: return s.repo.… → errors.As на AppError падает.
func TestAddMembers_DeleteChannel_WrapAppError_A3W3_2_006(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	repo.addMembersErr = errors.New("pg: connection reset")
	err := svc.AddMembers(ctx, channelID, ownerID, []string{uuid.New().String()})
	requireMessengerAppErrorCode(t, err, 500)
	var ae *apperrors.AppError
	require.True(t, errors.As(err, &ae))
	assert.NotContains(t, ae.Message, "pg:")

	repo.addMembersErr = nil
	repo.deleteChannelErr = &ent.NotFoundError{}
	err = svc.DeleteChannel(ctx, channelID, ownerID)
	requireMessengerAppErrorCode(t, err, 404)

	channelID2 := setupChannelWithMembers(repo, "public", false, ownerID)
	repo.deleteChannelErr = errors.New("disk full")
	err = svc.DeleteChannel(ctx, channelID2, ownerID)
	requireMessengerAppErrorCode(t, err, 500)
}
