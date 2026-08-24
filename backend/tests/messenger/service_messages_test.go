package messenger_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

func TestSendMessageRejectsForeignThreadParent(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	foreignChannelID := setupChannelWithMembers(repo, "public", false, senderID)
	parent := &messenger.MessageDocument{ID: uuid.NewString(), ChannelID: foreignChannelID, SenderID: senderID, ContentText: "foreign root"}
	repo.messages[parent.ID] = parent
	service := messenger.ExportNewMessengerServiceTest(repo)

	response, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{ContentText: "reply", ThreadID: parent.ID})

	require.Error(t, err)
	assert.Nil(t, response)
	requireAppErrorCode(t, err, 404)
	assert.Len(t, repo.messages, 1)
	assert.Zero(t, parent.ReplyCount)
}

func TestSendMessageRejectsMissingThreadParent(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	untouchedParent := &messenger.MessageDocument{ID: uuid.NewString(), ChannelID: channelID, SenderID: senderID, ContentText: "untouched root"}
	repo.messages[untouchedParent.ID] = untouchedParent
	service := messenger.ExportNewMessengerServiceTest(repo)

	response, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{ContentText: "reply", ThreadID: uuid.NewString()})

	require.Error(t, err)
	assert.Nil(t, response)
	requireAppErrorCode(t, err, 404)
	assert.Len(t, repo.messages, 1)
	assert.Zero(t, untouchedParent.ReplyCount)
}

func TestSendMessageRejectsDeletedThreadParent(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	repo.findMessageIncludeDeleted = true
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	deletedAt := time.Now().UTC()
	parent := &messenger.MessageDocument{ID: uuid.NewString(), ChannelID: channelID, SenderID: senderID, ContentText: "deleted root", DeletedAt: &deletedAt}
	repo.messages[parent.ID] = parent
	service := messenger.ExportNewMessengerServiceTest(repo)

	response, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{ContentText: "reply", ThreadID: parent.ID})

	require.Error(t, err)
	assert.Nil(t, response)
	requireAppErrorCode(t, err, 404)
	assert.Len(t, repo.messages, 1)
	assert.Zero(t, parent.ReplyCount)
}

func TestSendMessageRejectsNestedThreadParent(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	root := &messenger.MessageDocument{ID: uuid.NewString(), ChannelID: channelID, SenderID: senderID, ContentText: "root"}
	nestedParent := &messenger.MessageDocument{ID: uuid.NewString(), ChannelID: channelID, SenderID: senderID, ContentText: "nested", ThreadID: root.ID}
	repo.messages[root.ID] = root
	repo.messages[nestedParent.ID] = nestedParent
	service := messenger.ExportNewMessengerServiceTest(repo)

	response, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{ContentText: "reply", ThreadID: nestedParent.ID})

	require.Error(t, err)
	assert.Nil(t, response)
	requireAppErrorCode(t, err, 400)
	assert.Len(t, repo.messages, 2)
	assert.Zero(t, nestedParent.ReplyCount)
}

func TestSendMessageAcceptsLiveRootInSameChannel(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	service := messenger.ExportNewMessengerServiceTest(repo)

	root, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{ContentText: "root"})
	require.NoError(t, err)
	assert.Empty(t, root.ThreadID)

	reply, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{ContentText: "reply", ThreadID: root.ID})
	require.NoError(t, err)
	assert.Equal(t, root.ID, reply.ThreadID)
	require.Len(t, repo.messages, 2)
	assert.Equal(t, 1, repo.messages[root.ID].ReplyCount)

	replies := 0
	for _, document := range repo.messages {
		if document.ThreadID == root.ID {
			replies++
		}
	}
	assert.Equal(t, 1, replies)
}

func requireAppErrorCode(t *testing.T, err error, expectedCode int) {
	t.Helper()
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok, "expected AppError, got %T", err)
	assert.Equal(t, expectedCode, appErr.Code)
}

