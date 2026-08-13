package notification_test

// Тесты модуля уведомлений: интеграционные и unit-сценарии testframework.

import (
	"context"
	"testing"
	"time"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/valerii/hub/backend/internal/platform/config"
)

// TestNewPushSender_NilWhenKeysMissing проверяет отклонение при new push sender nil when keys отсутствующий.
func TestNewPushSender_NilWhenKeysMissing(t *testing.T) {
	_, repo, _, _ := newNotificationTestEnv(t)
	assert.Nil(t, notification.NewPushSender(repo, config.WebPushConfig{}))
}

// TestPushSender_VAPIDPublicKeyAndNilSafeSend проверяет сценарий: push sender vapid public key and nil safe отправка.
func TestPushSender_VAPIDPublicKeyAndNilSafeSend(t *testing.T) {
	_, repo, _, userID := newNotificationTestEnv(t)
	priv, pub, err := webpush.GenerateVAPIDKeys()
	require.NoError(t, err)

	push := notification.NewPushSender(repo, config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)
	assert.Equal(t, pub, push.VAPIDPublicKey())

	var nilPush *notification.PushSender
	assert.NotPanics(t, func() { nilPush.Send(userID, "t", "b", "/") })
	assert.Equal(t, "", nilPush.VAPIDPublicKey())
}

// TestPushSender_StartStopLifecycle проверяет сценарий: push sender start stop жизненный цикл.
func TestPushSender_StartStopLifecycle(t *testing.T) {
	_, repo, _, userID := newNotificationTestEnv(t)
	priv, pub, err := webpush.GenerateVAPIDKeys()
	require.NoError(t, err)

	push := notification.NewPushSender(repo, config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)

	ctx, cancel := context.WithCancel(context.Background())
	push.Start(ctx)
	push.Send(userID, "title", "body", "/")
	push.Stop()
	cancel()

	time.Sleep(10 * time.Millisecond)
}
