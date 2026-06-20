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

func setupTMSImportHandlerApp(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID) {
	t.Helper()
	userID := uuid.New()
	projectID := uuid.New()
	svc := NewTMSService(newMockRepo(), nil)
	h := NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	tms := api.Group("/projects/:projectId/tms", authMw)
	tms.Post("/allure/import", h.ImportAllureResults)
	tms.Post("/import", h.ImportTestCases)
	tms.Get("/export", h.ExportTestCases)
	tms.Post("/runs", h.CreateTestRun)
	tms.Post("/runs/:runId/allure", h.ImportAllureToExistingRun)
	tms.Get("/runs/:runId/export", h.ExportTestRunReport)
	return app, projectID, userID
}

func TestHandler_ImportAllure_InvalidBody(t *testing.T) {
	app, projectID, _ := setupTMSImportHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/allure/import"

	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_ExportTestCases_CSV(t *testing.T) {
	app, projectID, _ := setupTMSImportHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/tms/export?format=csv", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Disposition"), "attachment")
}

func TestHandler_ImportTestCases_MissingMultipart(t *testing.T) {
	app, projectID, _ := setupTMSImportHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/tms/import", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.GreaterOrEqual(t, resp.StatusCode, 400)
}

func TestHandler_ExportTestCases_JSONAndXML(t *testing.T) {
	app, projectID, _ := setupTMSImportHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/export"

	req := httptest.NewRequest(http.MethodGet, base+"?format=json", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"?format=xml", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_ImportAllure_Success(t *testing.T) {
	app, projectID, _ := setupTMSImportHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/allure/import"

	body := fmt.Sprintf(`{"results":%s,"title":"Handler import"}`, validAllureJSON)
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestHandler_ImportAllureToExistingRun_Success(t *testing.T) {
	app, projectID, _ := setupTMSImportHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms"

	runBody := `{"title":"Existing run"}`
	req := httptest.NewRequest(http.MethodPost, base+"/runs", bytes.NewReader([]byte(runBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var run TestRunResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&run))
	resp.Body.Close()

	importBody := fmt.Sprintf(`{"results":%s,"title":"Append"}`, validAllureJSON)
	req = httptest.NewRequest(http.MethodPost, base+"/runs/"+run.ID+"/allure", bytes.NewReader([]byte(importBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_ExportTestRunReport(t *testing.T) {
	app, projectID, _ := setupTMSImportHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms"

	runBody := `{"title":"Export report run"}`
	req := httptest.NewRequest(http.MethodPost, base+"/runs", bytes.NewReader([]byte(runBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var run TestRunResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&run))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, base+"/runs/"+run.ID+"/export?format=json", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Disposition"), "run-report.json")
}
