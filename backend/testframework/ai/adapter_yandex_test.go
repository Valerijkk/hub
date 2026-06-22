package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseYandexCredentials(t *testing.T) {
	folder, key, ok := ai.ExportParseYandexCredentials("folder123:secret-key")
	assert.True(t, ok)
	assert.Equal(t, "folder123", folder)
	assert.Equal(t, "secret-key", key)

	folder, key, ok = ai.ExportParseYandexCredentials("only-key")
	assert.True(t, ok)
	assert.Equal(t, "", folder)
	assert.Equal(t, "only-key", key)

	_, _, ok = ai.ExportParseYandexCredentials("")
	assert.False(t, ok)
}

func TestIsYandexProvider(t *testing.T) {
	assert.True(t, ai.IsYandexProvider("https://llm.api.cloud.yandex.net"))
	assert.True(t, ai.IsYandexProvider("yandexgpt"))
	assert.False(t, ai.IsYandexProvider("https://api.openai.com/v1"))
}
