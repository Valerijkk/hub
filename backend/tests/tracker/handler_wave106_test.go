package tracker_test

// Тесты HTTP-обработчика трекера задач (волна 106): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

// TestHandler_Wave106_ListCommentsAfterCreate проверяет HTTP-слой: HTTP-обработчик wave 106 список comments after создание.
func TestHandler_Wave106_ListCommentsAfterCreate(t *testing.T) {
	app, projectID, issueKey, _ := setupTrackerCommentsApp(t)
	base := fmt.Sprintf("/api/projects/%s/issues/%s/comments", projectID, issueKey)

	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(`{"body_text":"listed comment"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, base, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var comments []tracker.CommentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&comments))
	require.Len(t, comments, 1)
	assert.Equal(t, "listed comment", comments[0].BodyText)
}

// TestHandler_Wave106_IssueLinkCRUD проверяет HTTP-слой: HTTP-обработчик wave 106 issue link crud.
func TestHandler_Wave106_IssueLinkCRUD(t *testing.T) {
	app, projectID, _, repo := setupTrackerExtendedApp(t)
	base := "/api/projects/" + projectID.String()

	issueA := createIssueViaHandler(t, app, projectID, "LinkA")
	issueB := createIssueViaHandler(t, app, projectID, "LinkB")

	linkBody := fmt.Sprintf(`{"target_issue_id":%q,"link_type":"relates_to"}`, issueB.ID)
	req := httptest.NewRequest(http.MethodPost, base+"/issues/"+issueA.IssueKey+"/links", bytes.NewReader([]byte(linkBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var link tracker.IssueLinkResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&link))
	require.NotEmpty(t, link.ID)

	req = httptest.NewRequest(http.MethodGet, base+"/issues/"+issueA.IssueKey+"/links", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, base+"/issues/"+issueA.IssueKey+"/links/"+link.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = repo
}

// TestHandler_Wave106_TimeLogsFlow проверяет HTTP-слой: HTTP-обработчик wave 106 time logs поток.
func TestHandler_Wave106_TimeLogsFlow(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	issue := createIssueViaHandler(t, app, projectID, "TimeLog")

	base := fmt.Sprintf("/api/projects/%s/issues/%s", projectID, issue.IssueKey)
	body := `{"minutes":45,"description":"investigation"}`
	req := httptest.NewRequest(http.MethodPost, base+"/time", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/time-logs", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave106_GetIssue_NotFound проверяет HTTP-слой: HTTP-обработчик wave 106 получение issue not found.
func TestHandler_Wave106_GetIssue_NotFound(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/issues/MISSING-99", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave106_CreateIssueLink_InvalidBody проверяет отклонение при HTTP-обработчик wave 106 создание issue link некорректный body.
func TestHandler_Wave106_CreateIssueLink_InvalidBody(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	issue := createIssueViaHandler(t, app, projectID, "BadLink")

	url := fmt.Sprintf("/api/projects/%s/issues/%s/links", projectID, issue.IssueKey)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave106_WorkflowStatusUpdate проверяет HTTP-слой: HTTP-обработчик wave 106 workflow status обновление.
func TestHandler_Wave106_WorkflowStatusUpdate(t *testing.T) {
	app, projectID, _, repo := setupTrackerExtendedApp(t)
	statusID := uuid.New()
	repo.statuses[statusID] = &ent.WorkflowStatus{
		ID: statusID, Name: "Review", Category: "in_progress",
	}

	body := `{"name":"In Review","category":"in_progress"}`
	url := fmt.Sprintf("/api/projects/%s/workflow/statuses/%s", projectID, statusID)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
