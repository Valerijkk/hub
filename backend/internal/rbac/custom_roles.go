package rbac

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
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

	r, err := s.repo.CreateCustomRole(ctx, projectID, RoleWriteInput{
		Name:            req.Name,
		Description:     req.Description,
		Permissions:     uint64(perm),
		PermissionNames: req.Permissions,
	})
	if err != nil {
		return nil, err
	}

	return toCustomRoleResponse(r), nil
}

// ListCustomRoles возвращает все кастомные (не системные) роли проекта.
func (s *RBACService) ListCustomRoles(ctx context.Context, projectID uuid.UUID) ([]CustomRoleResponse, error) {
	roles, err := s.repo.ListCustomRoles(ctx, projectID)
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
// projectID из URL обязан совпадать с проектом роли; системные роли не обновляются.
func (s *RBACService) UpdateCustomRole(ctx context.Context, projectID, roleID uuid.UUID, req CustomRoleRequest) (*CustomRoleResponse, error) {
	existing, err := s.repo.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if existing.IsSystem {
		return nil, ErrSystemRoleUpdate
	}
	roleProj, err := s.repo.GetRoleProjectID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if roleProj != projectID {
		return nil, ErrRoleProjectMismatch
	}

	if err := validatePermissionNames(req.Permissions); err != nil {
		return nil, err
	}
	perm := permissionsFromNames(req.Permissions)

	r, err := s.repo.UpdateCustomRole(ctx, roleID, RoleWriteInput{
		Name:            req.Name,
		Description:     req.Description,
		Permissions:     uint64(perm),
		PermissionNames: req.Permissions,
	})
	if err != nil {
		return nil, err
	}

	members, _ := s.repo.ListRoleMembersWithUser(ctx, roleID)
	if projectID != uuid.Nil {
		for _, m := range members {
			if u := m.Edges.User; u != nil {
				_ = s.cache.Invalidate(ctx, u.ID.String(), projectID.String())
			}
		}
	}

	return toCustomRoleResponse(r), nil
}

// DeleteCustomRole удаляет кастомную роль. Участники с этой ролью сбрасываются на базовую роль.
func (s *RBACService) DeleteCustomRole(ctx context.Context, projectID, roleID uuid.UUID) error {
	r, err := s.repo.GetRole(ctx, roleID)
	if err != nil {
		return err
	}
	if r.IsSystem {
		return ErrSystemRoleDelete
	}
	roleProj, err := s.repo.GetRoleProjectID(ctx, roleID)
	if err != nil {
		return err
	}
	if roleProj != projectID {
		return ErrRoleProjectMismatch
	}

	memberIDs, _ := s.repo.ListRoleMemberIDs(ctx, roleID)
	for _, id := range memberIDs {
		_ = s.repo.ClearMemberCustomRoleByID(ctx, id)
	}

	_ = s.cache.InvalidateProject(ctx, projectID.String())

	return s.repo.DeleteRole(ctx, roleID)
}

// AssignCustomRole назначает кастомную роль участнику проекта.
func (s *RBACService) AssignCustomRole(ctx context.Context, projectID, userID, roleID uuid.UUID) error {
	roleProj, err := s.repo.GetRoleProjectID(ctx, roleID)
	if err != nil {
		return err
	}
	if roleProj != projectID {
		return ErrRoleProjectMismatch
	}

	mem, err := s.repo.GetProjectMember(ctx, projectID, userID)
	if err != nil {
		return err
	}

	err = s.repo.SetMemberCustomRole(ctx, mem.ID, roleID)
	if err != nil {
		return err
	}

	return s.cache.Invalidate(ctx, userID.String(), projectID.String())
}

// ClearCustomRole снимает кастомную роль с участника проекта (возвращает базовую роль).
func (s *RBACService) ClearCustomRole(ctx context.Context, projectID, userID uuid.UUID) error {
	mem, err := s.repo.GetProjectMember(ctx, projectID, userID)
	if err != nil {
		return err
	}

	err = s.repo.ClearMemberCustomRole(ctx, mem.ID)
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
