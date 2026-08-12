package imex_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/internal/imex"
	"github.com/valerii/hub/backend/internal/rbac"
)

// TestImexExport_RequiresProjectMembership_A5_002 — auth alone is not enough;
// non-member of :projectId must get 403 (PermProjectView).
func TestImexExport_RequiresProjectMembership_A5_002(t *testing.T) {
	dsn := "file:imex_a5002_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))
	ctx := t.Context()

	member, err := client.User.Create().
		SetEmail("member-" + uuid.New().String()[:8] + "@test.com").
		SetName("Member").
		Save(ctx)
	require.NoError(t, err)
	outsider, err := client.User.Create().
		SetEmail("out-" + uuid.New().String()[:8] + "@test.com").
		SetName("Outsider").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Imex RBAC").
		SetKey("IX").
		Save(ctx)
	require.NoError(t, err)
	_, err = client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetUserID(member.ID).
		SetProjectID(proj.ID).
		SetProjectRole("developer").
		Save(ctx)
	require.NoError(t, err)

	h := imex.NewHandler(imex.NewImExService(imex.NewEntRepository(client)))
	app := fiber.New()
	api := app.Group("/api")

	// Outsider JWT — member of nothing in this project.
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", outsider.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+proj.ID.String()+"/export/issues?format=csv", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "A5-002: non-member must not export")
}

// TestImexExport_MemberAllowed_A5_002 — project member with view can export.
func TestImexExport_MemberAllowed_A5_002(t *testing.T) {
	dsn := "file:imex_a5002ok_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))
	ctx := t.Context()

	member, err := client.User.Create().
		SetEmail("ok-" + uuid.New().String()[:8] + "@test.com").
		SetName("Member").
		Save(ctx)
	require.NoError(t, err)
	proj, err := client.Project.Create().
		SetName("Imex OK").
		SetKey("OK").
		Save(ctx)
	require.NoError(t, err)
	_, err = client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetUserID(member.ID).
		SetProjectID(proj.ID).
		SetProjectRole("developer").
		Save(ctx)
	require.NoError(t, err)

	h := imex.NewHandler(imex.NewImExService(imex.NewEntRepository(client)))
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", member.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+proj.ID.String()+"/export/issues?format=csv", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
