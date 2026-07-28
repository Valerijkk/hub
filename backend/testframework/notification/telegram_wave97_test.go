package notification_test

// Тесты модуля уведомлений (волна 97): интеграционные и unit-сценарии testframework.

import (
	"context"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTelegramBotToken = "123456789:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc"

// TestTelegramSender_sendMessage_ContextCancelled проверяет сценарий: Telegram sender отправка message context cancelled.
func TestTelegramSender_sendMessage_ContextCancelled(t *testing.T) {
	sender := notification.NewTelegramSender(testTelegramBotToken)
	require.NotNil(t, sender)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// W7-A1-006: cancel path не паникует.
	assert.NotPanics(t, func() {
		notification.ExportTelegramSendMessage(sender, ctx, 12345, "Shutdown", "Should abort")
	})
}

// TestTelegramSender_sendMessage_FormatsAndAttemptsSend проверяет сценарий: Telegram sender отправка message formats and attempts отправка.
func TestTelegramSender_sendMessage_FormatsAndAttemptsSend(t *testing.T) {
	sender := notification.NewTelegramSender(testTelegramBotToken)
	require.NotNil(t, sender)

	// W7-A1-006: format observable + err branch (invalid token/chat) без panic.
	formatted := notification.ExportFormatTelegramMessage("Alert", "Body text")
	assert.Contains(t, formatted, "Alert")
	assert.Contains(t, formatted, "Body text")

	assert.NotPanics(t, func() {
		notification.ExportTelegramSendMessage(sender, context.Background(), 1, "Alert", "Body text")
	})
}
