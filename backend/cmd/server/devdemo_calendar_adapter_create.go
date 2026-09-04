//go:build !qa && !release

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
	requestedOrder := make([]uuid.UUID, 0, len(input.Attendees))
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
		requestedOrder = append(requestedOrder, userID)
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
		var conflict *calendarmod.CalendarEventCreateConflictError
		if errors.As(err, &conflict) && conflict != nil && conflict.Kind() == calendarmod.CalendarEventCreateConflictIdempotencyKeyReused {
			recovered, recoverErr := a.recoverExpiredCalendarEventCreate(
				ctx,
				actorID,
				input.IdempotencyKey,
				command,
				requested,
				requestedOrder,
			)
			if recoverErr == nil {
				return recovered, nil
			}
		}
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

		expected, exists := requested[userID]
		if !exists {
			return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("unexpected calendar demo create receipt attendee"))
		}
		if attendee.Email != expected.email || attendee.Role != expected.role || (!created.Replayed() && attendee.RSVPStatus != calendarmod.RSVPNeedsAction) {
			return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("mismatched calendar demo create receipt attendee"))
		}

		attendees = append(attendees, devdemo.CalendarDemoAttendeeState{
			UserID:     userID,
			Email:      attendee.Email,
			Role:       role,
			RSVPStatus: rsvp,
			Version:    attendeeVersion,
		})
	}
	if len(seen) != len(requested) {
		return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("incomplete calendar demo create receipt"))
	}
	if created.Replayed() {
		live, err := a.GetEvent(ctx, actorID, receipt.ID)
		if err != nil {
			return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, err)
		}
		if live.EventID != receipt.ID || live.Status != "confirmed" {
			return zero, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("invalid calendar demo replay event"))
		}
		liveAttendees, err := calendarDemoReplayCreateAttendees(actorID, requested, requestedOrder, live.Attendees)
		if err != nil {
			return zero, err
		}
		return devdemo.CalendarDemoCreateResult{
			EventID:   live.EventID,
			Version:   live.Version,
			Sequence:  live.Sequence,
			Attendees: liveAttendees,
			Replayed:  true,
		}, nil
	}

	return devdemo.CalendarDemoCreateResult{
		EventID:   receipt.ID,
		Version:   version,
		Sequence:  sequence,
		Attendees: attendees,
		Replayed:  created.Replayed(),
	}, nil
}

func calendarDemoReplayCreateAttendees(
	actorID uuid.UUID,
	requested map[uuid.UUID]calendarDemoRequestedAttendee,
	requestedOrder []uuid.UUID,
	liveAttendees []devdemo.CalendarDemoAttendeeState,
) ([]devdemo.CalendarDemoAttendeeState, error) {
	if actorID == uuid.Nil || len(requested) != len(requestedOrder) {
		return nil, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("invalid calendar demo replay attendees"))
	}

	found := make(map[uuid.UUID]devdemo.CalendarDemoAttendeeState, len(requested))
	for _, attendee := range liveAttendees {
		if attendee.UserID == actorID {
			continue
		}
		if attendee.Role == "organizer" {
			return nil, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("invalid calendar demo replay organizer"))
		}
		expected, wanted := requested[attendee.UserID]
		if !wanted {
			return nil, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("unexpected calendar demo replay attendee"))
		}
		expectedRole, err := calendarDemoOutputRole(calendarDemoStageCreate, expected.role)
		if err != nil {
			return nil, err
		}
		if attendee.Email != expected.email || attendee.Role != expectedRole {
			return nil, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("mismatched calendar demo replay attendee"))
		}
		if _, exists := found[attendee.UserID]; exists {
			return nil, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("duplicate calendar demo replay attendee"))
		}
		found[attendee.UserID] = attendee
	}
	if len(found) != len(requested) {
		return nil, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("incomplete calendar demo replay attendees"))
	}

	attendees := make([]devdemo.CalendarDemoAttendeeState, 0, len(requestedOrder))
	for _, userID := range requestedOrder {
		attendee, exists := found[userID]
		if !exists {
			return nil, newCalendarDemoReconcileError(calendarDemoStageCreate, errors.New("incomplete calendar demo replay attendees"))
		}
		attendees = append(attendees, attendee)
	}
	return attendees, nil
}

func calendarDemoUintToInt(stage calendarDemoStage, value uint) (int, error) {
	converted := int(value)
	if converted < 0 || uint(converted) != value {
		return 0, newCalendarDemoReconcileError(stage, errors.New("calendar demo numeric value exceeds platform range"))
	}
	return converted, nil
}
