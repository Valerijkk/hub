package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch_DefaultLimitWhenZero(t *testing.T) {
	// A2W5-A5-006: без entity tables issue shard fails → total outage.
	db := newScopeTestDB(t)
	userID := uuid.New()
	projectID := uuid.New()
	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?)`,
		projectID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	_, _, err = svc.Search(context.Background(), "query", userID, []string{"issue"}, 0)
	require.Error(t, err)
	assert.ErrorIs(t, err, search.ErrSearchUnavailable)
}

func TestSearch_ProjectIDsFailure(t *testing.T) {
	db := newScopeTestDB(t)
	require.NoError(t, db.Close())

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	_, _, err := svc.Search(context.Background(), "x", uuid.New(), nil, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "проекты")
}
