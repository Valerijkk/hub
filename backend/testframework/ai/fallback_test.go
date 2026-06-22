package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
)

func TestIsRetryableProviderError(t *testing.T) {
	assert.True(t, ai.ExportIsRetryableProviderError(fmt.Errorf("yandexgpt HTTP 503: unavailable")))
	assert.True(t, ai.ExportIsRetryableProviderError(ai.ErrProviderRateLimited))
	assert.True(t, ai.ExportIsRetryableProviderError(fmt.Errorf("dial tcp: connection refused")))
	assert.False(t, ai.ExportIsRetryableProviderError(fmt.Errorf("yandexgpt HTTP 401: unauthorized")))
	assert.False(t, ai.ExportIsRetryableProviderError(ai.ErrContentPolicyViolation))
	assert.False(t, ai.ExportIsRetryableProviderError(nil))
}

func TestIsOfflineProvider(t *testing.T) {
	assert.True(t, ai.ExportIsOfflineProvider(&ent.AIProvider{
		ProviderType: aiprovider.ProviderTypeOllama,
		BaseURL:      "http://127.0.0.1:11434/v1",
	}))
	assert.True(t, ai.ExportIsOfflineProvider(&ent.AIProvider{
		ProviderType: aiprovider.ProviderTypeOpenaiCompatible,
		BaseURL:      "http://localhost:11434/v1",
	}))
	assert.False(t, ai.ExportIsOfflineProvider(&ent.AIProvider{
		ProviderType: aiprovider.ProviderTypeOpenaiCompatible,
		BaseURL:      "https://api.deepseek.com/v1",
	}))
}

func TestErrAllProvidersFailedWrap(t *testing.T) {
	err := fmt.Errorf("%w: timeout", ai.ErrAllProvidersFailed)
	assert.True(t, errors.Is(err, ai.ErrAllProvidersFailed))
}
