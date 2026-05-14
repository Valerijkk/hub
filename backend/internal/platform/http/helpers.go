// Package http содержит общие HTTP-хелперы для Fiber-обработчиков.
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
//
// ME-12: дополнительно отклоняем uuid.Nil (все нули). uuid.Parse сам
// по себе пропускает "00000000-0000-0000-0000-000000000000" — но zero
// UUID не должен попадать в запросы к БД, так как поведение DB-движка
// для нулевого ID может неприятно отличаться от 404 "not found"
// (например, точные совпадения с pre-existing test fixtures, или
// первая запись в незаполненной таблице).
func ParseUUIDParam(c fiber.Ctx, name string) (uuid.UUID, error) {
	idStr := utils.CopyString(c.Params(name))
	if idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, name+" is required")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	if id == uuid.Nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return id, nil
}
