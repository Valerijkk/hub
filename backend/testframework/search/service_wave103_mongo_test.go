package search_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/messenger"
	"github.com/valerii/hub/backend/internal/search"
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
	_, err = sqlDB.Exec(`
		CREATE TABLE channels (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			channel_type TEXT NOT NULL DEFAULT 'public',
			project_id TEXT,
			deleted_at TEXT
		);
		CREATE TABLE channel_members (channel_id TEXT NOT NULL, user_id TEXT NOT NULL);
	`)
	require.NoError(t, err)
	_, err = sqlDB.Exec(
		`INSERT INTO channels (id, name, channel_type, deleted_at) VALUES (?, 'own', 'public', NULL)`,
		channelID.String(),
	)
	require.NoError(t, err)
	_, err = sqlDB.Exec(
		`INSERT INTO channel_members (channel_id, user_id) VALUES (?, ?)`,
		channelID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, sqlDB, db)
	results, err := search.ExportSearchMessages(svc, ctx, "wave103 searchable", userID)
	require.NoError(t, err)
	// Len==1 — mutation-ready: drop channel_id $in → 2 hits (W5-A5-002).
	require.Len(t, results, 1, "foreign channel must be excluded (NF-03)")
	assert.Equal(t, "message", results[0].Type)
	assert.Contains(t, results[0].Title, "wave103")
	assert.NotContains(t, results[0].Title, "hidden")
}

// TestSearchService_searchMessages_ExcludesSoftDeleted_P1BE32 — W5-A4-003:
// soft-deleted Mongo messages must yield 0 hits. Mutation: drop deleted_at:nil → FAIL.
func TestSearchService_searchMessages_ExcludesSoftDeleted_P1BE32(t *testing.T) {
	db := mongotest.Connect(t, "hub_test_search_")
	ctx := context.Background()

	msgRepo := messenger.NewMongoRepository(db)
	require.NoError(t, msgRepo.EnsureIndexes(ctx))

	userID := uuid.New()
	channelID := uuid.New()
	now := time.Now().UTC()

	liveID, err := msgRepo.InsertMessage(ctx, &messenger.MessageDocument{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "p1be32 softdel live message body",
		CreatedAt:   now,
	})
	require.NoError(t, err)

	deletedID, err := msgRepo.InsertMessage(ctx, &messenger.MessageDocument{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "p1be32 softdel trash message body",
		CreatedAt:   now,
	})
	require.NoError(t, err)
	require.NoError(t, msgRepo.SoftDeleteMessage(ctx, deletedID))

	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	_, err = sqlDB.Exec(`
		CREATE TABLE channels (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			channel_type TEXT NOT NULL DEFAULT 'public',
			project_id TEXT,
			deleted_at TEXT
		);
		CREATE TABLE channel_members (channel_id TEXT NOT NULL, user_id TEXT NOT NULL);
	`)
	require.NoError(t, err)
	_, err = sqlDB.Exec(
		`INSERT INTO channels (id, name, channel_type, deleted_at) VALUES (?, 'ch', 'public', NULL)`,
		channelID.String(),
	)
	require.NoError(t, err)
	_, err = sqlDB.Exec(
		`INSERT INTO channel_members (channel_id, user_id) VALUES (?, ?)`,
		channelID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, sqlDB, db)
	results, err := search.ExportSearchMessages(svc, ctx, "p1be32 softdel", userID)
	require.NoError(t, err)
	require.Len(t, results, 1, "soft-deleted message must be excluded")
	assert.Equal(t, liveID, results[0].ID)
	assert.NotEqual(t, deletedID, results[0].ID)
	assert.Contains(t, results[0].Title, "live")
}

func TestSearchService_searchMessages_NoChannels(t *testing.T) {
	db := mongotest.Connect(t, "hub_test_search_")
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	_, err = sqlDB.Exec(`
		CREATE TABLE channels (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			channel_type TEXT NOT NULL DEFAULT 'public',
			project_id TEXT,
			deleted_at TEXT
		);
		CREATE TABLE channel_members (channel_id TEXT NOT NULL, user_id TEXT NOT NULL);
	`)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, sqlDB, db)
	results, err := search.ExportSearchMessages(svc, context.Background(), "anything", uuid.New())
	require.NoError(t, err)
	assert.Nil(t, results)
}
