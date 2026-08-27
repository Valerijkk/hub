package calendar

import (
	"context"
	"reflect"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendareventexception"
	"github.com/valerii/hub/backend/ent/calendarreplydelivery"
)

var _ CalendarEventSeriesRSVPTransactionBackend = (*eventEntTransactionBackend)(nil)
var _ CalendarEventOccurrenceRSVPTransactionBackend = (*eventEntTransactionBackend)(nil)

func (backend *eventEntTransactionBackend) ResolveCalendarEventSeriesRSVP(ctx context.Context, scope EventMutationScope, request CalendarEventSeriesRSVPAuthorizationRequest) (CalendarEventSeriesRSVPAuthorization, error) {
	if backend == nil || backend.tx == nil || request.EventID == uuid.Nil {
		return CalendarEventSeriesRSVPAuthorization{}, ErrEventInvalid
	}
	if _, err := backend.activeOrganizationMemberForUpdate(ctx, scope.ActorID(), scope.OrganizationID()); err != nil {
		return CalendarEventSeriesRSVPAuthorization{}, err
	}
	event, err := backend.tx.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(request.EventID),
			calendarevent.OrganizationIDEQ(scope.OrganizationID()),
			calendarevent.DeletedAtIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return CalendarEventSeriesRSVPAuthorization{}, calendarEntSafeError(err)
	}
	attendees, err := backend.tx.CalendarAttendee.Query().
		Where(calendarattendee.CalendarEventIDEQ(event.ID)).
		Order(ent.Asc(calendarattendee.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return CalendarEventSeriesRSVPAuthorization{}, calendarEntSafeError(err)
	}
	attendee, found := eventEntSeriesAttendeeForActor(attendees, scope.ActorID())
	if !found {
		return CalendarEventSeriesRSVPAuthorization{}, ErrEventNotFound
	}
	current, currentErr := eventEntCalendarEventInput(event)
	if currentErr != nil {
		return CalendarEventSeriesRSVPAuthorization{}, currentErr
	}
	snapshot, snapshotErr := eventEntSeriesRSVPAttendeeSnapshot(attendee)
	if snapshotErr != nil {
		return CalendarEventSeriesRSVPAuthorization{}, snapshotErr
	}
	return CalendarEventSeriesRSVPAuthorization{Current: current, Attendee: snapshot}, nil
}

func (backend *eventEntTransactionBackend) PersistCalendarEventSeriesRSVP(ctx context.Context, scope EventMutationScope, command CalendarEventSeriesRSVPPersistenceCommand) (CalendarEventSeriesRSVPPersistenceResult, error) {
	if backend == nil || backend.tx == nil || !command.valid() {
		return CalendarEventSeriesRSVPPersistenceResult{}, ErrEventInvalid
	}
	requested := command.Attendee()
	if requested.UserID == nil || *requested.UserID != scope.ActorID() {
		return CalendarEventSeriesRSVPPersistenceResult{}, ErrEventForbidden
	}
	if _, err := backend.activeOrganizationMemberForUpdate(ctx, scope.ActorID(), scope.OrganizationID()); err != nil {
		return CalendarEventSeriesRSVPPersistenceResult{}, err
	}
	event, err := backend.tx.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(requested.EventID),
			calendarevent.OrganizationIDEQ(scope.OrganizationID()),
			calendarevent.DeletedAtIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return CalendarEventSeriesRSVPPersistenceResult{}, calendarEntSafeError(err)
	}
	sequence, sequenceOK := eventEntUint(event.Sequence)
	if !sequenceOK || sequence != command.EventSequence() {
		return CalendarEventSeriesRSVPPersistenceResult{}, ErrEventConflict
	}
	attendees, err := backend.tx.CalendarAttendee.Query().
		Where(calendarattendee.CalendarEventIDEQ(event.ID)).
		Order(ent.Asc(calendarattendee.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return CalendarEventSeriesRSVPPersistenceResult{}, calendarEntSafeError(err)
	}
	attendee, found := eventEntSeriesAttendeeForActor(attendees, scope.ActorID())
	if !found || attendee.ID != requested.AttendeeID {
		return CalendarEventSeriesRSVPPersistenceResult{}, ErrEventConflict
	}
	locked, lockedErr := eventEntSeriesRSVPAttendeeSnapshot(attendee)
	if lockedErr != nil || !eventEntSeriesRSVPImmutableIdentityMatches(locked, requested) || locked.Version != command.ExpectedVersion() {
		return CalendarEventSeriesRSVPPersistenceResult{}, ErrEventConflict
	}
	expectedVersion, expectedOK := eventEntInt(command.ExpectedVersion())
	version, versionOK := eventEntInt(requested.Version)
	if !expectedOK || !versionOK || requested.RespondedAt == nil {
		return CalendarEventSeriesRSVPPersistenceResult{}, ErrEventInvalid
	}
	update := backend.tx.CalendarAttendee.UpdateOneID(attendee.ID).
		Where(
			calendarattendee.CalendarEventIDEQ(event.ID),
			calendarattendee.UserIDEQ(scope.ActorID()),
			calendarattendee.RecurrenceIDIsNil(),
			calendarattendee.RemovedAtIsNil(),
			calendarattendee.VersionEQ(expectedVersion),
		).
		SetRsvpStatus(calendarattendee.RsvpStatus(requested.Status)).
		SetRespondedAt(requested.RespondedAt.UTC()).
		SetVersion(version)
	reply, hasReply := command.Reply()
	if hasReply {
		update.SetReplyDeliveryStatus(calendarattendee.ReplyDeliveryStatusPending)
	} else {
		update.SetReplyDeliveryStatus(calendarattendee.ReplyDeliveryStatusNotRequired)
	}
	if err := update.Exec(ctx); err != nil {
		return CalendarEventSeriesRSVPPersistenceResult{}, calendarEntCASFailure(err)
	}
	if hasReply {
		if err := backend.persistCalendarReplyDelivery(ctx, event.OrganizationID, event.ID, attendee.ID, command.EventSequence(), requested.Version, nil, requested.Status); err != nil {
			return CalendarEventSeriesRSVPPersistenceResult{}, err
		}
	}
	if err := backend.AppendOutbox(ctx, scope, command.Outbox()); err != nil {
		return CalendarEventSeriesRSVPPersistenceResult{}, err
	}
	return NewCalendarEventSeriesRSVPPersistenceResult(requested, command.EventSequence(), command.Outbox(), eventEntSeriesReplyPointer(reply, hasReply))
}

