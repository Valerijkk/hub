package search

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
)

// mockSearchRepo implements Repository for scope-related unit tests.
type mockSearchRepo struct {
	isAdmin bool
	adminErr error
}

func (m *mockSearchRepo) IsSystemAdmin(_ context.Context, _ uuid.UUID) (bool, error) {
	if m.adminErr != nil {
		return false, m.adminErr
	}
	return m.isAdmin, nil
}

func (m *mockSearchRepo) CreateSavedFilter(context.Context, CreateSavedFilterInput) (*ent.SavedFilter, error) {
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
		CREATE TABLE channel_members (channel_id TEXT NOT NULL, user_id TEXT NOT NULL);
		CREATE TABLE project_members (project_members TEXT NOT NULL, user_project_members TEXT NOT NULL);
	`)
	require.NoError(t, err)
	return db
}

func TestIsSystemAdmin(t *testing.T) {
	svc := &SearchService{repo: &mockSearchRepo{isAdmin: true}}
	assert.True(t, svc.isSystemAdmin(context.Background(), uuid.New()))

	svc.repo = &mockSearchRepo{isAdmin: false}
	assert.False(t, svc.isSystemAdmin(context.Background(), uuid.New()))

	svc.repo = nil
	assert.False(t, svc.isSystemAdmin(context.Background(), uuid.New()))

	svc.repo = &mockSearchRepo{adminErr: assert.AnError}
	assert.False(t, svc.isSystemAdmin(context.Background(), uuid.New()))
}

func TestGetUserChannelIDs_ReturnsOnlyUserMemberships(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	ownChannel := uuid.New()
	otherChannel := uuid.New()

	_, err := db.Exec(
		`INSERT INTO channel_members (channel_id, user_id) VALUES (?, ?), (?, ?)`,
		ownChannel.String(), userID.String(),
		otherChannel.String(), uuid.New().String(),
	)
	require.NoError(t, err)

	svc := &SearchService{sqlDB: db}
	ids, err := svc.getUserChannelIDs(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, ids, 1)
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

	svc := &SearchService{sqlDB: db}
	ids, err := svc.getUserProjectIDs(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, ids, 1)
	assert.Equal(t, projectID, ids[0])
}

func TestSearchScopedEntities_EmptyProjectIDsReturnNothing(t *testing.T) {
	svc := &SearchService{sqlDB: newScopeTestDB(t)}
	ctx := context.Background()

	issues, err := svc.searchIssues(ctx, "test", nil)
	require.NoError(t, err)
	assert.Nil(t, issues)

	issues, err = svc.searchIssues(ctx, "test", []uuid.UUID{})
	require.NoError(t, err)
	assert.Nil(t, issues)

	cases, err := svc.searchTestCases(ctx, "test", nil)
	require.NoError(t, err)
	assert.Nil(t, cases)

	projects, err := svc.searchProjects(ctx, "test", nil)
	require.NoError(t, err)
	assert.Nil(t, projects)

	files, err := svc.searchFiles(ctx, "test", nil)
	require.NoError(t, err)
	assert.Nil(t, files)

	channels, err := svc.searchChannels(ctx, "test", nil)
	require.NoError(t, err)
	assert.Nil(t, channels)
}

func TestSearchMessages_NilMongoReturnsEmpty(t *testing.T) {
	svc := &SearchService{sqlDB: newScopeTestDB(t), mongoDB: nil}
	results, err := svc.searchMessages(context.Background(), "hello", uuid.New())
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearchMessages_NoChannelsReturnsEmpty(t *testing.T) {
	db := newScopeTestDB(t)
	svc := &SearchService{sqlDB: db, mongoDB: nil}
	userID := uuid.New()

	results, err := svc.searchMessages(context.Background(), "hello", userID)
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

	svc := &SearchService{repo: &mockSearchRepo{}, sqlDB: db, mongoDB: nil}
	results, err := svc.Search(context.Background(), "hub", userID, []string{"issue"}, 20)
	require.NoError(t, err)
	assert.Empty(t, results)
}
