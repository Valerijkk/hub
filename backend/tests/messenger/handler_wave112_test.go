package messenger_test

// Тесты HTTP-обработчика мессенджера (волна 112): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_MarkChannelRead_Success проверяет HTTP-слой: HTTP-обработчик mark channel read успех.
func TestHandler_MarkChannelRead_Success(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	postChannelMessage(t, app, channelID, "unread ping")

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/read", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// TestHandler_GetUnreadCounts_Success проверяет HTTP-слой: HTTP-обработчик получение непрочитанные counts успех.
func TestHandler_GetUnreadCounts_Success(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	postChannelMessage(t, app, channelID, "count me")

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/unread", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_DeleteMessage_Success проверяет HTTP-слой: HTTP-обработчик удаление message успех.
func TestHandler_DeleteMessage_Success(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	msg := postChannelMessage(t, app, channelID, "delete me")

	req := httptest.NewRequest(http.MethodDelete, "/api/messenger/channels/"+channelID.String()+"/messages/"+msg.ID, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

// TestHandler_RemoveBookmark_Success проверяет HTTP-слой: HTTP-обработчик remove bookmark успех.
func TestHandler_RemoveBookmark_Success(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	msg := postChannelMessage(t, app, channelID, "bookmark rm")

	addBody, _ := json.Marshal(map[string]string{"message_id": msg.ID, "channel_id": channelID.String()})
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/bookmarks", bytes.NewReader(addBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, "/api/messenger/bookmarks", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var bookmarks []messenger.BookmarkResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&bookmarks))
	require.Len(t, bookmarks, 1)

	req = httptest.NewRequest(http.MethodDelete, "/api/messenger/bookmarks/"+bookmarks[0].ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

// TestHandler_UpdateChannel_InvalidJSON проверяет отклонение при HTTP-обработчик обновление channel некорректный json.
func TestHandler_UpdateChannel_InvalidJSON(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	req := httptest.NewRequest(http.MethodPut, "/api/messenger/channels/"+channelID.String(), bytes.NewReader([]byte("[")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_PromoteToIssue_InvalidJSON проверяет отклонение при HTTP-обработчик promote to issue некорректный json.
func TestHandler_PromoteToIssue_InvalidJSON(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	msg := postChannelMessage(t, app, channelID, "promote me")

	req := httptest.NewRequest(http.MethodPost, "/api/messages/"+msg.ID+"/promote-to-issue", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_CreatePoll_InvalidJSON проверяет отклонение при HTTP-обработчик создание опрос некорректный json.
func TestHandler_CreatePoll_InvalidJSON(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/polls", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_AcceptInvite_InvalidToken проверяет отклонение при HTTP-обработчик accept приглашение некорректный token.
func TestHandler_AcceptInvite_InvalidToken(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/invites/"+uuid.New().String()+"/accept", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}