func (backend *eventEntTransactionBackend) ResolveCalendarEventOccurrenceRSVP(ctx context.Context, scope EventMutationScope, request CalendarEventOccurrenceRSVPAuthorizationRequest) (CalendarEventOccurrenceRSVPAuthorization, error) {
	if backend == nil || backend.tx == nil || request.EventID == uuid.Nil {
		return CalendarEventOccurrenceRSVPAuthorization{}, ErrEventInvalid
	}
	if _, err := backend.activeOrganizationMemberForUpdate(ctx, scope.ActorID(), scope.OrganizationID()); err != nil {
		return CalendarEventOccurrenceRSVPAuthorization{}, err
	}
	event, err := backend.tx.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(request.EventID),
			calendarevent.OrganizationIDEQ(scope.OrganizationID()),
			calendarevent.DeletedAtIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return CalendarEventOccurrenceRSVPAuthorization{}, calendarEntSafeError(err)
	}
	attendees, err := backend.tx.CalendarAttendee.Query().
		Where(calendarattendee.CalendarEventIDEQ(event.ID)).
		Order(ent.Asc(calendarattendee.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return CalendarEventOccurrenceRSVPAuthorization{}, calendarEntSafeError(err)
	}
	series, occurrence, attendeeErr := eventEntOccurrenceAttendeesForActor(attendees, scope.ActorID(), request.RecurrenceID)
	if attendeeErr != nil {
		return CalendarEventOccurrenceRSVPAuthorization{}, attendeeErr
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
		return CalendarEventOccurrenceRSVPAuthorization{}, calendarEntSafeError(err)
	}
	if len(exceptions) > 1 {
		return CalendarEventOccurrenceRSVPAuthorization{}, ErrEventInvalid
	}
	current, currentErr := eventEntCalendarEventInput(event)
	if currentErr != nil {
		return CalendarEventOccurrenceRSVPAuthorization{}, currentErr
	}
	seriesSnapshot, seriesErr := eventEntSeriesRSVPAttendeeSnapshot(series)
	if seriesErr != nil {
		return CalendarEventOccurrenceRSVPAuthorization{}, seriesErr
	}
	resolved := CalendarEventOccurrenceRSVPAuthorization{Current: current, SeriesAttendee: seriesSnapshot}
	if occurrence != nil {
		occurrenceSnapshot, occurrenceErr := eventEntOccurrenceRSVPAttendeeSnapshot(occurrence)
		if occurrenceErr != nil {
			return CalendarEventOccurrenceRSVPAuthorization{}, occurrenceErr
		}
		resolved.OccurrenceAttendee = &occurrenceSnapshot
	}
	if len(exceptions) == 1 {
		exception, exceptionErr := eventEntCalendarEventExceptionInput(exceptions[0])
		if exceptionErr != nil {
			return CalendarEventOccurrenceRSVPAuthorization{}, exceptionErr
		}
		resolved.Exception = &exception
	}
	return resolved, nil
}

