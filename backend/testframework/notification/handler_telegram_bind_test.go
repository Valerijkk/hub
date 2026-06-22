package notification_test

// Тесты HTTP-обработчика уведомлений: интеграционные и unit-сценарии testframework.

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/enttest"
)

func testTelegramBindHandlerApp(t *testing.T) (*fiber.App, uuid.UUID, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	dsn := "file:tghandler_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("tg-notification.Handler-" + uuid.New().String()[:8] + "@test.com").
		SetName("TG User").
		Save(ctx)
	require.NoError(t, err)

	repo := notification.NewRepository(client)
	svc := notification.NewNotificationService(repo, nil, nil, nil)
	tgBind := notification.NewTelegramBindService(
		repo, rdb, nil,
		"123456789:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc",
		"HubNotifyBot",
	)
	require.NotNil(t, tgBind)

	h := notification.NewHandler(svc, tgBind)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", user.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, user.ID, mr
}

// TestHandler_TelegramBind_Success проверяет HTTP-слой: HTTP-обработчик Telegram привязка успех.
func TestHandler_TelegramBind_Success(t *testing.T) {
	app, _, _ := testTelegramBindHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/notifications/telegram/bind", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result notification.TelegramBindResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result.BindToken)
	assert.Contains(t, result.DeepLink, "bind_")
}

// TestHandler_TelegramStatus_NotLinked проверяет HTTP-слой: HTTP-обработчик Telegram status not linked.
func TestHandler_TelegramStatus_NotLinked(t *testing.T) {
	app, _, _ := testTelegramBindHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/notifications/telegram/status", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result notification.TelegramStatusResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.False(t, result.Linked)
}

// TestHandler_TelegramConfirm_InvalidCode проверяет отклонение при HTTP-обработчик Telegram confirm некорректный code.
func TestHandler_TelegramConfirm_InvalidCode(t *testing.T) {
	app, _, _ := testTelegramBindHandlerApp(t)

	body, err := json.Marshal(notification.TelegramConfirmRequest{Code: "000000"})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/telegram/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_TelegramConfirm_Success проверяет HTTP-слой: HTTP-обработчик Telegram confirm успех.
func TestHandler_TelegramConfirm_Success(t *testing.T) {
	app, _, mr := testTelegramBindHandlerApp(t)
	require.NoError(t, mr.Set("tg:linkcode:112233", "44556677"))

	body, err := json.Marshal(notification.TelegramConfirmRequest{Code: "112233"})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/telegram/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_TelegramUnlink_NoContent проверяет HTTP-слой: HTTP-обработчик Telegram unlink no content.
func TestHandler_TelegramUnlink_NoContent(t *testing.T) {
	app, _, _ := testTelegramBindHandlerApp(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/notifications/telegram", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

// TestHandler_TelegramTest_NotLinkedOrUnavailable проверяет HTTP-слой: HTTP-обработчик Telegram test not linked or unavailable.
func TestHandler_TelegramTest_NotLinkedOrUnavailable(t *testing.T) {
	app, _, _ := testTelegramBindHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/notifications/telegram/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	// sender=nil в тестовом окружении → 503/500; без привязки → 400
	assert.True(t, resp.StatusCode == fiber.StatusBadRequest ||
		resp.StatusCode == fiber.StatusInternalServerError ||
		resp.StatusCode == fiber.StatusServiceUnavailable)
}
