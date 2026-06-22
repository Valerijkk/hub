package messenger_test

// Тесты сервиса мессенджера (волна 115): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

// TestAcceptInvite_Revoked проверяет сценарий: accept приглашение revoked.
func TestAcceptInvite_Revoked(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)

	inv := &ent.ChannelInvite{
		ID: uuid.New(), ChannelID: channelID, Token: "revoked-token",
		IsRevoked: true, CreatedAt: time.Now(),
	}
	repo.invites[inv.ID] = inv
	repo.invitesByToken[inv.Token] = inv

	_, err := svc.AcceptInvite(ctx, inv.Token, joinerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "отозвано")
}

// TestAcceptInvite_Expired проверяет отклонение при accept приглашение просроченный.
func TestAcceptInvite_Expired(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)

	past := time.Now().UTC().Add(-time.Hour)
	inv := &ent.ChannelInvite{
		ID: uuid.New(), ChannelID: channelID, Token: "expired-accept",
		ExpiresAt: &past, CreatedAt: time.Now(),
	}
	repo.invites[inv.ID] = inv
	repo.invitesByToken[inv.Token] = inv

	_, err := svc.AcceptInvite(ctx, inv.Token, joinerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "истёк")
}

// TestAcceptInvite_MaxUsesExceeded проверяет сценарий: accept приглашение max uses exceeded.
func TestAcceptInvite_MaxUsesExceeded(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)

	inv := &ent.ChannelInvite{
		ID: uuid.New(), ChannelID: channelID, Token: "maxed-accept",
		MaxUses: 1, UseCount: 1, CreatedAt: time.Now(),
	}
	repo.invites[inv.ID] = inv
	repo.invitesByToken[inv.Token] = inv

	_, err := svc.AcceptInvite(ctx, inv.Token, joinerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "лимит")
}

// TestAcceptInvite_NotFound проверяет сценарий: accept приглашение not found.
func TestAcceptInvite_NotFound(t *testing.T) {
	svc, _, _ := newTestService()
	_, err := svc.AcceptInvite(context.Background(), "missing-token", uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдено")
}

// TestFindMessageForBot_NotFound проверяет сценарий: find message for бот not found.
func TestFindMessageForBot_NotFound(t *testing.T) {
	svc, _, _ := newTestService()
	svc.SetSystemBotID(uuid.New())
	_, err := svc.FindMessageForBot(context.Background(), uuid.New().String())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдено")
}

// TestJoinChannel_NotFound проверяет сценарий: join channel not found.
func TestJoinChannel_NotFound(t *testing.T) {
	svc, _, _ := newTestService()
	err := svc.JoinChannel(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

// TestGetChannel_NonMemberForbidden проверяет отклонение при получение channel non участник запрещённый.
func TestGetChannel_NonMemberForbidden(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	outsiderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	_, err := svc.GetChannel(ctx, channelID, outsiderID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "участником")
}

// TestAddMembers_InvalidMemberID проверяет отклонение при add members некорректный участник id.
func TestAddMembers_InvalidMemberID(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	err := svc.AddMembers(ctx, channelID, ownerID, []string{"not-a-uuid"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "member ID")
}

// TestCreateChannel_SkipsInvalidMemberIDs проверяет отклонение при создание channel skips некорректный участник i ds.
func TestCreateChannel_SkipsInvalidMemberIDs(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	validMember := uuid.New()

	resp, err := svc.CreateChannel(ctx, ownerID, messenger.CreateChannelRequest{
		ChannelType: "public",
		Name:        "with-members",
		MemberIDs:   []string{"bad-id", validMember.String()},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	members := repo.members[uuid.MustParse(resp.ID)]
	assert.Contains(t, members, ownerID)
	assert.Contains(t, members, validMember)
}
