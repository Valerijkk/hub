package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchIssues_EmptyProjectIDs(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	results, err := search.ExportSearchIssues(svc, context.Background(), "bug", nil)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearchTestCases_EmptyProjectIDs(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	results, err := search.ExportSearchTestCases(svc, context.Background(), "login", nil)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearchChannels_EmptyProjectIDs(t *testing.T) {
	svc := search.ExportNewSearchServiceSQL(newScopeTestDB(t))
	results, err := search.ExportSearchChannels(svc, context.Background(), "general", nil)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestDeleteSavedFilter_FilterMissing(t *testing.T) {
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
