package rbac

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// PermissionCache provides Redis-backed caching for effective permissions.
type PermissionCache struct {
	rdb *redis.Client
}

// NewPermissionCache creates a new permission cache backed by Redis.
func NewPermissionCache(rdb *redis.Client) *PermissionCache {
	return &PermissionCache{rdb: rdb}
}

func cacheKey(userID, projectID string) string {
	return fmt.Sprintf("rbac:%s:%s", userID, projectID)
}

// Get retrieves cached permissions for a user in a project.
// Returns the permission and true if found, or zero and false on cache miss.
func (c *PermissionCache) Get(ctx context.Context, userID, projectID string) (Permission, bool) {
	data, err := c.rdb.Get(ctx, cacheKey(userID, projectID)).Bytes()
	if err != nil {
		return 0, false
	}
	if len(data) != 8 {
		return 0, false
	}
	return Permission(binary.LittleEndian.Uint64(data)), true
}

// Set stores permissions in cache with the given TTL.
func (c *PermissionCache) Set(ctx context.Context, userID, projectID string, perm Permission, ttl time.Duration) error {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(perm))
	return c.rdb.Set(ctx, cacheKey(userID, projectID), data, ttl).Err()
}

// Invalidate removes the cached permissions for a specific user+project.
func (c *PermissionCache) Invalidate(ctx context.Context, userID, projectID string) error {
	return c.rdb.Del(ctx, cacheKey(userID, projectID)).Err()
}

// InvalidateProject removes all cached permissions for a project using SCAN+DEL.
func (c *PermissionCache) InvalidateProject(ctx context.Context, projectID string) error {
	pattern := fmt.Sprintf("rbac:*:%s", projectID)
	iter := c.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := c.rdb.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
