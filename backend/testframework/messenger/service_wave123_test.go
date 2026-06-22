package messenger_test

// Тесты сервиса мессенджера (волна 123): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


// TestService_Wave123_SendBotMessage_ThreadReply проверяет бизнес-логику: сервис wave 123 отправка бот message thread reply.
func TestService_Wave123_SendBotMessage_ThreadReply(t *testing.T) {
	repo := newMockRepository()
	svc := messenger.ExportNewMessengerServiceTest(repo)
	botID := uuid.New()
	svc.SetSystemBotID(botID)

	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	parent, err := svc.SendBotMessage(context.Background(), channelID, messenger.SendMessageRequest{
		ContentText: "parent",
		ContentJSON: map[string]interface{}{"text": "parent"},
	})
	require.NoError(t, err)

	reply, err := svc.SendBotMessage(context.Background(), channelID, messenger.SendMessageRequest{
		ContentText: "reply",
		ContentJSON: map[string]interface{}{"text": "reply"},
		ThreadID:    parent.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, parent.ID, reply.ThreadID)
}

// TestService_Wave123_MuteChannel_AllDurations проверяет бизнес-логику: сервис wave 123 mute channel all durations.
func TestService_Wave123_MuteChannel_AllDurations(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	for _, dur := range []string{"2h", "8h", "1d", "forever"} {
		_, err := svc.MuteChannel(ctx, channelID, ownerID, dur)
		require.NoError(t, err, dur)
		require.NoError(t, svc.UnmuteChannel(ctx, channelID, ownerID))
	}
}

// TestService_Wave123_SetPromoteDependencies проверяет бизнес-логику: сервис wave 123 set promote dependencies.
func TestService_Wave123_SetPromoteDependencies(t *testing.T) {
	svc, _, _ := newTestService()
	svc.SetIssueThreadLinker(&stubIssueThreadLinker{})
	svc.SetPromotePermissionChecker(&stubPromoteChecker{allowed: true})
	assert.NotNil(t, messenger.ExportMessengerServiceIssueThreadLinker(svc))
	assert.NotNil(t, messenger.ExportMessengerServicePromoteCheck(svc))
}

type stubIssueThreadLinker struct{}

func (s *stubIssueThreadLinker) LinkThreadToIssue(_ context.Context, _, _, _ uuid.UUID, _ uuid.UUID, _ string) error {
	return nil
}

type stubPromoteChecker struct {
	allowed bool
}

func (s *stubPromoteChecker) CanCreateTaskInProject(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return s.allowed, nil
}

// TestService_Wave123_CreateInviteLink проверяет бизнес-логику: сервис wave 123 создание приглашение link.
func TestService_Wave123_CreateInviteLink(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	inv, err := svc.CreateInviteLink(ctx, channelID, ownerID, messenger.CreateInviteLinkRequest{MaxUses: 3, ExpiresIn: "1d"})
	require.NoError(t, err)
	assert.NotEmpty(t, inv.Token)

	list, err := svc.ListChannelInvites(ctx, channelID, ownerID)
	require.NoError(t, err)
	assert.NotEmpty(t, list)
}

// TestService_Wave123_ValidateInvite проверяет бизнес-логику: сервис wave 123 validate приглашение.
func TestService_Wave123_ValidateInvite(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	created, err := svc.CreateInviteLink(ctx, channelID, ownerID, messenger.CreateInviteLinkRequest{MaxUses: 1})
	require.NoError(t, err)

	valid, err := svc.ValidateInvite(ctx, created.Token)
	require.NoError(t, err)
	assert.True(t, valid.Valid)
}
