package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
)

type mockRepo struct {
	providers       []*ent.AIProvider
	ownerByProv     map[uuid.UUID]uuid.UUID
	prefModel       map[uuid.UUID]string
	allowExternal   map[uuid.UUID]bool
	usageIn         int
	usageOut        int
	requestCount    int
}

func (m *mockRepo) ListProviders(_ context.Context, scope string) ([]*ent.AIProvider, error) {
	var out []*ent.AIProvider
	for _, p := range m.providers {
		if scope == "" || string(p.Scope) == scope {
			out = append(out, p)
		}
	}
	return out, nil
}

func (m *mockRepo) GetProvider(_ context.Context, id uuid.UUID) (*ent.AIProvider, error) {
	for _, p := range m.providers {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, &ent.NotFoundError{}
}

func (m *mockRepo) CreateProvider(_ context.Context, input ai.CreateProviderInput) (*ent.AIProvider, error) {
	scope := aiprovider.ScopeSystem
	if input.Scope == "user" {
		scope = aiprovider.ScopeUser
	}
	p := &ent.AIProvider{
		ID:           uuid.New(),
		Name:         input.Name,
		ProviderType: aiprovider.ProviderType(input.ProviderType),
		BaseURL:      input.BaseURL,
		Models:       input.Models,
		Enabled:      input.Enabled,
		Scope:        scope,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if input.OwnerUserID != nil {
		if m.ownerByProv == nil {
			m.ownerByProv = make(map[uuid.UUID]uuid.UUID)
		}
		m.ownerByProv[p.ID] = *input.OwnerUserID
	}
	if input.APIKey != "" {
		p.APIKeyEnc = input.APIKey
	}
	m.providers = append(m.providers, p)
	return p, nil
}

func (m *mockRepo) UpdateProvider(_ context.Context, id uuid.UUID, input ai.UpdateProviderInput) (*ent.AIProvider, error) {
	p, err := m.GetProvider(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		p.Name = *input.Name
	}
	if input.Enabled != nil {
		p.Enabled = *input.Enabled
	}
	return p, nil
}

func (m *mockRepo) FindProviderForModel(_ context.Context, userID uuid.UUID, model string) (*ent.AIProvider, error) {
	list, err := m.ListProvidersForModel(context.Background(), userID, model)
	if err != nil {
		return nil, err
	}
	return list[0], nil
}

func (m *mockRepo) ListProvidersForModel(_ context.Context, userID uuid.UUID, model string) ([]*ent.AIProvider, error) {
	match := func(p *ent.AIProvider) bool {
		if !p.Enabled {
			return false
		}
		for _, mdl := range p.Models {
			if mdl == model {
				return true
			}
		}
		return false
	}
	var out []*ent.AIProvider
	for _, p := range m.providers {
		if p.Scope == aiprovider.ScopeUser && m.ownerByProv[p.ID] == userID && match(p) {
			out = append(out, p)
		}
	}
	for _, p := range m.providers {
		scope := p.Scope
		if scope == "" {
			scope = aiprovider.ScopeSystem
		}
		if scope == aiprovider.ScopeSystem && match(p) {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil, &ent.NotFoundError{}
	}
	return out, nil
}

func (m *mockRepo) ListUserProviders(_ context.Context, userID uuid.UUID) ([]*ent.AIProvider, error) {
	var out []*ent.AIProvider
	for _, p := range m.providers {
		if p.Scope == aiprovider.ScopeUser && m.ownerByProv[p.ID] == userID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (m *mockRepo) GetUserProvider(_ context.Context, userID, providerID uuid.UUID) (*ent.AIProvider, error) {
	p, err := m.GetProvider(context.Background(), providerID)
	if err != nil {
		return nil, err
	}
	if p.Scope != aiprovider.ScopeUser || m.ownerByProv[p.ID] != userID {
		return nil, &ent.NotFoundError{}
	}
	return p, nil
}

func (m *mockRepo) GetDailyUsage(_ context.Context, _ uuid.UUID) (int, int, int, error) {
	return m.usageIn, m.usageOut, m.requestCount, nil
}

func (m *mockRepo) GetUserPreference(_ context.Context, userID uuid.UUID) (string, *uuid.UUID, error) {
	rec, err := m.GetUserPreferenceRecord(context.Background(), userID)
	if err != nil || rec == nil {
		return "", nil, err
	}
	return rec.DefaultModel, rec.DefaultProviderID, nil
}

func (m *mockRepo) GetUserPreferenceRecord(_ context.Context, userID uuid.UUID) (*ai.UserPreferenceRecord, error) {
	if m.prefModel == nil && m.allowExternal == nil {
		return nil, nil
	}
	model := ""
	if m.prefModel != nil {
		model = m.prefModel[userID]
	}
	allow := true
	if m.allowExternal != nil {
		if v, ok := m.allowExternal[userID]; ok {
			allow = v
		}
	}
	return &ai.UserPreferenceRecord{
		DefaultModel:    model,
		AllowExternalAI: allow,
	}, nil
}

func (m *mockRepo) SaveUserPreference(_ context.Context, userID uuid.UUID, input ai.SaveUserPreferenceInput) error {
	if input.DefaultModel != nil {
		if m.prefModel == nil {
			m.prefModel = make(map[uuid.UUID]string)
		}
		m.prefModel[userID] = *input.DefaultModel
	}
	if input.AllowExternalAI != nil {
		if m.allowExternal == nil {
			m.allowExternal = make(map[uuid.UUID]bool)
		}
		m.allowExternal[userID] = *input.AllowExternalAI
	}
	return nil
}

func (m *mockRepo) GetUsageSummary(_ context.Context, _ time.Time) (int, int, int, int, error) {
	return m.requestCount, m.usageIn, m.usageOut, 1, nil
}

func (m *mockRepo) ListUsageHistory(_ context.Context, days int) ([]ai.UsageHistoryDay, error) {
	if days <= 0 {
		days = 7
	}
	day := time.Now().UTC().Format("2006-01-02")
	return []ai.UsageHistoryDay{{
		Date:        day,
		Requests:    m.requestCount,
		TokensIn:    m.usageIn,
		TokensOut:   m.usageOut,
		ActiveUsers: 1,
	}}, nil
}

func (m *mockRepo) IncrementUsage(_ context.Context, _ uuid.UUID, tokensIn, tokensOut int) error {
	m.usageIn += tokensIn
	m.usageOut += tokensOut
	m.requestCount++
	return nil
}

func (m *mockRepo) ListModels(_ context.Context, _ *uuid.UUID) ([]ai.ModelInfo, error) {
	var out []ai.ModelInfo
	for _, p := range m.providers {
		if !p.Enabled {
			continue
		}
		for _, mdl := range p.Models {
			out = append(out, ai.ModelInfo{ID: mdl, ProviderID: p.ID.String(), ProviderName: p.Name})
		}
	}
	return out, nil
}

type mockAdapter struct {
	resp      *ai.ChatCompletionResponse
	err       error
	calls     int
	failUntil int
}

func (a *mockAdapter) ChatCompletion(_ context.Context, _, _, _ string, _ ai.ChatCompletionRequest) (*ai.ChatCompletionResponse, error) {
	a.calls++
	if a.failUntil > 0 && a.calls <= a.failUntil {
		return nil, fmt.Errorf("HTTP 503 service unavailable")
	}
	if a.err != nil {
		return nil, a.err
	}
	return a.resp, nil
}

func (a *mockAdapter) ChatCompletionStream(_ context.Context, _, _, _ string, _ ai.ChatCompletionRequest, onDelta func(ai.StreamDelta) error) error {
	if a.err != nil {
		return a.err
	}
	if a.resp != nil && len(a.resp.Choices) > 0 {
		content := a.resp.Choices[0].Message.Content
		if content != "" {
			if err := onDelta(ai.StreamDelta{Content: content}); err != nil {
				return err
			}
		}
		return onDelta(ai.StreamDelta{Done: true})
	}
	return onDelta(ai.StreamDelta{Done: true})
}

func TestChatCompletion_IncrementsUsage(t *testing.T) {
	pid := uuid.New()
	repo := &mockRepo{providers: []*ent.AIProvider{{
		ID: pid, Name: "Ollama", BaseURL: "http://127.0.0.1:11434/v1",
		Models: []string{"llama3"}, Enabled: true, Scope: aiprovider.ScopeSystem,
	}}}
	adapter := &mockAdapter{resp: &ai.ChatCompletionResponse{}}
	adapter.resp.Usage.PromptTokens = 10
	adapter.resp.Usage.CompletionTokens = 5
	svc := ai.NewService(repo, adapter, nil)
	userID := uuid.New()
	resp, err := svc.ChatCompletion(context.Background(), userID, ai.ChatCompletionRequest{
		Model:    "llama3",
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	}, "", "")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 10, repo.usageIn)
	assert.Equal(t, 5, repo.usageOut)
}

func TestChatCompletion_FallbackOn503(t *testing.T) {
	repo := &mockRepo{providers: []*ent.AIProvider{
		{
			ID: uuid.New(), Name: "Primary", BaseURL: "https://api.deepseek.com/v1",
			Models: []string{"deepseek-chat"}, Enabled: true, Scope: aiprovider.ScopeSystem, Priority: 0,
		},
		{
			ID: uuid.New(), Name: "Backup", BaseURL: "https://openrouter.ai/api/v1",
			Models: []string{"deepseek-chat"}, Enabled: true, Scope: aiprovider.ScopeSystem, Priority: 1,
		},
	}}
	adapter := &mockAdapter{failUntil: 1, resp: &ai.ChatCompletionResponse{}}
	adapter.resp.Choices = []struct {
		Index        int         `json:"index"`
		Message      ai.ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	}{{Message: ai.ChatMessage{Role: "assistant", Content: "ok"}}}
	svc := ai.NewService(repo, adapter, nil)
	resp, err := svc.ChatCompletion(context.Background(), uuid.New(), ai.ChatCompletionRequest{
		Model:    "deepseek-chat",
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	}, "", "")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 2, adapter.calls)
}

func TestChatCompletion_NoProvider(t *testing.T) {
	svc := ai.NewService(&mockRepo{}, &mockAdapter{}, nil)
	_, err := svc.ChatCompletion(context.Background(), uuid.New(), ai.ChatCompletionRequest{
		Model:    "unknown",
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	}, "", "")
	require.ErrorIs(t, err, ai.ErrNoProvider)
}

func TestNormalizeModels(t *testing.T) {
	out := ai.ExportNormalizeModels([]string{" b ", "a", "a", ""})
	assert.Equal(t, []string{"a", "b"}, out)
}

type mockStreamPublisher struct {
	mu     sync.Mutex
	events []streamEvent
}

type streamEvent struct {
	userID   uuid.UUID
	streamID string
	event    string
	payload  map[string]interface{}
}

func (m *mockStreamPublisher) PublishStreamEvent(userID uuid.UUID, streamID, event string, payload map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, streamEvent{userID, streamID, event, payload})
}

func (m *mockStreamPublisher) eventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.events)
}

