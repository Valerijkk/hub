package notification_test

// Тесты модуля уведомлений (волна 102): интеграционные и unit-сценарии testframework.

import (
	"context"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mymmrac/telego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/enttest"
)

// TestNewTelegramBindService_RepoServiceWhenBotUnavailable проверяет, что
// repository-backed API остаётся доступным при недоступной bot-конфигурации.
func TestNewTelegramBindService_RepoServiceWhenBotUnavailable(t *testing.T) {
	dsn := "file:tgbindcfg_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	repo := notification.NewRepository(client)

	assert.NotNil(t, notification.NewTelegramBindService(repo, nil, nil, "token", "bot"))
	assert.NotNil(t, notification.NewTelegramBindService(repo, nil, nil, "", "bot"))
	assert.NotNil(t, notification.NewTelegramBindService(repo, nil, nil, "not-a-valid-token", "bot"))
}

// TestTelegramBindService_handleUnlinkCommand_NotLinked проверяет бизнес-логику: Telegram привязка сервис handle unlink command not linked.
func TestTelegramBindService_handleUnlinkCommand_NotLinked(t *testing.T) {
	svc, _, ctx, _ := setupTelegramBindForHandleMessage(t)
	notification.ExportTelegramHandleUnlinkCommand(svc, ctx, 9001, "999888777")
}

// TestTelegramBindService_handleStartBind_TelegramIDTaken проверяет бизнес-логику: Telegram привязка сервис handle start привязка Telegram id taken.
func TestTelegramBindService_handleStartBind_TelegramIDTaken(t *testing.T) {
	svc, rdb, ctx, userID := setupTelegramBindForHandleMessage(t)

	other, err := notification.ExportRepositoryDB(notification.ExportTelegramBindRepository(svc)).User.Create().
		SetEmail("other-tg-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, notification.ExportTelegramBindRepository(svc).SetUserTelegram(ctx, other.ID, "555001"))

	token := "deadbeefcafebabe"
	require.NoError(t, rdb.Set(ctx, notification.ExportTGBindPrefix()+token, userID.String(), notification.ExportTGBindTTL()).Err())

	notification.ExportTelegramHandleMessage(svc, ctx, &telego.Message{
		Text: "/start bind_" + token,
		Chat: telego.Chat{ID: 5001},
		From: &telego.User{ID: 555001},
	})

	updated, err := notification.ExportRepositoryDB(notification.ExportTelegramBindRepository(svc)).User.Get(ctx, userID)
	require.NoError(t, err)
	assert.True(t, updated.TelegramID == nil || *updated.TelegramID == "")
}

// TestTelegramBindService_CreateBindToken_NilService проверяет бизнес-логику: Telegram привязка сервис создание привязка token nil сервис.
func TestTelegramBindService_CreateBindToken_NilService(t *testing.T) {
	var nilSvc *notification.TelegramBindService
	_, err := nilSvc.CreateBindToken(context.Background(), uuid.New())
	assert.ErrorIs(t, err, notification.ErrTelegramBindUnavailable)
}

// TestTelegramBindService_ConfirmLinkCode_InvalidLength проверяет отклонение при Telegram привязка сервис confirm link code некорректный length.
func TestTelegramBindService_ConfirmLinkCode_InvalidLength(t *testing.T) {
	svc, _, ctx, userID := setupTelegramBindForHandleMessage(t)
	err := svc.ConfirmLinkCode(ctx, userID, "12345")
	assert.ErrorIs(t, err, notification.ErrTelegramInvalidCode)
}
