package project_test

// Общие хелперы RBAC для HTTP-тестов модуля project (TASK-009).
import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/rbac"
)

// newRBACServiceForClient поднимает RBAC поверх уже открытого ent-клиента + miniredis.
func newRBACServiceForClient(t *testing.T, client *ent.Client) *rbac.RBACService {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))
}

// newFounderBypassRBAC создаёт RBAC, где userID — system founder (полный доступ).
// Нужен для mock-сервисных handler-тестов без реального ProjectMember.
func newFounderBypassRBAC(t *testing.T, userID uuid.UUID) *rbac.RBACService {
	t.Helper()
	dsn := "file:prrbac_bypass_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	_, err := client.User.Create().
		SetID(userID).
		SetEmail("founder-bypass-" + uuid.New().String()[:8] + "@test.com").
		SetName("Founder Bypass").
		SetSystemRole(user.SystemRoleFounder).
		Save(context.Background())
	require.NoError(t, err)

	return newRBACServiceForClient(t, client)
}
