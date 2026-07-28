package rbac_test

// TASK-115 / A2W6-A3-001..003: unknown perm→400, delete tx, clear cache hygiene.
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/rbac"
)

// TestHandler_CreateRole_UnknownPermission_400_TASK115 — A2W6-A3-001:
// unknown permission name → 400 (не 500).
// Mutation: убрать errors.Is(ErrUnknownPermission) в CreateRole → 500.
func TestHandler_CreateRole_UnknownPermission_400_TASK115(t *testing.T) {
	f := newRBACFixture(t)
	app, _ := newRoleHandlerApp(t, f, f.ownerID)

	body, _ := json.Marshal(rbac.CustomRoleRequest{
		Name: "Bad", Permissions: []string{"not.a.real.permission"},
	})
	req := httptest.NewRequest(fiber.MethodPost, "/api/projects/"+f.projectID.String()+"/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	assert.Contains(t, string(respBody), "unknown permission")
}

// TestHandler_UpdateRole_UnknownPermission_400_TASK115 — A2W6-A3-001 Update path.
// Mutation: убрать errors.Is(ErrUnknownPermission) в UpdateRole → 500.
func TestHandler_UpdateRole_UnknownPermission_400_TASK115(t *testing.T) {
	f := newRBACFixture(t)
	ctx := t.Context()
	role, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name: "Ok", Permissions: []string{"project.view"},
	})
	require.NoError(t, err)

	app, _ := newRoleHandlerApp(t, f, f.ownerID)
	body, _ := json.Marshal(rbac.CustomRoleRequest{
		Name: "Ok", Permissions: []string{"totally.fake.perm"},
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
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestValidatePermissionNames_TypedErr_TASK115 — typed ErrUnknownPermission для handler mapping.
func TestValidatePermissionNames_TypedErr_TASK115(t *testing.T) {
	err := rbac.ExportValidatePermissionNames([]string{"project.view", "nope.x"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, rbac.ErrUnknownPermission))
}

// TestDeleteCustomRole_AtomicClearAndDelete_TASK115 — A2W6-A3-002:
// после Delete роль отсутствует и custom_role edge снят у участников.
// Mutation: вернуть non-tx clear→delete без ClearMembersAndDeleteRole → при mid-fail
// members cleared, role остаётся (inconsistency).
func TestDeleteCustomRole_AtomicClearAndDelete_TASK115(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	role, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name:        "AtomicDel",
		Permissions: []string{"task.create"},
	})
	require.NoError(t, err)
	roleID := uuid.MustParse(role.ID)

	mem, err := f.client.ProjectMember.Create().
		SetUserID(f.strangerID).
		SetProjectID(f.projectID).
		SetProjectRole("viewer").
		SetCustomRoleID(roleID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, f.svc.DeleteCustomRole(ctx, f.projectID, roleID))

	_, err = f.client.Role.Get(ctx, roleID)
	assert.True(t, ent.IsNotFound(err), "role must be deleted")

	updated, err := f.client.ProjectMember.Query().
		Where(projectmember.IDEQ(mem.ID)).
		WithCustomRole().
		Only(ctx)
	require.NoError(t, err)
	assert.Nil(t, updated.Edges.CustomRole, "custom_role edge must be cleared")
}

// atomicFailStub — ClearMembersAndDeleteRole падает целиком (tx rollback semantics).
type atomicFailStub struct {
	stubRepo
	clearByIDCalled bool
}

func (s *atomicFailStub) ListRoleMemberIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return []uuid.UUID{uuid.MustParse("11111111-1111-1111-1111-111111111111")}, nil
}

func (s *atomicFailStub) ClearMemberCustomRoleByID(context.Context, uuid.UUID) error {
	s.clearByIDCalled = true
	return nil
}

func (s *atomicFailStub) ClearMembersAndDeleteRole(context.Context, uuid.UUID, []uuid.UUID) error {
	return errors.New("simulated tx rollback: delete failed")
}

// TestDeleteCustomRole_UsesAtomicRepoMethod_TASK115 — сервис зовёт ClearMembersAndDeleteRole,
// а не раздельные ClearMemberCustomRoleByID + DeleteRole.
// Mutation: вернуть цикл clear+DeleteRole → clearByIDCalled станет true.
func TestDeleteCustomRole_UsesAtomicRepoMethod_TASK115(t *testing.T) {
	proj := uuid.New()
	roleID := uuid.New()
	repo := &atomicFailStub{
		stubRepo: stubRepo{
			roles: map[uuid.UUID]*ent.Role{
				roleID: {ID: roleID, Name: "X", IsSystem: false},
			},
			proj: map[uuid.UUID]uuid.UUID{roleID: proj},
		},
	}
	mrCache, _ := newCacheTestEnv(t)
	svc := rbac.NewRBACService(repo, mrCache)

	err := svc.DeleteCustomRole(context.Background(), proj, roleID)
	require.Error(t, err)
	assert.False(t, repo.clearByIDCalled, "atomic path must not call ClearMemberCustomRoleByID separately")
	assert.Contains(t, err.Error(), "simulated tx rollback")
}

// TestClearCustomRole_CacheFailNot500_TASK115 — A2W6-A3-003:
// demotion в БД + cache fail → nil (не ошибка), роль снята.
// Mutation: вернуть `return s.cache.Invalidate(...)` → Error при закрытом Redis.
func TestClearCustomRole_CacheFailNot500_TASK115(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()

	role, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name:        "ClearCacheFail",
		Permissions: []string{"task.create"},
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

	perm, err := f.svc.GetEffectivePermissions(ctx, f.strangerID, f.projectID)
	require.NoError(t, err)
	require.True(t, perm.Has(rbac.PermTaskCreate))

	require.NoError(t, f.rdb.Close())

	err = f.svc.ClearCustomRole(ctx, f.projectID, f.strangerID, roleID)
	require.NoError(t, err, "cache-only fail must not surface as error after DB demotion")

	pm, err := f.client.ProjectMember.Query().
		Where(projectmember.HasUserWith(user.IDEQ(f.strangerID))).
		WithCustomRole().
		Only(ctx)
	require.NoError(t, err)
	assert.Nil(t, pm.Edges.CustomRole, "DB demotion must succeed even if Redis is down")
}

// TestClearCustomRole_CacheFailHTTP200_TASK115 — HTTP unassign не 500 при мёртвом Redis.
func TestClearCustomRole_CacheFailHTTP200_TASK115(t *testing.T) {
	f := newRBACFixture(t)
	ctx := context.Background()
	role, err := f.svc.CreateCustomRole(ctx, f.projectID, rbac.CustomRoleRequest{
		Name: "HTTPCacheFail", Permissions: []string{"task.create"},
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

	require.NoError(t, f.rdb.Close())

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
