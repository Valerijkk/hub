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

func TestFilterProvidersByConsentBlocksCloud(t *testing.T) {
	cloud := &ent.AIProvider{
		Name:         "DeepSeek",
		BaseURL:      "https://api.deepseek.com/v1",
		ProviderType: aiprovider.ProviderTypeOpenaiCompatible,
	}
	local := &ent.AIProvider{
		Name:         "Ollama",
		BaseURL:      "http://127.0.0.1:11434/v1",
		ProviderType: aiprovider.ProviderTypeOllama,
	}
	filtered := ai.ExportFilterProvidersByConsent([]*ent.AIProvider{cloud, local}, false)
	require.Len(t, filtered, 1)
	assert.Equal(t, "Ollama", filtered[0].Name)
}

func TestResolveProvidersConsentRequired(t *testing.T) {
	userID := uuid.New()
	repo := &mockRepo{
		providers: []*ent.AIProvider{{
			ID: uuid.New(), Name: "Cloud", BaseURL: "https://api.deepseek.com/v1",
			Models: []string{"deepseek-chat"}, Enabled: true, Scope: aiprovider.ScopeSystem,
		}},
		allowExternal: map[uuid.UUID]bool{userID: false},
	}
	svc := ai.NewService(repo, &mockAdapter{}, nil)
	_, err := ai.ExportServiceResolveProviders(svc, context.Background(), userID, "deepseek-chat")
	require.Error(t, err)
	assert.ErrorIs(t, err, ai.ErrExternalAIConsentRequired)
}

func TestSaveUserPreferenceAllowExternal(t *testing.T) {
	userID := uuid.New()
	repo := &mockRepo{}
	svc := ai.NewService(repo, &mockAdapter{}, nil)
	falseVal := false
	resp, err := svc.SaveUserPreference(context.Background(), userID, ai.UpdateUserPreferenceRequest{
		AllowExternalAI: &falseVal,
	})
	require.NoError(t, err)
	assert.False(t, resp.AllowExternalAI)
}

func TestConsentErrorForModel(t *testing.T) {
	require.NoError(t, ai.ExportConsentErrorForModel("", true, nil))
	require.NoError(t, ai.ExportConsentErrorForModel("gpt-4", false, []*ent.AIProvider{{Name: "local"}}))

	err := ai.ExportConsentErrorForModel("gpt-4", false, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ai.ErrExternalAIConsentRequired)
	assert.Contains(t, err.Error(), "gpt-4")

	err = ai.ExportConsentErrorForModel("", false, nil)
	require.ErrorIs(t, err, ai.ErrExternalAIConsentRequired)
}
