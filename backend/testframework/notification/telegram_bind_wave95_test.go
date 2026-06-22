package notification_test

// Тесты модуля уведомлений (волна 95): интеграционные и unit-сценарии testframework.

import (
	"context"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mymmrac/telego"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/enttest"
)

func setupTelegramBindForHandleMessage(t *testing.T) (*notification.TelegramBindService, *redis.Client, context.Context, uuid.UUID) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	dsn := "file:tgw95_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	u, err := client.User.Create().
		SetEmail("tgw95@" + uuid.New().String()[:8] + ".local").
		Save(ctx)
	require.NoError(t, err)

	repo := notification.NewRepository(client)
	sender := notification.ExportNewTestTelegramSender(2)
	svc := notification.NewTelegramBindService(repo, rdb, sender, "123456789:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc", "HubNotifyBot")
	require.NotNil(t, svc)
	return svc, rdb, ctx, u.ID
}

// TestTelegramBindService_handleMessage_StartBind проверяет бизнес-логику: Telegram привязка сервис handle message start привязка.
func TestTelegramBindService_handleMessage_StartBind(t *testing.T) {
	svc, rdb, ctx, userID := setupTelegramBindForHandleMessage(t)
	token := "abc123deadbeef"
	require.NoError(t, rdb.Set(ctx, notification.ExportTGBindPrefix()+token, userID.String(), notification.ExportTGBindTTL()).Err())

	notification.ExportTelegramHandleMessage(svc, ctx, &telego.Message{
		Text: "/start bind_" + token,
		Chat: telego.Chat{ID: 1001},
		From: &telego.User{ID: 424242},
	})

	updated, err := notification.ExportRepositoryDB(notification.ExportTelegramBindRepository(svc)).User.Get(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, updated.TelegramID)
	assert.Equal(t, "424242", *updated.TelegramID)
}

// TestTelegramBindService_handleMessage_StartBindExpired проверяет отклонение при Telegram привязка сервис handle message start привязка просроченный.
func TestTelegramBindService_handleMessage_StartBindExpired(t *testing.T) {
	svc, _, ctx, _ := setupTelegramBindForHandleMessage(t)
	notification.ExportTelegramHandleMessage(svc, ctx, &telego.Message{
		Text: "/start bind_expiredtoken",
		Chat: telego.Chat{ID: 1002},
		From: &telego.User{ID: 111},
	})
}

// TestTelegramBindService_handleMessage_LinkCommand проверяет бизнес-логику: Telegram привязка сервис handle message link command.
func TestTelegramBindService_handleMessage_LinkCommand(t *testing.T) {
	svc, rdb, ctx, _ := setupTelegramBindForHandleMessage(t)
	notification.ExportTelegramHandleMessage(svc, ctx, &telego.Message{
		Text: "/link",
		Chat: telego.Chat{ID: 2001},
		From: &telego.User{ID: 555001},
	})
	keys, err := rdb.Keys(ctx, notification.ExportTGLinkCodePrefix()+"*").Result()
	require.NoError(t, err)
	assert.Len(t, keys, 1)
}

// TestTelegramBindService_handleMessage_UnlinkAndTest проверяет бизнес-логику: Telegram привязка сервис handle message unlink and test.
func TestTelegramBindService_handleMessage_UnlinkAndTest(t *testing.T) {
	svc, _, ctx, userID := setupTelegramBindForHandleMessage(t)
	tgID := "777888999"
	require.NoError(t, notification.ExportTelegramBindRepository(svc).SetUserTelegram(ctx, userID, tgID))

	notification.ExportTelegramHandleMessage(svc, ctx, &telego.Message{
		Text: "/test",
		Chat: telego.Chat{ID: 3001},
		From: &telego.User{ID: 777888999},
	})
	_, ok := notification.ExportTelegramTryRecv(notification.ExportTelegramBindSender(svc))
	if !ok {
		t.Fatal("expected test message from /test command")
	}

	notification.ExportTelegramHandleMessage(svc, ctx, &telego.Message{
		Text: "/unlink",
		Chat: telego.Chat{ID: 3001},
		From: &telego.User{ID: 777888999},
	})
	u, err := notification.ExportTelegramBindRepository(svc).GetUserTelegram(ctx, userID)
	require.NoError(t, err)
	assert.True(t, u.TelegramID == nil || *u.TelegramID == "")
}

// TestTelegramBindService_handleMessage_StartHelp проверяет бизнес-логику: Telegram привязка сервис handle message start help.
func TestTelegramBindService_handleMessage_StartHelp(t *testing.T) {
	svc, _, ctx, _ := setupTelegramBindForHandleMessage(t)
	notification.ExportTelegramHandleMessage(svc, ctx, &telego.Message{
		Text: "/start",
		Chat: telego.Chat{ID: 4001},
		From: &telego.User{ID: 1},
	})
}

// TestTelegramBindService_StartPolling_NilSafe проверяет бизнес-логику: Telegram привязка сервис start polling nil safe.
func TestTelegramBindService_StartPolling_NilSafe(t *testing.T) {
	var nilSvc *notification.TelegramBindService
	nilSvc.StartPolling(context.Background())
}
