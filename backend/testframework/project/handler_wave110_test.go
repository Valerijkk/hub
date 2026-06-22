package project_test

// Тесты HTTP-обработчика проектов (волна 110): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
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

// TestHandler_DeleteProject_Success проверяет HTTP-слой: HTTP-обработчик удаление project успех.
func TestHandler_DeleteProject_Success(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "DELP")

	req := httptest.NewRequest(http.MethodDelete, "/api/projects/"+created.ID, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/projects/trash", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var trash []project.ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&trash))
	require.NotEmpty(t, trash)
}

// TestHandler_DeleteProject_NotFound проверяет HTTP-слой: HTTP-обработчик удаление project not found.
func TestHandler_DeleteProject_NotFound(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/projects/"+uuid.New().String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_CreateInvite_InvalidJSON проверяет отклонение при HTTP-обработчик создание приглашение некорректный json.
func TestHandler_CreateInvite_InvalidJSON(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "INVB")

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/invites", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_JoinByInvite_InvalidCode проверяет отклонение при HTTP-обработчик join by приглашение некорректный code.
func TestHandler_JoinByInvite_InvalidCode(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	body := `{"code":"DOESNOTEXIST"}`
	req := httptest.NewRequest(http.MethodPost, "/api/invites/join", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_CreateTeam_InvalidJSON проверяет отклонение при HTTP-обработчик создание team некорректный json.
func TestHandler_CreateTeam_InvalidJSON(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/teams/", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_AddTeamMember_InvalidUserID проверяет отклонение при HTTP-обработчик add team участник некорректный user id.
func TestHandler_AddTeamMember_InvalidUserID(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	team := createTeamViaHandler(t, app, "BadMember")

	req := httptest.NewRequest(http.MethodPost, "/api/teams/"+team.ID+"/members",
		bytes.NewReader([]byte(`{"user_id":"not-a-uuid"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusCreated, resp.StatusCode)
}

// TestHandler_RestoreProject_AfterDelete проверяет HTTP-слой: HTTP-обработчик restore project after удаление.
func TestHandler_RestoreProject_AfterDelete(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "RSTR")

	req := httptest.NewRequest(http.MethodDelete, "/api/projects/"+created.ID, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/restore", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var restored project.ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&restored))
	assert.Equal(t, "active", restored.Status)
}
