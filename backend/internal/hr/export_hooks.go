//go:build hubtest

package hr

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/valerii/hub/backend/ent"
)

func ExportMapServiceError(err error) error {
	return mapServiceError(err)
}

func ExportDisplayName(u *ent.User) string {
	return displayName(u)
}

func ExportParseDateRange(startRaw, endRaw string) (time.Time, time.Time, error) {
	return parseDateRange(startRaw, endRaw)
}

func ExportIsValidLeaveType(t LeaveType) bool {
	return isValidLeaveType(t)
}

func ExportFiberStatusFromError(err error) int {
	if fe, ok := err.(*fiber.Error); ok {
		return fe.Code
	}
	return 0
}
