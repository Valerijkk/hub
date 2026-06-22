package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"

	messengermod "github.com/valerii/hub/backend/internal/messenger"
)

func TestService_UpdateProvider(t *testing.T) {
	pid := uuid.New()
	name := "Updated"
	repo := &mockRepo{providers: []*ent.AIProvider{{
		ID: pid, Name: "Old", BaseURL: "http://127.0.0.1/v1",
		Models: []string{"m1"}, Enabled: true, Scope: aiprovider.ScopeSystem,
	}}}
	svc := ai.NewService(repo, &mockAdapter{resp: &ai.ChatCompletionResponse{}}, nil)

	updated, err := svc.UpdateProvider(context.Background(), pid, ai.UpdateProviderRequest{Name: &name})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
}

func TestHandler_ChatCompletions_Success(t *testing.T) {
	app, _ := testAIHandlerApp(t)
	body, _ := json.Marshal(ai.ChatCompletionRequest{
		Model:    "deepseek-chat",
		Messages: []ai.ChatMessage{{Role: "user", Content: "hello"}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/ai/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_ChatCompletions_StreamAccepted(t *testing.T) {
	svc, userID := testAIService(t)
	svc.SetStreamPublisher(ai.NewWSStreamPublisher(nil))
	h := ai.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	body, _ := json.Marshal(ai.ChatCompletionRequest{
		Model:    "deepseek-chat",
		Stream:   true,
		StreamID: "stream-abc",
		Messages: []ai.ChatMessage{{Role: "user", Content: "stream"}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/ai/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestHandler_UpdatePreferences_Success(t *testing.T) {
	app, _ := testAIHandlerApp(t)
	model := "deepseek-chat"
	body, _ := json.Marshal(map[string]interface{}{"default_model": model})
	req := httptest.NewRequest(http.MethodPatch, "/api/ai/preferences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_AssistJQL_Success(t *testing.T) {
	app, _ := testAIHandlerApp(t)
	body, _ := json.Marshal(ai.JQLAssistRequest{Query: "open bugs assigned to me"})
	req := httptest.NewRequest(http.MethodPost, "/api/ai/jql/assist", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_SuggestComposer_Success(t *testing.T) {
	app, _ := testAIHandlerApp(t)
	body, _ := json.Marshal(ai.ComposerSuggestRequest{Prefix: "Fix login bug"})
	req := httptest.NewRequest(http.MethodPost, "/api/ai/composer/suggest", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAdminHandler_UpdateProvider(t *testing.T) {
	app, _ := testAdminAIApp(t, true)

	createBody, _ := json.Marshal(ai.CreateProviderRequest{
		Name: "PatchMe", ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL: "https://api.example.com/v1", Models: []string{"m"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/ai/providers", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var created ai.ProviderResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	newName := "Patched"
	patchBody, _ := json.Marshal(ai.UpdateProviderRequest{Name: &newName})
	req = httptest.NewRequest(http.MethodPatch, "/api/admin/ai/providers/"+created.ID, bytes.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRedisProviderRPMLimiter_AllowAndBlock(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	limiter := ai.NewRedisProviderRPMLimiter(rdb)
	pid := uuid.New()
	ctx := context.Background()

	require.NoError(t, limiter.Allow(ctx, pid, 1))
	err = limiter.Allow(ctx, pid, 1)
	require.ErrorIs(t, err, ai.ErrProviderRateLimited)
}

func TestNewUnifiedCodeReviewProvider(t *testing.T) {
	p := ai.NewUnifiedCodeReviewProvider(nil, nil)
	require.NotNil(t, p)
	_, err := p.FetchReview(context.Background(), "https://github.com/o/r/pull/1")
	require.Error(t, err)
}

func TestBuildChatRequestAndExtractAssistantText(t *testing.T) {
	botID := uuid.New()
	req := ai.ExportBuildChatRequest("system", []messengermod.MessageResponse{
		{SenderID: uuid.New().String(), ContentText: "hi"},
		{SenderID: botID.String(), ContentText: "hello"},
	}, botID, "next")
	require.Len(t, req.Messages, 4)
	assert.Equal(t, "assistant", req.Messages[2].Role)

	assert.Empty(t, ai.ExportExtractAssistantText(nil))
	assert.Equal(t, "done", ai.ExportExtractAssistantText(&ai.ChatCompletionResponse{
		Choices: []struct {
			Index        int         `json:"index"`
			Message      ai.ChatMessage `json:"message"`
			FinishReason string      `json:"finish_reason"`
		}{{Message: ai.ChatMessage{Content: "done"}}},
	}))
}

func TestMessengerBot_Setters(t *testing.T) {
	bot := ai.NewMessengerBot(ai.NewService(&mockRepo{}, &mockAdapter{}, nil), nil, uuid.New())
	bot.SetIssueLookup(nil)
	bot.SetDocSearch(stubDocSearch{})
	bot.SetCodeReview(&ai.UnifiedCodeReviewProvider{})
	assert.NotNil(t, bot)
}

type stubDocSearch struct{}

func (stubDocSearch) SearchContext(_ context.Context, _ uuid.UUID, _ string) (string, error) {
	return "doc context", nil
}

func TestSplitAnthropicMessagesAndAnthropicText(t *testing.T) {
	sys, msgs := ai.ExportSplitAnthropicMessages([]ai.ChatMessage{
		{Role: "system", Content: "a"},
		{Role: "system", Content: "b"},
		{Role: "assistant", Content: "hi"},
		{Role: "user", Content: "q"},
	})
	assert.Contains(t, sys, "a")
	assert.Len(t, msgs, 2)

	text := ai.ExportAnthropicTextFromBlocks("hello")
	assert.Equal(t, "hello", text)
}

func TestHandler_ChatCompletions_InvalidBody(t *testing.T) {
	app := fiber.New()
	svc, userID := testAIService(t)
	h := ai.NewHandler(svc)
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	req := httptest.NewRequest(http.MethodPost, "/api/ai/v1/chat/completions", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
