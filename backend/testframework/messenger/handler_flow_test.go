package messenger_test

// Тесты HTTP-обработчика мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

func setupMessengerHandlerApp(t *testing.T) (*fiber.App, *mockRepository, uuid.UUID) {
	t.Helper()
	svc, repo, _ := newTestService()
	h := messenger.NewHandler(svc)
	userID := uuid.New()

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, nil)
	return app, repo, userID
}

// TestHandler_CreateChannelAndList проверяет HTTP-слой: HTTP-обработчик создание channel and список.
func TestHandler_CreateChannelAndList(t *testing.T) {
	app, _, userID := setupMessengerHandlerApp(t)

	body := `{"name":"general","channel_type":"public","description":"team chat"}`
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/messenger/channels", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var channels []messenger.ChannelResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&channels))
	require.NotEmpty(t, channels)
	_ = userID
}

// TestHandler_SendAndListMessages проверяет HTTP-слой: HTTP-обработчик отправка and список сообщения.
func TestHandler_SendAndListMessages(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	payload := `{"content_text":"hello via handler","content_json":{"text":"hello via handler"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/messages", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/messenger/channels/"+channelID.String()+"/messages?limit=20", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_CreateOrGetDM проверяет HTTP-слой: HTTP-обработчик создание or получение dm.
func TestHandler_CreateOrGetDM(t *testing.T) {
	app, _, userID := setupMessengerHandlerApp(t)
	targetID := uuid.New()

	body, _ := json.Marshal(map[string]string{"target_user_id": targetID.String()})
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/dm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = userID
}

// TestHandler_SearchMessages проверяет HTTP-слой: HTTP-обработчик search сообщения.
func TestHandler_SearchMessages(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	payload := `{"content_text":"searchable-keyword-xyz","content_json":{"text":"searchable-keyword-xyz"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/messages", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, "/api/messenger/search?query=searchable-keyword&channel_id="+channelID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_BookmarksFlow проверяет HTTP-слой: HTTP-обработчик bookmarks поток.
func TestHandler_BookmarksFlow(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	payload := `{"content_text":"bookmark me","content_json":{"text":"bookmark me"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/messages", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var msg messenger.MessageResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&msg))
	resp.Body.Close()

	addBody, _ := json.Marshal(map[string]string{"message_id": msg.ID, "channel_id": channelID.String()})
	req = httptest.NewRequest(http.MethodPost, "/api/messenger/bookmarks", bytes.NewReader(addBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/messenger/bookmarks", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ValidateInvite_Public проверяет HTTP-слой: HTTP-обработчик validate приглашение public.
func TestHandler_ValidateInvite_Public(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "private", false, userID)
	inv := &ent.ChannelInvite{
		ID:        uuid.New(),
		ChannelID: channelID,
		Token:     "public-token",
		MaxUses:   5,
		CreatedAt: time.Now(),
	}
	repo.invites[inv.ID] = inv
	repo.invitesByToken[inv.Token] = inv

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/invites/public-token", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Unauthorized проверяет отклонение при HTTP-обработчик неавторизованный.
func TestHandler_Unauthorized(t *testing.T) {
	svc, _, _ := newTestService()
	h := messenger.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() }, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/channels", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestHandler_CreatePoll проверяет HTTP-слой: HTTP-обработчик создание опрос.
func TestHandler_CreatePoll(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	body := `{"question":"Lunch?","options":["Pizza","Salad"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/polls", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	raw, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(raw), "Lunch?")
}

func postChannelMessage(t *testing.T, app *fiber.App, channelID uuid.UUID, content string) messenger.MessageResponse {
	t.Helper()
	payload := `{"content_text":"` + content + `","content_json":{"text":"` + content + `"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/messages", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var msg messenger.MessageResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&msg))
	return msg
}

// TestHandler_Presence_GetAndHeartbeat проверяет HTTP-слой: HTTP-обработчик присутствие получение and heartbeat.
func TestHandler_Presence_GetAndHeartbeat(t *testing.T) {
	app, _, userID := setupMessengerHandlerApp(t)
	otherID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/presence?user_ids="+userID.String()+","+otherID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var statuses map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&statuses))
	assert.NotNil(t, statuses)

	req = httptest.NewRequest(http.MethodPost, "/api/messenger/presence/heartbeat", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Presence_MissingUserIDs проверяет отклонение при HTTP-обработчик присутствие отсутствующий user i ds.
func TestHandler_Presence_MissingUserIDs(t *testing.T) {
	app, _, _ := setupMessengerHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/presence", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_ToggleReaction проверяет HTTP-слой: HTTP-обработчик toggle reaction.
func TestHandler_ToggleReaction(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	msg := postChannelMessage(t, app, channelID, "react via handler")

	body := `{"emoji":"??"}`
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/messages/"+msg.ID+"/reactions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]bool
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result["added"])
}

// TestHandler_TogglePinAndListPins проверяет HTTP-слой: HTTP-обработчик toggle pin and список pins.
func TestHandler_TogglePinAndListPins(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	msg := postChannelMessage(t, app, channelID, "pin via handler")

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/messages/"+msg.ID+"/pin", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var pinResult map[string]bool
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pinResult))
	assert.True(t, pinResult["pinned"])

	req = httptest.NewRequest(http.MethodGet, "/api/messenger/channels/"+channelID.String()+"/pins", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var pins []messenger.PinnedMessageResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pins))
	require.Len(t, pins, 1)
	assert.Equal(t, msg.ID, pins[0].MessageID)
}

// TestHandler_ForwardMessage проверяет HTTP-слой: HTTP-обработчик forward message.
func TestHandler_ForwardMessage(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	sourceChannelID := setupChannelWithMembers(repo, "public", false, userID)
	targetChannelID := setupChannelWithMembers(repo, "public", false, userID)
	msg := postChannelMessage(t, app, sourceChannelID, "forward me")

	body, _ := json.Marshal(messenger.ForwardMessageRequest{TargetChannelID: targetChannelID.String()})
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+sourceChannelID.String()+"/messages/"+msg.ID+"/forward", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var forwarded messenger.MessageResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&forwarded))
	assert.Contains(t, forwarded.ContentText, "forward me")
}

// TestHandler_GetAndUpdateChannel проверяет HTTP-слой: HTTP-обработчик получение and обновление channel.
func TestHandler_GetAndUpdateChannel(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/channels/"+channelID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var channel messenger.ChannelResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&channel))
	assert.Equal(t, channelID.String(), channel.ID)

	updateBody := `{"name":"renamed-via-handler","description":"updated desc"}`
	req = httptest.NewRequest(http.MethodPut, "/api/messenger/channels/"+channelID.String(), bytes.NewReader([]byte(updateBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&channel))
	assert.Equal(t, "renamed-via-handler", channel.Name)
}

// TestHandler_UnreadCountsAndMarkRead проверяет HTTP-слой: HTTP-обработчик непрочитанные counts and mark read.
func TestHandler_UnreadCountsAndMarkRead(t *testing.T) {
	repo := newMockRepository()
	_, rdb := newTestRedis(t)
	svc := messenger.NewMessengerService(repo, nil, messenger.NewPresenceService(rdb), messenger.NewUnreadService(rdb))
	h := messenger.NewHandler(svc)
	userID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	ctx := t.Context()
	require.NoError(t, messenger.ExportMessengerServiceUnread(svc).Increment(ctx, channelID.String(), []string{userID.String()}, uuid.New().String()))

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/messenger/unread", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var counts map[string]int64
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&counts))
	assert.Equal(t, int64(1), counts[channelID.String()])

	req = httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/read", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/messenger/unread", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	counts = make(map[string]int64)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&counts))
	assert.Zero(t, counts[channelID.String()])
}

// TestHandler_PromoteToIssue_Success проверяет HTTP-слой: HTTP-обработчик promote to issue успех.
func TestHandler_PromoteToIssue_Success(t *testing.T) {
	svc, repo, _ := newTestService()
	userID := uuid.New()
	projectID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	svc.SetIssuePromoter(&mockIssuePromoter{issueKey: "HUB-42", issueID: uuid.New().String()})

	msgID := uuid.New().String()
	repo.messages[msgID] = &messenger.MessageDocument{
		ID: msgID, ChannelID: channelID, SenderID: userID,
		ContentText: "promote me", CreatedAt: time.Now(),
	}

	h := messenger.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, nil)

	body := `{"project_id":"` + projectID.String() + `","title":"From chat","issue_type":"task"}`
	req := httptest.NewRequest(http.MethodPost, "/api/messages/"+msgID+"/promote-to-issue", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var promoted messenger.PromoteToIssueResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&promoted))
	assert.Equal(t, "HUB-42", promoted.IssueKey)
}

// TestHandler_AcceptInvite_Success проверяет HTTP-слой: HTTP-обработчик accept приглашение успех.
func TestHandler_AcceptInvite_Success(t *testing.T) {
	svc, repo, _ := newTestService()
	ownerID := uuid.New()
	joinerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)

	inv := &ent.ChannelInvite{
		ID: uuid.New(), ChannelID: channelID, Token: "handler-join",
		MaxUses: 10, UseCount: 0, CreatedAt: time.Now(),
	}
	repo.invites[inv.ID] = inv
	repo.invitesByToken[inv.Token] = inv

	h := messenger.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", joinerID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/messenger/invites/handler-join/accept", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var ch messenger.ChannelResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&ch))
	assert.Equal(t, channelID.String(), ch.ID)
}
