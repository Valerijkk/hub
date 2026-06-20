package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/internal/platform/config"
)

func testNotificationHandlerApp(t *testing.T) (*fiber.App, uuid.UUID) {
	t.Helper()
	svc, _, ctx, userID := newNotificationTestEnv(t)
	require.NoError(t, svc.Dispatch(ctx, DispatchEvent{
		UserID: userID, Type: TypeAssigned, Title: "Hello", EntityType: "issue", EntityID: uuid.New().String(),
	}))

	h := NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, userID
}

func TestHandler_List_Unauthorized(t *testing.T) {
	svc, _, _, _ := newNotificationTestEnv(t)
	h := NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_List_Success(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/notifications?page=1&per_page=10", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result NotificationListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.GreaterOrEqual(t, result.Total, 1)
}

func TestHandler_UnreadCount_Success(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/notifications/unread-count", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result UnreadCountResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 1, result.Count)
}

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

	var result UnreadCountResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 0, result.Count)
}

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

func TestIsQuietHours_NoConfig(t *testing.T) {
	svc, _, ctx, userID := newNotificationTestEnv(t)
	assert.False(t, svc.isQuietHours(ctx, userID))
}

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

	var result NotificationListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.NotEmpty(t, result.Notifications)
	return result.Notifications[0].ID
}

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

	var count UnreadCountResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&count))
	assert.Equal(t, 0, count.Count)
}

func TestHandler_MarkRead_InvalidID(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/notifications/not-a-uuid/read", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

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

	var result NotificationListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 0, result.Total)
}

func TestHandler_GetPreferences_Success(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications/preferences", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var prefs []PreferenceDTO
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&prefs))
	assert.NotNil(t, prefs)
}

func TestHandler_UpdatePreference_Success(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	payload := `{"notification_type":"call_started","channel":"in_app","enabled":false}`
	req := httptest.NewRequest(http.MethodPut, "/api/notifications/preferences", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var pref PreferenceDTO
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pref))
	assert.Equal(t, TypeCallStarted, pref.NotificationType)
	assert.False(t, pref.Enabled)
}

func TestHandler_UpdatePreference_InvalidBody(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodPut, "/api/notifications/preferences", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

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

func TestHandler_PushSubscribe_InvalidBody(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/notifications/push/subscribe", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_PushSubscribe_InvalidEndpoint(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	priv, pub, err := generateTestVAPIDKeys(t)
	require.NoError(t, err)
	push := NewPushSender(repo.db, config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)
	svc.push = push

	h := NewHandler(svc)
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
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestHandler_PushSubscribe_Success(t *testing.T) {
	svc, repo, _, userID := newNotificationTestEnv(t)
	priv, pub, err := generateTestVAPIDKeys(t)
	require.NoError(t, err)
	push := NewPushSender(repo.db, config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)
	svc.push = push

	h := NewHandler(svc)
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

func TestHandler_PushUnsubscribe_MissingEndpoint(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/notifications/push/unsubscribe", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_PushUnsubscribe_Success(t *testing.T) {
	svc, repo, _, userID := newNotificationTestEnv(t)
	priv, pub, err := generateTestVAPIDKeys(t)
	require.NoError(t, err)
	push := NewPushSender(repo.db, config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)
	svc.push = push

	endpoint := "https://push.example.com/sub/unsub"
	require.NoError(t, push.SaveSubscription(context.Background(), userID, endpoint, "p256", "auth", "ua"))

	h := NewHandler(svc)
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

	count, err := repo.db.PushSubscription.Query().Count(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
