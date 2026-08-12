package project_test

// Тесты HTTP-обработчика проектов: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/invite"
)

func setupProjectHandlerApp(t *testing.T) (*fiber.App, uuid.UUID) {
	t.Helper()
	svc, _, _ := newTestService()
	h := project.NewHandler(svc)
	userID := uuid.New()
	rbacSvc := newFounderBypassRBAC(t, userID)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)
	return app, userID
}

// TestHandler_CreateAndListProjects проверяет HTTP-слой: HTTP-обработчик создание and список projects.
func TestHandler_CreateAndListProjects(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	body := `{"name":"project.Handler Project","key":"HPRJ","description":"via handler test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/projects/", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_GetProject проверяет HTTP-слой: HTTP-обработчик получение project.
func TestHandler_GetProject(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	body := `{"name":"Get Me","key":"GETM"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)

	var created project.ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, "/api/projects/"+created.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ListTrash проверяет HTTP-слой: HTTP-обработчик список корзина.
func TestHandler_ListTrash(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/trash", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Unauthorized проверяет отклонение при HTTP-обработчик неавторизованный.
func TestHandler_Unauthorized(t *testing.T) {
	svc, _, _ := newTestService()
	h := project.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() }, newFounderBypassRBAC(t, uuid.New()))

	req := httptest.NewRequest(http.MethodGet, "/api/projects/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func createProjectViaHandler(t *testing.T, app *fiber.App, key string) project.ProjectResponse {
	t.Helper()
	body := `{"name":"project.Handler Project","key":"` + key + `","description":"via handler test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created project.ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	return created
}

// TestHandler_UpdateProject проверяет HTTP-слой: HTTP-обработчик обновление project.
func TestHandler_UpdateProject(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "UPDH")

	body := `{"name":"Updated Name","description":"new desc","version":1}`
	req := httptest.NewRequest(http.MethodPut, "/api/projects/"+created.ID, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated project.ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&updated))
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "new desc", updated.Description)
}

// TestHandler_ArchiveProject проверяет HTTP-слой: HTTP-обработчик archive project.
func TestHandler_ArchiveProject(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "ARCH")

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/archive", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var archived project.ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&archived))
	assert.Equal(t, "archived", archived.Status)
}

// TestHandler_AddMember проверяет HTTP-слой: HTTP-обработчик add участник.
func TestHandler_AddMember(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "ADDM")
	memberID := uuid.New()

	body := `{"user_id":"` + memberID.String() + `","role":"developer"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/members", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var member project.MemberResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&member))
	assert.Equal(t, memberID.String(), member.UserID)
	assert.Equal(t, "developer", member.Role)
}

// TestHandler_AddMember_InvalidUserID проверяет отклонение при HTTP-обработчик add участник некорректный user id.
func TestHandler_AddMember_InvalidUserID(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "BADU")

	body := `{"user_id":"not-uuid","role":"developer"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/members", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_RemoveMember проверяет HTTP-слой: HTTP-обработчик remove участник.
func TestHandler_RemoveMember(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "RMEM")
	memberID := uuid.New()

	addBody := `{"user_id":"` + memberID.String() + `","role":"tester"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/members", bytes.NewReader([]byte(addBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodDelete, "/api/projects/"+created.ID+"/members/"+memberID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

// TestHandler_CreateInvite проверяет HTTP-слой: HTTP-обработчик создание приглашение.
func TestHandler_CreateInvite(t *testing.T) {
	app, userID := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "INVH")

	body := `{"role":"developer","max_uses":5}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/invites", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var invite project.InviteResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&invite))
	assert.NotEmpty(t, invite.Code)
	assert.Equal(t, "developer", invite.Role)
	_ = userID
}

// TestHandler_ListMembers проверяет HTTP-слой: HTTP-обработчик список members.
func TestHandler_ListMembers(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "LSTM")

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+created.ID+"/members", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ListInvites проверяет HTTP-слой: HTTP-обработчик список invites.
func TestHandler_ListInvites(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "INVL")

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+created.ID+"/invites", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_UnarchiveProject проверяет HTTP-слой: HTTP-обработчик unarchive project.
func TestHandler_UnarchiveProject(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "UNAR")

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/archive", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/unarchive", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var restored project.ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&restored))
	assert.Equal(t, "active", restored.Status)
}

// TestHandler_CreateTeam проверяет HTTP-слой: HTTP-обработчик создание team.
func TestHandler_CreateTeam(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	body := `{"name":"Platform","description":"infra"}`
	req := httptest.NewRequest(http.MethodPost, "/api/teams/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var team project.TeamResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&team))
	assert.Equal(t, "Platform", team.Name)
}

// TestHandler_ListTeams проверяет HTTP-слой: HTTP-обработчик список teams.
func TestHandler_ListTeams(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/teams/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_JoinByInvite_Success проверяет HTTP-слой: HTTP-обработчик join by приглашение успех.
func TestHandler_JoinByInvite_Success(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Invite Project", "INVH", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	inv := &ent.Invite{
		ID: uuid.New(), Code: "JOINME", MaxUses: 10, UseCount: 0,
		ProjectRole: invite.ProjectRoleDeveloper, CreatedAt: time.Now(),
		Edges: ent.InviteEdges{Project: repo.projects[projectID]},
	}
	repo.invites["JOINME"] = inv

	joinerID := uuid.New()
	h := project.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", joinerID.String())
		return c.Next()
	}
	invites := api.Group("/invites", authMw)
	invites.Post("/join", h.JoinByInvite)

	body := `{"code":"JOINME"}`
	req := httptest.NewRequest(http.MethodPost, "/api/invites/join", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var joined project.ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&joined))
	assert.Equal(t, created.ID, joined.ID)
}

func createTeamViaHandler(t *testing.T, app *fiber.App, name string) project.TeamResponse {
	t.Helper()
	body := `{"name":"` + name + `","description":"team desc"}`
	req := httptest.NewRequest(http.MethodPost, "/api/teams/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var team project.TeamResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&team))
	return team
}

// TestHandler_TeamMembersFlow проверяет HTTP-слой: HTTP-обработчик team members поток.
func TestHandler_TeamMembersFlow(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	team := createTeamViaHandler(t, app, "DevOps")
	memberID := uuid.New()

	body := `{"user_id":"` + memberID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/teams/"+team.ID+"/members", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, "/api/teams/"+team.ID+"/members/"+memberID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestHandler_AddTeamToProject проверяет HTTP-слой: HTTP-обработчик add team to project.
func TestHandler_AddTeamToProject(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	project := createProjectViaHandler(t, app, "TMAP")
	team := createTeamViaHandler(t, app, "Mappers")

	req := httptest.NewRequest(http.MethodPost, "/api/teams/"+team.ID+"/projects/"+project.ID, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_DeleteAndRestoreProject проверяет HTTP-слой: HTTP-обработчик удаление and restore project.
func TestHandler_DeleteAndRestoreProject(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "DRES")

	req := httptest.NewRequest(http.MethodDelete, "/api/projects/"+created.ID, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/restore", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var restored project.ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&restored))
	assert.Equal(t, "active", restored.Status)
}

// TestHandler_UpdateMemberRole проверяет HTTP-слой: HTTP-обработчик обновление участник role.
func TestHandler_UpdateMemberRole(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "ROLE")
	memberID := uuid.New()

	addBody := `{"user_id":"` + memberID.String() + `","role":"developer"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/members", bytes.NewReader([]byte(addBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	body := `{"role":"tester"}`
	req = httptest.NewRequest(http.MethodPut, "/api/projects/"+created.ID+"/members/"+memberID.String(), bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var member project.MemberResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&member))
	assert.Equal(t, "tester", member.Role)
}

// TestHandler_RevokeInvite проверяет HTTP-слой: HTTP-обработчик revoke приглашение.
func TestHandler_RevokeInvite(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "REVK")

	body := `{"role":"developer","max_uses":3}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/invites", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)

	var inv project.InviteResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&inv))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodDelete, "/api/projects/"+created.ID+"/invites/"+inv.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}
