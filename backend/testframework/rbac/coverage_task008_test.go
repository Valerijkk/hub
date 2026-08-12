package rbac_test

// TASK-008 coverage: close rbac gaps (InvalidateUserCache, CanCreateTaskInProject,
// DeleteRole/IsNotFound, handler error branches, RequireSystemRole edges).

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/rbac"
)

// TestInvalidateUserCache_AndCanCreateTask_TASK008 covers service methods at 0%.
func TestInvalidateUserCache_AndCanCreateTask_TASK008(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	_, err := f.svc.GetEffectivePermissions(ctx, f.viewerID, f.projectID)
	require.NoError(t, err)
	require.NoError(t, f.svc.InvalidateUserCache(ctx, f.viewerID.String()))

	ok, err := f.svc.CanCreateTaskInProject(ctx, f.ownerID, f.projectID)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = f.svc.CanCreateTaskInProject(ctx, f.viewerID, f.projectID)
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = f.svc.CanCreateTaskInProject(ctx, f.strangerID, f.projectID)
	require.NoError(t, err)
	assert.False(t, ok)
}

// TestEntRepository_DeleteRole_AndIsNotFound_TASK008 covers orphan DeleteRole + helper.
func TestEntRepository_DeleteRole_AndIsNotFound_TASK008(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()
	repo := rbac.NewEntRepository(f.client)

	role, err := f.client.Role.Create().
		SetName("orphan-del").
		SetPermissions(0).
		SetPermissionNames([]string{"project.view"}).
		SetIsSystem(false).
		SetProjectID(f.projectID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, repo.DeleteRole(ctx, role.ID))
	_, err = f.client.Role.Get(ctx, role.ID)
	assert.True(t, ent.IsNotFound(err))

	assert.True(t, rbac.IsNotFound(&ent.NotFoundError{}))
	assert.False(t, rbac.IsNotFound(errors.New("other")))
	assert.False(t, rbac.IsNotFound(nil))
}

