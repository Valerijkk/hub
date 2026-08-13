package messenger_test

// Тесты HTTP-обработчика мессенджера (волна 100): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_SetCustomStatus проверяет HTTP-слой: HTTP-обработчик set custom status.
func TestHandler_SetCustomStatus(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	body := `{"text":"In a meeting","duration":"1h"}`
	req := httptest.NewRequest(http.MethodPut, "/api/messenger/presence/status", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_SetCustomStatus_BadJSON проверяет HTTP-слой: HTTP-обработчик set custom status bad json.
func TestHandler_SetCustomStatus_BadJSON(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodPut, "/api/messenger/presence/status", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_RemoveBookmark проверяет HTTP-слой: HTTP-обработчик remove bookmark.
func TestHandler_RemoveBookmark(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	msg := postChannelMessage(t, app, channelID, "bookmark me")

	addBody, _ := json.Marshal(map[string]string{"message_id": msg.ID, "channel_id": channelID.String()})
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/bookmarks", bytes.NewReader(addBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/messenger/bookmarks", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var bookmarks []messenger.BookmarkResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&bookmarks))
	require.Len(t, bookmarks, 1)

	req = httptest.NewRequest(http.MethodDelete, "/api/messenger/bookmarks/"+bookmarks[0].ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestHandler_RemoveBookmark_InvalidID проверяет отклонение при HTTP-обработчик remove bookmark некорректный id.
func TestHandler_RemoveBookmark_InvalidID(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/messenger/bookmarks/not-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
