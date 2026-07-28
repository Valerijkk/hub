package messenger_test

// Тесты модуля мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRedis создаёт miniredis и клиент для тестов.
func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return mr, rdb
}

// TestHeartbeat_SetsScore проверяет, что heartbeat записывает score в sorted set.
func TestHeartbeat_SetsScore(t *testing.T) {
	mr, rdb := newTestRedis(t)
	svc := messenger.NewPresenceService(rdb)
	ctx := context.Background()

	err := svc.Heartbeat(ctx, "user-1")
	require.NoError(t, err)

	// Проверяем, что запись существует в sorted set
	members, err := mr.ZMembers(messenger.ExportPresenceOnlineKey())
	require.NoError(t, err)
	assert.Contains(t, members, "user-1")
}

// TestGetStatus_Online проверяет, что свежий heartbeat возвращает "online".
func TestGetStatus_Online(t *testing.T) {
	_, rdb := newTestRedis(t)
	svc := messenger.NewPresenceService(rdb)
	ctx := context.Background()

	err := svc.Heartbeat(ctx, "user-1")
	require.NoError(t, err)

	status, err := svc.GetStatus(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "online", status)
}

// TestGetStatus_Away проверяет, что heartbeat 6 минут назад возвращает "away".
func TestGetStatus_Away(t *testing.T) {
	mr, rdb := newTestRedis(t)
	svc := messenger.NewPresenceService(rdb)
	ctx := context.Background()

	// Записываем heartbeat 6 минут назад
	sixMinAgo := time.Now().Add(-6 * time.Minute).Unix()
	mr.ZAdd(messenger.ExportPresenceOnlineKey(), float64(sixMinAgo), "user-1")

	status, err := svc.GetStatus(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "away", status)
}

// TestGetStatus_Offline проверяет, что отсутствие записи возвращает "offline".
func TestGetStatus_Offline(t *testing.T) {
	_, rdb := newTestRedis(t)
	svc := messenger.NewPresenceService(rdb)
	ctx := context.Background()

	status, err := svc.GetStatus(ctx, "nonexistent-user")
	require.NoError(t, err)
	assert.Equal(t, "offline", status)
}

// TestCleanupStale проверяет, что старые записи удаляются при очистке.
func TestCleanupStale(t *testing.T) {
	mr, rdb := newTestRedis(t)
	svc := messenger.NewPresenceService(rdb)
	ctx := context.Background()

	// Свежий heartbeat
	err := svc.Heartbeat(ctx, "fresh-user")
	require.NoError(t, err)

	// Старый heartbeat (2 минуты назад, > 90 секунд порога)
	twoMinAgo := time.Now().Add(-2 * time.Minute).Unix()
	mr.ZAdd(messenger.ExportPresenceOnlineKey(), float64(twoMinAgo), "stale-user")

	// Очистка
	err = svc.CleanupStale(ctx)
	require.NoError(t, err)

	// Свежий должен остаться
	members, _ := mr.ZMembers(messenger.ExportPresenceOnlineKey())
	assert.Contains(t, members, "fresh-user")
	assert.NotContains(t, members, "stale-user")
}

// TestSetGetCustomStatus проверяет запись и чтение кастомного текста статуса.
func TestSetGetCustomStatus(t *testing.T) {
	_, rdb := newTestRedis(t)
	svc := messenger.NewPresenceService(rdb)
	ctx := context.Background()

	// Устанавливаем кастомный статус
	err := svc.SetCustomStatus(ctx, "user-1", "На обеде", "")
	require.NoError(t, err)

	// Читаем
	text, err := svc.GetCustomStatus(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "На обеде", text)

	// Очищаем статус
	err = svc.SetCustomStatus(ctx, "user-1", "", "")
	require.NoError(t, err)

	text, err = svc.GetCustomStatus(ctx, "user-1")
	require.NoError(t, err)
	assert.Empty(t, text)
}

// TestGetBulkStatus проверяет получение статусов для нескольких пользователей.
func TestGetBulkStatus(t *testing.T) {
	mr, rdb := newTestRedis(t)
	svc := messenger.NewPresenceService(rdb)
	ctx := context.Background()

	// online пользователь
	err := svc.Heartbeat(ctx, "online-user")
	require.NoError(t, err)

	// away пользователь (6 минут назад)
	sixMinAgo := time.Now().Add(-6 * time.Minute).Unix()
	mr.ZAdd(messenger.ExportPresenceOnlineKey(), float64(sixMinAgo), "away-user")

	statuses, err := svc.GetBulkStatus(ctx, []string{"online-user", "away-user", "offline-user"})
	require.NoError(t, err)

	assert.Equal(t, "online", statuses["online-user"])
	assert.Equal(t, "away", statuses["away-user"])
	assert.Equal(t, "offline", statuses["offline-user"])
}

// TestGetBulkStatus_RedisDown_ReturnsError — W3-A3-004 / P2-BE-10:
// ошибка pipeline не глотается (не silent offline).
func TestGetBulkStatus_RedisDown_ReturnsError(t *testing.T) {
	mr, rdb := newTestRedis(t)
	svc := messenger.NewPresenceService(rdb)
	ctx := context.Background()

	require.NoError(t, svc.Heartbeat(ctx, "u1"))
	mr.Close()
	_ = rdb.Close()

	_, err := svc.GetBulkStatus(ctx, []string{"u1", "u2"})
	require.Error(t, err)
}
