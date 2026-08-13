package tracker_test

// Тесты HTTP-обработчика трекера задач (волна 119): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func wave119CreateIssue(t *testing.T, app *fiber.App, projectID uuid.UUID) tracker.IssueResponse {
	t.Helper()
	base := "/api/projects/" + projectID.String()
	body := `{"title":"Wave119 issue","issue_type":"task","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base+"/issues/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created tracker.IssueResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	return created
}

// TestHandler_Wave119_ListCommentsAndActivity проверяет HTTP-слой: HTTP-обработчик wave 119 список comments and activity.
func TestHandler_Wave119_ListCommentsAndActivity(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	issue := wave119CreateIssue(t, app, projectID)
	base := "/api/projects/" + projectID.String() + "/issues/" + issue.IssueKey

	body := `{"body":"Wave119 comment"}`
	req := httptest.NewRequest(http.MethodPost, base+"/comments", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/comments", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/activity", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave119_MessageLinksValidation проверяет HTTP-слой: HTTP-обработчик wave 119 message links валидация.
func TestHandler_Wave119_MessageLinksValidation(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	issue := wave119CreateIssue(t, app, projectID)
	base := "/api/projects/" + projectID.String() + "/issues/" + issue.IssueKey + "/message-links"

	req := httptest.NewRequest(http.MethodGet, base, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, base+"/not-uuid", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave119_ListThreadIssueLinks_Validation проверяет HTTP-слой: HTTP-обработчик wave 119 список thread issue links валидация.
func TestHandler_Wave119_ListThreadIssueLinks_Validation(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	base := "/api/projects/" + projectID.String() + "/message-links"

	req := httptest.NewRequest(http.MethodGet, base, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"?channel_id=bad&message_id=root", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	chID := uuid.New()
	req = httptest.NewRequest(http.MethodGet, base+"?channel_id="+chID.String()+"&message_id=root-msg", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave119_UpdateComment_InvalidJSON проверяет отклонение при HTTP-обработчик wave 119 обновление comment некорректный json.
func TestHandler_Wave119_UpdateComment_InvalidJSON(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	issue := wave119CreateIssue(t, app, projectID)
	base := "/api/projects/" + projectID.String() + "/issues/" + issue.IssueKey + "/comments/" + uuid.New().String()

	req := httptest.NewRequest(http.MethodPut, base, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave119_ToggleReaction_InvalidJSON проверяет отклонение при HTTP-обработчик wave 119 toggle reaction некорректный json.
func TestHandler_Wave119_ToggleReaction_InvalidJSON(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	issue := wave119CreateIssue(t, app, projectID)
	base := "/api/projects/" + projectID.String() + "/issues/" + issue.IssueKey + "/comments/" + uuid.New().String() + "/reactions"

	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave119_UpdateAutomationRule_NotFound проверяет HTTP-слой: HTTP-обработчик wave 119 обновление автоматизация rule not found.
func TestHandler_Wave119_UpdateAutomationRule_NotFound(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	base := "/api/projects/" + projectID.String() + "/automation-rules/" + uuid.New().String()

	body := `{"name":"Renamed"}`
	req := httptest.NewRequest(http.MethodPatch, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave119_DeleteAutomationRule_NotFound проверяет HTTP-слой: HTTP-обработчик wave 119 удаление автоматизация rule not found.
func TestHandler_Wave119_DeleteAutomationRule_NotFound(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	base := "/api/projects/" + projectID.String() + "/automation-rules/" + uuid.New().String()

	req := httptest.NewRequest(http.MethodDelete, base, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusNoContent, resp.StatusCode)
}

// TestHandler_Wave119_LogTime_InvalidJSON проверяет отклонение при HTTP-обработчик wave 119 log time некорректный json.
func TestHandler_Wave119_LogTime_InvalidJSON(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	issue := wave119CreateIssue(t, app, projectID)
	base := "/api/projects/" + projectID.String() + "/issues/" + issue.IssueKey + "/time"

	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
