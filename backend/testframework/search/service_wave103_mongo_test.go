package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/internal/messenger"
	"github.com/valerii/hub/backend/testframework/mongotest"
)

func TestSearchService_searchMessages_MongoIntegration(t *testing.T) {
	db := mongotest.Connect(t, "hub_test_search_")
	ctx := context.Background()

	msgRepo := messenger.NewMongoRepository(db)
	require.NoError(t, msgRepo.EnsureIndexes(ctx))

	userID := uuid.New()
	channelID := uuid.New()
	otherChannel := uuid.New()

	now := time.Now().UTC()
	_, err := msgRepo.InsertMessage(ctx, &messenger.MessageDocument{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "unique wave103 searchable phrase",
		CreatedAt:   now,
	})
	require.NoError(t, err)
	_, err = msgRepo.InsertMessage(ctx, &messenger.MessageDocument{
		ChannelID:   otherChannel,
		SenderID:    uuid.New(),
		ContentText: "unique wave103 searchable phrase hidden",
		CreatedAt:   now,
	})
	require.NoError(t, err)

	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	_, err = sqlDB.Exec(`CREATE TABLE channel_members (channel_id TEXT NOT NULL, user_id TEXT NOT NULL)`)
	require.NoError(t, err)
	_, err = sqlDB.Exec(
		`INSERT INTO channel_members (channel_id, user_id) VALUES (?, ?)`,
		channelID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, sqlDB, db)
	results, err := search.ExportSearchMessages(svc, ctx, "wave103 searchable", userID)
	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.Equal(t, "message", results[0].Type)
	assert.Contains(t, results[0].Title, "wave103")
}

func TestSearchService_searchMessages_NoChannels(t *testing.T) {
	db := mongotest.Connect(t, "hub_test_search_")
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	_, err = sqlDB.Exec(`CREATE TABLE channel_members (channel_id TEXT NOT NULL, user_id TEXT NOT NULL)`)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, sqlDB, db)
	results, err := search.ExportSearchMessages(svc, context.Background(), "anything", uuid.New())
	require.NoError(t, err)
	assert.Nil(t, results)
}
