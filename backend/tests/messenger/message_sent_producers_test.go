package messenger_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

type capturedPublication struct {
	subject string
	payload any
}

// capturePublisher is the in-memory EventPublisher boundary for producer units.
// It intentionally has no NATS client, network, or nil-publisher fallback.
type capturePublisher struct {
	mu    sync.Mutex
	calls []capturedPublication
}

func (p *capturePublisher) PublishJSON(subject string, payload any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls = append(p.calls, capturedPublication{subject: subject, payload: payload})
}

func (p *capturePublisher) snapshot() []capturedPublication {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]capturedPublication(nil), p.calls...)
}

func TestMessageSentProducerCaptureConstructorRejectsNilPublisher(t *testing.T) {
	var typedNil *capturePublisher

	for _, testCase := range []struct {
		name      string
		publisher messenger.EventPublisher
	}{
		{name: "nil interface", publisher: nil},
		{name: "typed nil", publisher: typedNil},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			assert.PanicsWithValue(t, "messenger hubtest event publisher must not be nil or typed-nil", func() {
				messenger.ExportNewMessengerServiceWithEventPublisher(newMockRepository(), testCase.publisher)
			})
		})
	}
}

func TestMessageSentProducerNormalSendPublishesCanonicalNestedMessageOnceAndDoesNotRepublishRetry(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	mentionedID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID, mentionedID)
	root := addLiveThreadRoot(repo, channelID, senderID)
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)
	request := messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{
			"type": "doc",
			"children": []interface{}{
				map[string]interface{}{"type": "p", "children": []interface{}{map[string]interface{}{"text": "canonical rich reply"}}},
			},
		},
		ContentText:   "canonical rich reply",
		ThreadID:      root.ID,
		ClientNonce:   "message-sent-producer-retry-0001",
		Mentions:      []string{mentionedID.String()},
		MentionGroups: []string{"all"},
		Attachments: []messenger.AttachmentRef{
			{FileID: "file-1", FileName: "one.pdf", FileSize: 10, MimeType: "application/pdf", URL: "https://example.test/one.pdf"},
			{FileID: "file-2", FileName: "two.ogg", FileSize: 20, MimeType: "audio/ogg", URL: "https://example.test/two.ogg"},
		},
		VoiceURL:        "https://example.test/voice.ogg",
		VoiceDurationMs: 1234,
	}
	canonical := messageDocumentForRequest("canonical-rich-reaction", channelID, senderID, request)
	canonical.Reactions = []messenger.ReactionEntry{{Emoji: "thumbsup", UserIDs: []string{mentionedID.String()}}}
	require.NotEmpty(t, canonical.Reactions)
	repo.createMessageResultOverride = &messenger.CreateMessageResult{Message: canonical, Created: true}

	response, err := service.SendMessage(ctx, channelID, senderID, request)
	require.NoError(t, err)
	require.NotNil(t, response)

	expected := messenger.MessageResponse{
		ID:              response.ID,
		ChannelID:       channelID.String(),
		SenderID:        senderID.String(),
		ContentJSON:     request.ContentJSON,
		ContentText:     "canonical rich reply",
		ThreadID:        root.ID,
		ClientNonce:     "message-sent-producer-retry-0001",
		ReplyCount:      0,
		VoiceURL:        "https://example.test/voice.ogg",
		VoiceDurationMs: 1234,
		Attachments:     request.Attachments,
		Reactions:       canonical.Reactions,
		Mentions:        []string{mentionedID.String()},
		MentionGroups:   []string{"all"},
		CreatedAt:       response.CreatedAt,
	}
	assert.Equal(t, expected, *response)
	assert.Equal(t, canonical.Reactions, response.Reactions)
	event := assertCanonicalMessageSent(t, publisher, channelID, expected, []string{mentionedID.String()})
	assert.Equal(t, canonical.Reactions, event.Message.Reactions)
	repo.messages[canonical.ID] = canonical
	repo.messagesBySenderNonce[messageSenderNonceKey(senderID, request.ClientNonce)] = canonical.ID

	retry, err := service.SendMessage(ctx, channelID, senderID, request)
	require.NoError(t, err)
	require.NotNil(t, retry)
	assert.Equal(t, response.ID, retry.ID)
	assertCanonicalMessageSent(t, publisher, channelID, expected, []string{mentionedID.String()})
	assert.Equal(t, 1, repo.createMessageCalls)
}

