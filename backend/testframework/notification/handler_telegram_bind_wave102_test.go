package notification_test

// Тесты HTTP-обработчика уведомлений (волна 102): интеграционные и unit-сценарии testframework.

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

// TestHandler_TelegramBind_Unauthorized проверяет отклонение при HTTP-обработчик Telegram привязка неавторизованный.
func TestHandler_TelegramBind_Unauthorized(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	dsn := "file:tghunauth_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	repo := notification.NewRepository(client)
	tgBind := notification.NewTelegramBindService(
		repo, rdb, nil,
		"123456789:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc",
		"HubNotifyBot",
	)
	require.NotNil(t, tgBind)

	svc := notification.NewNotificationService(repo, nil, nil, nil)
	h := notification.NewHandler(svc, tgBind)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest(http.MethodPost, "/api/notifications/telegram/bind", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestHandler_TelegramTest_LinkedSuccess проверяет HTTP-слой: HTTP-обработчик Telegram test linked успех.
func TestHandler_TelegramTest_LinkedSuccess(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	dsn := "file:tghtest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("tgtest-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	repo := notification.NewRepository(client)
	tgID := "42424242"
	require.NoError(t, repo.SetUserTelegram(ctx, user.ID, tgID))

	sender := notification.ExportNewTestTelegramSender(1)
	tgBind := notification.NewTelegramBindService(
		repo, rdb, sender,
		"123456789:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc",
		"HubNotifyBot",
	)
	require.NotNil(t, tgBind)

	svc := notification.NewNotificationService(repo, nil, nil, nil)
	h := notification.NewHandler(svc, tgBind)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", user.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	req := httptest.NewRequest(http.MethodPost, "/api/notifications/telegram/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "test sent", result["message"])
}

// TestHandler_TelegramConfirm_InvalidBody проверяет отклонение при HTTP-обработчик Telegram confirm некорректный body.
func TestHandler_TelegramConfirm_InvalidBody(t *testing.T) {
	app, _, _ := testTelegramBindHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/notifications/telegram/confirm", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
