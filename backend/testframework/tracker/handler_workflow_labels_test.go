package tracker_test

// Тесты HTTP-обработчика трекера задач: интеграционные и unit-сценарии testframework.
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
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func setupTrackerExtendedApp(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID, *mockRepo) {
	t.Helper()
	svc, repo := newTestService()
	h := tracker.NewHandler(svc, nil)
	projectID := uuid.New()
	userID := uuid.New()

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	proj := api.Group("/projects/:projectId", authMw)

	proj.Post("/issues", h.CreateIssue)
	proj.Get("/issues/:issueKey", h.GetIssue)

	proj.Post("/issues/:issueKey/comments", h.CreateComment)
	proj.Get("/issues/:issueKey/comments", h.ListComments)
	proj.Put("/issues/:issueKey/comments/:commentId", h.UpdateComment)
	proj.Delete("/issues/:issueKey/comments/:commentId", h.DeleteComment)

	proj.Get("/labels", h.ListLabels)
	proj.Post("/labels", h.CreateLabel)
	proj.Delete("/labels/:labelId", h.DeleteLabel)
	proj.Get("/labels/groups", h.ListLabelGroups)
	proj.Post("/labels/groups", h.CreateLabelGroup)

	proj.Get("/workflow/statuses", h.ListWorkflowStatuses)
	proj.Post("/workflow/statuses", h.CreateWorkflowStatus)
	proj.Put("/workflow/statuses/:statusId", h.UpdateWorkflowStatus)
	proj.Delete("/workflow/statuses/:statusId", h.DeleteWorkflowStatus)
	proj.Get("/workflow/transitions", h.ListWorkflowTransitions)
	proj.Post("/workflow/transitions", h.CreateWorkflowTransition)
	proj.Delete("/workflow/transitions/:transitionId", h.DeleteWorkflowTransition)

	proj.Post("/issues/:issueKey/links", h.CreateIssueLink)
	proj.Get("/issues/:issueKey/links", h.ListIssueLinks)
	proj.Delete("/issues/:issueKey/links/:linkId", h.DeleteIssueLink)

	proj.Post("/issues/:issueKey/time", h.LogTime)
	proj.Get("/issues/:issueKey/time-logs", h.ListTimeLogs)

	return app, projectID, userID, repo
}

// TestHandler_LabelCRUDFlow проверяет HTTP-слой: HTTP-обработчик label crud поток.
func TestHandler_LabelCRUDFlow(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	base := "/api/projects/" + projectID.String()

	body := `{"name":"phase59-label","color":"#336699"}`
	req := httptest.NewRequest(http.MethodPost, base+"/labels", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var label tracker.LabelResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&label))
	require.NotEmpty(t, label.ID)

	req = httptest.NewRequest(http.MethodGet, base+"/labels", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var labels []tracker.LabelResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&labels))
	assert.NotEmpty(t, labels)

	req = httptest.NewRequest(http.MethodDelete, base+"/labels/"+label.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestHandler_LabelGroupsFlow проверяет HTTP-слой: HTTP-обработчик label groups поток.
func TestHandler_LabelGroupsFlow(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	base := "/api/projects/" + projectID.String()

	body := `{"name":"Components"}`
	req := httptest.NewRequest(http.MethodPost, base+"/labels/groups", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/labels/groups", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var groups []tracker.LabelGroupResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&groups))
	assert.Len(t, groups, 1)
	assert.Equal(t, "Components", groups[0].Name)
}

// TestHandler_WorkflowStatusAndTransitionFlow проверяет HTTP-слой: HTTP-обработчик workflow status and transition поток.
func TestHandler_WorkflowStatusAndTransitionFlow(t *testing.T) {
	app, projectID, _, repo := setupTrackerExtendedApp(t)
	base := "/api/projects/" + projectID.String()

	body := `{"name":"Review","category":"in_progress","color":"#AABBCC"}`
	req := httptest.NewRequest(http.MethodPost, base+"/workflow/statuses", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var status tracker.WorkflowStatusResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&status))
	require.NotEmpty(t, status.ID)

	updateBody := `{"name":"Code Review","category":"in_progress","color":"#112233"}`
	req = httptest.NewRequest(http.MethodPut, base+"/workflow/statuses/"+status.ID, bytes.NewReader([]byte(updateBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	fromID := repo.getDefaultStatusID()
	toID := uuid.MustParse(status.ID)
	transBody := fmt.Sprintf(`{"from_status_id":%q,"to_status_id":%q}`, fromID.String(), toID.String())
	req = httptest.NewRequest(http.MethodPost, base+"/workflow/transitions", bytes.NewReader([]byte(transBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var transition tracker.WorkflowTransitionResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&transition))
	require.NotEmpty(t, transition.ID)

	req = httptest.NewRequest(http.MethodDelete, base+"/workflow/transitions/"+transition.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, base+"/workflow/statuses/"+status.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	_ = workflowstatus.CategoryInProgress
}

// TestHandler_DeleteIssueLink проверяет HTTP-слой: HTTP-обработчик удаление issue link.
func TestHandler_DeleteIssueLink(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	base := "/api/projects/" + projectID.String()

	source := createIssueViaHandler(t, app, projectID, "Link source")
	target := createIssueViaHandler(t, app, projectID, "Link target")

	linkBody := fmt.Sprintf(`{"target_issue_id":%q,"link_type":"relates_to"}`, target.ID)
	req := httptest.NewRequest(http.MethodPost, base+"/issues/"+source.IssueKey+"/links", bytes.NewReader([]byte(linkBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var link tracker.IssueLinkResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&link))

	req = httptest.NewRequest(http.MethodDelete, base+"/issues/"+source.IssueKey+"/links/"+link.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestHandler_LogTimeAndList проверяет HTTP-слой: HTTP-обработчик log time and список.
func TestHandler_LogTimeAndList(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	base := "/api/projects/" + projectID.String()

	issue := createIssueViaHandler(t, app, projectID, "Time log target")

	body := `{"minutes":45,"description":"implementation"}`
	req := httptest.NewRequest(http.MethodPost, base+"/issues/"+issue.IssueKey+"/time", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/issues/"+issue.IssueKey+"/time-logs", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var logs []tracker.TimeLogResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&logs))
	assert.NotEmpty(t, logs)
	assert.Equal(t, 45, logs[0].Minutes)
}
