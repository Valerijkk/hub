package notification_test

// Тесты модуля уведомлений (волна 119): интеграционные и unit-сценарии testframework.

import (
	"context"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/config"
)

// TestPushSender_NilStartStop проверяет сценарий: push sender nil start stop.
func TestPushSender_NilStartStop(t *testing.T) {
	var nilPush *notification.PushSender
	assert.NotPanics(t, func() {
		nilPush.Start(context.Background())
		nilPush.Stop()
	})
}

// TestPushSender_QueueOverflow проверяет сценарий: push sender queue overflow.
func TestPushSender_QueueOverflow(t *testing.T) {
	_, repo, _, userID := newNotificationTestEnv(t)
	priv, pub, err := webpush.GenerateVAPIDKeys()
	require.NoError(t, err)

	push := notification.NewPushSender(notification.ExportRepositoryDB(repo), config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)

	for i := 0; i < 201; i++ {
		push.Send(userID, "title", "body", "/")
	}
}

// TestPushSender_sendToUser_NoSubscriptions проверяет сценарий: push sender отправка to user no subscriptions.
func TestPushSender_sendToUser_NoSubscriptions(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	priv, pub, err := webpush.GenerateVAPIDKeys()
	require.NoError(t, err)

	push := notification.NewPushSender(notification.ExportRepositoryDB(repo), config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)

	assert.NotPanics(t, func() {
		notification.ExportPushSendToUser(push, ctx, userID, "t", "b", "/")
	})
}
