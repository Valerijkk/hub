//go:build !qa && !release

package main

import (
	"errors"

	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	"github.com/valerii/hub/backend/internal/devdemo"
)

type calendarDemoAdapter struct {
	eventService     *calendarmod.EventService
	eventReadService *calendarmod.EventReadService
}

func newCalendarDemoAdapter(
	eventService *calendarmod.EventService,
	eventReadService *calendarmod.EventReadService,
) (*calendarDemoAdapter, error) {
	if eventService == nil || eventReadService == nil {
		return nil, errors.New("calendar demo adapter dependencies are required")
	}

	return &calendarDemoAdapter{
		eventService:     eventService,
		eventReadService: eventReadService,
	}, nil
}

type calendarDemoStage string

const (
	calendarDemoStageCreate calendarDemoStage = "create"
	calendarDemoStageRead   calendarDemoStage = "read"
	calendarDemoStageRSVP   calendarDemoStage = "rsvp"
	calendarDemoStageCancel calendarDemoStage = "cancel"
)

type calendarDemoReconcileError struct {
	stage calendarDemoStage
	cause error
}

func (err *calendarDemoReconcileError) Error() string {
	return "calendar demo " + string(err.stage) + " reconciliation failed"
}

func (err *calendarDemoReconcileError) Unwrap() []error {
	if err == nil || err.cause == nil {
		return []error{devdemo.ErrCalendarDemoReconcile}
	}

	return []error{devdemo.ErrCalendarDemoReconcile, err.cause}
}

func newCalendarDemoReconcileError(stage calendarDemoStage, cause error) error {
	return &calendarDemoReconcileError{stage: stage, cause: cause}
}

func calendarDemoInputRole(stage calendarDemoStage, role string) (calendarmod.CalendarEventAttendeeRole, error) {
	switch role {
	case "required":
		return calendarmod.CalendarEventAttendeeRoleRequired, nil
	case "optional":
		return calendarmod.CalendarEventAttendeeRoleOptional, nil
	default:
		return "", newCalendarDemoReconcileError(stage, nil)
	}
}

func calendarDemoOutputRole(stage calendarDemoStage, role calendarmod.CalendarEventAttendeeRole) (string, error) {
	switch role {
	case calendarmod.CalendarEventAttendeeRoleOrganizer:
		return "organizer", nil
	case calendarmod.CalendarEventAttendeeRoleRequired:
		return "required", nil
	case calendarmod.CalendarEventAttendeeRoleOptional:
		return "optional", nil
	default:
		return "", newCalendarDemoReconcileError(stage, nil)
	}
}

func calendarDemoInputRSVP(stage calendarDemoStage, status string) (calendarmod.RSVPStatus, error) {
	switch status {
	case "accepted":
		return calendarmod.RSVPAccepted, nil
	case "tentative":
		return calendarmod.RSVPTentative, nil
	default:
		return "", newCalendarDemoReconcileError(stage, nil)
	}
}

func calendarDemoOutputRSVP(stage calendarDemoStage, status calendarmod.RSVPStatus) (string, error) {
	switch status {
	case calendarmod.RSVPNeedsAction:
		return "needs_action", nil
	case calendarmod.RSVPAccepted:
		return "accepted", nil
	case calendarmod.RSVPTentative:
		return "tentative", nil
	case calendarmod.RSVPDeclined:
		return "declined", nil
	default:
		return "", newCalendarDemoReconcileError(stage, nil)
	}
}

func calendarDemoOutputEventStatus(stage calendarDemoStage, status calendarmod.CalendarEventStatus) (string, error) {
	switch status {
	case calendarmod.CalendarEventStatusConfirmed:
		return "confirmed", nil
	case calendarmod.CalendarEventStatusTentative:
		return "tentative", nil
	case calendarmod.CalendarEventStatusCancelled:
		return "cancelled", nil
	default:
		return "", newCalendarDemoReconcileError(stage, nil)
	}
}
