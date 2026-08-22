package calendar

import (
	"bytes"
	"context"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendareventexception"
	"github.com/valerii/hub/backend/ent/calendarinboundreplyreceipt"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/platform/outbox"
)

var _ CalendarInboundReplyTransactionBackend = (*eventEntTransactionBackend)(nil)

func (backend *eventEntTransactionBackend) ResolveCalendarInboundReplyOwner(ctx context.Context, ownerID uuid.UUID) (uuid.UUID, bool, error) {
	if backend == nil || backend.tx == nil || ownerID == uuid.Nil {
		return uuid.Nil, false, ErrEventInvalid
	}
	organizationRow, err := backend.tx.Organization.Query().
		Where(organization.HasUsersWith(
			user.IDEQ(ownerID),
			user.StatusEQ(user.StatusActive),
		)).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, calendarEntSafeError(err)
	}
	if _, err := backend.tx.User.Query().
		Where(
			user.IDEQ(ownerID),
			user.StatusEQ(user.StatusActive),
		).
		ForUpdate().
		Only(ctx); ent.IsNotFound(err) {
		return uuid.Nil, false, nil
	} else if err != nil {
		return uuid.Nil, false, calendarEntSafeError(err)
	}
	return organizationRow.ID, true, nil
}

func (backend *eventEntTransactionBackend) ResolveCalendarInboundReply(ctx context.Context, scope CalendarInboundReplyScope, request CalendarInboundReplyAuthorizationRequest) (CalendarInboundReplyAuthorization, error) {
	if backend == nil || backend.tx == nil || scope.OwnerID() == uuid.Nil || scope.OrganizationID() == uuid.Nil {
		return CalendarInboundReplyAuthorization{}, ErrEventInvalid
	}
	organizationID, found, ownerErr := backend.ResolveCalendarInboundReplyOwner(ctx, scope.OwnerID())
	if ownerErr != nil {
		return CalendarInboundReplyAuthorization{}, ownerErr
	}
	if !found {
		return CalendarInboundReplyAuthorization{}, nil
	}
	if organizationID != scope.OrganizationID() {
		return CalendarInboundReplyAuthorization{}, ErrEventInvalid
	}
	event, err := backend.tx.CalendarEvent.Query().
		Where(
			calendarevent.OrganizationIDEQ(scope.OrganizationID()),
			calendarevent.CalendarOwnerIDEQ(scope.OwnerID()),
			calendarevent.UIDEQ(request.UID),
			calendarevent.OrganizerIDEQ(scope.OwnerID()),
			calendarevent.OrganizerEmailEQ(scope.organizerEmail),
			calendarevent.DeletedAtIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return CalendarInboundReplyAuthorization{}, nil
	}
	if err != nil {
		return CalendarInboundReplyAuthorization{}, calendarEntSafeError(err)
	}
	attendees, err := backend.tx.CalendarAttendee.Query().
		Where(
			calendarattendee.CalendarEventIDEQ(event.ID),
			calendarattendee.NormalizedEmailEQ(request.Sender),
		).
		Order(ent.Asc(calendarattendee.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return CalendarInboundReplyAuthorization{}, calendarEntSafeError(err)
	}
	target, found, targetErr := eventEntInboundReplyAttendee(attendees, request.HasRecurrenceID, request.RecurrenceID)
	if targetErr != nil {
		return CalendarInboundReplyAuthorization{}, targetErr
	}
	if !found {
		return CalendarInboundReplyAuthorization{}, nil
	}
	var exception *CalendarEventExceptionCommandInput
	if request.HasRecurrenceID {
		exceptions, exceptionErr := backend.tx.CalendarEventException.Query().
			Where(
				calendareventexception.CalendarEventIDEQ(event.ID),
				calendareventexception.RecurrenceIDEQ(request.RecurrenceID),
			).
			Order(ent.Asc(calendareventexception.FieldID)).
			ForUpdate().
			All(ctx)
		if exceptionErr != nil {
			return CalendarInboundReplyAuthorization{}, calendarEntSafeError(exceptionErr)
		}
		if len(exceptions) > 1 {
			return CalendarInboundReplyAuthorization{}, ErrEventInvalid
		}
		if len(exceptions) == 1 {
			value, valueErr := eventEntCalendarEventExceptionInput(exceptions[0])
			if valueErr != nil {
				return CalendarInboundReplyAuthorization{}, valueErr
			}
			exception = &value
		}
	}
	requestSequence, requestSequenceOK := eventEntInt(request.EventSequence)
	if !requestSequenceOK {
		return CalendarInboundReplyAuthorization{}, ErrEventInvalid
	}
	receipts, err := backend.tx.CalendarInboundReplyReceipt.Query().
		Where(
			calendarinboundreplyreceipt.CalendarAttendeeIDEQ(target.ID),
			calendarinboundreplyreceipt.EventSequenceEQ(requestSequence),
		).
		Order(ent.Asc(calendarinboundreplyreceipt.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return CalendarInboundReplyAuthorization{}, calendarEntSafeError(err)
	}
	current, currentErr := eventEntCalendarEventInput(event)
	if currentErr != nil {
		return CalendarInboundReplyAuthorization{}, currentErr
	}
	targetSnapshot, targetSnapshotErr := eventEntInboundReplyAttendeeSnapshot(target)
	if targetSnapshotErr != nil {
		return CalendarInboundReplyAuthorization{}, targetSnapshotErr
	}
	keyHash := request.Receipt.PayloadHash()
	existing := false
	for _, receipt := range receipts {
		if receipt.PayloadHashVersion == request.Receipt.CanonicalVersion() && bytes.Equal(receipt.PayloadHash, keyHash[:]) {
			existing = true
			break
		}
	}
	if len(receipts) > int(^uint8(0)) {
		return CalendarInboundReplyAuthorization{}, ErrEventInvalid
	}
	return CalendarInboundReplyAuthorization{
		Matched:              true,
		Current:              current,
		Target:               targetSnapshot,
		Exception:            exception,
		ExistingReceipt:      existing,
		DistinctReceiptCount: uint8(len(receipts)),
	}, nil
}

func (backend *eventEntTransactionBackend) PersistCalendarInboundReply(ctx context.Context, scope CalendarInboundReplyScope, command CalendarInboundReplyPersistenceCommand) (CalendarInboundReplyPersistenceResult, error) {
	if backend == nil || backend.tx == nil || scope.OwnerID() == uuid.Nil || scope.OrganizationID() == uuid.Nil {
		return CalendarInboundReplyPersistenceResult{}, ErrEventInvalid
	}
	organizationID, found, ownerErr := backend.ResolveCalendarInboundReplyOwner(ctx, scope.OwnerID())
	if ownerErr != nil {
		return CalendarInboundReplyPersistenceResult{}, ownerErr
	}
	if !found {
		return CalendarInboundReplyPersistenceResult{}, ErrEventNotFound
	}
	if organizationID != scope.OrganizationID() {
		return CalendarInboundReplyPersistenceResult{}, ErrEventInvalid
	}
	prior := command.Prior()
	event, err := backend.tx.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(prior.EventID),
			calendarevent.OrganizationIDEQ(scope.OrganizationID()),
			calendarevent.CalendarOwnerIDEQ(scope.OwnerID()),
			calendarevent.OrganizerIDEQ(scope.OwnerID()),
			calendarevent.OrganizerEmailEQ(scope.organizerEmail),
			calendarevent.DeletedAtIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return CalendarInboundReplyPersistenceResult{}, calendarEntSafeError(err)
	}
	attendees, err := backend.tx.CalendarAttendee.Query().
		Where(calendarattendee.CalendarEventIDEQ(event.ID)).
		Order(ent.Asc(calendarattendee.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return CalendarInboundReplyPersistenceResult{}, calendarEntSafeError(err)
	}
	target, found := eventEntInboundReplyAttendeeByID(attendees, prior.AttendeeID)
	if !found {
		return CalendarInboundReplyPersistenceResult{}, ErrEventConflict
	}
	durablePrior, durablePriorErr := eventEntInboundReplyAttendeeSnapshot(target)
	if durablePriorErr != nil {
		return CalendarInboundReplyPersistenceResult{}, durablePriorErr
	}
	if !eventEntInboundReplyIdentityMatches(prior, durablePrior) {
		return CalendarInboundReplyPersistenceResult{}, ErrEventConflict
	}
	effectiveSequence, sequenceErr := backend.calendarInboundReplyEffectiveSequence(ctx, event, prior.RecurrenceID)
	if sequenceErr != nil {
		return CalendarInboundReplyPersistenceResult{}, sequenceErr
	}
	if effectiveSequence != command.EventSequence() {
		return CalendarInboundReplyPersistenceResult{}, ErrEventConflict
	}
	effectiveSequenceInt, effectiveSequenceOK := eventEntInt(effectiveSequence)
	if !effectiveSequenceOK {
		return CalendarInboundReplyPersistenceResult{}, ErrEventInvalid
	}
	receipts, err := backend.tx.CalendarInboundReplyReceipt.Query().
		Where(
			calendarinboundreplyreceipt.CalendarAttendeeIDEQ(target.ID),
			calendarinboundreplyreceipt.EventSequenceEQ(effectiveSequenceInt),
		).
		Order(ent.Asc(calendarinboundreplyreceipt.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return CalendarInboundReplyPersistenceResult{}, calendarEntSafeError(err)
	}
	keyHash := command.Receipt().PayloadHash()
	for _, receipt := range receipts {
		if receipt.PayloadHashVersion == command.Receipt().CanonicalVersion() && bytes.Equal(receipt.PayloadHash, keyHash[:]) {
			return CalendarInboundReplyPersistenceResult{}, ErrEventConflict
		}
	}

	resultAttendee := durablePrior
	if command.ReceiptOutcome() == CalendarInboundReplyReceiptApplied {
		updated, updatedFound := command.Updated()
		if !updatedFound || updated.RespondedAt == nil {
			return CalendarInboundReplyPersistenceResult{}, ErrEventInvalid
		}
		expectedVersion, expectedOK := eventEntInt(command.ExpectedVersion())
		updatedVersion, updatedOK := eventEntInt(updated.Version)
		if !expectedOK || !updatedOK {
			return CalendarInboundReplyPersistenceResult{}, ErrEventInvalid
		}
		update := backend.tx.CalendarAttendee.UpdateOneID(target.ID).
			Where(
				calendarattendee.CalendarEventIDEQ(event.ID),
				calendarattendee.VersionEQ(expectedVersion),
			).
			SetRsvpStatus(calendarattendee.RsvpStatus(updated.Status)).
			SetRespondedAt(updated.RespondedAt.UTC()).
			SetVersion(updatedVersion)
		if err := update.Exec(ctx); err != nil {
			return CalendarInboundReplyPersistenceResult{}, calendarEntCASFailure(err)
		}
		resultAttendee = updated
	}
	if err := backend.persistCalendarInboundReplyReceipt(ctx, target.ID, effectiveSequence, command.Receipt(), command.ReceiptOutcome()); err != nil {
		return CalendarInboundReplyPersistenceResult{}, err
	}
	if descriptor, hasOutbox := command.Outbox(); hasOutbox {
		if err := backend.appendCalendarInboundReplyOutbox(ctx, descriptor); err != nil {
			return CalendarInboundReplyPersistenceResult{}, err
		}
	}
	descriptor, hasOutbox := command.Outbox()
	return NewCalendarInboundReplyPersistenceResult(resultAttendee, effectiveSequence, command.Receipt(), command.ReceiptOutcome(), eventEntInboundOutboxPointer(descriptor, hasOutbox))
}

func eventEntInboundReplyAttendee(attendees []*ent.CalendarAttendee, hasRecurrenceID bool, recurrenceID string) (*ent.CalendarAttendee, bool, error) {
	var active *ent.CalendarAttendee
	var removed *ent.CalendarAttendee
	for _, attendee := range attendees {
		matchesRecurrence := attendee.RecurrenceID == nil
		if hasRecurrenceID {
			matchesRecurrence = attendee.RecurrenceID != nil && *attendee.RecurrenceID == recurrenceID
		}
		if !matchesRecurrence {
			continue
		}
		if attendee.RemovedAt == nil {
			if active != nil {
				return nil, false, ErrEventInvalid
			}
			active = attendee
			continue
		}
		if removed != nil {
			return nil, false, nil
		}
		removed = attendee
	}
	if active != nil {
		return active, true, nil
	}
	return removed, removed != nil, nil
}

func eventEntInboundReplyAttendeeByID(attendees []*ent.CalendarAttendee, attendeeID uuid.UUID) (*ent.CalendarAttendee, bool) {
	for _, attendee := range attendees {
		if attendee.ID == attendeeID {
			return attendee, true
		}
	}
	return nil, false
}

func eventEntInboundReplyAttendeeSnapshot(attendee *ent.CalendarAttendee) (CalendarInboundReplyAttendeeSnapshot, error) {
	if attendee == nil {
		return CalendarInboundReplyAttendeeSnapshot{}, ErrEventInvalid
	}
	version, ok := eventEntUint(attendee.Version)
	if !ok {
		return CalendarInboundReplyAttendeeSnapshot{}, ErrEventInvalid
	}
	return CalendarInboundReplyAttendeeSnapshot{
		AttendeeID:   attendee.ID,
		EventID:      attendee.CalendarEventID,
		Email:        attendee.NormalizedEmail,
		Status:       RSVPStatus(attendee.RsvpStatus),
		RecurrenceID: eventEntStringPointer(attendee.RecurrenceID),
		RespondedAt:  eventEntTimePointer(attendee.RespondedAt),
		Version:      version,
		RemovedAt:    eventEntTimePointer(attendee.RemovedAt),
	}, nil
}

func eventEntInboundReplyIdentityMatches(expected, actual CalendarInboundReplyAttendeeSnapshot) bool {
	if expected.AttendeeID != actual.AttendeeID || expected.EventID != actual.EventID || expected.Email != actual.Email || expected.Version != actual.Version || expected.Status != actual.Status {
		return false
	}
	if (expected.RecurrenceID == nil) != (actual.RecurrenceID == nil) {
		return false
	}
	return expected.RecurrenceID == nil || *expected.RecurrenceID == *actual.RecurrenceID
}

func (backend *eventEntTransactionBackend) calendarInboundReplyEffectiveSequence(ctx context.Context, event *ent.CalendarEvent, recurrenceID *string) (uint, error) {
	if event == nil {
		return 0, ErrEventInvalid
	}
	sequence, ok := eventEntUint(event.Sequence)
	if !ok {
		return 0, ErrEventInvalid
	}
	if recurrenceID == nil {
		return sequence, nil
	}
	exceptions, err := backend.tx.CalendarEventException.Query().
		Where(
			calendareventexception.CalendarEventIDEQ(event.ID),
			calendareventexception.RecurrenceIDEQ(*recurrenceID),
		).
		Order(ent.Asc(calendareventexception.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return 0, calendarEntSafeError(err)
	}
	if len(exceptions) > 1 {
		return 0, ErrEventInvalid
	}
	if len(exceptions) == 0 {
		return sequence, nil
	}
	value, valueOK := eventEntUint(exceptions[0].Sequence)
	if !valueOK || value > sequence {
		return 0, ErrEventInvalid
	}
	return value, nil
}

func (backend *eventEntTransactionBackend) persistCalendarInboundReplyReceipt(ctx context.Context, attendeeID uuid.UUID, sequence uint, key CalendarInboundReplyReceiptKey, outcome CalendarInboundReplyReceiptOutcome) error {
	eventSequence, sequenceOK := eventEntInt(sequence)
	if !sequenceOK {
		return ErrEventInvalid
	}
	hash := key.PayloadHash()
	if _, err := backend.tx.CalendarInboundReplyReceipt.Create().
		SetCalendarAttendeeID(attendeeID).
		SetEventSequence(eventSequence).
		SetPayloadHashVersion(key.CanonicalVersion()).
		SetPayloadHash(hash[:]).
		SetOutcome(calendarinboundreplyreceipt.Outcome(outcome)).
		Save(ctx); err != nil {
		return calendarEntSafeError(err)
	}
	return nil
}

func (backend *eventEntTransactionBackend) appendCalendarInboundReplyOutbox(ctx context.Context, descriptor EventOutboxDescriptor) error {
	envelope := descriptor.Envelope()
	return calendarEntSafeError(outbox.NewPublisher().PublishInTx(ctx, backend.tx, envelope.Subject(), envelope))
}

func eventEntInboundOutboxPointer(descriptor EventOutboxDescriptor, found bool) *EventOutboxDescriptor {
	if !found {
		return nil
	}
	return &descriptor
}
