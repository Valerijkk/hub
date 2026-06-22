//go:build hubtest

package marketplace

import (
	"github.com/gofiber/fiber/v3"
	"github.com/valerii/hub/backend/ent"
)

func ExportMapServiceError(err error) error {
	return mapServiceError(err)
}

func ExportDisplayName(u *ent.User) string {
	return displayName(u)
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

func ExportFormatPrice(price *int64, currency string) string {
	return formatPrice(price, currency)
}

func ExportValidateListingInput(title, description string, category ListingCategory) error {
	return validateListingInput(title, description, category)
}

func ExportFiberStatusFromError(err error) int {
	if fe, ok := err.(*fiber.Error); ok {
		return fe.Code
	}
	return 0
}
