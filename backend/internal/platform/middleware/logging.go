package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Logger логирует каждый запрос: метод, путь, статус, длительность и request_id через slog.
func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		requestID, _ := c.Locals("requestID").(string)

		slog.Info("request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration", duration.String(),
			"request_id", requestID,
		)

		return err
	}
}
