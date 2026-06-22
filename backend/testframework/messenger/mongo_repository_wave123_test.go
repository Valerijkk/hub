package messenger_test

// Тесты репозитория мессенджера (волна 123): интеграционные и unit-сценарии testframework.
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


// TestMongoRepository_Wave123_ListMessagesWithCursor проверяет слой репозитория: mongo репозиторий wave 123 список сообщения with cursor.
func TestMongoRepository_Wave123_ListMessagesWithCursor(t *testing.T) {
	db := mongotest.Connect(t, "hub_test_msg123_")
	repo := messenger.NewMongoRepository(db)
	ctx := context.Background()
	channelID := uuid.New()
	senderID := uuid.New()
	require.NoError(t, repo.EnsureIndexes(ctx))

	t1 := time.Now().UTC().Add(-2 * time.Minute)
	t2 := time.Now().UTC().Add(-1 * time.Minute)
	for i, ts := range []time.Time{t1, t2} {
		_, err := repo.InsertMessage(ctx, &messenger.MessageDocument{
			ChannelID: channelID, SenderID: senderID,
			ContentText: "msg " + string(rune('a'+i)),
			ContentJSON: map[string]interface{}{"text": "msg"},
			CreatedAt:   ts,
		})
		require.NoError(t, err)
	}

	all, err := repo.ListMessages(ctx, channelID, nil, 10)
	require.NoError(t, err)
	require.Len(t, all, 2)

	before := all[0].CreatedAt
	page, err := repo.ListMessages(ctx, channelID, &before, 10)
	require.NoError(t, err)
	assert.NotEmpty(t, page)
}

// TestMongoRepository_Wave123_FindMessageNotFound проверяет слой репозитория: mongo репозиторий wave 123 find message not found.
func TestMongoRepository_Wave123_FindMessageNotFound(t *testing.T) {
	db := mongotest.Connect(t, "hub_test_msg123_")
	repo := messenger.NewMongoRepository(db)
	_, err := repo.FindMessageByID(context.Background(), uuid.New().String())
	require.Error(t, err)
}

// TestMongoRepository_Wave123_SearchWithoutChannelFilter проверяет слой репозитория: mongo репозиторий wave 123 search without channel filter.
func TestMongoRepository_Wave123_SearchWithoutChannelFilter(t *testing.T) {
	db := mongotest.Connect(t, "hub_test_msg123_")
	repo := messenger.NewMongoRepository(db)
	ctx := context.Background()
	require.NoError(t, repo.EnsureIndexes(ctx))
	channelID := uuid.New()
	_, err := repo.InsertMessage(ctx, &messenger.MessageDocument{
		ChannelID: channelID, SenderID: uuid.New(),
		ContentText: "unique-search-token-123",
		ContentJSON: map[string]interface{}{"text": "unique-search-token-123"},
		CreatedAt:   time.Now().UTC(),
	})
	require.NoError(t, err)

	hits, err := repo.SearchMessages(ctx, "unique-search-token", nil, 5)
	require.NoError(t, err)
	assert.NotEmpty(t, hits)
}
