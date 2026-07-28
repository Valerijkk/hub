package rbac

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/role"
	"github.com/valerii/hub/backend/ent/user"
)

// RoleWriteInput — payload создания/обновления кастомной роли.
type RoleWriteInput struct {
	Name            string
	Description     string
	Permissions     uint64
	PermissionNames []string
}

// Repository — контракт persistence RBAC (ARCH-01).
type Repository interface {
	GetUser(ctx context.Context, userID uuid.UUID) (*ent.User, error)
	GetProjectMemberWithCustomRole(ctx context.Context, userID, projectID uuid.UUID) (*ent.ProjectMember, error)
	CreateCustomRole(ctx context.Context, projectID uuid.UUID, input RoleWriteInput) (*ent.Role, error)
	ListCustomRoles(ctx context.Context, projectID uuid.UUID) ([]*ent.Role, error)
	UpdateCustomRole(ctx context.Context, roleID uuid.UUID, input RoleWriteInput) (*ent.Role, error)
	GetRole(ctx context.Context, roleID uuid.UUID) (*ent.Role, error)
	ListRoleMembersWithUser(ctx context.Context, roleID uuid.UUID) ([]*ent.ProjectMember, error)
	GetRoleProjectID(ctx context.Context, roleID uuid.UUID) (uuid.UUID, error)
	ListRoleMemberIDs(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error)
	ClearMemberCustomRoleByID(ctx context.Context, memberID uuid.UUID) error
	// ClearMembersAndDeleteRole атомарно снимает custom_role с участников и удаляет роль (A2W6-A3-002).
	ClearMembersAndDeleteRole(ctx context.Context, roleID uuid.UUID, memberIDs []uuid.UUID) error
	DeleteRole(ctx context.Context, roleID uuid.UUID) error
	GetProjectMember(ctx context.Context, projectID, userID uuid.UUID) (*ent.ProjectMember, error)
	SetMemberCustomRole(ctx context.Context, memberID, roleID uuid.UUID) error
	ClearMemberCustomRole(ctx context.Context, memberID uuid.UUID) error
}

// EntRepository — реализация Repository через Ent ORM.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт Ent-репозиторий rbac.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

func (r *EntRepository) GetUser(ctx context.Context, userID uuid.UUID) (*ent.User, error) {
	return r.client.User.Get(ctx, userID)
}

func (r *EntRepository) GetProjectMemberWithCustomRole(ctx context.Context, userID, projectID uuid.UUID) (*ent.ProjectMember, error) {
	return r.client.ProjectMember.Query().
		Where(
			projectmember.HasUserWith(user.IDEQ(userID)),
			projectmember.HasProjectWith(project.IDEQ(projectID)),
		).
		WithCustomRole().
		Only(ctx)
}

func (r *EntRepository) CreateCustomRole(ctx context.Context, projectID uuid.UUID, input RoleWriteInput) (*ent.Role, error) {
	return r.client.Role.Create().
		SetName(input.Name).
		SetDescription(input.Description).
		SetPermissions(input.Permissions).
		SetPermissionNames(input.PermissionNames).
		SetIsSystem(false).
		SetProjectID(projectID).
		Save(ctx)
}

func (r *EntRepository) ListCustomRoles(ctx context.Context, projectID uuid.UUID) ([]*ent.Role, error) {
	return r.client.Role.Query().
		Where(
			role.HasProjectWith(project.IDEQ(projectID)),
			role.IsSystemEQ(false),
		).
		All(ctx)
}

func (r *EntRepository) UpdateCustomRole(ctx context.Context, roleID uuid.UUID, input RoleWriteInput) (*ent.Role, error) {
	return r.client.Role.UpdateOneID(roleID).
		SetName(input.Name).
		SetDescription(input.Description).
		SetPermissions(input.Permissions).
		SetPermissionNames(input.PermissionNames).
		Save(ctx)
}

