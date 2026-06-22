package tms_test

// Тесты HTTP-обработчика TMS (волна 114): интеграционные и unit-сценарии testframework.
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

// TestHandler_ExecuteTestResult_InvalidJSON проверяет отклонение при HTTP-обработчик execute test result некорректный json.
func TestHandler_ExecuteTestResult_InvalidJSON(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/runs/" + uuid.New().String() + "/results/" + uuid.New().String()

	req := httptest.NewRequest(http.MethodPut, base, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_AssignTesters_InvalidJSON проверяет отклонение при HTTP-обработчик assign testers некорректный json.
func TestHandler_AssignTesters_InvalidJSON(t *testing.T) {
	app, projectID, _ := setupTMSHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/runs/" + uuid.New().String() + "/assignees"

	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte("[")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
