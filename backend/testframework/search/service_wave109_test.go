package search_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Empty-scope duplicates removed (W5-A5-005) — канон: TestSearchScopedEntities_EmptyProjectIDsReturnNothing.

func TestDeleteSavedFilter_FilterMissing(t *testing.T) {
	// W5-A5-006 / P3-BE-02: единственный NotFound assert с Contains("не найден").
	svc, _, ctx, userID, _ := newSearchServiceTestEnv(t)
	err := svc.DeleteSavedFilter(ctx, userID, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

func TestShareFilter_NotFound(t *testing.T) {
	svc, _, ctx, userID, _ := newSearchServiceTestEnv(t)
	_, err := svc.ShareFilter(ctx, userID, uuid.New(), true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}
