package tms

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartTestRun_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	run, err := svc.CreateTestRun(ctx, projectID, userID, CreateTestRunRequest{Title: "Start me"})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)

	started, err := svc.StartTestRun(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, "in_progress", started.Status)
}

func TestGetTestRunDetail_WithResults(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	tc, err := svc.CreateTestCase(ctx, projectID, userID, CreateTestCaseRequest{
		Title: "Detail case", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)

	run, err := svc.CreateTestRun(ctx, projectID, userID, CreateTestRunRequest{Title: "Detail run"})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)

	detail, err := svc.AddTestCasesToRun(ctx, runID, AddTestCasesToRunRequest{
		CaseIDs: []string{tc.ID},
	})
	require.NoError(t, err)
	require.Len(t, detail.Results, 1)

	got, err := svc.GetTestRunDetail(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, "Detail run", got.Title)
	assert.Len(t, got.Results, 1)
}

func TestGetTestRunDetail_NotFound(t *testing.T) {
	svc := NewTMSService(newMockRepo(), nil)
	_, err := svc.GetTestRunDetail(context.Background(), uuid.New())
	require.Error(t, err)
}

func TestAssignTestersToRun_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	run, err := svc.CreateTestRun(ctx, projectID, userID, CreateTestRunRequest{Title: "Assigned run"})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)

	testerID := uuid.New()
	detail, err := svc.AssignTestersToRun(ctx, runID, AssignTestersRequest{
		UserIDs: []string{testerID.String()},
	})
	require.NoError(t, err)
	require.NotNil(t, detail)
	assert.Equal(t, run.ID, detail.ID)
}

func TestAssignTestersToRun_InvalidUserID(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	run, err := svc.CreateTestRun(ctx, projectID, userID, CreateTestRunRequest{Title: "Bad assign"})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)

	_, err = svc.AssignTestersToRun(ctx, runID, AssignTestersRequest{UserIDs: []string{"not-a-uuid"}})
	require.Error(t, err)
}

func TestCreateEnvironment(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	env, err := svc.CreateEnvironment(ctx, projectID, CreateEnvironmentRequest{
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

func TestListEnvironments(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	_, err := svc.CreateEnvironment(ctx, projectID, CreateEnvironmentRequest{
		Name: "Dev", Label: "dev",
	})
	require.NoError(t, err)
	_, err = svc.CreateEnvironment(ctx, projectID, CreateEnvironmentRequest{
		Name: "Prod", Label: "prod",
	})
	require.NoError(t, err)

	envs, err := svc.ListEnvironments(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, envs, 2)
}

func TestListEnvironments_EmptyProject(t *testing.T) {
	svc := NewTMSService(newMockRepo(), nil)
	envs, err := svc.ListEnvironments(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Empty(t, envs)
}