// TestHandler_ErrorBranches_TASK008 hits handler 4xx paths under-covered previously.
func TestHandler_ErrorBranches_TASK008(t *testing.T) {
	f := newRBACFixture(t)
	ctx := t.Context()
	app, _ := newRoleHandlerApp(t, f, f.ownerID)

	t.Run("create bad json", func(t *testing.T) {
		req := httptest.NewRequest(fiber.MethodPost, "/api/projects/"+f.projectID.String()+"/roles", bytes.NewReader([]byte("{")))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 5*time.Second)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("create unknown permission", func(t *testing.T) {
		body, _ := json.Marshal(rbac.CustomRoleRequest{Name: "X", Permissions: []string{"not.real"}})
		req := httptest.NewRequest(fiber.MethodPost, "/api/projects/"+f.projectID.String()+"/roles", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 5*time.Second)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("list bad project id", func(t *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, "/api/projects/not-a-uuid/roles", nil)
		resp, err := app.Test(req, 5*time.Second)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, fiber.StatusOK, resp.StatusCode)
	})

	role, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name: "ErrRole", Permissions: []string{"project.view"},
	})
	require.NoError(t, err)
	roleID := role.ID

	otherProj, err := f.client.Project.Create().
		SetName("Other").
		SetKey(randomProjectKey()).
		Save(ctx)
	require.NoError(t, err)
	otherRole, err := f.client.Role.Create().
		SetName("foreign").
		SetPermissions(0).
		SetPermissionNames([]string{"project.view"}).
		SetIsSystem(false).
		SetProjectID(otherProj.ID).
		Save(ctx)
	require.NoError(t, err)

	sysRole, err := f.client.Role.Create().
		SetName("system-role").
		SetPermissions(0).
		SetPermissionNames([]string{"project.view"}).
		SetIsSystem(true).
		SetProjectID(f.projectID).
		Save(ctx)
	require.NoError(t, err)

	t.Run("update bad json", func(t *testing.T) {
		req := httptest.NewRequest(fiber.MethodPut, "/api/projects/"+f.projectID.String()+"/roles/"+roleID, bytes.NewReader([]byte("{")))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 5*time.Second)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("update unknown perm", func(t *testing.T) {
		body, _ := json.Marshal(rbac.CustomRoleRequest{Name: "X", Permissions: []string{"nope"}})
		req := httptest.NewRequest(fiber.MethodPut, "/api/projects/"+f.projectID.String()+"/roles/"+roleID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 5*time.Second)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("update project mismatch", func(t *testing.T) {
		body, _ := json.Marshal(rbac.CustomRoleRequest{Name: "X", Permissions: []string{"project.view"}})
		req := httptest.NewRequest(fiber.MethodPut, "/api/projects/"+f.projectID.String()+"/roles/"+otherRole.ID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 5*time.Second)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("update system role", func(t *testing.T) {
		body, _ := json.Marshal(rbac.CustomRoleRequest{Name: "X", Permissions: []string{"project.view"}})
		req := httptest.NewRequest(fiber.MethodPut, "/api/projects/"+f.projectID.String()+"/roles/"+sysRole.ID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 5*time.Second)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run("delete project mismatch", func(t *testing.T) {
		req := httptest.NewRequest(fiber.MethodDelete, "/api/projects/"+f.projectID.String()+"/roles/"+otherRole.ID.String(), nil)
		resp, err := app.Test(req, 5*time.Second)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("delete system role", func(t *testing.T) {
		req := httptest.NewRequest(fiber.MethodDelete, "/api/projects/"+f.projectID.String()+"/roles/"+sysRole.ID.String(), nil)
		resp, err := app.Test(req, 5*time.Second)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run("assign project mismatch", func(t *testing.T) {
		req := httptest.NewRequest(fiber.MethodPost,
			"/api/projects/"+f.projectID.String()+"/roles/"+otherRole.ID.String()+"/assign/"+f.viewerID.String(), nil)
		resp, err := app.Test(req, 5*time.Second)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("unassign project mismatch", func(t *testing.T) {
		req := httptest.NewRequest(fiber.MethodDelete,
			"/api/projects/"+f.projectID.String()+"/roles/"+otherRole.ID.String()+"/assign/"+f.viewerID.String(), nil)
		resp, err := app.Test(req, 5*time.Second)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("delete bad role id", func(t *testing.T) {
		req := httptest.NewRequest(fiber.MethodDelete, "/api/projects/"+f.projectID.String()+"/roles/not-uuid", nil)
		resp, err := app.Test(req, 5*time.Second)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("assign bad user id", func(t *testing.T) {
		req := httptest.NewRequest(fiber.MethodPost,
			"/api/projects/"+f.projectID.String()+"/roles/"+roleID+"/assign/not-uuid", nil)
		resp, err := app.Test(req, 5*time.Second)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.NotEqual(t, fiber.StatusOK, resp.StatusCode)
	})
}

// TestRequireSystemRole_Edges_TASK008 covers auth/parse/role-mismatch branches.
func TestRequireSystemRole_Edges_TASK008(t *testing.T) {
	memberID := uuid.New()
	adminID := uuid.New()
	repo := &systemRoleRepo{
		users: map[uuid.UUID]*ent.User{
			memberID: {ID: memberID, SystemRole: user.SystemRoleMember},
			adminID:  {ID: adminID, SystemRole: user.SystemRoleAdmin},
		},
	}

	t.Run("missing userID", func(t *testing.T) {
		app := fiber.New()
		app.Get("/x", func(c fiber.Ctx) error { return c.SendString("ok") }, rbac.RequireSystemRole("admin", repo))
		status, _ := doGet(t, app, "/x")
		assert.Equal(t, fiber.StatusUnauthorized, status)
	})

	t.Run("invalid userID", func(t *testing.T) {
		app := newAppWithUserID("not-uuid", rbac.RequireSystemRole("admin", repo), "/x")
		status, _ := doGet(t, app, "/x")
		assert.Equal(t, fiber.StatusUnauthorized, status)
	})

	t.Run("user not found", func(t *testing.T) {
		app := newAppWithUserID(uuid.New().String(), rbac.RequireSystemRole("admin", repo), "/x")
		status, _ := doGet(t, app, "/x")
		assert.Equal(t, fiber.StatusUnauthorized, status)
	})

	t.Run("exact role match non-admin", func(t *testing.T) {
		guard := rbac.RequireSystemRole("member", repo)
		app := newAppWithUserID(memberID.String(), guard, "/x")
		status, body := doGet(t, app, "/x")
		assert.Equal(t, fiber.StatusOK, status)
		assert.Equal(t, "ok", body)

		app2 := newAppWithUserID(adminID.String(), guard, "/x")
		status, _ = doGet(t, app2, "/x")
		assert.Equal(t, fiber.StatusForbidden, status)
	})
}

// TestUpdateCustomRole_MismatchAndSystem_TASK008 service-level mismatch/system paths.
func TestUpdateCustomRole_MismatchAndSystem_TASK008(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	otherProj, err := f.client.Project.Create().
		SetName("Other2").
		SetKey(randomProjectKey()).
		Save(ctx)
	require.NoError(t, err)
	foreign, err := f.client.Role.Create().
		SetName("foreign2").
		SetPermissions(0).
		SetPermissionNames([]string{"project.view"}).
		SetIsSystem(false).
		SetProjectID(otherProj.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = f.svc.UpdateCustomRole(ctx, f.projectID, foreign.ID, rbac.CustomRoleRequest{
		Name: "x", Permissions: []string{"project.view"},
	})
	require.ErrorIs(t, err, rbac.ErrRoleProjectMismatch)

	sys, err := f.client.Role.Create().
		SetName("sys2").
		SetPermissions(0).
		SetPermissionNames([]string{"project.view"}).
		SetIsSystem(true).
		SetProjectID(f.projectID).
		Save(ctx)
	require.NoError(t, err)
	_, err = f.svc.UpdateCustomRole(ctx, f.projectID, sys.ID, rbac.CustomRoleRequest{
		Name: "x", Permissions: []string{"project.view"},
	})
	require.ErrorIs(t, err, rbac.ErrSystemRoleUpdate)

	require.ErrorIs(t, f.svc.DeleteCustomRole(ctx, f.projectID, sys.ID), rbac.ErrSystemRoleDelete)
	require.ErrorIs(t, f.svc.DeleteCustomRole(ctx, f.projectID, foreign.ID), rbac.ErrRoleProjectMismatch)
}

// TestCustomRole_LegacyBitmaskResponse_TASK008 covers toCustomRoleResponse empty names branch.
func TestCustomRole_LegacyBitmaskResponse_TASK008(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	perm := rbac.ExportPermissionsFromNames([]string{"task.create", "project.view"})
	role, err := f.client.Role.Create().
		SetName("legacy-names").
		SetPermissions(uint64(perm)).
		SetPermissionNames([]string{}). // empty → derive from bitmask
		SetIsSystem(false).
		SetProjectID(f.projectID).
		Save(ctx)
	require.NoError(t, err)

	list, err := f.svc.ListCustomRoles(ctx, f.projectID)
	require.NoError(t, err)
	var found bool
	for _, r := range list {
		if r.ID == role.ID.String() {
			found = true
			assert.NotEmpty(t, r.PermList)
			assert.Contains(t, r.PermList, "task.create")
		}
	}
	assert.True(t, found)
}

// TestClearCustomRole_NotFoundRole_TASK008 maps missing role → ErrRoleProjectMismatch.
func TestClearCustomRole_NotFoundRole_TASK008(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()
	err := f.svc.ClearCustomRole(ctx, f.projectID, f.viewerID, uuid.New())
	require.ErrorIs(t, err, rbac.ErrRoleProjectMismatch)
}

// TestGetEffectivePermissions_CustomRoleBitmaskOnly_TASK008 empty PermissionNames fallback.
func TestGetEffectivePermissions_CustomRoleBitmaskOnly_TASK008(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	perm := rbac.ExportPermissionsFromNames([]string{"task.create"})
	role, err := f.client.Role.Create().
		SetName("bitmask-only").
		SetPermissions(uint64(perm)).
		SetPermissionNames(nil).
		SetIsSystem(false).
		SetProjectID(f.projectID).
		Save(ctx)
	require.NoError(t, err)

	_, err = f.client.ProjectMember.Create().
		SetUserID(f.strangerID).
		SetProjectID(f.projectID).
		SetProjectRole("viewer").
		SetCustomRoleID(role.ID).
		Save(ctx)
	require.NoError(t, err)

	got, err := f.svc.GetEffectivePermissions(ctx, f.strangerID, f.projectID)
	require.NoError(t, err)
	assert.True(t, got.Has(rbac.PermTaskCreate))
}
