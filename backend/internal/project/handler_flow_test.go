package project

import (
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
	h := NewHandler(svc)
	userID := uuid.New()

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, userID
}

func TestHandler_CreateAndListProjects(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	body := `{"name":"Handler Project","key":"HPRJ","description":"via handler test"}`
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

func TestHandler_GetProject(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	body := `{"name":"Get Me","key":"GETM"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)

	var created ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, "/api/projects/"+created.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_ListTrash(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/trash", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_Unauthorized(t *testing.T) {
	svc, _, _ := newTestService()
	h := NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest(http.MethodGet, "/api/projects/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func createProjectViaHandler(t *testing.T, app *fiber.App, key string) ProjectResponse {
	t.Helper()
	body := `{"name":"Handler Project","key":"` + key + `","description":"via handler test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	return created
}

func TestHandler_UpdateProject(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "UPDH")

	body := `{"name":"Updated Name","description":"new desc"}`
	req := httptest.NewRequest(http.MethodPut, "/api/projects/"+created.ID, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&updated))
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "new desc", updated.Description)
}

func TestHandler_ArchiveProject(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "ARCH")

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/archive", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var archived ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&archived))
	assert.Equal(t, "archived", archived.Status)
}

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

	var member MemberResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&member))
	assert.Equal(t, memberID.String(), member.UserID)
	assert.Equal(t, "developer", member.Role)
}

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

	var invite InviteResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&invite))
	assert.NotEmpty(t, invite.Code)
	assert.Equal(t, "developer", invite.Role)
	_ = userID
}

func TestHandler_ListMembers(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "LSTM")

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+created.ID+"/members", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_ListInvites(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "INVL")

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+created.ID+"/invites", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

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

	var restored ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&restored))
	assert.Equal(t, "active", restored.Status)
}

func TestHandler_CreateTeam(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	body := `{"name":"Platform","description":"infra"}`
	req := httptest.NewRequest(http.MethodPost, "/api/teams/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var team TeamResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&team))
	assert.Equal(t, "Platform", team.Name)
}

func TestHandler_ListTeams(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/teams/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

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
	h := NewHandler(svc)
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

	var joined ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&joined))
	assert.Equal(t, created.ID, joined.ID)
}

func createTeamViaHandler(t *testing.T, app *fiber.App, name string) TeamResponse {
	t.Helper()
	body := `{"name":"` + name + `","description":"team desc"}`
	req := httptest.NewRequest(http.MethodPost, "/api/teams/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var team TeamResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&team))
	return team
}

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
