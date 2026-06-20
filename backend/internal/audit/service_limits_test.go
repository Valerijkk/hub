package audit

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/user"
)

func TestRepository_GetSystemSetting(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	ctx := context.Background()

	val, err := repo.GetSystemSetting(ctx, "missing.key")
	require.NoError(t, err)
	assert.Equal(t, "", val)

	_, err = repo.(*entRepository).client.SystemSetting.Create().
		SetKey("limit.test").
		SetValue("42").
		SetDescription("test").
		Save(ctx)
	require.NoError(t, err)

	val, err = repo.GetSystemSetting(ctx, "limit.test")
	require.NoError(t, err)
	assert.Equal(t, "42", val)
}

func TestRepository_UpsertSystemSetting(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	ctx := context.Background()

	require.NoError(t, repo.UpsertSystemSetting(ctx, "limit.test", "10", "desc"))
	val, err := repo.GetSystemSetting(ctx, "limit.test")
	require.NoError(t, err)
	assert.Equal(t, "10", val)

	require.NoError(t, repo.UpsertSystemSetting(ctx, "limit.test", "20", "desc"))
	val, err = repo.GetSystemSetting(ctx, "limit.test")
	require.NoError(t, err)
	assert.Equal(t, "20", val)
}

func TestRepository_IsSystemAdmin(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	ctx := context.Background()
	r := repo.(*entRepository)

	admin, err := r.client.User.Create().
		SetEmail("admin-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)
	member, err := r.client.User.Create().
		SetEmail("member-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	ok, err := repo.IsSystemAdmin(ctx, admin.ID)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = repo.IsSystemAdmin(ctx, member.ID)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestAuditService_GetResourceLimits(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := NewAuditService(repo, nil)
	ctx := context.Background()

	limits, err := svc.GetResourceLimits(ctx)
	require.NoError(t, err)
	require.Len(t, limits, 3)
	assert.Equal(t, "50", limits[0].Value)
}

func TestAuditService_CheckLimit(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := NewAuditService(repo, nil)
	ctx := context.Background()

	require.NoError(t, repo.UpsertSystemSetting(ctx, "limit.user.max_projects", "5", "test"))

	tests := []struct {
		name    string
		current int
		wantOK  bool
	}{
		{"under limit", 3, true},
		{"at limit", 5, false},
		{"over limit", 6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := svc.CheckLimit(ctx, "limit.user.max_projects", tt.current)
			require.NoError(t, err)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestAuditService_UpdateResourceLimit(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := NewAuditService(repo, nil)
	ctx := context.Background()

	require.NoError(t, svc.UpdateResourceLimit(ctx, LimitFileMaxSizeMB, "100"))
	val, err := repo.GetSystemSetting(ctx, LimitFileMaxSizeMB)
	require.NoError(t, err)
	assert.Equal(t, "100", val)
}

func TestAuditService_CheckLimit_NoSetting(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := NewAuditService(repo, nil)
	ctx := context.Background()

	ok, err := svc.CheckLimit(ctx, "limit.missing", 999)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestAuditService_CheckLimit_InvalidValue(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := NewAuditService(repo, nil)
	ctx := context.Background()

	require.NoError(t, repo.UpsertSystemSetting(ctx, "limit.bad", "not-a-number", "test"))
	ok, err := svc.CheckLimit(ctx, "limit.bad", 999)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestAuditService_ListTrash_UnknownType(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := NewAuditService(repo, nil)
	ctx := context.Background()

	resp, err := svc.ListTrash(ctx, "unknown", nil, 1, 50)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
	assert.Empty(t, resp.Items)
}

func TestAuditService_DeactivateUser(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := NewAuditService(repo, nil)
	ctx := context.Background()
	r := repo.(*entRepository)

	u, err := r.client.User.Create().SetEmail("deact-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.DeactivateUser(ctx, u.ID))

	updated, err := r.client.User.Get(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, user.StatusDeactivated, updated.Status)
}

func TestAuditService_ReassignTasks(t *testing.T) {
	ctx := context.Background()
	from := uuid.New()
	to := uuid.New()
	repo := &mockAuditRepo{}
	svc := NewAuditService(repo, nil)

	count, err := svc.ReassignTasks(ctx, from, to)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
