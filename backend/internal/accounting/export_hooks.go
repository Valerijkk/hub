//go:build hubtest

package accounting

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

func ExportMapServiceError(err error) error {
	return mapServiceError(err)
}

func ExportParseEntryDate(raw string) (time.Time, error) {
	return parseEntryDate(raw)
}

func ExportFiberStatusFromError(err error) int {
	if fe, ok := err.(*fiber.Error); ok {
		return fe.Code
	}
	return 0
}
