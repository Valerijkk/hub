package notification_test

// Тесты HTTP-обработчика уведомлений (волна 99): интеграционные и unit-сценарии testframework.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_MarkAllRead_Unauthorized проверяет отклонение при HTTP-обработчик mark all read неавторизованный.
func TestHandler_MarkAllRead_Unauthorized(t *testing.T) {
	svc, _, _, _ := newNotificationTestEnv(t)
	h := notification.NewHandler(svc, nil)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest(http.MethodPatch, "/api/notifications/read-all", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestHandler_Delete_Unauthorized проверяет отклонение при HTTP-обработчик удаление неавторизованный.
func TestHandler_Delete_Unauthorized(t *testing.T) {
	svc, _, _, _ := newNotificationTestEnv(t)
	h := notification.NewHandler(svc, nil)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest(http.MethodDelete, "/api/notifications/00000000-0000-0000-0000-000000000001", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
