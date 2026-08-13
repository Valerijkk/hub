package tracker_test

// Тесты HTTP-обработчика трекера задач (волна 118): интеграционные и unit-сценарии testframework.
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

// TestHandler_Wave118_CreateAutomationRule_InvalidJSON проверяет отклонение при HTTP-обработчик wave 118 создание автоматизация rule некорректный json.
func TestHandler_Wave118_CreateAutomationRule_InvalidJSON(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	base := "/api/projects/" + projectID.String()

	req := httptest.NewRequest(http.MethodPost, base+"/automation-rules", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave118_CreateBoard_InvalidJSON проверяет отклонение при HTTP-обработчик wave 118 создание board некорректный json.
func TestHandler_Wave118_CreateBoard_InvalidJSON(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/boards/", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave118_DeleteIssue_NotFound проверяет HTTP-слой: HTTP-обработчик wave 118 удаление issue not found.
func TestHandler_Wave118_DeleteIssue_NotFound(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/projects/"+projectID.String()+"/issues/NOPE-99", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusNoContent, resp.StatusCode)
}

// TestHandler_Wave118_ListBoards_InvalidProject проверяет отклонение при HTTP-обработчик wave 118 список boards некорректный project.
func TestHandler_Wave118_ListBoards_InvalidProject(t *testing.T) {
	app, _ := newTrackerRoutesEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/not-uuid/boards/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave118_CreateCustomField_InvalidJSON проверяет отклонение при HTTP-обработчик wave 118 создание custom field некорректный json.
func TestHandler_Wave118_CreateCustomField_InvalidJSON(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	base := "/api/projects/" + projectID.String() + "/custom-fields"

	req := httptest.NewRequest(http.MethodPost, base+"/", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave118_DeleteCustomField_InvalidID проверяет отклонение при HTTP-обработчик wave 118 удаление custom field некорректный id.
func TestHandler_Wave118_DeleteCustomField_InvalidID(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	base := "/api/projects/" + projectID.String() + "/custom-fields"

	req := httptest.NewRequest(http.MethodDelete, base+"/bad-id", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave118_ListAutomationRules_InvalidProject проверяет отклонение при HTTP-обработчик wave 118 список автоматизация rules некорректный project.
func TestHandler_Wave118_ListAutomationRules_InvalidProject(t *testing.T) {
	app, _ := newTrackerRoutesEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/not-uuid/automation-rules", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave118_UpdateAutomationRule_InvalidJSON проверяет отклонение при HTTP-обработчик wave 118 обновление автоматизация rule некорректный json.
func TestHandler_Wave118_UpdateAutomationRule_InvalidJSON(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	base := "/api/projects/" + projectID.String()

	req := httptest.NewRequest(http.MethodPatch, base+"/automation-rules/"+uuid.New().String(), bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
