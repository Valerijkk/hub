package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A2W5-A5-006 / A2W5-A4-003: missing entity tables → total outage error, не empty success.

func TestSearch_AllShardsWithMembership_RecoversFromMissingTables(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	projectID := uuid.New()
	channelID := uuid.New()

	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?)`,
		projectID.String(), userID.String(),
	)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO channel_members (channel_id, user_id) VALUES (?, ?)`,
		channelID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	// searchAll включает message (nil mongo = OK) — исключаем, чтобы проверить total outage.
	_, _, err = svc.Search(context.Background(), "hub query", userID,
		[]string{"issue", "test_case", "user", "project", "file", "channel"}, 20)
	require.Error(t, err)
	assert.ErrorIs(t, err, search.ErrSearchUnavailable)
}

func TestGetUserChannelIDs_DBClosed(t *testing.T) {
	db := newScopeTestDB(t)
	require.NoError(t, db.Close())
	svc := search.ExportNewSearchServiceSQL(db)
	_, err := search.ExportGetUserChannelIDs(svc, context.Background(), uuid.New())
	require.Error(t, err)
}

func TestSearchMessages_NoChannelMembership(t *testing.T) {
	db := newScopeTestDB(t)
	svc := search.ExportNewSearchServiceFull(nil, db, nil)
	results, err := search.ExportSearchMessages(svc, context.Background(), "hello", uuid.New())
	require.NoError(t, err)
	assert.Nil(t, results)
}
