package tms_test

// Тесты HTTP-обработчика TMS (волна 122): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Фаза 122 — runs/results handler and ent-репозиторий coverage..

func TestHandler_Wave122_CreateSuiteAndRun(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms"

	suiteBody := `{"name":"Regression","description":"core flows"}`
	req := httptest.NewRequest(http.MethodPost, base+"/suites", bytes.NewReader([]byte(suiteBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	runBody := `{"title":"Sprint run"}`
	req = httptest.NewRequest(http.MethodPost, base+"/runs", bytes.NewReader([]byte(runBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var run tms.TestRunResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&run))
	require.NotEmpty(t, run.ID)

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s/runs/%s", base, run.ID), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestEntRepository_Wave122_CreateTestRunAndResult проверяет слой репозитория: ent репозиторий wave 122 создание test run and result.
func TestEntRepository_Wave122_CreateTestRunAndResult(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Manual run", CreatedByID: userID,
	})
	require.NoError(t, err)
	require.NotNil(t, run)

	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Login check", CaseType: "classic", Priority: "high", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	result, err := repo.CreateTestResult(ctx, run.ID, tms.CreateTestResultInput{
		Status: "unknown", TestCaseID: &tc.ID, ExecutedByID: &userID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	results, err := repo.ListTestResults(ctx, run.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

// TestService_Wave122_ReopenTestRun проверяет бизнес-логику: сервис wave 122 reopen test run.
func TestService_Wave122_ReopenTestRun(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	projectID := uuid.New()
	userID := uuid.New()
	ctx := context.Background()

	run, err := svc.CreateTestRun(ctx, projectID, userID, tms.CreateTestRunRequest{Title: "Reopen me"})
	require.NoError(t, err)
	runID := uuid.MustParse(run.ID)

	_, err = svc.StartTestRun(ctx, runID)
	require.NoError(t, err)

	_, err = svc.CompleteTestRun(ctx, runID)
	require.NoError(t, err)

	reopened, err := svc.ReopenTestRun(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, "in_progress", reopened.Status)
}
