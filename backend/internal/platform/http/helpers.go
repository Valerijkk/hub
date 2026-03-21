package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
)

// UserIDFromCtx извлекает ID аутентифицированного пользователя из контекста Fiber.
// Читает c.Locals("userID") и парсит как UUID.
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

// ParseProjectID извлекает и валидирует параметр пути projectId.
func ParseProjectID(c fiber.Ctx) (uuid.UUID, error) {
	return ParseUUIDParam(c, "projectId")
}

// ParseUUIDParam извлекает и валидирует UUID-параметр пути по имени.
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
