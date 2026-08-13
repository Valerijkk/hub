package marketplace_test
import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/marketplacelisting"
)

func TestService_WithdrawListing_Approved(t *testing.T) {
	svc, adminID, memberID, _ := setupMarketplaceService(t)
	ctx := context.Background()

	created := createTestListing(t, svc, memberID)
	approved, err := svc.ApproveListing(ctx, adminID, created.ID, ReviewListingRequest{Version: created.Version})
	require.NoError(t, err)

	withdrawn, err := svc.WithdrawListing(ctx, memberID, approved.ID, VersionRequest{Version: approved.Version})
	require.NoError(t, err)
	assert.Equal(t, StatusWithdrawn, withdrawn.Status)
}

func TestService_WithdrawListing_InvalidStatus(t *testing.T) {
	svc, adminID, memberID, _ := setupMarketplaceService(t)
	ctx := context.Background()

	created := createTestListing(t, svc, memberID)
	approved, err := svc.ApproveListing(ctx, adminID, created.ID, ReviewListingRequest{Version: created.Version})
	require.NoError(t, err)

	sold, err := svc.MarkSold(ctx, memberID, approved.ID, VersionRequest{Version: approved.Version})
	require.NoError(t, err)

	_, err = svc.WithdrawListing(ctx, memberID, sold.ID, VersionRequest{Version: sold.Version})
	assert.ErrorIs(t, err, ErrInvalidStatus)
}

func TestService_MarkSold_InvalidStatus(t *testing.T) {
	svc, _, memberID, _ := setupMarketplaceService(t)
	ctx := context.Background()

	created := createTestListing(t, svc, memberID)
	_, err := svc.MarkSold(ctx, memberID, created.ID, VersionRequest{Version: created.Version})
	assert.ErrorIs(t, err, ErrInvalidStatus)
}

func TestService_ApproveListing_InvalidStatus(t *testing.T) {
	svc, adminID, memberID, _ := setupMarketplaceService(t)
	ctx := context.Background()

	created := createTestListing(t, svc, memberID)
	rejected, err := svc.RejectListing(ctx, adminID, created.ID, ReviewListingRequest{
		Version:    created.Version,
		ReviewNote: "no",
	})
	require.NoError(t, err)
	assert.Equal(t, marketplacelisting.StatusRejected, marketplacelisting.Status(rejected.Status))

	_, err = svc.ApproveListing(ctx, adminID, created.ID, ReviewListingRequest{Version: rejected.Version})
	assert.ErrorIs(t, err, ErrInvalidStatus)
}

func TestService_OwnerStatusChange_NotOwner(t *testing.T) {
	svc, adminID, memberID, _ := setupMarketplaceService(t)
	ctx := context.Background()

	created := createTestListing(t, svc, memberID)
	approved, err := svc.ApproveListing(ctx, adminID, created.ID, ReviewListingRequest{Version: created.Version})
	require.NoError(t, err)

	_, err = svc.MarkSold(ctx, uuid.New(), approved.ID, VersionRequest{Version: approved.Version})
	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestService_TransitionReview_VersionConflict(t *testing.T) {
	svc, adminID, memberID, _ := setupMarketplaceService(t)
	ctx := context.Background()

	created := createTestListing(t, svc, memberID)
	_, err := svc.ApproveListing(ctx, adminID, created.ID, ReviewListingRequest{Version: created.Version + 10})
	assert.ErrorIs(t, err, ErrVersionConflict)
}
