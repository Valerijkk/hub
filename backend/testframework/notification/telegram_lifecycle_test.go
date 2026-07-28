package notification_test

// Тесты модуля уведомлений: интеграционные и unit-сценарии testframework.

import (
	"context"
	"testing"
	"time"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTelegramSender_NilWhenTokenEmpty проверяет сценарий: new Telegram sender nil when token пустой.
func TestNewTelegramSender_NilWhenTokenEmpty(t *testing.T) {
	assert.Nil(t, notification.NewTelegramSender(""))
}

// TestTelegramSender_NilSafeMethods проверяет сценарий: Telegram sender nil safe methods.
func TestTelegramSender_NilSafeMethods(t *testing.T) {
	var sender *notification.TelegramSender
	assert.NotPanics(t, func() {
		sender.Start(context.Background())
		sender.Stop()
		sender.Send(1, "t", "b")
	})
}

// TestTelegramSender_SendAndQueueOverflow проверяет сценарий: Telegram sender отправка and queue overflow.
func TestTelegramSender_SendAndQueueOverflow(t *testing.T) {
	sender := notification.ExportNewTestTelegramSender(1)
	sender.Send(100, "title", "body")
	sender.Send(200, "drop", "me") // очередь переполнена — срабатывает ветка по умолчанию

	// W7-A1-007: первая в очереди, вторая дропнута.
	msg, ok := notification.ExportTelegramTryRecv(sender)
	require.True(t, ok, "first message must be queued")
	assert.Equal(t, int64(100), msg.ChatID)
	assert.Equal(t, "title", msg.Title)
	_, ok = notification.ExportTelegramTryRecv(sender)
	assert.False(t, ok, "overflow message must be dropped")
}

// TestTelegramSender_StartStopWithoutWorkerSend проверяет сценарий: Telegram sender start stop without worker отправка.
func TestTelegramSender_StartStopWithoutWorkerSend(t *testing.T) {
	sender := notification.ExportNewTestTelegramSender(2)
	ctx, cancel := context.WithCancel(context.Background())
	assert.NotPanics(t, func() {
		sender.Start(ctx)
		sender.Stop()
		cancel()
		time.Sleep(5 * time.Millisecond)
	})
}

// TestNotificationService_SendTelegram проверяет бизнес-логику: notification сервис отправка Telegram.
func TestNotificationService_SendTelegram(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)

	_, err := notification.ExportRepositoryDB(repo).User.UpdateOneID(userID).SetTelegramID("987654321").Save(ctx)
	require.NoError(t, err)

	sender := notification.ExportNewTestTelegramSender(1)
	svc = notification.NewNotificationService(repo, nil, sender, nil)

	notification.ExportSendTelegram(svc, ctx, userID, "Alert", "Details")
	msg, ok := notification.ExportTelegramTryRecv(sender)
	require.True(t, ok, "expected telegram message in queue")
	assert.Equal(t, int64(987654321), msg.ChatID)
	assert.Equal(t, "Alert", msg.Title)
}

// TestNotificationService_SendTelegram_SkipsWithoutTelegramID проверяет бизнес-логику: notification сервис отправка Telegram skips without Telegram id.
func TestNotificationService_SendTelegram_SkipsWithoutTelegramID(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	sender := notification.ExportNewTestTelegramSender(1)
	svc = notification.NewNotificationService(repo, nil, sender, nil)

	notification.ExportSendTelegram(svc, ctx, userID, "Alert", "Details")
	_, ok := notification.ExportTelegramTryRecv(sender)
	assert.False(t, ok, "unexpected telegram message")
}

// TestNotificationService_SendTelegram_InvalidTelegramID проверяет отклонение при notification сервис отправка Telegram некорректный Telegram id.
func TestNotificationService_SendTelegram_InvalidTelegramID(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	invalid := "not-a-number"
	_, err := notification.ExportRepositoryDB(repo).User.UpdateOneID(userID).SetTelegramID(invalid).Save(ctx)
	require.NoError(t, err)

	sender := notification.ExportNewTestTelegramSender(1)
	svc = notification.NewNotificationService(repo, nil, sender, nil)

	notification.ExportSendTelegram(svc, ctx, userID, "Alert", "Details")
	_, ok := notification.ExportTelegramTryRecv(sender)
	assert.False(t, ok, "unexpected telegram message for invalid id")
}

// TestNotificationService_SendTelegram_MissingUser проверяет отклонение при notification сервис отправка Telegram отсутствующий user.
func TestNotificationService_SendTelegram_MissingUser(t *testing.T) {
	svc, repo, ctx, _ := newNotificationTestEnv(t)
	sender := notification.ExportNewTestTelegramSender(1)
	svc = notification.NewNotificationService(repo, nil, sender, nil)

	assert.NotPanics(t, func() {
		notification.ExportSendTelegram(svc, ctx, uuid.New(), "Alert", "Details")
	})
	// W7-A1-007: missing user → ничего не ставится в очередь.
	_, ok := notification.ExportTelegramTryRecv(sender)
	assert.False(t, ok, "no telegram message for missing user")
}
