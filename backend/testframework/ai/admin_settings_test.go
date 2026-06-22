package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
)

type memSettings struct {
	data map[string]string
}

func (m *memSettings) Get(_ context.Context, key string) (string, error) {
	return m.data[key], nil
}

func (m *memSettings) Upsert(_ context.Context, key, value, _ string) error {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = value
	return nil
}

func TestBotSystemPromptDefault(t *testing.T) {
	svc := ai.NewService(&mockRepo{}, &mockAdapter{}, nil)
	assert.Equal(t, ai.DefaultBotSystemPrompt, svc.BotSystemPrompt(context.Background()))
}

func TestBotSystemPromptCustom(t *testing.T) {
	store := &memSettings{data: map[string]string{ai.ExportSettingBotSystemPrompt(): "Custom prompt"}}
	svc := ai.NewService(&mockRepo{}, &mockAdapter{}, nil)
	svc.SetSettingsStore(store)
	assert.Equal(t, "Custom prompt", svc.BotSystemPrompt(context.Background()))
}

func TestUpdateAdminSettings(t *testing.T) {
	store := &memSettings{data: map[string]string{}}
	svc := ai.NewService(&mockRepo{}, &mockAdapter{}, nil)
	svc.SetSettingsStore(store)
	prompt := "Review code carefully"
	resp, err := svc.UpdateAdminSettings(context.Background(), ai.UpdateAdminSettingsRequest{
		SystemPrompt: &prompt,
	})
	require.NoError(t, err)
	assert.Equal(t, prompt, resp.SystemPrompt)
	assert.Equal(t, prompt, svc.BotSystemPrompt(context.Background()))
}

func TestOfflineOnlyFiltersCloudProviders(t *testing.T) {
	repo := &mockRepo{providers: []*ent.AIProvider{
		{
			ID: uuid.New(), Name: "Cloud", BaseURL: "https://api.deepseek.com/v1",
			Models: []string{"deepseek-chat"}, Enabled: true, Scope: aiprovider.ScopeSystem,
		},
		{
			ID: uuid.New(), Name: "Ollama", BaseURL: "http://127.0.0.1:11434/v1",
			ProviderType: aiprovider.ProviderTypeOllama,
			Models: []string{"deepseek-chat"}, Enabled: true, Scope: aiprovider.ScopeSystem, Priority: 1,
		},
	}}
	svc := ai.NewService(repo, &mockAdapter{resp: &ai.ChatCompletionResponse{}}, nil)
	svc.SetOfflineOnly(true)
	providers, err := ai.ExportServiceResolveProviders(svc, context.Background(), uuid.New(), "deepseek-chat")
	require.NoError(t, err)
	require.Len(t, providers, 1)
	assert.Equal(t, "Ollama", providers[0].Name)
}
