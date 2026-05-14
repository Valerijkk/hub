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

// CustomRoleRequest содержит данные для создания или обновления кастомной роли.
type CustomRoleRequest struct {
	Name        string   `json:"name" validate:"required,min=1,max=100"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions" validate:"required"`
}

// CustomRoleResponse — DTO для ответа с информацией о кастомной роли.
type CustomRoleResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Permissions uint64   `json:"permissions"`
	PermList    []string `json:"permission_names"`
	IsSystem    bool     `json:"is_system"`
}

// CreateCustomRole создаёт кастомную роль в проекте с заданным набором разрешений.
//
// Имена разрешений проходят строгую валидацию против реестра (ValidPermissionName) —
// при наличии неизвестного имени возвращается ошибка, роль не создаётся. Это защищает
// от опечаток в API-клиентах и от устаревших имён после refactor реестра.
//
// Dual-write: одновременно сохраняются (а) `permissions uint64` — кешированная битовая
// проекция, и (б) `permission_names []string` — полный авторитетный список имён. Имена
// являются источником истины; uint64 — fast-path для middleware.
func (s *RBACService) CreateCustomRole(ctx context.Context, projectID uuid.UUID, req CustomRoleRequest) (*CustomRoleResponse, error) {
	if err := validatePermissionNames(req.Permissions); err != nil {
		return nil, err
	}
	perm := permissionsFromNames(req.Permissions)

	r, err := s.client.Role.Create().
		SetName(req.Name).
		SetDescription(req.Description).
		SetPermissions(uint64(perm)).
		SetPermissionNames(req.Permissions).
		SetIsSystem(false).
		SetProjectID(projectID).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return toCustomRoleResponse(r), nil
}

// ListCustomRoles возвращает все кастомные (не системные) роли проекта.
func (s *RBACService) ListCustomRoles(ctx context.Context, projectID uuid.UUID) ([]CustomRoleResponse, error) {
	roles, err := s.client.Role.Query().
		Where(
			role.HasProjectWith(project.IDEQ(projectID)),
			role.IsSystemEQ(false),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]CustomRoleResponse, 0, len(roles))
	for _, r := range roles {
		result = append(result, *toCustomRoleResponse(r))
	}
	return result, nil
}

// UpdateCustomRole обновляет кастомную роль (имя, описание, набор разрешений).
//
// Та же валидация и dual-write, что и в CreateCustomRole.
func (s *RBACService) UpdateCustomRole(ctx context.Context, roleID uuid.UUID, req CustomRoleRequest) (*CustomRoleResponse, error) {
	if err := validatePermissionNames(req.Permissions); err != nil {
		return nil, err
	}
	perm := permissionsFromNames(req.Permissions)

	r, err := s.client.Role.UpdateOneID(roleID).
		SetName(req.Name).
		SetDescription(req.Description).
		SetPermissions(uint64(perm)).
		SetPermissionNames(req.Permissions).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// Инвалидируем кэш — разрешения участников с этой ролью могли измениться.
	// Используем eager loading WithUser(), чтобы избежать N+1 запросов.
	members, _ := r.QueryMembers().WithUser().All(ctx)
	proj, _ := r.QueryProject().Only(ctx)
	if proj != nil {
		for _, m := range members {
			if u := m.Edges.User; u != nil {
				_ = s.cache.Invalidate(ctx, u.ID.String(), proj.ID.String())
			}
		}
	}

	return toCustomRoleResponse(r), nil
}