func TestMessageSentProducerForwardPublishesCanonicalForwardedResponse(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	userID := uuid.New()
	sourceChannelID := setupChannelWithMembers(repo, "public", false, userID)
	targetChannelID := setupChannelWithMembers(repo, "public", false, userID)
	source := &messenger.MessageDocument{
		ID:          "source-message-1",
		ChannelID:   sourceChannelID,
		SenderID:    uuid.New(),
		ContentJSON: map[string]interface{}{"type": "doc", "children": []interface{}{map[string]interface{}{"text": "forward source"}}},
		ContentText: "forward source",
		Attachments: []messenger.AttachmentRef{{FileID: "source-file", FileName: "source.txt", FileSize: 11, MimeType: "text/plain", URL: "https://example.test/source.txt"}},
	}
	repo.messages[source.ID] = source
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

	response, err := service.ForwardMessage(ctx, sourceChannelID, userID, source.ID, targetChannelID.String())
	require.NoError(t, err)
	require.NotNil(t, response)

	expected := messenger.MessageResponse{
		ID:              response.ID,
		ChannelID:       targetChannelID.String(),
		SenderID:        userID.String(),
		ContentJSON:     source.ContentJSON,
		ContentText:     "forward source",
		ReplyCount:      0,
		ForwardedFromID: "source-message-1",
		Attachments:     source.Attachments,
		Reactions:       nil,
		CreatedAt:       response.CreatedAt,
	}
	assert.Equal(t, expected, *response)
	assertCanonicalMessageSent(t, publisher, targetChannelID, expected, nil)
}

func TestMessageSentProducerBotPublishesCanonicalNestedMessage(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	channelID := setupChannelWithMembers(repo, "public", false, uuid.New())
	botID := uuid.New()
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)
	service.SetSystemBotID(botID)
	request := messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{"type": "doc", "children": []interface{}{map[string]interface{}{"text": "bot delivery"}}},
		ContentText: "bot delivery",
	}

	response, err := service.SendBotMessage(ctx, channelID, request)
	require.NoError(t, err)
	require.NotNil(t, response)

	expected := messenger.MessageResponse{
		ID:          response.ID,
		ChannelID:   channelID.String(),
		SenderID:    botID.String(),
		ContentJSON: request.ContentJSON,
		ContentText: "bot delivery",
		ReplyCount:  0,
		Reactions:   nil,
		CreatedAt:   response.CreatedAt,
	}
	assert.Equal(t, expected, *response)
	assertCanonicalMessageSent(t, publisher, channelID, expected, nil)
}

func TestMessageSentProducerPollWrapperPublishesCanonicalNestedMessageAlongsidePollEvent(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	creatorID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, creatorID)
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

	poll, err := service.CreatePoll(ctx, channelID, creatorID, messenger.CreatePollRequest{
		Question:      "Ship canonical events?",
		Options:       []string{"yes", "also yes"},
		IsAnonymous:   false,
		IsMultiChoice: true,
	})
	require.NoError(t, err)
	require.NotNil(t, poll)

	event := requireOneMessageSent(t, publisher, channelID)
	assertCanonicalMessageSentJSON(t, event)
	assert.NotEmpty(t, event.Message.ID)
	assert.Equal(t, channelID.String(), event.Message.ChannelID)
	assert.Equal(t, creatorID.String(), event.Message.SenderID)
	assert.Equal(t, "poll", event.Message.MessageType)
	assert.Equal(t, poll.ID, event.Message.PollID)
	assert.Equal(t, "Ship canonical events?", event.Message.ContentText)
	assert.Equal(t, map[string]interface{}{
		"type": "doc",
		"children": []interface{}{
			map[string]interface{}{"type": "p", "children": []interface{}{map[string]interface{}{"text": "Ship canonical events?"}}},
		},
	}, event.Message.ContentJSON)
	assert.Zero(t, event.Message.ReplyCount)
	assert.Empty(t, event.Message.Attachments)
	assert.Empty(t, event.Message.Reactions)
	assert.NotZero(t, event.Message.CreatedAt)
}

