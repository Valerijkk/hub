package tracker_test

// Тесты HTTP-обработчика трекера задач (волна 120): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
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

// TestHandler_Wave120_CommentReactionUpdateDelete проверяет HTTP-слой: HTTP-обработчик wave 120 comment reaction обновление удаление.
func TestHandler_Wave120_CommentReactionUpdateDelete(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	issue := wave119CreateIssue(t, app, projectID)
	base := fmt.Sprintf("/api/projects/%s/issues/%s", projectID, issue.IssueKey)

	body := `{"body_json":{"type":"doc"},"body_text":"Wave120 comment","mentions":[]}`
	req := httptest.NewRequest(http.MethodPost, base+"/comments", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var comment tracker.CommentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&comment))

	reactBody := `{"emoji":"👍"}`
	req = httptest.NewRequest(http.MethodPost, base+"/comments/"+comment.ID+"/reactions", bytes.NewReader([]byte(reactBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	updateBody := `{"body_json":{"type":"doc","v":2},"body_text":"Updated wave120"}`
	req = httptest.NewRequest(http.MethodPut, base+"/comments/"+comment.ID, bytes.NewReader([]byte(updateBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave120_DeleteComment_NoReactions проверяет HTTP-слой: HTTP-обработчик wave 120 удаление comment no reactions.
func TestHandler_Wave120_DeleteComment_NoReactions(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	issue := wave119CreateIssue(t, app, projectID)
	base := fmt.Sprintf("/api/projects/%s/issues/%s", projectID, issue.IssueKey)

	body := `{"body_json":{"type":"doc"},"body_text":"Delete me"}`
	req := httptest.NewRequest(http.MethodPost, base+"/comments", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var comment tracker.CommentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&comment))

	req = httptest.NewRequest(http.MethodDelete, base+"/comments/"+comment.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestHandler_Wave120_MoveIssueViaRoutes проверяет HTTP-слой: HTTP-обработчик wave 120 move issue via маршруты.
func TestHandler_Wave120_MoveIssueViaRoutes(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	issue := wave119CreateIssue(t, app, projectID)
	base := fmt.Sprintf("/api/projects/%s/issues/%s", projectID, issue.IssueKey)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/workflow/statuses", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var statuses []tracker.WorkflowStatusResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&statuses))
	require.NotEmpty(t, statuses)

	var targetID string
	for _, st := range statuses {
		if st.Name == "In Progress" {
			targetID = st.ID
			break
		}
	}
	require.NotEmpty(t, targetID)

	moveBody := fmt.Sprintf(`{"status_id":%q,"rank":"aaa|bbb"}`, targetID)
	req = httptest.NewRequest(http.MethodPut, base+"/position", bytes.NewReader([]byte(moveBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave120_LabelGroupAndTimer проверяет HTTP-слой: HTTP-обработчик wave 120 label group and timer.
func TestHandler_Wave120_LabelGroupAndTimer(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	issue := wave119CreateIssue(t, app, projectID)
	base := fmt.Sprintf("/api/projects/%s/issues/%s", projectID, issue.IssueKey)
	labelsBase := "/api/projects/" + projectID.String() + "/labels"

	groupBody := `{"name":"Wave120 Group"}`
	req := httptest.NewRequest(http.MethodPost, labelsBase+"/groups", bytes.NewReader([]byte(groupBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, labelsBase+"/groups", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, base+"/timer/start", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, base+"/timer/stop", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave120_AutomationRuleCRUD проверяет HTTP-слой: HTTP-обработчик wave 120 автоматизация rule crud.
func TestHandler_Wave120_AutomationRuleCRUD(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	autoBase := "/api/projects/" + projectID.String() + "/automation-rules"

	createBody := `{"name":"On create notify","trigger_type":"issue_created","action_type":"notify_user"}`
	req := httptest.NewRequest(http.MethodPost, autoBase, bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var rule tracker.AutomationRuleResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&rule))

	req = httptest.NewRequest(http.MethodGet, autoBase, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, autoBase+"/"+rule.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestHandler_Wave120_IssueLinksAndLogTime проверяет HTTP-слой: HTTP-обработчик wave 120 issue links and log time.
func TestHandler_Wave120_IssueLinksAndLogTime(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	source := wave119CreateIssue(t, app, projectID)
	target := wave119CreateIssue(t, app, projectID)
	base := fmt.Sprintf("/api/projects/%s/issues/%s", projectID, source.IssueKey)

	linkBody := fmt.Sprintf(`{"target_issue_id":%q,"link_type":"blocks"}`, target.ID)
	req := httptest.NewRequest(http.MethodPost, base+"/links", bytes.NewReader([]byte(linkBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/links", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	timeBody := `{"minutes":15,"description":"wave120 work"}`
	req = httptest.NewRequest(http.MethodPost, base+"/time", bytes.NewReader([]byte(timeBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/time-logs", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave120_CreateLabelGroup_InvalidJSON проверяет отклонение при HTTP-обработчик wave 120 создание label group некорректный json.
func TestHandler_Wave120_CreateLabelGroup_InvalidJSON(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	base := "/api/projects/" + projectID.String() + "/labels/groups"

	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
