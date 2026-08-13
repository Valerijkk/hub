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

func setupTrackerHandlerApp(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID, *mockRepo) {
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
	proj.Get("/issues", h.ListIssues)
	proj.Get("/issues/:issueKey", h.GetIssue)
	proj.Put("/issues/:issueKey", h.UpdateIssue)
	proj.Put("/issues/:issueKey/position", h.MoveIssue)
	proj.Get("/issues/:issueKey/activity", h.GetActivity)

	proj.Post("/issues/:issueKey/comments", h.CreateComment)
	proj.Get("/issues/:issueKey/comments", h.ListComments)

	proj.Post("/issues/:issueKey/links", h.CreateIssueLink)
	proj.Get("/issues/:issueKey/links", h.ListIssueLinks)

	proj.Post("/issues/:issueKey/timer/start", h.StartTimer)
	proj.Post("/issues/:issueKey/timer/stop", h.StopTimer)

	proj.Get("/labels", h.ListLabels)
	proj.Post("/labels", h.CreateLabel)
	proj.Get("/workflow/statuses", h.ListWorkflowStatuses)
	proj.Get("/workflow/transitions", h.ListWorkflowTransitions)

	proj.Get("/boards", h.ListBoards)
	proj.Post("/boards", h.CreateBoard)
	proj.Get("/boards/:boardId", h.GetBoardData)

	return app, projectID, userID, repo
}

func createIssueViaHandler(t *testing.T, app *fiber.App, projectID uuid.UUID, title string) tracker.IssueResponse {
	t.Helper()
	body := fmt.Sprintf(`{"title":%q,"issue_type":"task","priority":"medium"}`, title)
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/issues", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created tracker.IssueResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEmpty(t, created.IssueKey)
	return created
}

// TestHandler_CreateAndGetIssue проверяет HTTP-слой: HTTP-обработчик создание and получение issue.
func TestHandler_CreateAndGetIssue(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)

	body := `{"title":"tracker.Handler issue","issue_type":"task","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/issues", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created tracker.IssueResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEmpty(t, created.IssueKey)

	req = httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/issues/"+created.IssueKey, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ListIssues проверяет HTTP-слой: HTTP-обработчик список issues.
func TestHandler_ListIssues(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)

	createIssueViaHandler(t, app, projectID, "Listed issue")

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/issues?limit=10", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_UpdateIssue проверяет HTTP-слой: HTTP-обработчик обновление issue.
func TestHandler_UpdateIssue(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)
	created := createIssueViaHandler(t, app, projectID, "Original title")

	body := fmt.Sprintf(`{"title":"Updated title","version":%d}`, created.Version)
	req := httptest.NewRequest(http.MethodPut, "/api/projects/"+projectID.String()+"/issues/"+created.IssueKey, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated tracker.IssueResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&updated))
	assert.Equal(t, "Updated title", updated.Title)
}

// TestHandler_CreateAndListComments проверяет HTTP-слой: HTTP-обработчик создание and список comments.
func TestHandler_CreateAndListComments(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)
	created := createIssueViaHandler(t, app, projectID, "Comment target")

	body := `{"body_json":{"text":"hello"},"body_text":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/issues/"+created.IssueKey+"/comments", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/issues/"+created.IssueKey+"/comments", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var comments []tracker.CommentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&comments))
	assert.Len(t, comments, 1)
	assert.Equal(t, "hello", comments[0].BodyText)
}

// TestHandler_GetActivity проверяет HTTP-слой: HTTP-обработчик получение activity.
func TestHandler_GetActivity(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)
	created := createIssueViaHandler(t, app, projectID, "Activity issue")

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/issues/"+created.IssueKey+"/activity", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var activity []tracker.ActivityLogResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&activity))
	assert.NotEmpty(t, activity)
}

// TestHandler_MoveIssue проверяет HTTP-слой: HTTP-обработчик move issue.
func TestHandler_MoveIssue(t *testing.T) {
	app, projectID, _, repo := setupTrackerHandlerApp(t)
	created := createIssueViaHandler(t, app, projectID, "Move me")

	defaultID := repo.getDefaultStatusID()
	inProgressID := repo.addStatusWithTransition("In Progress", workflowstatus.CategoryInProgress, defaultID)

	body := fmt.Sprintf(`{"status_id":%q,"rank":"aaa|bbb"}`, inProgressID.String())
	req := httptest.NewRequest(http.MethodPut, "/api/projects/"+projectID.String()+"/issues/"+created.IssueKey+"/position", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var moved tracker.IssueResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&moved))
	assert.Equal(t, "In Progress", moved.Status.Name)
}

// TestHandler_ListWorkflowTransitions проверяет HTTP-слой: HTTP-обработчик список workflow transitions.
func TestHandler_ListWorkflowTransitions(t *testing.T) {
	app, projectID, _, repo := setupTrackerHandlerApp(t)
	repo.addStatusWithTransition("Doing", workflowstatus.CategoryInProgress, repo.getDefaultStatusID())

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/workflow/transitions", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_CreateAndListIssueLinks проверяет HTTP-слой: HTTP-обработчик создание and список issue links.
func TestHandler_CreateAndListIssueLinks(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)
	source := createIssueViaHandler(t, app, projectID, "Blocker")
	target := createIssueViaHandler(t, app, projectID, "Blocked")

	body := fmt.Sprintf(`{"target_issue_id":%q,"link_type":"blocks"}`, target.ID)
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/issues/"+source.IssueKey+"/links", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/issues/"+source.IssueKey+"/links", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var links []tracker.IssueLinkResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&links))
	assert.Len(t, links, 1)
	assert.Equal(t, "blocks", links[0].LinkType)
}

// TestHandler_StartAndStopTimer проверяет HTTP-слой: HTTP-обработчик start and stop timer.
func TestHandler_StartAndStopTimer(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)
	created := createIssueViaHandler(t, app, projectID, "Timer issue")

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/issues/"+created.IssueKey+"/timer/start", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/issues/"+created.IssueKey+"/timer/stop", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ListWorkflowStatuses проверяет HTTP-слой: HTTP-обработчик список workflow statuses.
func TestHandler_ListWorkflowStatuses(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/workflow/statuses", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_CreateLabel проверяет HTTP-слой: HTTP-обработчик создание label.
func TestHandler_CreateLabel(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)

	body := `{"name":"handler-label","color":"#FF0000"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/labels", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// TestHandler_BoardFlow проверяет HTTP-слой: HTTP-обработчик board поток.
func TestHandler_BoardFlow(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)

	body := `{"name":"tracker.Handler Board","filters":{}}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/boards", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var board tracker.BoardResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&board))
	require.NotEmpty(t, board.ID)

	req = httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/boards", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/boards/"+board.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Unauthorized проверяет отклонение при HTTP-обработчик неавторизованный.
func TestHandler_Unauthorized(t *testing.T) {
	svc, _ := newTestService()
	h := tracker.NewHandler(svc, nil)
	app := fiber.New()
	api := app.Group("/api")
	projectID := uuid.New()
	api.Post("/projects/:projectId/issues", h.CreateIssue)

	body := `{"title":"No auth","issue_type":"task"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/issues", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