func TestSendMessageRetryReturnsCanonicalMessageWithoutSideEffects(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	root := addLiveThreadRoot(repo, channelID, senderID)
	service := messenger.ExportNewMessengerServiceTest(repo)
	req := messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{"type": "text", "body": "reply"},
		ContentText: "reply",
		ThreadID:    root.ID,
		ClientNonce: "retry-canonical-0001",
	}

	first, err := service.SendMessage(ctx, channelID, senderID, req)
	require.NoError(t, err)
	retry, err := service.SendMessage(ctx, channelID, senderID, req)

	require.NoError(t, err)
	require.NotNil(t, retry)
	assert.Equal(t, first.ID, retry.ID)
	assert.Equal(t, first.CreatedAt, retry.CreatedAt)
	assert.Equal(t, 1, repo.createMessageCalls)
	assert.Len(t, repo.messages, 2)
	assert.Equal(t, 1, root.ReplyCount)
}

func TestSendMessageRetryAfterParentDeletionReturnsCanonical(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	root := addLiveThreadRoot(repo, channelID, senderID)
	service := messenger.ExportNewMessengerServiceTest(repo)
	req := messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{"type": "text", "body": "reply"},
		ContentText: "reply",
		ThreadID:    root.ID,
		ClientNonce: "retry-parent-deleted-0001",
	}

	first, err := service.SendMessage(ctx, channelID, senderID, req)
	require.NoError(t, err)
	deletedAt := time.Now().UTC()
	root.DeletedAt = &deletedAt

	retry, err := service.SendMessage(ctx, channelID, senderID, req)

	require.NoError(t, err)
	require.NotNil(t, retry)
	assert.Equal(t, first.ID, retry.ID)
	assert.Equal(t, 1, repo.createMessageCalls)
	assert.Len(t, repo.messages, 2)
	assert.Equal(t, 1, root.ReplyCount)
}

func TestSendMessageConcurrentRetryCreatesOneMessage(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	root := addLiveThreadRoot(repo, channelID, senderID)
	service := messenger.ExportNewMessengerServiceTest(repo)
	req := messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{"type": "text", "body": "reply"},
		ContentText: "reply",
		ThreadID:    root.ID,
		ClientNonce: "concurrent-retry-0001",
	}

	const attempts = 16
	responses := make([]*messenger.MessageResponse, attempts)
	errs := make([]error, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			responses[index], errs[index] = service.SendMessage(ctx, channelID, senderID, req)
		}(i)
	}
	wg.Wait()

	for i := range attempts {
		require.NoError(t, errs[i])
		require.NotNil(t, responses[i])
		assert.Equal(t, responses[0].ID, responses[i].ID)
	}
	assert.GreaterOrEqual(t, repo.createMessageCalls, 1)
	assert.LessOrEqual(t, repo.createMessageCalls, attempts)
	assert.Len(t, repo.messages, 2)
	assert.Equal(t, 1, root.ReplyCount)
}

func TestSendMessageNonceReuseWithDifferentChannelConflicts(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	firstChannelID := setupChannelWithMembers(repo, "public", false, senderID)
	secondChannelID := setupChannelWithMembers(repo, "public", false, senderID)
	service := messenger.ExportNewMessengerServiceTest(repo)
	req := simpleSendRequest("nonce-channel-conflict-0001")

	_, err := service.SendMessage(ctx, firstChannelID, senderID, req)
	require.NoError(t, err)
	response, err := service.SendMessage(ctx, secondChannelID, senderID, req)

	assert.Nil(t, response)
	requireAppError(t, err, 409, "client_nonce уже использован для другого сообщения")
	assert.Equal(t, 1, repo.createMessageCalls)
	assert.Len(t, repo.messages, 1)
}

