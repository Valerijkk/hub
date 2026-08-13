package tms_test

// Тесты HTTP-обработчика TMS (волна 120): интеграционные и unit-сценарии testframework.
import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_Wave120_CompareEnvironments_Validation проверяет HTTP-слой: HTTP-обработчик wave 120 compare environments валидация.
func TestHandler_Wave120_CompareEnvironments_Validation(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/reports/env-comparison"

	req := httptest.NewRequest(http.MethodGet, base, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"?env1=bad&env2="+uuid.New().String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave120_GetDefectCategories проверяет HTTP-слой: HTTP-обработчик wave 120 получение defect categories.
func TestHandler_Wave120_GetDefectCategories(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/reports/defect-categories"

	req := httptest.NewRequest(http.MethodGet, base, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave120_GetTraceabilityMatrix проверяет HTTP-слой: HTTP-обработчик wave 120 получение traceability matrix.
func TestHandler_Wave120_GetTraceabilityMatrix(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/reports/traceability"

	req := httptest.NewRequest(http.MethodGet, base, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave120_ListSuites_Empty проверяет HTTP-слой: HTTP-обработчик wave 120 список suites пустой.
func TestHandler_Wave120_ListSuites_Empty(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/suites"

	req := httptest.NewRequest(http.MethodGet, base, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave120_ListTestRuns_Empty проверяет HTTP-слой: HTTP-обработчик wave 120 список test runs пустой.
func TestHandler_Wave120_ListTestRuns_Empty(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/runs"

	req := httptest.NewRequest(http.MethodGet, base, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave120_GetFlakyTests_Empty проверяет HTTP-слой: HTTP-обработчик wave 120 получение flaky tests пустой.
func TestHandler_Wave120_GetFlakyTests_Empty(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/reports/flaky"

	req := httptest.NewRequest(http.MethodGet, base, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