func TestMessageSentProducerPromotionBacklinkPublishesCanonicalNestedMessage(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	userID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	source := &messenger.MessageDocument{
		ID:          "promotion-source-1",
		ChannelID:   channelID,
		SenderID:    userID,
		ContentJSON: map[string]interface{}{"type": "doc", "children": []interface{}{map[string]interface{}{"text": "promote this"}}},
		ContentText: "promote this",
	}
	repo.messages[source.ID] = source
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)
	service.SetIssuePromoter(fixedIssuePromoter{issueKey: "HUB-149", issueID: uuid.NewString()})
	service.SetPromotePermissionChecker(allowPromotionChecker{})

	response, err := service.PromoteToIssue(ctx, source.ID, userID, messenger.PromoteToIssueRequest{
		ProjectID: uuid.NewString(),
		Title:     "Canonical publication",
	})
	require.NoError(t, err)
	require.Equal(t, "HUB-149", response.IssueKey)

	expectedText := "📌 Промоутнуто в задачу HUB-149"
	event := requireOneMessageSent(t, publisher, channelID)
	assertCanonicalMessageSentJSON(t, event)
	assert.NotEmpty(t, event.Message.ID)
	assert.Equal(t, channelID.String(), event.Message.ChannelID)
	assert.Equal(t, userID.String(), event.Message.SenderID)
	assert.Equal(t, expectedText, event.Message.ContentText)
	assert.Equal(t, source.ID, event.Message.ThreadID)
	assert.Equal(t, map[string]interface{}{
		"type": "doc",
		"children": []interface{}{
			map[string]interface{}{"type": "p", "children": []interface{}{map[string]interface{}{"text": expectedText}}},
		},
	}, event.Message.ContentJSON)
	assert.Zero(t, event.Message.ReplyCount)
	assert.Empty(t, event.Message.Attachments)
	assert.Empty(t, event.Message.Reactions)
	assert.NotZero(t, event.Message.CreatedAt)
}

func assertCanonicalMessageSent(t *testing.T, publisher *capturePublisher, channelID uuid.UUID, expected messenger.MessageResponse, expectedMentionedUserIDs []string) messenger.MessageSentEvent {
	t.Helper()
	event := requireOneMessageSent(t, publisher, channelID)
	assert.Equal(t, expected, event.Message)
	assert.Equal(t, expectedMentionedUserIDs, event.MentionedUserIDs)
	assertCanonicalMessageSentJSON(t, event)
	return event
}

func requireOneMessageSent(t *testing.T, publisher *capturePublisher, channelID uuid.UUID) messenger.MessageSentEvent {
	t.Helper()
	wantSubject := fmt.Sprintf(messenger.SubjectMessageSent, channelID.String())
	var events []messenger.MessageSentEvent
	for _, call := range publisher.snapshot() {
		if call.subject != wantSubject {
			continue
		}
		event, ok := call.payload.(messenger.MessageSentEvent)
		require.Truef(t, ok, "message_sent payload must be messenger.MessageSentEvent, got %T", call.payload)
		events = append(events, event)
	}
	require.Len(t, events, 1, "one logical message must publish one matching message_sent event")
	return events[0]
}

func assertCanonicalMessageSentJSON(t *testing.T, event messenger.MessageSentEvent) {
	t.Helper()
	raw, err := json.Marshal(event)
	require.NoError(t, err)

	var encoded map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &encoded))
	messageJSON, ok := encoded["message"]
	require.True(t, ok, "canonical message_sent JSON must contain nested message")
	assert.NotEmpty(t, messageJSON)
	assert.NotContains(t, encoded, "message_id")
	assert.NotContains(t, encoded, "channel_id")
	assert.NotContains(t, encoded, "sender_id")
	assert.NotContains(t, encoded, "content_text")

	var nested messenger.MessageResponse
	require.NoError(t, json.Unmarshal(messageJSON, &nested))
	assert.Equal(t, event.Message, nested)
}

type fixedIssuePromoter struct {
	issueKey string
	issueID  string
}

func (p fixedIssuePromoter) CreateIssueFromPromotion(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string, _ string, _ map[string]interface{}) (string, string, error) {
	return p.issueKey, p.issueID, nil
}

type allowPromotionChecker struct{}

func (allowPromotionChecker) CanCreateTaskInProject(_ context.Context, _ uuid.UUID, _ uuid.UUID) (bool, error) {
	return true, nil
}
