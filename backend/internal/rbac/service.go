// Package rbac реализует ролевое управление доступом: системные и кастомные роли, права, кеширование.
package rbac

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent/user"
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
		// P1-BE-33 / W6-A3-003: full-mask (^0) мог быть закэширован для admin/founder.
		// При demote SystemRole нельзя отдавать stale bypass до TTL — перепроверяем роль.
		if perm != ^Permission(0) {
			return perm, nil
		}
		u, err := s.repo.GetUser(ctx, userID)
		if err != nil {
			return 0, err
		}
		if auth.IsSystemAdmin(u.SystemRole) {
			return perm, nil
		}
		_ = s.cache.Invalidate(ctx, userStr, projectStr)
		// fall through: пересчитать по membership (owner тоже может дать ^0)
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

	// Кастомная роль: permission_names — source of truth (W6-A3-004 / P2-BE-19).
	if customRole := member.Edges.CustomRole; customRole != nil {
		if len(customRole.PermissionNames) > 0 {
			perm = permissionsFromNames(customRole.PermissionNames)
		} else {
			perm = Permission(customRole.Permissions)
		}
	}

	// Кэшируем эффективную маску на defaultCacheTTL.
	_ = s.cache.Set(ctx, userStr, projectStr, perm, s.ttl)

	return perm, nil
}

// InvalidateCache удаляет кешированные разрешения пользователя в проекте.
func (s *RBACService) InvalidateCache(ctx context.Context, userID, projectID string) error {
	return s.cache.Invalidate(ctx, userID, projectID)
}

// InvalidateUserCache удаляет все кешированные разрешения пользователя (все проекты).
// Вызывать при смене system_role / demote admin→member (P1-BE-33).
func (s *RBACService) InvalidateUserCache(ctx context.Context, userID string) error {
	return s.cache.InvalidateUser(ctx, userID)
}

// AuthorizePromotionTarget возвращает свежую видимость target-проекта и право task.create.
func (s *RBACService) AuthorizePromotionTarget(ctx context.Context, userID, projectID uuid.UUID) (bool, bool, error) {
	if s == nil || s.repo == nil {
		return false, false, fmt.Errorf("promotion target authority: service or repository is not configured")
	}

	facts, err := s.repo.ResolveFreshPromotionTargetAuthority(ctx, userID, projectID)
	if err != nil {
		return false, false, err
	}
	if !facts.TargetFound {
		if facts.Member != nil {
			return false, false, fmt.Errorf("promotion target authority: member without target")
		}
		return false, false, nil
	}

	switch facts.ActorSystemRole {
	case user.SystemRoleFounder, user.SystemRoleAdmin:
		return true, true, nil
	case user.SystemRoleMember:
		if facts.Member == nil {
			return true, false, nil
		}
	default:
		return false, false, fmt.Errorf("promotion target authority: invalid actor system role %q", facts.ActorSystemRole)
	}

	perm := DefaultRolePermissions[string(facts.Member.ProjectRole)]
	if customRole := facts.Member.Edges.CustomRole; customRole != nil {
		if len(customRole.PermissionNames) > 0 {
			perm = permissionsFromNames(customRole.PermissionNames)
		} else {
			perm = Permission(customRole.Permissions)
		}
	}
	return true, perm.Has(PermTaskCreate), nil
}
