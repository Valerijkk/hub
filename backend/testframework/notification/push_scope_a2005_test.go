package notification_test

// Тесты модуля уведомлений: A2-005 push subscribe hijack / unsubscribe scope.

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/pushsubscription"
	"github.com/valerii/hub/backend/internal/notification"
	"github.com/valerii/hub/backend/internal/platform/config"
)

// TestPushSubscription_RejectsForeignSubscribeAndUnsubscribe_A2_005 —
// чужой endpoint нельзя перепривязать (SetUserID) и нельзя удалить без UserID.
// Mutation: снять UserID из Where в Update/Delete → foreign hijack/unsub проходит → FAIL.
func TestPushSubscription_RejectsForeignSubscribeAndUnsubscribe_A2_005(t *testing.T) {
	_, repo, ctx, ownerID := newNotificationTestEnv(t)
	db := notification.ExportRepositoryDB(repo)

	attacker, err := db.User.Create().
		SetEmail("attacker-" + uuid.New().String()[:8] + "@test.com").
		SetName("Attacker").
		Save(ctx)
	require.NoError(t, err)

	priv, pub, err := generateTestVAPIDKeys(t)
	require.NoError(t, err)
	push := notification.NewPushSender(repo, config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)

	endpoint := "https://push.example.com/sub/a2005-owner"
	require.NoError(t, push.SaveSubscription(ctx, ownerID, endpoint, "owner-p256", "owner-auth", "owner-ua"))

	// Hijack: attacker tries SaveSubscription on owner's endpoint → Forbidden, no rebind.
	err = push.SaveSubscription(ctx, attacker.ID, endpoint, "hijack-p256", "hijack-auth", "hijack-ua")
	require.ErrorIs(t, err, notification.ErrSubscriptionForbidden)

	sub, err := db.PushSubscription.Query().
		Where(pushsubscription.Endpoint(endpoint)).
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, ownerID, sub.UserID, "owner must keep endpoint")
	assert.Equal(t, "owner-p256", sub.P256dhKey, "keys must not be overwritten by foreign subscribe")
	assert.Equal(t, "owner-auth", sub.AuthKey)

	// Unsubscribe without ownership → NotFound; row survives.
	err = push.RemoveSubscription(ctx, attacker.ID, endpoint)
	require.ErrorIs(t, err, notification.ErrSubscriptionNotFound)

	still, err := db.PushSubscription.Query().
		Where(pushsubscription.Endpoint(endpoint)).
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, ownerID, still.UserID)

	// Owner can still update keys and unsubscribe.
	require.NoError(t, push.SaveSubscription(ctx, ownerID, endpoint, "owner-p256-v2", "owner-auth-v2", "owner-ua"))
	updated, err := db.PushSubscription.Query().
		Where(pushsubscription.Endpoint(endpoint)).
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, ownerID, updated.UserID)
	assert.Equal(t, "owner-p256-v2", updated.P256dhKey)

	require.NoError(t, push.RemoveSubscription(ctx, ownerID, endpoint))
	count, err := db.PushSubscription.Query().
		Where(pushsubscription.Endpoint(endpoint)).
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)
}