func (backend *eventEntTransactionBackend) PersistCalendarEventOccurrenceRSVP(ctx context.Context, scope EventMutationScope, command CalendarEventOccurrenceRSVPPersistenceCommand) (CalendarEventOccurrenceRSVPPersistenceResult, error) {
	if backend == nil || backend.tx == nil || !command.valid() {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventInvalid
	}
	series := command.SeriesAttendee()
	requested := command.Attendee()
	if series.UserID == nil || *series.UserID != scope.ActorID() || requested.UserID == nil || *requested.UserID != scope.ActorID() {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventForbidden
	}
	if _, err := backend.activeOrganizationMemberForUpdate(ctx, scope.ActorID(), scope.OrganizationID()); err != nil {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, err
	}
	event, err := backend.tx.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(requested.EventID),
			calendarevent.OrganizationIDEQ(scope.OrganizationID()),
			calendarevent.DeletedAtIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, calendarEntSafeError(err)
	}
	attendees, err := backend.tx.CalendarAttendee.Query().
		Where(calendarattendee.CalendarEventIDEQ(event.ID)).
		Order(ent.Asc(calendarattendee.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, calendarEntSafeError(err)
	}
	sequence, sequenceOK := eventEntUint(event.Sequence)
	if !sequenceOK {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventInvalid
	}
	exceptions, err := backend.tx.CalendarEventException.Query().
		Where(
			calendareventexception.CalendarEventIDEQ(event.ID),
			calendareventexception.RecurrenceIDEQ(requested.RecurrenceID),
		).
		Order(ent.Asc(calendareventexception.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, calendarEntSafeError(err)
	}
	if len(exceptions) > 1 {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventInvalid
	}
	if len(exceptions) == 1 {
		value, valueOK := eventEntUint(exceptions[0].Sequence)
		if !valueOK {
			return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventInvalid
		}
		sequence = value
	}
	if sequence != command.EventSequence() {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventConflict
	}
	currentSeries, currentOccurrence, attendeeErr := eventEntOccurrenceAttendeesForActor(attendees, scope.ActorID(), requested.RecurrenceID)
	if attendeeErr != nil {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventConflict
	}
	durableSeries, durableSeriesErr := eventEntSeriesRSVPAttendeeSnapshot(currentSeries)
	if durableSeriesErr != nil || !reflect.DeepEqual(durableSeries, series) {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventConflict
	}
	prior, hadPrior := command.Prior()
	if hadPrior {
		durablePrior, durablePriorErr := eventEntOccurrenceRSVPAttendeeSnapshot(currentOccurrence)
		if durablePriorErr != nil || !reflect.DeepEqual(durablePrior, prior) {
			return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventConflict
		}
	} else if !validCalendarEventOccurrenceRSVPAttendeeForSeries(durableSeries, requested) {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventConflict
	}
	appliedAttendeeID := requested.AttendeeID
	if hadPrior {
		appliedAttendeeID = currentOccurrence.ID
	}
	if err := backend.persistCalendarOccurrenceRSVPAttendee(ctx, event.ID, currentOccurrence, prior, hadPrior, requested, command.ExpectedVersion(), command.Reply()); err != nil {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, err
	}
	reply, hasReply := command.Reply()
	if hasReply {
		recurrenceID := requested.RecurrenceID
		if err := backend.persistCalendarReplyDelivery(ctx, event.OrganizationID, event.ID, appliedAttendeeID, command.EventSequence(), requested.Version, &recurrenceID, requested.Status); err != nil {
			return CalendarEventOccurrenceRSVPPersistenceResult{}, err
		}
	}
	if err := backend.AppendOutbox(ctx, scope, command.Outbox()); err != nil {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, err
	}
	return NewCalendarEventOccurrenceRSVPPersistenceResult(requested, command.EventSequence(), command.Outbox(), eventEntOccurrenceReplyPointer(reply, hasReply))
}

func eventEntSeriesAttendeeForActor(attendees []*ent.CalendarAttendee, actorID uuid.UUID) (*ent.CalendarAttendee, bool) {
	var matched *ent.CalendarAttendee
	for _, attendee := range attendees {
		if attendee.UserID == nil || *attendee.UserID != actorID || attendee.RecurrenceID != nil || attendee.RemovedAt != nil {
			continue
		}
		if matched != nil {
			return nil, false
		}
		matched = attendee
	}
	return matched, matched != nil
}

func eventEntOccurrenceAttendeesForActor(attendees []*ent.CalendarAttendee, actorID uuid.UUID, recurrenceID string) (*ent.CalendarAttendee, *ent.CalendarAttendee, error) {
	series, found := eventEntSeriesAttendeeForActor(attendees, actorID)
	if !found {
		return nil, nil, ErrEventNotFound
	}
	var occurrence *ent.CalendarAttendee
	for _, attendee := range attendees {
		if attendee.UserID == nil || *attendee.UserID != actorID || attendee.RemovedAt != nil || attendee.RecurrenceID == nil || *attendee.RecurrenceID != recurrenceID {
			continue
		}
		if occurrence != nil {
			return nil, nil, ErrEventInvalid
		}
		occurrence = attendee
	}
	return series, occurrence, nil
}

func eventEntSeriesRSVPAttendeeSnapshot(attendee *ent.CalendarAttendee) (CalendarEventSeriesRSVPAttendeeSnapshot, error) {
	if attendee == nil || attendee.RecurrenceID != nil {
		return CalendarEventSeriesRSVPAttendeeSnapshot{}, ErrEventInvalid
	}
	version, ok := eventEntUint(attendee.Version)
	if !ok {
		return CalendarEventSeriesRSVPAttendeeSnapshot{}, ErrEventInvalid
	}
	return CalendarEventSeriesRSVPAttendeeSnapshot{
		AttendeeID:  attendee.ID,
		EventID:     attendee.CalendarEventID,
		UserID:      eventEntUUIDPointer(attendee.UserID),
		Email:       attendee.NormalizedEmail,
		DisplayName: eventEntStringValue(attendee.DisplayName),
		Role:        CalendarEventAttendeeRole(attendee.Role),
		Status:      RSVPStatus(attendee.RsvpStatus),
		RespondedAt: eventEntTimePointer(attendee.RespondedAt),
		Version:     version,
		RemovedAt:   eventEntTimePointer(attendee.RemovedAt),
	}, nil
}

func eventEntSeriesRSVPImmutableIdentityMatches(durable, requested CalendarEventSeriesRSVPAttendeeSnapshot) bool {
	return durable.AttendeeID == requested.AttendeeID &&
		durable.EventID == requested.EventID &&
		durable.UserID != nil && requested.UserID != nil && *durable.UserID == *requested.UserID &&
		durable.Email == requested.Email &&
		durable.DisplayName == requested.DisplayName &&
		durable.Role == requested.Role &&
		durable.RemovedAt == nil && requested.RemovedAt == nil
}

func eventEntOccurrenceRSVPAttendeeSnapshot(attendee *ent.CalendarAttendee) (CalendarEventOccurrenceRSVPAttendeeSnapshot, error) {
	if attendee == nil || attendee.RecurrenceID == nil {
		return CalendarEventOccurrenceRSVPAttendeeSnapshot{}, ErrEventInvalid
	}
	version, ok := eventEntUint(attendee.Version)
	if !ok {
		return CalendarEventOccurrenceRSVPAttendeeSnapshot{}, ErrEventInvalid
	}
	return CalendarEventOccurrenceRSVPAttendeeSnapshot{
		AttendeeID:   attendee.ID,
		EventID:      attendee.CalendarEventID,
		UserID:       eventEntUUIDPointer(attendee.UserID),
		Email:        attendee.NormalizedEmail,
		DisplayName:  eventEntStringValue(attendee.DisplayName),
		Role:         CalendarEventAttendeeRole(attendee.Role),
		Status:       RSVPStatus(attendee.RsvpStatus),
		RecurrenceID: *attendee.RecurrenceID,
		RespondedAt:  eventEntTimePointer(attendee.RespondedAt),
		Version:      version,
		RemovedAt:    eventEntTimePointer(attendee.RemovedAt),
	}, nil
}

func (backend *eventEntTransactionBackend) persistCalendarOccurrenceRSVPAttendee(ctx context.Context, eventID uuid.UUID, existing *ent.CalendarAttendee, prior CalendarEventOccurrenceRSVPAttendeeSnapshot, hadPrior bool, requested CalendarEventOccurrenceRSVPAttendeeSnapshot, expectedVersion uint, reply CalendarEventOccurrenceRSVPReplyDeliveryDescriptor, hasReply bool) error {
	expected, expectedOK := eventEntInt(expectedVersion)
	version, versionOK := eventEntInt(requested.Version)
	if !expectedOK || !versionOK || requested.RespondedAt == nil {
		return ErrEventInvalid
	}
	if !hadPrior {
		if existing != nil {
			return ErrEventConflict
		}
		create := backend.tx.CalendarAttendee.Create().
			SetID(requested.AttendeeID).
			SetCalendarEventID(eventID).
			SetNormalizedEmail(requested.Email).
			SetRole(calendarattendee.Role(requested.Role)).
			SetRsvpStatus(calendarattendee.RsvpStatus(requested.Status)).
			SetRecurrenceID(requested.RecurrenceID).
			SetRespondedAt(requested.RespondedAt.UTC()).
			SetVersion(version)
		if requested.UserID != nil {
			create.SetUserID(*requested.UserID)
		}
		if requested.DisplayName != "" {
			create.SetDisplayName(requested.DisplayName)
		}
	if hasReply {
		create.SetReplyDeliveryStatus(calendarattendee.ReplyDeliveryStatusPending)
	} else {
		create.SetReplyDeliveryStatus(calendarattendee.ReplyDeliveryStatusNotRequired)
		}
		if _, err := create.Save(ctx); err != nil {
			return calendarEntSafeError(err)
		}
		return nil
	}
	if existing == nil || existing.ID != prior.AttendeeID || existing.ID != requested.AttendeeID {
		return ErrEventConflict
	}
	update := backend.tx.CalendarAttendee.UpdateOneID(existing.ID).
		Where(
			calendarattendee.CalendarEventIDEQ(eventID),
			calendarattendee.RecurrenceIDEQ(requested.RecurrenceID),
			calendarattendee.VersionEQ(expected),
			calendarattendee.RemovedAtIsNil(),
		).
		SetRsvpStatus(calendarattendee.RsvpStatus(requested.Status)).
		SetRespondedAt(requested.RespondedAt.UTC()).
		SetVersion(version)
	if hasReply {
		update.SetReplyDeliveryStatus(calendarattendee.ReplyDeliveryStatusPending)
	} else {
		update.SetReplyDeliveryStatus(calendarattendee.ReplyDeliveryStatusNotRequired)
	}
	if err := update.Exec(ctx); err != nil {
		return calendarEntCASFailure(err)
	}
	return nil
}

func (backend *eventEntTransactionBackend) persistCalendarReplyDelivery(ctx context.Context, organizationID, eventID, attendeeID uuid.UUID, eventSequence, attendeeVersion uint, recurrenceID *string, status RSVPStatus) error {
	sequence, sequenceOK := eventEntInt(eventSequence)
	version, versionOK := eventEntInt(attendeeVersion)
	if organizationID == uuid.Nil || !sequenceOK || !versionOK {
		return ErrEventInvalid
	}
	databaseNow, err := backend.calendarInboundMessageDatabaseClock(ctx)
	if err != nil {
		return err
	}
	create := backend.tx.CalendarReplyDelivery.Create().
		SetOrganizationID(organizationID).
		SetCalendarEventID(eventID).
		SetCalendarAttendeeID(attendeeID).
		SetEventSequence(sequence).
		SetAttendeeVersion(version).
		SetRsvpStatus(calendarreplydelivery.RsvpStatus(status)).
		SetNextAttemptAt(databaseNow)
	if recurrenceID != nil {
		create.SetRecurrenceID(*recurrenceID)
	}
	if _, err := create.Save(ctx); err != nil {
		return calendarEntSafeError(err)
	}
	return nil
}

func eventEntSeriesReplyPointer(reply CalendarEventSeriesRSVPReplyDeliveryDescriptor, found bool) *CalendarEventSeriesRSVPReplyDeliveryDescriptor {
	if !found {
		return nil
	}
	return &reply
}

func eventEntOccurrenceReplyPointer(reply CalendarEventOccurrenceRSVPReplyDeliveryDescriptor, found bool) *CalendarEventOccurrenceRSVPReplyDeliveryDescriptor {
	if !found {
		return nil
	}
	return &reply
}
