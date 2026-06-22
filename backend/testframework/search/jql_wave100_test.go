package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseJQL_OrderByMissingBy(t *testing.T) {
	_, err := search.ParseJQL(`status = todo ORDER`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "BY")
}

func TestParseJQL_OrderByDesc(t *testing.T) {
	q, err := search.ParseJQL(`status = todo ORDER BY updated DESC`)
	require.NoError(t, err)
	require.NotNil(t, q.OrderBy)
	assert.Equal(t, "updated", q.OrderBy.Field)
	assert.Equal(t, "DESC", q.OrderBy.Direction)
}

func TestParseJQL_AndOrChain(t *testing.T) {
	q, err := search.ParseJQL(`status = todo AND priority = high OR type = bug`)
	require.NoError(t, err)
	require.Len(t, q.Conditions, 3)
	assert.Equal(t, "AND", q.Conditions[1].Logic)
	assert.Equal(t, "OR", q.Conditions[2].Logic)
}

func TestParseJQL_QuotedValueWithSpaces(t *testing.T) {
	q, err := search.ParseJQL(`summary = "fix login bug"`)
	require.NoError(t, err)
	require.Len(t, q.Conditions, 1)
	assert.Equal(t, "fix login bug", q.Conditions[0].Value)
}

func TestToEntPredicates_AssigneeUUID(t *testing.T) {
	userID := uuid.New()
	q, err := search.ParseJQL(`assignee = ` + userID.String())
	require.NoError(t, err)
	preds := search.ToEntPredicates(q, userID)
	require.Len(t, preds, 1)
}

func TestParseJQL_EmptyInputError(t *testing.T) {
	_, err := search.ParseJQL(`   `)
	require.Error(t, err)
}

func TestParseJQL_INValueList(t *testing.T) {
	q, err := search.ParseJQL(`status IN (todo, "in progress", done)`)
	require.NoError(t, err)
	require.Len(t, q.Conditions, 1)
	assert.Equal(t, []string{"todo", "in progress", "done"}, q.Conditions[0].Values)
}
