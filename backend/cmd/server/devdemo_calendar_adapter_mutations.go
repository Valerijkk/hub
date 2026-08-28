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

var _ devdemo.CalendarDemoPort = (*calendarDemoAdapter)(nil)

func (adapter *calendarDemoAdapter) RespondSeries(
	ctx context.Context,
	actorID uuid.UUID,
	input devdemo.CalendarDemoRSVPInput,
) (devdemo.CalendarDemoRSVPResult, error) {
	const stage = calendarDemoStageRSVP

	if adapter == nil || adapter.eventService == nil || ctx == nil || actorID == uuid.Nil {
		return devdemo.CalendarDemoRSVPResult{}, newCalendarDemoReconcileError(stage, errors.New("invalid calendar demo RSVP context"))
	}
	if len(input.IdempotencyKey) < 16 || len(input.IdempotencyKey) > 128 || input.EventID == uuid.Nil {
		return devdemo.CalendarDemoRSVPResult{}, newCalendarDemoReconcileError(stage, errors.New("invalid calendar demo RSVP input"))
	}

	expectedVersion, err := calendarDemoPositiveIntToUint(stage, input.ExpectedAttendeeVersion)
	if err != nil {
		return devdemo.CalendarDemoRSVPResult{}, err
	}
	status, err := calendarDemoInputRSVP(stage, input.Status)
	if err != nil {
		return devdemo.CalendarDemoRSVPResult{}, err
	}

	command, err := calendarmod.NewCalendarEventSeriesRSVPCommand(calendarmod.CalendarEventSeriesRSVPInput{
		IdempotencyKey:          input.IdempotencyKey,
		EventID:                 input.EventID,
		ExpectedAttendeeVersion: expectedVersion,
		Status:                  status,
	})
	if err != nil {
		return devdemo.CalendarDemoRSVPResult{}, newCalendarDemoReconcileError(stage, err)
	}

	serviceResult, err := adapter.eventService.RespondToCalendarSeries(ctx, actorID, command)
	if err != nil {
		return devdemo.CalendarDemoRSVPResult{}, newCalendarDemoReconcileError(stage, err)
	}

	response := serviceResult.Response()
	if response.Status() != http.StatusOK {
		return devdemo.CalendarDemoRSVPResult{}, newCalendarDemoReconcileError(stage, errors.New("unexpected calendar demo RSVP response"))
	}

	var output struct {
		ID                  uuid.UUID                              `json:"id"`
		UserID              *uuid.UUID                             `json:"user_id"`
		Email               string                                 `json:"email"`
		Role                calendarmod.CalendarEventAttendeeRole `json:"role"`
		RSVPStatus          calendarmod.RSVPStatus                 `json:"rsvp_status"`
		RecurrenceID        *string                                `json:"recurrence_id"`
		RespondedAt         *time.Time                             `json:"responded_at"`
		Version             uint                                   `json:"version"`
		ReplyDeliveryStatus string                                 `json:"reply_delivery_status"`
	}
	if err := json.Unmarshal(response.JSON(), &output); err != nil {
		return devdemo.CalendarDemoRSVPResult{}, newCalendarDemoReconcileError(stage, err)
	}

	if output.ID == uuid.Nil || output.UserID == nil || *output.UserID != actorID || strings.TrimSpace(output.Email) == "" || output.RecurrenceID != nil || output.RespondedAt == nil || output.RespondedAt.IsZero() || output.RespondedAt.Location() != time.UTC || output.Version <= expectedVersion {
		return devdemo.CalendarDemoRSVPResult{}, newCalendarDemoReconcileError(stage, errors.New("invalid calendar demo RSVP response"))
	}
	if output.ReplyDeliveryStatus != "not_required" && output.ReplyDeliveryStatus != "pending" {
		return devdemo.CalendarDemoRSVPResult{}, newCalendarDemoReconcileError(stage, errors.New("invalid calendar demo RSVP reply delivery status"))
	}
	if _, err := calendarDemoOutputRole(stage, output.Role); err != nil {
		return devdemo.CalendarDemoRSVPResult{}, err
	}
	outputStatus, err := calendarDemoOutputRSVP(stage, output.RSVPStatus)
	if err != nil {
		return devdemo.CalendarDemoRSVPResult{}, err
	}
	if outputStatus != input.Status {
		return devdemo.CalendarDemoRSVPResult{}, newCalendarDemoReconcileError(stage, errors.New("calendar demo RSVP status mismatch"))
	}
	version, err := calendarDemoUintToInt(stage, output.Version)
	if err != nil {
		return devdemo.CalendarDemoRSVPResult{}, err
	}

	return devdemo.CalendarDemoRSVPResult{
		UserID:   actorID,
		Status:   outputStatus,
		Version:  version,
		Replayed: serviceResult.Replayed(),
	}, nil
}

