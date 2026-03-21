package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// RequestID генерирует уникальный X-Request-ID для каждого запроса и сохраняет в c.Locals.
// Использует uuid.NewString(), возвращающий copy-safe строку (Pitfall 1: переиспользование Fiber RequestCtx).
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Locals("requestID", requestID)
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}
