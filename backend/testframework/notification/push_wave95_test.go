package notification_test

// Тесты модуля уведомлений (волна 95): интеграционные и unit-сценарии testframework.

import (
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/platform/config"
)

// TestPushSender_sendToSubscription_ValidEndpointSendFails проверяет сценарий: push sender отправка to subscription корректный endpoint отправка fails.
func TestPushSender_sendToSubscription_ValidEndpointSendFails(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	priv, pub, err := webpush.GenerateVAPIDKeys()
	require.NoError(t, err)

	push := notification.NewPushSender(notification.ExportRepositoryDB(repo), config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)

	sub := &ent.PushSubscription{
		ID:        uuid.New(),
		UserID:    userID,
		Endpoint:  "https://push.example.com/sub/wave95",
		P256dhKey: "invalid-p256",
		AuthKey:   "invalid-auth",
	}
	payload := []byte(`{"title":"t","body":"b"}`)
	notification.ExportPushSendToSubscription(push, ctx, sub, payload)
}

// TestPushSender_sendToSubscription_UnsafeEndpointDeleted проверяет сценарий: push sender отправка to subscription unsafe endpoint deleted.
func TestPushSender_sendToSubscription_UnsafeEndpointDeleted(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	priv, pub, err := webpush.GenerateVAPIDKeys()
	require.NoError(t, err)

	push := notification.NewPushSender(notification.ExportRepositoryDB(repo), config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)

	saved, err := notification.ExportRepositoryDB(repo).PushSubscription.Create().
		SetUserID(userID).
		SetEndpoint("file:///tmp/push-dead").
		SetP256dhKey("k").
		SetAuthKey("a").
		SetUserAgent("ua").
		Save(ctx)
	require.NoError(t, err)

	notification.ExportPushSendToSubscription(push, ctx, saved, []byte(`{"title":"x"}`))

	exists, err := notification.ExportRepositoryDB(repo).PushSubscription.Query().Exist(ctx)
	require.NoError(t, err)
	require.False(t, exists)
}
