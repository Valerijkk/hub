package messenger_test

// Тесты HTTP-обработчика мессенджера (волна 93): интеграционные и unit-сценарии testframework.
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

// TestHandler_Wave93_ListThreadRepliesAndRevokeInvite проверяет HTTP-слой: HTTP-обработчик wave 93 список thread replies and revoke приглашение.
func TestHandler_Wave93_ListThreadRepliesAndRevokeInvite(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	parent := postChannelMessage(t, app, channelID, "thread root")
	replyBody, _ := json.Marshal(map[string]any{
		"content_text": "thread reply",
		"content_json": map[string]string{"text": "thread reply"},
		"thread_id":    parent.ID,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/messages", bytes.NewReader(replyBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/messenger/channels/"+channelID.String()+"/messages/"+parent.ID+"/replies?limit=10", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var replies []messenger.MessageResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&replies))
	require.Len(t, replies, 1)
	assert.Equal(t, "thread reply", replies[0].ContentText)

	privateID := setupChannelWithMembers(repo, "private", false, userID)
	inviteBody := `{"max_uses":3}`
	req = httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+privateID.String()+"/invites", bytes.NewReader([]byte(inviteBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var invite messenger.ChannelInviteResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&invite))

	req = httptest.NewRequest(http.MethodDelete, "/api/messenger/channels/"+privateID.String()+"/invites/"+invite.ID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave93_ListThreadReplies_InvalidBefore проверяет отклонение при HTTP-обработчик wave 93 список thread replies некорректный before.
func TestHandler_Wave93_ListThreadReplies_InvalidBefore(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	parent := postChannelMessage(t, app, channelID, "root")

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/channels/"+channelID.String()+"/messages/"+parent.ID+"/replies?before=not-rfc3339", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave93_RevokeInvite_InvalidID проверяет отклонение при HTTP-обработчик wave 93 revoke приглашение некорректный id.
func TestHandler_Wave93_RevokeInvite_InvalidID(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "private", false, userID)

	req := httptest.NewRequest(http.MethodDelete, "/api/messenger/channels/"+channelID.String()+"/invites/not-a-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave93_PollVoteCloseAndResults проверяет HTTP-слой: HTTP-обработчик wave 93 опрос vote close and results.
func TestHandler_Wave93_PollVoteCloseAndResults(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	createBody := `{"question":"Pick one","options":["A","B"],"is_anonymous":false,"is_multi_choice":false}`
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/polls", bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var poll messenger.PollResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&poll))
	require.Len(t, poll.Options, 2)

	voteBody, _ := json.Marshal(messenger.VotePollRequest{OptionIDs: []string{poll.Options[0].ID}})
	req = httptest.NewRequest(http.MethodPost, "/api/messenger/polls/"+poll.ID+"/vote", bytes.NewReader(voteBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/messenger/polls/"+poll.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/messenger/polls/"+poll.ID+"/close", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
