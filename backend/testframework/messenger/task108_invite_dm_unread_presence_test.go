package messenger_test

// TASK-108: invite atomic / orphan create / unread TTL / presence scope / bot / reaction channel.
import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
)

func newUnreadWiredService(t *testing.T) (*messenger.MessengerService, *mockRepository, *redis.Client) {
	t.Helper()
	_, rdb := newTestRedis(t)
	repo := newMockRepository()
	unread := messenger.NewUnreadService(rdb)
	presence := messenger.NewPresenceService(rdb)
	svc := messenger.NewMessengerService(repo, nil, presence, unread)
	return svc, repo, rdb
}

// TestAcceptInvite_IsMemberError_A2W3_2_004 — err IsMember → 500, без Increment.
func TestAcceptInvite_IsMemberError_A2W3_2_004(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)
	token := "tok-ismember-err"
	inv := &ent.ChannelInvite{
		ID: uuid.New(), ChannelID: channelID, Token: token,
		MaxUses: 3, UseCount: 0, CreatedAt: time.Now().UTC(),
	}
	repo.invites[inv.ID] = inv
	repo.invitesByToken[token] = inv
	repo.isMemberErr = errors.New("db down")

	_, err := svc.AcceptInvite(ctx, token, joinerID)
	requireMessengerAppErrorCode(t, err, 500)
	assert.Equal(t, 0, inv.UseCount, "use_count не должен расти при IsMember err")
}

// TestAcceptInvite_AddMembersFailRollsBackUse_A2W3_2_001 — fail AddMembers → Decrement.
func TestAcceptInvite_AddMembersFailRollsBackUse_A2W3_2_001(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)
	token := "tok-addmembers-fail"
	inv := &ent.ChannelInvite{
		ID: uuid.New(), ChannelID: channelID, Token: token,
		MaxUses: 2, UseCount: 0, CreatedAt: time.Now().UTC(),
	}
	repo.invites[inv.ID] = inv
	repo.invitesByToken[token] = inv
	repo.addMembersErr = errors.New("add members failed")

	_, err := svc.AcceptInvite(ctx, token, joinerID)
	require.Error(t, err)
	assert.Equal(t, 0, inv.UseCount, "слот должен откатиться после fail AddMembers")
}

// TestAcceptInvite_AtomicMaxUses_A2W3_3_001 — UseCountLT: exhausted.
func TestAcceptInvite_AtomicMaxUses_A2W3_3_001(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)
	token := "tok-exhausted"
	inv := &ent.ChannelInvite{
		ID: uuid.New(), ChannelID: channelID, Token: token,
		MaxUses: 1, UseCount: 1, CreatedAt: time.Now().UTC(),
	}
	repo.invites[inv.ID] = inv
	repo.invitesByToken[token] = inv

	_, err := svc.AcceptInvite(ctx, token, joinerID)
	requireMessengerAppErrorCode(t, err, 400)
	assert.Equal(t, 1, inv.UseCount)
}

// TestCreateChannel_OwnerAddMembersFailCleansOrphan_A2W3_2_002
func TestCreateChannel_OwnerAddMembersFailCleansOrphan_A2W3_2_002(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	repo.addMembersErr = errors.New("owner add failed")

	_, err := svc.CreateChannel(ctx, ownerID, messenger.CreateChannelRequest{
		ChannelType: "public",
		Name:        "orphan-check",
	})
	require.Error(t, err)
	require.Len(t, repo.channels, 1)
	for _, ch := range repo.channels {
		assert.NotNil(t, ch.DeletedAt, "orphan channel должен быть soft-deleted")
	}
}

// TestCreateOrGetDM_AddMembersFailCleansOrphan_A2W3_2_002
func TestCreateOrGetDM_AddMembersFailCleansOrphan_A2W3_2_002(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	a, b := uuid.New(), uuid.New()
	repo.addMembersErr = errors.New("dm members failed")

	_, err := svc.CreateOrGetDM(ctx, a, b)
	require.Error(t, err)
	require.NotEmpty(t, repo.channels)
	for _, ch := range repo.channels {
		assert.NotNil(t, ch.DeletedAt, "orphan DM должен быть soft-deleted")
	}
}

