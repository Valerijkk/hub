package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ErrorHandler — централизованный обработчик ошибок Fiber.
// Проверяет тип ошибки (AppError, fiber.Error, generic) и возвращает JSON с кодом, сообщением и request_id.
func ErrorHandler(c fiber.Ctx, err error) error {
	requestID, _ := c.Locals("requestID").(string)

	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		appErr.RequestID = requestID
		return c.Status(appErr.Code).JSON(appErr)
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(apperrors.AppError{
			Code:      fiberErr.Code,
			Message:   fiberErr.Message,
			RequestID: requestID,
		})
	}

	return c.Status(500).JSON(apperrors.AppError{
		Code:      500,
		Message:   "internal server error",
		RequestID: requestID,
	})
}
