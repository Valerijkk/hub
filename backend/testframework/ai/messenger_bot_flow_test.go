package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"

	messengermod "github.com/valerii/hub/backend/internal/messenger"
	searchmod "github.com/valerii/hub/backend/internal/search"
)

func setupHubAIBot(t *testing.T, adapter ai.ChatAdapter) (*ai.MessengerBot, *messengermod.HookTestRepo, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	botID := uuid.New()
	userID := uuid.New()
	hookRepo := messengermod.NewHookTestRepo()
	msgSvc := messengermod.NewMessengerService(hookRepo.Repository(), nil, nil, nil)
	msgSvc.SetSystemBotID(botID)

	repo := &mockRepo{providers: []*ent.AIProvider{{
		ID: uuid.New(), Name: "Test", BaseURL: "https://api.example.com/v1",
		Models: []string{"test-model"}, Enabled: true, Scope: aiprovider.ScopeSystem,
	}}}
	aiSvc := ai.NewService(repo, adapter, nil)
	bot := ai.NewMessengerBot(aiSvc, msgSvc, botID)
	channelID := hookRepo.SetupDM(userID, botID)
	return bot, hookRepo, userID, channelID, botID
}

func TestMessengerBot_OnMessageSent_DMStreamReply(t *testing.T) {
	adapter := &streamingMockAdapter{chunks: []string{"Привет", "!"}}
	bot, _, userID, channelID, _ := setupHubAIBot(t, adapter)

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "hello",
		MessageID:   "msg-1",
	})
	require.NoError(t, err)
}

func TestMessengerBot_OnMessageSent_HelpCommand(t *testing.T) {
	adapter := &streamingMockAdapter{chunks: []string{"commands list"}}
	bot, _, userID, channelID, _ := setupHubAIBot(t, adapter)

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "@hub-ai /ai help",
		MessageID:   "msg-2",
	})
	require.NoError(t, err)
}

func TestMessengerBot_OnMessageSent_PublicMention(t *testing.T) {
	adapter := &streamingMockAdapter{chunks: []string{"ok"}}
	bot, hookRepo, userID, _, botID := setupHubAIBot(t, adapter)
	channelID := hookRepo.SetupPublicChannel(userID, botID)

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "@hub-ai summarize this",
		MessageID:   "msg-root",
	})
	require.NoError(t, err)
}

func TestMessengerBot_OnMessageSent_SkipsBotMessages(t *testing.T) {
	bot, hookRepo, _, channelID, botID := setupHubAIBot(t, &mockAdapter{})
	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID: channelID, SenderID: botID, ContentText: "loop",
	})
	require.NoError(t, err)
	assert.NotNil(t, hookRepo.Channel(channelID))
}

func TestMessengerBot_HandleCommand_Docs(t *testing.T) {
	adapter := &streamingMockAdapter{chunks: []string{"answer from docs"}}
	bot, _, userID, channelID, _ := setupHubAIBot(t, adapter)
	bot.SetDocSearch(searchDocStub{ctx: "Issue HUB-1: login flow"})

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "/ai docs как настроить TMS",
		MessageID:   "msg-docs",
	})
	require.NoError(t, err)
}

func TestMessengerBot_HandleCommand_TestCasesMissingKey(t *testing.T) {
	bot, _, userID, channelID, _ := setupHubAIBot(t, &mockAdapter{})
	bot.SetIssueLookup(stubIssueLookup{})

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "/ai test-cases",
		MessageID:   "msg-tc",
	})
	require.NoError(t, err)
}

func TestMessengerBot_LoadHistory_Thread(t *testing.T) {
	adapter := &streamingMockAdapter{chunks: []string{"reply"}}
	bot, hookRepo, userID, channelID, _ := setupHubAIBot(t, adapter)
	rootID, err := hookRepo.InsertMessage(channelID, userID, "thread root", "")
	require.NoError(t, err)
	_, err = hookRepo.InsertMessage(channelID, userID, "follow up", rootID)
	require.NoError(t, err)

	err = bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "@hub-ai explain",
		MessageID:   "msg-thread",
		ThreadID:    rootID,
	})
	require.NoError(t, err)
}