func TestSendMessageNonceReuseWithDifferentPayloadConflicts(t *testing.T) {
	tests := []struct {
		name          string
		initial       func(string) messenger.SendMessageRequest
		retry         func(messenger.SendMessageRequest) messenger.SendMessageRequest
		setup         func(*mockRepository, uuid.UUID, uuid.UUID)
		wantConflict  bool
		wantDocuments int
	}{
		{
			name:    "json number serialization is compatible",
			initial: simpleSendRequest,
			retry: func(req messenger.SendMessageRequest) messenger.SendMessageRequest {
				req.ContentJSON = map[string]interface{}{"count": json.Number("1"), "type": "text", "body": "message"}
				return req
			},
			wantDocuments: 1,
		},
		{
			name:    "content json differs",
			initial: simpleSendRequest,
			retry: func(req messenger.SendMessageRequest) messenger.SendMessageRequest {
				req.ContentJSON = map[string]interface{}{"type": "text", "body": "changed", "count": float64(1)}
				return req
			},
			wantConflict:  true,
			wantDocuments: 1,
		},
		{
			name:    "content text differs",
			initial: simpleSendRequest,
			retry: func(req messenger.SendMessageRequest) messenger.SendMessageRequest {
				req.ContentText = "changed"
				return req
			},
			wantConflict:  true,
			wantDocuments: 1,
		},
		{
			name:    "thread id differs",
			initial: simpleSendRequest,
			setup: func(repo *mockRepository, channelID, senderID uuid.UUID) {
				addLiveThreadRoot(repo, channelID, senderID)
			},
			retry: func(req messenger.SendMessageRequest) messenger.SendMessageRequest {
				req.ThreadID = ""
				return req
			},
			wantConflict:  true,
			wantDocuments: 2,
		},
		{
			name:    "attachment order differs",
			initial: richSendRequest,
			retry: func(req messenger.SendMessageRequest) messenger.SendMessageRequest {
				attachments := append([]messenger.AttachmentRef(nil), req.Attachments...)
				attachments[0], attachments[1] = attachments[1], attachments[0]
				req.Attachments = attachments
				return req
			},
			wantConflict:  true,
			wantDocuments: 1,
		},
		{
			name:    "attachment field differs",
			initial: richSendRequest,
			retry: func(req messenger.SendMessageRequest) messenger.SendMessageRequest {
				attachments := append([]messenger.AttachmentRef(nil), req.Attachments...)
				attachments[0].FileSize++
				req.Attachments = attachments
				return req
			},
			wantConflict:  true,
			wantDocuments: 1,
		},
		{
			name:    "voice url differs",
			initial: richSendRequest,
			retry: func(req messenger.SendMessageRequest) messenger.SendMessageRequest {
				req.VoiceURL = "https://example.test/changed.ogg"
				return req
			},
			wantConflict:  true,
			wantDocuments: 1,
		},
		{
			name:    "voice duration differs",
			initial: richSendRequest,
			retry: func(req messenger.SendMessageRequest) messenger.SendMessageRequest {
				req.VoiceDurationMs++
				return req
			},
			wantConflict:  true,
			wantDocuments: 1,
		},
		{
			name:    "mentions order and duplicates are compatible",
			initial: richSendRequest,
			retry: func(req messenger.SendMessageRequest) messenger.SendMessageRequest {
				req.Mentions = []string{"user-2", "user-1"}
				return req
			},
			wantDocuments: 1,
		},
		{
			name:    "mention groups order and duplicates are compatible",
			initial: richSendRequest,
			retry: func(req messenger.SendMessageRequest) messenger.SendMessageRequest {
				req.MentionGroups = []string{"online", "all"}
				return req
			},
			wantDocuments: 1,
		},
		{
			name: "nil and empty collections are compatible",
			initial: func(nonce string) messenger.SendMessageRequest {
				return messenger.SendMessageRequest{
					ContentJSON: map[string]interface{}{"type": "text", "body": "message", "count": float64(1)},
					ContentText: "message",
					ClientNonce: nonce,
				}
			},
			retry: func(req messenger.SendMessageRequest) messenger.SendMessageRequest {
				req.Attachments = []messenger.AttachmentRef{}
				req.Mentions = []string{}
				req.MentionGroups = []string{}
				return req
			},
			wantDocuments: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := newMockRepository()
			senderID := uuid.New()
			channelID := setupChannelWithMembers(repo, "public", false, senderID)
			if tt.setup != nil {
				tt.setup(repo, channelID, senderID)
			}
			service := messenger.ExportNewMessengerServiceTest(repo)
			req := tt.initial("payload-conflict-0001")
			if tt.name == "thread id differs" {
				for messageID := range repo.messages {
					req.ThreadID = messageID
					break
				}
			}

			first, err := service.SendMessage(ctx, channelID, senderID, req)
			require.NoError(t, err)
			retryReq := tt.retry(req)
			response, err := service.SendMessage(ctx, channelID, senderID, retryReq)

			if tt.wantConflict {
				assert.Nil(t, response)
				requireAppError(t, err, 409, "client_nonce уже использован для другого сообщения")
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, first.ID, response.ID)
			}
			assert.Equal(t, 1, repo.createMessageCalls)
			assert.Len(t, repo.messages, tt.wantDocuments)
		})
	}
}

