package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToEntPredicates_StatusOperators(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		jql     string
		wantLen int
	}{
		{`status = todo`, 1},
		{`status != done`, 1},
		{`status IN (todo, done)`, 1},
		{`status NOT IN (blocked)`, 1},
	}

	for _, tt := range tests {
		q, err := search.ParseJQL(tt.jql)
		require.NoError(t, err, tt.jql)
		preds := search.ToEntPredicates(q, userID)
		assert.Len(t, preds, tt.wantLen, tt.jql)
	}
}

func TestToEntPredicates_PriorityAndType(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		jql     string
		wantLen int
	}{
		{`priority = high`, 1},
		{`priority != low`, 1},
		{`priority IN (high, critical)`, 1},
		{`type = bug`, 1},
		{`type != epic`, 1},
		{`type IN (bug, task)`, 1},
	}

	for _, tt := range tests {
		q, err := search.ParseJQL(tt.jql)
		require.NoError(t, err, tt.jql)
		preds := search.ToEntPredicates(q, userID)
		assert.Len(t, preds, tt.wantLen, tt.jql)
	}
}

func TestToEntPredicates_AssigneeCurrentUser(t *testing.T) {
	userID := uuid.New()
	q, err := search.ParseJQL(`assignee = currentUser()`)
	require.NoError(t, err)
	preds := search.ToEntPredicates(q, userID)
	require.Len(t, preds, 1)
}

func TestToEntPredicates_UpdatedDateOps(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		jql     string
		wantLen int
	}{
		{`updated = 2024-06-01`, 1},
		{`updated > 2024-01-01`, 1},
		{`updated < 2024-12-31`, 1},
		{`updated >= 2024-06-01`, 1},
		{`updated <= 2024-06-30`, 1},
	}

	for _, tt := range tests {
		q, err := search.ParseJQL(tt.jql)
		require.NoError(t, err, tt.jql)
		preds := search.ToEntPredicates(q, userID)
		assert.Len(t, preds, tt.wantLen, tt.jql)
	}
}

func TestToEntPredicates_UnknownFieldReturnsNil(t *testing.T) {
	q, err := search.ParseJQL(`unknown_field = value`)
	require.NoError(t, err)
	preds := search.ToEntPredicates(q, uuid.New())
	assert.Empty(t, preds)
}

func TestDateFieldPredicate_InvalidDate(t *testing.T) {
	assert.Nil(t, search.ExportDateFieldPredicate("not-a-date", "=", "created"))
}
