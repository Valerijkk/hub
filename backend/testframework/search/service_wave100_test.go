package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureSearchIndexes_SQLiteFails(t *testing.T) {
	db := newScopeTestDB(t)
	err := search.EnsureSearchIndexes(context.Background(), db)
	require.Error(t, err)
}

func TestSearchUsers_AdminUnscopedQueryFails(t *testing.T) {
	db := newScopeTestDB(t)
	svc := search.ExportNewSearchServiceRepoSQL(&adminSearchRepo{admin: true}, db)
	_, err := search.ExportSearchUsers(svc, context.Background(), "admin-user", uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

func TestSearchChannels_EmptyProjects(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	results, err := search.ExportSearchChannels(svc, context.Background(), "general", nil)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearchIssues_EmptyProjects(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	results, err := search.ExportSearchIssues(svc, context.Background(), "bug", nil)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearchTestCases_EmptyProjects(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	results, err := search.ExportSearchTestCases(svc, context.Background(), "case", nil)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearchProjects_EmptyProjects(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	results, err := search.ExportSearchProjects(svc, context.Background(), "hub", nil)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearchFiles_EmptyProjects(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	results, err := search.ExportSearchFiles(svc, context.Background(), "doc", nil)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func wave100ProjectScope() []uuid.UUID {
	return []uuid.UUID{uuid.New()}
}

func TestSearchChannels_QueryFailsWithProjects(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	_, err := search.ExportSearchChannels(svc, context.Background(), "general", wave100ProjectScope())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "каналов")
}

func TestSearchIssues_QueryFailsWithProjects(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	_, err := search.ExportSearchIssues(svc, context.Background(), "bug", wave100ProjectScope())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "задач")
}

func TestSearchTestCases_QueryFailsWithProjects(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	_, err := search.ExportSearchTestCases(svc, context.Background(), "case", wave100ProjectScope())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "тест")
}

func TestSearchProjects_QueryFailsWithProjects(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	_, err := search.ExportSearchProjects(svc, context.Background(), "hub", wave100ProjectScope())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "проект")
}

func TestSearchFiles_QueryFailsWithProjects(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	_, err := search.ExportSearchFiles(svc, context.Background(), "doc", wave100ProjectScope())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "файл")
}

func TestSearchUsers_ScopedQueryFails(t *testing.T) {
	svc := search.ExportNewSearchServiceRepoSQL(&mockSearchRepo{}, newScopeTestDB(t))
	_, err := search.ExportSearchUsers(svc, context.Background(), "alice", uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "пользователей")
}

func TestUuidSliceToPgArray(t *testing.T) {
	id1, id2 := uuid.New(), uuid.New()
	got := search.ExportUUIDSliceToPgArray([]uuid.UUID{id1, id2})
	assert.Equal(t, "{"+id1.String()+","+id2.String()+"}", got)
	assert.Equal(t, "{}", search.ExportUUIDSliceToPgArray(nil))
}
