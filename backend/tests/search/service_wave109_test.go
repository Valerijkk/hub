package search_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/search"
)

// Empty-scope duplicates removed (W5-A5-005) — канон: TestSearchScopedEntities_EmptyProjectIDsReturnNothing.

func TestDeleteSavedFilter_FilterMissing(t *testing.T) {
	// W5-A5-006 / P3-BE-02 / TASK-092: ErrFilterNotFound sentinel.
	svc, _, ctx, userID, _ := newSearchServiceTestEnv(t)
	err := svc.DeleteSavedFilter(ctx, userID, uuid.New())
	require.ErrorIs(t, err, search.ErrFilterNotFound)
}

func TestShareFilter_NotFound(t *testing.T) {
	svc, _, ctx, userID, _ := newSearchServiceTestEnv(t)
	_, err := svc.ShareFilter(ctx, userID, uuid.New(), true)
	require.ErrorIs(t, err, search.ErrFilterNotFound)
}
