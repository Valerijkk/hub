package tms_test

// Тесты модуля TMS (волна 105): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/testresult"
)

// TestHandler_CreateEnvironment_InvalidProjectID проверяет отклонение при HTTP-обработчик создание environment некорректный project id.
func TestHandler_CreateEnvironment_InvalidProjectID(t *testing.T) {
	app, _, _ := setupTMSHandlerApp(t)

	body := `{"name":"Dev","label":"dev"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/not-a-uuid/environments", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_CreateEnvironment_NilProjectID проверяет HTTP-слой: HTTP-обработчик создание environment nil project id.
func TestHandler_CreateEnvironment_NilProjectID(t *testing.T) {
	app, _, _ := setupTMSHandlerApp(t)

	body := `{"name":"Dev","label":"dev"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/00000000-0000-0000-0000-000000000000/environments", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_CreateEnvironment_MissingName проверяет отклонение при HTTP-обработчик создание environment отсутствующий name.
func TestHandler_CreateEnvironment_MissingName(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/environments"

	body := `{"name":"","label":"dev"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_UpdateEnvironment_InvalidEnvID проверяет отклонение при HTTP-обработчик обновление environment некорректный env id.
func TestHandler_UpdateEnvironment_InvalidEnvID(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/environments"

	body := `{"name":"Renamed"}`
	req := httptest.NewRequest(http.MethodPut, base+"/not-uuid", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_DeleteEnvironment_NotFound проверяет HTTP-слой: HTTP-обработчик удаление environment not found.
func TestHandler_DeleteEnvironment_NotFound(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/environments"

	req := httptest.NewRequest(http.MethodDelete, base+"/"+uuid.New().String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

// TestService_UpdateEnvironment_NotFound проверяет бизнес-логику: сервис обновление environment not found.
func TestService_UpdateEnvironment_NotFound(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	name := "X"
	_, err := svc.UpdateEnvironment(context.Background(), uuid.New(), tms.UpdateEnvironmentRequest{Name: &name})
	require.Error(t, err)
	var fe *fiber.Error
	require.ErrorAs(t, err, &fe)
	assert.Equal(t, fiber.StatusNotFound, fe.Code)
}

// TestService_DeleteEnvironment_NotFound проверяет бизнес-логику: сервис удаление environment not found.
func TestService_DeleteEnvironment_NotFound(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	err := svc.DeleteEnvironment(context.Background(), uuid.New())
	require.Error(t, err)
	var fe *fiber.Error
	require.ErrorAs(t, err, &fe)
	assert.Equal(t, fiber.StatusNotFound, fe.Code)
}

type envCompareRepo struct {
	mockRepository
	envRuns map[uuid.UUID]*ent.TestRun
}

func (r *envCompareRepo) GetLatestTestRunForEnvironment(_ context.Context, _, envID uuid.UUID) (*ent.TestRun, error) {
	if run, ok := r.envRuns[envID]; ok {
		return run, nil
	}
	return nil, nil
}

// TestCompareEnvironments_WithResults проверяет сценарий: compare environments with results.
func TestCompareEnvironments_WithResults(t *testing.T) {
	repo := &envCompareRepo{
		mockRepository: *newMockRepo(),
		envRuns:        make(map[uuid.UUID]*ent.TestRun),
	}
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	env1ID := uuid.New()
	env2ID := uuid.New()
	repo.environments[env1ID] = &ent.Environment{ID: env1ID, Name: "Dev", EnvLabel: "dev"}
	repo.environments[env2ID] = &ent.Environment{ID: env2ID, Name: "Prod", EnvLabel: "prod"}

	tc1ID := uuid.New()
	tc2ID := uuid.New()
	repo.testCases[tc1ID] = &ent.TestCase{ID: tc1ID, CaseKey: "HUB-TC-1", Title: "Login"}
	repo.testCases[tc2ID] = &ent.TestCase{ID: tc2ID, CaseKey: "HUB-TC-2", Title: "Checkout"}

	run1ID := uuid.New()
	run2ID := uuid.New()
	repo.envRuns[env1ID] = &ent.TestRun{ID: run1ID, CreatedAt: time.Now()}
	repo.envRuns[env2ID] = &ent.TestRun{ID: run2ID, CreatedAt: time.Now()}

	r1 := uuid.New()
	r2 := uuid.New()
	repo.results[r1] = &ent.TestResult{
		ID: r1, Status: testresult.StatusPassed,
		Edges: ent.TestResultEdges{TestCase: repo.testCases[tc1ID]},
	}
	repo.results[r2] = &ent.TestResult{
		ID: r2, Status: testresult.StatusFailed,
		Edges: ent.TestResultEdges{TestCase: repo.testCases[tc1ID]},
	}
	repo.runResults[run1ID] = []uuid.UUID{r1}
	repo.runResults[run2ID] = []uuid.UUID{r2}

	cmp, err := svc.CompareEnvironments(ctx, projectID, env1ID, env2ID)
	require.NoError(t, err)
	require.Len(t, cmp.Items, 1)
	assert.Equal(t, "HUB-TC-1", cmp.Items[0].CaseKey)
	assert.Equal(t, "passed", cmp.Items[0].StatusEnv1)
	assert.Equal(t, "failed", cmp.Items[0].StatusEnv2)
}

// TestCompareEnvironments_EnvNotFound проверяет сценарий: compare environments env not found.
func TestCompareEnvironments_EnvNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	env2ID := uuid.New()
	repo.environments[env2ID] = &ent.Environment{ID: env2ID, Name: "Prod", EnvLabel: "prod"}

	_, err := svc.CompareEnvironments(ctx, projectID, uuid.New(), env2ID)
	require.Error(t, err)
}

// TestHandler_CompareEnvironments_Success проверяет HTTP-слой: HTTP-обработчик compare environments успех.
func TestHandler_CompareEnvironments_Success(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	envBase := "/api/projects/" + projectID.String() + "/environments"

	createBody := `{"name":"Dev","label":"dev"}`
	req := httptest.NewRequest(http.MethodPost, envBase, bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var env1 tms.EnvironmentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env1))
	resp.Body.Close()

	createBody = `{"name":"Prod","label":"prod"}`
	req = httptest.NewRequest(http.MethodPost, envBase, bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	var env2 tms.EnvironmentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env2))
	resp.Body.Close()

	reportURL := "/api/projects/" + projectID.String() + "/tms/reports/env-comparison?env1=" + env1.ID + "&env2=" + env2.ID
	req = httptest.NewRequest(http.MethodGet, reportURL, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
