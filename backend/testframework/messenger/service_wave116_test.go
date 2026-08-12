package messenger_test

// Тесты сервиса мессенджера (волна 116): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channelmember"
)

// TestUpdateChannel_ForbiddenForMember проверяет отклонение при обновление channel запрещённый for участник.
func TestUpdateChannel_ForbiddenForMember(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)

	_, err := svc.UpdateChannel(ctx, channelID, memberID, messenger.CreateChannelRequest{Name: "nope", Version: 1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

// TestDeleteChannel_ForbiddenForAdmin проверяет отклонение при удаление channel запрещённый for админ.
func TestDeleteChannel_ForbiddenForAdmin(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	adminID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)
	repo.members[channelID][adminID] = &ent.ChannelMember{
		ID: uuid.New(), ChannelID: channelID, UserID: adminID,
		Role: channelmember.RoleAdmin, JoinedAt: repo.members[channelID][ownerID].JoinedAt,
	}

	err := svc.DeleteChannel(ctx, channelID, adminID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "owner")
}

// TestDeleteChannel_SuccessForOwner проверяет сценарий: удаление channel успех for owner.
func TestDeleteChannel_SuccessForOwner(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	require.NoError(t, svc.DeleteChannel(ctx, channelID, ownerID))
	ch, ok := repo.channels[channelID]
	require.True(t, ok, "soft-delete keeps row")
	require.NotNil(t, ch.DeletedAt)
	_, err := repo.GetChannel(ctx, channelID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}

// TestAddMembers_ForbiddenForMember проверяет отклонение при add members запрещённый for участник.
func TestAddMembers_ForbiddenForMember(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)

	err := svc.AddMembers(ctx, channelID, memberID, []string{uuid.New().String()})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

// TestToggleStar_Success проверяет сценарий: toggle star успех.
func TestToggleStar_Success(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	starred, err := svc.ToggleStar(ctx, channelID, ownerID)
	require.NoError(t, err)
	assert.True(t, starred)

	starred, err = svc.ToggleStar(ctx, channelID, ownerID)
	require.NoError(t, err)
	assert.False(t, starred)
}

// TestUnmuteChannel_Success проверяет сценарий: unmute channel успех.
func TestUnmuteChannel_Success(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	_, err := svc.MuteChannel(ctx, channelID, ownerID, "1h")
	require.NoError(t, err)
	require.NoError(t, svc.UnmuteChannel(ctx, channelID, ownerID))

	muted, err := svc.IsChannelMuted(ctx, channelID, ownerID)
	require.NoError(t, err)
	assert.False(t, muted)
}
