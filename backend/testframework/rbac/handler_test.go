package rbac_test

// Тесты HTTP-обработчика RBAC: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/rbac"
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRoleHandlerApp(t *testing.T, f *rbacFixture, userID uuid.UUID) (*fiber.App, *rbac.CustomRoleHandler) {
	t.Helper()
	h := rbac.NewCustomRoleHandler(f.svc)
	app := fiber.New()
	localsMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	api := app.Group("/api", localsMw)
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() })
	return app, h
}

// TestHandler_CreateRole_Success проверяет HTTP-слой: HTTP-обработчик создание role успех.
func TestHandler_CreateRole_Success(t *testing.T) {
	f := newRBACFixture(t)
	app, _ := newRoleHandlerApp(t, f, f.ownerID)

	body, _ := json.Marshal(rbac.CustomRoleRequest{
		Name: "Reviewer", Permissions: []string{"task.create"},
	})
	req := httptest.NewRequest(fiber.MethodPost, "/api/projects/"+f.projectID.String()+"/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

// TestHandler_ListRoles_Success проверяет HTTP-слой: HTTP-обработчик список roles успех.
func TestHandler_ListRoles_Success(t *testing.T) {
	f := newRBACFixture(t)
	ctx := t.Context()
	_, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name: "Listed", Permissions: []string{"project.view"},
	})
	require.NoError(t, err)

	app, _ := newRoleHandlerApp(t, f, f.viewerID)
	req := httptest.NewRequest(fiber.MethodGet, "/api/projects/"+f.projectID.String()+"/roles", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(respBody), "Listed")
}

// TestHandler_UpdateRole_Success проверяет HTTP-слой: HTTP-обработчик обновление role успех.
func TestHandler_UpdateRole_Success(t *testing.T) {
	f := newRBACFixture(t)
	ctx := t.Context()
	role, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name: "Old", Permissions: []string{"project.view"},
	})
	require.NoError(t, err)

	app, _ := newRoleHandlerApp(t, f, f.ownerID)
	body, _ := json.Marshal(rbac.CustomRoleRequest{
		Name: "Renamed", Permissions: []string{"task.create"},
	})
	req := httptest.NewRequest(
		fiber.MethodPut,
		"/api/projects/"+f.projectID.String()+"/roles/"+role.ID,
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// TestHandler_DeleteRole_Success проверяет HTTP-слой: HTTP-обработчик удаление role успех.
func TestHandler_DeleteRole_Success(t *testing.T) {
	f := newRBACFixture(t)
	ctx := t.Context()
	role, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name: "Temp", Permissions: []string{"project.view"},
	})
	require.NoError(t, err)

	app, _ := newRoleHandlerApp(t, f, f.ownerID)
	req := httptest.NewRequest(
		fiber.MethodDelete,
		"/api/projects/"+f.projectID.String()+"/roles/"+role.ID,
		nil,
	)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// TestHandler_AssignRole_Success проверяет HTTP-слой: HTTP-обработчик assign role успех.
func TestHandler_AssignRole_Success(t *testing.T) {
	f := newRBACFixture(t)
	ctx := t.Context()
	role, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name: "Assign", Permissions: []string{"task.create"},
	})
	require.NoError(t, err)

	_, err = f.client.ProjectMember.Create().
		SetUserID(f.strangerID).
		SetProjectID(f.projectID).
		SetProjectRole("viewer").
		Save(ctx)
	require.NoError(t, err)

	app, _ := newRoleHandlerApp(t, f, f.ownerID)
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api/projects/"+f.projectID.String()+"/roles/"+role.ID+"/assign/"+f.strangerID.String(),
		nil,
	)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// TestUpdateCustomRole_Success проверяет сценарий: обновление custom role успех.
func TestUpdateCustomRole_Success(t *testing.T) {
	f := newRBACFixture(t)
	ctx := t.Context()
	role, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name: "Before", Permissions: []string{"project.view"},
	})
	require.NoError(t, err)

	updated, err := f.svc.UpdateCustomRole(ctx, uuid.MustParse(role.ID), rbac.CustomRoleRequest{
		Name: "After", Permissions: []string{"task.create", "file.upload"},
	})
	require.NoError(t, err)
	assert.Equal(t, "After", updated.Name)
	assert.True(t, rbac.Permission(updated.Permissions).Has(rbac.PermTaskCreate))
}

// TestHandler_UnassignRole_Success проверяет HTTP-слой: HTTP-обработчик unassign role успех.
func TestHandler_UnassignRole_Success(t *testing.T) {
	f := newRBACFixture(t)
	ctx := t.Context()
	role, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name: "Unassign", Permissions: []string{"task.create"},
	})
	require.NoError(t, err)

	_, err = f.client.ProjectMember.Create().
		SetUserID(f.strangerID).
		SetProjectID(f.projectID).
		SetProjectRole("viewer").
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, f.svc.AssignCustomRole(ctx, f.projectID, f.strangerID, uuid.MustParse(role.ID)))

	app, _ := newRoleHandlerApp(t, f, f.ownerID)
	req := httptest.NewRequest(
		fiber.MethodDelete,
		"/api/projects/"+f.projectID.String()+"/roles/"+role.ID+"/assign/"+f.strangerID.String(),
		nil,
	)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// TestUpdateCustomRole_InvalidPermission проверяет отклонение при обновление custom role некорректный permission.
func TestUpdateCustomRole_InvalidPermission(t *testing.T) {
	f := newRBACFixture(t)
	ctx := t.Context()
	role, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name: "BadPerm", Permissions: []string{"project.view"},
	})
	require.NoError(t, err)

	_, err = f.svc.UpdateCustomRole(ctx, uuid.MustParse(role.ID), rbac.CustomRoleRequest{
		Name: "BadPerm", Permissions: []string{"not.a.real.permission"},
	})
	require.Error(t, err)
}

// TestClearCustomRole_Success проверяет сценарий: clear custom role успех.
func TestClearCustomRole_Success(t *testing.T) {
	f := newRBACFixture(t)
	ctx := t.Context()
	role, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name: "ClearMe", Permissions: []string{"task.create"},
	})
	require.NoError(t, err)

	_, err = f.client.ProjectMember.Create().
		SetUserID(f.strangerID).
		SetProjectID(f.projectID).
		SetProjectRole("viewer").
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, f.svc.AssignCustomRole(ctx, f.projectID, f.strangerID, uuid.MustParse(role.ID)))

	require.NoError(t, f.svc.ClearCustomRole(ctx, f.projectID, f.strangerID))

	perm, err := f.svc.GetEffectivePermissions(ctx, f.strangerID, f.projectID)
	require.NoError(t, err)
	assert.False(t, perm.Has(rbac.PermTaskCreate))
}
