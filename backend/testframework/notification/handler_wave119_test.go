package notification_test

// Тесты HTTP-обработчика уведомлений (волна 119): интеграционные и unit-сценарии testframework.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_MarkRead_Unauthorized проверяет отклонение при HTTP-обработчик mark read неавторизованный.
func TestHandler_MarkRead_Unauthorized(t *testing.T) {
	svc, _, _, _ := newNotificationTestEnv(t)
	h := notification.NewHandler(svc, nil)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest(http.MethodPatch, "/api/notifications/"+uuid.New().String()+"/read", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestHandler_List_InvalidIsReadFilterIgnored проверяет отклонение при HTTP-обработчик список некорректный is read filter ignored.
func TestHandler_List_InvalidIsReadFilterIgnored(t *testing.T) {
	app, _ := testNotificationHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications?is_read=maybe", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
