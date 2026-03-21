package rbac

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/user"
)

const defaultCacheTTL = 5 * time.Minute

// RBACService предоставляет проверку разрешений с кешированием в Redis.
type RBACService struct {
	client *ent.Client
	cache  *PermissionCache
	ttl    time.Duration
}

// NewRBACService создаёт новый RBAC-сервис.
func NewRBACService(client *ent.Client, cache *PermissionCache) *RBACService {
	return &RBACService{
		client: client,
		cache:  cache,
		ttl:    defaultCacheTTL,
	}
}

// GetEffectivePermissions возвращает эффективные разрешения пользователя в проекте.
// Сначала проверяет кеш Redis, затем обращается к базе данных.
// Системные администраторы получают все разрешения.
func (s *RBACService) GetEffectivePermissions(ctx context.Context, userID, projectID uuid.UUID) (Permission, error) {
	userStr := userID.String()
	projectStr := projectID.String()

	// Check cache first
	if perm, ok := s.cache.Get(ctx, userStr, projectStr); ok {
		return perm, nil
	}

	// Check if user is system admin
	u, err := s.client.User.Get(ctx, userID)
	if err != nil {
		return 0, err
	}
	if u.SystemRole == user.SystemRoleAdmin {
		allPerms := ^Permission(0)
		_ = s.cache.Set(ctx, userStr, projectStr, allPerms, s.ttl)
		return allPerms, nil
	}

	// Look up project membership with optional custom role
	member, err := s.client.ProjectMember.Query().
		Where(
			projectmember.HasUserWith(user.IDEQ(userID)),
			projectmember.HasProjectWith(project.IDEQ(projectID)),
		).
		WithCustomRole().
		Only(ctx)
	if err != nil {
		return 0, err
	}

	// Determine base permissions from project role
	perm := DefaultRolePermissions[string(member.ProjectRole)]

	// If custom role is assigned, use its permissions instead
	if customRole := member.Edges.CustomRole; customRole != nil {
		perm = Permission(customRole.Permissions)
	}

	// Cache the result
	_ = s.cache.Set(ctx, userStr, projectStr, perm, s.ttl)

	return perm, nil
}

// InvalidateCache удаляет кешированные разрешения пользователя в проекте.
func (s *RBACService) InvalidateCache(ctx context.Context, userID, projectID string) error {
	return s.cache.Invalidate(ctx, userID, projectID)
}
