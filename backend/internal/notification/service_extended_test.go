package notification

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/platform/config"
)

func newNotificationTestEnv(t *testing.T) (*NotificationService, *Repository, context.Context, uuid.UUID) {
	t.Helper()
	dsn := "file:notiftest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("notif-" + uuid.New().String()[:8] + "@test.com").
		SetName("Recipient").
		Save(ctx)
	require.NoError(t, err)

	repo := NewRepository(client)
	svc := NewNotificationService(repo, nil, nil, nil)
	return svc, repo, ctx, user.ID
}

func TestBuildPushURL(t *testing.T) {
	assert.Equal(t, "/tracker/issues/abc", buildPushURL("issue", "abc"))
	assert.Equal(t, "/messenger/channels/ch1", buildPushURL("channel", "ch1"))
	assert.Equal(t, "/tms/runs/run1", buildPushURL("testrun", "run1"))
	assert.Equal(t, "/calls/call1", buildPushURL("call", "call1"))
	assert.Equal(t, "/", buildPushURL("unknown", "x"))
}

func TestNotifToDTO(t *testing.T) {
	projectID := uuid.New()
	actorID := uuid.New()
	now := time.Now()
	dto := notifToDTO(&ent.Notification{
		ID: uuid.New(), Type: TypeAssigned, Title: "Assigned", Body: "body",
		EntityType: "issue", EntityID: "iss-1", ProjectID: &projectID, ActorID: &actorID,
		IsRead: false, IsUrgent: true, CreatedAt: now,
	})
	assert.Equal(t, TypeAssigned, dto.Type)
	assert.True(t, dto.IsUrgent)
	require.NotNil(t, dto.ProjectID)
}

