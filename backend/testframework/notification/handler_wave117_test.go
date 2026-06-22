package notification_test

// Тесты HTTP-обработчика уведомлений (волна 117): интеграционные и unit-сценарии testframework.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_List_IsReadFilterTrue проверяет HTTP-слой: HTTP-обработчик список is read filter true.
func TestHandler_List_IsReadFilterTrue(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications?is_read=false", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result notification.NotificationListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.GreaterOrEqual(t, result.Total, 1)
}

// TestHandler_List_InvalidPageDefaultsToOne проверяет отклонение при HTTP-обработчик список некорректный page defaults to one.
func TestHandler_List_InvalidPageDefaultsToOne(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications?page=0&per_page=0", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result notification.NotificationListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PerPage)
}

// TestHandler_PushUnsubscribe_EmptyBody проверяет HTTP-слой: HTTP-обработчик push unsubscribe пустой body.
func TestHandler_PushUnsubscribe_EmptyBody(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/notifications/push/unsubscribe", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
