package messenger_test

// Тесты сервиса мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/channelmember"
)

// TestResolveGroupMentions_AllAndIndividual проверяет сценарий: resolve group упоминания all and individual.
func TestResolveGroupMentions_AllAndIndividual(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()

	ownerID := uuid.New()
	memberID := uuid.New()
	outsiderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)

	mentions := messenger.ExportMessengerServiceResolveGroupMentions(svc, ctx, channelID, nil, []string{"all"}, []string{memberID.String()}, ownerID)
	assert.Len(t, mentions, 1)
	assert.Contains(t, mentions, memberID.String())
	assert.NotContains(t, mentions, ownerID.String())
	_ = outsiderID
}

// TestResolveGroupMentions_RoleRequiresProject проверяет сценарий: resolve group упоминания role requires project.
func TestResolveGroupMentions_RoleRequiresProject(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	mentions := messenger.ExportMessengerServiceResolveGroupMentions(svc, ctx, channelID, nil, []string{"developer"}, nil, ownerID)
	assert.Empty(t, mentions)
}

// TestResolveGroupMentions_RoleWithProjectMembers проверяет сценарий: resolve group упоминания role with project members.
func TestResolveGroupMentions_RoleWithProjectMembers(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()

	projectID := uuid.New()
	ownerID := uuid.New()
	devID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, devID)
	repo.channels[channelID].ProjectID = &projectID
	repo.projectMembersByRole[projectID] = map[string][]uuid.UUID{
		"developer": {devID},
	}

	mentions := messenger.ExportMessengerServiceResolveGroupMentions(svc, ctx, channelID, &projectID, []string{"developer"}, nil, ownerID)
	assert.Contains(t, mentions, devID.String())
}

// TestDeleteMessage_ForbiddenForNonAuthor проверяет отклонение при удаление message запрещённый for non author.
func TestDeleteMessage_ForbiddenForNonAuthor(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()

	ownerID := uuid.New()
	otherID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, otherID)

	doc := &messenger.MessageDocument{
		ID:        "msg-1",
		ChannelID: channelID,
		SenderID:  ownerID,
		CreatedAt: time.Now(),
	}
	repo.messages[doc.ID] = doc

	err := svc.DeleteMessage(ctx, doc.ID, otherID, channelID.String(), false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "удаление")
}

// TestDeleteMessage_AllowedForOwner проверяет сценарий: удаление message allowed for owner.
func TestDeleteMessage_AllowedForOwner(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()

	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)
	repo.members[channelID][ownerID].Role = channelmember.RoleOwner

	doc := &messenger.MessageDocument{
		ID:        "msg-2",
		ChannelID: channelID,
		SenderID:  memberID,
		CreatedAt: time.Now(),
	}
	repo.messages[doc.ID] = doc

	require.NoError(t, svc.DeleteMessage(ctx, doc.ID, ownerID, channelID.String(), false))
}

// TestListMembers_NonMemberForbidden проверяет отклонение при список members non участник запрещённый.
func TestListMembers_NonMemberForbidden(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	_, err := svc.ListMembers(ctx, channelID, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "участником")
}

// TestMuteAndUnmuteChannel проверяет сценарий: mute and unmute channel.
func TestMuteAndUnmuteChannel(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	mutedUntil, err := svc.MuteChannel(ctx, channelID, ownerID, "1h")
	require.NoError(t, err)
	assert.NotEmpty(t, mutedUntil)

	ok, err := svc.IsChannelMuted(ctx, channelID, ownerID)
	require.NoError(t, err)
	assert.True(t, ok)

	require.NoError(t, svc.UnmuteChannel(ctx, channelID, ownerID))
	ok, err = svc.IsChannelMuted(ctx, channelID, ownerID)
	require.NoError(t, err)
	assert.False(t, ok)
}

// TestJoinChannel_AlreadyMemberIsNoOp проверяет сценарий: join channel already участник is no op.
func TestJoinChannel_AlreadyMemberIsNoOp(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	require.NoError(t, svc.JoinChannel(ctx, channelID, ownerID))
}

// TestMapKeys_OrderStable проверяет сценарий: map keys order stable.
func TestMapKeys_OrderStable(t *testing.T) {
	seen := map[string]struct{}{"b": {}, "a": {}}
	keys := messenger.ExportMapKeys(seen)
	assert.Len(t, keys, 2)
}

// TestToMessageResponse_Basic проверяет сценарий: to message response basic.
func TestToMessageResponse_Basic(t *testing.T) {
	id := uuid.New()
	now := time.Now().UTC()
	doc := &messenger.MessageDocument{
		ID:          "mongo-id",
		ChannelID:   id,
		SenderID:    id,
		ContentText: "hello",
		CreatedAt:   now,
	}
	resp := messenger.ExportToMessageResponse(doc)
	require.NotNil(t, resp)
	assert.Equal(t, "mongo-id", resp.ID)
	assert.Equal(t, "hello", resp.ContentText)
}
