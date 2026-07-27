package rbac

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/auth"
)

// RequireAuth возвращает middleware для валидации JWT-токенов и сохранения userID в Locals.
// Детали валидации JWT (метод подписи, парсинг claims) реализованы в модуле auth.
// Извлекает токен из заголовка Authorization и устанавливает "userID" в Locals.
func RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := utils.CopyString(c.Get("Authorization"))
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}

		// Extract Bearer token — извлекаем Bearer-токен из Authorization.
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid authorization format")
		}

		token := parts[1]
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "empty token")
		}

		// JWT валидация выполняется в auth.RequireAuth middleware.
		// Здесь сохраняем токен для RBAC-проверок.
		c.Locals("token", token)

		// userID в Locals выставляет auth.RequireAuth; здесь только token для цепочки middleware.

		return c.Next()
	}
}

// RequireSystemRole возвращает middleware, проверяющий системную роль пользователя.
func RequireSystemRole(role string, repo Repository) fiber.Handler {
	return func(c fiber.Ctx) error {
		userIDStr, ok := c.Locals("userID").(string)
		if !ok || userIDStr == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
		}

		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid user ID")
		}

		u, err := repo.GetUser(c.Context(), userUUID)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "user not found")
		}

		if string(u.SystemRole) != role {
			return fiber.NewError(fiber.StatusForbidden, "insufficient system role")
		}

		return c.Next()
	}
}

// RequirePermission возвращает middleware, проверяющий наличие у пользователя указанного
// разрешения в проекте, определяемом параметром пути "projectId".
// Использует utils.CopyString для безопасности контекста Fiber (Pitfall 1).
func RequirePermission(perm Permission, svc *RBACService) fiber.Handler {
	return func(c fiber.Ctx) error {
		userIDStr, ok := c.Locals("userID").(string)
		if !ok || userIDStr == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
		}

		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid user ID")
		}

		// Копируем projectId до async — fasthttp переиспользует RequestCtx (Pitfall 1).
		projectIDStr := utils.CopyString(c.Params("projectId"))
		if projectIDStr == "" {
			return fiber.NewError(fiber.StatusBadRequest, "project ID required")
		}

		projectUUID, err := uuid.Parse(projectIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid project ID")
		}

		effective, err := svc.GetEffectivePermissions(c.Context(), userUUID, projectUUID)
		if err != nil {
			if ent.IsNotFound(err) {
				return fiber.NewError(fiber.StatusForbidden, "not a project member")
			}
			return fiber.NewError(fiber.StatusInternalServerError, "failed to check permissions")
		}

		if !effective.Has(perm) {
			return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
		}

		return c.Next()
	}
}

// RequireAdmin — middleware для системных админов: founder || admin (ORG-04 / auth.IsSystemAdmin).
func RequireAdmin(repo Repository) fiber.Handler {
	return func(c fiber.Ctx) error {
		userIDStr, ok := c.Locals("userID").(string)
		if !ok || userIDStr == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
		}

		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid user ID")
		}

		u, err := repo.GetUser(c.Context(), userUUID)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "user not found")
		}

		if !auth.IsSystemAdmin(u.SystemRole) {
			return fiber.NewError(fiber.StatusForbidden, "insufficient system role")
		}

		return c.Next()
	}
}

// UserIDFromCtx извлекает строку userID из Fiber Locals.
func UserIDFromCtx(c fiber.Ctx) string {
	if id, ok := c.Locals("userID").(string); ok {
		return id
	}
	return ""
}
