package messenger_test

// Тесты модуля мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIncrement_SkipsSender проверяет, что sender не получает инкремент непрочитанных.
func TestIncrement_SkipsSender(t *testing.T) {
	_, rdb := newTestRedis(t)
	svc := messenger.NewUnreadService(rdb)
	ctx := context.Background()

	memberIDs := []string{"user-1", "user-2", "user-3"}
	senderID := "user-1"

	err := svc.Increment(ctx, "channel-1", memberIDs, senderID)
	require.NoError(t, err)

	// sender не должен иметь непрочитанных
	counts, err := svc.GetAll(ctx, "user-1")
	require.NoError(t, err)
	assert.Empty(t, counts, "sender не должен получать unread")

	// остальные должны иметь 1 непрочитанное
	counts2, err := svc.GetAll(ctx, "user-2")
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts2["channel-1"])
}

// TestIncrement_AllMembers проверяет, что все non-sender участники получают +1.
func TestIncrement_AllMembers(t *testing.T) {
	_, rdb := newTestRedis(t)
	svc := messenger.NewUnreadService(rdb)
	ctx := context.Background()

	memberIDs := []string{"user-a", "user-b", "user-c"}
	senderID := "user-a"

	// Два инкремента
	err := svc.Increment(ctx, "ch-1", memberIDs, senderID)
	require.NoError(t, err)
	err = svc.Increment(ctx, "ch-1", memberIDs, senderID)
	require.NoError(t, err)

	for _, uid := range []string{"user-b", "user-c"} {
		counts, err := svc.GetAll(ctx, uid)
		require.NoError(t, err)
		assert.Equal(t, int64(2), counts["ch-1"], "участник %s должен иметь 2 непрочитанных", uid)
	}
}

// TestMarkRead_ResetsToZero проверяет, что MarkRead удаляет запись (HDel).
func TestMarkRead_ResetsToZero(t *testing.T) {
	_, rdb := newTestRedis(t)
	svc := messenger.NewUnreadService(rdb)
	ctx := context.Background()

	// Инкрементируем
	err := svc.Increment(ctx, "ch-1", []string{"user-1", "user-2"}, "user-2")
	require.NoError(t, err)

	// Проверяем, что есть непрочитанные
	counts, err := svc.GetAll(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts["ch-1"])

	// Сбрасываем
	err = svc.MarkRead(ctx, "user-1", "ch-1")
	require.NoError(t, err)

	// Проверяем, что сброшено
	counts, err = svc.GetAll(ctx, "user-1")
	require.NoError(t, err)
	assert.Zero(t, counts["ch-1"], "после MarkRead счётчик должен быть 0")
}

// TestGetAll_ReturnsAllChannels проверяет, что GetAll возвращает все каналы с counts.
func TestGetAll_ReturnsAllChannels(t *testing.T) {
	_, rdb := newTestRedis(t)
	svc := messenger.NewUnreadService(rdb)
	ctx := context.Background()

	// Инкрементируем для нескольких каналов
	err := svc.Increment(ctx, "ch-1", []string{"user-1"}, "other")
	require.NoError(t, err)
	err = svc.Increment(ctx, "ch-2", []string{"user-1"}, "other")
	require.NoError(t, err)
	err = svc.Increment(ctx, "ch-2", []string{"user-1"}, "other")
	require.NoError(t, err)
	err = svc.Increment(ctx, "ch-3", []string{"user-1"}, "other")
	require.NoError(t, err)

	counts, err := svc.GetAll(ctx, "user-1")
	require.NoError(t, err)
	assert.Len(t, counts, 3, "должно быть 3 канала")
	assert.Equal(t, int64(1), counts["ch-1"])
	assert.Equal(t, int64(2), counts["ch-2"])
	assert.Equal(t, int64(1), counts["ch-3"])
}
