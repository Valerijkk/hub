package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// stripQuery возвращает путь без query-строки.
// Используется в access-логах, чтобы не утекали токены через
// `?token=...` (например, WebSocket-апгрейд: /ws?token=<JWT>).
// См. ME-03.
func stripQuery(p string) string {
	if i := strings.IndexByte(p, '?'); i >= 0 {
		return p[:i]
	}
	return p
}

// Logger логирует каждый запрос: метод, путь, статус, длительность и request_id через slog.
// Путь логируется без query string во избежание утечки токенов в логи (ME-03).
func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		requestID, _ := c.Locals("requestID").(string)

		slog.Info("request",
			"method", c.Method(),
			"path", stripQuery(c.Path()),
			"status", c.Response().StatusCode(),
			"duration", duration.String(),
			"duration_ms", duration.Milliseconds(),
			"request_id", requestID,
		)

		return err
	}
}
