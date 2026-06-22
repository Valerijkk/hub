package notification_test

// Тесты модуля уведомлений: интеграционные и unit-сценарии testframework.

import (
	"context"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/enttest"
)

// TestTelegramBindService_CreateBindToken проверяет бизнес-логику: Telegram привязка сервис создание привязка token.
func TestTelegramBindService_CreateBindToken(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	dsn := "file:tgbind_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	repo := notification.NewRepository(client)
	svc := notification.NewTelegramBindService(repo, rdb, nil, "123456789:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc", "HubNotifyBot")
	require.NotNil(t, svc)

	userID := uuid.New()
	resp, err := svc.CreateBindToken(context.Background(), userID)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.BindToken)
	assert.Contains(t, resp.DeepLink, "bind_")
}

// TestTelegramBindService_ConfirmLinkCode проверяет бизнес-логику: Telegram привязка сервис confirm link code.
func TestTelegramBindService_ConfirmLinkCode(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	dsn := "file:tgbind2_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	u, err := client.User.Create().SetEmail("tg@" + uuid.New().String()[:8] + ".local").Save(ctx)
	require.NoError(t, err)

	repo := notification.NewRepository(client)
	svc := notification.NewTelegramBindService(repo, rdb, nil, "123456789:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc", "")
	require.NotNil(t, svc)

	require.NoError(t, mr.Set("tg:linkcode:654321", "99887766"))
	err = svc.ConfirmLinkCode(ctx, u.ID, "654321")
	require.NoError(t, err)

	updated, err := client.User.Get(ctx, u.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.TelegramID)
	assert.Equal(t, "99887766", *updated.TelegramID)
}

// TestTelegramBindService_SendTest проверяет бизнес-логику: Telegram привязка сервис отправка test.
func TestTelegramBindService_SendTest(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	dsn := "file:tgbind3_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	u, err := client.User.Create().
		SetEmail("sendtest@" + uuid.New().String()[:8] + ".local").
		SetTelegramID("555666777").
		Save(ctx)
	require.NoError(t, err)

	repo := notification.NewRepository(client)
	sender := notification.ExportNewTestTelegramSender(1)
	svc := notification.NewTelegramBindService(repo, rdb, sender, "123456789:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc", "")
	require.NotNil(t, svc)

	require.NoError(t, svc.SendTest(ctx, u.ID))
	msg, ok := notification.ExportTelegramTryRecv(sender)
	require.True(t, ok, "expected test message in queue")
	assert.Equal(t, int64(555666777), msg.ChatID)
}

// TestTelegramBindService_SendTest_NotLinked проверяет бизнес-логику: Telegram привязка сервис отправка test not linked.
func TestTelegramBindService_SendTest_NotLinked(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	dsn := "file:tgbind4_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	u, err := client.User.Create().SetEmail("nolink@" + uuid.New().String()[:8] + ".local").Save(ctx)
	require.NoError(t, err)

	repo := notification.NewRepository(client)
	sender := notification.ExportNewTestTelegramSender(1)
	svc := notification.NewTelegramBindService(repo, rdb, sender, "123456789:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc", "")
	require.NotNil(t, svc)

	err = svc.SendTest(ctx, u.ID)
	require.ErrorIs(t, err, notification.ErrTelegramNotLinked)
}
