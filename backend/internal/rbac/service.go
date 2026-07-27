// Package rbac реализует ролевое управление доступом: системные и кастомные роли, права, кеширование.
package rbac

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/auth"
)

const defaultCacheTTL = 5 * time.Minute

// RBACService предоставляет проверку разрешений с кешированием в Redis.
type RBACService struct {
	repo  Repository
	cache *PermissionCache
	ttl   time.Duration
}

// NewRBACService создаёт новый RBAC-сервис.
func NewRBACService(repo Repository, cache *PermissionCache) *RBACService {
	return &RBACService{
		repo:  repo,
		cache: cache,
		ttl:   defaultCacheTTL,
	}
}

// GetEffectivePermissions возвращает эффективные разрешения пользователя в проекте.
// Сначала проверяет кеш Redis, затем обращается к базе данных.
// Системные администраторы получают все разрешения.
func (s *RBACService) GetEffectivePermissions(ctx context.Context, userID, projectID uuid.UUID) (Permission, error) {
	userStr := userID.String()
	projectStr := projectID.String()

	// Сначала Redis-кэш — снижает нагрузку на БД при частых RBAC-проверках.
	if perm, ok := s.cache.Get(ctx, userStr, projectStr); ok {
		return perm, nil
	}

	// Системный admin/founder обходит project membership — полный набор разрешений (ORG-04).
	u, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	if auth.IsSystemAdmin(u.SystemRole) {
		allPerms := ^Permission(0)
		_ = s.cache.Set(ctx, userStr, projectStr, allPerms, s.ttl)
		return allPerms, nil
	}

	// Членство в проекте + опциональная кастомная роль.
	member, err := s.repo.GetProjectMemberWithCustomRole(ctx, userID, projectID)
	if err != nil {
		return 0, err
	}

	// Базовые права по фиксированной роли проекта (owner/member/viewer).
	perm := DefaultRolePermissions[string(member.ProjectRole)]

	// Кастомная роль перекрывает базовую bitmask проекта.
	if customRole := member.Edges.CustomRole; customRole != nil {
		perm = Permission(customRole.Permissions)
	}

	// Кэшируем эффективную маску на defaultCacheTTL.
	_ = s.cache.Set(ctx, userStr, projectStr, perm, s.ttl)

	return perm, nil
}

// InvalidateCache удаляет кешированные разрешения пользователя в проекте.
func (s *RBACService) InvalidateCache(ctx context.Context, userID, projectID string) error {
	return s.cache.Invalidate(ctx, userID, projectID)
}
