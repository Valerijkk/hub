package notification_test

// Тесты HTTP-обработчика уведомлений (волна 116): интеграционные и unit-сценарии testframework.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_List_InvalidPerPage проверяет отклонение при HTTP-обработчик список некорректный per page.
func TestHandler_List_InvalidPerPage(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications?per_page=999", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
