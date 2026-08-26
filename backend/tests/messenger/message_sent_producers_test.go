package messenger_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

func TestMessageSentProducerStorageRaceReturnsCanonicalWithoutRepublishing(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	recipientID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID, recipientID)
	request := messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{"type": "doc", "children": []interface{}{map[string]interface{}{"text": "losing request"}}},
		ContentText: "losing request",
		ClientNonce: "message-sent-storage-race-0001",
	}
	canonical := messageDocumentForRequest("storage-race-canonical", channelID, senderID, request)
	canonical.ReplyCount = 3
	canonical.Reactions = []messenger.ReactionEntry{{Emoji: "white_check_mark", UserIDs: []string{recipientID.String()}}}
	repo.createMessageResultOverride = &messenger.CreateMessageResult{Message: canonical, Created: false}
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

	response, err := service.SendMessage(ctx, channelID, senderID, request)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, canonical.ID, response.ID)
	require.Equal(t, canonical.ChannelID.String(), response.ChannelID)
	require.Equal(t, canonical.SenderID.String(), response.SenderID)
	require.Equal(t, canonical.ContentJSON, response.ContentJSON)
	require.Equal(t, canonical.ContentText, response.ContentText)
	require.Equal(t, canonical.ClientNonce, response.ClientNonce)
	require.Equal(t, canonical.ReplyCount, response.ReplyCount)
	require.Equal(t, canonical.Reactions, response.Reactions)
	require.Equal(t, canonical.CreatedAt, response.CreatedAt)
	require.Equal(t, 1, repo.createMessageCalls)
	requireNoMessageSentForChannel(t, publisher, channelID)
	requireNoNewMessageNotifications(t, publisher)
}

func TestMessageSentProducerNormalSendPublishesCanonicalNestedMessageOnceAndDoesNotRepublishRetry(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	mentionedID := uuid.New()
	recipientID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID, mentionedID, recipientID)
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
	event := assertCanonicalMessageSent(t, publisher, channelID, expected, []string{mentionedID.String(), recipientID.String()})
	assert.Equal(t, canonical.Reactions, event.Message.Reactions)
	requireNewMessageNotifications(t, publisher, channelID, response.ID, mentionedID, recipientID)
	repo.messages[canonical.ID] = canonical
	repo.messagesBySenderNonce[messageSenderNonceKey(senderID, request.ClientNonce)] = canonical.ID

	retry, err := service.SendMessage(ctx, channelID, senderID, request)
	require.NoError(t, err)
	require.NotNil(t, retry)
	assert.Equal(t, response.ID, retry.ID)
	assertCanonicalMessageSent(t, publisher, channelID, expected, []string{mentionedID.String(), recipientID.String()})
	requireNewMessageNotifications(t, publisher, channelID, response.ID, mentionedID, recipientID)
	assert.Equal(t, 1, repo.createMessageCalls)
}

func TestMessageSentProducerForwardPublishesCanonicalForwardedResponse(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	userID := uuid.New()
	recipientID := uuid.New()
	sourceChannelID := setupChannelWithMembers(repo, "public", false, userID)
	targetChannelID := setupChannelWithMembers(repo, "public", false, userID, recipientID)
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
	requireNewMessageNotifications(t, publisher, targetChannelID, response.ID, recipientID)
}

func TestMessageSentProducerBotPublishesCanonicalNestedMessage(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	botID := uuid.New()
	recipientID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, botID, recipientID)
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
	requireNewMessageNotifications(t, publisher, channelID, response.ID, recipientID)
}

func TestMessageSentProducerPollWrapperPublishesCanonicalNestedMessageAlongsidePollEvent(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	creatorID := uuid.New()
	recipientID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, creatorID, recipientID)
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
	requireNewMessageNotifications(t, publisher, channelID, event.Message.ID, recipientID)
}

