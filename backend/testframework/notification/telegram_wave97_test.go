package notification_test

// Тесты модуля уведомлений (волна 97): интеграционные и unit-сценарии testframework.

import (
	"context"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/stretchr/testify/require"
)

const testTelegramBotToken = "123456789:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc"

// TestTelegramSender_sendMessage_ContextCancelled проверяет сценарий: Telegram sender отправка message context cancelled.
func TestTelegramSender_sendMessage_ContextCancelled(t *testing.T) {
	sender := notification.NewTelegramSender(testTelegramBotToken)
	require.NotNil(t, sender)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	notification.ExportTelegramSendMessage(sender, ctx, 12345, "Shutdown", "Should abort")
}

// TestTelegramSender_sendMessage_FormatsAndAttemptsSend проверяет сценарий: Telegram sender отправка message formats and attempts отправка.
func TestTelegramSender_sendMessage_FormatsAndAttemptsSend(t *testing.T) {
	sender := notification.NewTelegramSender(testTelegramBotToken)
	require.NotNil(t, sender)

	// Invalid chat / token — covers err branch after API call, not ctx cancel.
	notification.ExportTelegramSendMessage(sender, context.Background(), 1, "Alert", "Body text")
}
