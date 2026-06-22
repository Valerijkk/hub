package rbac_test

// Тесты сервиса RBAC: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/rbac"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/user"
)

// TestGetEffectivePermissions_ViewerRole проверяет сценарий: получение effective права viewer role.
func TestGetEffectivePermissions_ViewerRole(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	perm, err := f.svc.GetEffectivePermissions(ctx, f.viewerID, f.projectID)
	require.NoError(t, err)
	assert.True(t, perm.Has(rbac.PermProjectView))
	assert.False(t, perm.Has(rbac.PermProjectEdit))
	assert.False(t, perm.Has(rbac.PermTaskCreate))
}

// TestGetEffectivePermissions_OwnerRole проверяет сценарий: получение effective права owner role.
func TestGetEffectivePermissions_OwnerRole(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	perm, err := f.svc.GetEffectivePermissions(ctx, f.ownerID, f.projectID)
	require.NoError(t, err)
	assert.Equal(t, ^rbac.Permission(0), perm)
}

// TestGetEffectivePermissions_SystemAdminAllPerms проверяет сценарий: получение effective права system админ all perms.
func TestGetEffectivePermissions_SystemAdminAllPerms(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	perm, err := f.svc.GetEffectivePermissions(ctx, f.adminID, f.projectID)
	require.NoError(t, err)
	assert.Equal(t, ^rbac.Permission(0), perm)
	assert.True(t, perm.Has(rbac.PermProjectDelete))
	assert.True(t, perm.Has(rbac.PermTeamManage))
}

// TestGetEffectivePermissions_NonMemberNotFound проверяет сценарий: получение effective права non участник not found.
func TestGetEffectivePermissions_NonMemberNotFound(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	_, err := f.svc.GetEffectivePermissions(ctx, f.strangerID, f.projectID)
	require.Error(t, err)
}

// TestGetEffectivePermissions_CacheHit проверяет сценарий: получение effective права кэш hit.
func TestGetEffectivePermissions_CacheHit(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	first, err := f.svc.GetEffectivePermissions(ctx, f.viewerID, f.projectID)
	require.NoError(t, err)

	_, err = f.client.ProjectMember.Update().
		Where(
			projectmember.HasUserWith(user.IDEQ(f.viewerID)),
			projectmember.HasProjectWith(project.IDEQ(f.projectID)),
		).
		SetProjectRole("developer").
		Save(ctx)
	require.NoError(t, err)

	second, err := f.svc.GetEffectivePermissions(ctx, f.viewerID, f.projectID)
	require.NoError(t, err)
	assert.Equal(t, first, second)
	assert.False(t, second.Has(rbac.PermTaskCreate))
}

// TestGetEffectivePermissions_CustomRoleOverridesProjectRole проверяет сценарий: получение effective права custom role overrides project role.
func TestGetEffectivePermissions_CustomRoleOverridesProjectRole(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	customPerm := rbac.ExportPermissionsFromNames([]string{"task.create", "file.upload"})
	role, err := f.client.Role.Create().
		SetName("Uploader").
		SetPermissions(uint64(customPerm)).
		SetPermissionNames([]string{"task.create", "file.upload"}).
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

	perm, err := f.svc.GetEffectivePermissions(ctx, f.strangerID, f.projectID)
	require.NoError(t, err)
	assert.True(t, perm.Has(rbac.PermTaskCreate))
	assert.True(t, perm.Has(rbac.PermFileUpload))
	assert.False(t, perm.Has(rbac.PermProjectEdit))
}

// TestInvalidateCache_ForcesReload проверяет сценарий: invalidate кэш forces reload.
func TestInvalidateCache_ForcesReload(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	_, err := f.svc.GetEffectivePermissions(ctx, f.viewerID, f.projectID)
	require.NoError(t, err)

	_, err = f.client.ProjectMember.Update().
		Where(
			projectmember.HasUserWith(user.IDEQ(f.viewerID)),
			projectmember.HasProjectWith(project.IDEQ(f.projectID)),
		).
		SetProjectRole("developer").
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, f.svc.InvalidateCache(ctx, f.viewerID.String(), f.projectID.String()))

	perm, err := f.svc.GetEffectivePermissions(ctx, f.viewerID, f.projectID)
	require.NoError(t, err)
	assert.True(t, perm.Has(rbac.PermTaskCreate))
	assert.True(t, perm.Has(rbac.PermFileUpload))
}

// TestGetEffectivePermissions_UnknownUser проверяет сценарий: получение effective права unknown user.
func TestGetEffectivePermissions_UnknownUser(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	_, err := f.svc.GetEffectivePermissions(ctx, uuid.New(), f.projectID)
	require.Error(t, err)
}

// TestGetEffectivePermissions_AdminCachedInRedis проверяет сценарий: получение effective права админ cached in redis.
func TestGetEffectivePermissions_AdminCachedInRedis(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	_, err := f.svc.GetEffectivePermissions(ctx, f.adminID, f.projectID)
	require.NoError(t, err)

	u, err := f.client.User.Get(ctx, f.adminID)
	require.NoError(t, err)
	_, err = f.client.User.UpdateOne(u).SetSystemRole(user.SystemRoleMember).Save(ctx)
	require.NoError(t, err)

	cached, ok := rbac.ExportRBACServiceCache(f.svc).Get(ctx, f.adminID.String(), f.projectID.String())
	require.True(t, ok)
	assert.Equal(t, ^rbac.Permission(0), cached)
}
