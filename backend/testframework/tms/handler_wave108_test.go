package tms_test

// Тесты HTTP-обработчика TMS (волна 108): интеграционные и unit-сценарии testframework.
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

// TestHandler_CreateSuite_InvalidJSON проверяет отклонение при HTTP-обработчик создание suite некорректный json.
func TestHandler_CreateSuite_InvalidJSON(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/suites"

	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_CreateComponent_InvalidJSON проверяет отклонение при HTTP-обработчик создание component некорректный json.
func TestHandler_CreateComponent_InvalidJSON(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/components"

	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_GetTestCase_NotFound проверяет HTTP-слой: HTTP-обработчик получение test case not found.
func TestHandler_GetTestCase_NotFound(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	url := "/api/projects/" + projectID.String() + "/tms/cases/HUB-TC-404"

	req := httptest.NewRequest(http.MethodGet, url, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_DeleteSuite_InvalidID проверяет отклонение при HTTP-обработчик удаление suite некорректный id.
func TestHandler_DeleteSuite_InvalidID(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	url := "/api/projects/" + projectID.String() + "/tms/suites/not-a-uuid"

	req := httptest.NewRequest(http.MethodDelete, url, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_DeleteComponent_InvalidID проверяет отклонение при HTTP-обработчик удаление component некорректный id.
func TestHandler_DeleteComponent_InvalidID(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	url := "/api/projects/" + projectID.String() + "/tms/components/" + uuid.New().String()

	req := httptest.NewRequest(http.MethodDelete, url, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}
