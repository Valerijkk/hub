//go:build hubtest

package messenger_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/testframework/mongotest"
)

func requireMessengerAppErrorCode(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr), "expected AppError, got %T: %v", err, err)
	require.Equal(t, code, appErr.Code, "unexpected AppError: %v", err)
}

// ---------------------------------------------------------------------------
// P1-BE-15 — RemoveBookmark owner-scope (Ent)
// Mutation: drop UserIDEQ from Where → FAIL.
// ---------------------------------------------------------------------------

func TestRemoveBookmark_RejectsForeignOwner_P1BE15(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "bm-idor")
	msgID := uuid.New().String()

	bm, err := env.repo.AddBookmark(env.ctx, env.owner.ID, msgID, ch.ID)
	require.NoError(t, err)

	err = env.repo.RemoveBookmark(env.ctx, env.peer.ID, bm.ID)
	require.True(t, ent.IsNotFound(err), "foreign user must not delete bookmark, got %v", err)

	list, err := env.repo.ListBookmarks(env.ctx, env.owner.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NoError(t, env.repo.RemoveBookmark(env.ctx, env.owner.ID, bm.ID))
}

func TestRemoveBookmark_ServiceRejectsForeignOwner_P1BE15(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	outsiderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	require.NoError(t, svc.AddBookmark(ctx, ownerID, "msg-a", channelID))
	list, err := svc.ListBookmarks(ctx, ownerID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	bmID := uuid.MustParse(list[0].ID)

	err = svc.RemoveBookmark(ctx, outsiderID, bmID)
	requireMessengerAppErrorCode(t, err, 404)

	list, err = svc.ListBookmarks(ctx, ownerID)
	require.NoError(t, err)
	require.Len(t, list, 1)
}

// ---------------------------------------------------------------------------
// P1-BE-16 — Reaction / Pin / AddBookmark membership
// Mutation: drop IsMember → FAIL.
// ---------------------------------------------------------------------------

func TestToggleReaction_RejectsNonMember_P1BE16(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	outsiderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	doc := &messenger.MessageDocument{
		ID: uuid.New().String(), ChannelID: channelID, SenderID: ownerID, CreatedAt: time.Now(),
	}
	repo.messages[doc.ID] = doc

	_, err := svc.ToggleReaction(ctx, channelID, doc.ID, outsiderID, "🔥")
	requireMessengerAppErrorCode(t, err, 403)
}

func TestTogglePin_RejectsNonMember_P1BE16(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	outsiderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)
	msgID := "pin-idor"
	repo.messages[msgID] = &messenger.MessageDocument{
		ID: msgID, ChannelID: channelID, SenderID: ownerID, CreatedAt: time.Now(),
	}

	_, err := svc.TogglePin(ctx, channelID, msgID, outsiderID)
	requireMessengerAppErrorCode(t, err, 403)
}

func TestAddBookmark_RejectsNonMember_P1BE16(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	outsiderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	err := svc.AddBookmark(ctx, outsiderID, "msg-x", channelID)
	requireMessengerAppErrorCode(t, err, 403)
}

// ---------------------------------------------------------------------------
// P1-BE-17 — DeleteMessage: URL channel ≠ msg.ChannelID
// Mutation: drop ChannelID match → FAIL.
// ---------------------------------------------------------------------------

func TestDeleteMessage_RejectsChannelMismatch_P1BE17(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerA := uuid.New()
	ownerB := uuid.New()
	channelA := setupChannelWithMembers(repo, "public", false, ownerA)
	channelB := setupChannelWithMembers(repo, "public", false, ownerB)

	doc := &messenger.MessageDocument{
		ID: uuid.New().String(), ChannelID: channelB, SenderID: ownerB, CreatedAt: time.Now(),
	}
	repo.messages[doc.ID] = doc

	// Owner A пытается удалить сообщение канала B, подставляя channelA в URL.
	err := svc.DeleteMessage(ctx, doc.ID, ownerA, channelA.String(), false)
	requireMessengerAppErrorCode(t, err, 404)

	still, err := repo.FindMessageByID(ctx, doc.ID)
	require.NoError(t, err)
	assert.Nil(t, still.DeletedAt)
}

