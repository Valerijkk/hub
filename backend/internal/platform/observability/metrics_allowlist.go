package observability

import (
	"net"

	"github.com/gofiber/fiber/v3"
)

// MetricsAllowMiddleware ограничивает /metrics CIDR (пусто = все).
func MetricsAllowMiddleware(cidrs []string) fiber.Handler {
	if len(cidrs) == 0 {
		return func(c fiber.Ctx) error { return c.Next() }
	}

	nets := parseCIDRs(cidrs)
	denyAll := len(nets) == 0

	return func(c fiber.Ctx) error {
		if denyAll {
			return fiber.NewError(fiber.StatusForbidden, "metrics endpoint forbidden")
		}

		ip := net.ParseIP(c.IP())
		if ip == nil || !ipAllowed(ip, nets) {
			return fiber.NewError(fiber.StatusForbidden, "metrics endpoint forbidden")
		}
		return c.Next()
	}
}

func parseCIDRs(cidrs []string) []*net.IPNet {
	out := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		out = append(out, network)
	}
	return out
}

func ipAllowed(ip net.IP, nets []*net.IPNet) bool {
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
