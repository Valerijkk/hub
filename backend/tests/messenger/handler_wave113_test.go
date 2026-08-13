package messenger_test

// Тесты HTTP-обработчика мессенджера (волна 113): интеграционные и unit-сценарии testframework.
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

// TestHandler_LeaveChannel_InvalidChannelID проверяет отклонение при HTTP-обработчик leave channel некорректный channel id.
func TestHandler_LeaveChannel_InvalidChannelID(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/not-a-uuid/leave", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_DeleteChannel_InvalidChannelID проверяет отклонение при HTTP-обработчик удаление channel некорректный channel id.
func TestHandler_DeleteChannel_InvalidChannelID(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/messenger/channels/bad-id", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_ForwardMessage_InvalidJSON проверяет отклонение при HTTP-обработчик forward message некорректный json.
func TestHandler_ForwardMessage_InvalidJSON(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	msg := postChannelMessage(t, app, channelID, "forward me")

	path := "/api/messenger/channels/" + channelID.String() + "/messages/" + msg.ID + "/forward"
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_ForwardMessage_Success проверяет HTTP-слой: HTTP-обработчик forward message успех.
func TestHandler_ForwardMessage_Success(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	sourceID := setupChannelWithMembers(repo, "public", false, userID)
	targetID := setupChannelWithMembers(repo, "public", false, userID)
	msg := postChannelMessage(t, app, sourceID, "ship it")

	body, _ := json.Marshal(messenger.ForwardMessageRequest{TargetChannelID: targetID.String()})
	path := "/api/messenger/channels/" + sourceID.String() + "/messages/" + msg.ID + "/forward"
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// TestHandler_UnmuteChannel_Success проверяет HTTP-слой: HTTP-обработчик unmute channel успех.
func TestHandler_UnmuteChannel_Success(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	muteBody := `{"duration":"1h"}`
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/mute", bytes.NewReader([]byte(muteBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/unmute", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