func (m *mockStreamPublisher) eventAt(i int) (streamEvent, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if i < 0 || i >= len(m.events) {
		return streamEvent{}, false
	}
	return m.events[i], true
}

func TestStartChatStream_PublishesEvents(t *testing.T) {
	pid := uuid.New()
	repo := &mockRepo{providers: []*ent.AIProvider{{
		ID: pid, Name: "Ollama", BaseURL: "http://127.0.0.1:11434/v1",
		Models: []string{"llama3"}, Enabled: true, Scope: aiprovider.ScopeSystem,
	}}}
	adapter := &mockAdapter{resp: &ai.ChatCompletionResponse{
		Choices: []struct {
			Index        int         `json:"index"`
			Message      ai.ChatMessage `json:"message"`
			FinishReason string      `json:"finish_reason"`
		}{{Message: ai.ChatMessage{Role: "assistant", Content: "hello"}}},
	}}
	svc := ai.NewService(repo, adapter, nil)
	streams := &mockStreamPublisher{}
	svc.SetStreamPublisher(streams)

	userID := uuid.New()
	err := svc.StartChatStream(context.Background(), userID, ai.ChatCompletionRequest{
		Model:    "llama3",
		StreamID: "stream-1",
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	}, "127.0.0.1", "test")
	require.NoError(t, err)

	deadline := time.Now().Add(2 * time.Second)
	for streams.eventCount() < 2 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	require.GreaterOrEqual(t, streams.eventCount(), 2)
	first, ok := streams.eventAt(0)
	require.True(t, ok)
	assert.Equal(t, "start", first.event)
	last, ok := streams.eventAt(streams.eventCount() - 1)
	require.True(t, ok)
	assert.Equal(t, "done", last.event)
	assert.Equal(t, "hello", last.payload["content"])
}

