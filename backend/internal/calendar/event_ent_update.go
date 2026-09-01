package calendar

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendareventexception"
)

var _ CalendarEventUpdateTransactionBackend = (*eventEntTransactionBackend)(nil)
var _ CalendarEventOccurrenceCancelTransactionBackend = (*eventEntTransactionBackend)(nil)

func (backend *eventEntTransactionBackend) ResolveCalendarEventUpdate(ctx context.Context, scope EventMutationScope, request CalendarEventUpdateAuthorizationRequest) (CalendarEventUpdateAuthorization, error) {
	if backend == nil || backend.tx == nil || request.EventID == uuid.Nil {
		return CalendarEventUpdateAuthorization{}, ErrEventInvalid
	}
	if _, err := backend.activeOrganizationMemberForUpdate(ctx, scope.ActorID(), scope.OrganizationID()); err != nil {
		return CalendarEventUpdateAuthorization{}, err
	}
	event, err := backend.tx.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(request.EventID),
			calendarevent.OrganizationIDEQ(scope.OrganizationID()),
			calendarevent.OrganizerIDEQ(scope.ActorID()),
			calendarevent.DeletedAtIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return CalendarEventUpdateAuthorization{}, calendarEntSafeError(err)
	}
	current, err := eventEntCalendarEventInput(event)
	if err != nil {
		return CalendarEventUpdateAuthorization{}, err
	}
	attendees, err := backend.tx.CalendarAttendee.Query().
		Where(
			calendarattendee.CalendarEventIDEQ(event.ID),
			calendarattendee.RemovedAtIsNil(),
			calendarattendee.RecurrenceIDIsNil(),
		).
		Order(ent.Asc(calendarattendee.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return CalendarEventUpdateAuthorization{}, calendarEntSafeError(err)
	}
	exceptions, err := backend.tx.CalendarEventException.Query().
		Where(calendareventexception.CalendarEventIDEQ(event.ID)).
		Order(ent.Asc(calendareventexception.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return CalendarEventUpdateAuthorization{}, calendarEntSafeError(err)
	}

	resolved := CalendarEventUpdateAuthorization{
		Current:                  current,
		Attendees:                make([]CalendarEventUpdateAttendeeSnapshot, 0, len(attendees)),
		ResolvedAttendees:        make([]CalendarEventCreateResolvedAttendee, len(request.Attendees)),
		HasMaterializedException: len(exceptions) != 0,
	}
	for _, attendee := range attendees {
		snapshot, snapshotErr := eventEntUpdateAttendeeSnapshot(attendee)
		if snapshotErr != nil {
			return CalendarEventUpdateAuthorization{}, snapshotErr
		}
		resolved.Attendees = append(resolved.Attendees, snapshot)
	}
	for index, candidate := range request.Attendees {
		attendee := CalendarEventCreateResolvedAttendee{
			Email:       candidate.Email,
			DisplayName: eventEntStringPointer(candidate.DisplayName),
			Role:        candidate.Role,
		}
		if candidate.UserID != nil {
			member, memberErr := backend.activeOrganizationMemberForUpdate(ctx, *candidate.UserID, scope.OrganizationID())
			if memberErr != nil {
				return CalendarEventUpdateAuthorization{}, memberErr
			}
			memberID := member.ID
			attendee.UserID = &memberID
			attendee.Email = member.Email
			attendee.DisplayName = eventEntOptionalDisplayName(member.DisplayName)
		}
		resolved.ResolvedAttendees[index] = attendee
	}
	return resolved, nil
}

func (backend *eventEntTransactionBackend) PersistCalendarEventUpdate(ctx context.Context, scope EventMutationScope, command CalendarEventUpdatePersistenceCommand) (CalendarEventUpdatePersistenceResult, error) {
	if backend == nil || backend.tx == nil {
		return CalendarEventUpdatePersistenceResult{}, ErrEventInvalid
	}
	input := command.Event().Snapshot()
	if input.EventID == uuid.Nil || input.OrganizationID != scope.OrganizationID() || input.OrganizerID == nil || *input.OrganizerID != scope.ActorID() {
		return CalendarEventUpdatePersistenceResult{}, ErrEventForbidden
	}
	if _, err := backend.activeOrganizationMemberForUpdate(ctx, scope.ActorID(), scope.OrganizationID()); err != nil {
		return CalendarEventUpdatePersistenceResult{}, err
	}
	event, err := backend.tx.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(input.EventID),
			calendarevent.OrganizationIDEQ(scope.OrganizationID()),
			calendarevent.OrganizerIDEQ(scope.ActorID()),
			calendarevent.DeletedAtIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return CalendarEventUpdatePersistenceResult{}, calendarEntSafeError(err)
	}
	lockedAttendees, err := backend.tx.CalendarAttendee.Query().
		Where(calendarattendee.CalendarEventIDEQ(event.ID)).
		Order(ent.Asc(calendarattendee.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return CalendarEventUpdatePersistenceResult{}, calendarEntSafeError(err)
	}
	byID := make(map[uuid.UUID]*ent.CalendarAttendee, len(lockedAttendees))
	for _, attendee := range lockedAttendees {
		byID[attendee.ID] = attendee
	}

	mutations := command.Attendees()
	results := make([]CalendarEventUpdateAttendeePersistenceResult, 0, len(mutations))
	for _, mutation := range mutations {
		if err := backend.persistCalendarEventUpdateAttendee(ctx, event.ID, byID, mutation); err != nil {
			return CalendarEventUpdatePersistenceResult{}, err
		}
		results = append(results, CalendarEventUpdateAttendeePersistenceResult{
			AttendeeID: mutation.AttendeeID,
			Version:    mutation.Version,
			Removed:    mutation.Kind == CalendarEventUpdateAttendeeRemove,
		})
	}
	if err := backend.persistCalendarEventUpdateMaster(ctx, event.ID, command.ExpectedVersion(), input); err != nil {
		return CalendarEventUpdatePersistenceResult{}, err
	}
	for _, descriptor := range command.Outbox() {
		if err := backend.AppendOutbox(ctx, scope, descriptor); err != nil {
			return CalendarEventUpdatePersistenceResult{}, err
		}
	}
	return NewCalendarEventUpdatePersistenceResult(input.EventID, input.Version, input.Sequence, results)
}

func (backend *eventEntTransactionBackend) ResolveCalendarEventOccurrenceCancel(ctx context.Context, scope EventMutationScope, request CalendarEventOccurrenceCancelAuthorizationRequest) (CalendarEventOccurrenceCancelAuthorization, error) {
	if backend == nil || backend.tx == nil || request.EventID == uuid.Nil {
		return CalendarEventOccurrenceCancelAuthorization{}, ErrEventInvalid
	}
	if _, err := backend.activeOrganizationMemberForUpdate(ctx, scope.ActorID(), scope.OrganizationID()); err != nil {
		return CalendarEventOccurrenceCancelAuthorization{}, err
	}
	event, err := backend.tx.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(request.EventID),
			calendarevent.OrganizationIDEQ(scope.OrganizationID()),
			calendarevent.CalendarOwnerIDEQ(scope.ActorID()),
			calendarevent.OrganizerIDEQ(scope.ActorID()),
			calendarevent.DeletedAtIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return CalendarEventOccurrenceCancelAuthorization{}, calendarEntSafeError(err)
	}
	current, err := eventEntCalendarEventInput(event)
	if err != nil {
		return CalendarEventOccurrenceCancelAuthorization{}, err
	}
	attendees, err := backend.tx.CalendarAttendee.Query().
		Where(calendarattendee.CalendarEventIDEQ(event.ID)).
		Order(ent.Asc(calendarattendee.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return CalendarEventOccurrenceCancelAuthorization{}, calendarEntSafeError(err)
	}
	exceptions, err := backend.tx.CalendarEventException.Query().
		Where(
			calendareventexception.CalendarEventIDEQ(event.ID),
			calendareventexception.RecurrenceIDEQ(request.RecurrenceID),
		).
		Order(ent.Asc(calendareventexception.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return CalendarEventOccurrenceCancelAuthorization{}, calendarEntSafeError(err)
	}
	if len(exceptions) > 1 {
		return CalendarEventOccurrenceCancelAuthorization{}, ErrEventInvalid
	}

	resolved := CalendarEventOccurrenceCancelAuthorization{
		Current:   current,
		Attendees: make([]CalendarEventOccurrenceCancelAttendeeSnapshot, 0, len(attendees)),
	}
	for _, attendee := range attendees {
		snapshot, snapshotErr := eventEntOccurrenceCancelAttendeeSnapshot(attendee)
		if snapshotErr != nil {
			return CalendarEventOccurrenceCancelAuthorization{}, snapshotErr
		}
		resolved.Attendees = append(resolved.Attendees, snapshot)
	}
	if len(exceptions) == 1 {
		exception, exceptionErr := eventEntCalendarEventExceptionInput(exceptions[0])
		if exceptionErr != nil {
			return CalendarEventOccurrenceCancelAuthorization{}, exceptionErr
		}
		resolved.Exception = &exception
	}
	return resolved, nil
}

func (backend *eventEntTransactionBackend) PersistCalendarEventOccurrenceCancel(ctx context.Context, scope EventMutationScope, command CalendarEventOccurrenceCancelPersistenceCommand) (CalendarEventOccurrenceCancelPersistenceResult, error) {
	if backend == nil || backend.tx == nil {
		return CalendarEventOccurrenceCancelPersistenceResult{}, ErrEventInvalid
	}
	input := command.Event().Snapshot()
	exceptionInput := command.Exception().Snapshot()
	if input.EventID == uuid.Nil || input.OrganizationID != scope.OrganizationID() || input.OwnerID != scope.ActorID() || input.OrganizerID == nil || *input.OrganizerID != scope.ActorID() {
		return CalendarEventOccurrenceCancelPersistenceResult{}, ErrEventForbidden
	}
	if _, err := backend.activeOrganizationMemberForUpdate(ctx, scope.ActorID(), scope.OrganizationID()); err != nil {
		return CalendarEventOccurrenceCancelPersistenceResult{}, err
	}
	event, err := backend.tx.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(input.EventID),
			calendarevent.OrganizationIDEQ(scope.OrganizationID()),
			calendarevent.CalendarOwnerIDEQ(scope.ActorID()),
			calendarevent.OrganizerIDEQ(scope.ActorID()),
			calendarevent.DeletedAtIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return CalendarEventOccurrenceCancelPersistenceResult{}, calendarEntSafeError(err)
	}
	exceptions, err := backend.tx.CalendarEventException.Query().
		Where(
			calendareventexception.CalendarEventIDEQ(event.ID),
			calendareventexception.RecurrenceIDEQ(exceptionInput.RecurrenceID),
		).
		Order(ent.Asc(calendareventexception.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return CalendarEventOccurrenceCancelPersistenceResult{}, calendarEntSafeError(err)
	}
	if len(exceptions) > 1 || len(exceptions) == 1 && exceptions[0].ID != exceptionInput.ExceptionID {
		return CalendarEventOccurrenceCancelPersistenceResult{}, ErrEventConflict
	}
	if err := backend.persistCalendarEventOccurrenceException(ctx, exceptions, exceptionInput); err != nil {
		return CalendarEventOccurrenceCancelPersistenceResult{}, err
	}
	if err := backend.persistCalendarEventUpdateMaster(ctx, event.ID, command.ExpectedVersion(), input); err != nil {
		return CalendarEventOccurrenceCancelPersistenceResult{}, err
	}
	if err := backend.AppendOutbox(ctx, scope, command.Outbox()); err != nil {
		return CalendarEventOccurrenceCancelPersistenceResult{}, err
	}
	eventCommand, eventErr := NewCalendarEventCommand(input)
	if eventErr != nil {
		return CalendarEventOccurrenceCancelPersistenceResult{}, ErrEventInvalid
	}
	exceptionCommand, exceptionErr := NewCalendarEventExceptionCommand(exceptionInput)
	if exceptionErr != nil {
		return CalendarEventOccurrenceCancelPersistenceResult{}, ErrEventInvalid
	}
	return NewCalendarEventOccurrenceCancelPersistenceResult(eventCommand, exceptionCommand, command.Outbox())
}

func (backend *eventEntTransactionBackend) persistCalendarEventUpdateAttendee(ctx context.Context, eventID uuid.UUID, locked map[uuid.UUID]*ent.CalendarAttendee, mutation CalendarEventUpdateAttendeeMutation) error {
	version, ok := eventEntInt(mutation.Version)
	if !ok {
		return ErrEventInvalid
	}
	switch mutation.Kind {
	case CalendarEventUpdateAttendeeAdd:
		create := backend.tx.CalendarAttendee.Create().
			SetID(mutation.AttendeeID).
			SetCalendarEventID(eventID).
			SetNormalizedEmail(mutation.Email).
			SetRole(calendarattendee.Role(mutation.Role)).
			SetRsvpStatus(calendarattendee.RsvpStatus(mutation.Status)).
			SetVersion(version)
		if mutation.UserID != nil {
			create.SetUserID(*mutation.UserID)
		}
		if mutation.DisplayName != "" {
			create.SetDisplayName(mutation.DisplayName)
		}
		if mutation.RespondedAt != nil {
			create.SetRespondedAt(mutation.RespondedAt.UTC())
		}
		if _, err := create.Save(ctx); err != nil {
			return calendarEntSafeError(err)
		}
		return nil
	case CalendarEventUpdateAttendeeReset, CalendarEventUpdateAttendeeRemove:
		if mutation.Version == 0 {
			return ErrEventInvalid
		}
		if _, found := locked[mutation.AttendeeID]; !found {
			return ErrEventConflict
		}
		expectedVersion, expectedOK := eventEntInt(mutation.Version - 1)
		if !expectedOK {
			return ErrEventInvalid
		}
		update := backend.tx.CalendarAttendee.UpdateOneID(mutation.AttendeeID).
			Where(
				calendarattendee.CalendarEventIDEQ(eventID),
				calendarattendee.VersionEQ(expectedVersion),
			).
			SetVersion(version)
		if mutation.Kind == CalendarEventUpdateAttendeeReset {
			update.SetNormalizedEmail(mutation.Email).
				SetRole(calendarattendee.Role(mutation.Role)).
				SetRsvpStatus(calendarattendee.RsvpStatus(mutation.Status)).
				ClearRemovedAt()
			if mutation.UserID == nil {
				update.ClearUserID()
			} else {
				update.SetUserID(*mutation.UserID)
			}
			if mutation.DisplayName == "" {
				update.ClearDisplayName()
			} else {
				update.SetDisplayName(mutation.DisplayName)
			}
			if mutation.RespondedAt == nil {
				update.ClearRespondedAt()
			} else {
				update.SetRespondedAt(mutation.RespondedAt.UTC())
			}
		} else {
			if mutation.RemovedAt == nil {
				return ErrEventInvalid
			}
			update.SetRemovedAt(mutation.RemovedAt.UTC())
		}
		if err := update.Exec(ctx); err != nil {
			return calendarEntCASFailure(err)
		}
		return nil
	default:
		return ErrEventInvalid
	}
}

func (backend *eventEntTransactionBackend) persistCalendarEventUpdateMaster(ctx context.Context, eventID uuid.UUID, expectedVersion uint, input CalendarEventCommandInput) error {
	expected, expectedOK := eventEntInt(expectedVersion)
	version, versionOK := eventEntInt(input.Version)
	sequence, sequenceOK := eventEntInt(input.Sequence)
	if !expectedOK || !versionOK || !sequenceOK {
		return ErrEventInvalid
	}
	update := backend.tx.CalendarEvent.UpdateOneID(eventID).
		Where(calendarevent.VersionEQ(expected)).
		SetTitle(input.Title).
		SetDescription(input.Description).
		SetLocation(input.Location).
		SetAllDay(input.Temporal.AllDay).
		SetTimezone(input.Temporal.Timezone).
		SetExdates(append([]string{}, input.Exdates...)).
		SetSequence(sequence).
		SetVersion(version).
		SetStatus(calendarevent.Status(input.Status))
	if input.Temporal.AllDay {
		startDate, startErr := time.Parse(time.DateOnly, input.Temporal.StartDate)
		endDate, endErr := time.Parse(time.DateOnly, input.Temporal.EndDateExclusive)
		if startErr != nil || endErr != nil {
			return ErrEventInvalid
		}
		update.ClearStartsAt().
			ClearEndsAt().
			SetStartDate(startDate).
			SetEndDateExclusive(endDate)
	} else {
		update.ClearStartDate().
			ClearEndDateExclusive().
			SetStartsAt(input.Temporal.StartsAt.UTC()).
			SetEndsAt(input.Temporal.EndsAt.UTC())
	}
	if input.RRULE == "" {
		update.ClearRrule()
	} else {
		update.SetRrule(input.RRULE)
	}
	if input.ReminderMinutes == nil {
		update.ClearReminderMinutes()
	} else {
		update.SetReminderMinutes(*input.ReminderMinutes)
	}
	if err := update.Exec(ctx); err != nil {
		return calendarEntCASFailure(err)
	}
	return nil
}

func (backend *eventEntTransactionBackend) persistCalendarEventOccurrenceException(ctx context.Context, existing []*ent.CalendarEventException, input CalendarEventExceptionCommandInput) error {
	sequence, ok := eventEntInt(input.Sequence)
	if !ok {
		return ErrEventInvalid
	}
	if len(existing) == 0 {
		create := backend.tx.CalendarEventException.Create().
			SetID(input.ExceptionID).
			SetCalendarEventID(input.EventID).
			SetRecurrenceID(input.RecurrenceID).
			SetSequence(sequence).
			SetStatus(calendareventexception.Status(input.Status))
		if err := eventEntSetExceptionCreateTemporal(create, input.Temporal); err != nil {
			return err
		}
		if _, err := create.Save(ctx); err != nil {
			return calendarEntSafeError(err)
		}
		return nil
	}
	update := backend.tx.CalendarEventException.UpdateOneID(existing[0].ID).
		Where(
			calendareventexception.CalendarEventIDEQ(input.EventID),
			calendareventexception.RecurrenceIDEQ(input.RecurrenceID),
		).
		SetSequence(sequence).
		SetStatus(calendareventexception.Status(input.Status))
	if err := eventEntSetExceptionUpdateTemporal(update, input.Temporal); err != nil {
		return err
	}
	if err := update.Exec(ctx); err != nil {
		return calendarEntCASFailure(err)
	}
	return nil
}

func eventEntSetExceptionCreateTemporal(target *ent.CalendarEventExceptionCreate, temporal *CalendarEvent) error {
	if temporal == nil {
		return nil
	}
	if temporal.Validate() != nil {
		return ErrEventInvalid
	}
	if temporal.AllDay {
		startDate, startErr := time.Parse(time.DateOnly, temporal.StartDate)
		endDate, endErr := time.Parse(time.DateOnly, temporal.EndDateExclusive)
		if startErr != nil || endErr != nil {
			return ErrEventInvalid
		}
		target.SetStartDate(startDate).SetEndDateExclusive(endDate)
	} else {
		target.SetStartsAt(temporal.StartsAt.UTC()).SetEndsAt(temporal.EndsAt.UTC())
	}
	target.SetTimezone(temporal.Timezone)
	return nil
}

func eventEntSetExceptionUpdateTemporal(target *ent.CalendarEventExceptionUpdateOne, temporal *CalendarEvent) error {
	target.ClearStartsAt().ClearEndsAt().ClearStartDate().ClearEndDateExclusive().ClearTimezone()
	if temporal == nil {
		return nil
	}
	if temporal.Validate() != nil {
		return ErrEventInvalid
	}
	if temporal.AllDay {
		startDate, startErr := time.Parse(time.DateOnly, temporal.StartDate)
		endDate, endErr := time.Parse(time.DateOnly, temporal.EndDateExclusive)
		if startErr != nil || endErr != nil {
			return ErrEventInvalid
		}
		target.SetStartDate(startDate).SetEndDateExclusive(endDate)
	} else {
		target.SetStartsAt(temporal.StartsAt.UTC()).SetEndsAt(temporal.EndsAt.UTC())
	}
	target.SetTimezone(temporal.Timezone)
	return nil
}

func eventEntCalendarEventInput(event *ent.CalendarEvent) (CalendarEventCommandInput, error) {
	if event == nil {
		return CalendarEventCommandInput{}, ErrEventInvalid
	}
	version, versionOK := eventEntUint(event.Version)
	sequence, sequenceOK := eventEntUint(event.Sequence)
	if !versionOK || !sequenceOK {
		return CalendarEventCommandInput{}, ErrEventInvalid
	}
	temporal := CalendarEvent{AllDay: event.AllDay, Timezone: event.Timezone}
	if event.AllDay {
		if event.StartDate == nil || event.EndDateExclusive == nil || event.StartsAt != nil || event.EndsAt != nil {
			return CalendarEventCommandInput{}, ErrEventInvalid
		}
		temporal.StartDate = event.StartDate.Format(time.DateOnly)
		temporal.EndDateExclusive = event.EndDateExclusive.Format(time.DateOnly)
	} else {
		if event.StartsAt == nil || event.EndsAt == nil || event.StartDate != nil || event.EndDateExclusive != nil {
			return CalendarEventCommandInput{}, ErrEventInvalid
		}
		temporal.StartsAt = event.StartsAt.UTC()
		temporal.EndsAt = event.EndsAt.UTC()
	}
	input := CalendarEventCommandInput{
		EventID:         event.ID,
		OrganizationID:  event.OrganizationID,
		OwnerID:         event.CalendarOwnerID,
		UID:             event.UID,
		Title:           event.Title,
		Description:     event.Description,
		Location:        event.Location,
		Temporal:        temporal,
		Exdates:         append([]string(nil), event.Exdates...),
		Sequence:        sequence,
		Version:         version,
		Source:          CalendarEventSource(event.Source),
		Status:          CalendarEventStatus(event.Status),
		OrganizerID:     eventEntUUIDPointer(event.OrganizerID),
		OrganizerEmail:  event.OrganizerEmail,
		ScheduledCallID: eventEntUUIDPointer(event.ScheduledCallID),
		ReminderMinutes: eventEntIntPointer(event.ReminderMinutes),
	}
	if event.Rrule != nil {
		input.RRULE = *event.Rrule
	}
	if _, err := NewCalendarEventCommand(input); err != nil {
		return CalendarEventCommandInput{}, ErrEventInvalid
	}
	return input, nil
}

func eventEntUpdateAttendeeSnapshot(attendee *ent.CalendarAttendee) (CalendarEventUpdateAttendeeSnapshot, error) {
	if attendee == nil {
		return CalendarEventUpdateAttendeeSnapshot{}, ErrEventInvalid
	}
	var recurrenceID string
	if attendee.RecurrenceID != nil {
		recurrenceID = *attendee.RecurrenceID
	}
	role := CalendarEventAttendeeRole(attendee.Role)
	status := RSVPStatus(attendee.RsvpStatus)
	if !validCalendarEventStoredAttendeeRoleState(role, status, recurrenceID) {
		return CalendarEventUpdateAttendeeSnapshot{}, ErrEventInvalid
	}
	version, ok := eventEntUint(attendee.Version)
	if !ok {
		return CalendarEventUpdateAttendeeSnapshot{}, ErrEventInvalid
	}
	return CalendarEventUpdateAttendeeSnapshot{
		AttendeeID:  attendee.ID,
		UserID:      eventEntUUIDPointer(attendee.UserID),
		Email:       attendee.NormalizedEmail,
		DisplayName: eventEntStringValue(attendee.DisplayName),
		Role:        role,
		Status:      status,
		RespondedAt: eventEntTimePointer(attendee.RespondedAt),
		Version:     version,
	}, nil
}

func eventEntOccurrenceCancelAttendeeSnapshot(attendee *ent.CalendarAttendee) (CalendarEventOccurrenceCancelAttendeeSnapshot, error) {
	if attendee == nil {
		return CalendarEventOccurrenceCancelAttendeeSnapshot{}, ErrEventInvalid
	}
	version, ok := eventEntUint(attendee.Version)
	if !ok {
		return CalendarEventOccurrenceCancelAttendeeSnapshot{}, ErrEventInvalid
	}
	return CalendarEventOccurrenceCancelAttendeeSnapshot{
		AttendeeID: attendee.ID,
		EventID:    attendee.CalendarEventID,
		UserID:     eventEntUUIDPointer(attendee.UserID),
		Status:     RSVPStatus(attendee.RsvpStatus),
		Version:    version,
		RemovedAt:  eventEntTimePointer(attendee.RemovedAt),
	}, nil
}

func eventEntCalendarEventExceptionInput(exception *ent.CalendarEventException) (CalendarEventExceptionCommandInput, error) {
	if exception == nil || exception.Status == nil {
		return CalendarEventExceptionCommandInput{}, ErrEventInvalid
	}
	sequence, ok := eventEntUint(exception.Sequence)
	if !ok {
		return CalendarEventExceptionCommandInput{}, ErrEventInvalid
	}
	input := CalendarEventExceptionCommandInput{
		ExceptionID:  exception.ID,
		EventID:      exception.CalendarEventID,
		RecurrenceID: exception.RecurrenceID,
		Sequence:     sequence,
		Status:       CalendarEventStatus(*exception.Status),
		Cancelled:    CalendarEventStatus(*exception.Status) == CalendarEventStatusCancelled,
	}
	if exception.StartsAt != nil || exception.EndsAt != nil || exception.StartDate != nil || exception.EndDateExclusive != nil || exception.Timezone != nil {
		if exception.Timezone == nil {
			return CalendarEventExceptionCommandInput{}, ErrEventInvalid
		}
		temporal := CalendarEvent{Timezone: *exception.Timezone}
		switch {
		case exception.StartsAt != nil && exception.EndsAt != nil && exception.StartDate == nil && exception.EndDateExclusive == nil:
			temporal.StartsAt = exception.StartsAt.UTC()
			temporal.EndsAt = exception.EndsAt.UTC()
		case exception.StartsAt == nil && exception.EndsAt == nil && exception.StartDate != nil && exception.EndDateExclusive != nil:
			temporal.AllDay = true
			temporal.StartDate = exception.StartDate.Format(time.DateOnly)
			temporal.EndDateExclusive = exception.EndDateExclusive.Format(time.DateOnly)
		default:
			return CalendarEventExceptionCommandInput{}, ErrEventInvalid
		}
		input.Temporal = &temporal
	}
	if _, err := NewCalendarEventExceptionCommand(input); err != nil {
		return CalendarEventExceptionCommandInput{}, ErrEventInvalid
	}
	return input, nil
}

func calendarEntCASFailure(err error) error {
	if ent.IsNotFound(err) || ent.IsConstraintError(err) {
		return ErrEventConflict
	}
	return calendarEntSafeError(err)
}

func eventEntInt(value uint) (int, bool) {
	maximum := uint(^uint(0) >> 1)
	if value > maximum {
		return 0, false
	}
	return int(value), true
}

func eventEntUint(value int) (uint, bool) {
	if value < 0 {
		return 0, false
	}
	return uint(value), true
}

func eventEntIntPointer(value *int) *int {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func eventEntStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func eventEntTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := value.UTC()
	return &copied
}
