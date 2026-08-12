package rbac

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// invalidateByPatternScript — атомарный SCAN+DEL в одном EVAL.
// W6-A3-014: без Lua окно SCAN→DEL позволяет concurrent Set оставить stale key до TTL.
// A3W6-3-003: комментарий и скрипт согласованы на DEL (не UNLINK).
const invalidateByPatternScript = `
local cursor = "0"
local n = 0
repeat
  local result = redis.call("SCAN", cursor, "MATCH", ARGV[1], "COUNT", 100)
  cursor = result[1]
  local keys = result[2]
  for i = 1, #keys do
    redis.call("DEL", keys[i])
    n = n + 1
  end
until cursor == "0"
return n
`

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

// InvalidateProject удаляет все кешированные разрешения проекта атомарно (Lua SCAN+DEL).
func (c *PermissionCache) InvalidateProject(ctx context.Context, projectID string) error {
	pattern := fmt.Sprintf("rbac:*:%s", projectID)
	return c.invalidateByPattern(ctx, pattern)
}

// InvalidateUser удаляет все кешированные разрешения пользователя во всех проектах
// (ключ `rbac:<userID>:<projectID>`). Нужно при demote system_role / смене прав,
// когда projectID заранее неизвестен (P1-BE-33 / W6-A3-003).
func (c *PermissionCache) InvalidateUser(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("rbac:%s:*", userID)
	return c.invalidateByPattern(ctx, pattern)
}

func (c *PermissionCache) invalidateByPattern(ctx context.Context, pattern string) error {
	return c.rdb.Eval(ctx, invalidateByPatternScript, nil, pattern).Err()
}
