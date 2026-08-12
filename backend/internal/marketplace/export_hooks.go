//go:build hubtest

package marketplace

import (
	"github.com/gofiber/fiber/v3"
	"github.com/valerii/hub/backend/ent"
)

func ExportMapServiceError(err error) error {
	return mapServiceError(err)
}

// ExportDisplayName — публичный fallback (без email).
func ExportDisplayName(u *ent.User) string {
	return displayName(u, false)
}

// ExportDisplayNamePrivate — ListMine/admin fallback (email OK).
func ExportDisplayNamePrivate(u *ent.User) string {
	return displayName(u, true)
}

func ExportFormatMinorUnits(cents int64) string {
	return formatMinorUnits(cents)
}

func ExportItoa(n int64) string {
	return itoa(n)
}

func ExportPad2(n int64) string {
	return pad2(n)
}

func ExportIsValidCategory(c ListingCategory) bool {
	return isValidCategory(c)
}

func ExportIsValidCurrency(c string) bool {
	return isValidCurrency(c)
}

func ExportNormalizeCurrency(c string) string {
	return normalizeCurrency(c)
}

func ExportMaxReviewNoteRunes() int {
	return maxReviewNoteRunes
}

func ExportToListingResponse(l *ent.MarketplaceListing) ListingResponse {
	return toListingResponse(l)
}

func ExportToListingResponsePrivate(l *ent.MarketplaceListing) ListingResponse {
	return toListingResponsePrivate(l)
}

func ExportFormatPrice(price *int64, currency string) string {
	return formatPrice(price, currency)
}

func ExportValidateListingInput(title, description string, category ListingCategory, priceCents *int64, currency string) error {
	return validateListingInput(title, description, category, priceCents, currency)
}

func ExportFiberStatusFromError(err error) int {
	if fe, ok := err.(*fiber.Error); ok {
		return fe.Code
	}
	return 0
}
