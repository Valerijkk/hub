package auth

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent/user"
)

// RequireAuth возвращает Fiber middleware для валидации JWT access-токенов.
// Сохраняет userID и sessionID в Locals для последующих обработчиков.
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

		// Проверяем, не деактивирован ли пользователь
		uid, parseErr := uuid.Parse(claims.UserID)
		if parseErr == nil {
			u, uErr := svc.db.User.Get(c.Context(), uid)
			if uErr == nil && u.Status == user.StatusDeactivated {
				return fiber.NewError(fiber.StatusForbidden, "user account is deactivated")
			}
		}

		c.Locals("userID", claims.UserID)
		c.Locals("sessionID", claims.SessionID)

		return c.Next()
	}
}
