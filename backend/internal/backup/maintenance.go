package backup

import (
	"strings"
	"sync/atomic"

	"github.com/gofiber/fiber/v3"
)

// MaintenanceGate — in-process флаг режима обслуживания на время restore (A7-003).
// Без реального гейта setMaintenanceMode только логировал, а pg_restore/--drop
// шли на живом трафике.
type MaintenanceGate struct {
	on atomic.Bool
}

// NewMaintenanceGate создаёт выключенный гейт.
func NewMaintenanceGate() *MaintenanceGate {
	return &MaintenanceGate{}
}

// Set включает/выключает режим обслуживания.
func (g *MaintenanceGate) Set(enabled bool) {
	if g == nil {
		return
	}
	g.on.Store(enabled)
}

// Enabled сообщает, активен ли режим обслуживания.
func (g *MaintenanceGate) Enabled() bool {
	if g == nil {
		return false
	}
	return g.on.Load()
}

// MaintenanceMiddleware отклоняет non-health API с 503, пока gate включён.
func MaintenanceMiddleware(gate *MaintenanceGate) fiber.Handler {
	return func(c fiber.Ctx) error {
		if gate == nil || !gate.Enabled() {
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
