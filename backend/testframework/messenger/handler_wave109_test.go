package messenger_test

// Тесты HTTP-обработчика мессенджера (волна 109): интеграционные и unit-сценарии testframework.
import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_SendTyping_Success проверяет HTTP-слой: HTTP-обработчик отправка typing успех.
func TestHandler_SendTyping_Success(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/typing", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

// TestHandler_Heartbeat_Success проверяет HTTP-слой: HTTP-обработчик heartbeat успех.
func TestHandler_Heartbeat_Success(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/presence/heartbeat", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ListMessages_InvalidBefore проверяет отклонение при HTTP-обработчик список сообщения некорректный before.
func TestHandler_ListMessages_InvalidBefore(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/channels/"+channelID.String()+"/messages?before=bad-date", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_SendMessage_InvalidJSON проверяет отклонение при HTTP-обработчик отправка message некорректный json.
func TestHandler_SendMessage_InvalidJSON(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/messages", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_ToggleReaction_InvalidJSON проверяет отклонение при HTTP-обработчик toggle reaction некорректный json.
func TestHandler_ToggleReaction_InvalidJSON(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	msg := postChannelMessage(t, app, channelID, "react target")

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/messages/"+msg.ID+"/reactions", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_SearchMessages_MissingQuery проверяет отклонение при HTTP-обработчик search сообщения отсутствующий query.
func TestHandler_SearchMessages_MissingQuery(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/search", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_SendTyping_InvalidChannelID проверяет отклонение при HTTP-обработчик отправка typing некорректный channel id.
func TestHandler_SendTyping_InvalidChannelID(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/not-a-uuid/typing", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_EditMessage_InvalidJSON проверяет отклонение при HTTP-обработчик edit message некорректный json.
func TestHandler_EditMessage_InvalidJSON(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	msg := postChannelMessage(t, app, channelID, "edit me")

	req := httptest.NewRequest(http.MethodPut, "/api/messenger/channels/"+channelID.String()+"/messages/"+msg.ID, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
