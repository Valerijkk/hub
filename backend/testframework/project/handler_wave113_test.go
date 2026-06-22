package project_test

// Тесты HTTP-обработчика проектов (волна 113): интеграционные и unit-сценарии testframework.
import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_UpdateProject_InvalidJSON проверяет отклонение при HTTP-обработчик обновление project некорректный json.
func TestHandler_UpdateProject_InvalidJSON(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "UPDJ")

	req := httptest.NewRequest(http.MethodPut, "/api/projects/"+created.ID, bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_AddTeamToProject_InvalidTeamID проверяет отклонение при HTTP-обработчик add team to project некорректный team id.
func TestHandler_AddTeamToProject_InvalidTeamID(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "ATTP")

	req := httptest.NewRequest(http.MethodPost, "/api/teams/not-a-uuid/projects/"+created.ID, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_RevokeInvite_InvalidInviteID проверяет отклонение при HTTP-обработчик revoke приглашение некорректный приглашение id.
func TestHandler_RevokeInvite_InvalidInviteID(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "RVKI")

	req := httptest.NewRequest(http.MethodDelete, "/api/projects/"+created.ID+"/invites/not-a-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_UpdateMemberRole_InvalidJSON проверяет отклонение при HTTP-обработчик обновление участник role некорректный json.
func TestHandler_UpdateMemberRole_InvalidJSON(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "UMRJ")

	req := httptest.NewRequest(http.MethodPut, "/api/projects/"+created.ID+"/members/"+uuid.New().String(),
		bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_CreateProject_InvalidJSON проверяет отклонение при HTTP-обработчик создание project некорректный json.
func TestHandler_CreateProject_InvalidJSON(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/projects/", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
