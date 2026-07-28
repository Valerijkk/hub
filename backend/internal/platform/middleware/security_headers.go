package middleware

import "github.com/gofiber/fiber/v3"

// SecurityHeaders — базовые HTTP security headers на каждый ответ.
// CSP default-src 'self' — JSON API с того же origin.
//
// A2W4-1-001 / NFR-22: Strict-Transport-Security намеренно НЕ выставляется
// здесь. API слушает HTTP за TLS-terminating Caddy; HSTS на origin —
// только на edge (`deployments/Caddyfile.prod`). Дублировать HSTS в Fiber
// при прямом HTTP expose дало бы браузеру ложный «HTTPS forever» сигнал.
func SecurityHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Content-Security-Policy", "default-src 'self'")
		return c.Next()
	}
}
