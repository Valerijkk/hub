package tracker_test

// Тесты HTTP-обработчика трекера задач (волна 112): интеграционные и unit-сценарии testframework.
import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_UpdateIssue_InvalidJSON проверяет отклонение при HTTP-обработчик обновление issue некорректный json.
func TestHandler_UpdateIssue_InvalidJSON(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)
	issue := createIssueViaHandler(t, app, projectID, "JSON issue")

	req := httptest.NewRequest(http.MethodPut, "/api/projects/"+projectID.String()+"/issues/"+issue.IssueKey, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_MoveIssue_InvalidJSON проверяет отклонение при HTTP-обработчик move issue некорректный json.
func TestHandler_MoveIssue_InvalidJSON(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)
	issue := createIssueViaHandler(t, app, projectID, "Move issue")

	req := httptest.NewRequest(http.MethodPut, "/api/projects/"+projectID.String()+"/issues/"+issue.IssueKey+"/position", bytes.NewReader([]byte("[")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_ListIssues_Success проверяет HTTP-слой: HTTP-обработчик список issues успех.
func TestHandler_ListIssues_Success(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)
	createIssueViaHandler(t, app, projectID, "Listed issue")

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/issues?limit=10", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.NotNil(t, body["items"])
}

// TestHandler_DeleteComment_InvalidCommentID проверяет отклонение при HTTP-обработчик удаление comment некорректный comment id.
func TestHandler_DeleteComment_InvalidCommentID(t *testing.T) {
	app, projectID, issueKey, _ := setupTrackerCommentsApp(t)

	url := fmt.Sprintf("/api/projects/%s/issues/%s/comments/not-a-uuid", projectID, issueKey)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_CreateWorkflowStatus_InvalidJSON проверяет отклонение при HTTP-обработчик создание workflow status некорректный json.
func TestHandler_CreateWorkflowStatus_InvalidJSON(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/workflow/statuses", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_CreateLabel_InvalidJSON проверяет отклонение при HTTP-обработчик создание label некорректный json.
func TestHandler_CreateLabel_InvalidJSON(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/labels", bytes.NewReader([]byte("[")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_DeleteIssueLink_InvalidLinkID проверяет отклонение при HTTP-обработчик удаление issue link некорректный link id.
func TestHandler_DeleteIssueLink_InvalidLinkID(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	issue := createIssueViaHandler(t, app, projectID, "Link host")

	url := fmt.Sprintf("/api/projects/%s/issues/%s/links/not-a-uuid", projectID, issue.IssueKey)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_LogTime_InvalidJSON проверяет отклонение при HTTP-обработчик log time некорректный json.
func TestHandler_LogTime_InvalidJSON(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	issue := createIssueViaHandler(t, app, projectID, "Time log")

	url := fmt.Sprintf("/api/projects/%s/issues/%s/time", projectID, issue.IssueKey)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
