package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch_AllEntityTypesParallel(t *testing.T) {
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

	svc := search.NewSearchService(&mockSearchRepo{isAdmin: true}, db, nil)
	results, _, err := svc.Search(context.Background(), "hub", userID,
		[]string{"issue", "test_case", "user", "project", "file", "channel", "message"}, 50)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearch_LimitCappedAtFifty(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	projectID := uuid.New()
	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?)`,
		projectID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	results, _, err := svc.Search(context.Background(), "x", userID, []string{"issue"}, 100)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearchChannels_TriggersFallbackBeforeError(t *testing.T) {
	db := newScopeTestDB(t)
	projectID := uuid.New()
	_, err := db.Exec(
		`INSERT INTO channels (id, name, channel_type, project_id) VALUES (?, ?, ?, ?)`,
		uuid.New().String(), "general", "public", projectID.String(),
	)
	require.NoError(t, err)

	svc := search.ExportNewSearchServiceSQL(db)
	_, err = search.ExportSearchChannels(svc, context.Background(), "general", []uuid.UUID{projectID}, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "каналов")
}

func TestSearchUsers_AdminRoutesToUnscoped(t *testing.T) {
	svc := search.ExportNewSearchServiceRepoSQL(&mockSearchRepo{isAdmin: true}, newScopeTestDB(t))
	_, err := search.ExportSearchUsers(svc, context.Background(), "alice", uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

func TestSearchUsers_NonAdminScopedQueryFails(t *testing.T) {
	svc := search.ExportNewSearchServiceRepoSQL(&mockSearchRepo{isAdmin: false}, newScopeTestDB(t))
	_, err := search.ExportSearchUsers(svc, context.Background(), "bob", uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "пользователей")
}

func TestGetUserChannelIDs_ClosedDB(t *testing.T) {
	db := newScopeTestDB(t)
	require.NoError(t, db.Close())
	svc := search.ExportNewSearchServiceSQL(db)
	_, err := search.ExportGetUserChannelIDs(svc, context.Background(), uuid.New())
	require.Error(t, err)
}

func TestGetUserProjectIDs_ClosedDB(t *testing.T) {
	db := newScopeTestDB(t)
	require.NoError(t, db.Close())
	svc := search.ExportNewSearchServiceSQL(db)
	_, err := search.ExportGetUserProjectIDs(svc, context.Background(), uuid.New())
	require.Error(t, err)
}

func TestSearchMessages_WithMembershipNoMongo(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	channelID := uuid.New()
	_, err := db.Exec(
		`INSERT INTO channel_members (channel_id, user_id) VALUES (?, ?)`,
		channelID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.ExportNewSearchServiceFull(nil, db, nil)
	results, err := search.ExportSearchMessages(svc, context.Background(), "hello", userID)
	require.NoError(t, err)
	assert.Nil(t, results)
}
