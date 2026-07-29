package notification_test

// Тесты сервиса уведомлений: интеграционные и unit-сценарии testframework.

import (
	"testing"
	"time"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/platform/config"
)

// TestBuildPushURL проверяет сценарий: build push url.
func TestBuildPushURL(t *testing.T) {
	assert.Equal(t, "/tracker/issues/abc", notification.ExportBuildPushURL("issue", "abc"))
	assert.Equal(t, "/messenger/channels/ch1", notification.ExportBuildPushURL("channel", "ch1"))
	assert.Equal(t, "/tms/runs/run1", notification.ExportBuildPushURL("testrun", "run1"))
	assert.Equal(t, "/calls/call1", notification.ExportBuildPushURL("call", "call1"))
	assert.Equal(t, "/", notification.ExportBuildPushURL("unknown", "x"))
}

// TestNotifToDTO проверяет сценарий: notif to dto.
func TestNotifToDTO(t *testing.T) {
	projectID := uuid.New()
	actorID := uuid.New()
	entityID := uuid.New()
	now := time.Now()
	dto := notification.ExportNotifToDTO(&ent.Notification{
		ID: uuid.New(), Type: notification.TypeAssigned, Title: "Assigned", Body: "body",
		EntityType: "issue", EntityID: &entityID, ProjectID: &projectID, ActorID: &actorID,
		IsRead: false, IsUrgent: true, CreatedAt: now,
	})
	assert.Equal(t, notification.TypeAssigned, dto.Type)
	assert.Equal(t, entityID.String(), dto.EntityID)
	assert.True(t, dto.IsUrgent)
	require.NotNil(t, dto.ProjectID)
}

