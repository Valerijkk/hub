//go:build hubtest

package calendar

import (
	"time"

	"github.com/teambition/rrule-go"
	"github.com/valerii/hub/backend/ent"
)

// ExportExpandCallEvents оборачивает expandCallEvents для hubtest.
func ExportExpandCallEvents(sc *ent.ScheduledCall, from, to time.Time) []EventResponse {
	return expandCallEvents(sc, from, to)
}

// ExportParseRRule оборачивает parseRRule для hubtest.
func ExportParseRRule(base time.Time, rruleStr string) (*rrule.RRule, error) {
	return parseRRule(base, rruleStr)
}

// ExportLeaveEvent оборачивает leaveEvent для hubtest.
func ExportLeaveEvent(lv *ent.LeaveRequest) EventResponse {
	return leaveEvent(lv)
}

// ExportValidateRange оборачивает validateRange для hubtest.
func ExportValidateRange(from, to time.Time) error {
	return validateRange(from, to)
}
