package messenger_test

// Тесты сервиса мессенджера (волна 114): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRevokeInvite_ForbiddenForMember проверяет отклонение при revoke приглашение запрещённый for участник.
func TestRevokeInvite_ForbiddenForMember(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID, memberID)

	created, err := svc.CreateInviteLink(ctx, channelID, ownerID, messenger.CreateInviteLinkRequest{
		MaxUses:   3,
		ExpiresIn: "1d",
	})
	require.NoError(t, err)

	err = svc.RevokeInvite(ctx, channelID, created.ID, memberID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "owner")
}

// TestRevokeInvite_NotFound проверяет сценарий: revoke приглашение not found.
func TestRevokeInvite_NotFound(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)

	err := svc.RevokeInvite(ctx, channelID, uuid.New(), ownerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдено")
}

// TestListChannelInvites_ForbiddenForMember проверяет отклонение при список channel invites запрещённый for участник.
func TestListChannelInvites_ForbiddenForMember(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID, memberID)

	_, err := svc.ListChannelInvites(ctx, channelID, memberID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "owner")
}
