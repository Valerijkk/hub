package notification_test

// Тесты HTTP-обработчика уведомлений: интеграционные и unit-сценарии testframework.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_List_FilterByIsRead проверяет HTTP-слой: HTTP-обработчик список filter by is read.
func TestHandler_List_FilterByIsRead(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications?is_read=false&page=1&per_page=10", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var unreadList notification.NotificationListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&unreadList))
	assert.GreaterOrEqual(t, unreadList.Total, 1)

	notifID := unreadList.Notifications[0].ID
	req = httptest.NewRequest(http.MethodPatch, "/api/notifications/"+notifID.String()+"/read", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/notifications?is_read=true&page=1&per_page=10", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var readList notification.NotificationListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&readList))
	assert.GreaterOrEqual(t, readList.Total, 1)
	assert.True(t, readList.Notifications[0].IsRead)
}

// TestHandler_List_InvalidIsReadParam_Ignored проверяет отклонение при HTTP-обработчик список некорректный is read param ignored.
func TestHandler_List_InvalidIsReadParam_Ignored(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications?is_read=maybe", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result notification.NotificationListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.GreaterOrEqual(t, result.Total, 1)
}