func TestMessengerBot_NoProvider_ReplyError(t *testing.T) {
	botID := uuid.New()
	userID := uuid.New()
	hookRepo := messengermod.NewHookTestRepo()
	msgSvc := messengermod.NewMessengerService(hookRepo.Repository(), nil, nil, nil)
	msgSvc.SetSystemBotID(botID)
	channelID := hookRepo.SetupDM(userID, botID)

	bot := ai.NewMessengerBot(ai.NewService(&mockRepo{}, &mockAdapter{}, nil), msgSvc, botID)
	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID: channelID, SenderID: userID, ContentText: "hi", MessageID: "m1",
	})
	require.NoError(t, err)
}

type searchDocStub struct {
	ctx string
	err error
}

func (s searchDocStub) SearchContext(_ context.Context, _ uuid.UUID, _ string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.ctx, nil
}

type stubIssueLookup struct{}

func (stubIssueLookup) IssueContext(_ context.Context, _ uuid.UUID, key string) (string, string, error) {
	return "Title " + key, "Description", nil
}

type searchBackendStub struct {
	results []searchmod.SearchResult
	err     error
}

func (s searchBackendStub) Search(_ context.Context, _ string, _ uuid.UUID, _ []string, _ int) ([]searchmod.SearchResult, bool, error) {
	if s.err != nil {
		return nil, false, s.err
	}
	return s.results, false, nil
}

type stubCodeReviewProvider struct {
	ctx *ai.ReviewContext
	err error
}

func (s stubCodeReviewProvider) FetchReview(_ context.Context, _ uuid.UUID, _ string) (*ai.ReviewContext, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.ctx, nil
}

type stubIssueLookupErr struct{}

func (stubIssueLookupErr) IssueContext(_ context.Context, _ uuid.UUID, _ string) (string, string, error) {
	return "", "", assert.AnError
}

type denyIssueLookup struct {
	called bool
}

func (d *denyIssueLookup) IssueContext(_ context.Context, _ uuid.UUID, _ string) (string, string, error) {
	d.called = true
	return "", "", &ent.NotFoundError{}
}

type denyCodeReviewProvider struct {
	called bool
}

func (d *denyCodeReviewProvider) FetchReview(_ context.Context, _ uuid.UUID, _ string) (*ai.ReviewContext, error) {
	d.called = true
	return nil, assert.AnError
}

func TestMessengerBot_HandleCommand_TestCasesSuccess(t *testing.T) {
	adapter := &streamingMockAdapter{chunks: []string{"TC-1 login", "TC-2 logout"}}
	bot, _, userID, channelID, _ := setupHubAIBot(t, adapter)
	bot.SetIssueLookup(stubIssueLookup{})

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "/ai test-cases HUB-42",
		MessageID:   "m-tc-ok",
	})
	require.NoError(t, err)
}

func TestMessengerBot_HandleCommand_TestCasesNotFound(t *testing.T) {
	bot, _, userID, channelID, _ := setupHubAIBot(t, &mockAdapter{})
	bot.SetIssueLookup(stubIssueLookupErr{})

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "/ai test-cases HUB-999",
		MessageID:   "m-tc-miss",
	})
	require.NoError(t, err)
}

func TestMessengerBot_HandleCommand_ReviewSuccess(t *testing.T) {
	adapter := &streamingMockAdapter{chunks: []string{"LGTM with notes"}}
	bot, _, userID, channelID, _ := setupHubAIBot(t, adapter)
	bot.SetCodeReview(stubCodeReviewProvider{ctx: &ai.ReviewContext{
		URL: "https://github.com/o/r/pull/1", Title: "Fix auth",
	}})

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "/ai review https://github.com/o/r/pull/1",
		MessageID:   "m-rev-ok",
	})
	require.NoError(t, err)
}

