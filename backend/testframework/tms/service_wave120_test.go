package tms_test

// Тесты сервиса TMS (волна 120): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEntRepository_Wave120_UpdateEnvironment проверяет слой репозитория: ent репозиторий wave 120 обновление environment.
func TestEntRepository_Wave120_UpdateEnvironment(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)

	env, err := repo.CreateEnvironment(ctx, projectID, tms.CreateEnvironmentInput{
		Name: "Staging", Label: "staging",
	})
	require.NoError(t, err)

	newName := "Staging v2"
	newLabel := "staging"
	cfg := map[string]interface{}{"url": "https://staging.example"}
	updated, err := repo.UpdateEnvironment(ctx, env.ID, tms.UpdateEnvironmentInput{
		Name:   &newName,
		Label:  &newLabel,
		Config: &cfg,
	})
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)
}

// TestEntRepository_Wave120_GetLatestTestRunForEnvironment_Empty проверяет слой репозитория: ent репозиторий wave 120 получение latest test run for environment пустой.
func TestEntRepository_Wave120_GetLatestTestRunForEnvironment_Empty(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)

	env, err := repo.CreateEnvironment(ctx, projectID, tms.CreateEnvironmentInput{
		Name: "QA Env", Label: "dev",
	})
	require.NoError(t, err)

	run, err := repo.GetLatestTestRunForEnvironment(ctx, projectID, env.ID)
	require.NoError(t, err)
	assert.Nil(t, run)
}

// TestListTestRuns_Wave120_DefaultLimit проверяет сценарий: список test runs wave 120 default limit.
func TestListTestRuns_Wave120_DefaultLimit(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	projectID := uuid.New()

	page, err := svc.ListTestRuns(context.Background(), projectID, tms.ListTestRunsRequest{})
	require.NoError(t, err)
	assert.Equal(t, 50, page.Limit)
}

// TestGetCoverageReport_Wave120_Empty проверяет сценарий: получение покрытие report wave 120 пустой.
func TestGetCoverageReport_Wave120_Empty(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	report, err := svc.GetCoverageReport(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.NotNil(t, report)
}
