package messenger_test

// Тесты репозитория мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/testframework/mongotest"
)

// TestMongoRepository_MessageLifecycle проверяет слой репозитория: mongo репозиторий message жизненный цикл.
func TestMongoRepository_MessageLifecycle(t *testing.T) {
	db := mongotest.Connect(t, "hub_test_msg_")
	repo := messenger.NewMongoRepository(db)
	ctx := context.Background()
	channelID := uuid.New()
	senderID := uuid.New()

	require.NoError(t, repo.EnsureIndexes(ctx))

	now := time.Now().UTC().Truncate(time.Millisecond)
	doc := &messenger.MessageDocument{
		ChannelID:   channelID,
		SenderID:    senderID,
		ContentJSON: map[string]interface{}{"text": "hello"},
		ContentText: "hello world searchable",
		CreatedAt:   now,
	}
	msgID, err := repo.InsertMessage(ctx, doc)
	require.NoError(t, err)
	require.NotEmpty(t, msgID)

	found, err := repo.FindMessageByID(ctx, msgID)
	require.NoError(t, err)
	assert.Equal(t, "hello world searchable", found.ContentText)

	updated, err := repo.UpdateMessage(ctx, msgID, map[string]interface{}{"text": "edited"}, "edited text")
	require.NoError(t, err)
	assert.Equal(t, "edited text", updated.ContentText)

	listed, err := repo.ListMessages(ctx, channelID, nil, 10)
	require.NoError(t, err)
	require.Len(t, listed, 1)

	require.NoError(t, repo.IncrementReplyCount(ctx, msgID))
	require.NoError(t, repo.DecrementReplyCount(ctx, msgID))

	userID := uuid.New().String()
	require.NoError(t, repo.AddReaction(ctx, msgID, userID, "??"))
	has, err := repo.FindReaction(ctx, msgID, userID, "??")
	require.NoError(t, err)
	assert.True(t, has)
	require.NoError(t, repo.RemoveReaction(ctx, msgID, userID, "??"))

	threadRoot := msgID
	reply := &messenger.MessageDocument{
		ChannelID:   channelID,
		SenderID:    senderID,
		ThreadID:    threadRoot,
		ContentJSON: map[string]interface{}{"text": "reply"},
		ContentText: "thread reply",
		CreatedAt:   now.Add(time.Second),
	}
	replyID, err := repo.InsertMessage(ctx, reply)
	require.NoError(t, err)

	replies, err := repo.ListThreadReplies(ctx, threadRoot, nil, 10)
	require.NoError(t, err)
	require.Len(t, replies, 1)
	assert.Equal(t, replyID, replies[0].ID)

	searchHits, err := repo.SearchMessages(ctx, "edited", []uuid.UUID{channelID}, 5)
	require.NoError(t, err)
	assert.NotEmpty(t, searchHits)

	require.NoError(t, repo.SoftDeleteMessage(ctx, msgID))
}
