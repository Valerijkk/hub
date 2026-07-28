package observability

import (
	"net"

	"github.com/gofiber/fiber/v3"
)

// MetricsAllowAllSentinel — явный opt-in на открытый /metrics (dev/tests).
// Пустой список CIDR = fail-closed (W4-A3-004).
const MetricsAllowAllSentinel = "*"

// MetricsAllowAllCIDRs — явный allow-all для wiring/тестов (не пустой слайс).
var MetricsAllowAllCIDRs = []string{MetricsAllowAllSentinel}

func metricsAllowsAll(cidrs []string) bool {
	return len(cidrs) == 1 && cidrs[0] == MetricsAllowAllSentinel
}

// authorizeMetricsIP — единая CIDR-проверка для /metrics (W4-A3-007 / P3-BE-07).
// Раньше дублировалась в MetricsEndpoint и мёртвом MetricsAllowMiddleware.
func authorizeMetricsIP(c fiber.Ctx, cidrs []string) error {
	if metricsAllowsAll(cidrs) {
		return nil
	}
	if len(cidrs) == 0 {
		return fiber.NewError(fiber.StatusForbidden, "metrics endpoint forbidden")
	}
	nets := parseCIDRs(cidrs)
	if len(nets) == 0 {
		return fiber.NewError(fiber.StatusForbidden, "metrics endpoint forbidden")
	}
	ip := net.ParseIP(c.IP())
	if ip == nil || !ipAllowed(ip, nets) {
		return fiber.NewError(fiber.StatusForbidden, "metrics endpoint forbidden")
	}
	return nil
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
