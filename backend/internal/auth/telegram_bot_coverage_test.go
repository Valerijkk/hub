package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelegramBotAuth_StartPolling_NilReceiver(t *testing.T) {
	var bot *TelegramBotAuth
	assert.NotPanics(t, func() {
		bot.StartPolling(context.Background())
	})
}

func TestTelegramBotAuth_VerifyCode_NilReceiver(t *testing.T) {
	var bot *TelegramBotAuth
	_, err := bot.VerifyCode(context.Background(), "123456", "")
	require.Error(t, err)
}