func TestSendMessageNonceOfDeletedMessageConflictsWithoutContent(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	nonce := "deleted-canonical-0001"
	deletedAt := time.Now().UTC()
	deleted := messageDocumentForRequest(uuid.NewString(), channelID, senderID, simpleSendRequest(nonce))
	deleted.DeletedAt = &deletedAt
	repo.messages[deleted.ID] = deleted
	repo.messagesBySenderNonce[messageSenderNonceKey(senderID, nonce)] = deleted.ID
	service := messenger.ExportNewMessengerServiceTest(repo)

	response, err := service.SendMessage(ctx, channelID, senderID, simpleSendRequest(nonce))

	assert.Nil(t, response)
	requireAppError(t, err, 409, "client_nonce уже принадлежит удалённому сообщению")
	assert.Zero(t, repo.createMessageCalls)
	assert.Len(t, repo.messages, 1)
}

func TestSendMessageCreatedFalseSuppressesEverySideEffect(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	root := addLiveThreadRoot(repo, channelID, senderID)
	req := simpleSendRequest("created-false-0001")
	req.ThreadID = root.ID
	canonical := messageDocumentForRequest("canonical-created-false", channelID, senderID, req)
	repo.createMessageResultOverride = &messenger.CreateMessageResult{Message: canonical, Created: false}
	service := messenger.ExportNewMessengerServiceTest(repo)

	response, err := service.SendMessage(ctx, channelID, senderID, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, canonical.ID, response.ID)
	assert.Equal(t, 1, repo.createMessageCalls)
	assert.Len(t, repo.messages, 1)
	assert.Zero(t, root.ReplyCount)
}

func TestSendMessageCreatedTrueDeletedCanonicalSuppressesSideEffects(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	root := addLiveThreadRoot(repo, channelID, senderID)
	req := simpleSendRequest("created-true-deleted-0001")
	req.ThreadID = root.ID
	canonical := messageDocumentForRequest("deleted-created-true", channelID, senderID, req)
	deletedAt := time.Now().UTC()
	canonical.DeletedAt = &deletedAt
	repo.createMessageResultOverride = &messenger.CreateMessageResult{Message: canonical, Created: true}
	service := messenger.ExportNewMessengerServiceTest(repo)

	response, err := service.SendMessage(ctx, channelID, senderID, req)

	assert.Nil(t, response)
	requireAppError(t, err, 409, "client_nonce уже принадлежит удалённому сообщению")
	assert.Equal(t, 1, repo.createMessageCalls)
	assert.Len(t, repo.messages, 1)
	assert.Zero(t, root.ReplyCount)
}

