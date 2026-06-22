package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitAnthropicMessages(t *testing.T) {
	sys, msgs := ai.ExportSplitAnthropicMessages([]ai.ChatMessage{
		{Role: "system", Content: "Be helpful"},
		{Role: "user", Content: "Hi"},
		{Role: "assistant", Content: "Hello"},
		{Role: "user", Content: "Bye"},
	})
	assert.Equal(t, "Be helpful", sys)
	assert.Len(t, msgs, 3)
	assert.Equal(t, "user", msgs[0].Role)
	assert.Equal(t, "assistant", msgs[1].Role)
}

func TestAnthropicAdapter_RequiresAPIKey(t *testing.T) {
	adapter := ai.NewAnthropicAdapter()
	_, err := adapter.ChatCompletion(t.Context(), "https://api.anthropic.com", "", "claude-3-5-haiku-20241022", ai.ChatCompletionRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi"}},
	})
	assert.Error(t, err)
}

func TestExtractPRURL(t *testing.T) {
	assert.Contains(t, ai.ExportExtractPRURL("see https://github.com/org/repo/pull/9 please"), "github.com")
	assert.Empty(t, ai.ExportExtractPRURL("no url"))
}

func TestParseBotCommand_Help(t *testing.T) {
	cmd := ai.ExportParseBotCommand("/ai help")
	assert.NotNil(t, cmd)
	assert.Equal(t, "help", ai.ExportBotCommandName(cmd))
}

func TestBuildReviewPrompt(t *testing.T) {
	_, user := ai.ExportBuildReviewPrompt(nil)
	assert.NotEmpty(t, user)
}
