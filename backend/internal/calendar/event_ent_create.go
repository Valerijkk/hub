package calendar

import (
	"bytes"
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendareventexception"
	"github.com/valerii/hub/backend/ent/calendarmutationreceipt"
	"github.com/valerii/hub/backend/internal/platform/outbox"
)

func (backend *eventEntTransactionBackend) ReserveMutationReceipt(ctx context.Context, scope EventMutationScope, reservation MutationReceiptReservation) (MutationReceiptReservationResult, error) {
	if backend == nil || backend.tx == nil || backend.now == nil {
		return MutationReceiptReservationResult{}, ErrEventInvalid
	}

	keyHash := reservation.KeyHash()
	requestHash := reservation.RequestHash()
	operation := calendarmutationreceipt.Operation(scope.Operation())
	now := backend.now().UTC().Truncate(time.Second)
	receipt, err := backend.tx.CalendarMutationReceipt.Query().
		Where(
			calendarmutationreceipt.OrganizationIDEQ(scope.OrganizationID()),
			calendarmutationreceipt.ActorUserIDEQ(scope.ActorID()),
			calendarmutationreceipt.OperationEQ(operation),
			calendarmutationreceipt.IdempotencyKeyHashEQ(keyHash[:]),
		).
		ForUpdate().
		Only(ctx)
	if err == nil {
		if !bytes.Equal(receipt.RequestHash, requestHash[:]) ||
			receipt.State != calendarmutationreceipt.StateCompleted ||
			receipt.ResponseStatus == nil ||
			receipt.ResponseJSON == nil ||
			!receipt.ExpiresAt.After(now) {
			return MutationReceiptReservationResult{}, newCalendarEventCreateConflict(CalendarEventCreateConflictIdempotencyKeyReused)
		}
		response, err := NewMutationReceiptResponse(*receipt.ResponseStatus, *receipt.ResponseJSON)
		if err != nil {
			return MutationReceiptReservationResult{}, ErrEventInvalid
		}
		return NewReplayedMutationReceiptResult(receipt.ID, response)
	}
	if !ent.IsNotFound(err) {
		return MutationReceiptReservationResult{}, calendarEntSafeError(err)
	}

	expiresAt := now.Add(calendarMutationReceiptTTL)
	receipt, err = backend.tx.CalendarMutationReceipt.Create().
		SetOrganizationID(scope.OrganizationID()).
		SetActorUserID(scope.ActorID()).
		SetOperation(operation).
		SetIdempotencyKeyHash(keyHash[:]).
		SetRequestHash(requestHash[:]).
		SetExpiresAt(expiresAt).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return MutationReceiptReservationResult{}, newCalendarEventCreateConflict(CalendarEventCreateConflictIdempotencyKeyReused)
		}
		return MutationReceiptReservationResult{}, calendarEntSafeError(err)
	}
	return NewReservedMutationReceiptResult(receipt.ID)
}

func (backend *eventEntTransactionBackend) CompleteMutationReceipt(ctx context.Context, scope EventMutationScope, completion MutationReceiptCompletion) error {
	if backend == nil || backend.tx == nil {
		return ErrEventInvalid
	}

	operation := calendarmutationreceipt.Operation(scope.Operation())
	receipt, err := backend.tx.CalendarMutationReceipt.Query().
		Where(
			calendarmutationreceipt.IDEQ(completion.ReceiptID()),
			calendarmutationreceipt.OrganizationIDEQ(scope.OrganizationID()),
			calendarmutationreceipt.ActorUserIDEQ(scope.ActorID()),
			calendarmutationreceipt.OperationEQ(operation),
		).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return ErrEventConflict
	}
	if err != nil {
		return calendarEntSafeError(err)
	}
	if receipt.State != calendarmutationreceipt.StateInProgress {
		return ErrEventConflict
	}

	response := completion.Response()
	update := backend.tx.CalendarMutationReceipt.UpdateOne(receipt).
		SetState(calendarmutationreceipt.StateCompleted).
		SetResponseStatus(response.Status()).
		SetResponseJSON(response.JSON())
	if eventID, found := completion.EventID(); found {
		update.SetCalendarEventID(eventID)
	}
	if err := update.Exec(ctx); err != nil {
		return calendarEntSafeError(err)
	}
	return nil
}

