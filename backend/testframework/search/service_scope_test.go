package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
)

// mockSearchRepo реализует search.Repository для unit-тестов области видимости.
type mockSearchRepo struct {
	isAdmin  bool
	adminErr error
	// memberOK: nil = allow (legacy tests); иначе *memberOK.
	memberOK  *bool
	memberErr error
}

func (m *mockSearchRepo) IsSystemAdmin(_ context.Context, _ uuid.UUID) (bool, error) {
	if m.adminErr != nil {
		return false, m.adminErr
	}
	return m.isAdmin, nil
}

func (m *mockSearchRepo) IsProjectMember(_ context.Context, _, _ uuid.UUID) (bool, error) {
	if m.memberErr != nil {
		return false, m.memberErr
	}
	if m.memberOK != nil {
		return *m.memberOK, nil
	}
	return true, nil
}

func (m *mockSearchRepo) CreateSavedFilter(context.Context, search.CreateSavedFilterInput) (*ent.SavedFilter, error) {
	return nil, nil
}
func (m *mockSearchRepo) ListSavedFilters(context.Context, uuid.UUID, uuid.UUID) ([]*ent.SavedFilter, error) {
	return nil, nil
}
func (m *mockSearchRepo) GetSavedFilter(context.Context, uuid.UUID) (*ent.SavedFilter, error) {
	return nil, nil
}
func (m *mockSearchRepo) DeleteSavedFilter(context.Context, uuid.UUID) error { return nil }
func (m *mockSearchRepo) UpdateSavedFilterShared(context.Context, uuid.UUID, bool) (*ent.SavedFilter, error) {
	return nil, nil
}

func newScopeTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`
		CREATE TABLE channels (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			channel_type TEXT NOT NULL DEFAULT 'public',
			project_id TEXT,
			deleted_at TEXT
		);
		CREATE TABLE channel_members (channel_id TEXT NOT NULL, user_id TEXT NOT NULL);
		CREATE TABLE project_members (project_members TEXT NOT NULL, user_project_members TEXT NOT NULL);
	`)
	require.NoError(t, err)
	return db
}

func TestIsSystemAdmin(t *testing.T) {
	svc := search.ExportNewSearchServiceRepoOnly(&mockSearchRepo{isAdmin: true})
	assert.True(t, search.ExportIsSystemAdmin(svc, context.Background(), uuid.New()))

	search.ExportSetSearchServiceRepo(svc, &mockSearchRepo{isAdmin: false})
	assert.False(t, search.ExportIsSystemAdmin(svc, context.Background(), uuid.New()))

	search.ExportSetSearchServiceRepo(svc, nil)
	assert.False(t, search.ExportIsSystemAdmin(svc, context.Background(), uuid.New()))

	search.ExportSetSearchServiceRepo(svc, &mockSearchRepo{adminErr: assert.AnError})
	assert.False(t, search.ExportIsSystemAdmin(svc, context.Background(), uuid.New()))
}

func TestGetUserChannelIDs_ReturnsOnlyUserMemberships(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	ownChannel := uuid.New()
	otherChannel := uuid.New()
	deletedChannel := uuid.New()

	_, err := db.Exec(
		`INSERT INTO channels (id, name, channel_type, deleted_at) VALUES (?, 'own', 'public', NULL), (?, 'other', 'public', NULL), (?, 'gone', 'public', '2026-01-01')`,
		ownChannel.String(), otherChannel.String(), deletedChannel.String(),
	)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO channel_members (channel_id, user_id) VALUES (?, ?), (?, ?), (?, ?)`,
		ownChannel.String(), userID.String(),
		otherChannel.String(), uuid.New().String(),
		deletedChannel.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.ExportNewSearchServiceSQL(db)
	ids, err := search.ExportGetUserChannelIDs(svc, context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, ids, 1, "soft-deleted channel membership must be excluded")
	assert.Equal(t, ownChannel, ids[0])
}

func TestGetUserProjectIDs_ReturnsOnlyUserMemberships(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	projectID := uuid.New()
	otherProject := uuid.New()

	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?), (?, ?)`,
		projectID.String(), userID.String(),
		otherProject.String(), uuid.New().String(),
	)
	require.NoError(t, err)

	svc := search.ExportNewSearchServiceSQL(db)
	ids, err := search.ExportGetUserProjectIDs(svc, context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, ids, 1)
	assert.Equal(t, projectID, ids[0])
}

func TestSearchScopedEntities_EmptyProjectIDsReturnNothing(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	ctx := context.Background()

	issues, err := search.ExportSearchIssues(svc, ctx, "test", nil)
	require.NoError(t, err)
	assert.Nil(t, issues)

	issues, err = search.ExportSearchIssues(svc, ctx, "test", []uuid.UUID{})
	require.NoError(t, err)
	assert.Nil(t, issues)

	cases, err := search.ExportSearchTestCases(svc, ctx, "test", nil)
	require.NoError(t, err)
	assert.Nil(t, cases)

	projects, err := search.ExportSearchProjects(svc, ctx, "test", nil)
	require.NoError(t, err)
	assert.Nil(t, projects)

	files, err := search.ExportSearchFiles(svc, ctx, "test", nil)
	require.NoError(t, err)
	assert.Nil(t, files)

	channels, err := search.ExportSearchChannels(svc, ctx, "test", nil, uuid.Nil)
	require.NoError(t, err)
	assert.Nil(t, channels)
}

func TestSearchMessages_NilMongoReturnsEmpty(t *testing.T) {
	svc := search.ExportNewSearchServiceFull(nil, newScopeTestDB(t), nil)
	results, err := search.ExportSearchMessages(svc, context.Background(), "hello", uuid.New())
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearchMessages_NoChannelsReturnsEmpty(t *testing.T) {
	db := newScopeTestDB(t)
	svc := search.ExportNewSearchServiceFull(nil, db, nil)
	userID := uuid.New()

	results, err := search.ExportSearchMessages(svc, context.Background(), "hello", userID)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearch_UsesProjectScopeForIssues(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	allowedProject := uuid.New()

	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?)`,
		allowedProject.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.ExportNewSearchServiceFull(&mockSearchRepo{}, db, nil)
	results, err := svc.Search(context.Background(), "hub", userID, []string{"issue"}, 20)
	require.NoError(t, err)
	assert.Empty(t, results)
}
