package notification_test

// Тесты сервиса уведомлений (волна 119): интеграционные и unit-сценарии testframework.

import (
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMarkRead_NonExistentNotificationNoError проверяет сценарий: mark read non existent notification no ошибка.
func TestMarkRead_NonExistentNotificationNoError(t *testing.T) {
	svc, _, ctx, userID := newNotificationTestEnv(t)

	err := svc.MarkRead(ctx, uuid.New(), userID)
	require.NoError(t, err)
}

// TestDeleteNotification_NonExistentNoError проверяет сценарий: удаление notification non existent no ошибка.
func TestDeleteNotification_NonExistentNoError(t *testing.T) {
	svc, _, ctx, userID := newNotificationTestEnv(t)

	err := svc.DeleteNotification(ctx, uuid.New(), userID)
	require.NoError(t, err)
}

// TestListNotifications_ReadUnreadSplit проверяет сценарий: список уведомления read непрочитанные split.
func TestListNotifications_ReadUnreadSplit(t *testing.T) {
	svc, _, ctx, userID := newNotificationTestEnv(t)

	require.NoError(t, svc.Dispatch(ctx, notification.DispatchEvent{
		UserID: userID, Type: notification.TypeAssigned, Title: "Unread", EntityType: "issue", EntityID: uuid.New().String(),
	}))

	read := true
	result, err := svc.ListNotifications(ctx, userID, 1, 20, &read)
	require.NoError(t, err)
	assert.Empty(t, result.Notifications)

	unread := false
	result, err = svc.ListNotifications(ctx, userID, 1, 20, &unread)
	require.NoError(t, err)
	assert.Len(t, result.Notifications, 1)
}

// TestDispatch_UrgentFlagPersisted проверяет сценарий: dispatch urgent flag persisted.
func TestDispatch_UrgentFlagPersisted(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)

	require.NoError(t, svc.Dispatch(ctx, notification.DispatchEvent{
		UserID: userID, Type: notification.TypeBlockerAdded, Title: "Blocked", EntityType: "issue",
		EntityID: uuid.New().String(), IsUrgent: true,
	}))

	n, err := notification.ExportRepositoryDB(repo).Notification.Query().Only(ctx)
	require.NoError(t, err)
	assert.True(t, n.IsUrgent)
}
