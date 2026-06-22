package tms_test

// Тесты сервиса TMS: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStartTestRun_Success проверяет сценарий: start test run успех.
func TestStartTestRun_Success(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	run, err := svc.CreateTestRun(ctx, projectID, userID, tms.CreateTestRunRequest{Title: "Start me"})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)

	started, err := svc.StartTestRun(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, "in_progress", started.Status)
}

// TestGetTestRunDetail_WithResults проверяет сценарий: получение test run detail with results.
func TestGetTestRunDetail_WithResults(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	tc, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Detail case", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)

	run, err := svc.CreateTestRun(ctx, projectID, userID, tms.CreateTestRunRequest{Title: "Detail run"})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)

	detail, err := svc.AddTestCasesToRun(ctx, runID, tms.AddTestCasesToRunRequest{
		CaseIDs: []string{tc.ID},
	})
	require.NoError(t, err)
	require.Len(t, detail.Results, 1)

	got, err := svc.GetTestRunDetail(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, "Detail run", got.Title)
	assert.Len(t, got.Results, 1)
}

// TestGetTestRunDetail_NotFound проверяет сценарий: получение test run detail not found.
func TestGetTestRunDetail_NotFound(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	_, err := svc.GetTestRunDetail(context.Background(), uuid.New())
	require.Error(t, err)
}

// TestAssignTestersToRun_Success проверяет сценарий: assign testers to run успех.
func TestAssignTestersToRun_Success(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	run, err := svc.CreateTestRun(ctx, projectID, userID, tms.CreateTestRunRequest{Title: "Assigned run"})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)

	testerID := uuid.New()
	detail, err := svc.AssignTestersToRun(ctx, runID, tms.AssignTestersRequest{
		UserIDs: []string{testerID.String()},
	})
	require.NoError(t, err)
	require.NotNil(t, detail)
	assert.Equal(t, run.ID, detail.ID)
}

// TestAssignTestersToRun_InvalidUserID проверяет отклонение при assign testers to run некорректный user id.
func TestAssignTestersToRun_InvalidUserID(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	run, err := svc.CreateTestRun(ctx, projectID, userID, tms.CreateTestRunRequest{Title: "Bad assign"})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)

	_, err = svc.AssignTestersToRun(ctx, runID, tms.AssignTestersRequest{UserIDs: []string{"not-a-uuid"}})
	require.Error(t, err)
}

// TestCreateEnvironment проверяет сценарий: создание environment.
func TestCreateEnvironment(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	env, err := svc.CreateEnvironment(ctx, projectID, tms.CreateEnvironmentRequest{
		Name:  "Staging EU",
		Label: "staging",
		Config: map[string]interface{}{
			"browser": "chrome",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "Staging EU", env.Name)
	assert.Equal(t, "staging", env.Label)
	assert.NotEmpty(t, env.ID)
}

// TestListEnvironments проверяет сценарий: список environments.
func TestListEnvironments(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	_, err := svc.CreateEnvironment(ctx, projectID, tms.CreateEnvironmentRequest{
		Name: "Dev", Label: "dev",
	})
	require.NoError(t, err)
	_, err = svc.CreateEnvironment(ctx, projectID, tms.CreateEnvironmentRequest{
		Name: "Prod", Label: "prod",
	})
	require.NoError(t, err)

	envs, err := svc.ListEnvironments(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, envs, 2)
}

// TestListEnvironments_EmptyProject проверяет сценарий: список environments пустой project.
func TestListEnvironments_EmptyProject(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	envs, err := svc.ListEnvironments(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Empty(t, envs)
}
