package tracker_test

// Тесты HTTP-обработчика трекера задач (волна 114): интеграционные и unit-сценарии testframework.
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

// TestHandler_UpdateWorkflowStatus_InvalidJSON проверяет отклонение при HTTP-обработчик обновление workflow status некорректный json.
func TestHandler_UpdateWorkflowStatus_InvalidJSON(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	base := "/api/projects/" + projectID.String()

	req := httptest.NewRequest(http.MethodPut, base+"/workflow/statuses/"+uuid.New().String(), bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_CreateWorkflowTransition_InvalidFromID проверяет отклонение при HTTP-обработчик создание workflow transition некорректный from id.
func TestHandler_CreateWorkflowTransition_InvalidFromID(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	base := "/api/projects/" + projectID.String()

	body := `{"from_status_id":"bad","to_status_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, base+"/workflow/transitions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_DeleteWorkflowStatus_InvalidID проверяет отклонение при HTTP-обработчик удаление workflow status некорректный id.
func TestHandler_DeleteWorkflowStatus_InvalidID(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	base := "/api/projects/" + projectID.String()

	req := httptest.NewRequest(http.MethodDelete, base+"/workflow/statuses/not-a-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_DeleteLabel_InvalidID проверяет отклонение при HTTP-обработчик удаление label некорректный id.
func TestHandler_DeleteLabel_InvalidID(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	base := "/api/projects/" + projectID.String()

	req := httptest.NewRequest(http.MethodDelete, base+"/labels/not-a-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
