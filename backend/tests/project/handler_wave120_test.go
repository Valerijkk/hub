package project_test

// Тесты HTTP-обработчика проектов (волна 120): интеграционные и unit-сценарии testframework.
import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_Wave120_RemoveMember_InvalidUserID проверяет отклонение при HTTP-обработчик wave 120 remove участник некорректный user id.
func TestHandler_Wave120_RemoveMember_InvalidUserID(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "RMMB")

	req := httptest.NewRequest(http.MethodDelete, "/api/projects/"+created.ID+"/members/not-a-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave120_ListTeams_Empty проверяет HTTP-слой: HTTP-обработчик wave 120 список teams пустой.
func TestHandler_Wave120_ListTeams_Empty(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/teams/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave120_CreateInvite_InvalidJSON проверяет отклонение при HTTP-обработчик wave 120 создание приглашение некорректный json.
func TestHandler_Wave120_CreateInvite_InvalidJSON(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "INVJ")

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/invites", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_Wave120_ListMembers_EmptyProject проверяет HTTP-слой: HTTP-обработчик wave 120 список members пустой project.
func TestHandler_Wave120_ListMembers_EmptyProject(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "LMBR")

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+created.ID+"/members", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave120_GetProject_InvalidID проверяет отклонение при HTTP-обработчик wave 120 получение project некорректный id.
func TestHandler_Wave120_GetProject_InvalidID(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/not-a-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
