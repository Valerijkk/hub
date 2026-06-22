package notification_test

// Тесты HTTP-обработчика уведомлений (волна 118): интеграционные и unit-сценарии testframework.

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

// TestHandler_UnreadCount_Unauthorized проверяет отклонение при HTTP-обработчик непрочитанные count неавторизованный.
func TestHandler_UnreadCount_Unauthorized(t *testing.T) {
	svc, _, _, _ := newNotificationTestEnv(t)
	h := notification.NewHandler(svc, nil)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest(http.MethodGet, "/api/notifications/unread-count", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestHandler_GetPreferences_Unauthorized проверяет отклонение при HTTP-обработчик получение preferences неавторизованный.
func TestHandler_GetPreferences_Unauthorized(t *testing.T) {
	svc, _, _, _ := newNotificationTestEnv(t)
	h := notification.NewHandler(svc, nil)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest(http.MethodGet, "/api/notifications/preferences", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestGetUserID_UUIDLocal проверяет сценарий: получение user id uuid local.
func TestGetUserID_UUIDLocal(t *testing.T) {
	app := fiber.New()
	t.Cleanup(func() { _ = app.Shutdown() })

	id := uuid.New()
	var got uuid.UUID
	var gotErr error

	app.Get("/uid", func(c fiber.Ctx) error {
		c.Locals("userID", id)
		got, gotErr = notification.ExportGetUserID(c)
		return c.SendStatus(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/uid", nil)
	_, err := app.Test(req)
	require.NoError(t, err)
	require.NoError(t, gotErr)
	assert.Equal(t, id, got)
}
