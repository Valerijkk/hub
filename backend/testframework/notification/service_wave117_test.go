package notification_test

// Тесты сервиса уведомлений (волна 117): интеграционные и unit-сценарии testframework.

import (
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIsQuietHours_InvalidTimeFormat проверяет отклонение при is quiet hours некорректный time форматирование.
func TestIsQuietHours_InvalidTimeFormat(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	_, err := notification.ExportRepositoryDB(repo).User.UpdateOneID(userID).
		SetQuietHoursStart("invalid").
		SetQuietHoursEnd("06:00").
		SetTimezone("UTC").
		Save(ctx)
	require.NoError(t, err)

	assert.False(t, notification.ExportIsQuietHours(svc, ctx, userID))
}

// TestIsQuietHours_SameDayWindow проверяет сценарий: is quiet hours same day window.
func TestIsQuietHours_SameDayWindow(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	_, err := notification.ExportRepositoryDB(repo).User.UpdateOneID(userID).
		SetQuietHoursStart("00:00").
		SetQuietHoursEnd("23:59").
		SetTimezone("UTC").
		Save(ctx)
	require.NoError(t, err)

	assert.True(t, notification.ExportIsQuietHours(svc, ctx, userID))
}

// TestListNotifications_IsReadFilter проверяет сценарий: список уведомления is read filter.
func TestListNotifications_IsReadFilter(t *testing.T) {
	svc, _, ctx, userID := newNotificationTestEnv(t)

	read := true
	result, err := svc.ListNotifications(ctx, userID, 1, 20, &read)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Page)
}

// TestGetPreferences_ReturnsDTOs проверяет сценарий: получение preferences returns dt os.
func TestGetPreferences_ReturnsDTOs(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	projID := uuid.New()
	_, err := repo.UpsertPreference(ctx, userID, notification.TypeAssigned, notification.ChannelInApp, true, &projID)
	require.NoError(t, err)

	prefs, err := svc.GetPreferences(ctx, userID)
	require.NoError(t, err)
	require.NotEmpty(t, prefs)
}

// TestUpdatePreference_WithProjectScope проверяет сценарий: обновление preference with project scope.
func TestUpdatePreference_WithProjectScope(t *testing.T) {
	svc, _, ctx, userID := newNotificationTestEnv(t)
	projID := uuid.New()

	dto, err := svc.UpdatePreference(ctx, userID, notification.PreferenceUpdateRequest{
		NotificationType: notification.TypeStatusChanged,
		Channel:          notification.ChannelInApp,
		Enabled:          false,
		ProjectID:        &projID,
	})
	require.NoError(t, err)
	require.NotNil(t, dto.ProjectID)
	assert.Equal(t, projID, *dto.ProjectID)
	assert.False(t, dto.Enabled)
}

// TestDispatch_SuppressedDuringQuietHours проверяет сценарий: dispatch suppressed during quiet hours.
func TestDispatch_SuppressedDuringQuietHours(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)

	_, err := notification.ExportRepositoryDB(repo).User.UpdateOneID(userID).
		SetQuietHoursStart("00:00").
		SetQuietHoursEnd("23:59").
		SetTimezone("UTC").
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.Dispatch(ctx, notification.DispatchEvent{
		UserID:     userID,
		Type:       notification.TypeAssigned,
		Title:      "Normal during quiet hours",
		EntityType: "issue",
		EntityID:   uuid.New().String(),
	}))

	count, err := repo.CountUnread(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
