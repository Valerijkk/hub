package tracker_test

// Тесты HTTP-обработчика трекера задач (волна 103): интеграционные и unit-сценарии testframework.
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
	"github.com/valerii/hub/backend/ent"
)

func setupTrackerCommentsApp(t *testing.T) (*fiber.App, uuid.UUID, string, uuid.UUID) {
	t.Helper()
	svc, repo := newTestService()
	h := tracker.NewHandler(svc, nil)
	projectID := uuid.New()
	userID := uuid.New()

	issueID := uuid.New()
	issueKey := "CMT-1"
	repo.issuesByKey[issueKey] = &ent.Issue{
		ID: issueID, IssueKey: issueKey, Title: "Comment issue",
		Edges: ent.IssueEdges{Project: &ent.Project{ID: projectID}},
	}
	repo.issues[issueID] = repo.issuesByKey[issueKey]

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	proj := api.Group("/projects/:projectId", authMw)
	issues := proj.Group("/issues")
	issues.Post("/:issueKey/comments", h.CreateComment)
	issues.Get("/:issueKey/comments", h.ListComments)
	issues.Put("/:issueKey/comments/:commentId", h.UpdateComment)
	issues.Delete("/:issueKey/comments/:commentId", h.DeleteComment)
	issues.Post("/:issueKey/comments/:commentId/reactions", h.ToggleReaction)
	issues.Get("/:issueKey/activity", h.GetActivity)

	return app, projectID, issueKey, userID
}

// TestHandler_Wave103_UpdateAndDeleteComment проверяет HTTP-слой: HTTP-обработчик wave 103 обновление and удаление comment.
func TestHandler_Wave103_UpdateAndDeleteComment(t *testing.T) {
	app, projectID, issueKey, _ := setupTrackerCommentsApp(t)
	base := fmt.Sprintf("/api/projects/%s/issues/%s/comments", projectID, issueKey)

	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(`{"body_text":"first"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created tracker.CommentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEmpty(t, created.ID)

	upURL := fmt.Sprintf("%s/%s", base, created.ID)
	req = httptest.NewRequest(http.MethodPut, upURL, bytes.NewReader([]byte(`{"body_text":"updated"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated tracker.CommentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&updated))
	assert.Equal(t, "updated", updated.BodyText)

	req = httptest.NewRequest(http.MethodDelete, upURL, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

// TestHandler_CreateComment_InvalidJSON проверяет отклонение при HTTP-обработчик создание comment некорректный json.
func TestHandler_CreateComment_InvalidJSON(t *testing.T) {
	app, projectID, issueKey, _ := setupTrackerCommentsApp(t)
	url := fmt.Sprintf("/api/projects/%s/issues/%s/comments", projectID, issueKey)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_ToggleReaction_InvalidCommentID проверяет отклонение при HTTP-обработчик toggle reaction некорректный comment id.
func TestHandler_ToggleReaction_InvalidCommentID(t *testing.T) {
	app, projectID, issueKey, _ := setupTrackerCommentsApp(t)
	url := fmt.Sprintf("/api/projects/%s/issues/%s/comments/not-uuid/reactions", projectID, issueKey)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(`{"emoji":"👍"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_GetActivity_Empty проверяет HTTP-слой: HTTP-обработчик получение activity пустой.
func TestHandler_GetActivity_Empty(t *testing.T) {
	app, projectID, issueKey, _ := setupTrackerCommentsApp(t)
	url := fmt.Sprintf("/api/projects/%s/issues/%s/activity", projectID, issueKey)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
