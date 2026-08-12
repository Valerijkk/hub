package middleware

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

// playgroundValidator адаптирует go-playground/validator к Fiber StructValidator.
// Без него теги `validate:"…"` на DTO после Bind().JSON не исполняются (P2-BE-01).
type playgroundValidator struct {
	validate *validator.Validate
}

// Validate проверяет структуру по тегам validate.
func (v *playgroundValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

// NewStructValidator создаёт Fiber StructValidator на базе go-playground/validator/v10.
func NewStructValidator() fiber.StructValidator {
	return &playgroundValidator{validate: validator.New()}
}
