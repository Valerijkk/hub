package observability

import (
	"github.com/gofiber/fiber/v3"
)

const unmatchedRouteLabel = "unmatched"

// httpRouteLabel — шаблон маршрута Fiber без сырого path/UUID (W4-A3-001/002).
// Unmatched / 404 в Fiber v3 часто отдаёт Route().Path == "/" при другом c.Path() —
// такие запросы → "unmatched", никогда сырой UUID path.
func httpRouteLabel(c fiber.Ctx) string {
	r := c.Route()
	if r == nil || r.Path == "" {
		return unmatchedRouteLabel
	}
	// Catch-all «нет совпадения»: Fiber оставляет Path="/" при фактическом path ≠ "/".
	if r.Path == "/" && c.Path() != "/" {
		return unmatchedRouteLabel
	}
	return r.Path
}
