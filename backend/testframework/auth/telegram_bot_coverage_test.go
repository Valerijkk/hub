package auth_test

// Тесты модуля аутентификации: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTelegramBotAuth_StartPolling_NilReceiver проверяет сценарий: Telegram бот auth start polling nil receiver.
func TestTelegramBotAuth_StartPolling_NilReceiver(t *testing.T) {
	var bot *auth.TelegramBotAuth
	assert.NotPanics(t, func() {
		bot.StartPolling(context.Background())
	})
}

// TestTelegramBotAuth_VerifyCode_NilReceiver проверяет сценарий: Telegram бот auth verify code nil receiver.
func TestTelegramBotAuth_VerifyCode_NilReceiver(t *testing.T) {
	var bot *auth.TelegramBotAuth
	_, err := bot.VerifyCode(context.Background(), "123456", "")
	require.Error(t, err)
}
