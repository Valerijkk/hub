package rbac_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/rbac"
)

// TestUpdateCustomRole_InvalidatesMemberCache_P1BE33 — сужение permissions
// роли → следующий GetEffectivePermissions сразу без elevated mask (W6-A3-001).
func TestUpdateCustomRole_InvalidatesMemberCache_P1BE33(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	role, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name:        "Elevated",
		Permissions: []string{"task.create", "project.edit"},
	})
	require.NoError(t, err)
	roleID := uuid.MustParse(role.ID)

	_, err = f.client.ProjectMember.Create().
		SetUserID(f.strangerID).
		SetProjectID(f.projectID).
		SetProjectRole("viewer").
		SetCustomRoleID(roleID).
		Save(ctx)
	require.NoError(t, err)

	before, err := f.svc.GetEffectivePermissions(ctx, f.strangerID, f.projectID)
	require.NoError(t, err)
	require.True(t, before.Has(rbac.PermProjectEdit))
	require.True(t, before.Has(rbac.PermTaskCreate))

	// Сужаем permissions — без ручного Invalidate.
	_, err = f.svc.UpdateCustomRole(ctx, f.projectID, roleID, rbac.CustomRoleRequest{
		Name:        "Elevated",
		Permissions: []string{"task.create"},
	})
	require.NoError(t, err)

	after, err := f.svc.GetEffectivePermissions(ctx, f.strangerID, f.projectID)
	require.NoError(t, err)
	assert.True(t, after.Has(rbac.PermTaskCreate))
	assert.False(t, after.Has(rbac.PermProjectEdit), "stale elevated mask after UpdateCustomRole")
}

// TestDeleteCustomRole_InvalidatesCache_P1BE33 — после DeleteCustomRole
// участник теряет custom perms сразу (W6-A3-002).
func TestDeleteCustomRole_InvalidatesCache_P1BE33(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	role, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name:        "TempElevated",
		Permissions: []string{"task.create", "project.edit"},
	})
	require.NoError(t, err)
	roleID := uuid.MustParse(role.ID)

	_, err = f.client.ProjectMember.Create().
		SetUserID(f.strangerID).
		SetProjectID(f.projectID).
		SetProjectRole("viewer").
		SetCustomRoleID(roleID).
		Save(ctx)
	require.NoError(t, err)

	before, err := f.svc.GetEffectivePermissions(ctx, f.strangerID, f.projectID)
	require.NoError(t, err)
	require.True(t, before.Has(rbac.PermProjectEdit))

	require.NoError(t, f.svc.DeleteCustomRole(ctx, f.projectID, roleID))

	after, err := f.svc.GetEffectivePermissions(ctx, f.strangerID, f.projectID)
	require.NoError(t, err)
	assert.False(t, after.Has(rbac.PermProjectEdit))
	assert.False(t, after.Has(rbac.PermTaskCreate))
	assert.True(t, after.Has(rbac.PermProjectView)) // viewer baseline
}

// TestDemoteAdmin_RequirePermission403_P1BE33 — admin → member без membership:
// RequirePermission сразу 403 (не stale full-mask до TTL). W6-A3-003.
func TestDemoteAdmin_RequirePermission403_P1BE33(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	// Прогрев кэша admin bypass (^0).
	perm, err := f.svc.GetEffectivePermissions(ctx, f.adminID, f.projectID)
	require.NoError(t, err)
	require.Equal(t, ^rbac.Permission(0), perm)

	u, err := f.client.User.Get(ctx, f.adminID)
	require.NoError(t, err)
	_, err = f.client.User.UpdateOne(u).SetSystemRole(user.SystemRoleMember).Save(ctx)
	require.NoError(t, err)

	guard := rbac.RequirePermission(rbac.PermProjectView, f.svc)
	app := newAppWithUserID(f.adminID.String(), guard, "/projects/:projectId/x")
	status, _ := doGet(t, app, "/projects/"+f.projectID.String()+"/x")
	assert.Equal(t, http.StatusForbidden, status,
		"demoted admin without membership must get 403 immediately")
}

// TestGetEffectivePermissions_AdminDemoteDropsStaleCache_P1BE33 —
// после demote кэш full-mask не возвращается как hit.
func TestGetEffectivePermissions_AdminDemoteDropsStaleCache_P1BE33(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	_, err := f.svc.GetEffectivePermissions(ctx, f.adminID, f.projectID)
	require.NoError(t, err)

	u, err := f.client.User.Get(ctx, f.adminID)
	require.NoError(t, err)
	_, err = f.client.User.UpdateOne(u).SetSystemRole(user.SystemRoleMember).Save(ctx)
	require.NoError(t, err)

	_, err = f.svc.GetEffectivePermissions(ctx, f.adminID, f.projectID)
	require.Error(t, err, "demoted admin without membership must miss / not-found")

	_, ok := rbac.ExportRBACServiceCache(f.svc).Get(ctx, f.adminID.String(), f.projectID.String())
	assert.False(t, ok, "stale admin full-mask must be invalidated")
}

// TestPermissionCache_InvalidateUser_P1BE33 проверяет SCAN+DEL по user.
func TestPermissionCache_InvalidateUser_P1BE33(t *testing.T) {
	cache, _ := newCacheTestEnv(t)
	ctx := context.Background()
	userID := uuid.New().String()
	p1 := uuid.New().String()
	p2 := uuid.New().String()
	other := uuid.New().String()

	require.NoError(t, cache.Set(ctx, userID, p1, rbac.PermProjectView, time.Minute))
	require.NoError(t, cache.Set(ctx, userID, p2, rbac.PermTaskCreate, time.Minute))
	require.NoError(t, cache.Set(ctx, other, p1, rbac.PermFileUpload, time.Minute))

	require.NoError(t, cache.InvalidateUser(ctx, userID))

	_, ok := cache.Get(ctx, userID, p1)
	assert.False(t, ok)
	_, ok = cache.Get(ctx, userID, p2)
	assert.False(t, ok)
	got, ok := cache.Get(ctx, other, p1)
	require.True(t, ok)
	assert.Equal(t, rbac.PermFileUpload, got)
}
