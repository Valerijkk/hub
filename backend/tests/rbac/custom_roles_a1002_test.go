package rbac_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/rbac"
)

// stubRepo — минимальный Repository для юнит-тестов A1-002/A1-003 без Postgres.
type stubRepo struct {
	roles map[uuid.UUID]*ent.Role
	proj  map[uuid.UUID]uuid.UUID
	// updateCalled — true, если UpdateCustomRole на репозитории был вызван (до фикса IDOR).
	updateCalled bool
	deleteCalled bool
	assignCalled bool
}

func (s *stubRepo) GetUser(context.Context, uuid.UUID) (*ent.User, error) { return nil, errors.New("unused") }
func (s *stubRepo) GetProjectMemberWithCustomRole(context.Context, uuid.UUID, uuid.UUID) (*ent.ProjectMember, error) {
	return nil, errors.New("unused")
}
func (s *stubRepo) ResolveFreshPromotionTargetAuthority(context.Context, uuid.UUID, uuid.UUID) (rbac.PromotionTargetAuthorityFacts, error) {
	return rbac.PromotionTargetAuthorityFacts{}, errors.New("unused")
}
func (s *stubRepo) CreateCustomRole(context.Context, uuid.UUID, rbac.RoleWriteInput) (*ent.Role, error) {
	return nil, errors.New("unused")
}
func (s *stubRepo) ListCustomRoles(context.Context, uuid.UUID) ([]*ent.Role, error) {
	return nil, errors.New("unused")
}
func (s *stubRepo) UpdateCustomRole(_ context.Context, roleID uuid.UUID, input rbac.RoleWriteInput) (*ent.Role, error) {
	s.updateCalled = true
	r := s.roles[roleID]
	r.Name = input.Name
	r.Permissions = input.Permissions
	r.PermissionNames = input.PermissionNames
	return r, nil
}
func (s *stubRepo) GetRole(_ context.Context, roleID uuid.UUID) (*ent.Role, error) {
	r, ok := s.roles[roleID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return r, nil
}
func (s *stubRepo) ListRoleMembersWithUser(context.Context, uuid.UUID) ([]*ent.ProjectMember, error) {
	return nil, nil
}
func (s *stubRepo) GetRoleProjectID(_ context.Context, roleID uuid.UUID) (uuid.UUID, error) {
	p, ok := s.proj[roleID]
	if !ok {
		return uuid.Nil, &ent.NotFoundError{}
	}
	return p, nil
}
func (s *stubRepo) ListRoleMemberIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (s *stubRepo) ClearMemberCustomRoleByID(context.Context, uuid.UUID) error { return nil }
func (s *stubRepo) ClearMembersAndDeleteRole(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error {
	s.deleteCalled = true
	return nil
}
func (s *stubRepo) DeleteRole(_ context.Context, _ uuid.UUID) error {
	s.deleteCalled = true
	return nil
}
func (s *stubRepo) GetProjectMember(_ context.Context, _, _ uuid.UUID) (*ent.ProjectMember, error) {
	return &ent.ProjectMember{ID: uuid.New()}, nil
}
func (s *stubRepo) SetMemberCustomRole(_ context.Context, _, _ uuid.UUID) error {
	s.assignCalled = true
	return nil
}
func (s *stubRepo) ClearMemberCustomRole(context.Context, uuid.UUID) error { return nil }

func newStubRBAC(t *testing.T, repo *stubRepo) *rbac.RBACService {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return rbac.NewRBACService(repo, rbac.NewPermissionCache(rdb))
}

// TestUpdateCustomRole_RejectsForeignProjectRole — A1-002: роль чужого проекта нельзя обновить через projectId URL.
func TestUpdateCustomRole_RejectsForeignProjectRole(t *testing.T) {
	projA := uuid.New()
	projB := uuid.New()
	roleID := uuid.New()
	repo := &stubRepo{
		roles: map[uuid.UUID]*ent.Role{
			roleID: {ID: roleID, Name: "Foreign", IsSystem: false, Permissions: 1},
		},
		proj: map[uuid.UUID]uuid.UUID{roleID: projB},
	}
	svc := newStubRBAC(t, repo)

	_, err := svc.UpdateCustomRole(context.Background(), projA, roleID, rbac.CustomRoleRequest{
		Name: "Hacked", Permissions: []string{"project.view"},
	})
	require.ErrorIs(t, err, rbac.ErrRoleProjectMismatch)
	assert.False(t, repo.updateCalled, "repo update must not run on project mismatch")
}

// TestAssignCustomRole_RejectsForeignProjectRole — A1-002: нельзя назначить роль другого проекта.
func TestAssignCustomRole_RejectsForeignProjectRole(t *testing.T) {
	projA := uuid.New()
	projB := uuid.New()
	roleID := uuid.New()
	repo := &stubRepo{
		roles: map[uuid.UUID]*ent.Role{
			roleID: {ID: roleID, Name: "Foreign", IsSystem: false},
		},
		proj: map[uuid.UUID]uuid.UUID{roleID: projB},
	}
	svc := newStubRBAC(t, repo)

	err := svc.AssignCustomRole(context.Background(), projA, uuid.New(), roleID)
	require.ErrorIs(t, err, rbac.ErrRoleProjectMismatch)
	assert.False(t, repo.assignCalled)
}

// TestDeleteCustomRole_RejectsForeignProjectRole — A1-003: удаление роли чужого проекта.
func TestDeleteCustomRole_RejectsForeignProjectRole(t *testing.T) {
	projA := uuid.New()
	projB := uuid.New()
	roleID := uuid.New()
	repo := &stubRepo{
		roles: map[uuid.UUID]*ent.Role{
			roleID: {ID: roleID, Name: "Foreign", IsSystem: false},
		},
		proj: map[uuid.UUID]uuid.UUID{roleID: projB},
	}
	svc := newStubRBAC(t, repo)

	err := svc.DeleteCustomRole(context.Background(), projA, roleID)
	require.ErrorIs(t, err, rbac.ErrRoleProjectMismatch)
	assert.False(t, repo.deleteCalled)
}

// TestUpdateCustomRole_RejectsSystemRole — audit gap: Update должен блокировать IsSystem как Delete.
func TestUpdateCustomRole_RejectsSystemRole(t *testing.T) {
	proj := uuid.New()
	roleID := uuid.New()
	repo := &stubRepo{
		roles: map[uuid.UUID]*ent.Role{
			roleID: {ID: roleID, Name: "Owner", IsSystem: true, Permissions: 1},
		},
		proj: map[uuid.UUID]uuid.UUID{roleID: proj},
	}
	svc := newStubRBAC(t, repo)

	_, err := svc.UpdateCustomRole(context.Background(), proj, roleID, rbac.CustomRoleRequest{
		Name: "Rewritten", Permissions: []string{"project.view"},
	})
	require.ErrorIs(t, err, rbac.ErrSystemRoleUpdate)
	assert.False(t, repo.updateCalled)
}
