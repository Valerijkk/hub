//go:build !qa && !release

package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/devdemo"
)

func (a *calendarDemoAdapter) GetEvent(ctx context.Context, actorID uuid.UUID, eventID uuid.UUID) (devdemo.CalendarDemoEventState, error) {
	if a == nil || a.eventReadService == nil || ctx == nil || actorID == uuid.Nil || eventID == uuid.Nil {
		return devdemo.CalendarDemoEventState{}, newCalendarDemoReconcileError(calendarDemoStageRead, errors.New("invalid calendar read request"))
	}

	detail, err := a.eventReadService.GetCalendarEvent(ctx, actorID, eventID)
	if err != nil {
		return devdemo.CalendarDemoEventState{}, newCalendarDemoReconcileError(calendarDemoStageRead, err)
	}
	if detail.ID != eventID || detail.Version == 0 {
		return devdemo.CalendarDemoEventState{}, newCalendarDemoReconcileError(calendarDemoStageRead, errors.New("invalid calendar event detail"))
	}

	version, err := calendarDemoUintToInt(calendarDemoStageRead, detail.Version)
	if err != nil {
		return devdemo.CalendarDemoEventState{}, err
	}
	sequence, err := calendarDemoUintToInt(calendarDemoStageRead, detail.Sequence)
	if err != nil {
		return devdemo.CalendarDemoEventState{}, err
	}
	status, err := calendarDemoOutputEventStatus(calendarDemoStageRead, detail.Status)
	if err != nil {
		return devdemo.CalendarDemoEventState{}, err
	}

	rrule := ""
	if detail.RRULE != nil {
		if strings.TrimSpace(*detail.RRULE) == "" {
			return devdemo.CalendarDemoEventState{}, newCalendarDemoReconcileError(calendarDemoStageRead, errors.New("invalid calendar event detail"))
		}
		rrule = *detail.RRULE
	}

	exdates := make([]time.Time, 0, len(detail.Exdates))
	for _, source := range detail.Exdates {
		parsed, err := time.Parse(time.RFC3339, source)
		if err != nil || parsed.IsZero() || parsed.Location() != time.UTC || parsed.Format(time.RFC3339) != source {
			return devdemo.CalendarDemoEventState{}, newCalendarDemoReconcileError(calendarDemoStageRead, errors.New("invalid calendar event detail"))
		}
		exdates = append(exdates, parsed)
	}

	attendees := make([]devdemo.CalendarDemoAttendeeState, 0, len(detail.Attendees))
	seenAttendees := make(map[uuid.UUID]struct{}, len(detail.Attendees))
	for _, attendee := range detail.Attendees {
		if attendee.UserID == nil || *attendee.UserID == uuid.Nil || strings.TrimSpace(attendee.Email) == "" || attendee.Version == 0 {
			return devdemo.CalendarDemoEventState{}, newCalendarDemoReconcileError(calendarDemoStageRead, errors.New("invalid calendar event detail"))
		}
		if _, exists := seenAttendees[*attendee.UserID]; exists {
			return devdemo.CalendarDemoEventState{}, newCalendarDemoReconcileError(calendarDemoStageRead, errors.New("invalid calendar event detail"))
		}

		role, err := calendarDemoOutputRole(calendarDemoStageRead, attendee.Role)
		if err != nil {
			return devdemo.CalendarDemoEventState{}, err
		}
		rsvp, err := calendarDemoOutputRSVP(calendarDemoStageRead, attendee.RSVPStatus)
		if err != nil {
			return devdemo.CalendarDemoEventState{}, err
		}
		attendeeVersion, err := calendarDemoUintToInt(calendarDemoStageRead, attendee.Version)
		if err != nil {
			return devdemo.CalendarDemoEventState{}, err
		}

		seenAttendees[*attendee.UserID] = struct{}{}
		attendees = append(attendees, devdemo.CalendarDemoAttendeeState{
			UserID:     *attendee.UserID,
			Email:      attendee.Email,
			Role:       role,
			RSVPStatus: rsvp,
			Version:    attendeeVersion,
		})
	}

	return devdemo.CalendarDemoEventState{
		EventID:   detail.ID,
		Version:   version,
		Sequence:  sequence,
		Status:    status,
		RRULE:     rrule,
		Exdates:   exdates,
		Attendees: attendees,
	}, nil
}
