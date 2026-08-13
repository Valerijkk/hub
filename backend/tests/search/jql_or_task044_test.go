//go:build hubtest

package search_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/search"
)

// TestToEntPredicates_ORBuildsUnion — W5-A4-005 / P2-BE-18: OR не схлопывается в AND.
func TestToEntPredicates_ORBuildsUnion(t *testing.T) {
	q, err := search.ParseJQL(`priority = high OR type = bug`)
	require.NoError(t, err)
	require.Len(t, q.Conditions, 2)
	assert.Equal(t, "OR", q.Conditions[1].Logic)

	preds := search.ToEntPredicates(q, uuid.New())
	require.Len(t, preds, 1, "OR-группа должна схлопнуться в один issue.Or предикат")

	// AND → тоже один сгруппированный предикат (не два независимых AND-flat).
	qAnd, err := search.ParseJQL(`priority = high AND type = bug`)
	require.NoError(t, err)
	predsAnd := search.ToEntPredicates(qAnd, uuid.New())
	require.Len(t, predsAnd, 1)
}
