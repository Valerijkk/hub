package tracker_test

// Тесты HTTP-обработчика трекера задач (волна 109): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
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

// TestHandler_Wave109_GetBoardData_SwimlaneQuery проверяет HTTP-слой: HTTP-обработчик wave 109 получение board data swimlane query.
func TestHandler_Wave109_GetBoardData_SwimlaneQuery(t *testing.T) {
	app, projectID, userID, _ := setupTrackerHandlerApp(t)

	body := `{"name":"Swim tracker.Handler Board","filters":{}}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/boards", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var board tracker.BoardResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&board))
	resp.Body.Close()

	issueBody := fmt.Sprintf(`{"title":"Board issue","issue_type":"task","priority":"medium","assignee_id":%q}`, userID.String())
	req = httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/issues", bytes.NewReader([]byte(issueBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	url := fmt.Sprintf("/api/projects/%s/boards/%s?swimlane=assignee", projectID, board.ID)
	req = httptest.NewRequest(http.MethodGet, url, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave109_LogTime_IssueNotFound проверяет HTTP-слой: HTTP-обработчик wave 109 log time issue not found.
func TestHandler_Wave109_LogTime_IssueNotFound(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)

	url := fmt.Sprintf("/api/projects/%s/issues/MISSING-99/time", projectID)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(`{"minutes":30}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusCreated, resp.StatusCode)
}

// TestHandler_Wave109_LogTime_InvalidJSON проверяет отклонение при HTTP-обработчик wave 109 log time некорректный json.
func TestHandler_Wave109_LogTime_InvalidJSON(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	issue := createIssueViaHandler(t, app, projectID, "BadJSON")

	url := fmt.Sprintf("/api/projects/%s/issues/%s/time", projectID, issue.IssueKey)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave109_UpdateComment_NotFound проверяет HTTP-слой: HTTP-обработчик wave 109 обновление comment not found.
func TestHandler_Wave109_UpdateComment_NotFound(t *testing.T) {
	app, projectID, issueKey, _ := setupTrackerCommentsApp(t)
	url := fmt.Sprintf("/api/projects/%s/issues/%s/comments/%s", projectID, issueKey, uuid.New().String())

	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader([]byte(`{"body_text":"ghost"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave109_GetBoardData_NotFound проверяет HTTP-слой: HTTP-обработчик wave 109 получение board data not found.
func TestHandler_Wave109_GetBoardData_NotFound(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/boards/"+uuid.New().String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}