// TestUnread_TTLAndLeaveCleanup_A2W3_3_003
func TestUnread_TTLAndLeaveCleanup_A2W3_3_003(t *testing.T) {
	svc, repo, rdb := newUnreadWiredService(t)
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)

	unread := messenger.ExportMessengerServiceUnread(svc)
	require.NoError(t, unread.Increment(
		ctx, channelID.String(),
		[]string{ownerID.String(), memberID.String()},
		ownerID.String(),
	))
	// Mutation: убрать Expire → TTL < 0 и тест RED.
	ttl, err := rdb.TTL(ctx, "unread:"+memberID.String()).Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0), "unread key должен иметь TTL")

	require.NoError(t, svc.LeaveChannel(ctx, channelID, memberID))
	counts, err := unread.GetAll(ctx, memberID.String())
	require.NoError(t, err)
	_, still := counts[channelID.String()]
	assert.False(t, still, "после leave поле канала должно исчезнуть")
}

// TestUnread_DeleteChannelClearsAll_A2W3_3_003
func TestUnread_DeleteChannelClearsAll_A2W3_3_003(t *testing.T) {
	svc, repo, _ := newUnreadWiredService(t)
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)
	unread := messenger.ExportMessengerServiceUnread(svc)
	require.NoError(t, unread.Increment(ctx, channelID.String(),
		[]string{ownerID.String(), memberID.String()}, ownerID.String()))

	require.NoError(t, svc.DeleteChannel(ctx, channelID, ownerID))
	for _, uid := range []string{ownerID.String(), memberID.String()} {
		counts, err := unread.GetAll(ctx, uid)
		require.NoError(t, err)
		_, ok := counts[channelID.String()]
		assert.False(t, ok, "unread поля канала удалены для %s", uid)
	}
}

// TestGetPresence_FiltersSharedChannel_A2W3_1_001
func TestGetPresence_FiltersSharedChannel_A2W3_1_001(t *testing.T) {
	svc, repo, _ := newUnreadWiredService(t)
	ctx := context.Background()
	caller := uuid.New()
	sharedPeer := uuid.New()
	stranger := uuid.New()
	setupChannelWithMembers(repo, "public", false, caller, sharedPeer)

	require.NoError(t, svc.Heartbeat(ctx, sharedPeer.String()))
	require.NoError(t, svc.Heartbeat(ctx, stranger.String()))

	statuses, err := svc.GetPresence(ctx, caller, []string{sharedPeer.String(), stranger.String()})
	require.NoError(t, err)
	_, hasShared := statuses[sharedPeer.String()]
	_, hasStranger := statuses[stranger.String()]
	assert.True(t, hasShared, "peer из общего канала должен быть виден")
	assert.False(t, hasStranger, "чужой UUID не должен просачиваться")
}

// TestToggleReaction_ChannelMismatch_A2W3_1_003
func TestToggleReaction_ChannelMismatch_A2W3_1_003(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()
	realCh := setupChannelWithMembers(repo, "public", false, userID)
	otherCh := setupChannelWithMembers(repo, "public", false, userID)
	doc := &messenger.MessageDocument{
		ID: uuid.New().String(), ChannelID: realCh, SenderID: userID, CreatedAt: time.Now(),
	}
	repo.messages[doc.ID] = doc

	_, err := svc.ToggleReaction(ctx, otherCh, doc.ID, userID, "👍")
	requireMessengerAppErrorCode(t, err, 404)
}

// TestEntIncrementInviteUses_UseCountLT_A2W3_3_001 — слой репозитория.
func TestEntIncrementInviteUses_UseCountLT_A2W3_3_001(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "invite-atomic")
	inv, err := env.repo.CreateChannelInvite(env.ctx, messenger.CreateChannelInviteInput{
		ChannelID: ch.ID,
		CreatorID: env.owner.ID,
		Token:     "tok-" + uuid.New().String()[:8],
		MaxUses:   1,
	})
	require.NoError(t, err)

	require.NoError(t, env.repo.IncrementInviteUses(env.ctx, inv.ID, 1))
	err = env.repo.IncrementInviteUses(env.ctx, inv.ID, 1)
	require.Error(t, err)
	assert.True(t, errors.Is(err, messenger.ErrInviteUsesExhausted))

	found, err := env.repo.FindChannelInviteByToken(env.ctx, inv.Token)
	require.NoError(t, err)
	assert.Equal(t, 1, found.UseCount)
}

// TestEntListMembers_HasHardCap_A2W3_3_002 — Limit присутствует (не unbounded All).
func TestEntListMembers_HasHardCap_A2W3_3_002(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "cap-members")
	require.NoError(t, env.repo.AddMembers(env.ctx, ch.ID, []uuid.UUID{env.peer.ID}, "member"))

	members, err := env.repo.ListMembers(env.ctx, ch.ID)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(members), 500)
	assert.GreaterOrEqual(t, len(members), 1)
}