func (adapter *calendarDemoAdapter) CancelOccurrence(
	ctx context.Context,
	actorID uuid.UUID,
	input devdemo.CalendarDemoOccurrenceCancelInput,
) (devdemo.CalendarDemoOccurrenceCancelResult, error) {
	const stage = calendarDemoStageCancel

	if adapter == nil || adapter.eventService == nil || ctx == nil || actorID == uuid.Nil {
		return devdemo.CalendarDemoOccurrenceCancelResult{}, newCalendarDemoReconcileError(stage, errors.New("invalid calendar demo occurrence cancellation context"))
	}
	if input.EventID == uuid.Nil || input.RecurrenceID.IsZero() || input.RecurrenceID.Location() != time.UTC || input.RecurrenceID.Nanosecond() != 0 {
		return devdemo.CalendarDemoOccurrenceCancelResult{}, newCalendarDemoReconcileError(stage, errors.New("invalid calendar demo occurrence cancellation input"))
	}

	expectedVersion, err := calendarDemoPositiveIntToUint(stage, input.ExpectedMasterVersion)
	if err != nil {
		return devdemo.CalendarDemoOccurrenceCancelResult{}, err
	}
	recurrenceID := input.RecurrenceID.Format(time.RFC3339)
	parsedRecurrenceID, err := time.Parse(time.RFC3339, recurrenceID)
	if err != nil || parsedRecurrenceID.IsZero() || parsedRecurrenceID.Location() != time.UTC || !parsedRecurrenceID.Equal(input.RecurrenceID) || parsedRecurrenceID.Format(time.RFC3339) != recurrenceID {
		return devdemo.CalendarDemoOccurrenceCancelResult{}, newCalendarDemoReconcileError(stage, errors.New("invalid calendar demo occurrence recurrence"))
	}

	command, err := calendarmod.NewCalendarEventOccurrenceCancelCommand(calendarmod.CalendarEventOccurrenceCancelInput{
		EventID:         input.EventID,
		RecurrenceID:    recurrenceID,
		ExpectedVersion: expectedVersion,
	})
	if err != nil {
		return devdemo.CalendarDemoOccurrenceCancelResult{}, newCalendarDemoReconcileError(stage, err)
	}

	serviceResult, err := adapter.eventService.CancelCalendarEventOccurrence(ctx, actorID, command)
	if err != nil {
		return devdemo.CalendarDemoOccurrenceCancelResult{}, newCalendarDemoReconcileError(stage, err)
	}
	if serviceResult.EventID() != input.EventID || serviceResult.RecurrenceID() != recurrenceID || serviceResult.Status() != calendarmod.CalendarEventStatusCancelled || serviceResult.Sequence() == 0 || serviceResult.EventVersion() == 0 {
		return devdemo.CalendarDemoOccurrenceCancelResult{}, newCalendarDemoReconcileError(stage, errors.New("invalid calendar demo occurrence cancellation response"))
	}

	status, err := calendarDemoOutputEventStatus(stage, serviceResult.Status())
	if err != nil {
		return devdemo.CalendarDemoOccurrenceCancelResult{}, err
	}
	sequence, err := calendarDemoUintToInt(stage, serviceResult.Sequence())
	if err != nil {
		return devdemo.CalendarDemoOccurrenceCancelResult{}, err
	}
	eventVersion, err := calendarDemoUintToInt(stage, serviceResult.EventVersion())
	if err != nil {
		return devdemo.CalendarDemoOccurrenceCancelResult{}, err
	}
	returnedRecurrenceID, err := time.Parse(time.RFC3339, serviceResult.RecurrenceID())
	if err != nil || returnedRecurrenceID.IsZero() || returnedRecurrenceID.Location() != time.UTC || returnedRecurrenceID.Format(time.RFC3339) != serviceResult.RecurrenceID() || !returnedRecurrenceID.Equal(input.RecurrenceID) {
		return devdemo.CalendarDemoOccurrenceCancelResult{}, newCalendarDemoReconcileError(stage, errors.New("invalid calendar demo occurrence recurrence response"))
	}

	return devdemo.CalendarDemoOccurrenceCancelResult{
		EventID:        serviceResult.EventID(),
		RecurrenceID:   returnedRecurrenceID,
		Status:         status,
		Sequence:       sequence,
		EventVersion:   eventVersion,
	}, nil
}

func calendarDemoPositiveIntToUint(stage calendarDemoStage, value int) (uint, error) {
	if value <= 0 {
		return 0, newCalendarDemoReconcileError(stage, errors.New("invalid calendar demo version"))
	}

	converted := uint(value)
	if int(converted) != value {
		return 0, newCalendarDemoReconcileError(stage, errors.New("calendar demo version overflow"))
	}

	return converted, nil
}