// ---------------------------------------------------------------------------
// P1-BE-18 — GetPollResults membership
// Mutation: drop IsMember → FAIL.
// ---------------------------------------------------------------------------

func TestGetPollResults_RejectsNonMember_P1BE18(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	outsiderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	poll, err := svc.CreatePoll(ctx, channelID, ownerID, messenger.CreatePollRequest{
		Question:    "Secret?",
		Options:     []string{"Yes", "No"},
		IsAnonymous: false,
	})
	require.NoError(t, err)
	pollID := uuid.MustParse(poll.ID)

	_, err = svc.GetPollResults(ctx, pollID, outsiderID)
	requireMessengerAppErrorCode(t, err, 403)
}

// ---------------------------------------------------------------------------
// P1-BE-19 — Channel invite revoke ChannelIDEQ (Ent)
// Mutation: drop ChannelIDEQ → FAIL.
// ---------------------------------------------------------------------------

func TestRevokeChannelInvite_RejectsForeignChannel_P1BE19(t *testing.T) {
	env := newMessengerRepoEnv(t)
	chA := env.createPublicChannel(t, "invite-a")
	chB := env.createPublicChannel(t, "invite-b")

	inv, err := env.repo.CreateChannelInvite(env.ctx, messenger.CreateChannelInviteInput{
		ChannelID: chB.ID,
		CreatorID: env.owner.ID,
		Token:     "tok-" + uuid.New().String()[:8],
		MaxUses:   3,
	})
	require.NoError(t, err)

	err = env.repo.RevokeChannelInvite(env.ctx, chA.ID, inv.ID)
	require.True(t, ent.IsNotFound(err), "cross-channel revoke must 404, got %v", err)

	list, err := env.repo.ListChannelInvites(env.ctx, chB.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NoError(t, env.repo.RevokeChannelInvite(env.ctx, chB.ID, inv.ID))
}

func TestRevokeInvite_ServiceRejectsForeignChannel_P1BE19(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerA := uuid.New()
	ownerB := uuid.New()
	channelA := setupChannelWithMembers(repo, "private", false, ownerA)
	channelB := setupChannelWithMembers(repo, "private", false, ownerB)

	inv, err := svc.CreateInviteLink(ctx, channelB, ownerB, messenger.CreateInviteLinkRequest{MaxUses: 2})
	require.NoError(t, err)

	err = svc.RevokeInvite(ctx, channelA, inv.ID, ownerA)
	requireMessengerAppErrorCode(t, err, 404)

	list, err := svc.ListChannelInvites(ctx, channelB, ownerB)
	require.NoError(t, err)
	require.Len(t, list, 1)
}

// ---------------------------------------------------------------------------
// P1-BE-20 — AddReaction emoji dedup (Mongo)
// Mutation: push when ModifiedCount==0 → FAIL (2 entries).
// ---------------------------------------------------------------------------

func TestAddReaction_DedupsSameEmoji_P1BE20(t *testing.T) {
	db := mongotest.Connect(t, "hub_test_react_dedup_")
	repo := messenger.NewMongoRepository(db)
	ctx := context.Background()

	msgID, err := repo.InsertMessage(ctx, &messenger.MessageDocument{
		ChannelID:   uuid.New(),
		SenderID:    uuid.New(),
		ContentJSON: map[string]interface{}{"text": "hi"},
		ContentText: "hi",
		CreatedAt:   time.Now().UTC(),
	})
	require.NoError(t, err)

	userID := uuid.New().String()
	require.NoError(t, repo.AddReaction(ctx, msgID, userID, "👍"))
	require.NoError(t, repo.AddReaction(ctx, msgID, userID, "👍")) // retry / double-click

	found, err := repo.FindMessageByID(ctx, msgID)
	require.NoError(t, err)

	var thumbEntries int
	for _, r := range found.Reactions {
		if r.Emoji == "👍" {
			thumbEntries++
			assert.Equal(t, []string{userID}, r.UserIDs)
		}
	}
	assert.Equal(t, 1, thumbEntries, "same emoji must not create duplicate ReactionEntry")
}
