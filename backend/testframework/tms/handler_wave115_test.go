package tms_test

// Тесты HTTP-обработчика TMS (волна 115): интеграционные и unit-сценарии testframework.
import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_ReopenTestRun_InvalidRunID проверяет отклонение при HTTP-обработчик reopen test run некорректный run id.
func TestHandler_ReopenTestRun_InvalidRunID(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/runs"

	req := httptest.NewRequest(http.MethodPost, base+"/not-a-uuid/reopen", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_CompleteTestRun_InvalidRunID проверяет отклонение при HTTP-обработчик complete test run некорректный run id.
func TestHandler_CompleteTestRun_InvalidRunID(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/runs"

	req := httptest.NewRequest(http.MethodPost, base+"/bad-id/complete", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_GetTestRunDetail_InvalidRunID проверяет отклонение при HTTP-обработчик получение test run detail некорректный run id.
func TestHandler_GetTestRunDetail_InvalidRunID(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/runs"

	req := httptest.NewRequest(http.MethodGet, base+"/not-uuid/detail", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
