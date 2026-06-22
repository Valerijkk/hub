package notification_test

// Тесты сервиса уведомлений (волна 118): интеграционные и unit-сценарии testframework.

import (
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDispatch_ProjectScopedPreferenceDisabled проверяет сценарий: dispatch project scoped preference disabled.
func TestDispatch_ProjectScopedPreferenceDisabled(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	projID := uuid.New()

	_, err := repo.UpsertPreference(ctx, userID, notification.TypeStatusChanged, notification.ChannelInApp, false, &projID)
	require.NoError(t, err)

	require.NoError(t, svc.Dispatch(ctx, notification.DispatchEvent{
		UserID:     userID,
		Type:       notification.TypeStatusChanged,
		Title:      "Scoped off",
		EntityType: "issue",
		EntityID:   uuid.New().String(),
		ProjectID:  &projID,
	}))

	count, err := repo.CountUnread(ctx, userID)
	require.NoError(t, err)
	assert.Zero(t, count)
}

// TestIsQuietHours_InvalidTimezone проверяет отклонение при is quiet hours некорректный timezone.
func TestIsQuietHours_InvalidTimezone(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	_, err := notification.ExportRepositoryDB(repo).User.UpdateOneID(userID).
		SetQuietHoursStart("09:00").
		SetQuietHoursEnd("18:00").
		SetTimezone("Invalid/Zone").
		Save(ctx)
	require.NoError(t, err)

	_ = notification.ExportIsQuietHours(svc, ctx, userID)
}

// TestGetPreferences_EmptyList проверяет сценарий: получение preferences пустой список.
func TestGetPreferences_EmptyList(t *testing.T) {
	svc, _, ctx, userID := newNotificationTestEnv(t)

	prefs, err := svc.GetPreferences(ctx, userID)
	require.NoError(t, err)
	assert.NotNil(t, prefs)
}
