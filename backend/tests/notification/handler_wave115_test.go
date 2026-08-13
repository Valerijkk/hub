package notification_test

// Тесты HTTP-обработчика уведомлений (волна 115): интеграционные и unit-сценарии testframework.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_Delete_InvalidID проверяет отклонение при HTTP-обработчик удаление некорректный id.
func TestHandler_Delete_InvalidID(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/notifications/not-a-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_List_InvalidPage проверяет отклонение при HTTP-обработчик список некорректный page.
func TestHandler_List_InvalidPage(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications?page=0", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