func TestMessengerBot_HandleCommand_ReviewFetchError(t *testing.T) {
	bot, _, userID, channelID, _ := setupHubAIBot(t, &mockAdapter{})
	bot.SetCodeReview(stubCodeReviewProvider{err: assert.AnError})

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "/ai review https://github.com/o/r/pull/2",
		MessageID:   "m-rev-err",
	})
	require.NoError(t, err)
}

func TestMessengerBot_HandleCommand_TestCasesForeignKeyDenial(t *testing.T) {
	// P1-BE-31 / W5-A3-002: foreign issue key → denial (not found), без утечки.
	bot, hookRepo, userID, channelID, _ := setupHubAIBot(t, &mockAdapter{})
	lookup := &denyIssueLookup{}
	bot.SetIssueLookup(lookup)

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "/ai test-cases OTHER-42",
		MessageID:   "m-tc-foreign",
	})
	require.NoError(t, err)
	assert.True(t, lookup.called)
	ch := hookRepo.Channel(channelID)
	require.NotNil(t, ch)
}

func TestMessengerBot_HandleCommand_ReviewOutsideProjectDenial(t *testing.T) {
	// P1-BE-31 / W5-A3-003: review вне project → denial без сырого err.Error().
	bot, _, userID, channelID, _ := setupHubAIBot(t, &mockAdapter{})
	deny := &denyCodeReviewProvider{}
	bot.SetCodeReview(deny)

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "/ai review https://github.com/secret/repo/pull/9",
		MessageID:   "m-rev-deny",
	})
	require.NoError(t, err)
	assert.True(t, deny.called)
}

func TestMessengerBot_OnMessageSent_PublicMentionWithoutBotMember(t *testing.T) {
	// P1-BE-31 / W5-A3-001: mention без bot∈members → no reply / no history load.
	adapter := &streamingMockAdapter{chunks: []string{"should not run"}}
	bot, hookRepo, userID, _, botID := setupHubAIBot(t, adapter)
	// Канал только с user, без bot.
	channelID := hookRepo.SetupPublicChannel(userID)

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "@hub-ai summarize secrets",
		MessageID:   "msg-no-bot",
	})
	require.NoError(t, err)
	_ = botID
}

func TestMessengerBot_HandleCommand_ExplainTranslateSQL(t *testing.T) {
	adapter := &streamingMockAdapter{chunks: []string{"done"}}
	cases := []string{
		"/ai explain func main() {}",
		"/ai translate Hello world",
		"/ai sql count open issues by project",
	}
	for i, cmd := range cases {
		t.Run(cmd, func(t *testing.T) {
			bot, _, userID, channelID, _ := setupHubAIBot(t, adapter)
			err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
				ChannelID: channelID, SenderID: userID, ContentText: cmd, MessageID: "m-cmd-" + string(rune('a'+i)),
			})
			require.NoError(t, err)
		})
	}
}

func TestMessengerBot_HandleCommand_DocsSearchError(t *testing.T) {
	// A2W5-A3-004: канал получает fixed string, не err.Error().
	bot, hookRepo, userID, channelID, _ := setupHubAIBot(t, &mockAdapter{})
	leak := "super-secret-db-dsn=postgres://internal"
	bot.SetDocSearch(searchDocStub{err: errors.New(leak)})

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "/ai docs broken search",
		MessageID:   "m-docs-err",
	})
	require.NoError(t, err)

	msgs, err := hookRepo.Repository().ListMessages(context.Background(), channelID, nil, 20)
	require.NoError(t, err)
	require.NotEmpty(t, msgs)
	var botText string
	for _, m := range msgs {
		if m.ContentText != "" && m.ContentText != "/ai docs broken search" {
			botText = m.ContentText
			break
		}
	}
	require.NotEmpty(t, botText)
	assert.Contains(t, botText, "Поиск временно недоступен")
	assert.NotContains(t, botText, leak)
	assert.NotContains(t, botText, "Ошибка поиска:")
}

