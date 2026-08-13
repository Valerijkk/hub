package messenger_test

// Тесты HTTP-обработчика мессенджера (волна 106): интеграционные и unit-сценарии testframework.
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

// TestHandler_GetChannel_NotFound проверяет HTTP-слой: HTTP-обработчик получение channel not found.
func TestHandler_GetChannel_NotFound(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/channels/"+uuid.New().String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_CreateChannel_InvalidJSON проверяет отклонение при HTTP-обработчик создание channel некорректный json.
func TestHandler_CreateChannel_InvalidJSON(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_JoinPrivateChannel_Forbidden проверяет отклонение при HTTP-обработчик join private channel запрещённый.
func TestHandler_JoinPrivateChannel_Forbidden(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/join", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	_ = userID
}

// TestHandler_CreateOrGetDM_InvalidTarget проверяет отклонение при HTTP-обработчик создание or получение dm некорректный target.
func TestHandler_CreateOrGetDM_InvalidTarget(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	body := `{"target_user_id":"not-a-uuid"}`
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/dm", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_MuteChannel_InvalidJSON проверяет отклонение при HTTP-обработчик mute channel некорректный json.
func TestHandler_MuteChannel_InvalidJSON(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/mute", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_AddMembers_InvalidMemberID проверяет отклонение при HTTP-обработчик add members некорректный участник id.
func TestHandler_AddMembers_InvalidMemberID(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	body, _ := json.Marshal(map[string]any{"member_ids": []string{"bad-uuid"}})
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/members", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ValidateInvite_MissingToken проверяет отклонение при HTTP-обработчик validate приглашение отсутствующий token.
func TestHandler_ValidateInvite_MissingToken(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/invites/missing-token", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result messenger.InviteValidationResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.False(t, result.Valid)
}
