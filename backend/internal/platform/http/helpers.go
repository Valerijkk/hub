package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
)

// UserIDFromCtx extracts the authenticated user ID from the Fiber context.
// It reads c.Locals("userID") and parses it as a UUID.
func UserIDFromCtx(c fiber.Ctx) (uuid.UUID, error) {
	idStr, ok := c.Locals("userID").(string)
	if !ok || idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "invalid user ID")
	}
	return id, nil
}

// ParseProjectID extracts and validates the projectId path parameter.
func ParseProjectID(c fiber.Ctx) (uuid.UUID, error) {
	return ParseUUIDParam(c, "projectId")
}

// ParseUUIDParam extracts and validates a UUID path parameter by name.
func ParseUUIDParam(c fiber.Ctx, name string) (uuid.UUID, error) {
	idStr := utils.CopyString(c.Params(name))
	if idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, name+" is required")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return id, nil
}
