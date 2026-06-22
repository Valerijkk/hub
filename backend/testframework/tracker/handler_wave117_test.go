package tracker_test

// Тесты HTTP-обработчика трекера задач (волна 117): интеграционные и unit-сценарии testframework.
import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_CreateIssue_InvalidJSON проверяет отклонение при HTTP-обработчик создание issue некорректный json.
func TestHandler_CreateIssue_InvalidJSON(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/issues", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_StopTimer_InvalidProjectID проверяет отклонение при HTTP-обработчик stop timer некорректный project id.
func TestHandler_StopTimer_InvalidProjectID(t *testing.T) {
	app, _, _, _ := setupTrackerHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/projects/not-uuid/issues/HUB-1/timer/stop", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_CreateIssueLink_InvalidJSON проверяет отклонение при HTTP-обработчик создание issue link некорректный json.
func TestHandler_CreateIssueLink_InvalidJSON(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	base := "/api/projects/" + projectID.String()
	created := createIssueViaHandler(t, app, projectID, "Link source")

	req := httptest.NewRequest(http.MethodPost, base+"/issues/"+created.IssueKey+"/links", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
