package middleware

import "github.com/gofiber/fiber/v3"

// SecurityHeaders sets baseline HTTP security headers on every response.
// CSP default-src 'self' is appropriate for a JSON API served from the same origin.
func SecurityHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Content-Security-Policy", "default-src 'self'")
		return c.Next()
	}
}