func TestMessageSentProducerPromotionBacklinkPublishesCanonicalNestedMessage(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	userID := uuid.New()
	recipientID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, userID, recipientID)
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

	response, err := service.PromoteToIssue(ctx, source.ChannelID, source.ID, userID, messenger.PromoteToIssueRequest{
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
	requireNewMessageNotifications(t, publisher, channelID, event.Message.ID, recipientID)
}

func TestMessageSentProducerUserNotificationsRemainBestEffortAndPostPersistence(t *testing.T) {
	ctx := context.Background()

	t.Run("poll wrapper insert failure keeps successful poll but publishes no message notification", func(t *testing.T) {
		base := newMockRepository()
		creatorID := uuid.New()
		recipientID := uuid.New()
		channelID := setupChannelWithMembers(base, "public", false, creatorID, recipientID)
		publisher := &capturePublisher{}
		repo := &messageNotificationFailureRepository{
			mockRepository: base,
			insertMessageErr: errors.New("poll wrapper insert failed"),
		}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		poll, err := service.CreatePoll(ctx, channelID, creatorID, messenger.CreatePollRequest{
			Question: "Persist before notify?",
			Options:  []string{"yes", "always"},
		})
		require.NoError(t, err)
		require.NotNil(t, poll)
		requireNoNewMessageNotifications(t, publisher)
	})

	t.Run("promotion backlink insert failure keeps successful promotion but publishes no message notification", func(t *testing.T) {
		base := newMockRepository()
		userID := uuid.New()
		recipientID := uuid.New()
		channelID := setupChannelWithMembers(base, "public", false, userID, recipientID)
		source := &messenger.MessageDocument{
			ID:          "notification-promotion-source",
			ChannelID:   channelID,
			SenderID:    userID,
			ContentJSON: map[string]interface{}{"type": "doc", "children": []interface{}{map[string]interface{}{"text": "promote this"}}},
			ContentText: "promote this",
		}
		base.messages[source.ID] = source
		publisher := &capturePublisher{}
		repo := &messageNotificationFailureRepository{
			mockRepository: base,
			insertMessageErr: errors.New("promotion backlink insert failed"),
		}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)
		service.SetIssuePromoter(fixedIssuePromoter{issueKey: "HUB-149", issueID: uuid.NewString()})
		service.SetPromotePermissionChecker(allowPromotionChecker{})

		response, err := service.PromoteToIssue(ctx, source.ChannelID, source.ID, userID, messenger.PromoteToIssueRequest{
			ProjectID: uuid.NewString(),
			Title:     "Notification failure",
		})
		require.NoError(t, err)
		require.Equal(t, "HUB-149", response.IssueKey)
		requireNoNewMessageNotifications(t, publisher)
	})

	t.Run("recipient lookup failure keeps successful forward and channel event", func(t *testing.T) {
		base := newMockRepository()
		userID := uuid.New()
		recipientID := uuid.New()
		sourceChannelID := setupChannelWithMembers(base, "public", false, userID)
		targetChannelID := setupChannelWithMembers(base, "public", false, userID, recipientID)
		source := &messenger.MessageDocument{
			ID:          "notification-forward-lookup-source",
			ChannelID:   sourceChannelID,
			SenderID:    uuid.New(),
			ContentJSON: map[string]interface{}{"type": "doc", "children": []interface{}{map[string]interface{}{"text": "forward despite notification lookup"}}},
			ContentText: "forward despite notification lookup",
		}
		base.messages[source.ID] = source
		publisher := &capturePublisher{}
		repo := &messageNotificationFailureRepository{
			mockRepository:          base,
			getChannelMemberIDsErr: errors.New("forward recipient lookup failed"),
		}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		response, err := service.ForwardMessage(ctx, sourceChannelID, userID, source.ID, targetChannelID.String())
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Equal(t, targetChannelID.String(), response.ChannelID)
		requireOneMessageSent(t, publisher, targetChannelID)
		requireNoNewMessageNotifications(t, publisher)
	})

	t.Run("recipient lookup failure keeps successful bot send and channel event", func(t *testing.T) {
		base := newMockRepository()
		botID := uuid.New()
		recipientID := uuid.New()
		channelID := setupChannelWithMembers(base, "public", false, botID, recipientID)
		publisher := &capturePublisher{}
		repo := &messageNotificationFailureRepository{
			mockRepository:          base,
			getChannelMemberIDsErr: errors.New("bot recipient lookup failed"),
		}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)
		service.SetSystemBotID(botID)

		response, err := service.SendBotMessage(ctx, channelID, messenger.SendMessageRequest{
			ContentJSON: map[string]interface{}{"type": "doc", "children": []interface{}{map[string]interface{}{"text": "bot send despite notification lookup"}}},
			ContentText: "bot send despite notification lookup",
		})
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Equal(t, botID.String(), response.SenderID)
		requireOneMessageSent(t, publisher, channelID)
		requireNoNewMessageNotifications(t, publisher)
	})

	t.Run("recipient lookup failure keeps successful poll wrapper and channel event", func(t *testing.T) {
		base := newMockRepository()
		creatorID := uuid.New()
		recipientID := uuid.New()
		channelID := setupChannelWithMembers(base, "public", false, creatorID, recipientID)
		publisher := &capturePublisher{}
		repo := &messageNotificationFailureRepository{
			mockRepository:          base,
			getChannelMemberIDsErr: errors.New("poll recipient lookup failed"),
		}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		poll, err := service.CreatePoll(ctx, channelID, creatorID, messenger.CreatePollRequest{
			Question: "Keep poll after lookup failure?",
			Options:  []string{"yes", "always"},
		})
		require.NoError(t, err)
		require.NotNil(t, poll)
		event := requireOneMessageSent(t, publisher, channelID)
		require.Equal(t, poll.ID, event.Message.PollID)
		requireNoNewMessageNotifications(t, publisher)
	})

	t.Run("recipient lookup failure keeps successful promotion backlink and channel event", func(t *testing.T) {
		base := newMockRepository()
		userID := uuid.New()
		recipientID := uuid.New()
		channelID := setupChannelWithMembers(base, "public", false, userID, recipientID)
		source := &messenger.MessageDocument{
			ID:          "notification-promotion-lookup-source",
			ChannelID:   channelID,
			SenderID:    userID,
			ContentJSON: map[string]interface{}{"type": "doc", "children": []interface{}{map[string]interface{}{"text": "promote despite notification lookup"}}},
			ContentText: "promote despite notification lookup",
		}
		base.messages[source.ID] = source
		publisher := &capturePublisher{}
		repo := &messageNotificationFailureRepository{
			mockRepository:          base,
			getChannelMemberIDsErr: errors.New("promotion recipient lookup failed"),
		}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)
		service.SetIssuePromoter(fixedIssuePromoter{issueKey: "HUB-149", issueID: uuid.NewString()})
		service.SetPromotePermissionChecker(allowPromotionChecker{})

		response, err := service.PromoteToIssue(ctx, source.ChannelID, source.ID, userID, messenger.PromoteToIssueRequest{
			ProjectID: uuid.NewString(),
			Title:     "Lookup failure promotion",
		})
		require.NoError(t, err)
		require.Equal(t, "HUB-149", response.IssueKey)
		requireOneMessageSent(t, publisher, channelID)
		requireNoNewMessageNotifications(t, publisher)
	})

	t.Run("recipient snapshot filters sender duplicate recipients and nil uuid", func(t *testing.T) {
		base := newMockRepository()
		senderID := uuid.New()
		recipientID := uuid.New()
		channelID := setupChannelWithMembers(base, "public", false, senderID, recipientID)
		publisher := &capturePublisher{}
		repo := &messageNotificationFailureRepository{
			mockRepository: base,
			getChannelMemberIDsOverride: []uuid.UUID{
				senderID,
				recipientID,
				recipientID,
				uuid.Nil,
			},
		}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		response, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{
			ContentJSON: map[string]interface{}{"type": "doc", "children": []interface{}{map[string]interface{}{"text": "deduplicated recipients"}}},
			ContentText: "deduplicated recipients",
			ClientNonce: "notification-recipient-filter-0001",
		})
		require.NoError(t, err)
		require.NotNil(t, response)
		requireOneMessageSent(t, publisher, channelID)
		requireNewMessageNotifications(t, publisher, channelID, response.ID, recipientID)
		requireNoNewMessageNotificationsFor(t, publisher, senderID, uuid.Nil)
	})

	t.Run("recipient lookup failure does not change successful send but publishes no notification", func(t *testing.T) {
		base := newMockRepository()
		senderID := uuid.New()
		recipientID := uuid.New()
		channelID := setupChannelWithMembers(base, "public", false, senderID, recipientID)
		publisher := &capturePublisher{}
		repo := &messageNotificationFailureRepository{
			mockRepository:          base,
			getChannelMemberIDsErr: errors.New("recipient lookup failed"),
		}
		service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)

		response, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{
			ContentJSON: map[string]interface{}{"type": "doc", "children": []interface{}{map[string]interface{}{"text": "primary send persists"}}},
			ContentText: "primary send persists",
			ClientNonce: "notification-member-lookup-0001",
		})
		require.NoError(t, err)
		require.NotNil(t, response)
		requireOneMessageSent(t, publisher, channelID)
		requireNoNewMessageNotifications(t, publisher)
	})
}

