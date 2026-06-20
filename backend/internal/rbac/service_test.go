package rbac

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/user"
)

func TestGetEffectivePermissions_ViewerRole(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	perm, err := f.svc.GetEffectivePermissions(ctx, f.viewerID, f.projectID)
	require.NoError(t, err)
	assert.True(t, perm.Has(PermProjectView))
	assert.False(t, perm.Has(PermProjectEdit))
	assert.False(t, perm.Has(PermTaskCreate))
}

func TestGetEffectivePermissions_OwnerRole(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	perm, err := f.svc.GetEffectivePermissions(ctx, f.ownerID, f.projectID)
	require.NoError(t, err)
	assert.Equal(t, ^Permission(0), perm)
}

func TestGetEffectivePermissions_SystemAdminAllPerms(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	perm, err := f.svc.GetEffectivePermissions(ctx, f.adminID, f.projectID)
	require.NoError(t, err)
	assert.Equal(t, ^Permission(0), perm)
	assert.True(t, perm.Has(PermProjectDelete))
	assert.True(t, perm.Has(PermTeamManage))
}

func TestGetEffectivePermissions_NonMemberNotFound(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	_, err := f.svc.GetEffectivePermissions(ctx, f.strangerID, f.projectID)
	require.Error(t, err)
}

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
	assert.False(t, second.Has(PermTaskCreate))
}

func TestGetEffectivePermissions_CustomRoleOverridesProjectRole(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	customPerm := permissionsFromNames([]string{"task.create", "file.upload"})
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
	assert.True(t, perm.Has(PermTaskCreate))
	assert.True(t, perm.Has(PermFileUpload))
	assert.False(t, perm.Has(PermProjectEdit))
}

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
	assert.True(t, perm.Has(PermTaskCreate))
	assert.True(t, perm.Has(PermFileUpload))
}

func TestGetEffectivePermissions_UnknownUser(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	_, err := f.svc.GetEffectivePermissions(ctx, uuid.New(), f.projectID)
	require.Error(t, err)
}

func TestGetEffectivePermissions_AdminCachedInRedis(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	_, err := f.svc.GetEffectivePermissions(ctx, f.adminID, f.projectID)
	require.NoError(t, err)

	u, err := f.client.User.Get(ctx, f.adminID)
	require.NoError(t, err)
	_, err = f.client.User.UpdateOne(u).SetSystemRole(user.SystemRoleMember).Save(ctx)
	require.NoError(t, err)

	cached, ok := f.svc.cache.Get(ctx, f.adminID.String(), f.projectID.String())
	require.True(t, ok)
	assert.Equal(t, ^Permission(0), cached)
}