func (r *EntRepository) GetRole(ctx context.Context, roleID uuid.UUID) (*ent.Role, error) {
	return r.client.Role.Get(ctx, roleID)
}

func (r *EntRepository) ListRoleMembersWithUser(ctx context.Context, roleID uuid.UUID) ([]*ent.ProjectMember, error) {
	roleEnt, err := r.client.Role.Get(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return roleEnt.QueryMembers().WithUser().All(ctx)
}

func (r *EntRepository) GetRoleProjectID(ctx context.Context, roleID uuid.UUID) (uuid.UUID, error) {
	roleEnt, err := r.client.Role.Get(ctx, roleID)
	if err != nil {
		return uuid.Nil, err
	}
	proj, err := roleEnt.QueryProject().Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return proj.ID, nil
}

func (r *EntRepository) ListRoleMemberIDs(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	roleEnt, err := r.client.Role.Get(ctx, roleID)
	if err != nil {
		return nil, err
	}
	members, err := roleEnt.QueryMembers().All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(members))
	for _, m := range members {
		ids = append(ids, m.ID)
	}
	return ids, nil
}

func (r *EntRepository) ClearMemberCustomRoleByID(ctx context.Context, memberID uuid.UUID) error {
	return r.client.ProjectMember.UpdateOneID(memberID).
		ClearCustomRole().
		Exec(ctx)
}

// ClearMembersAndDeleteRole — clear members + delete role в одной Ent-транзакции (A2W6-A3-002).
// Без tx partial clear + fail delete оставлял dangling edges и stale elevated cache.
func (r *EntRepository) ClearMembersAndDeleteRole(ctx context.Context, roleID uuid.UUID, memberIDs []uuid.UUID) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin clear+delete role tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, id := range memberIDs {
		if err := tx.ProjectMember.UpdateOneID(id).ClearCustomRole().Exec(ctx); err != nil {
			return fmt.Errorf("clear member custom role %s: %w", id, err)
		}
	}
	if err := tx.Role.DeleteOneID(roleID).Exec(ctx); err != nil {
		return fmt.Errorf("delete role %s: %w", roleID, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit clear+delete role: %w", err)
	}
	return nil
}

func (r *EntRepository) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	return r.client.Role.DeleteOneID(roleID).Exec(ctx)
}

func (r *EntRepository) GetProjectMember(ctx context.Context, projectID, userID uuid.UUID) (*ent.ProjectMember, error) {
	return r.client.ProjectMember.Query().
		Where(
			projectmember.HasUserWith(user.IDEQ(userID)),
			projectmember.HasProjectWith(project.IDEQ(projectID)),
		).
		Only(ctx)
}

func (r *EntRepository) SetMemberCustomRole(ctx context.Context, memberID, roleID uuid.UUID) error {
	return r.client.ProjectMember.UpdateOneID(memberID).
		SetCustomRoleID(roleID).
		Exec(ctx)
}

func (r *EntRepository) ClearMemberCustomRole(ctx context.Context, memberID uuid.UUID) error {
	return r.ClearMemberCustomRoleByID(ctx, memberID)
}

// IsNotFound сообщает об ошибках ent not-found.
func IsNotFound(err error) bool {
	return ent.IsNotFound(err)
}

// ErrSystemRoleDelete возвращается при попытке удалить системную роль.
var ErrSystemRoleDelete = fmt.Errorf("system roles cannot be deleted")

// ErrSystemRoleUpdate возвращается при попытке изменить системную роль.
var ErrSystemRoleUpdate = fmt.Errorf("system roles cannot be updated")

// ErrRoleProjectMismatch — роль не принадлежит указанному проекту (IDOR guard).
var ErrRoleProjectMismatch = fmt.Errorf("role does not belong to project")

// ErrUnknownPermission — имя разрешения отсутствует в реестре (клиентская валидация → 400).
var ErrUnknownPermission = fmt.Errorf("unknown permission name")