func TestStartChatStream_RequiresPublisher(t *testing.T) {
	svc := ai.NewService(&mockRepo{}, &mockAdapter{}, nil)
	err := svc.StartChatStream(context.Background(), uuid.New(), ai.ChatCompletionRequest{
		Model:    "m",
		StreamID: "s",
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	}, "", "")
	require.Error(t, err)
}

func TestCreateUserProvider(t *testing.T) {
	svc := ai.NewService(&mockRepo{}, &mockAdapter{}, nil)
	userID := uuid.New()
	resp, err := svc.CreateUserProvider(context.Background(), userID, ai.CreateProviderRequest{
		Name:    "My DeepSeek",
		BaseURL: "https://api.deepseek.com/v1",
		Models:  []string{"deepseek-chat"},
		APIKey:  "sk-test",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "user", resp.Scope)
	assert.True(t, resp.HasAPIKey)

	items, err := svc.ListUserProviders(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestAssistJQL_ValidQuery(t *testing.T) {
	repo := &mockRepo{providers: []*ent.AIProvider{{
		ID: uuid.New(), Name: "Ollama", BaseURL: "http://127.0.0.1:11434/v1",
		Models: []string{"llama3"}, Enabled: true, Scope: aiprovider.ScopeSystem,
	}}}
	adapter := &mockAdapter{resp: &ai.ChatCompletionResponse{
		Choices: []struct {
			Index        int         `json:"index"`
			Message      ai.ChatMessage `json:"message"`
			FinishReason string      `json:"finish_reason"`
		}{{Message: ai.ChatMessage{Role: "assistant", Content: "priority = high"}}},
	}}
	svc := ai.NewService(repo, adapter, nil)
	resp, err := svc.AssistJQL(context.Background(), uuid.New(), ai.JQLAssistRequest{
		Query: "покажи высокий приоритет",
	})
	require.NoError(t, err)
	assert.Equal(t, "priority = high", resp.JQL)
	assert.Empty(t, resp.Explanation)
}

func TestAssistJQL_InvalidJQL_ReturnsWarning(t *testing.T) {
	repo := &mockRepo{providers: []*ent.AIProvider{{
		ID: uuid.New(), Name: "Ollama", BaseURL: "http://127.0.0.1:11434/v1",
		Models: []string{"llama3"}, Enabled: true, Scope: aiprovider.ScopeSystem,
	}}}
	adapter := &mockAdapter{resp: &ai.ChatCompletionResponse{
		Choices: []struct {
			Index        int         `json:"index"`
			Message      ai.ChatMessage `json:"message"`
			FinishReason string      `json:"finish_reason"`
		}{{Message: ai.ChatMessage{Role: "assistant", Content: "not valid jql!!!"}}},
	}}
	svc := ai.NewService(repo, adapter, nil)
	resp, err := svc.AssistJQL(context.Background(), uuid.New(), ai.JQLAssistRequest{Query: "bad"})
	require.NoError(t, err)
	assert.Contains(t, resp.Explanation, "warning")
}

func TestGetUsageToday(t *testing.T) {
	repo := &mockRepo{usageIn: 100, usageOut: 50, requestCount: 3}
	svc := ai.NewService(repo, &mockAdapter{}, nil)
	resp, err := svc.GetUsageToday(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, 100, resp.TokensIn)
	assert.Equal(t, 50, resp.TokensOut)
	assert.Equal(t, 3, resp.RequestCount)
	assert.Equal(t, ai.DefaultDailyRequestLimit, resp.DailyLimit)
}

func TestStreamChatCompletion_AccumulatesText(t *testing.T) {
	pid := uuid.New()
	repo := &mockRepo{providers: []*ent.AIProvider{{
		ID: pid, Name: "Ollama", BaseURL: "http://127.0.0.1:11434/v1",
		Models: []string{"llama3"}, Enabled: true, Scope: aiprovider.ScopeSystem,
	}}}
	adapter := &streamingMockAdapter{chunks: []string{"hel", "lo"}}
	svc := ai.NewService(repo, adapter, nil)
	out, err := svc.StreamChatCompletion(context.Background(), uuid.New(), ai.ChatCompletionRequest{
		Model:    "llama3",
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	}, "", "", nil)
	require.NoError(t, err)
	assert.Equal(t, "hello", out)
}

type streamingMockAdapter struct {
	chunks []string
}

func (a *streamingMockAdapter) ChatCompletion(_ context.Context, _, _, _ string, _ ai.ChatCompletionRequest) (*ai.ChatCompletionResponse, error) {
	return &ai.ChatCompletionResponse{}, nil
}

func (a *streamingMockAdapter) ChatCompletionStream(_ context.Context, _, _, _ string, _ ai.ChatCompletionRequest, onDelta func(ai.StreamDelta) error) error {
	for _, c := range a.chunks {
		if err := onDelta(ai.StreamDelta{Content: c}); err != nil {
			return err
		}
	}
	return onDelta(ai.StreamDelta{Done: true})
}

func TestGigaChatAdapter_RequiresCredentials(t *testing.T) {
	adapter := ai.NewGigaChatAdapter()
	_, err := adapter.ChatCompletion(context.Background(), "https://gigachat.devices.sberbank.ru/api/v1", "", "GigaChat", ai.ChatCompletionRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client_id:client_secret")
}