func TestDispatch_CreatesInAppNotification(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)

	err := svc.Dispatch(ctx, DispatchEvent{
		UserID:     userID,
		Type:       TypeAssigned,
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

func TestDispatch_QuietHoursSuppressesNonUrgent(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)

	start := "23:00"
	end := "07:00"
	_, err := repo.db.User.UpdateOneID(userID).
		SetQuietHoursStart(start).
		SetQuietHoursEnd(end).
		SetTimezone("UTC").
		Save(ctx)
	require.NoError(t, err)

	// Force quiet hours by setting user timezone and using current time —
	// if we're outside quiet hours, skip assertion on count.
	user, err := repo.GetUser(ctx, userID)
	require.NoError(t, err)
	inQuiet := svc.isQuietHours(ctx, userID)
	if !inQuiet {
		t.Skip("current UTC time is outside configured quiet hours")
	}

	err = svc.Dispatch(ctx, DispatchEvent{
		UserID: userID, Type: TypeAssigned, Title: "Quiet", EntityType: "issue", EntityID: uuid.New().String(),
	})
	require.NoError(t, err)

	count, err := repo.CountUnread(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
	_ = user
}

func TestDispatch_UrgentBypassesQuietHours(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)

	start := "00:00"
	end := "23:59"
	_, err := repo.db.User.UpdateOneID(userID).
		SetQuietHoursStart(start).
		SetQuietHoursEnd(end).
		SetTimezone("UTC").
		Save(ctx)
	require.NoError(t, err)

	err = svc.Dispatch(ctx, DispatchEvent{
		UserID: userID, Type: TypeAssigned, Title: "Urgent", EntityType: "issue",
		EntityID: uuid.New().String(), IsUrgent: true,
	})
	require.NoError(t, err)

	count, err := repo.CountUnread(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestListNotifications_Pagination(t *testing.T) {
	svc, _, ctx, userID := newNotificationTestEnv(t)

	for i := 0; i < 3; i++ {
		require.NoError(t, svc.Dispatch(ctx, DispatchEvent{
			UserID: userID, Type: TypeCommentAdded, Title: "C", EntityType: "issue", EntityID: uuid.New().String(),
		}))
	}

	resp, err := svc.ListNotifications(ctx, userID, 1, 2, nil)
	require.NoError(t, err)
	assert.Len(t, resp.Notifications, 2)
	assert.Equal(t, 3, resp.Total)
}

func TestMarkReadAndDelete(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	entityID := uuid.New().String()
	require.NoError(t, svc.Dispatch(ctx, DispatchEvent{
		UserID: userID, Type: TypeStatusChanged, Title: "Moved", EntityType: "issue", EntityID: entityID,
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

func TestRepository_IsEnabled_DefaultTrue(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	enabled, err := repo.IsEnabled(ctx, userID, TypeAssigned, ChannelInApp, nil)
	require.NoError(t, err)
	assert.True(t, enabled)
}

func TestRepository_UpsertPreference(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)

	pref, err := repo.UpsertPreference(ctx, userID, TypeAssigned, ChannelPush, false, nil)
	require.NoError(t, err)
	assert.False(t, pref.Enabled)

	updated, err := repo.UpsertPreference(ctx, userID, TypeAssigned, ChannelPush, true, nil)
	require.NoError(t, err)
	assert.True(t, updated.Enabled)

	enabled, err := repo.IsEnabled(ctx, userID, TypeAssigned, ChannelPush, nil)
	require.NoError(t, err)
	assert.True(t, enabled)
}

func TestRepository_MarkAllRead(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)

	for i := 0; i < 2; i++ {
		_, err := repo.CreateNotification(ctx, userID, TypeAssigned, "t", "b", "issue", uuid.New().String(), nil, nil, false)
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

func TestGetPreferences(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	_, err := repo.db.NotificationPreference.Create().
		SetUserID(userID).
		SetNotificationType(TypeMentioned).
		SetChannel(ChannelTelegram).
		SetEnabled(false).
		Save(ctx)
	require.NoError(t, err)

	prefs, err := svc.GetPreferences(ctx, userID)
	require.NoError(t, err)
	require.Len(t, prefs, 1)
	assert.Equal(t, ChannelTelegram, prefs[0].Channel)
}

func TestUpdatePreference(t *testing.T) {
	svc, _, ctx, userID := newNotificationTestEnv(t)
	dto, err := svc.UpdatePreference(ctx, userID, PreferenceUpdateRequest{
		NotificationType: TypeCallStarted,
		Channel:          ChannelInApp,
		Enabled:          false,
	})
	require.NoError(t, err)
	assert.False(t, dto.Enabled)
}

func TestSubjectToType_AuthSession(t *testing.T) {
	assert.Equal(t, TypeNewLogin, SubjectToType("auth.session.created"))
}

func generateTestVAPIDKeys(t *testing.T) (privateKey, publicKey string, err error) {
	t.Helper()
	return webpush.GenerateVAPIDKeys()
}

func TestUpdatePreference_WithProjectID(t *testing.T) {
	svc, _, ctx, userID := newNotificationTestEnv(t)
	projectID := uuid.New()

	dto, err := svc.UpdatePreference(ctx, userID, PreferenceUpdateRequest{
		NotificationType: TypeMentioned,
		Channel:          ChannelPush,
		Enabled:          true,
		ProjectID:        &projectID,
	})
	require.NoError(t, err)
	require.NotNil(t, dto.ProjectID)
	assert.Equal(t, projectID, *dto.ProjectID)
}

func TestSaveSubscription_ValidatesEndpoint(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	priv, pub, err := generateTestVAPIDKeys(t)
	require.NoError(t, err)
	push := NewPushSender(repo.db, config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)

	err = push.SaveSubscription(ctx, userID, "file:///etc/passwd", "k", "a", "test-agent")
	require.Error(t, err)
}

func TestSaveSubscription_CreateAndUpdate(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	priv, pub, err := generateTestVAPIDKeys(t)
	require.NoError(t, err)
	push := NewPushSender(repo.db, config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)

	endpoint := "https://push.example.com/sub/user-1"
	require.NoError(t, push.SaveSubscription(ctx, userID, endpoint, "key1", "auth1", "agent1"))
	require.NoError(t, push.SaveSubscription(ctx, userID, endpoint, "key2", "auth2", "agent2"))

	subs, err := repo.db.PushSubscription.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, subs, 1)
	assert.Equal(t, "key2", subs[0].P256dhKey)
	assert.Equal(t, userID, subs[0].UserID)
}

func TestRemoveSubscription_DeletesByEndpoint(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	priv, pub, err := generateTestVAPIDKeys(t)
	require.NoError(t, err)
	push := NewPushSender(repo.db, config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)

	endpoint := "https://push.example.com/sub/remove"
	require.NoError(t, push.SaveSubscription(ctx, userID, endpoint, "k", "a", "ua"))
	require.NoError(t, push.RemoveSubscription(ctx, endpoint))

	count, err := repo.db.PushSubscription.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
