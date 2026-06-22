package tracker_test

// Тесты HTTP-обработчика трекера задач (волна 121): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Фаза 121 — покрытие маршрутов доски..

func TestHandler_RegisterRoutes_ListBoards(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/boards/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_RegisterRoutes_GetBoardAfterCreate проверяет HTTP-слой: HTTP-обработчик register маршруты получение board after создание.
func TestHandler_RegisterRoutes_GetBoardAfterCreate(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	base := "/api/projects/" + projectID.String() + "/boards/"

	createBody := `{"name":"Kanban"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var board struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&board))

	req = httptest.NewRequest(http.MethodGet, base+board.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_RegisterRoutes_GetActivity проверяет HTTP-слой: HTTP-обработчик register маршруты получение activity.
func TestHandler_RegisterRoutes_GetActivity(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	base := "/api/projects/" + projectID.String()

	createBody := `{"title":"Activity issue","issue_type":"task","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base+"/issues/", bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var issue tracker.IssueResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&issue))

	req = httptest.NewRequest(http.MethodGet, base+"/issues/"+issue.IssueKey+"/activity", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_RegisterRoutes_ListComments проверяет HTTP-слой: HTTP-обработчик register маршруты список comments.
func TestHandler_RegisterRoutes_ListComments(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	base := "/api/projects/" + projectID.String()

	createBody := `{"title":"Comment target","issue_type":"task","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base+"/issues/", bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var issue tracker.IssueResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&issue))

	req = httptest.NewRequest(http.MethodGet, base+"/issues/"+issue.IssueKey+"/comments", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