func assertCanonicalMessageSent(t *testing.T, publisher *capturePublisher, channelID uuid.UUID, expected messenger.MessageResponse, expectedMentionedUserIDs []string) messenger.MessageSentEvent {
	t.Helper()
	event := requireOneMessageSent(t, publisher, channelID)
	assert.Equal(t, expected, event.Message)
	assert.ElementsMatch(t, expectedMentionedUserIDs, event.MentionedUserIDs)
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

func requireNoMessageSentForChannel(t *testing.T, publisher *capturePublisher, channelID uuid.UUID) {
	t.Helper()
	wantSubject := fmt.Sprintf(messenger.SubjectMessageSent, channelID.String())
	for _, publication := range publisher.snapshot() {
		require.NotEqual(t, wantSubject, publication.subject, "idempotency loser must not republish channel message_sent")
	}
}

func requireNewMessageNotifications(t *testing.T, publisher *capturePublisher, channelID uuid.UUID, messageID string, expectedRecipients ...uuid.UUID) {
	t.Helper()
	expected := make(map[string]struct{}, len(expectedRecipients))
	for _, recipientID := range expectedRecipients {
		expected[fmt.Sprintf("messenger.user.%s.new_message_notification", recipientID.String())] = struct{}{}
	}

	seen := make(map[string]int, len(expected))
	for _, publication := range publisher.snapshot() {
		if !strings.HasSuffix(publication.subject, ".new_message_notification") {
			continue
		}
		require.Containsf(t, expected, publication.subject, "unexpected new-message notification subject: %s", publication.subject)
		seen[publication.subject]++
		requireNewMessageNotificationPayload(t, publication.payload, channelID.String(), messageID)
	}

	require.Len(t, seen, len(expected), "every non-sender current member receives one new-message notification")
	for subject := range expected {
		require.Equalf(t, 1, seen[subject], "%s must receive exactly one notification", subject)
	}
}

func requireNoNewMessageNotifications(t *testing.T, publisher *capturePublisher) {
	t.Helper()
	for _, publication := range publisher.snapshot() {
		require.Falsef(t, strings.HasSuffix(publication.subject, ".new_message_notification"), "unexpected new-message notification: %s", publication.subject)
	}
}

func requireNoNewMessageNotificationsFor(t *testing.T, publisher *capturePublisher, forbiddenRecipients ...uuid.UUID) {
	t.Helper()
	for _, recipientID := range forbiddenRecipients {
		forbiddenSubject := fmt.Sprintf("messenger.user.%s.new_message_notification", recipientID.String())
		for _, publication := range publisher.snapshot() {
			require.NotEqualf(t, forbiddenSubject, publication.subject, "recipient %s must not receive a new-message notification", recipientID)
		}
	}
}

func requireNewMessageNotificationPayload(t *testing.T, payload any, channelID, messageID string) {
	t.Helper()
	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	var object map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &object))
	require.Len(t, object, 2, "new-message notification payload must contain only channel_id and message_id")

	var actualChannelID, actualMessageID string
	require.NoError(t, json.Unmarshal(object["channel_id"], &actualChannelID))
	require.NoError(t, json.Unmarshal(object["message_id"], &actualMessageID))
	require.Equal(t, channelID, actualChannelID)
	require.Equal(t, messageID, actualMessageID)
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

