package tracker_test

// Тесты HTTP-обработчика трекера задач (волна 95): интеграционные и unit-сценарии testframework.
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

func setupTrackerMessageLinkApp(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID, *mockRepo, string) {
	t.Helper()
	svc, repo := newTestService()
	h := tracker.NewHandler(svc, nil)
	projectID := uuid.New()
	userID := uuid.New()

	issueID := uuid.New()
	issueKey := "MSG-1"
	repo.issuesByKey[issueKey] = &ent.Issue{
		ID: issueID, IssueKey: issueKey, Title: "Linked issue",
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
	issues.Get("/:issueKey/message-links", h.ListIssueMessageLinks)
	issues.Post("/:issueKey/message-links", h.CreateIssueMessageLink)
	issues.Delete("/:issueKey/message-links/:linkId", h.DeleteIssueMessageLink)
	proj.Get("/message-links", h.ListThreadIssueLinks)
	issues.Post("/:issueKey/comments/:commentId/reactions", h.ToggleReaction)
	proj.Delete("/custom-fields/:fieldId", h.DeleteCustomField)

	return app, projectID, userID, repo, issueKey
}

// TestHandler_IssueMessageLinks_CRUD проверяет HTTP-слой: HTTP-обработчик issue message links crud.
func TestHandler_IssueMessageLinks_CRUD(t *testing.T) {
	app, projectID, _, _, issueKey := setupTrackerMessageLinkApp(t)
	base := fmt.Sprintf("/api/projects/%s/issues/%s/message-links", projectID, issueKey)
	channelID := uuid.New()

	body := fmt.Sprintf(`{"channel_id":%q,"message_id":"root-msg-1","mirror_to_thread":true}`, channelID)
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created tracker.IssueMessageLinkResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	assert.Equal(t, "root-msg-1", created.MessageID)

	req = httptest.NewRequest(http.MethodGet, base, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listed []tracker.IssueMessageLinkResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&listed))
	require.Len(t, listed, 1)

	threadURL := fmt.Sprintf("/api/projects/%s/message-links?channel_id=%s&message_id=root-msg-1",
		projectID, channelID)
	req = httptest.NewRequest(http.MethodGet, threadURL, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	delURL := fmt.Sprintf("%s/%s", base, created.ID)
	req = httptest.NewRequest(http.MethodDelete, delURL, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestHandler_ListThreadIssueLinks_Validation проверяет HTTP-слой: HTTP-обработчик список thread issue links валидация.
func TestHandler_ListThreadIssueLinks_Validation(t *testing.T) {
	app, projectID, _, _, _ := setupTrackerMessageLinkApp(t)
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/projects/%s/message-links", projectID), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestHandler_ToggleReaction_Success проверяет HTTP-слой: HTTP-обработчик toggle reaction успех.
func TestHandler_ToggleReaction_Success(t *testing.T) {
	app, projectID, _, _, issueKey := setupTrackerMessageLinkApp(t)
	commentID := uuid.New()
	url := fmt.Sprintf("/api/projects/%s/issues/%s/comments/%s/reactions",
		projectID, issueKey, commentID)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(`{"emoji":"👍"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_DeleteCustomField_Success проверяет HTTP-слой: HTTP-обработчик удаление custom field успех.
func TestHandler_DeleteCustomField_Success(t *testing.T) {
	app, projectID, _, repo, _ := setupTrackerMessageLinkApp(t)
	fieldID := uuid.New()
	repo.customFields[fieldID] = &ent.CustomFieldDefinition{ID: fieldID, Name: "Sprint"}

	url := fmt.Sprintf("/api/projects/%s/custom-fields/%s", projectID, fieldID)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
