package messenger_test

// Тесты HTTP-обработчика мессенджера (волна 114): интеграционные и unit-сценарии testframework.
import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_ListPins_Success проверяет HTTP-слой: HTTP-обработчик список pins успех.
func TestHandler_ListPins_Success(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	msg := postChannelMessage(t, app, channelID, "pin target")

	require.NoError(t, repo.AddPin(t.Context(), channelID, msg.ID, userID))

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/channels/"+channelID.String()+"/pins", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ListPins_InvalidChannelID проверяет отклонение при HTTP-обработчик список pins некорректный channel id.
func TestHandler_ListPins_InvalidChannelID(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/channels/bad-id/pins", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_GetPollResults_InvalidChannelID проверяет отклонение при HTTP-обработчик получение опрос results некорректный channel id.
func TestHandler_GetPollResults_InvalidChannelID(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/channels/not-a-uuid/polls/"+uuid.New().String()+"/results", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}
