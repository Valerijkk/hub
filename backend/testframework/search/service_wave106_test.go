package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A2W5-A5-006: sqlite Empty без entity tables — НЕ NF-03 proof.
// После A2W5-A4-003 missing-table shards → ErrSearchUnavailable.

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
	// Без message: nil-mongo shard = OK и маскирует total outage.
	_, _, err = svc.Search(context.Background(), "hub", userID,
		[]string{"issue", "test_case", "user", "project", "file", "channel"}, 50)
	require.Error(t, err)
	assert.ErrorIs(t, err, search.ErrSearchUnavailable)
}

// A3W5-5-001: бывший TestSearch_LimitCappedAtFifty — имя врало про clamp;
// поведение = Unavailable на missing issues. Clamp → limit_clamp_a3w55_test.go.
func TestSearch_MissingIssuesTable_Unavailable(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	projectID := uuid.New()
	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?)`,
		projectID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	_, _, err = svc.Search(context.Background(), "x", userID, []string{"issue"}, 100)
	require.Error(t, err)
	assert.ErrorIs(t, err, search.ErrSearchUnavailable)
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

// A3W5-5-002: канон closed-DB GetUserChannelIDs (дубль shard_errors удалён).
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

// A3W5-5-002: дубль nil-mongo WithMembershipNoMongo снят —
// канон TestSearchMessages_NilMongoReturnsEmpty (service_scope_test.go).
