package backup

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

const (
	// maintenanceRedisKey — cluster-wide флаг restore (ADR-BACKUP-MAINTENANCE-GATE-2026 / U04).
	maintenanceRedisKey = "hub:backup:maintenance"
	// maintenanceLeaseTTL — safety lease: если под умер без Set(false), ключ истечёт.
	maintenanceLeaseTTL = 2 * time.Hour
)

// MaintenanceGate — режим обслуживания на время restore (A7-003 / W7-A4-003).
// С Redis — shared между репликами; без Redis — process-local (только тесты).
type MaintenanceGate struct {
	rdb   *redis.Client
	local atomic.Bool // write-through + fallback при сбое Redis / nil rdb
}

// NewMaintenanceGate создаёт выключенный in-process гейт (unit-тесты без Redis).
func NewMaintenanceGate() *MaintenanceGate {
	return &MaintenanceGate{}
}

// NewRedisMaintenanceGate — production гейт с Redis как source of truth (U04).
func NewRedisMaintenanceGate(rdb *redis.Client) *MaintenanceGate {
	return &MaintenanceGate{rdb: rdb}
}

// Set включает/выключает режим обслуживания.
// При rdb != nil пишет в Redis (SET+TTL / DEL); локальный кэш всегда обновляется.
func (g *MaintenanceGate) Set(ctx context.Context, enabled bool) {
	if g == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	g.local.Store(enabled)
	if g.rdb == nil {
		return
	}
	if enabled {
		if err := g.rdb.Set(ctx, maintenanceRedisKey, "1", maintenanceLeaseTTL).Err(); err != nil {
			slog.ErrorContext(ctx, "maintenance gate: redis SET failed", "error", err)
		}
		return
	}
	if err := g.rdb.Del(ctx, maintenanceRedisKey).Err(); err != nil {
		slog.ErrorContext(ctx, "maintenance gate: redis DEL failed", "error", err)
	}
}

// Enabled сообщает, активен ли режим обслуживания.
// С Redis читает ключ; redis.Nil → false; infra-ошибка → локальный кэш.
func (g *MaintenanceGate) Enabled(ctx context.Context) bool {
	if g == nil {
		return false
	}
	if g.rdb == nil {
		return g.local.Load()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	val, err := g.rdb.Get(ctx, maintenanceRedisKey).Result()
	if errors.Is(err, redis.Nil) {
		return false
	}
	if err != nil {
		slog.WarnContext(ctx, "maintenance gate: redis GET failed, using local cache", "error", err)
		return g.local.Load()
	}
	return val == "1"
}

// MaintenanceMiddleware отклоняет non-health API с 503, пока gate включён.
func MaintenanceMiddleware(gate *MaintenanceGate) fiber.Handler {
	return func(c fiber.Ctx) error {
		if gate == nil || !gate.Enabled(c.Context()) {
			return c.Next()
		}
		path := c.Path()
		if isMaintenanceBypassPath(path) {
			return c.Next()
		}
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":      "service unavailable: maintenance mode (backup restore in progress)",
			"request_id": c.Locals("request_id"),
		})
	}
}

func isMaintenanceBypassPath(path string) bool {
	p := strings.TrimSuffix(path, "/")
	switch p {
	case "/api/health", "/api/live", "/api/ready", "/health", "/metrics":
		return true
	default:
		return false
	}
}
