package notification_test

// Тесты сервиса уведомлений (волна 111): интеграционные и unit-сценарии testframework.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/config"
)

// TestDispatch_TelegramAndPushChannels проверяет сценарий: dispatch Telegram and push каналы.
func TestDispatch_TelegramAndPushChannels(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)

	tgID := "123456789"
	_, err := notification.ExportRepositoryDB(repo).User.UpdateOneID(userID).SetTelegramID(tgID).Save(ctx)
	require.NoError(t, err)

	telegram := notification.NewTelegramSender("123456789:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc")
	require.NotNil(t, telegram)

	priv, pub, err := webpush.GenerateVAPIDKeys()
	require.NoError(t, err)
	push := notification.NewPushSender(notification.ExportRepositoryDB(repo), config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)
	require.NoError(t, push.SaveSubscription(ctx, userID, "https://push.example.com/sub/w111", "k", "a", "test"))

	svc = notification.NewNotificationService(repo, nil, telegram, push)

	require.NoError(t, svc.Dispatch(ctx, notification.DispatchEvent{
		UserID:     userID,
		Type:       notification.TypeAssigned,
		Title:      "Multi-channel",
		Body:       "hello",
		EntityType: "issue",
		EntityID:   uuid.New().String(),
	}))

	count, err := repo.CountUnread(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestDispatch_InAppDisabledStillPersistsWhenEnabledViaDefault проверяет сценарий: dispatch in app disabled still persists when enabled via default.
func TestDispatch_InAppDisabledStillPersistsWhenEnabledViaDefault(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)

	_, err := repo.UpsertPreference(ctx, userID, notification.TypeMentioned, notification.ChannelInApp, false, nil)
	require.NoError(t, err)

	require.NoError(t, svc.Dispatch(ctx, notification.DispatchEvent{
		UserID:     userID,
		Type:       notification.TypeMentioned,
		Title:      "Mention",
		EntityType: "issue",
		EntityID:   uuid.New().String(),
	}))

	count, err := repo.CountUnread(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestDispatch_WithProjectAndActor проверяет сценарий: dispatch with project and actor.
func TestDispatch_WithProjectAndActor(t *testing.T) {
	nc := startWave79NATSServer(t)
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	svc = notification.NewNotificationService(repo, nc, nil, nil)

	actor, err := notification.ExportRepositoryDB(repo).User.Create().
		SetEmail("actor-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := notification.ExportRepositoryDB(repo).Project.Create().
		SetName("Notify Proj").
		SetKey(testNotifyProjectKey("N")).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.Dispatch(ctx, notification.DispatchEvent{
		UserID:     userID,
		Type:       notification.TypeStatusChanged,
		Title:      "Moved",
		Body:       "status",
		EntityType: "issue",
		EntityID:   uuid.New().String(),
		ProjectID:  &proj.ID,
		ActorID:    &actor.ID,
	}))

	notifs, err := notification.ExportRepositoryDB(repo).Notification.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	require.NotNil(t, notifs[0].ProjectID)
	require.NotNil(t, notifs[0].ActorID)
}

// TestHandler_PushSubscribe_NoPushSender проверяет HTTP-слой: HTTP-обработчик push subscribe no push sender.
func TestHandler_PushSubscribe_NoPushSender(t *testing.T) {
	svc, _, _, _ := newNotificationTestEnv(t)
	h := notification.NewHandler(svc, nil)
	app := fiber.New()
	api := app.Group("/api")
	userID := uuid.New()
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	body := []byte(`{"endpoint":"https://push.example.com/sub/x","p256dh_key":"k","auth_key":"a"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/push/subscribe", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
}

// TestHandler_TelegramConfirm_InvalidJSON проверяет отклонение при HTTP-обработчик Telegram confirm некорректный json.
func TestHandler_TelegramConfirm_InvalidJSON(t *testing.T) {
	app, _, _ := testTelegramBindHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/notifications/telegram/confirm", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_TelegramStatus_LinkedAfterConfirm проверяет HTTP-слой: HTTP-обработчик Telegram status linked after confirm.
func TestHandler_TelegramStatus_LinkedAfterConfirm(t *testing.T) {
	app, _, mr := testTelegramBindHandlerApp(t)
	require.NoError(t, mr.Set("tg:linkcode:998877", "11223344"))

	body, err := json.Marshal(notification.TelegramConfirmRequest{Code: "998877"})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/telegram/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, "/api/notifications/telegram/status", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var status notification.TelegramStatusResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&status))
	assert.True(t, status.Linked)
}

// TestPublishToWS_NilNatsNoPanic гарантирует отсутствие паники/гонок при publish to ws nil nats no паника.
func TestPublishToWS_NilNatsNoPanic(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	notif, err := repo.CreateNotification(ctx, userID, notification.TypeAssigned, "t", "b", "issue", uuid.New().String(), nil, nil, false)
	require.NoError(t, err)
	require.NotNil(t, notif)
	notification.ExportPublishToWS(svc, userID, notif)
}
