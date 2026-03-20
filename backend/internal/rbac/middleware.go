package rbac

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
)

// RequireAuth returns middleware that validates JWT tokens and stores userID in Locals.
// JWT validation details (signing method, claims parsing) will be implemented in the auth module.
// This middleware extracts the token from the Authorization header and sets "userID" in Locals.
func RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := utils.CopyString(c.Get("Authorization"))
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}

		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid authorization format")
		}

		token := parts[1]
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "empty token")
		}

		// TODO: Validate JWT token and extract claims (implemented in auth module)
		// For now, store the token as a placeholder. The auth module will replace this
		// with actual JWT validation using the configured secret.
		c.Locals("token", token)

		// The actual userID extraction from JWT claims will be done by the auth module.
		// This placeholder ensures the middleware chain works during development.

		return c.Next()
	}
}

// RequireSystemRole returns middleware that checks the user's system role.
func RequireSystemRole(role string, client *ent.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		userIDStr, ok := c.Locals("userID").(string)
		if !ok || userIDStr == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
		}

		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid user ID")
		}

		u, err := client.User.Get(c.Context(), userUUID)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "user not found")
		}

		if string(u.SystemRole) != role {
			return fiber.NewError(fiber.StatusForbidden, "insufficient system role")
		}

		return c.Next()
	}
}

// RequirePermission returns middleware that checks the user has a specific permission
// in the project identified by the "projectId" route parameter.
// Uses utils.CopyString for Fiber ctx safety (Pitfall 1).
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

		// Copy string from Fiber ctx before any async usage (Pitfall 1)
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

// RequireAdmin is a convenience middleware that checks for system admin role.
func RequireAdmin(client *ent.Client) fiber.Handler {
	return RequireSystemRole(string(user.SystemRoleAdmin), client)
}

// UserIDFromCtx extracts the userID string from Fiber Locals.
func UserIDFromCtx(c fiber.Ctx) string {
	if id, ok := c.Locals("userID").(string); ok {
		return id
	}
	return ""
}
