package notification_test

// Тесты модуля уведомлений (волна 95): интеграционные и unit-сценарии testframework.

import (
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/pushsubscription"
	"github.com/valerii/hub/backend/internal/platform/config"
)

// TestPushSender_sendToSubscription_ValidEndpointSendFails проверяет сценарий: push sender отправка to subscription корректный endpoint отправка fails.
func TestPushSender_sendToSubscription_ValidEndpointSendFails(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	priv, pub, err := webpush.GenerateVAPIDKeys()
	require.NoError(t, err)

	push := notification.NewPushSender(repo, config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)

	endpoint := "https://push.example.com/sub/wave95-" + uuid.New().String()
	saved, err := notification.ExportRepositoryDB(repo).PushSubscription.Create().
		SetUserID(userID).
		SetEndpoint(endpoint).
		SetP256dhKey("invalid-p256").
		SetAuthKey("invalid-auth").
		SetUserAgent("ua").
		Save(ctx)
	require.NoError(t, err)

	payload := []byte(`{"title":"t","body":"b"}`)
	require.NotPanics(t, func() {
		notification.ExportPushSendToSubscription(push, ctx, notification.ExportPushSubFromEnt(saved), payload)
	})

	// W7-A1-008: https-endpoint при ошибке отправки не удаляется.
	exists, err := notification.ExportRepositoryDB(repo).PushSubscription.Query().
		Where(pushsubscription.IDEQ(saved.ID)).
		Exist(ctx)
	require.NoError(t, err)
	require.True(t, exists, "valid https subscription must remain after send failure")
}

// TestPushSender_sendToSubscription_UnsafeEndpointDeleted проверяет сценарий: push sender отправка to subscription unsafe endpoint deleted.
func TestPushSender_sendToSubscription_UnsafeEndpointDeleted(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	priv, pub, err := webpush.GenerateVAPIDKeys()
	require.NoError(t, err)

	push := notification.NewPushSender(repo, config.WebPushConfig{
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

	notification.ExportPushSendToSubscription(push, ctx, notification.ExportPushSubFromEnt(saved), []byte(`{"title":"x"}`))

	exists, err := notification.ExportRepositoryDB(repo).PushSubscription.Query().Exist(ctx)
	require.NoError(t, err)
	require.False(t, exists)
}
