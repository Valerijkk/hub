package tms_test

// Тесты HTTP-обработчика TMS (волна 113): интеграционные и unit-сценарии testframework.
import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_GetRunComparison_InvalidRun1 проверяет отклонение при HTTP-обработчик получение run comparison некорректный run 1.
func TestHandler_GetRunComparison_InvalidRun1(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/reports/comparison"

	req := httptest.NewRequest(http.MethodGet, base+"?run1=bad&run2="+uuid.New().String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_GetRunComparison_InvalidRun2 проверяет отклонение при HTTP-обработчик получение run comparison некорректный run 2.
func TestHandler_GetRunComparison_InvalidRun2(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/reports/comparison"

	req := httptest.NewRequest(http.MethodGet, base+"?run1="+uuid.New().String()+"&run2=bad", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_UpdateSuite_InvalidJSON проверяет отклонение при HTTP-обработчик обновление suite некорректный json.
func TestHandler_UpdateSuite_InvalidJSON(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/suites"

	req := httptest.NewRequest(http.MethodPut, base+"/"+uuid.New().String(), bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_GetCoverageReport_Success проверяет HTTP-слой: HTTP-обработчик получение покрытие report успех.
func TestHandler_GetCoverageReport_Success(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/reports/coverage"

	req := httptest.NewRequest(http.MethodGet, base, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_GetTestTrends_Success проверяет HTTP-слой: HTTP-обработчик получение test trends успех.
func TestHandler_GetTestTrends_Success(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/reports/trends"

	req := httptest.NewRequest(http.MethodGet, base+"?limit=10", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