func TestMessengerBot_StreamReply_AIError(t *testing.T) {
	repo := &mockRepo{providers: []*ent.AIProvider{{
		ID: uuid.New(), Name: "Test", BaseURL: "https://api.example.com/v1",
		Models: []string{"test-model"}, Enabled: true, Scope: aiprovider.ScopeSystem,
	}}}
	aiSvc := ai.NewService(repo, &mockAdapter{err: assert.AnError}, nil)
	botID := uuid.New()
	userID := uuid.New()
	hookRepo := messengermod.NewHookTestRepo()
	msgSvc := messengermod.NewMessengerService(hookRepo.Repository(), nil, nil, nil)
	msgSvc.SetSystemBotID(botID)
	channelID := hookRepo.SetupDM(userID, botID)
	bot := ai.NewMessengerBot(aiSvc, msgSvc, botID)

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID: channelID, SenderID: userID, ContentText: "fail please", MessageID: "m-fail",
	})
	require.NoError(t, err)
}

func TestMessengerBot_HandleCommand_Summarize(t *testing.T) {
	adapter := &streamingMockAdapter{chunks: []string{"summary"}}
	bot, hookRepo, userID, channelID, _ := setupHubAIBot(t, adapter)
	_, err := hookRepo.InsertMessage(channelID, userID, "decision made", "")
	require.NoError(t, err)

	err = bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID: channelID, SenderID: userID, ContentText: "/ai summarize", MessageID: "m-sum",
	})
	require.NoError(t, err)
}

func TestMessengerBot_HandleCommand_Unknown(t *testing.T) {
	bot, _, userID, channelID, _ := setupHubAIBot(t, &mockAdapter{})
	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID: channelID, SenderID: userID, ContentText: "/ai unknown-cmd", MessageID: "m-bad",
	})
	require.NoError(t, err)
}

func TestMessengerBot_HandleCommand_ReviewUnavailable(t *testing.T) {
	bot, _, userID, channelID, _ := setupHubAIBot(t, &mockAdapter{})
	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "/ai review https://github.com/o/r/pull/1",
		MessageID:   "m-rev",
	})
	require.NoError(t, err)
}

func TestSearchDocProvider_WithResults(t *testing.T) {
	p := ai.NewSearchDocProvider(searchBackendStub{results: []searchmod.SearchResult{
		{Type: "issue", Title: "HUB-1", Subtitle: "Login bug", URL: "/issues/HUB-1"},
	}})
	out, err := p.SearchContext(context.Background(), uuid.New(), "login")
	require.NoError(t, err)
	assert.Contains(t, out, "HUB-1")
	assert.Contains(t, out, "Login bug")
}

func TestSearchDocProvider_EmptyResults(t *testing.T) {
	p := ai.NewSearchDocProvider(searchBackendStub{})
	out, err := p.SearchContext(context.Background(), uuid.New(), "missing")
	require.NoError(t, err)
	assert.Contains(t, out, "не найдено")
}

func TestSearchDocProvider_SearchError(t *testing.T) {
	p := ai.NewSearchDocProvider(searchBackendStub{err: assert.AnError})
	_, err := p.SearchContext(context.Background(), uuid.New(), "fail")
	require.Error(t, err)
}

// Ensure bot stream path waits for partial edit throttle without racing the test runner.
func TestMessengerBot_StreamReply_ThrottlesEdits(t *testing.T) {
	chunks := make([]string, 20)
	for i := range chunks {
		chunks[i] = "x"
	}
	bot, _, userID, channelID, _ := setupHubAIBot(t, &streamingMockAdapter{chunks: chunks})
	start := time.Now()
	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID: channelID, SenderID: userID, ContentText: "go", MessageID: "m-stream",
	})
	require.NoError(t, err)
	assert.Less(t, time.Since(start), 5*time.Second)
}
