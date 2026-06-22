package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	messengermod "github.com/valerii/hub/backend/internal/messenger"
)

func TestExtractUserPrompt(t *testing.T) {
	assert.Equal(t, "help me", ai.ExportExtractUserPrompt("@hub-ai help me"))
	assert.Equal(t, "help", ai.ExportExtractUserPrompt("@HUB-AI help"))
	assert.Equal(t, "", ai.ExportExtractUserPrompt("@hub-ai"))
}

func TestMessengerBot_shouldRespond(t *testing.T) {
	botID := uuid.New()
	b := ai.ExportNewMessengerBotForShouldRespond(botID)

	dmChannel := &ent.Channel{ChannelType: channel.ChannelTypeDm}
	groupChannel := &ent.Channel{ChannelType: channel.ChannelTypePublic}

	assert.True(t, ai.ExportMessengerBotShouldRespond(b, messengermod.MessageHookEvent{
		ContentText: "hello",
	}, dmChannel, []uuid.UUID{uuid.New(), botID}))

	assert.True(t, ai.ExportMessengerBotShouldRespond(b, messengermod.MessageHookEvent{
		ContentText: "@hub-ai что такое REST?",
	}, groupChannel, []uuid.UUID{uuid.New()}))

	assert.True(t, ai.ExportMessengerBotShouldRespond(b, messengermod.MessageHookEvent{
		Mentions: []string{botID.String()},
	}, groupChannel, []uuid.UUID{uuid.New()}))

	assert.False(t, ai.ExportMessengerBotShouldRespond(b, messengermod.MessageHookEvent{
		ContentText: "hello team",
	}, groupChannel, []uuid.UUID{uuid.New()}))
}

func TestMarkdownToPlateJSON(t *testing.T) {
	doc := ai.ExportMarkdownToPlateJSON("**hi**")
	assert.Equal(t, true, doc["hub_ai_markdown"])
	children := doc["children"].([]map[string]interface{})
	assert.NotEmpty(t, children)
}
