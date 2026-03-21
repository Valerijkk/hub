package rbac

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// PermissionCache предоставляет кеширование эффективных разрешений в Redis.
type PermissionCache struct {
	rdb *redis.Client
}

// NewPermissionCache создаёт новый кеш разрешений на основе Redis.
func NewPermissionCache(rdb *redis.Client) *PermissionCache {
	return &PermissionCache{rdb: rdb}
}

// cacheKey формирует Redis-ключ кеша разрешений для пользователя в проекте.
func cacheKey(userID, projectID string) string {
	return fmt.Sprintf("rbac:%s:%s", userID, projectID)
}

// Get извлекает кешированные разрешения пользователя в проекте.
// Возвращает разрешения и true при попадании в кеш, или ноль и false при промахе.
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

// Set сохраняет разрешения в кеше с указанным TTL.
func (c *PermissionCache) Set(ctx context.Context, userID, projectID string, perm Permission, ttl time.Duration) error {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(perm))
	return c.rdb.Set(ctx, cacheKey(userID, projectID), data, ttl).Err()
}

// Invalidate удаляет кешированные разрешения для конкретного пользователя и проекта.
func (c *PermissionCache) Invalidate(ctx context.Context, userID, projectID string) error {
	return c.rdb.Del(ctx, cacheKey(userID, projectID)).Err()
}

// InvalidateProject удаляет все кешированные разрешения проекта через SCAN+DEL.
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
