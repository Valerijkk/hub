package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIAdapter_ChatCompletion_StreamRejected(t *testing.T) {
	a := ai.NewOpenAIAdapter()
	_, err := a.ChatCompletion(context.Background(), "https://api.example.com/v1", "key", "m", ai.ChatCompletionRequest{
		Stream: true,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stream_id")
}

func TestOpenAIAdapter_ChatCompletion_InvalidURL(t *testing.T) {
	a := ai.NewOpenAIAdapter()
	_, err := a.ChatCompletion(context.Background(), "://bad", "key", "m", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid provider URL")
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "abc", ai.ExportTruncate("abc", 10))
	long := strings.Repeat("x", 600)
	out := ai.ExportTruncate(long, 512)
	assert.Len(t, out, 515)
	assert.True(t, strings.HasSuffix(out, "..."))
}
