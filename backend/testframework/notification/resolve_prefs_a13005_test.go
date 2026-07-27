//go:build hubtest

package notification_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/notification"
)

// TestResolveChannelPrefs_SingleQuery_A13_005 — один ResolveChannelPrefs покрывает
// 3 канала (раньше 3× IsEnabled × до 2 SQL). Project override побеждает global.
func TestResolveChannelPrefs_SingleQuery_A13_005(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	client := notification.ExportRepositoryDB(repo)

	projectID := uuid.New()
	_, err := client.NotificationPreference.Create().
		SetUserID(userID).
		SetNotificationType(notification.TypeAssigned).
		SetChannel(notification.ChannelInApp).
		SetEnabled(false).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.NotificationPreference.Create().
		SetUserID(userID).
		SetNotificationType(notification.TypeAssigned).
		SetChannel(notification.ChannelTelegram).
		SetEnabled(true).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.NotificationPreference.Create().
		SetUserID(userID).
		SetNotificationType(notification.TypeAssigned).
		SetChannel(notification.ChannelPush).
		SetEnabled(false).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	prefs, err := repo.ResolveChannelPrefs(ctx, userID, notification.TypeAssigned, &projectID)
	require.NoError(t, err)

	assert.False(t, prefs[notification.ChannelInApp], "global in_app=false")
	assert.True(t, prefs[notification.ChannelTelegram], "project telegram=true")
	assert.False(t, prefs[notification.ChannelPush], "project push=false")

	// Before: Dispatch → 3× IsEnabled (до 6 SQL) + 1–2 GetUser.
	// After: 1 ResolveChannelPrefs + 1 GetUser.
	inApp, err := repo.IsEnabled(ctx, userID, notification.TypeAssigned, notification.ChannelInApp, &projectID)
	require.NoError(t, err)
	assert.False(t, inApp)

	require.NoError(t, notification.NewNotificationService(repo, nil, nil, nil).Dispatch(ctx, notification.DispatchEvent{
		UserID:     userID,
		Type:       notification.TypeAssigned,
		Title:      "A13-005",
		EntityType: "issue",
		EntityID:   "iss-a13005",
		ProjectID:  &projectID,
	}))
}
