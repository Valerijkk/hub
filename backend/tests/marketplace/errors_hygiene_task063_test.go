//go:build hubtest

package marketplace_test

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/marketplace"
)

// Mutation: stranger DeleteListing → ErrForbidden → ErrorIs ErrNotOwner краснеет.
func TestMapServiceError_NotOwner_P3BE08(t *testing.T) {
	err := mapServiceError(marketplace.ErrNotOwner).(*fiber.Error)
	assert.Equal(t, fiber.StatusForbidden, err.Code)
	assert.Equal(t, "not listing owner", err.Message)

	adminErr := mapServiceError(marketplace.ErrForbidden).(*fiber.Error)
	assert.Equal(t, "admin access required", adminErr.Message)
}

// Mutation: убрать isValidCategory в ListApproved → invalid category доходит до Ent / не ErrInvalidInput.
func TestListApproved_InvalidCategory_P3BE09(t *testing.T) {
	svc, _, _, _ := setupMarketplaceServiceWithClient(t)
	_, err := svc.ListApproved(context.Background(), "not-a-category", 1, 10)
	require.ErrorIs(t, err, marketplace.ErrInvalidInput)
}
