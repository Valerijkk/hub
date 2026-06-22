package messenger_test

// Тесты HTTP-обработчика мессенджера (волна 115): интеграционные и unit-сценарии testframework.
import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_TogglePin_Success проверяет HTTP-слой: HTTP-обработчик toggle pin успех.
func TestHandler_TogglePin_Success(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	msg := postChannelMessage(t, app, channelID, "pin me")

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/messages/"+msg.ID+"/pin", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_TogglePin_InvalidChannelID проверяет отклонение при HTTP-обработчик toggle pin некорректный channel id.
func TestHandler_TogglePin_InvalidChannelID(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/bad-id/messages/msg-1/pin", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_DeleteMessage_InvalidMessageID проверяет отклонение при HTTP-обработчик удаление message некорректный message id.
func TestHandler_DeleteMessage_InvalidMessageID(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	req := httptest.NewRequest(http.MethodDelete, "/api/messenger/channels/"+channelID.String()+"/messages/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusNoContent, resp.StatusCode)
	_ = userID
}
