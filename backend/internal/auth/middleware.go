package auth

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
)

// RequireAuth returns Fiber middleware that validates JWT access tokens and
// stores userID and sessionID in Locals for downstream handlers.
func RequireAuth(svc *AuthService) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := utils.CopyString(c.Get("Authorization"))
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid authorization format")
		}

		tokenStr := parts[1]
		if tokenStr == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "empty token")
		}

		claims, err := svc.ParseAccessToken(tokenStr)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		c.Locals("userID", claims.UserID)
		c.Locals("sessionID", claims.SessionID)

		return c.Next()
	}
}