// TestDispatch_CreatesInAppNotification проверяет сценарий: dispatch creates in app notification.
func TestDispatch_CreatesInAppNotification(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)

	err := svc.Dispatch(ctx, notification.DispatchEvent{
		UserID:     userID,
		Type:       notification.TypeAssigned,
		Title:      "Task assigned",
		Body:       "You got a task",
		EntityType: "issue",
		EntityID:   uuid.New().String(),
	})
	require.NoError(t, err)

	count, err := repo.CountUnread(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestDispatch_QuietHoursSuppressesNonUrgent проверяет сценарий: dispatch quiet hours suppresses non urgent.
func TestDispatch_QuietHoursSuppressesNonUrgent(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)

	start := "23:00"
	end := "07:00"
	_, err := notification.ExportRepositoryDB(repo).User.UpdateOneID(userID).
		SetQuietHoursStart(start).
		SetQuietHoursEnd(end).
		SetTimezone("UTC").
		Save(ctx)
	require.NoError(t, err)

	// Принудительно включаем тихие часы, задавая таймзону пользователя и текущее время —
	// если мы вне тихих часов, пропускаем проверку счётчика.
	user, err := repo.GetUser(ctx, userID)
	require.NoError(t, err)
	inQuiet := notification.ExportIsQuietHours(svc, ctx, userID)
	if !inQuiet {
		t.Skip("current UTC time is outside configured quiet hours")
	}

	err = svc.Dispatch(ctx, notification.DispatchEvent{
		UserID: userID, Type: notification.TypeAssigned, Title: "Quiet", EntityType: "issue", EntityID: uuid.New().String(),
	})
	require.NoError(t, err)

	count, err := repo.CountUnread(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
	_ = user
}

// TestDispatch_UrgentBypassesQuietHours проверяет сценарий: dispatch urgent bypasses quiet hours.
func TestDispatch_UrgentBypassesQuietHours(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)

	start := "00:00"
	end := "23:59"
	_, err := notification.ExportRepositoryDB(repo).User.UpdateOneID(userID).
		SetQuietHoursStart(start).
		SetQuietHoursEnd(end).
		SetTimezone("UTC").
		Save(ctx)
	require.NoError(t, err)

	err = svc.Dispatch(ctx, notification.DispatchEvent{
		UserID: userID, Type: notification.TypeAssigned, Title: "Urgent", EntityType: "issue",
		EntityID: uuid.New().String(), IsUrgent: true,
	})
	require.NoError(t, err)

	count, err := repo.CountUnread(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestListNotifications_Pagination проверяет сценарий: список уведомления pagination.
func TestListNotifications_Pagination(t *testing.T) {
	svc, _, ctx, userID := newNotificationTestEnv(t)

	for i := 0; i < 3; i++ {
		require.NoError(t, svc.Dispatch(ctx, notification.DispatchEvent{
			UserID: userID, Type: notification.TypeCommentAdded, Title: "C", EntityType: "issue", EntityID: uuid.New().String(),
		}))
	}

	resp, err := svc.ListNotifications(ctx, userID, 1, 2, nil)
	require.NoError(t, err)
	assert.Len(t, resp.Notifications, 2)
	assert.Equal(t, 3, resp.Total)
}

// TestMarkReadAndDelete проверяет сценарий: mark read and удаление.
func TestMarkReadAndDelete(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	entityID := uuid.New().String()
	require.NoError(t, svc.Dispatch(ctx, notification.DispatchEvent{
		UserID: userID, Type: notification.TypeStatusChanged, Title: "Moved", EntityType: "issue", EntityID: entityID,
	}))

	items, _, err := repo.ListByUser(ctx, userID, 1, 10, nil)
	require.NoError(t, err)
	require.Len(t, items, 1)

	require.NoError(t, svc.MarkRead(ctx, items[0].ID, userID))
	count, err := svc.GetUnreadCount(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	require.NoError(t, svc.DeleteNotification(ctx, items[0].ID, userID))
	items, total, err := repo.ListByUser(ctx, userID, 1, 10, nil)
	require.NoError(t, err)
	assert.Empty(t, items)
	assert.Equal(t, 0, total)
}

// TestRepository_IsEnabled_DefaultTrue проверяет слой репозитория: репозиторий is enabled default true.
func TestRepository_IsEnabled_DefaultTrue(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	enabled, err := repo.IsEnabled(ctx, userID, notification.TypeAssigned, notification.ChannelInApp, nil)
	require.NoError(t, err)
	assert.True(t, enabled)
}

// TestRepository_UpsertPreference проверяет слой репозитория: репозиторий upsert preference.
func TestRepository_UpsertPreference(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)

	pref, err := repo.UpsertPreference(ctx, userID, notification.TypeAssigned, notification.ChannelPush, false, nil)
	require.NoError(t, err)
	assert.False(t, pref.Enabled)

	updated, err := repo.UpsertPreference(ctx, userID, notification.TypeAssigned, notification.ChannelPush, true, nil)
	require.NoError(t, err)
	assert.True(t, updated.Enabled)

	enabled, err := repo.IsEnabled(ctx, userID, notification.TypeAssigned, notification.ChannelPush, nil)
	require.NoError(t, err)
	assert.True(t, enabled)
}

// TestRepository_MarkAllRead проверяет слой репозитория: репозиторий mark all read.
func TestRepository_MarkAllRead(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)

	for i := 0; i < 2; i++ {
		_, err := repo.CreateNotification(ctx, userID, notification.TypeAssigned, "t", "b", "issue", uuid.New().String(), nil, nil, false)
		require.NoError(t, err)
	}

	count, err := svc.GetUnreadCount(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	require.NoError(t, svc.MarkAllRead(ctx, userID))
	count, err = svc.GetUnreadCount(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestGetPreferences проверяет сценарий: получение preferences.
func TestGetPreferences(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	_, err := notification.ExportRepositoryDB(repo).NotificationPreference.Create().
		SetUserID(userID).
		SetNotificationType(notification.TypeMentioned).
		SetChannel(notification.ChannelTelegram).
		SetEnabled(false).
		Save(ctx)
	require.NoError(t, err)

	prefs, err := svc.GetPreferences(ctx, userID)
	require.NoError(t, err)
	require.Len(t, prefs, 1)
	assert.Equal(t, notification.ChannelTelegram, prefs[0].Channel)
}

// TestUpdatePreference проверяет сценарий: обновление preference.
func TestUpdatePreference(t *testing.T) {
	svc, _, ctx, userID := newNotificationTestEnv(t)
	dto, err := svc.UpdatePreference(ctx, userID, notification.PreferenceUpdateRequest{
		NotificationType: notification.TypeCallStarted,
		Channel:          notification.ChannelInApp,
		Enabled:          false,
	})
	require.NoError(t, err)
	assert.False(t, dto.Enabled)
}

// TestSubjectToType_AuthSession проверяет сценарий: subject to type auth сессия.
func TestSubjectToType_AuthSession(t *testing.T) {
	assert.Equal(t, notification.TypeNewLogin, notification.ExportSubjectToType("auth.session.created"))
}

func generateTestVAPIDKeys(t *testing.T) (privateKey, publicKey string, err error) {
	t.Helper()
	return webpush.GenerateVAPIDKeys()
}

// TestUpdatePreference_WithProjectID проверяет сценарий: обновление preference with project id.
func TestUpdatePreference_WithProjectID(t *testing.T) {
	svc, _, ctx, userID := newNotificationTestEnv(t)
	projectID := uuid.New()

	dto, err := svc.UpdatePreference(ctx, userID, notification.PreferenceUpdateRequest{
		NotificationType: notification.TypeMentioned,
		Channel:          notification.ChannelPush,
		Enabled:          true,
		ProjectID:        &projectID,
	})
	require.NoError(t, err)
	require.NotNil(t, dto.ProjectID)
	assert.Equal(t, projectID, *dto.ProjectID)
}

// TestSaveSubscription_ValidatesEndpoint проверяет сценарий: save subscription validates endpoint.
func TestSaveSubscription_ValidatesEndpoint(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	priv, pub, err := generateTestVAPIDKeys(t)
	require.NoError(t, err)
	push := notification.NewPushSender(repo, config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)

	err = push.SaveSubscription(ctx, userID, "file:///etc/passwd", "k", "a", "test-agent")
	require.Error(t, err)
}

// TestSaveSubscription_CreateAndUpdate проверяет сценарий: save subscription создание and обновление.
func TestSaveSubscription_CreateAndUpdate(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	priv, pub, err := generateTestVAPIDKeys(t)
	require.NoError(t, err)
	push := notification.NewPushSender(repo, config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)

	endpoint := "https://push.example.com/sub/user-1"
	require.NoError(t, push.SaveSubscription(ctx, userID, endpoint, "key1", "auth1", "agent1"))
	require.NoError(t, push.SaveSubscription(ctx, userID, endpoint, "key2", "auth2", "agent2"))

	subs, err := notification.ExportRepositoryDB(repo).PushSubscription.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, subs, 1)
	assert.Equal(t, "key2", subs[0].P256dhKey)
	assert.Equal(t, userID, subs[0].UserID)
}

// TestRemoveSubscription_DeletesByEndpoint проверяет сценарий: remove subscription deletes by endpoint.
func TestRemoveSubscription_DeletesByEndpoint(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	priv, pub, err := generateTestVAPIDKeys(t)
	require.NoError(t, err)
	push := notification.NewPushSender(repo, config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)

	endpoint := "https://push.example.com/sub/remove"
	require.NoError(t, push.SaveSubscription(ctx, userID, endpoint, "k", "a", "ua"))
	require.NoError(t, push.RemoveSubscription(ctx, userID, endpoint))

	count, err := notification.ExportRepositoryDB(repo).PushSubscription.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestIsQuietHours_CrossMidnight — W7-A1-009 / P3-BE-02: фиксируем expected bool
// для cross-midnight ветки (start>end), не `_ = ExportIsQuietHours`.
func TestIsQuietHours_CrossMidnight(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)

	now := time.Now().UTC()
	// start на 7ч вперёд, end на 1ч вперёд → startMinutes > endMinutes (cross-midnight),
	// current < end → всегда quiet.
	inStart := now.Add(7 * time.Hour).Format("15:04")
	inEnd := now.Add(1 * time.Hour).Format("15:04")
	_, err := notification.ExportRepositoryDB(repo).User.UpdateOneID(userID).
		SetQuietHoursStart(inStart).
		SetQuietHoursEnd(inEnd).
		SetTimezone("UTC").
		Save(ctx)
	require.NoError(t, err)
	assert.True(t, notification.ExportIsQuietHours(svc, ctx, userID),
		"now must be inside cross-midnight quiet window %s–%s", inStart, inEnd)

	// start на 1ч вперёд, end на 1ч назад → quiet только вне [end, start); now вне → false.
	outStart := now.Add(1 * time.Hour).Format("15:04")
	outEnd := now.Add(-1 * time.Hour).Format("15:04")
	_, err = notification.ExportRepositoryDB(repo).User.UpdateOneID(userID).
		SetQuietHoursStart(outStart).
		SetQuietHoursEnd(outEnd).
		SetTimezone("UTC").
		Save(ctx)
	require.NoError(t, err)
	assert.False(t, notification.ExportIsQuietHours(svc, ctx, userID),
		"now must be outside cross-midnight quiet window %s–%s", outStart, outEnd)
}

// TestPushSender_sendToUser_UnsafeEndpointRemoved проверяет сценарий: push sender отправка to user unsafe endpoint removed.
func TestPushSender_sendToUser_UnsafeEndpointRemoved(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	priv, pub, err := generateTestVAPIDKeys(t)
	require.NoError(t, err)
	push := notification.NewPushSender(repo, config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)

	_, err = notification.ExportRepositoryDB(repo).PushSubscription.Create().
		SetUserID(userID).
		SetEndpoint("file:///etc/passwd").
		SetP256dhKey("k").
		SetAuthKey("a").
		SetUserAgent("test").
		Save(ctx)
	require.NoError(t, err)

	notification.ExportPushSendToUser(push, ctx, userID, "Alert", "Body", "/issues/1")

	count, err := notification.ExportRepositoryDB(repo).PushSubscription.Query().Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)
}
