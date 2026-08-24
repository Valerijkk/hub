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
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	_ messenger.MessageRepository = (*messenger.MongoRepository)(nil)
	_ messenger.MessageRepository = (*mockRepository)(nil)
)

func TestMessageClientNonceDTOPropagation(t *testing.T) {
	const nonce = "client-nonce-149"

	requestJSON, err := json.Marshal(messenger.SendMessageRequest{ClientNonce: nonce})
	require.NoError(t, err)
	assert.JSONEq(t, `{"content_json":null,"client_nonce":"client-nonce-149"}`, string(requestJSON))

	responseJSON, err := json.Marshal(messenger.MessageResponse{ClientNonce: nonce})
	require.NoError(t, err)
	assert.JSONEq(t, `{"id":"","channel_id":"","sender_id":"","content_json":null,"reply_count":0,"client_nonce":"client-nonce-149","created_at":"0001-01-01T00:00:00Z"}`, string(responseJSON))

	persisted, err := bson.Marshal(messenger.MessageDocument{ClientNonce: nonce})
	require.NoError(t, err)
	var persistedFields bson.M
	require.NoError(t, bson.Unmarshal(persisted, &persistedFields))
	assert.Equal(t, nonce, persistedFields["client_nonce"])
}

func TestMockRepositoryFindMessageByClientNonceIsScopedToSender(t *testing.T) {
	repo := newMockRepository()
	ctx := context.Background()
	nonce := "same-nonce-for-two-senders"
	firstSender := uuid.New()
	secondSender := uuid.New()

	first, err := repo.CreateMessageIdempotent(ctx, &messenger.MessageDocument{
		ChannelID:   uuid.New(),
		SenderID:    firstSender,
		ClientNonce: nonce,
	})
	require.NoError(t, err)
	require.True(t, first.Created)

	second, err := repo.CreateMessageIdempotent(ctx, &messenger.MessageDocument{
		ChannelID:   uuid.New(),
		SenderID:    secondSender,
		ClientNonce: nonce,
	})
	require.NoError(t, err)
	require.True(t, second.Created)

	found, err := repo.FindMessageByClientNonce(ctx, firstSender, nonce)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, first.Message.ID, found.ID)

	notFound, err := repo.FindMessageByClientNonce(ctx, uuid.New(), nonce)
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestMockRepositoryCreateMessageIdempotentWithEmptyNonceUsesTransitionalInsert(t *testing.T) {
	repo := newMockRepository()
	doc := &messenger.MessageDocument{
		ChannelID: uuid.New(),
		SenderID:  uuid.New(),
	}

	result, err := repo.CreateMessageIdempotent(context.Background(), doc)

	require.NoError(t, err)
	require.True(t, result.Created)
	require.Same(t, doc, result.Message)
	assert.NotEmpty(t, doc.ID)
	assert.False(t, doc.CreatedAt.IsZero())
	assert.WithinDuration(t, time.Now().UTC(), doc.CreatedAt, time.Second)
	assert.Len(t, repo.messages, 1)
	assert.Equal(t, 1, repo.createMessageCalls)
}

func TestMockRepositoryCreateMessageResultOverrideDoesNotPersist(t *testing.T) {
	repo := newMockRepository()
	doc := &messenger.MessageDocument{
		ChannelID:   uuid.New(),
		SenderID:    uuid.New(),
		ClientNonce: "override-client-nonce-149",
	}
	override := &messenger.CreateMessageResult{
		Message: &messenger.MessageDocument{ID: "canonical-override"},
		Created: false,
	}
	repo.createMessageResultOverride = override

	result, err := repo.CreateMessageIdempotent(context.Background(), doc)

	require.NoError(t, err)
	assert.Same(t, override, result)
	assert.Equal(t, 1, repo.createMessageCalls)
	assert.Empty(t, doc.ID)
	assert.True(t, doc.CreatedAt.IsZero())
	assert.Empty(t, repo.messages)
	found, err := repo.FindMessageByClientNonce(context.Background(), doc.SenderID, doc.ClientNonce)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestMockRepositoryCreateMessageIdempotentReturnsOneCanonicalMessageConcurrently(t *testing.T) {
	const attempts = 12

	repo := newMockRepository()
	ctx := context.Background()
	senderID := uuid.New()
	channelID := uuid.New()
	nonce := "concurrent-client-nonce-149"
	results := make(chan *messenger.CreateMessageResult, attempts)
	errors := make(chan error, attempts)
	start := make(chan struct{})
	var workers sync.WaitGroup

	for range attempts {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			result, err := repo.CreateMessageIdempotent(ctx, &messenger.MessageDocument{
				ChannelID:   channelID,
				SenderID:    senderID,
				ClientNonce: nonce,
			})
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}()
	}

	close(start)
	workers.Wait()
	close(results)
	close(errors)

	for err := range errors {
		require.NoError(t, err)
	}

	canonicalIDs := make(map[string]struct{})
	createdCount := 0
	for result := range results {
		require.NotNil(t, result)
		require.NotNil(t, result.Message)
		canonicalIDs[result.Message.ID] = struct{}{}
		if result.Created {
			createdCount++
		}
	}

	assert.Len(t, canonicalIDs, 1)
	assert.Equal(t, 1, createdCount)
	assert.Len(t, repo.messages, 1)
	assert.Equal(t, attempts, repo.createMessageCalls)
}

func TestMockRepositoryMessageMapAccessIsSynchronized(t *testing.T) {
	const attempts = 24

	repo := newMockRepository()
	ctx := context.Background()
	parent := &messenger.MessageDocument{
		ID:        "seeded-parent",
		ChannelID: uuid.New(),
		SenderID:  uuid.New(),
	}
	_, err := repo.InsertMessage(ctx, parent)
	require.NoError(t, err)

	start := make(chan struct{})
	failures := make(chan string, attempts*3)
	var workers sync.WaitGroup
	for range attempts {
		workers.Add(3)
		go func() {
			defer workers.Done()
			<-start
			result, err := repo.CreateMessageIdempotent(ctx, &messenger.MessageDocument{
				ChannelID:   parent.ChannelID,
				SenderID:    parent.SenderID,
				ClientNonce: "shared-client-nonce-149",
			})
			if err != nil || result == nil || result.Message == nil {
				failures <- "idempotent create did not return a canonical message"
			}
		}()
		go func() {
			defer workers.Done()
			<-start
			doc, err := repo.FindMessageByID(ctx, parent.ID)
			if err != nil || doc == nil || doc.ID != parent.ID {
				failures <- "find by ID did not return the seeded parent"
			}
		}()
		go func() {
			defer workers.Done()
			<-start
			if err := repo.IncrementReplyCount(ctx, parent.ID); err != nil {
				failures <- "increment reply count failed"
			}
		}()
	}

	close(start)
	workers.Wait()
	close(failures)
	for failure := range failures {
		assert.Fail(t, failure)
	}

	parentAfter, err := repo.FindMessageByID(ctx, parent.ID)
	require.NoError(t, err)
	assert.Equal(t, attempts, parentAfter.ReplyCount)
	canonical, err := repo.FindMessageByClientNonce(ctx, parent.SenderID, "shared-client-nonce-149")
	require.NoError(t, err)
	require.NotNil(t, canonical)
	assert.Equal(t, attempts, repo.createMessageCalls)
}