func (backend *eventEntTransactionBackend) PersistCalendarEvent(ctx context.Context, scope EventMutationScope, command CalendarEventCommand) (CalendarEventPersistenceResult, error) {
	if backend == nil || backend.tx == nil {
		return CalendarEventPersistenceResult{}, ErrEventInvalid
	}

	input := command.Snapshot()
	if input.OrganizationID != scope.OrganizationID() || input.OwnerID != scope.ActorID() || input.OrganizerID == nil || *input.OrganizerID != scope.ActorID() {
		return CalendarEventPersistenceResult{}, ErrEventForbidden
	}
	startsAt, endsAt, startDate, endDateExclusive, err := eventEntTemporalValues(input.Temporal)
	if err != nil {
		return CalendarEventPersistenceResult{}, err
	}
	create := backend.tx.CalendarEvent.Create().
		SetID(input.EventID).
		SetOrganizationID(input.OrganizationID).
		SetCalendarOwnerID(input.OwnerID).
		SetUID(input.UID).
		SetTitle(input.Title).
		SetDescription(input.Description).
		SetLocation(input.Location).
		SetAllDay(input.Temporal.AllDay).
		SetTimezone(input.Temporal.Timezone).
		SetExdates(append([]string{}, input.Exdates...)).
		SetSequence(int(input.Sequence)).
		SetVersion(int(input.Version)).
		SetSource(calendarevent.Source(input.Source)).
		SetStatus(calendarevent.Status(input.Status)).
		SetOrganizerID(*input.OrganizerID).
		SetOrganizerEmail(input.OrganizerEmail)
	if startsAt != nil {
		create.SetStartsAt(*startsAt).SetEndsAt(*endsAt)
	} else {
		create.SetStartDate(*startDate).SetEndDateExclusive(*endDateExclusive)
	}
	if input.RRULE != "" {
		create.SetRrule(input.RRULE)
	}
	if input.ScheduledCallID != nil {
		create.SetScheduledCallID(*input.ScheduledCallID)
	}
	if input.ReminderMinutes != nil {
		create.SetReminderMinutes(*input.ReminderMinutes)
	}
	event, err := create.Save(ctx)
	if err != nil {
		return CalendarEventPersistenceResult{}, calendarEntSafeError(err)
	}
	return NewCalendarEventPersistenceResult(event.ID, uint(event.Version))
}

func (backend *eventEntTransactionBackend) PersistCalendarEventAttendee(ctx context.Context, scope EventMutationScope, command CalendarEventAttendeeCommand) (CalendarEventAttendeePersistenceResult, error) {
	if backend == nil || backend.tx == nil || scope.OrganizationID() == uuid.Nil {
		return CalendarEventAttendeePersistenceResult{}, ErrEventInvalid
	}

	input := command.input
	input.UserID = eventEntUUIDPointer(input.UserID)
	create := backend.tx.CalendarAttendee.Create().
		SetID(input.AttendeeID).
		SetCalendarEventID(input.EventID).
		SetNormalizedEmail(input.Email).
		SetRole(calendarattendee.Role(input.Role)).
		SetRsvpStatus(calendarattendee.RsvpStatus(input.Status)).
		SetVersion(int(input.Version))
	if input.UserID != nil {
		create.SetUserID(*input.UserID)
	}
	if input.DisplayName != "" {
		create.SetDisplayName(input.DisplayName)
	}
	if input.RecurrenceID != "" {
		create.SetRecurrenceID(input.RecurrenceID)
	}
	attendee, err := create.Save(ctx)
	if err != nil {
		return CalendarEventAttendeePersistenceResult{}, calendarEntSafeError(err)
	}
	return NewCalendarEventAttendeePersistenceResult(attendee.ID, uint(attendee.Version))
}

func (backend *eventEntTransactionBackend) PersistCalendarEventException(ctx context.Context, scope EventMutationScope, command CalendarEventExceptionCommand) (CalendarEventExceptionPersistenceResult, error) {
	if backend == nil || backend.tx == nil || scope.OrganizationID() == uuid.Nil {
		return CalendarEventExceptionPersistenceResult{}, ErrEventInvalid
	}

	input := command.Snapshot()
	create := backend.tx.CalendarEventException.Create().
		SetID(input.ExceptionID).
		SetCalendarEventID(input.EventID).
		SetRecurrenceID(input.RecurrenceID).
		SetSequence(int(input.Sequence)).
		SetStatus(calendareventexception.Status(input.Status))
	if input.Temporal != nil {
		startsAt, endsAt, startDate, endDateExclusive, err := eventEntTemporalValues(*input.Temporal)
		if err != nil {
			return CalendarEventExceptionPersistenceResult{}, err
		}
		if startsAt != nil {
			create.SetStartsAt(*startsAt).SetEndsAt(*endsAt).SetTimezone(input.Temporal.Timezone)
		} else {
			create.SetStartDate(*startDate).SetEndDateExclusive(*endDateExclusive).SetTimezone(input.Temporal.Timezone)
		}
	}
	exception, err := create.Save(ctx)
	if err != nil {
		return CalendarEventExceptionPersistenceResult{}, calendarEntSafeError(err)
	}
	return NewCalendarEventExceptionPersistenceResult(exception.ID, uint(exception.Sequence))
}

func (backend *eventEntTransactionBackend) AppendOutbox(ctx context.Context, scope EventMutationScope, descriptor EventOutboxDescriptor) error {
	if backend == nil || backend.tx == nil || scope.OrganizationID() == uuid.Nil {
		return ErrEventInvalid
	}
	envelope := descriptor.Envelope()
	return calendarEntSafeError(outbox.NewPublisher().PublishInTx(ctx, backend.tx, envelope.Subject(), envelope))
}

func eventEntTemporalValues(temporal CalendarEvent) (*time.Time, *time.Time, *time.Time, *time.Time, error) {
	if !temporal.AllDay {
		startsAt := temporal.StartsAt.UTC()
		endsAt := temporal.EndsAt.UTC()
		return &startsAt, &endsAt, nil, nil, nil
	}
	startDate, err := time.Parse("2006-01-02", temporal.StartDate)
	if err != nil {
		return nil, nil, nil, nil, ErrEventInvalid
	}
	endDateExclusive, err := time.Parse("2006-01-02", temporal.EndDateExclusive)
	if err != nil {
		return nil, nil, nil, nil, ErrEventInvalid
	}
	return nil, nil, &startDate, &endDateExclusive, nil
}
