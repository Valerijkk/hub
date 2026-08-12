package apigateway

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

var (
	ErrNotFound                = errors.New("not found")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrForbidden               = errors.New("forbidden")
	ErrInvalidInput            = errors.New("invalid input")
	ErrVersionConflict         = errors.New("version conflict")
	ErrAPIKeyNotFound          = errors.New("api key not found")
	ErrAPIKeyInvalid           = errors.New("invalid api key")
	ErrAPIKeyExpired           = errors.New("expired api key")
	ErrAPIKeyUserNotFound      = errors.New("api key user not found")
	ErrAPIKeyUserDeactivated   = errors.New("api key user deactivated")
	ErrWebhookEndpointNotFound = errors.New("webhook endpoint not found")
)

// invalidInputError carries a user-facing message while matching ErrInvalidInput.
type invalidInputError struct {
	msg string
}

func (e *invalidInputError) Error() string { return e.msg }

func (e *invalidInputError) Is(target error) bool { return target == ErrInvalidInput }

func newInvalidInputError(msg string) error {
	return &invalidInputError{msg: msg}
}

// MapToFiber converts apigateway domain errors to Fiber HTTP errors (APIGW-05).
// Handlers and APIKeyAuth middleware call this; services stay HTTP-agnostic.
func MapToFiber(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, ErrAPIKeyNotFound):
		return fiber.NewError(fiber.StatusNotFound, "API-ключ не найден")
	case errors.Is(err, ErrAPIKeyInvalid):
		return fiber.NewError(fiber.StatusUnauthorized, "недействительный API-ключ")
	case errors.Is(err, ErrAPIKeyExpired):
		return fiber.NewError(fiber.StatusUnauthorized, "API-ключ истёк")
	case errors.Is(err, ErrAPIKeyUserNotFound):
		return fiber.NewError(fiber.StatusUnauthorized, "пользователь API-ключа не найден")
	case errors.Is(err, ErrAPIKeyUserDeactivated):
		return fiber.NewError(fiber.StatusForbidden, "user account is deactivated")
	case errors.Is(err, ErrWebhookEndpointNotFound):
		return fiber.NewError(fiber.StatusNotFound, "вебхук-эндпоинт не найден")
	case errors.Is(err, ErrVersionConflict):
		return fiber.NewError(fiber.StatusConflict, "version mismatch: вебхук был изменён другим пользователем")
	case errors.Is(err, ErrInvalidInput):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	case errors.Is(err, ErrUnauthorized):
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	case errors.Is(err, ErrForbidden):
		return fiber.NewError(fiber.StatusForbidden, "forbidden")
	case errors.Is(err, ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, "not found")
	default:
		return err
	}
}
