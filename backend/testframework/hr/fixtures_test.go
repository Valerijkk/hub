package hr_test

import (
	"time"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/hr"
)

type (
	Service              = hr.Service
	EntRepository        = hr.EntRepository
	Repository           = hr.Repository
	LeaveRequestResponse = hr.LeaveRequestResponse
	CreateLeaveRequest   = hr.CreateLeaveRequest
	ReviewLeaveRequest   = hr.ReviewLeaveRequest
	LeaveType            = hr.LeaveType
	LeaveStatus          = hr.LeaveStatus
)

const (
	LeaveVacation = hr.LeaveVacation
	LeaveSick     = hr.LeaveSick
	LeavePersonal = hr.LeavePersonal
	LeaveOther    = hr.LeaveOther

	StatusPending   = hr.StatusPending
	StatusApproved  = hr.StatusApproved
	StatusRejected  = hr.StatusRejected
	StatusCancelled = hr.StatusCancelled
)

var (
	NewService       = hr.NewService
	NewEntRepository = hr.NewEntRepository

	ErrNotFound        = hr.ErrNotFound
	ErrForbidden       = hr.ErrForbidden
	ErrInvalidInput    = hr.ErrInvalidInput
	ErrVersionConflict = hr.ErrVersionConflict
)

func mapServiceError(err error) error {
	return hr.ExportMapServiceError(err)
}

func displayName(u *ent.User) string {
	return hr.ExportDisplayName(u)
}

func parseDateRange(startRaw, endRaw string) (time.Time, time.Time, error) {
	return hr.ExportParseDateRange(startRaw, endRaw)
}

func isValidLeaveType(t LeaveType) bool {
	return hr.ExportIsValidLeaveType(t)
}