// DeleteCustomRole удаляет кастомную роль. Участники с этой ролью сбрасываются на базовую роль.
func (s *RBACService) DeleteCustomRole(ctx context.Context, roleID uuid.UUID) error {
	r, err := s.client.Role.Get(ctx, roleID)
	if err != nil {
		return err
	}
	if r.IsSystem {
		return fmt.Errorf("system roles cannot be deleted")
	}

	// Очищаем custom_role у всех участников с этой ролью
	members, _ := r.QueryMembers().All(ctx)
	for _, m := range members {
		_ = s.client.ProjectMember.UpdateOneID(m.ID).
			ClearCustomRole().
			Exec(ctx)
	}

	// Инвалидируем кэш для проекта
	proj, _ := r.QueryProject().Only(ctx)
	if proj != nil {
		_ = s.cache.InvalidateProject(ctx, proj.ID.String())
	}

	return s.client.Role.DeleteOneID(roleID).Exec(ctx)
}

// AssignCustomRole назначает кастомную роль участнику проекта.
func (s *RBACService) AssignCustomRole(ctx context.Context, projectID, userID, roleID uuid.UUID) error {
	mem, err := s.client.ProjectMember.Query().
		Where(
			projectmember.HasUserWith(user.IDEQ(userID)),
			projectmember.HasProjectWith(project.IDEQ(projectID)),
		).
		Only(ctx)
	if err != nil {
		return err
	}

	err = s.client.ProjectMember.UpdateOneID(mem.ID).
		SetCustomRoleID(roleID).
		Exec(ctx)
	if err != nil {
		return err
	}

	return s.cache.Invalidate(ctx, userID.String(), projectID.String())
}

// ClearCustomRole снимает кастомную роль с участника проекта (возвращает базовую роль).
func (s *RBACService) ClearCustomRole(ctx context.Context, projectID, userID uuid.UUID) error {
	mem, err := s.client.ProjectMember.Query().
		Where(
			projectmember.HasUserWith(user.IDEQ(userID)),
			projectmember.HasProjectWith(project.IDEQ(projectID)),
		).
		Only(ctx)
	if err != nil {
		return err
	}

	err = s.client.ProjectMember.UpdateOneID(mem.ID).
		ClearCustomRole().
		Exec(ctx)
	if err != nil {
		return err
	}

	return s.cache.Invalidate(ctx, userID.String(), projectID.String())
}

// permissionsFromNames конвертирует список имён разрешений в битовую маску.
func permissionsFromNames(names []string) Permission {
	var perm Permission
	for _, name := range names {
		if p, ok := permissionNameMap[name]; ok {
			perm = perm.Add(p)
		}
	}
	return perm
}

// permissionNames возвращает список имён разрешений из битовой маски.
func permissionNames(perm Permission) []string {
	var names []string
	for name, p := range permissionNameMap {
		if perm.Has(p) {
			names = append(names, name)
		}
	}
	return names
}

// validatePermissionNames отвергает запрос, в котором есть имя разрешения,
// не зарегистрированное в реестре. Возвращает ошибку с информативным сообщением
// — handler-слой должен превратить её в 400 Bad Request.
func validatePermissionNames(names []string) error {
	for _, n := range names {
		if !ValidPermissionName(n) {
			return fmt.Errorf("unknown permission name: %s", n)
		}
	}
	return nil
}

// toCustomRoleResponse конвертирует Ent Role в DTO ответа.
//
// Источник истины для списка разрешений — `r.PermissionNames`. Если поле непустое
// (новые роли, созданные после миграции на dual-write), используем его как есть.
// Иначе (legacy-роли, у которых JSON-колонка осталась NULL после ALTER TABLE)
// восстанавливаем список из bitmask — это работает для всех 25 текущих разрешений
// с Bit != 0. Когда появятся разрешения с Bit == 0, такие legacy-роли просто не
// будут содержать их — что корректно, потому что эти разрешения не существовали
// в момент создания роли.
func toCustomRoleResponse(r *ent.Role) *CustomRoleResponse {
	permList := r.PermissionNames
	if len(permList) == 0 {
		permList = permissionNames(Permission(r.Permissions))
	}
	return &CustomRoleResponse{
		ID:          r.ID.String(),
		Name:        r.Name,
		Description: r.Description,
		Permissions: r.Permissions,
		PermList:    permList,
		IsSystem:    r.IsSystem,
	}
}
