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

	// P1-BE-33 / W6-A3-001: ошибки list/invalidate нельзя глотать — иначе
	// участники держат elevated mask до TTL после сужения permissions.
	members, err := s.repo.ListRoleMembersWithUser(ctx, roleID)
	if err != nil {
		// Fallback: сбросить весь project-кэш, если не удалось перечислить members.
		if invErr := s.cache.InvalidateProject(ctx, projectID.String()); invErr != nil {
			return nil, fmt.Errorf("list role members after update: %w (invalidate project: %v)", err, invErr)
		}
		return nil, fmt.Errorf("list role members after update: %w", err)
	}
	for _, m := range members {
		if u := m.Edges.User; u != nil {
			if invErr := s.cache.Invalidate(ctx, u.ID.String(), projectID.String()); invErr != nil {
				if projErr := s.cache.InvalidateProject(ctx, projectID.String()); projErr != nil {
					return nil, fmt.Errorf("invalidate member cache: %w (invalidate project: %v)", invErr, projErr)
				}
				return nil, fmt.Errorf("invalidate member cache: %w", invErr)
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

	// A2W6-A3-002: clear members + delete role атомарно; invalidate только после commit.
	// Иначе partial clear + fail delete → dangling edges / inconsistency.
	memberIDs, err := s.repo.ListRoleMemberIDs(ctx, roleID)
	if err != nil {
		return fmt.Errorf("list role members before delete: %w", err)
	}
	if err := s.repo.ClearMembersAndDeleteRole(ctx, roleID, memberIDs); err != nil {
		return err
	}

	if invErr := s.cache.InvalidateProject(ctx, projectID.String()); invErr != nil {
		return fmt.Errorf("invalidate project cache after delete role: %w", invErr)
	}
	return nil
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

// ClearCustomRole снимает кастомную роль с участника, только если :roleId совпадает
// с текущей назначенной ролью (W6-A3-012).
func (s *RBACService) ClearCustomRole(ctx context.Context, projectID, userID, roleID uuid.UUID) error {
	roleProj, err := s.repo.GetRoleProjectID(ctx, roleID)
	if err != nil {
		if IsNotFound(err) {
			return ErrRoleProjectMismatch
		}
		return err
	}
	if roleProj != projectID {
		return ErrRoleProjectMismatch
	}

	mem, err := s.repo.GetProjectMemberWithCustomRole(ctx, userID, projectID)
	if err != nil {
		return err
	}
	custom := mem.Edges.CustomRole
	if custom == nil || custom.ID != roleID {
		return ErrRoleProjectMismatch
	}

	err = s.repo.ClearMemberCustomRole(ctx, mem.ID)
	if err != nil {
		return err
	}

	// A2W6-A3-003: demotion в БД уже выполнен — не отдаём 500 на cache-only fail
	// (иначе клиент думает, что unassign провалился, а роль уже снята).
	// Best-effort: Invalidate → InvalidateProject при ошибке точечного ключа.
	if invErr := s.cache.Invalidate(ctx, userID.String(), projectID.String()); invErr != nil {
		_ = s.cache.InvalidateProject(ctx, projectID.String())
	}
	return nil
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
// не зарегистрированное в реестре. Возвращает ErrUnknownPermission (handler → 400).
func validatePermissionNames(names []string) error {
	for _, n := range names {
		if !ValidPermissionName(n) {
			return fmt.Errorf("%w: %s", ErrUnknownPermission, n)
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
