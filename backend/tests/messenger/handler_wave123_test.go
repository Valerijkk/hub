package messenger_test

// Тесты HTTP-обработчика мессенджера (волна 123): интеграционные и unit-сценарии testframework.
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

// Фаза 123 — handler success paths for mute, bookmarks, invites..

func TestHandler_Wave123_MuteChannel_Success(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	for _, dur := range []string{"1h", "2h", "8h", "1d", "forever"} {
		body := `{"duration":"` + dur + `"}`
		req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/mute", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// W3-A1-012 / P3-BE-02: unmute = POST /unmute (DELETE /mute маршрута нет).
		req = httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/unmute", nil)
		resp, err = app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

// TestHandler_Wave123_AddAndListBookmarks проверяет HTTP-слой: HTTP-обработчик wave 123 add and список bookmarks.
func TestHandler_Wave123_AddAndListBookmarks(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	msg := postChannelMessage(t, app, channelID, "bookmark me")

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/bookmarks", bytes.NewReader([]byte(
		`{"message_id":"`+msg.ID+`","channel_id":"`+channelID.String()+`"}`,
	)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/messenger/bookmarks", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var bookmarks []messenger.BookmarkResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&bookmarks))
	assert.NotEmpty(t, bookmarks)
}

// TestHandler_Wave123_CreateChannelInvite проверяет HTTP-слой: HTTP-обработчик wave 123 создание channel приглашение.
func TestHandler_Wave123_CreateChannelInvite(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	body := `{"max_uses":5}`
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/invites", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/messenger/channels/"+channelID.String()+"/invites", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave123_ToggleStar_Success проверяет HTTP-слой: HTTP-обработчик wave 123 toggle star успех.
func TestHandler_Wave123_ToggleStar_Success(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/star", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
