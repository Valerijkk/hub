package tms_test

// Тесты репозитория TMS: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/testframework/mongotest"
)

// TestMongoRepository_AllureResultsLifecycle проверяет слой репозитория: mongo репозиторий allure results жизненный цикл.
func TestMongoRepository_AllureResultsLifecycle(t *testing.T) {
	db := mongotest.Connect(t, "hub_test_tms_")
	repo := tms.NewMongoRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.EnsureIndexes(ctx))
	require.NoError(t, repo.StoreAllureResults(ctx, uuid.New(), nil))

	runID := uuid.New()
	results := []tms.AllureResult{
		{UUID: "allure-uuid-1", Name: "testLogin", Status: "passed"},
		{UUID: "allure-uuid-2", Name: "testLogout", Status: "failed"},
	}
	require.NoError(t, repo.StoreAllureResults(ctx, runID, results))

	byRun, err := repo.GetAllureResultsByRunID(ctx, runID)
	require.NoError(t, err)
	require.Len(t, byRun, 2)
	assert.Equal(t, "testLogin", byRun[0].Name)

	one, err := repo.GetAllureResultByUUID(ctx, "allure-uuid-2")
	require.NoError(t, err)
	assert.Equal(t, "failed", one.Status)

	emptyRun := uuid.New()
	none, err := repo.GetAllureResultsByRunID(ctx, emptyRun)
	require.NoError(t, err)
	assert.Empty(t, none)
}

// TestNewMongoRepository проверяет слой репозитория: new mongo репозиторий.
func TestNewMongoRepository(t *testing.T) {
	db := mongotest.Connect(t, "hub_test_tms_")
	repo := tms.NewMongoRepository(db)
	require.NotNil(t, repo)
	require.NotNil(t, tms.ExportMongoRepositoryAllureResults(repo))
}
