package messenger_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/messenger"
	"github.com/valerii/hub/backend/testframework/mongotest"
)

// TestMongoRepository_SearchEmptyChannelIDs_ReturnsEmpty_A2_001 — пустой scope не
// должен открывать org-wide $text. Mutation: удалить early-return len==0 → FAIL.
func TestMongoRepository_SearchEmptyChannelIDs_ReturnsEmpty_A2_001(t *testing.T) {
	db := mongotest.Connect(t, "hub_test_a2001_")
	repo := messenger.NewMongoRepository(db)
	ctx := context.Background()
	require.NoError(t, repo.EnsureIndexes(ctx))

	channelID := uuid.New()
	_, err := repo.InsertMessage(ctx, &messenger.MessageDocument{
		ChannelID:   channelID,
		SenderID:    uuid.New(),
		ContentText: "secret-token-a2001-empty-scope",
		ContentJSON: map[string]interface{}{"text": "secret-token-a2001-empty-scope"},
		CreatedAt:   time.Now().UTC(),
	})
	require.NoError(t, err)

	hits, err := repo.SearchMessages(ctx, "secret-token-a2001-empty-scope", nil, 10)
	require.NoError(t, err)
	assert.Empty(t, hits, "empty channelIDs must not return messages")

	hits, err = repo.SearchMessages(ctx, "secret-token-a2001-empty-scope", []uuid.UUID{}, 10)
	require.NoError(t, err)
	assert.Empty(t, hits)

	hits, err = repo.SearchMessages(ctx, "secret-token-a2001-empty-scope", []uuid.UUID{channelID}, 10)
	require.NoError(t, err)
	assert.NotEmpty(t, hits)
}
