package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ErrorHandler is the centralized error handler for Fiber.
// It checks error types (AppError, fiber.Error, generic) and returns JSON with code, message, request_id.
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
