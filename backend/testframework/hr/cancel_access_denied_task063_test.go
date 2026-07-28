//go:build hubtest

package hr_test

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"

	"github.com/valerii/hub/backend/internal/hr"
)

// Mutation: map ErrAccessDenied → "admin access required" → Equal краснеет.
func TestMapServiceError_AccessDenied_P3BE08(t *testing.T) {
	err := mapServiceError(hr.ErrAccessDenied).(*fiber.Error)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, "leave access denied", err.Message)

	adminErr := mapServiceError(hr.ErrForbidden).(*fiber.Error)
	assert.Equal(t, "admin access required", adminErr.Message)
}
