package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBotCommand(t *testing.T) {
	cmd := ai.ExportParseBotCommand("/ai summarize")
	assert.NotNil(t, cmd)
	assert.Equal(t, "summarize", ai.ExportBotCommandName(cmd))

	cmd = ai.ExportParseBotCommand("@hub-ai /ai test-cases hub-42")
	assert.Nil(t, cmd)

	cmd = ai.ExportParseBotCommand("/ai test-cases HUB-123 extra")
	assert.Equal(t, "test-cases", ai.ExportBotCommandName(cmd))
	assert.Equal(t, "HUB-123 extra", ai.ExportBotCommandArgs(cmd))

	cmd = ai.ExportParseBotCommand("/ai explain func main() {}")
	assert.Equal(t, "explain", ai.ExportBotCommandName(cmd))
	assert.Contains(t, ai.ExportBotCommandArgs(cmd), "func main")

	assert.Nil(t, ai.ExportParseBotCommand("hello world"))
}

func TestExtractIssueKey(t *testing.T) {
	assert.Equal(t, "HUB-123", ai.ExportExtractIssueKey("hub-123"))
	assert.Equal(t, "PROJ-1", ai.ExportExtractIssueKey("PROJ-1 details"))
	assert.Equal(t, "", ai.ExportExtractIssueKey("no key here"))
}

func TestBuildSummarizePrompt(t *testing.T) {
	_, user := ai.ExportBuildSummarizePrompt([]string{"decided to ship", "need tests"})
	assert.Contains(t, user, "decided to ship")
}

func TestBuildTestCasesPrompt(t *testing.T) {
	_, user := ai.ExportBuildTestCasesPrompt("HUB-1", "Login", "User logs in via magic link")
	assert.Contains(t, user, "HUB-1")
	assert.Contains(t, user, "magic link")
}

func TestBuildExplainPrompt(t *testing.T) {
	_, user := ai.ExportBuildExplainPrompt("func main() {}")
	assert.Contains(t, user, "func main")

	_, empty := ai.ExportBuildExplainPrompt("")
	assert.Contains(t, empty, "код не указан")
}

func TestBuildTranslatePrompt(t *testing.T) {
	_, user := ai.ExportBuildTranslatePrompt("Hello world")
	assert.Equal(t, "Hello world", user)
}

func TestBuildDocsPrompt(t *testing.T) {
	_, user := ai.ExportBuildDocsPrompt("issue HUB-1", "What is the status?")
	assert.Contains(t, user, "HUB-1")
}

func TestBuildSQLPrompt(t *testing.T) {
	_, user := ai.ExportBuildSQLPrompt("count open issues")
	assert.Contains(t, user, "count open issues")
}

func TestBuildHelpPrompt(t *testing.T) {
	_, user := ai.ExportBuildHelpPrompt()
	assert.Contains(t, user, "/ai summarize")
}