func (allowPromotionChecker) AuthorizePromotionTarget(_ context.Context, _ uuid.UUID, _ uuid.UUID) (bool, bool, error) {
	return true, true, nil
}

// messageNotificationFailureRepository exposes failure seams only for this
// producer suite while preserving the shared fixture's ordinary behavior.
type messageNotificationFailureRepository struct {
	*mockRepository
	insertMessageErr            error
	getChannelMemberIDsErr      error
	getChannelMemberIDsOverride []uuid.UUID
}

func (r *messageNotificationFailureRepository) InsertMessage(ctx context.Context, doc *messenger.MessageDocument) (string, error) {
	if r.insertMessageErr != nil {
		return "", r.insertMessageErr
	}
	return r.mockRepository.InsertMessage(ctx, doc)
}

func (r *messageNotificationFailureRepository) GetChannelMemberIDs(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
	if r.getChannelMemberIDsErr != nil {
		return nil, r.getChannelMemberIDsErr
	}
	if r.getChannelMemberIDsOverride != nil {
		return append([]uuid.UUID(nil), r.getChannelMemberIDsOverride...), nil
	}
	return r.mockRepository.GetChannelMemberIDs(ctx, channelID)
}

func TestPhase151ExistingProducerSuiteUsesClosedPipelineAllowlist(t *testing.T) {
	assert.ElementsMatch(t, []messenger.TrustedMessageProducer{
		messenger.TrustedMessageProducerUser,
		messenger.TrustedMessageProducerForward,
		messenger.TrustedMessageProducerSystemBot,
		messenger.TrustedMessageProducerPoll,
		messenger.TrustedMessageProducerPromotionBacklink,
	}, messenger.ExportPhase151TrustedMessageProducers())
	assert.NotContains(t, messenger.ExportPhase151TrustedMessageProducers(), messenger.TrustedMessageProducer("generic"))
}
func TestPhase151SystemBotProducerRequiresAuthorizedBotScope(t *testing.T) {
	command := phase151BotMongoCommand(messenger.BotMongoPathSend)
	scope := phase151AuthorizedBotScope(command)
	producer, ok := messenger.ExportPhase151TrustedProducerForBotScope(scope)
	require.True(t, ok)
	assert.Equal(t, messenger.TrustedMessageProducerSystemBot, producer)

	scope.BotEnabled = false
	_, ok = messenger.ExportPhase151TrustedProducerForBotScope(scope)
	assert.False(t, ok)
	_, ok = messenger.ExportPhase151TrustedProducerForBotScope(nil)
	assert.False(t, ok)
}
