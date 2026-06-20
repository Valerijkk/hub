package rbac

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/user"
)

func newCacheTestEnv(t *testing.T) (*PermissionCache, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return NewPermissionCache(rdb), mr
}

func TestPermissionCache_GetSetInvalidate(t *testing.T) {
	cache, _ := newCacheTestEnv(t)
	ctx := context.Background()
	userID := uuid.New().String()
	projectID := uuid.New().String()
	perm := PermProjectView | PermTaskCreate

	_, ok := cache.Get(ctx, userID, projectID)
	assert.False(t, ok)

	require.NoError(t, cache.Set(ctx, userID, projectID, perm, time.Minute))

	got, ok := cache.Get(ctx, userID, projectID)
	require.True(t, ok)
	assert.Equal(t, perm, got)

	require.NoError(t, cache.Invalidate(ctx, userID, projectID))
	_, ok = cache.Get(ctx, userID, projectID)
	assert.False(t, ok)
}

func TestPermissionCache_GetInvalidPayload(t *testing.T) {
	cache, mr := newCacheTestEnv(t)
	ctx := context.Background()
	userID := uuid.New().String()
	projectID := uuid.New().String()

	require.NoError(t, mr.Set(cacheKey(userID, projectID), "short"))
	_, ok := cache.Get(ctx, userID, projectID)
	assert.False(t, ok)
}

func TestPermissionCache_InvalidateProject(t *testing.T) {
	cache, _ := newCacheTestEnv(t)
	ctx := context.Background()
	projectID := uuid.New().String()
	userA := uuid.New().String()
	userB := uuid.New().String()
	otherProject := uuid.New().String()

	require.NoError(t, cache.Set(ctx, userA, projectID, PermProjectView, time.Minute))
	require.NoError(t, cache.Set(ctx, userB, projectID, PermTaskCreate, time.Minute))
	require.NoError(t, cache.Set(ctx, userA, otherProject, PermFileUpload, time.Minute))

	require.NoError(t, cache.InvalidateProject(ctx, projectID))

	_, ok := cache.Get(ctx, userA, projectID)
	assert.False(t, ok)
	_, ok = cache.Get(ctx, userB, projectID)
	assert.False(t, ok)

	got, ok := cache.Get(ctx, userA, otherProject)
	require.True(t, ok)
	assert.Equal(t, PermFileUpload, got)
}

func TestCreateCustomRole_ValidatesPermissions(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	_, err := f.svc.CreateCustomRole(ctx, f.projectID, CustomRoleRequest{
		Name: "Bad", Permissions: []string{"not.a.real.perm"},
	})
	require.Error(t, err)
}

func TestCreateCustomRole_Success(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	role, err := f.svc.CreateCustomRole(ctx, f.projectID, CustomRoleRequest{
		Name:        "Uploader",
		Description: "Can upload files",
		Permissions: []string{"file.upload", "task.create"},
	})
	require.NoError(t, err)
	require.NotNil(t, role)
	assert.Equal(t, "Uploader", role.Name)
	assert.Contains(t, role.PermList, "file.upload")
	assert.True(t, Permission(role.Permissions).Has(PermFileUpload))
}

func TestListCustomRoles_ExcludesSystem(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	_, err := f.svc.CreateCustomRole(ctx, f.projectID, CustomRoleRequest{
		Name: "Custom", Permissions: []string{"project.view"},
	})
	require.NoError(t, err)

	roles, err := f.svc.ListCustomRoles(ctx, f.projectID)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	assert.Equal(t, "Custom", roles[0].Name)
}

func TestAssignCustomRole_UpdatesEffectivePermissions(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	role, err := f.svc.CreateCustomRole(ctx, f.projectID, CustomRoleRequest{
		Name: "Tasker", Permissions: []string{"task.create"},
	})
	require.NoError(t, err)

	_, err = f.client.ProjectMember.Create().
		SetUserID(f.strangerID).
		SetProjectID(f.projectID).
		SetProjectRole("viewer").
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, f.svc.AssignCustomRole(ctx, f.projectID, f.strangerID, uuid.MustParse(role.ID)))

	perm, err := f.svc.GetEffectivePermissions(ctx, f.strangerID, f.projectID)
	require.NoError(t, err)
	assert.True(t, perm.Has(PermTaskCreate))
	assert.False(t, perm.Has(PermProjectEdit))
}

func TestDeleteCustomRole_ClearsAssignments(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	role, err := f.svc.CreateCustomRole(ctx, f.projectID, CustomRoleRequest{
		Name: "Temp", Permissions: []string{"task.create"},
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

	require.NoError(t, f.svc.DeleteCustomRole(ctx, roleID))

	mem, err := f.client.ProjectMember.Query().
		Where(projectmember.HasUserWith(user.IDEQ(f.strangerID))).
		WithCustomRole().
		Only(ctx)
	require.NoError(t, err)
	assert.Nil(t, mem.Edges.CustomRole)

	roles, err := f.svc.ListCustomRoles(ctx, f.projectID)
	require.NoError(t, err)
	assert.Empty(t, roles)
}

func TestGetEffectivePermissions_DeveloperRole(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	_, err := f.client.ProjectMember.Update().
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
