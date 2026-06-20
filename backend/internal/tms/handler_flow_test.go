package tms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTMSHandlerApp(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID) {
	t.Helper()
	svc := NewTMSService(newMockRepo(), nil)
	h := NewHandler(svc)
	projectID := uuid.New()
	userID := uuid.New()

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}

	tms := api.Group("/projects/:projectId/tms", authMw)
	tms.Post("/cases", h.CreateTestCase)
	tms.Get("/cases", h.ListTestCases)
	tms.Get("/cases/:caseKey", h.GetTestCase)
	tms.Post("/runs", h.CreateTestRun)
	tms.Get("/runs/:runId", h.GetTestRun)
	tms.Post("/runs/:runId/start", h.StartTestRun)
	tms.Put("/runs/:runId/status", h.UpdateRunStatus)
	tms.Post("/runs/:runId/reopen", h.ReopenTestRun)
	tms.Get("/runs/:runId/detail", h.GetTestRunDetail)
	tms.Post("/runs/:runId/assignees", h.AssignTestersToRun)
	tms.Post("/runs/:runId/cases", h.AddTestCasesToRun)
	tms.Post("/runs/:runId/complete", h.CompleteTestRun)
	tms.Get("/runs/:runId/results", h.ListTestResults)
	tms.Post("/runs/:runId/results", h.CreateTestResult)
	tms.Put("/runs/:runId/results/:resultId", h.ExecuteTestResult)

	tms.Post("/suites", h.CreateSuite)
	tms.Get("/suites", h.ListSuites)
	tms.Put("/suites/:suiteId", h.UpdateSuite)
	tms.Delete("/suites/:suiteId", h.DeleteSuite)
	tms.Post("/components", h.CreateComponent)
	tms.Get("/components", h.ListComponents)
	tms.Delete("/components/:componentId", h.DeleteComponent)
	tms.Put("/cases/:caseKey", h.UpdateTestCase)
	tms.Delete("/cases/:caseKey", h.DeleteTestCase)
	tms.Get("/cases/:caseKey/versions", h.ListTestCaseVersions)
	tms.Post("/cases/:caseKey/links", h.LinkTestCaseToIssue)
	tms.Delete("/cases/:caseKey/links/:issueId", h.UnlinkTestCaseFromIssue)
	tms.Get("/cases/:caseKey/links", h.GetTestCaseLinks)
	tms.Get("/runs", h.ListTestRuns)

	tms.Get("/reports/history", h.GetTestRunHistory)
	tms.Get("/reports/flaky", h.GetFlakyTests)
	tms.Get("/reports/coverage", h.GetCoverageReport)
	tms.Get("/reports/comparison", h.GetRunComparison)
	tms.Get("/reports/traceability", h.GetTraceabilityMatrix)
	tms.Get("/reports/env-comparison", h.CompareEnvironments)
	tms.Get("/reports/defect-categories", h.GetDefectCategories)

	envs := api.Group("/projects/:projectId/environments", authMw)
	envs.Post("/", h.CreateEnvironmentHandler)
	envs.Get("/", h.ListEnvironmentsHandler)

	return app, projectID, userID
}

