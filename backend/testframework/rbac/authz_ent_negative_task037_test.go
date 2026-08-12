//go:build hubtest

package rbac_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/rbac"
)

// TASK-037 / P2-BE-02 / W6-A3-010 — CreateRole через реальный RequirePermission + Ent.
// Mutation: снять RequirePermission с Post Create → viewer/stranger получают 201.

func TestHandler_CreateRole_Forbidden_Viewer_TASK037(t *testing.T) {
	f := newRBACFixture(t)
	app, _ := newRoleHandlerApp(t, f, f.viewerID)

	body, _ := json.Marshal(rbac.CustomRoleRequest{
		Name: "HackerRole", Permissions: []string{"task.create"},
	})
	req := httptest.NewRequest(fiber.MethodPost, "/api/projects/"+f.projectID.String()+"/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	roles, err := f.svc.ListCustomRoles(t.Context(), f.projectID)
	require.NoError(t, err)
	assert.Empty(t, roles)
}

func TestHandler_CreateRole_Forbidden_Stranger_TASK037(t *testing.T) {
	f := newRBACFixture(t)
	app, _ := newRoleHandlerApp(t, f, f.strangerID)

	body, _ := json.Marshal(rbac.CustomRoleRequest{
		Name: "StrangerRole", Permissions: []string{"project.edit"},
	})
	req := httptest.NewRequest(fiber.MethodPost, "/api/projects/"+f.projectID.String()+"/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_DeleteRole_Forbidden_Viewer_TASK037(t *testing.T) {
	f := newRBACFixture(t)
	ctx := t.Context()
	role, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name: "KeepMe", Permissions: []string{"project.view"},
	})
	require.NoError(t, err)

	app, _ := newRoleHandlerApp(t, f, f.viewerID)
	req := httptest.NewRequest(
		fiber.MethodDelete,
		"/api/projects/"+f.projectID.String()+"/roles/"+role.ID,
		nil,
	)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	roles, err := f.svc.ListCustomRoles(ctx, f.projectID)
	require.NoError(t, err)
	require.Len(t, roles, 1)
}
