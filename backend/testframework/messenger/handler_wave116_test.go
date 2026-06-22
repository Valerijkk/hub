package messenger_test

// Тесты HTTP-обработчика мессенджера (волна 116): интеграционные и unit-сценарии testframework.
import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_JoinChannel_InvalidChannelID проверяет отклонение при HTTP-обработчик join channel некорректный channel id.
func TestHandler_JoinChannel_InvalidChannelID(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/bad-id/join", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_AddMembers_InvalidJSON проверяет отклонение при HTTP-обработчик add members некорректный json.
func TestHandler_AddMembers_InvalidJSON(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/members", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_ListMembers_InvalidChannelID проверяет отклонение при HTTP-обработчик список members некорректный channel id.
func TestHandler_ListMembers_InvalidChannelID(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/channels/not-uuid/members", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