func TestSendMessageWithoutNonceUsesTransitionalCreate(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	service := messenger.ExportNewMessengerServiceTest(repo)
	req := simpleSendRequest("")

	first, err := service.SendMessage(ctx, channelID, senderID, req)
	require.NoError(t, err)
	second, err := service.SendMessage(ctx, channelID, senderID, req)

	require.NoError(t, err)
	assert.NotEqual(t, first.ID, second.ID)
	assert.Equal(t, 2, repo.createMessageCalls)
	assert.Len(t, repo.messages, 2)
}

func TestSendMessageAuthorizationPrecedesClientNonceLookup(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, uuid.New())
	nonce := "unauthorized-nonce-0001"
	canonical := messageDocumentForRequest("protected-canonical", channelID, senderID, simpleSendRequest(nonce))
	repo.messages[canonical.ID] = canonical
	repo.messagesBySenderNonce[messageSenderNonceKey(senderID, nonce)] = canonical.ID
	service := messenger.ExportNewMessengerServiceTest(repo)

	response, err := service.SendMessage(ctx, channelID, senderID, simpleSendRequest(nonce))

	assert.Nil(t, response)
	requireAppErrorCode(t, err, 403)
	assert.Zero(t, repo.createMessageCalls)
	assert.Len(t, repo.messages, 1)
}

func simpleSendRequest(nonce string) messenger.SendMessageRequest {
	return messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{"type": "text", "body": "message", "count": float64(1)},
		ContentText: "message",
		ClientNonce: nonce,
	}
}

func richSendRequest(nonce string) messenger.SendMessageRequest {
	return messenger.SendMessageRequest{
		ContentJSON:     map[string]interface{}{"type": "text", "body": "message", "count": float64(1)},
		ContentText:     "message",
		ClientNonce:     nonce,
		Attachments:     []messenger.AttachmentRef{{FileID: "file-1", FileName: "one.pdf", FileSize: 10, MimeType: "application/pdf", URL: "https://example.test/one.pdf"}, {FileID: "file-2", FileName: "two.png", FileSize: 20, MimeType: "image/png", URL: "https://example.test/two.png"}},
		Mentions:        []string{"user-1", "user-2", "user-1"},
		MentionGroups:   []string{"all", "online", "all"},
		VoiceURL:        "https://example.test/message.ogg",
		VoiceDurationMs: 1000,
	}
}

func addLiveThreadRoot(repo *mockRepository, channelID, senderID uuid.UUID) *messenger.MessageDocument {
	root := &messenger.MessageDocument{
		ID:          uuid.NewString(),
		ChannelID:   channelID,
		SenderID:    senderID,
		ContentJSON: map[string]interface{}{"type": "text", "body": "root"},
		ContentText: "root",
		CreatedAt:   time.Now().UTC(),
	}
	repo.messages[root.ID] = root
	return root
}

func messageDocumentForRequest(id string, channelID, senderID uuid.UUID, req messenger.SendMessageRequest) *messenger.MessageDocument {
	return &messenger.MessageDocument{
		ID:              id,
		ChannelID:       channelID,
		SenderID:        senderID,
		ContentJSON:     req.ContentJSON,
		ContentText:     req.ContentText,
		ThreadID:        req.ThreadID,
		ClientNonce:     req.ClientNonce,
		Attachments:     req.Attachments,
		Mentions:        req.Mentions,
		MentionGroups:   req.MentionGroups,
		VoiceURL:        req.VoiceURL,
		VoiceDurationMs: req.VoiceDurationMs,
		CreatedAt:       time.Now().UTC(),
	}
}

func requireAppError(t *testing.T, err error, expectedCode int, expectedMessage string) {
	t.Helper()
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok, "expected AppError, got %T", err)
	assert.Equal(t, expectedCode, appErr.Code)
	assert.Equal(t, expectedMessage, appErr.Message)
}