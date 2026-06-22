package tms_test

// Тесты HTTP-обработчика TMS (волна 110): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
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

// TestHandler_ListSuites_Empty проверяет HTTP-слой: HTTP-обработчик список suites пустой.
func TestHandler_ListSuites_Empty(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/suites"

	req := httptest.NewRequest(http.MethodGet, base, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var suites []tms.SuiteResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&suites))
	assert.Empty(t, suites)
}

// TestHandler_CreateAndListSuite проверяет HTTP-слой: HTTP-обработчик создание and список suite.
func TestHandler_CreateAndListSuite(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/suites"

	body := `{"name":"Regression","description":"core flows"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var suites []tms.SuiteResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&suites))
	require.Len(t, suites, 1)
	assert.Equal(t, "Regression", suites[0].Name)
}

// TestHandler_CreateTestCase_InvalidJSON проверяет отклонение при HTTP-обработчик создание test case некорректный json.
func TestHandler_CreateTestCase_InvalidJSON(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/cases"

	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_UpdateTestCase_InvalidJSON проверяет отклонение при HTTP-обработчик обновление test case некорректный json.
func TestHandler_UpdateTestCase_InvalidJSON(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/cases"

	createBody := `{"title":"Update me","case_type":"classic","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var created tms.TestCaseResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPut, base+"/"+created.CaseKey, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_GetTestRun_NotFound проверяет HTTP-слой: HTTP-обработчик получение test run not found.
func TestHandler_GetTestRun_NotFound(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	url := fmt.Sprintf("/api/projects/%s/tms/runs/%s", projectID, uuid.New().String())

	req := httptest.NewRequest(http.MethodGet, url, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ListComponents_Empty проверяет HTTP-слой: HTTP-обработчик список components пустой.
func TestHandler_ListComponents_Empty(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/components"

	req := httptest.NewRequest(http.MethodGet, base, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var comps []tms.ComponentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&comps))
	assert.Empty(t, comps)
}

// TestHandler_CreateTestRun_InvalidJSON проверяет отклонение при HTTP-обработчик создание test run некорректный json.
func TestHandler_CreateTestRun_InvalidJSON(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/runs"

	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
