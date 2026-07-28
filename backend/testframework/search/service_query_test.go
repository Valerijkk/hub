package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch_NoProjectMemberships_ReturnsEmpty(t *testing.T) {
	db := newScopeTestDB(t)
	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	userID := uuid.New()

	results, _, err := svc.Search(context.Background(), "login", userID, nil, 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearch_TypeFilterIssueOnly(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	projectID := uuid.New()
	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?)`,
		projectID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	results, _, err := svc.Search(context.Background(), "bug", userID, []string{"issue"}, 5)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearch_LimitClampedToDefault(t *testing.T) {
	db := newScopeTestDB(t)
	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	userID := uuid.New()

	results, _, err := svc.Search(context.Background(), "x", userID, []string{"user"}, 0)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearch_MessageTypeWithoutMongo(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	projectID := uuid.New()
	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?)`,
		projectID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	results, _, err := svc.Search(context.Background(), "hello", userID, []string{"message"}, 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearch_ProjectIDsLookupError(t *testing.T) {
	db := newScopeTestDB(t)
	_ = db.Close()
	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	_, _, err := svc.Search(context.Background(), "q", uuid.New(), nil, 10)
	require.Error(t, err)
}
