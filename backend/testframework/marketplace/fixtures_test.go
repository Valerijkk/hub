package marketplace_test

import (
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/marketplace"
)

type (
	Service             = marketplace.Service
	EntRepository       = marketplace.EntRepository
	Repository          = marketplace.Repository
	ListingResponse     = marketplace.ListingResponse
	ListingListResponse = marketplace.ListingListResponse
	CreateListingRequest = marketplace.CreateListingRequest
	UpdateListingRequest = marketplace.UpdateListingRequest
	ReviewListingRequest = marketplace.ReviewListingRequest
	VersionRequest      = marketplace.VersionRequest
	ListingCategory     = marketplace.ListingCategory
	ListingStatus       = marketplace.ListingStatus
)

const (
	CategoryEquipment = marketplace.CategoryEquipment
	CategoryBooks     = marketplace.CategoryBooks
	CategoryServices  = marketplace.CategoryServices
	CategoryFurniture = marketplace.CategoryFurniture
	CategoryOther     = marketplace.CategoryOther

	StatusPending   = marketplace.StatusPending
	StatusApproved  = marketplace.StatusApproved
	StatusRejected  = marketplace.StatusRejected
	StatusSold      = marketplace.StatusSold
	StatusWithdrawn = marketplace.StatusWithdrawn
)

var (
	NewService       = marketplace.NewService
	NewEntRepository = marketplace.NewEntRepository

	ErrNotFound        = marketplace.ErrNotFound
	ErrForbidden       = marketplace.ErrForbidden
	ErrInvalidInput    = marketplace.ErrInvalidInput
	ErrVersionConflict = marketplace.ErrVersionConflict
	ErrNotOwner        = marketplace.ErrNotOwner
	ErrInvalidStatus   = marketplace.ErrInvalidStatus
)

func mapServiceError(err error) error {
	return marketplace.ExportMapServiceError(err)
}

func displayName(u *ent.User) string {
	return marketplace.ExportDisplayName(u)
}

func formatMinorUnits(cents int64) string {
	return marketplace.ExportFormatMinorUnits(cents)
}

func itoa(n int64) string {
	return marketplace.ExportItoa(n)
}

func pad2(n int64) string {
	return marketplace.ExportPad2(n)
}

func isValidCategory(c ListingCategory) bool {
	return marketplace.ExportIsValidCategory(c)
}

func formatPrice(price *int64, currency string) string {
	return marketplace.ExportFormatPrice(price, currency)
}
