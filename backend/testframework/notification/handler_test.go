package notification_test

// Тесты HTTP-обработчика уведомлений: интеграционные и unit-сценарии testframework.

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/config"
)

func testNotificationHandlerApp(t *testing.T) (*fiber.App, uuid.UUID) {
	t.Helper()
	svc, _, ctx, userID := newNotificationTestEnv(t)
	require.NoError(t, svc.Dispatch(ctx, notification.DispatchEvent{
		UserID: userID, Type: notification.TypeAssigned, Title: "Hello", EntityType: "issue", EntityID: uuid.New().String(),
	}))

	h := notification.NewHandler(svc, nil)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, userID
}

// TestHandler_List_Unauthorized проверяет отклонение при HTTP-обработчик список неавторизованный.
func TestHandler_List_Unauthorized(t *testing.T) {
	svc, _, _, _ := newNotificationTestEnv(t)
	h := notification.NewHandler(svc, nil)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestHandler_List_Success проверяет HTTP-слой: HTTP-обработчик список успех.
func TestHandler_List_Success(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/notifications?page=1&per_page=10", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result notification.NotificationListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.GreaterOrEqual(t, result.Total, 1)
}

// TestHandler_UnreadCount_Success проверяет HTTP-слой: HTTP-обработчик непрочитанные count успех.
func TestHandler_UnreadCount_Success(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/notifications/unread-count", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result notification.UnreadCountResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 1, result.Count)
}

// TestHandler_MarkAllRead_Success проверяет HTTP-слой: HTTP-обработчик mark all read успех.
func TestHandler_MarkAllRead_Success(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)
	req := httptest.NewRequest(http.MethodPatch, "/api/notifications/read-all", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/notifications/unread-count", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var result notification.UnreadCountResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 0, result.Count)
}

// TestHandler_UpdatePreference_Validation проверяет HTTP-слой: HTTP-обработчик обновление preference валидация.
func TestHandler_UpdatePreference_Validation(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)
	req := httptest.NewRequest(http.MethodPut, "/api/notifications/preferences",
		bytes.NewReader([]byte(`{"notification_type":"","channel":""}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_GetVAPIDKey_EmptyWhenPushDisabled проверяет HTTP-слой: HTTP-обработчик получение vapid key пустой when push disabled.
func TestHandler_GetVAPIDKey_EmptyWhenPushDisabled(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/notifications/push/vapid-key", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "vapid_public_key")
}

// TestHandler_PushSubscribe_ServiceUnavailable проверяет HTTP-слой: HTTP-обработчик push subscribe сервис unavailable.
func TestHandler_PushSubscribe_ServiceUnavailable(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)
	payload := `{"endpoint":"https://push.example/sub","p256dh_key":"k","auth_key":"a"}`
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/push/subscribe", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
}

// TestIsQuietHours_NoConfig проверяет сценарий: is quiet hours no config.
func TestIsQuietHours_NoConfig(t *testing.T) {
	svc, _, ctx, userID := newNotificationTestEnv(t)
	assert.False(t, notification.ExportIsQuietHours(svc, ctx, userID))
}

// TestCheckDeadlines_Empty проверяет сценарий: проверка deadlines пустой.
func TestCheckDeadlines_Empty(t *testing.T) {
	svc, _, ctx, _ := newNotificationTestEnv(t)
	require.NoError(t, svc.CheckDeadlines(ctx))
}

func notificationIDFromList(t *testing.T, app *fiber.App) uuid.UUID {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/notifications?page=1&per_page=1", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var result notification.NotificationListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.NotEmpty(t, result.Notifications)
	return result.Notifications[0].ID
}

// TestHandler_MarkRead_Success проверяет HTTP-слой: HTTP-обработчик mark read успех.
func TestHandler_MarkRead_Success(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)
	notifID := notificationIDFromList(t, app)

	req := httptest.NewRequest(http.MethodPatch, "/api/notifications/"+notifID.String()+"/read", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/notifications/unread-count", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var count notification.UnreadCountResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&count))
	assert.Equal(t, 0, count.Count)
}

// TestHandler_MarkRead_InvalidID проверяет отклонение при HTTP-обработчик mark read некорректный id.
func TestHandler_MarkRead_InvalidID(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/notifications/not-a-uuid/read", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_DeleteNotification_Success проверяет HTTP-слой: HTTP-обработчик удаление notification успех.
func TestHandler_DeleteNotification_Success(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)
	notifID := notificationIDFromList(t, app)

	req := httptest.NewRequest(http.MethodDelete, "/api/notifications/"+notifID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var result notification.NotificationListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 0, result.Total)
}

// TestHandler_GetPreferences_Success проверяет HTTP-слой: HTTP-обработчик получение preferences успех.
func TestHandler_GetPreferences_Success(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications/preferences", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var prefs []notification.PreferenceDTO
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&prefs))
	assert.NotNil(t, prefs)
}

// TestHandler_UpdatePreference_Success проверяет HTTP-слой: HTTP-обработчик обновление preference успех.
func TestHandler_UpdatePreference_Success(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	payload := `{"notification_type":"call_started","channel":"in_app","enabled":false}`
	req := httptest.NewRequest(http.MethodPut, "/api/notifications/preferences", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var pref notification.PreferenceDTO
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pref))
	assert.Equal(t, notification.TypeCallStarted, pref.NotificationType)
	assert.False(t, pref.Enabled)
}

// TestHandler_UpdatePreference_InvalidBody проверяет отклонение при HTTP-обработчик обновление preference некорректный body.
func TestHandler_UpdatePreference_InvalidBody(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodPut, "/api/notifications/preferences", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_PushSubscribe_ValidationMissingFields проверяет отклонение при HTTP-обработчик push subscribe валидация отсутствующий fields.
func TestHandler_PushSubscribe_ValidationMissingFields(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/notifications/push/subscribe",
		bytes.NewReader([]byte(`{"endpoint":"https://push.example/sub"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_PushSubscribe_InvalidBody проверяет отклонение при HTTP-обработчик push subscribe некорректный body.
func TestHandler_PushSubscribe_InvalidBody(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/notifications/push/subscribe", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_PushSubscribe_InvalidEndpoint проверяет отклонение при HTTP-обработчик push subscribe некорректный endpoint.
func TestHandler_PushSubscribe_InvalidEndpoint(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	priv, pub, err := generateTestVAPIDKeys(t)
	require.NoError(t, err)
	push := notification.NewPushSender(notification.ExportRepositoryDB(repo), config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)
	notification.ExportSetServicePush(svc, push)

	h := notification.NewHandler(svc, nil)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	_ = ctx

	payload := `{"endpoint":"javascript:alert(1)","p256dh_key":"k","auth_key":"a"}`
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/push/subscribe", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	// W7-A1-002: ValidateOutboundURL fail → 400 (не 500).
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_PushSubscribe_Success проверяет HTTP-слой: HTTP-обработчик push subscribe успех.
func TestHandler_PushSubscribe_Success(t *testing.T) {
	svc, repo, _, userID := newNotificationTestEnv(t)
	priv, pub, err := generateTestVAPIDKeys(t)
	require.NoError(t, err)
	push := notification.NewPushSender(notification.ExportRepositoryDB(repo), config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)
	notification.ExportSetServicePush(svc, push)

	h := notification.NewHandler(svc, nil)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	payload := `{"endpoint":"https://push.example.com/sub/abc","p256dh_key":"p256","auth_key":"auth"}`
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/push/subscribe", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

// TestHandler_PushUnsubscribe_ServiceUnavailable проверяет HTTP-слой: HTTP-обработчик push unsubscribe сервис unavailable.
func TestHandler_PushUnsubscribe_ServiceUnavailable(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)
	body := `{"endpoint":"https://push.example.com/sub/abc"}`
	req := httptest.NewRequest(http.MethodDelete, "/api/notifications/push/unsubscribe", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
}

// TestHandler_PushUnsubscribe_MissingEndpoint проверяет отклонение при HTTP-обработчик push unsubscribe отсутствующий endpoint.
func TestHandler_PushUnsubscribe_MissingEndpoint(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/notifications/push/unsubscribe", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_PushUnsubscribe_Success проверяет HTTP-слой: HTTP-обработчик push unsubscribe успех.
func TestHandler_PushUnsubscribe_Success(t *testing.T) {
	svc, repo, _, userID := newNotificationTestEnv(t)
	priv, pub, err := generateTestVAPIDKeys(t)
	require.NoError(t, err)
	push := notification.NewPushSender(notification.ExportRepositoryDB(repo), config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)
	notification.ExportSetServicePush(svc, push)

	endpoint := "https://push.example.com/sub/unsub"
	require.NoError(t, push.SaveSubscription(context.Background(), userID, endpoint, "p256", "auth", "ua"))

	h := notification.NewHandler(svc, nil)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	body := `{"endpoint":"` + endpoint + `"}`
	req := httptest.NewRequest(http.MethodDelete, "/api/notifications/push/unsubscribe", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	count, err := notification.ExportRepositoryDB(repo).PushSubscription.Query().Count(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
