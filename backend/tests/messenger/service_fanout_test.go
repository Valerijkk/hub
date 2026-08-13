package messenger_test

// Тесты сервиса мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSendMessage_UnreadFanOut_ExcludesSender проверяет сценарий: отправка message непрочитанные fan out excludes sender.
func TestSendMessage_UnreadFanOut_ExcludesSender(t *testing.T) {
	repo := newMockRepository()
	_, rdb := newTestRedis(t)
	unread := messenger.NewUnreadService(rdb)
	svc := messenger.NewMessengerService(repo, nil, nil, unread)
	ctx := context.Background()

	ownerID := uuid.New()
	memberB := uuid.New()
	memberC := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberB, memberC)

	_, err := svc.SendMessage(ctx, channelID, ownerID, messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "fan-out"},
		ContentText: "fan-out",
	})
	require.NoError(t, err)

	ownerCounts, err := unread.GetAll(ctx, ownerID.String())
	require.NoError(t, err)
	assert.Empty(t, ownerCounts, "sender must not receive unread")

	for _, uid := range []uuid.UUID{memberB, memberC} {
		counts, err := unread.GetAll(ctx, uid.String())
		require.NoError(t, err)
		assert.Equal(t, int64(1), counts[channelID.String()], "member %s should have 1 unread", uid)
	}
}

// TestSendMessage_UnreadFanOut_SecondMessage проверяет сценарий: отправка message непрочитанные fan out second message.
func TestSendMessage_UnreadFanOut_SecondMessage(t *testing.T) {
	repo := newMockRepository()
	_, rdb := newTestRedis(t)
	unread := messenger.NewUnreadService(rdb)
	svc := messenger.NewMessengerService(repo, nil, nil, unread)
	ctx := context.Background()

	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)

	for i := 0; i < 2; i++ {
		_, err := svc.SendMessage(ctx, channelID, ownerID, messenger.SendMessageRequest{
			ContentJSON: map[string]interface{}{"text": "msg"},
			ContentText: "msg",
		})
		require.NoError(t, err)
	}

	counts, err := unread.GetAll(ctx, memberID.String())
	require.NoError(t, err)
	assert.Equal(t, int64(2), counts[channelID.String()])
}
