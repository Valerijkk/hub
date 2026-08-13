package marketplace_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/marketplace"
)

// TestUpdateListing_OptimisticLock_A6_003 — version:0 / stale → conflict; VersionEQ всегда в WHERE.
// Mutation: убрать marketplacelisting.VersionEQ из UpdateListing → version:0 write проходит → FAIL.
func TestUpdateListing_OptimisticLock_A6_003(t *testing.T) {
	svc, _, memberID, _ := setupMarketplaceService(t)
	ctx := context.Background()

	created := createTestListing(t, svc, memberID)
	require.Equal(t, 1, created.Version)

	title := "Monitor 27 Pro"
	updated, err := svc.UpdateListing(ctx, memberID, created.ID, marketplace.UpdateListingRequest{
		Title: &title, Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, title, updated.Title)
	assert.Equal(t, 2, updated.Version)

	stale := "should-not-apply"
	_, err = svc.UpdateListing(ctx, memberID, created.ID, marketplace.UpdateListingRequest{
		Title: &stale, Version: created.Version,
	})
	require.ErrorIs(t, err, marketplace.ErrVersionConflict)

	_, err = svc.UpdateListing(ctx, memberID, created.ID, marketplace.UpdateListingRequest{
		Title: &stale, Version: 0,
	})
	require.ErrorIs(t, err, marketplace.ErrVersionConflict)

	mine, err := svc.ListMine(ctx, memberID, 1, 25)
	require.NoError(t, err)
	require.Len(t, mine.Items, 1)
	assert.Equal(t, title, mine.Items[0].Title)
	assert.Equal(t, 2, mine.Items[0].Version)
}
