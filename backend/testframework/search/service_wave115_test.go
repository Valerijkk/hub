package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch_UserShardOnly_Admin(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	svc := search.NewSearchService(&adminSearchRepo{admin: true}, db, nil)

	results, _, err := svc.Search(context.Background(), "alice", userID, []string{"user"}, 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearch_FileShardOnly(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	projectID := uuid.New()
	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?)`,
		projectID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	results, _, err := svc.Search(context.Background(), "readme", userID, []string{"file"}, 5)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearch_TestCaseShardOnly(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	projectID := uuid.New()
	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?)`,
		projectID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	results, _, err := svc.Search(context.Background(), "login", userID, []string{"test_case"}, 5)
	require.NoError(t, err)
	assert.Empty(t, results)
}
