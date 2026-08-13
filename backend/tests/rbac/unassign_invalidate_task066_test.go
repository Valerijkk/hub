package rbac_test

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/rbac"
)

// TestUnassignRole_WrongRoleID_404_TASK066 — W6-A3-012: чужой :roleId → 404, роль не снята.
func TestUnassignRole_WrongRoleID_404_TASK066(t *testing.T) {
	f := newRBACFixture(t)
	ctx := t.Context()

	assigned, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name: "Assigned", Permissions: []string{"task.create"},
	})
	require.NoError(t, err)
	other, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name: "Other", Permissions: []string{"file.upload"},
	})
	require.NoError(t, err)

	_, err = f.client.ProjectMember.Create().
		SetUserID(f.strangerID).
		SetProjectID(f.projectID).
		SetProjectRole("viewer").
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, f.svc.AssignCustomRole(ctx, f.projectID, f.strangerID, uuid.MustParse(assigned.ID)))

	app, _ := newRoleHandlerApp(t, f, f.ownerID)
	req := httptest.NewRequest(
		fiber.MethodDelete,
		"/api/projects/"+f.projectID.String()+"/roles/"+other.ID+"/assign/"+f.strangerID.String(),
		nil,
	)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	// Назначенная роль осталась.
	perm, err := f.svc.GetEffectivePermissions(ctx, f.strangerID, f.projectID)
	require.NoError(t, err)
	assert.True(t, perm.Has(rbac.PermTaskCreate))
}

// TestInvalidateProject_LuaClearsKeys_TASK066 — W6-A3-014: InvalidateProject чистит все ключи проекта.
func TestInvalidateProject_LuaClearsKeys_TASK066(t *testing.T) {
	cache, _ := newCacheTestEnv(t)
	ctx := context.Background()
	projectID := uuid.New().String()
	userA := uuid.New().String()
	userB := uuid.New().String()

	require.NoError(t, cache.Set(ctx, userA, projectID, rbac.PermProjectView, time.Minute))
	require.NoError(t, cache.Set(ctx, userB, projectID, rbac.PermTaskCreate, time.Minute))
	require.NoError(t, cache.InvalidateProject(ctx, projectID))

	_, ok := cache.Get(ctx, userA, projectID)
	assert.False(t, ok)
	_, ok = cache.Get(ctx, userB, projectID)
	assert.False(t, ok)
}
