// Package health — liveness/readiness probes (A14-001).
// /live = процесс жив; /ready = зависимости доступны.
package health

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Deps — ping-функции для readiness. Nil-пинг пропускается.
type Deps struct {
	PingPostgres func(ctx context.Context) error
	PingRedis    func(ctx context.Context) error
	PingMongo    func(ctx context.Context) error
	PingNATS     func(ctx context.Context) error
	Timeout      time.Duration
}

// Live — процесс up → всегда 200 (K8s/Docker liveness).
func Live(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// Ready — зависимости; 503 при любой деградации (K8s readiness / Compose healthcheck).
// W4-A4-007: ни одного ping → 503 (fail-closed miswiring), не 200 с пустым services.
// W4-A4-012 / P3-BE-12: timeout от c.Context() (RequestCtx), не Background.
func Ready(deps Deps) fiber.Handler {
	timeout := deps.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return func(c fiber.Ctx) error {
		healthCtx, cancel := context.WithTimeout(probeParentContext(c), timeout)
		defer cancel()

		services := fiber.Map{}
		overall := "ok"
		checked := 0

		ping := func(name string, fn func(context.Context) error) {
			if fn == nil {
				return
			}
			checked++
			if err := fn(healthCtx); err != nil {
				services[name] = "error"
				overall = "degraded"
			} else {
				services[name] = "ok"
			}
		}

		ping("postgres", deps.PingPostgres)
		ping("redis", deps.PingRedis)
		ping("mongo", deps.PingMongo)
		ping("nats", deps.PingNATS)

		if checked == 0 {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":   "degraded",
				"services": services,
				"error":    "no readiness probes configured",
			})
		}

		status := fiber.StatusOK
		if overall != "ok" {
			status = fiber.StatusServiceUnavailable
		}
		return c.Status(status).JSON(fiber.Map{
			"status":   overall,
			"services": services,
		})
	}
}

// probeParentContext — parent для WithTimeout (W4-A4-012). Вынесено для hubtest.
func probeParentContext(c fiber.Ctx) context.Context {
	return c.Context()
}
