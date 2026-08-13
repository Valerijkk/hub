package tms_test

// Тесты HTTP-обработчика TMS (волна 88): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/testresult"
)

// TestHandler_RollbackTestCase проверяет HTTP-слой: HTTP-обработчик rollback test case.
func TestHandler_RollbackTestCase(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/cases"

	body := `{"title":"Rollback me","case_type":"classic","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var created tms.TestCaseResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	updateBody := `{"title":"Updated title","version":1}`
	req = httptest.NewRequest(http.MethodPut, base+"/"+created.CaseKey, bytes.NewReader([]byte(updateBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, base+"/"+created.CaseKey+"/versions", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	var versions []tms.TestCaseVersionResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&versions))
	resp.Body.Close()
	require.NotEmpty(t, versions)

	req = httptest.NewRequest(http.MethodPost, base+"/"+created.CaseKey+"/rollback/"+versions[0].ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var rolled tms.TestCaseResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&rolled))
	assert.Equal(t, "Rollback me", rolled.Title)
}

// TestHandler_CreateBugFromResult проверяет HTTP-слой: HTTP-обработчик создание bug from result.
func TestHandler_CreateBugFromResult(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	h := tms.NewHandler(svc)
	projectID := uuid.New()
	userID := uuid.New()

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	tms := api.Group("/projects/:projectId/tms", authMw)
	tms.Post("/runs/:runId/results/:resultId/bug", h.CreateBugFromResult)

	tcID := uuid.New()
	repo.testCases[tcID] = &ent.TestCase{
		ID: tcID, CaseKey: "HUB-TC-1", Title: "Failed case",
	}
	runID := uuid.New()
	repo.runs[runID] = &ent.TestRun{
		ID: runID,
		Edges: ent.TestRunEdges{
			Project: &ent.Project{ID: projectID},
		},
	}
	resultID := uuid.New()
	repo.results[resultID] = &ent.TestResult{
		ID: resultID, Status: testresult.StatusFailed, ErrorMessage: "boom",
		Edges: ent.TestResultEdges{
			TestCase: repo.testCases[tcID],
			Run:      repo.runs[runID],
		},
	}

	body := `{"title":"Custom bug title"}`
	path := "/api/projects/" + projectID.String() + "/tms/runs/" + runID.String() + "/results/" + resultID.String() + "/bug"
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// TestHandler_EnvironmentUpdateAndDelete проверяет HTTP-слой: HTTP-обработчик environment обновление and удаление.
func TestHandler_EnvironmentUpdateAndDelete(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	envBase := "/api/projects/" + projectID.String() + "/environments"

	createBody := `{"name":"Staging","label":"staging"}`
	req := httptest.NewRequest(http.MethodPost, envBase, bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var env tms.EnvironmentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	resp.Body.Close()

	updateBody := `{"name":"Staging v2"}`
	req = httptest.NewRequest(http.MethodPut, envBase+"/"+env.ID, bytes.NewReader([]byte(updateBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated tms.EnvironmentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&updated))
	assert.Equal(t, "Staging v2", updated.Name)

	req = httptest.NewRequest(http.MethodDelete, envBase+"/"+env.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestHandler_GetRunComparison_MissingParams проверяет отклонение при HTTP-обработчик получение run comparison отсутствующий params.
func TestHandler_GetRunComparison_MissingParams(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/reports/comparison"

	req := httptest.NewRequest(http.MethodGet, base, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
