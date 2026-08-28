package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	"github.com/valerii/hub/backend/internal/devdemo"
)

type calendarDemoCreateReceiptAttendee struct {
	UserID     *uuid.UUID                              `json:"user_id"`
	Email      string                                  `json:"email"`
	Role       calendarmod.CalendarEventAttendeeRole   `json:"role"`
	RSVPStatus calendarmod.RSVPStatus                  `json:"rsvp_status"`
	Version    uint                                    `json:"version"`
}

type calendarDemoCreateReceipt struct {
	ID          uuid.UUID                               `json:"id"`
	Sequence    uint                                    `json:"sequence"`
	Status      calendarmod.CalendarEventStatus         `json:"status"`
	OrganizerID *uuid.UUID                              `json:"organizer_id"`
	Version     uint                                    `json:"version"`
	Attendees   []calendarDemoCreateReceiptAttendee     `json:"attendees"`
}

type calendarDemoRequestedAttendee struct {
	email string
	role  calendarmod.CalendarEventAttendeeRole
}

func (a *calendarDemoAdapter) CreateEvent(
	ctx context.Context,
	actorID uuid.UUID,
	input devdemo.CalendarDemoCreateInput,
) (devdemo.CalendarDemoCreateResult, error) {
	var zero devdemo.CalendarDemoCreateResult
	if a == nil || a.eventService == nil || a.eventReadService == nil || ctx == nil || actorID == uuid.Nil {
		return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("invalid calendar demo create dependencies"))
	}
	if len(input.IdempotencyKey) < 16 || len(input.IdempotencyKey) > 128 {
		return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("invalid calendar demo create key"))
	}
	if input.StartsAt.IsZero() || input.EndsAt.IsZero() || input.StartsAt.Location() != time.UTC || input.EndsAt.Location() != time.UTC || !input.StartsAt.Before(input.EndsAt) {
		return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("invalid calendar demo create interval"))
	}
	if strings.TrimSpace(input.Timezone) == "" {
		return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("invalid calendar demo create timezone"))
	}

	exdates := make([]string, len(input.Exdates))
	for index, exdate := range input.Exdates {
		if exdate.IsZero() || exdate.Location() != time.UTC {
			return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("invalid calendar demo create exclusion"))
		}
		exdates[index] = exdate.Format(time.RFC3339)
	}

	requested := make(map[uuid.UUID]calendarDemoRequestedAttendee, len(input.Attendees))
	candidates := make([]calendarmod.CalendarEventCreateAttendeeCandidate, 0, len(input.Attendees))
	for _, attendee := range input.Attendees {
		if attendee.UserID == uuid.Nil || strings.TrimSpace(attendee.Email) == "" {
			return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("invalid calendar demo create attendee"))
		}
		if _, exists := requested[attendee.UserID]; exists {
			return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("duplicate calendar demo create attendee"))
		}
		role, err := calendarDemoInputRole(calendarDemoStageCreate, attendee.Role)
		if err != nil {
			return zero, err
		}
		userID := attendee.UserID
		requested[userID] = calendarDemoRequestedAttendee{email: attendee.Email, role: role}
		candidates = append(candidates, calendarmod.CalendarEventCreateAttendeeCandidate{
			UserID:      &userID,
			Email:       attendee.Email,
			DisplayName: nil,
			Role:        role,
		})
	}

	command, err := calendarmod.NewCalendarEventCreateCommand(calendarmod.CalendarEventCreateInput{
		IdempotencyKey: input.IdempotencyKey,
		Title:          input.Title,
		Description:    input.Description,
		Location:       input.Location,
		Temporal: calendarmod.CalendarEvent{
			AllDay:    false,
			StartsAt:  input.StartsAt,
			EndsAt:    input.EndsAt,
			Timezone:  input.Timezone,
		},
		RRULE:           input.RRULE,
		Exdates:         exdates,
		ScheduledCallID: nil,
		ReminderMinutes: nil,
		Attendees:       candidates,
	})
	if err != nil {
		return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, err)
	}

	created, err := a.eventService.CreateCalendarEvent(ctx, actorID, command)
	if err != nil {
		return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, err)
	}
	response := created.Response()
	if response.Status() != http.StatusCreated {
		return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("unexpected calendar demo create status"))
	}

	var receipt calendarDemoCreateReceipt
	if err := json.Unmarshal(response.JSON(), &receipt); err != nil {
		return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, err)
	}
	if receipt.ID == uuid.Nil || receipt.Version == 0 || receipt.OrganizerID == nil || *receipt.OrganizerID != actorID {
		return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("invalid calendar demo create receipt"))
	}
	if receipt.Status != calendarmod.CalendarEventStatusConfirmed {
		return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("invalid calendar demo create receipt status"))
	}
	version, err := calendarDemoUintToInt(calendarDemoStageCreate, receipt.Version)
	if err != nil {
		return zero, err
	}
	sequence, err := calendarDemoUintToInt(calendarDemoStageCreate, receipt.Sequence)
	if err != nil {
		return zero, err
	}

	seen := make(map[uuid.UUID]struct{}, len(receipt.Attendees))
	attendees := make([]devdemo.CalendarDemoAttendeeState, 0, len(receipt.Attendees))
	organizerSeen := false
	for _, attendee := range receipt.Attendees {
		if attendee.UserID == nil || *attendee.UserID == uuid.Nil || attendee.Version == 0 || strings.TrimSpace(attendee.Email) == "" {
			return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("invalid calendar demo create receipt attendee"))
		}
		userID := *attendee.UserID
		if _, exists := seen[userID]; exists {
			return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("duplicate calendar demo create receipt attendee"))
		}
		seen[userID] = struct{}{}
		role, err := calendarDemoOutputRole(calendarDemoStageCreate, attendee.Role)
		if err != nil {
			return zero, err
		}
		rsvp, err := calendarDemoOutputRSVP(calendarDemoStageCreate, attendee.RSVPStatus)
		if err != nil {
			return zero, err
		}
		attendeeVersion, err := calendarDemoUintToInt(calendarDemoStageCreate, attendee.Version)
		if err != nil {
			return zero, err
		}

		if expected, exists := requested[userID]; exists {
			if attendee.Email != expected.email || attendee.Role != expected.role || attendee.RSVPStatus != calendarmod.RSVPNeedsAction {
				return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("mismatched calendar demo create receipt attendee"))
			}
			delete(requested, userID)
		} else {
			if userID != actorID || attendee.Role != calendarmod.CalendarEventAttendeeRoleOrganizer || attendee.RSVPStatus != calendarmod.RSVPAccepted || organizerSeen {
				return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("invalid calendar demo create receipt organizer"))
			}
			organizerSeen = true
		}

		attendees = append(attendees, devdemo.CalendarDemoAttendeeState{
			UserID:     userID,
			Email:      attendee.Email,
			Role:       role,
			RSVPStatus: rsvp,
			Version:    attendeeVersion,
		})
	}
	if len(requested) != 0 || !organizerSeen {
		return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("incomplete calendar demo create receipt"))
	}

	return devdemo.CalendarDemoCreateResult{
		EventID:   receipt.ID,
		Version:   version,
		Sequence:  sequence,
		Attendees: attendees,
		Replayed:  created.Replayed(),
	}, nil
}

func calendarDemoUintToInt(stage calendarDemoStage, value uint) (int, error) {
	converted := int(value)
	if converted < 0 || uint(converted) != value {
		return 0, newCalendarDemoReconcileError(stage, errors.New("calendar demo numeric value exceeds platform range"))
	}
	return converted, nil
}
