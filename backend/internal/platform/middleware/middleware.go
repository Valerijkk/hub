// Package middleware содержит HTTP-middleware: логирование, rate limiting, request ID.
package middleware

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"

	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)
// ErrorHandler — централизованный обработчик ошибок Fiber.
// Проверяет тип ошибки (AppError, fiber.Error, generic) и возвращает JSON
// строго в формате {"error","request_id"} (P2-BE-07).
func ErrorHandler(c fiber.Ctx, err error) error {
	requestID, _ := c.Locals("requestID").(string)

	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		appErr.RequestID = requestID
		if appErr.RetryAfter > 0 {
			c.Set("Retry-After", strconv.Itoa(appErr.RetryAfter))
		}
		return c.Status(appErr.Code).JSON(apperrors.PublicErrorBody{
			Error:     appErr.Message,
			RequestID: requestID,
		})
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(apperrors.PublicErrorBody{
			Error:     fiberErr.Message,
			RequestID: requestID,
		})
	}

	// Неизвестная ошибка — 500, детали не утекают клиенту.
	return c.Status(fiber.StatusInternalServerError).JSON(apperrors.PublicErrorBody{
		Error:     "internal server error",
		RequestID: requestID,
	})
}
