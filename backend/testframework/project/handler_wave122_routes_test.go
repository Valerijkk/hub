package project_test

// Тесты HTTP-обработчика проектов (волна 122): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/invite"
	"github.com/valerii/hub/backend/ent/projectmember"
)

type projectRoutesEnv struct {
	app      *fiber.App
	ownerID  uuid.UUID
	memberID uuid.UUID
	repo     *project.EntRepository
	ctx      context.Context
}

func newProjectRoutesEnv(t *testing.T) projectRoutesEnv {
	t.Helper()
	dsn := "file:prroutes_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:8] + "@test.com").
		SetName("Owner").
		Save(ctx)
	require.NoError(t, err)

	member, err := client.User.Create().
		SetEmail("member-" + uuid.New().String()[:8] + "@test.com").
		SetName("Member").
		Save(ctx)
	require.NoError(t, err)

	repo := project.NewEntRepository(client, nil)
	svc := project.NewProjectService(repo, nil, nil)
	h := project.NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	return projectRoutesEnv{app: app, ownerID: owner.ID, memberID: member.ID, repo: repo, ctx: ctx}
}

func createProjectViaRoutes(t *testing.T, env projectRoutesEnv, key string) project.ProjectResponse {
	t.Helper()
	body := fmt.Sprintf(`{"name":"%s Project","key":"%s"}`, key, key)
	req := httptest.NewRequest(http.MethodPost, "/api/projects/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created project.ProjectResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	return created
}

// Фаза 122 — ent-backed handler routes for archive/trash/teams/invites..

func TestHandler_Wave122_Routes_ArchiveUnarchive(t *testing.T) {
	env := newProjectRoutesEnv(t)
	created := createProjectViaRoutes(t, env, "ARC")

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/archive", nil)
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/unarchive", nil)
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave122_Routes_TrashRestore проверяет HTTP-слой: HTTP-обработчик wave 122 маршруты корзина restore.
func TestHandler_Wave122_Routes_TrashRestore(t *testing.T) {
	env := newProjectRoutesEnv(t)
	created := createProjectViaRoutes(t, env, "TRH")

	req := httptest.NewRequest(http.MethodDelete, "/api/projects/"+created.ID, nil)
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, "/api/projects/trash", nil)
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/restore", nil)
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave122_Routes_MembersAndTeams проверяет HTTP-слой: HTTP-обработчик wave 122 маршруты members and teams.
func TestHandler_Wave122_Routes_MembersAndTeams(t *testing.T) {
	env := newProjectRoutesEnv(t)
	created := createProjectViaRoutes(t, env, "MEM")

	addBody := fmt.Sprintf(`{"user_id":"%s","role":"developer"}`, env.memberID.String())
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/members", bytes.NewReader([]byte(addBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, "/api/projects/"+created.ID+"/members", nil)
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	teamBody := `{"name":"Route Team","description":"handlers"}`
	req = httptest.NewRequest(http.MethodPost, "/api/teams/", bytes.NewReader([]byte(teamBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	var team project.TeamResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&team))
	resp.Body.Close()

	memberBody := fmt.Sprintf(`{"user_id":"%s"}`, env.memberID.String())
	req = httptest.NewRequest(http.MethodPost, "/api/teams/"+team.ID+"/members", bytes.NewReader([]byte(memberBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, "/api/teams/", nil)
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, "/api/teams/"+team.ID+"/members/"+env.memberID.String(), nil)
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestHandler_Wave122_Routes_CreateAndListInvites проверяет HTTP-слой: HTTP-обработчик wave 122 маршруты создание and список invites.
func TestHandler_Wave122_Routes_CreateAndListInvites(t *testing.T) {
	env := newProjectRoutesEnv(t)
	created := createProjectViaRoutes(t, env, "INV")

	invBody := `{"role":"developer","max_uses":5}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/invites", bytes.NewReader([]byte(invBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/projects/"+created.ID+"/invites", nil)
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave122_Routes_RevokeInvite проверяет HTTP-слой: HTTP-обработчик wave 122 маршруты revoke приглашение.
func TestHandler_Wave122_Routes_RevokeInvite(t *testing.T) {
	env := newProjectRoutesEnv(t)
	proj, err := env.repo.Create(env.ctx, "Revoke", "RVK", "", env.ownerID)
	require.NoError(t, err)

	exp := time.Now().Add(24 * time.Hour)
	inv, err := env.repo.CreateInvite(env.ctx, proj.ID, env.ownerID, "JOINME1", invite.ProjectRoleDeveloper, 1, &exp)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/api/projects/"+proj.ID.String()+"/invites/"+inv.ID.String(), nil)
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestHandler_Wave122_Routes_UpdateMemberRole проверяет HTTP-слой: HTTP-обработчик wave 122 маршруты обновление участник role.
func TestHandler_Wave122_Routes_UpdateMemberRole(t *testing.T) {
	env := newProjectRoutesEnv(t)
	created := createProjectViaRoutes(t, env, "ROL")

	_, err := env.repo.AddMember(env.ctx, uuid.MustParse(created.ID), env.memberID, projectmember.ProjectRoleDeveloper)
	require.NoError(t, err)

	body := `{"role":"tester"}`
	req := httptest.NewRequest(http.MethodPut, "/api/projects/"+created.ID+"/members/"+env.memberID.String(), bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, "/api/projects/"+created.ID+"/members/"+env.memberID.String(), nil)
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
