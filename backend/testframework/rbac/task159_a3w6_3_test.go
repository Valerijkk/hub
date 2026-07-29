package rbac_test

// TASK-159 / A3W6-3-001..003: Delete cache-fail, RequireSystemRole founder, DEL comment sync.
// Stub/miniredis — без enttest fixture (изоляция от параллельных schema/codegen волн).
import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/rbac"
)

// deleteCacheRepo — stub: Tx-delete снимает роль из map (как успешный ClearMembersAndDeleteRole).
type deleteCacheRepo struct {
	stubRepo
}

func (s *deleteCacheRepo) ClearMembersAndDeleteRole(_ context.Context, roleID uuid.UUID, _ []uuid.UUID) error {
	s.deleteCalled = true
	delete(s.roles, roleID)
	delete(s.proj, roleID)
	return nil
}

// TestDeleteCustomRole_CacheFailNotError_TASK159 — A3W6-3-001:
// после успешного Tx delete InvalidateProject fail → nil, роль отсутствует.
// Mutation: вернуть `return fmt.Errorf("invalidate...: %w", invErr)` → Error при закрытом Redis.
func TestDeleteCustomRole_CacheFailNotError_TASK159(t *testing.T) {
	proj := uuid.New()
	roleID := uuid.New()
	repo := &deleteCacheRepo{
		stubRepo: stubRepo{
			roles: map[uuid.UUID]*ent.Role{
				roleID: {ID: roleID, Name: "DelCacheFail", IsSystem: false},
			},
			proj: map[uuid.UUID]uuid.UUID{roleID: proj},
		},
	}

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svc := rbac.NewRBACService(repo, rbac.NewPermissionCache(rdb))
	require.NoError(t, rdb.Close())

	err := svc.DeleteCustomRole(context.Background(), proj, roleID)
	require.NoError(t, err, "cache-only fail must not surface after successful Tx delete")
	assert.True(t, repo.deleteCalled, "Tx clear+delete must have run")
	_, getErr := repo.GetRole(context.Background(), roleID)
	assert.True(t, ent.IsNotFound(getErr), "role must be gone after delete")
}

// TestDeleteCustomRole_CacheFailHTTP200_TASK159 — A3W6-3-001 HTTP:
// Delete role не 500 при мёртвом Redis.
// Mutation: hard-fail InvalidateProject → HTTP 500.
func TestDeleteCustomRole_CacheFailHTTP200_TASK159(t *testing.T) {
	proj := uuid.New()
	roleID := uuid.New()
	repo := &deleteCacheRepo{
		stubRepo: stubRepo{
			roles: map[uuid.UUID]*ent.Role{
				roleID: {ID: roleID, Name: "HTTPDelCacheFail", IsSystem: false},
			},
			proj: map[uuid.UUID]uuid.UUID{roleID: proj},
		},
	}

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svc := rbac.NewRBACService(repo, rbac.NewPermissionCache(rdb))
	require.NoError(t, rdb.Close())

	h := rbac.NewCustomRoleHandler(svc)
	app := fiber.New()
	app.Delete("/api/projects/:projectId/roles/:roleId", h.DeleteRole)

	req := httptest.NewRequest(
		fiber.MethodDelete,
		"/api/projects/"+proj.String()+"/roles/"+roleID.String(),
		nil,
	)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode, "delete must succeed after Tx even if Redis down")
	assert.True(t, repo.deleteCalled)
	_, getErr := repo.GetRole(context.Background(), roleID)
	assert.True(t, ent.IsNotFound(getErr), "role must be gone in stub after delete")
}

// systemRoleRepo — GetUser для middleware RequireSystemRole / RequireAdmin.
type systemRoleRepo struct {
	stubRepo
	users map[uuid.UUID]*ent.User
}

func (s *systemRoleRepo) GetUser(_ context.Context, id uuid.UUID) (*ent.User, error) {
	u, ok := s.users[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return u, nil
}

// TestRequireSystemRole_AdminAllowsFounder_TASK159 — A3W6-3-002:
// founder проходит RequireSystemRole("admin") (ORG-04 / IsSystemAdmin).
// Mutation: вернуть pure string equality → founder 403.
func TestRequireSystemRole_AdminAllowsFounder_TASK159(t *testing.T) {
	founderID := uuid.New()
	adminID := uuid.New()
	memberID := uuid.New()
	repo := &systemRoleRepo{
		users: map[uuid.UUID]*ent.User{
			founderID: {ID: founderID, SystemRole: user.SystemRoleFounder},
			adminID:   {ID: adminID, SystemRole: user.SystemRoleAdmin},
			memberID:  {ID: memberID, SystemRole: user.SystemRoleMember},
		},
	}
	guard := rbac.RequireSystemRole("admin", repo)

	t.Run("founder allowed", func(t *testing.T) {
		app := newAppWithUserID(founderID.String(), guard, "/admin/sysrole")
		status, body := doGet(t, app, "/admin/sysrole")
		assert.Equal(t, fiber.StatusOK, status)
		assert.Equal(t, "ok", body)
	})

	t.Run("admin allowed", func(t *testing.T) {
		app := newAppWithUserID(adminID.String(), guard, "/admin/sysrole")
		status, body := doGet(t, app, "/admin/sysrole")
		assert.Equal(t, fiber.StatusOK, status)
		assert.Equal(t, "ok", body)
	})

	t.Run("member forbidden", func(t *testing.T) {
		app := newAppWithUserID(memberID.String(), guard, "/admin/sysrole")
		status, body := doGet(t, app, "/admin/sysrole")
		assert.Equal(t, fiber.StatusForbidden, status)
		assert.Contains(t, body, "insufficient system role")
	})
}

// TestInvalidateByPatternScript_DELNotUNLINK_TASK159 — A3W6-3-003:
// Lua-скрипт и комментарии в cache.go согласованы на DEL (не UNLINK).
// Mutation: вернуть UNLINK в comment при DEL в script (или наоборот) → FAIL.
func TestInvalidateByPatternScript_DELNotUNLINK_TASK159(t *testing.T) {
	script := rbac.ExportInvalidateByPatternScript()
	assert.Contains(t, script, `redis.call("DEL"`, "Lua must use DEL")
	assert.NotContains(t, script, "UNLINK", "Lua must not mention UNLINK (comment/script sync)")

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	cachePath := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "internal", "rbac", "cache.go"))
	src, err := os.ReadFile(cachePath)
	require.NoError(t, err, "read cache.go for comment/script consistency")
	text := string(src)
	assert.NotContains(t, text, "SCAN+UNLINK", "godoc/comment must not claim UNLINK when script uses DEL")
	assert.Contains(t, text, "SCAN+DEL", "comment must document DEL to match Lua")
}
