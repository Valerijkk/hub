package messenger_test

// Тесты HTTP-обработчика мессенджера: интеграционные и unit-сценарии testframework.
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

// TestHandler_JoinAndLeaveChannel проверяет HTTP-слой: HTTP-обработчик join and leave channel.
func TestHandler_JoinAndLeaveChannel(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/join", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	isMember, err := repo.IsMember(t.Context(), channelID, userID)
	require.NoError(t, err)
	assert.True(t, isMember)

	req = httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/leave", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ToggleStarAndMute проверяет HTTP-слой: HTTP-обработчик toggle star and mute.
func TestHandler_ToggleStarAndMute(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/star", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var starResult map[string]bool
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&starResult))
	assert.True(t, starResult["starred"])

	body := `{"duration":"1h"}`
	req = httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/mute", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/unmute", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ListMembersAndCreateInvite проверяет HTTP-слой: HTTP-обработчик список members and создание приглашение.
func TestHandler_ListMembersAndCreateInvite(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, userID, memberID)

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/channels/"+channelID.String()+"/members", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var members []messenger.ChannelMemberResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&members))
	assert.GreaterOrEqual(t, len(members), 2)

	inviteBody := `{"max_uses":10}`
	req = httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/invites", bytes.NewReader([]byte(inviteBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/messenger/channels/"+channelID.String()+"/invites", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_EditAndDeleteMessage проверяет HTTP-слой: HTTP-обработчик edit and удаление message.
func TestHandler_EditAndDeleteMessage(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	msg := postChannelMessage(t, app, channelID, "edit me")

	editBody := `{"content_text":"edited","content_json":{"text":"edited"}}`
	req := httptest.NewRequest(http.MethodPut, "/api/messenger/channels/"+channelID.String()+"/messages/"+msg.ID, bytes.NewReader([]byte(editBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, "/api/messenger/channels/"+channelID.String()+"/messages/"+msg.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

// TestHandler_SendTyping проверяет HTTP-слой: HTTP-обработчик отправка typing.
func TestHandler_SendTyping(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/typing", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
