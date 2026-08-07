package notification_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/notification"
)

func testTelegramUnavailableApp(t *testing.T, authenticated bool) (*fiber.App, *notification.TelegramBindService, uuid.UUID) {
	t.Helper()

	dsn := "file:tg_unavailable_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("tg-unavailable-" + uuid.New().String()[:8] + "@test.com").
		SetName("Telegram unavailable user").
		Save(ctx)
	require.NoError(t, err)

	repo := notification.NewRepository(client)
	service := notification.NewNotificationService(repo, nil, nil, nil)
	// Empty token is intentional: this fixture must not initialize a bot or sender.
	telegramBind := notification.NewTelegramBindService(repo, nil, nil, "", "")
	require.NotNil(t, telegramBind)
	require.Nil(t, notification.ExportTelegramBindSender(telegramBind))

	handler := notification.NewHandler(service, telegramBind)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		if authenticated {
			c.Locals("userID", user.ID.String())
		}
		return c.Next()
	}
	handler.RegisterRoutes(api, authMw)

	return app, telegramBind, user.ID
}

// TestHandler_TelegramRoutesPresentAndAuthenticatedWithoutBotToken фиксирует
// стабильный HTTP surface: отсутствие bot token не превращает маршруты в 404.
func TestHandler_TelegramRoutesPresentAndAuthenticatedWithoutBotToken(t *testing.T) {
	app, _, _ := testTelegramUnavailableApp(t, false)

	tests := []struct {
		method string
		path   string
		body   []byte
	}{
		{method: http.MethodPost, path: "/api/notifications/telegram/bind"},
		{method: http.MethodPost, path: "/api/notifications/telegram/confirm", body: []byte(`{"code":"123456"}`)},
		{method: http.MethodDelete, path: "/api/notifications/telegram"},
		{method: http.MethodPost, path: "/api/notifications/telegram/test"},
		{method: http.MethodGet, path: "/api/notifications/telegram/status"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(tt.body))
			if len(tt.body) > 0 {
				req.Header.Set("Content-Type", "application/json")
			}
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
		})
	}
}

// TestHandler_TelegramAvailabilityWithoutBotToken проверяет разделение:
// status/unlink используют repository, остальные операции fail closed с 503.
func TestHandler_TelegramAvailabilityWithoutBotToken(t *testing.T) {
	app, telegramBind, userID := testTelegramUnavailableApp(t, true)
	ctx := context.Background()
	repo := notification.ExportTelegramBindRepository(telegramBind)
	require.NoError(t, repo.SetUserTelegram(ctx, userID, "42424242"))

	statusReq := httptest.NewRequest(http.MethodGet, "/api/notifications/telegram/status", nil)
	statusResp, err := app.Test(statusReq)
	require.NoError(t, err)
	defer statusResp.Body.Close()
	require.Equal(t, http.StatusOK, statusResp.StatusCode)
	var status notification.TelegramStatusResponse
	require.NoError(t, json.NewDecoder(statusResp.Body).Decode(&status))
	assert.True(t, status.Linked)
	assert.Equal(t, "42424242", status.TelegramID)

	unavailable := []struct {
		method string
		path   string
		body   []byte
	}{
		{method: http.MethodPost, path: "/api/notifications/telegram/bind"},
		{method: http.MethodPost, path: "/api/notifications/telegram/confirm", body: []byte(`{"code":"123456"}`)},
		{method: http.MethodPost, path: "/api/notifications/telegram/test"},
	}
	for _, tt := range unavailable {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(tt.body))
			if len(tt.body) > 0 {
				req.Header.Set("Content-Type", "application/json")
			}
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
		})
	}

	unlinkReq := httptest.NewRequest(http.MethodDelete, "/api/notifications/telegram", nil)
	unlinkResp, err := app.Test(unlinkReq)
	require.NoError(t, err)
	defer unlinkResp.Body.Close()
	require.Equal(t, fiber.StatusNoContent, unlinkResp.StatusCode)

	statusReq = httptest.NewRequest(http.MethodGet, "/api/notifications/telegram/status", nil)
	statusResp, err = app.Test(statusReq)
	require.NoError(t, err)
	defer statusResp.Body.Close()
	require.Equal(t, http.StatusOK, statusResp.StatusCode)
	status = notification.TelegramStatusResponse{}
	require.NoError(t, json.NewDecoder(statusResp.Body).Decode(&status))
	assert.False(t, status.Linked)
	assert.Empty(t, status.TelegramID)
}