func TestHandler_CreateAndGetTestCase(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/cases"

	body := `{"title":"Handler TC","case_type":"classic","priority":"high"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created TestCaseResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEmpty(t, created.CaseKey)

	req = httptest.NewRequest(http.MethodGet, base+"/"+created.CaseKey, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_ListTestCases(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/cases"

	body := `{"title":"Listed TC","case_type":"classic","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, base+"?offset=0&limit=20", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_CreateAndStartTestRun(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/runs"

	body := `{"title":"Handler run","run_type":"manual"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created TestRunResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEmpty(t, created.ID)

	req = httptest.NewRequest(http.MethodPost, base+"/"+created.ID+"/start", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var started TestRunResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&started))
	assert.Equal(t, "in_progress", started.Status)
}

func TestHandler_GetTestRunDetail(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms"

	tcBody := `{"title":"Run case","case_type":"classic","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base+"/cases", bytes.NewReader([]byte(tcBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var tc TestCaseResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&tc))
	resp.Body.Close()

	runBody := `{"title":"Detail handler run"}`
	req = httptest.NewRequest(http.MethodPost, base+"/runs", bytes.NewReader([]byte(runBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	var run TestRunResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&run))
	resp.Body.Close()

	addBody := fmt.Sprintf(`{"case_ids":["%s"]}`, tc.ID)
	req = httptest.NewRequest(http.MethodPost, base+"/runs/"+run.ID+"/cases", bytes.NewReader([]byte(addBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, base+"/runs/"+run.ID+"/detail", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var detail TestRunDetailResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&detail))
	assert.Len(t, detail.Results, 1)
}

func TestHandler_AssignTestersToRun(t *testing.T) {
	app, projectID, userID := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/runs"

	body := `{"title":"Assign run"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var run TestRunResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&run))
	resp.Body.Close()

	assignBody := fmt.Sprintf(`{"user_ids":["%s"]}`, userID.String())
	req = httptest.NewRequest(http.MethodPost, base+"/"+run.ID+"/assignees", bytes.NewReader([]byte(assignBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_CreateAndListEnvironments(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/environments"

	body := `{"name":"Handler Staging","label":"staging"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var envs []EnvironmentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&envs))
	assert.NotEmpty(t, envs)
}

func TestHandler_Unauthorized(t *testing.T) {
	h := NewHandler(NewTMSService(newMockRepo(), nil))
	app := fiber.New()
	api := app.Group("/api")
	api.Post("/projects/:projectId/tms/cases", h.CreateTestCase)

	projectID := uuid.New()
	body := `{"title":"No auth","case_type":"classic","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/tms/cases", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_SuiteAndComponentFlow(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms"

	suiteBody := `{"name":"Handler Suite","description":"folder"}`
	req := httptest.NewRequest(http.MethodPost, base+"/suites", bytes.NewReader([]byte(suiteBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var suite SuiteResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&suite))

	req = httptest.NewRequest(http.MethodGet, base+"/suites", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	updateBody := `{"name":"Renamed Suite"}`
	req = httptest.NewRequest(http.MethodPut, base+"/suites/"+suite.ID, bytes.NewReader([]byte(updateBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	compBody := `{"name":"API","description":"backend"}`
	req = httptest.NewRequest(http.MethodPost, base+"/components", bytes.NewReader([]byte(compBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var comp ComponentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&comp))

	req = httptest.NewRequest(http.MethodGet, base+"/components", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, base+"/components/"+comp.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHandler_ExecuteAndCompleteRun(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms"

	tcBody := `{"title":"Exec case","case_type":"classic","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base+"/cases", bytes.NewReader([]byte(tcBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var tc TestCaseResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&tc))
	resp.Body.Close()

	runBody := `{"title":"Exec run"}`
	req = httptest.NewRequest(http.MethodPost, base+"/runs", bytes.NewReader([]byte(runBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	var run TestRunResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&run))
	resp.Body.Close()

	addBody := fmt.Sprintf(`{"case_ids":["%s"]}`, tc.ID)
	req = httptest.NewRequest(http.MethodPost, base+"/runs/"+run.ID+"/cases", bytes.NewReader([]byte(addBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, base+"/runs/"+run.ID+"/start", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, base+"/runs/"+run.ID+"/detail", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	var detail TestRunDetailResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&detail))
	resp.Body.Close()
	require.Len(t, detail.Results, 1)

	execBody := `{"status":"passed","comment":"ok via handler"}`
	req = httptest.NewRequest(http.MethodPut, base+"/runs/"+run.ID+"/results/"+detail.Results[0].ID, bytes.NewReader([]byte(execBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/runs/"+run.ID+"/results", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, base+"/runs/"+run.ID+"/complete", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_UpdateAndDeleteTestCase(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/cases"

	body := `{"title":"To update","case_type":"classic","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var created TestCaseResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	updateBody := fmt.Sprintf(`{"title":"Updated via handler","version":%d}`, created.Version)
	req = httptest.NewRequest(http.MethodPut, base+"/"+created.CaseKey, bytes.NewReader([]byte(updateBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, base+"/"+created.CaseKey, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHandler_ListTestRunsAndReports(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms"

	runBody := `{"title":"Report run"}`
	req := httptest.NewRequest(http.MethodPost, base+"/runs", bytes.NewReader([]byte(runBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, base+"/runs?offset=0&limit=10", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/reports/history?limit=5", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/reports/flaky", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/reports/coverage", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_GetTestRunAndStatusFlow(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/runs"

	body := `{"title":"Status flow run"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var run TestRunResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&run))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, base+"/"+run.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	statusBody := `{"status":"in_progress"}`
	req = httptest.NewRequest(http.MethodPut, base+"/"+run.ID+"/status", bytes.NewReader([]byte(statusBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, base+"/"+run.ID+"/complete", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, base+"/"+run.ID+"/reopen", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_CreateTestResultAndDeleteSuite(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms"

	runBody := `{"title":"Result handler run"}`
	req := httptest.NewRequest(http.MethodPost, base+"/runs", bytes.NewReader([]byte(runBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var run TestRunResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&run))
	resp.Body.Close()

	resultBody := `{"status":"passed","comment":"manual result"}`
	req = httptest.NewRequest(http.MethodPost, base+"/runs/"+run.ID+"/results", bytes.NewReader([]byte(resultBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	suiteBody := `{"name":"Delete me"}`
	req = httptest.NewRequest(http.MethodPost, base+"/suites", bytes.NewReader([]byte(suiteBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	var suite SuiteResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&suite))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodDelete, base+"/suites/"+suite.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHandler_ListTestCaseVersions(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/cases"

	body := `{"title":"Versioned","case_type":"classic","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var created TestCaseResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, base+"/"+created.CaseKey+"/versions", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_ExtendedReports(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms"
	envBase := "/api/projects/" + projectID.String() + "/environments"

	runBody := `{"title":"Compare A"}`
	req := httptest.NewRequest(http.MethodPost, base+"/runs", bytes.NewReader([]byte(runBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var run1 TestRunResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&run1))
	resp.Body.Close()

	runBody = `{"title":"Compare B"}`
	req = httptest.NewRequest(http.MethodPost, base+"/runs", bytes.NewReader([]byte(runBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	var run2 TestRunResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&run2))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, base+"/reports/comparison?run1="+run1.ID+"&run2="+run2.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/reports/traceability", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/reports/defect-categories", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	envBody := `{"name":"Env A","label":"a"}`
	req = httptest.NewRequest(http.MethodPost, envBase, bytes.NewReader([]byte(envBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	var env1 EnvironmentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env1))
	resp.Body.Close()

	envBody = `{"name":"Env B","label":"b"}`
	req = httptest.NewRequest(http.MethodPost, envBase, bytes.NewReader([]byte(envBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	var env2 EnvironmentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env2))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, base+"/reports/env-comparison?env1="+env1.ID+"&env2="+env2.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_LinkTestCaseToIssue(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/cases"

	body := `{"title":"Linked case","case_type":"classic","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var created TestCaseResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	issueID := uuid.New()
	linkBody := `{"issue_id":"` + issueID.String() + `"}`
	req = httptest.NewRequest(http.MethodPost, base+"/"+created.CaseKey+"/links", bytes.NewReader([]byte(linkBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/"+created.CaseKey+"/links", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, base+"/"+created.CaseKey+"/links/"+issueID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
