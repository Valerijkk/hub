package messenger_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

func phase151PreparedMessageFixture() (messenger.AuthorizedMessageProducerScope, messenger.PostCreatePipelineRequest) {
	tenantID := uuid.New()
	actorID := uuid.New()
	channelID := uuid.New()
	return messenger.AuthorizedMessageProducerScope{
		TenantID: tenantID,
		ActorID:  actorID,
		ChannelID: channelID,
		Producer: messenger.TrustedMessageProducerUser,
	}, messenger.PostCreatePipelineRequest{
		Producer: messenger.TrustedMessageProducerUser,
		Message: messenger.MessageDocument{
			ChannelID:       channelID,
			SenderID:        actorID,
			MessageType:     "rich",
			PollID:          "poll-151",
			ContentJSON:     map[string]interface{}{"doc": []interface{}{map[string]interface{}{"text": "original"}}},
			ContentText:     "original",
			ClientNonce:     "phase151-full-snapshot",
			ForwardedFromID: "source-message-151",
			VoiceURL:        "/storage/voice-151.ogg",
			VoiceDurationMs: 1510,
			Attachments: []messenger.AttachmentRef{{
				FileID: "file-151", FileName: "design.png", FileSize: 151, MimeType: "image/png", URL: "/storage/design.png",
			}},
			Mentions:      []string{"user-151"},
			MentionGroups: []string{"online"},
		},
	}
}

func TestPhase151PreparePendingCanonicalMessagePreservesFullPayloadWithoutAliases(t *testing.T) {
	scope, request := phase151PreparedMessageFixture()
	mentionedID := uuid.New()
	resolved := []uuid.UUID{mentionedID}
	now := time.Date(2026, time.August, 26, 1, 45, 0, 123, time.FixedZone("MSK", 3*60*60))

	document, err := messenger.ExportPhase151PreparePendingCanonicalMessage(scope, request, resolved, now)

	require.NoError(t, err)
	require.NotNil(t, document)
	assert.Equal(t, request.Message.ChannelID, document.ChannelID)
	assert.Equal(t, request.Message.SenderID, document.SenderID)
	assert.Equal(t, request.Message.MessageType, document.MessageType)
	assert.Equal(t, request.Message.PollID, document.PollID)
	assert.Equal(t, request.Message.ContentText, document.ContentText)
	assert.Equal(t, request.Message.ClientNonce, document.ClientNonce)
	assert.Equal(t, request.Message.ForwardedFromID, document.ForwardedFromID)
	assert.Equal(t, request.Message.VoiceURL, document.VoiceURL)
	assert.Equal(t, request.Message.VoiceDurationMs, document.VoiceDurationMs)
	assert.Equal(t, request.Message.Attachments, document.Attachments)
	assert.Equal(t, request.Message.Mentions, document.Mentions)
	assert.Equal(t, request.Message.MentionGroups, document.MentionGroups)
	assert.Equal(t, resolved, document.ResolvedMentionUserIDs)
	assert.Equal(t, messenger.PostCreateEffectsStatePending, document.PostCreateEffectsState)
	assert.Equal(t, now.UTC(), document.CreatedAt)
	assert.JSONEq(t, `{"doc":[{"text":"original"}]}`, phase151MustMarshalJSON(t, document.ContentJSON))

	request.Message.ContentJSON["doc"].([]interface{})[0].(map[string]interface{})["text"] = "mutated"
	request.Message.Attachments[0].FileName = "mutated.png"
	request.Message.Mentions[0] = "mutated-user"
	request.Message.MentionGroups[0] = "mutated-group"
	resolved[0] = uuid.New()

	assert.JSONEq(t, `{"doc":[{"text":"original"}]}`, phase151MustMarshalJSON(t, document.ContentJSON))
	assert.Equal(t, "design.png", document.Attachments[0].FileName)
	assert.Equal(t, []string{"user-151"}, document.Mentions)
	assert.Equal(t, []string{"online"}, document.MentionGroups)
	assert.Equal(t, []uuid.UUID{mentionedID}, document.ResolvedMentionUserIDs)
}

func phase151MustMarshalJSON(t *testing.T, value interface{}) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	return string(encoded)
}
func TestPhase151PreparePendingCanonicalMessageRejectsInvalidOrServerManagedFields(t *testing.T) {
	now := time.Now().UTC()
	cyclicContent := map[string]interface{}{}
	cyclicContent["self"] = cyclicContent
	tests := []struct {
		name   string
		mutate func(*messenger.MessageDocument)
	}{
		{name: "non-json content", mutate: func(message *messenger.MessageDocument) { message.ContentJSON = cyclicContent }},
		{name: "id", mutate: func(message *messenger.MessageDocument) { message.ID = uuid.NewString() }},
		{name: "reply count", mutate: func(message *messenger.MessageDocument) { message.ReplyCount = 1 }},
		{name: "reactions", mutate: func(message *messenger.MessageDocument) { message.Reactions = []messenger.ReactionEntry{{Emoji: "thumbsup"}} }},
		{name: "resolved mentions", mutate: func(message *messenger.MessageDocument) { message.ResolvedMentionUserIDs = []uuid.UUID{uuid.New()} }},
		{name: "effects state", mutate: func(message *messenger.MessageDocument) { message.PostCreateEffectsState = messenger.PostCreateEffectsStateComplete }},
		{name: "effects completed", mutate: func(message *messenger.MessageDocument) { message.PostCreateEffectsCompletedAt = &now }},
		{name: "edited at", mutate: func(message *messenger.MessageDocument) { message.EditedAt = &now }},
		{name: "deleted at", mutate: func(message *messenger.MessageDocument) { message.DeletedAt = &now }},
		{name: "created at", mutate: func(message *messenger.MessageDocument) { message.CreatedAt = now }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scope, request := phase151PreparedMessageFixture()
			test.mutate(&request.Message)

			document, err := messenger.ExportPhase151PreparePendingCanonicalMessage(scope, request, nil, now)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.Nil(t, document)
		})
	}
}
