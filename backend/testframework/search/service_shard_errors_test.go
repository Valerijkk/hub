package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	results, err := svc.Search(context.Background(), "hub query", userID, nil, 20)
	require.NoError(t, err)
	assert.Empty(t, results)
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

func TestDeleteSavedFilter_NotFound(t *testing.T) {
	svc, _, ctx, userID, _ := newSearchServiceTestEnv(t)
	err := svc.DeleteSavedFilter(ctx, userID, uuid.New())
	require.Error(t, err)
}
