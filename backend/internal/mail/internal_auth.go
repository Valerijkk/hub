package mail

import (
	"crypto/subtle"
	"strings"

	"github.com/gofiber/fiber/v3"
)

const internalTokenHeader = "X-Hub-Internal-Token"

// InternalTokenAuth — только внутренние эндпоинты hub-mail (SEC-15).
// Принимает X-Hub-Internal-Token или Authorization: Bearer.
func InternalTokenAuth(expected string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if expected == "" {
			return fiber.NewError(fiber.StatusServiceUnavailable, "internal token not configured")
		}
		token := strings.TrimSpace(c.Get(internalTokenHeader))
		if token == "" {
			auth := c.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				token = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			}
		}
		if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid internal token")
		}
		return c.Next()
	}
}
