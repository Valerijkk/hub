package mail

import (
	"context"
	"time"

	"github.com/google/uuid"
	ical "github.com/valerii/hub/backend/internal/ical"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendareventexception"
	"github.com/valerii/hub/backend/ent/calendarinvitedelivery"
	"github.com/valerii/hub/backend/ent/calendarreplydelivery"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

const mailboxEntCalendarDeliveryMaximumClaimGeneration int64 = 2147483647

// ClaimCalendarMailDelivery returns an external capability only for an
// unbound master delivery that this transaction has atomically fenced. Calendar
// facts and the canonical iTIP snapshot are validated before the state CAS so
// a malformed projection cannot strand a row in submitting.
func (repository *MailboxEntRepository) ClaimCalendarMailDelivery(ctx context.Context, deliveryID CalendarMailDeliveryID) (CalendarMailDeliveryClaimResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || !validCalendarMailDeliveryID(deliveryID) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	result := CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		var claimErr error
		switch deliveryID.Kind() {
		case CalendarMailDeliveryKindInvite:
			result, claimErr = mailboxEntClaimCalendarInviteDelivery(ctx, transaction, deliveryID)
		case CalendarMailDeliveryKindReply:
			result, claimErr = mailboxEntClaimCalendarReplyDelivery(ctx, transaction, deliveryID)
		default:
			claimErr = newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
		}
		return claimErr
	})
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, mailboxEntCalendarDeliveryRepositoryError(err)
	}
	if !validCalendarMailDeliveryClaimResult(result) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return result, nil
}

func mailboxEntClaimCalendarInviteDelivery(ctx context.Context, transaction *ent.Tx, deliveryID CalendarMailDeliveryID) (CalendarMailDeliveryClaimResult, error) {
	if ctx == nil || transaction == nil || deliveryID.Kind() != CalendarMailDeliveryKindInvite || !validCalendarMailDeliveryID(deliveryID) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	locator, found, err := mailboxEntReadCalendarInviteDeliveryLocator(ctx, transaction, deliveryID)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	if !found {
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
	}
	facts, err := mailboxEntLockCalendarInviteDeliveryFacts(ctx, transaction, locator)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	delivery, found, err := mailboxEntCalendarInviteDeliveryForUpdate(ctx, transaction, locator)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	if !found {
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
	}
	if !mailboxEntCalendarInviteDeliveryLocatorMatches(locator, delivery) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
	if err != nil || !mailboxEntValidTime(databaseNow) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	switch CalendarMailDeliveryState(delivery.State) {
	case CalendarMailDeliveryStateUnknown:
		if !mailboxEntCalendarInviteDeliveryUnknownBoundValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if delivery.NextAttemptAt.After(databaseNow) {
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
		}
		return mailboxEntObserveCalendarInviteDeliveryBoundSubmission(ctx, transaction, delivery, deliveryID, databaseNow)
	case CalendarMailDeliveryStateSent, CalendarMailDeliveryStateFailedPermanent:
		if !mailboxEntCalendarInviteDeliveryTerminalValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
	case CalendarMailDeliveryStateSuperseded:
		if !mailboxEntCalendarInviteDeliveryTerminalValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeSuperseded}, nil
	case CalendarMailDeliveryStatePending:
		if !mailboxEntCalendarInviteDeliveryPendingValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if delivery.NextAttemptAt.After(databaseNow) {
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
		}
		return mailboxEntClaimDueCalendarInviteDelivery(ctx, transaction, delivery, deliveryID, facts, databaseNow, false)
	case CalendarMailDeliveryStateSubmitting:
		if !mailboxEntCalendarInviteDeliverySubmittingValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if delivery.LeaseUntil.After(databaseNow) {
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeBusy}, nil
		}
		return mailboxEntClaimDueCalendarInviteDelivery(ctx, transaction, delivery, deliveryID, facts, databaseNow, true)
	default:
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
}

func mailboxEntClaimCalendarReplyDelivery(ctx context.Context, transaction *ent.Tx, deliveryID CalendarMailDeliveryID) (CalendarMailDeliveryClaimResult, error) {
	if ctx == nil || transaction == nil || deliveryID.Kind() != CalendarMailDeliveryKindReply || !validCalendarMailDeliveryID(deliveryID) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	locator, found, err := mailboxEntReadCalendarReplyDeliveryLocator(ctx, transaction, deliveryID)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	if !found {
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
	}
	facts, err := mailboxEntLockCalendarReplyDeliveryFacts(ctx, transaction, locator)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	delivery, found, err := mailboxEntCalendarReplyDeliveryForUpdate(ctx, transaction, locator)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	if !found {
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
	}
	if !mailboxEntCalendarReplyDeliveryClaimLocatorMatches(locator, delivery) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
	if err != nil || !mailboxEntValidTime(databaseNow) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	switch CalendarMailDeliveryState(delivery.State) {
	case CalendarMailDeliveryStateUnknown:
		if !mailboxEntCalendarReplyDeliveryUnknownBoundValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if delivery.NextAttemptAt.After(databaseNow) {
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
		}
		return mailboxEntObserveCalendarReplyDeliveryBoundSubmission(ctx, transaction, delivery, deliveryID, databaseNow)
	case CalendarMailDeliveryStateSent, CalendarMailDeliveryStateFailedPermanent:
		if !mailboxEntCalendarReplyDeliveryTerminalValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
	case CalendarMailDeliveryStateSuperseded:
		if !mailboxEntCalendarReplyDeliveryTerminalValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeSuperseded}, nil
	case CalendarMailDeliveryStatePending:
		if mailboxEntCalendarReplyDeliveryInitialPendingUnscheduledValid(delivery) {
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
		}
		if !mailboxEntCalendarReplyDeliveryScheduledPendingValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if delivery.NextAttemptAt.After(databaseNow) {
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
		}
		return mailboxEntClaimDueCalendarReplyDelivery(ctx, transaction, delivery, deliveryID, facts, databaseNow, false)
	case CalendarMailDeliveryStateSubmitting:
		if !mailboxEntCalendarReplyDeliverySubmittingValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if delivery.LeaseUntil.After(databaseNow) {
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeBusy}, nil
		}
		return mailboxEntClaimDueCalendarReplyDelivery(ctx, transaction, delivery, deliveryID, facts, databaseNow, true)
	default:
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
}

type mailboxEntCalendarDeliveryClaimSnapshot struct {
	organizationID  uuid.UUID
	eventID         uuid.UUID
	attendeeID      uuid.UUID
	organizerID     uuid.UUID
	accountID       uuid.UUID
	senderActorID   uuid.UUID
	fromIdentityID  uuid.UUID
	recipient       EngineAddress
	method          CalendarMailDeliveryMethod
	eventSequence   int
	attendeeVersion int
	recurrenceID    string
	message         ical.ITIPMessage
}

// mailboxEntCalendarInviteDeliveryLocator is an immutable navigation hint read
// without a row lock. Every value is revalidated after canonical Calendar
// locks and the final delivery lock have been acquired.
type mailboxEntCalendarInviteDeliveryLocator struct {
	id             uuid.UUID
	organizationID uuid.UUID
	eventID        uuid.UUID
	attendeeID     uuid.UUID
	kind           calendarinvitedelivery.Kind
	recurrenceID   *string
}

type mailboxEntCalendarInviteDeliveryLockedFacts struct {
	event     *ent.CalendarEvent
	attendees []*ent.CalendarAttendee
	exception *ent.CalendarEventException
}

type mailboxEntCalendarReplyDeliveryLockedFacts struct {
	event     *ent.CalendarEvent
	attendee  *ent.CalendarAttendee
	exception *ent.CalendarEventException
}

func mailboxEntReadCalendarReplyDeliveryLocator(ctx context.Context, transaction *ent.Tx, deliveryID CalendarMailDeliveryID) (mailboxEntCalendarReplyDeliveryLocator, bool, error) {
	if ctx == nil || transaction == nil || deliveryID.Kind() != CalendarMailDeliveryKindReply || !validCalendarMailDeliveryID(deliveryID) {
		return mailboxEntCalendarReplyDeliveryLocator{}, false, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	row, err := transaction.CalendarReplyDelivery.Query().
		Where(calendarreplydelivery.IDEQ(deliveryID.UUID())).
		Only(ctx)
	if ent.IsNotFound(err) {
		return mailboxEntCalendarReplyDeliveryLocator{}, false, nil
	}
	if err != nil || row == nil || row.ID != deliveryID.UUID() {
		return mailboxEntCalendarReplyDeliveryLocator{}, false, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	locator := mailboxEntCalendarReplyDeliveryLocator{id: row.ID, organizationID: row.OrganizationID, eventID: row.CalendarEventID, attendeeID: row.CalendarAttendeeID}
	if row.RecurrenceID != nil {
		recurrenceID := *row.RecurrenceID
		locator.recurrenceID = &recurrenceID
	}
	if !mailboxEntCalendarReplyDeliveryLocatorValid(locator) {
		return mailboxEntCalendarReplyDeliveryLocator{}, false, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return locator, true, nil
}

func mailboxEntLockCalendarReplyDeliveryFacts(ctx context.Context, transaction *ent.Tx, locator mailboxEntCalendarReplyDeliveryLocator) (mailboxEntCalendarReplyDeliveryLockedFacts, error) {
	if ctx == nil || transaction == nil || !mailboxEntCalendarReplyDeliveryLocatorValid(locator) {
		return mailboxEntCalendarReplyDeliveryLockedFacts{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	event, err := transaction.CalendarEvent.Query().
		Where(calendarevent.IDEQ(locator.eventID), calendarevent.OrganizationIDEQ(locator.organizationID)).
		ForUpdate().
		Only(ctx)
	if err != nil || event == nil || event.ID != locator.eventID || event.OrganizationID != locator.organizationID {
		return mailboxEntCalendarReplyDeliveryLockedFacts{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	attendeeRecurrencePredicate := calendarattendee.RecurrenceIDIsNil()
	if locator.recurrenceID != nil {
		attendeeRecurrencePredicate = calendarattendee.RecurrenceIDEQ(*locator.recurrenceID)
	}
	attendee, err := transaction.CalendarAttendee.Query().
		Where(calendarattendee.IDEQ(locator.attendeeID), calendarattendee.CalendarEventIDEQ(event.ID), attendeeRecurrencePredicate).
		ForUpdate().
		Only(ctx)
	if err != nil || attendee == nil || attendee.ID != locator.attendeeID || attendee.CalendarEventID != event.ID || !mailboxEntCalendarDeliveryRecurrencePointersEqual(attendee.RecurrenceID, locator.recurrenceID) {
		return mailboxEntCalendarReplyDeliveryLockedFacts{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	facts := mailboxEntCalendarReplyDeliveryLockedFacts{event: event, attendee: attendee}
	if locator.recurrenceID == nil {
		return facts, nil
	}
	exception, err := transaction.CalendarEventException.Query().
		Where(
			calendareventexception.CalendarEventIDEQ(event.ID),
			calendareventexception.RecurrenceIDEQ(*locator.recurrenceID),
		).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return facts, nil
	}
	if err != nil || exception == nil || exception.CalendarEventID != event.ID || exception.RecurrenceID != *locator.recurrenceID {
		return mailboxEntCalendarReplyDeliveryLockedFacts{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	facts.exception = exception
	return facts, nil
}

func mailboxEntCalendarReplyDeliveryForUpdate(ctx context.Context, transaction *ent.Tx, locator mailboxEntCalendarReplyDeliveryLocator) (*ent.CalendarReplyDelivery, bool, error) {
	if ctx == nil || transaction == nil || !mailboxEntCalendarReplyDeliveryLocatorValid(locator) {
		return nil, false, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	deliveryRecurrencePredicate := calendarreplydelivery.RecurrenceIDIsNil()
	if locator.recurrenceID != nil {
		deliveryRecurrencePredicate = calendarreplydelivery.RecurrenceIDEQ(*locator.recurrenceID)
	}
	delivery, err := transaction.CalendarReplyDelivery.Query().
		Where(calendarreplydelivery.IDEQ(locator.id), deliveryRecurrencePredicate).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, false, nil
	}
	if err != nil || delivery == nil || delivery.ID != locator.id {
		return nil, false, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return delivery, true, nil
}

func mailboxEntCalendarReplyDeliveryClaimLocatorMatches(locator mailboxEntCalendarReplyDeliveryLocator, delivery *ent.CalendarReplyDelivery) bool {
	return delivery != nil && delivery.ID == locator.id && delivery.OrganizationID == locator.organizationID && delivery.CalendarEventID == locator.eventID &&
		delivery.CalendarAttendeeID == locator.attendeeID && mailboxEntCalendarDeliveryRecurrencePointersEqual(delivery.RecurrenceID, locator.recurrenceID)
}

func mailboxEntReadCalendarInviteDeliveryLocator(ctx context.Context, transaction *ent.Tx, deliveryID CalendarMailDeliveryID) (mailboxEntCalendarInviteDeliveryLocator, bool, error) {
	if ctx == nil || transaction == nil || deliveryID.Kind() != CalendarMailDeliveryKindInvite || !validCalendarMailDeliveryID(deliveryID) {
		return mailboxEntCalendarInviteDeliveryLocator{}, false, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	row, err := transaction.CalendarInviteDelivery.Query().
		Where(calendarinvitedelivery.IDEQ(deliveryID.UUID())).
		Only(ctx)
	if ent.IsNotFound(err) {
		return mailboxEntCalendarInviteDeliveryLocator{}, false, nil
	}
	if err != nil || row == nil || row.ID != deliveryID.UUID() {
		return mailboxEntCalendarInviteDeliveryLocator{}, false, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	locator := mailboxEntCalendarInviteDeliveryLocator{
		id:             row.ID,
		organizationID: row.OrganizationID,
		eventID:        row.CalendarEventID,
		attendeeID:     row.CalendarAttendeeID,
		kind:           row.Kind,
	}
	if row.RecurrenceID != nil {
		recurrenceID := *row.RecurrenceID
		locator.recurrenceID = &recurrenceID
	}
	if locator.id == uuid.Nil || locator.organizationID == uuid.Nil || locator.eventID == uuid.Nil || locator.attendeeID == uuid.Nil {
		return mailboxEntCalendarInviteDeliveryLocator{}, false, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return locator, true, nil
}

func mailboxEntLockCalendarInviteDeliveryFacts(ctx context.Context, transaction *ent.Tx, locator mailboxEntCalendarInviteDeliveryLocator) (mailboxEntCalendarInviteDeliveryLockedFacts, error) {
	if ctx == nil || transaction == nil || locator.id == uuid.Nil || locator.organizationID == uuid.Nil || locator.eventID == uuid.Nil || locator.attendeeID == uuid.Nil {
		return mailboxEntCalendarInviteDeliveryLockedFacts{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	event, err := transaction.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(locator.eventID),
			calendarevent.OrganizationIDEQ(locator.organizationID),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || event == nil || event.ID != locator.eventID || event.OrganizationID != locator.organizationID {
		return mailboxEntCalendarInviteDeliveryLockedFacts{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	attendees, err := transaction.CalendarAttendee.Query().
		Where(
			calendarattendee.CalendarEventIDEQ(event.ID),
			calendarattendee.RecurrenceIDIsNil(),
		).
		Order(ent.Asc(calendarattendee.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil || len(attendees) == 0 {
		return mailboxEntCalendarInviteDeliveryLockedFacts{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	facts := mailboxEntCalendarInviteDeliveryLockedFacts{event: event, attendees: attendees}
	if locator.recurrenceID == nil || locator.kind != calendarinvitedelivery.KindCancelled {
		return facts, nil
	}
	exception, err := transaction.CalendarEventException.Query().
		Where(
			calendareventexception.CalendarEventIDEQ(event.ID),
			calendareventexception.RecurrenceIDEQ(*locator.recurrenceID),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || exception == nil || exception.CalendarEventID != event.ID || exception.RecurrenceID != *locator.recurrenceID {
		return mailboxEntCalendarInviteDeliveryLockedFacts{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	facts.exception = exception
	return facts, nil
}

func mailboxEntCalendarInviteDeliveryForUpdate(ctx context.Context, transaction *ent.Tx, locator mailboxEntCalendarInviteDeliveryLocator) (*ent.CalendarInviteDelivery, bool, error) {
	if ctx == nil || transaction == nil || locator.id == uuid.Nil {
		return nil, false, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	delivery, err := transaction.CalendarInviteDelivery.Query().
		Where(calendarinvitedelivery.IDEQ(locator.id)).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, false, nil
	}
	if err != nil || delivery == nil || delivery.ID != locator.id {
		return nil, false, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return delivery, true, nil
}

func mailboxEntCalendarInviteDeliveryLocatorMatches(locator mailboxEntCalendarInviteDeliveryLocator, delivery *ent.CalendarInviteDelivery) bool {
	return delivery != nil && delivery.ID == locator.id && delivery.OrganizationID == locator.organizationID &&
		delivery.CalendarEventID == locator.eventID && delivery.CalendarAttendeeID == locator.attendeeID &&
		delivery.Kind == locator.kind && mailboxEntCalendarDeliveryRecurrencePointersEqual(delivery.RecurrenceID, locator.recurrenceID)
}

func mailboxEntCalendarInviteDeliveryStaleSequenceFactsValid(delivery *ent.CalendarInviteDelivery, facts mailboxEntCalendarInviteDeliveryLockedFacts) bool {
	if delivery == nil || facts.event == nil || len(facts.attendees) == 0 || delivery.ID == uuid.Nil || delivery.OrganizationID == uuid.Nil ||
		delivery.CalendarEventID != facts.event.ID || facts.event.OrganizationID != delivery.OrganizationID || facts.event.DeletedAt != nil ||
		facts.event.OrganizerID == nil || *facts.event.OrganizerID == uuid.Nil || facts.event.Sequence <= delivery.EventSequence {
		return false
	}
	if delivery.RecurrenceID != nil {
		return mailboxEntCalendarInviteOccurrenceCancelDeliveryStaleSequenceFactsValid(delivery, facts)
	}
	var target *ent.CalendarAttendee
	for _, attendee := range facts.attendees {
		if attendee == nil || attendee.ID == uuid.Nil || attendee.CalendarEventID != facts.event.ID || attendee.RecurrenceID != nil {
			return false
		}
		if attendee.ID == delivery.CalendarAttendeeID {
			target = attendee
		}
		if attendee.RemovedAt != nil && attendee.ID != delivery.CalendarAttendeeID {
			continue
		}
		if _, ok := mailboxEntCalendarDeliveryITIPParticipant(attendee, attendee.ID == delivery.CalendarAttendeeID && delivery.Kind != calendarinvitedelivery.KindCancelled); !ok {
			return false
		}
	}
	if target == nil || target.CalendarEventID != facts.event.ID || target.RecurrenceID != nil {
		return false
	}
	switch delivery.Kind {
	case calendarinvitedelivery.KindCreated, calendarinvitedelivery.KindUpdated:
		return target.RemovedAt == nil && (facts.event.Status == calendarevent.StatusConfirmed || facts.event.Status == calendarevent.StatusTentative)
	case calendarinvitedelivery.KindCancelled:
		return (facts.event.Status == calendarevent.StatusCancelled && target.RemovedAt == nil) ||
			(target.RemovedAt != nil && (facts.event.Status == calendarevent.StatusConfirmed || facts.event.Status == calendarevent.StatusTentative))
	default:
		return false
	}
}

func mailboxEntCalendarInviteOccurrenceCancelDeliveryStaleSequenceFactsValid(delivery *ent.CalendarInviteDelivery, facts mailboxEntCalendarInviteDeliveryLockedFacts) bool {
	if delivery == nil || facts.event == nil || facts.exception == nil || delivery.Kind != calendarinvitedelivery.KindCancelled || delivery.RecurrenceID == nil || *delivery.RecurrenceID == "" ||
		facts.exception.CalendarEventID != facts.event.ID || facts.exception.RecurrenceID != *delivery.RecurrenceID || facts.exception.Sequence != delivery.EventSequence ||
		facts.exception.Status == nil || *facts.exception.Status != calendareventexception.StatusCancelled || facts.exception.StartsAt != nil || facts.exception.EndsAt != nil ||
		facts.exception.StartDate != nil || facts.exception.EndDateExclusive != nil || facts.exception.Timezone != nil || facts.exception.Location != nil ||
		(facts.event.Status != calendarevent.StatusConfirmed && facts.event.Status != calendarevent.StatusTentative) {
		return false
	}
	var target *ent.CalendarAttendee
	for _, attendee := range facts.attendees {
		if attendee == nil || attendee.ID == uuid.Nil || attendee.CalendarEventID != facts.event.ID || attendee.RecurrenceID != nil {
			return false
		}
		if attendee.ID == delivery.CalendarAttendeeID {
			target = attendee
		}
		if attendee.RemovedAt != nil {
			continue
		}
		participant, ok := mailboxEntCalendarDeliveryITIPParticipant(attendee, false)
		if !ok || participant.RSVP {
			return false
		}
	}
	return target != nil && target.CalendarEventID == facts.event.ID && target.RecurrenceID == nil && target.RemovedAt == nil
}

func mailboxEntCalendarReplyDeliveryStaleSequenceFactsValid(delivery *ent.CalendarReplyDelivery, facts mailboxEntCalendarReplyDeliveryLockedFacts) bool {
	return delivery != nil && facts.event != nil && facts.attendee != nil && delivery.ID != uuid.Nil && delivery.OrganizationID != uuid.Nil &&
		delivery.RecurrenceID == nil && delivery.CalendarEventID == facts.event.ID && facts.event.OrganizationID == delivery.OrganizationID &&
		facts.event.DeletedAt == nil && facts.event.OrganizerID == nil && facts.event.Sequence > delivery.EventSequence &&
		facts.attendee.ID == delivery.CalendarAttendeeID && facts.attendee.CalendarEventID == facts.event.ID && facts.attendee.RecurrenceID == nil &&
		facts.attendee.RemovedAt == nil && facts.attendee.UserID != nil && *facts.attendee.UserID != uuid.Nil &&
		facts.attendee.RsvpStatus == calendarattendee.RsvpStatus(delivery.RsvpStatus) && delivery.AttendeeVersion >= 1
}

type mailboxEntCalendarReplyOccurrenceFacts struct {
	sequence   int
	status     calendarevent.Status
	temporal   ical.ITIPTemporal
	recurrence ical.ITIPRecurrenceID
	superseded bool
}

// mailboxEntCalendarReplyDeliveryOccurrenceFacts resolves one durable RSVP
// against its nominal recurrence slot. The slot identity always comes from
// the master schedule; an active exception can replace only the temporal
// payload and sequence, never RECURRENCE-ID itself.
func mailboxEntCalendarReplyDeliveryOccurrenceFacts(delivery *ent.CalendarReplyDelivery, facts mailboxEntCalendarReplyDeliveryLockedFacts) (mailboxEntCalendarReplyOccurrenceFacts, bool) {
	if delivery == nil || facts.event == nil || facts.attendee == nil || delivery.ID == uuid.Nil || delivery.OrganizationID == uuid.Nil ||
		delivery.RecurrenceID == nil || *delivery.RecurrenceID == "" || !validCalendarMailDeliveryRecurrenceID(*delivery.RecurrenceID) ||
		delivery.EventSequence < 0 || delivery.AttendeeVersion < 1 || delivery.CalendarEventID != facts.event.ID || facts.event.OrganizationID != delivery.OrganizationID ||
		facts.event.DeletedAt != nil || facts.event.OrganizerID != nil || facts.event.Sequence < 0 ||
		facts.attendee.ID != delivery.CalendarAttendeeID || facts.attendee.CalendarEventID != facts.event.ID ||
		!mailboxEntCalendarDeliveryRecurrencePointersEqual(facts.attendee.RecurrenceID, delivery.RecurrenceID) {
		return mailboxEntCalendarReplyOccurrenceFacts{}, false
	}

	master, exdates, ok := mailboxEntCalendarDeliveryRecurrenceMaster(facts.event)
	if !ok {
		return mailboxEntCalendarReplyOccurrenceFacts{}, false
	}
	nominal, ok := ical.ResolveNominalOccurrence(master, mailboxEntCalendarEventRRULE(facts.event), *delivery.RecurrenceID)
	if !ok || nominal.RecurrenceID != *delivery.RecurrenceID || nominal.Temporal.AllDay != master.AllDay {
		return mailboxEntCalendarReplyOccurrenceFacts{}, false
	}
	temporal, recurrence, ok := mailboxEntCalendarDeliveryOccurrenceITIPTemporal(nominal)
	if !ok {
		return mailboxEntCalendarReplyOccurrenceFacts{}, false
	}
	if mailboxEntCalendarDeliveryRecurrenceExcluded(master.AllDay, exdates, *delivery.RecurrenceID) {
		return mailboxEntCalendarReplyOccurrenceFacts{superseded: true}, true
	}

	if facts.exception == nil {
		status, ok := mailboxEntCalendarReplyOccurrenceStatus(facts.event.Status)
		if !ok {
			return mailboxEntCalendarReplyOccurrenceFacts{}, false
		}
		return mailboxEntCalendarReplyOccurrenceFacts{sequence: facts.event.Sequence, status: status, temporal: temporal, recurrence: recurrence}, true
	}
	exception := facts.exception
	if exception.ID == uuid.Nil || exception.CalendarEventID != facts.event.ID || exception.RecurrenceID != *delivery.RecurrenceID ||
		exception.Status == nil || exception.Sequence < 0 || exception.Sequence > facts.event.Sequence {
		return mailboxEntCalendarReplyOccurrenceFacts{}, false
	}
	switch *exception.Status {
	case calendareventexception.StatusCancelled:
		return mailboxEntCalendarReplyOccurrenceFacts{superseded: true}, true
	case calendareventexception.StatusConfirmed, calendareventexception.StatusTentative:
		status, ok := mailboxEntCalendarReplyOccurrenceStatus(calendarevent.Status(*exception.Status))
		if !ok {
			return mailboxEntCalendarReplyOccurrenceFacts{}, false
		}
		effectiveTemporal, ok := mailboxEntCalendarReplyOccurrenceOverrideTemporal(nominal, exception)
		if !ok {
			return mailboxEntCalendarReplyOccurrenceFacts{}, false
		}
		return mailboxEntCalendarReplyOccurrenceFacts{sequence: exception.Sequence, status: status, temporal: effectiveTemporal, recurrence: recurrence}, true
	default:
		return mailboxEntCalendarReplyOccurrenceFacts{}, false
	}
}

func mailboxEntCalendarReplyOccurrenceStatus(status calendarevent.Status) (calendarevent.Status, bool) {
	switch status {
	case calendarevent.StatusConfirmed, calendarevent.StatusTentative:
		return status, true
	default:
		return "", false
	}
}

// mailboxEntCalendarReplyOccurrenceOverrideTemporal accepts only a complete
// override in the same timed/all-day class as the nominal instance. A partial
// or cross-class exception cannot be represented safely in a REPLY.
func mailboxEntCalendarReplyOccurrenceOverrideTemporal(nominal ical.NominalOccurrence, exception *ent.CalendarEventException) (ical.ITIPTemporal, bool) {
	if exception == nil || nominal.RecurrenceID == "" || nominal.Temporal.Validate() != nil {
		return ical.ITIPTemporal{}, false
	}
	hasTimed := exception.StartsAt != nil || exception.EndsAt != nil
	hasAllDay := exception.StartDate != nil || exception.EndDateExclusive != nil
	if !hasTimed && !hasAllDay && exception.Timezone == nil {
		temporal, _, ok := mailboxEntCalendarDeliveryOccurrenceITIPTemporal(nominal)
		return temporal, ok
	}
	if hasTimed == hasAllDay || exception.Timezone == nil || *exception.Timezone == "" {
		return ical.ITIPTemporal{}, false
	}
	if hasTimed {
		if nominal.Temporal.AllDay || exception.StartsAt == nil || exception.EndsAt == nil ||
			!mailboxEntValidTime(*exception.StartsAt) || !mailboxEntValidTime(*exception.EndsAt) || !exception.EndsAt.After(*exception.StartsAt) {
			return ical.ITIPTemporal{}, false
		}
		return ical.ITIPTemporal{Timed: &ical.ITIPTimed{StartsAt: *exception.StartsAt, EndsAt: *exception.EndsAt, Timezone: *exception.Timezone}}, true
	}
	if !nominal.Temporal.AllDay || exception.StartDate == nil || exception.EndDateExclusive == nil || exception.StartDate.Format("2006-01-02") == "" ||
		!exception.EndDateExclusive.After(*exception.StartDate) {
		return ical.ITIPTemporal{}, false
	}
	startDate, err := ical.ParseCalendarDate(exception.StartDate.Format("2006-01-02"))
	if err != nil || startDate.Format("2006-01-02") != exception.StartDate.Format("2006-01-02") {
		return ical.ITIPTemporal{}, false
	}
	endDate, err := ical.ParseCalendarDate(exception.EndDateExclusive.Format("2006-01-02"))
	if err != nil || endDate.Format("2006-01-02") != exception.EndDateExclusive.Format("2006-01-02") || !endDate.After(startDate) {
		return ical.ITIPTemporal{}, false
	}
	return ical.ITIPTemporal{AllDay: &ical.ITIPAllDay{
		StartDate:        startDate.Format("2006-01-02"),
		EndDateExclusive: endDate.Format("2006-01-02"),
		Timezone:         *exception.Timezone,
	}}, true
}

func mailboxEntCalendarReplyOccurrenceDeliveryStaleFactsValid(delivery *ent.CalendarReplyDelivery, facts mailboxEntCalendarReplyDeliveryLockedFacts, occurrence mailboxEntCalendarReplyOccurrenceFacts) bool {
	if delivery == nil || facts.event == nil || facts.attendee == nil || occurrence.superseded || delivery.RecurrenceID == nil ||
		occurrence.sequence < delivery.EventSequence || facts.attendee.Version < delivery.AttendeeVersion ||
		delivery.EventSequence < 0 || delivery.AttendeeVersion < 1 || facts.attendee.Version < 1 ||
		facts.event.DeletedAt != nil || facts.event.OrganizerID != nil || facts.attendee.RemovedAt != nil || facts.attendee.UserID == nil || *facts.attendee.UserID == uuid.Nil ||
		!mailboxEntCalendarDeliveryRecurrencePointersEqual(delivery.RecurrenceID, facts.attendee.RecurrenceID) {
		return false
	}
	_, deliveryStatusOK := mailboxEntCalendarDeliveryRSVPStatus(calendarattendee.RsvpStatus(delivery.RsvpStatus))
	_, attendeeStatusOK := mailboxEntCalendarDeliveryRSVPStatus(facts.attendee.RsvpStatus)
	return deliveryStatusOK && attendeeStatusOK && (occurrence.sequence > delivery.EventSequence || facts.attendee.Version > delivery.AttendeeVersion)
}

func mailboxEntClaimDueCalendarInviteDelivery(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery, deliveryID CalendarMailDeliveryID, facts mailboxEntCalendarInviteDeliveryLockedFacts, databaseNow time.Time, reclaim bool) (CalendarMailDeliveryClaimResult, error) {
	if ctx == nil || transaction == nil || delivery == nil || facts.event == nil || len(facts.attendees) == 0 || deliveryID.Kind() != CalendarMailDeliveryKindInvite || !mailboxEntValidTime(databaseNow) || reclaim != (CalendarMailDeliveryState(delivery.State) == CalendarMailDeliveryStateSubmitting) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if facts.event.Sequence < 0 || delivery.EventSequence < 0 || facts.event.Sequence < delivery.EventSequence {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if facts.event.Sequence > delivery.EventSequence {
		if !mailboxEntCalendarInviteDeliveryStaleSequenceFactsValid(delivery, facts) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		updated, err := mailboxEntCASSupersedeCalendarInviteDelivery(ctx, transaction, delivery, databaseNow, reclaim)
		if err != nil {
			return CalendarMailDeliveryClaimResult{}, err
		}
		if updated != 1 {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
		}
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeSuperseded}, nil
	}
	nextGeneration, leaseUntil, token, ok := mailboxEntCalendarDeliveryClaimLease(delivery.ClaimGeneration, databaseNow)
	if !ok || (!reclaim && delivery.Attempts >= calendarMailDeliveryMaximumAttempts) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	snapshot, err := mailboxEntCalendarInviteDeliveryClaimSnapshot(ctx, transaction, delivery, facts, databaseNow)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	claim, err := mailboxEntCalendarInviteDeliveryClaim(delivery, deliveryID, snapshot, token, nextGeneration, leaseUntil)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}

	updated, err := mailboxEntCASCalendarInviteDeliveryClaim(ctx, transaction, delivery, databaseNow, token, nextGeneration, leaseUntil, reclaim)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	if updated != 1 {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeClaimed, claim: claim}, nil
}

func mailboxEntClaimDueCalendarReplyDelivery(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarReplyDelivery, deliveryID CalendarMailDeliveryID, facts mailboxEntCalendarReplyDeliveryLockedFacts, databaseNow time.Time, reclaim bool) (CalendarMailDeliveryClaimResult, error) {
	if ctx == nil || transaction == nil || delivery == nil || facts.event == nil || facts.attendee == nil || deliveryID.Kind() != CalendarMailDeliveryKindReply || !mailboxEntValidTime(databaseNow) || reclaim != (CalendarMailDeliveryState(delivery.State) == CalendarMailDeliveryStateSubmitting) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if delivery.RecurrenceID != nil {
		occurrence, ok := mailboxEntCalendarReplyDeliveryOccurrenceFacts(delivery, facts)
		if !ok {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if occurrence.superseded {
			updated, err := mailboxEntCASSupersedeCalendarReplyDelivery(ctx, transaction, delivery, databaseNow, reclaim)
			if err != nil {
				return CalendarMailDeliveryClaimResult{}, err
			}
			if updated != 1 {
				return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
			}
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeSuperseded}, nil
		}
		if occurrence.sequence < delivery.EventSequence || facts.attendee.Version < delivery.AttendeeVersion {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if occurrence.sequence > delivery.EventSequence || facts.attendee.Version > delivery.AttendeeVersion {
			if !mailboxEntCalendarReplyOccurrenceDeliveryStaleFactsValid(delivery, facts, occurrence) {
				return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
			}
			updated, err := mailboxEntCASSupersedeCalendarReplyDelivery(ctx, transaction, delivery, databaseNow, reclaim)
			if err != nil {
				return CalendarMailDeliveryClaimResult{}, err
			}
			if updated != 1 {
				return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
			}
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeSuperseded}, nil
		}
	} else {
		if facts.event.Sequence < 0 || delivery.EventSequence < 0 || facts.event.Sequence < delivery.EventSequence {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if facts.event.Sequence > delivery.EventSequence {
			if !mailboxEntCalendarReplyDeliveryStaleSequenceFactsValid(delivery, facts) {
				return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
			}
			updated, err := mailboxEntCASSupersedeCalendarReplyDelivery(ctx, transaction, delivery, databaseNow, reclaim)
			if err != nil {
				return CalendarMailDeliveryClaimResult{}, err
			}
			if updated != 1 {
				return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
			}
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeSuperseded}, nil
		}
	}
	nextGeneration, leaseUntil, token, ok := mailboxEntCalendarDeliveryClaimLease(delivery.ClaimGeneration, databaseNow)
	if !ok || (!reclaim && delivery.Attempts >= calendarMailDeliveryMaximumAttempts) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	snapshot, err := mailboxEntCalendarReplyDeliveryClaimSnapshot(ctx, transaction, delivery, facts, databaseNow)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	claim, err := mailboxEntCalendarReplyDeliveryClaim(delivery, deliveryID, snapshot, token, nextGeneration, leaseUntil)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}

	updated, err := mailboxEntCASCalendarReplyDeliveryClaim(ctx, transaction, delivery, databaseNow, token, nextGeneration, leaseUntil, reclaim)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	if updated != 1 {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeClaimed, claim: claim}, nil
}

func mailboxEntCalendarDeliveryClaimLease(generation int64, databaseNow time.Time) (CalendarMailDeliveryClaimGeneration, time.Time, CalendarMailDeliveryClaimToken, bool) {
	if generation < 0 || generation >= mailboxEntCalendarDeliveryMaximumClaimGeneration || !mailboxEntValidTime(databaseNow) {
		return 0, time.Time{}, CalendarMailDeliveryClaimToken{}, false
	}
	leaseUntil, ok := mailboxEntCheckedTimeAdd(databaseNow, calendarMailDeliveryLease)
	if !ok || !leaseUntil.After(databaseNow) {
		return 0, time.Time{}, CalendarMailDeliveryClaimToken{}, false
	}
	token := CalendarMailDeliveryClaimToken(uuid.New())
	if uuid.UUID(token) == uuid.Nil {
		return 0, time.Time{}, CalendarMailDeliveryClaimToken{}, false
	}
	return CalendarMailDeliveryClaimGeneration(generation + 1), leaseUntil, token, true
}

func mailboxEntCalendarInviteDeliveryClaimSnapshot(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery, facts mailboxEntCalendarInviteDeliveryLockedFacts, databaseNow time.Time) (mailboxEntCalendarDeliveryClaimSnapshot, error) {
	if ctx == nil || transaction == nil || delivery == nil || facts.event == nil || len(facts.attendees) == 0 || !mailboxEntValidTime(databaseNow) {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	stamp, ok := mailboxEntCalendarDeliveryStableITIPStamp(delivery.CreatedAt)
	if !ok {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	event := facts.event
	if event.DeletedAt != nil || event.OrganizerID == nil || *event.OrganizerID == uuid.Nil ||
		event.Sequence != delivery.EventSequence || event.ID != delivery.CalendarEventID || event.OrganizationID != delivery.OrganizationID {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if delivery.RecurrenceID != nil {
		return mailboxEntCalendarInviteOccurrenceCancelDeliveryClaimSnapshot(ctx, transaction, delivery, event, facts.attendees, facts.exception, stamp)
	}

	attendees := facts.attendees
	if len(attendees) == 0 {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	var target *ent.CalendarAttendee
	var targetParticipant ical.ITIPParticipant
	targetParticipantOK := false
	participants := make([]ical.ITIPParticipant, 0, len(attendees))
	for _, attendee := range attendees {
		if attendee == nil || attendee.ID == uuid.Nil || attendee.CalendarEventID != event.ID || attendee.RecurrenceID != nil {
			return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if attendee.ID == delivery.CalendarAttendeeID {
			target = attendee
		}
		if attendee.RemovedAt != nil && attendee.ID != delivery.CalendarAttendeeID {
			continue
		}
		participant, ok := mailboxEntCalendarDeliveryITIPParticipant(attendee, attendee.ID == delivery.CalendarAttendeeID && delivery.Kind != calendarinvitedelivery.KindCancelled)
		if !ok {
			return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if attendee.ID == delivery.CalendarAttendeeID {
			targetParticipant = participant
			targetParticipantOK = true
		}
		participants = append(participants, participant)
	}
	if target == nil || target.CalendarEventID != event.ID || target.RecurrenceID != nil ||
		!mailboxEntCalendarInviteDeliveryFactsMatch(delivery, event, target) {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	sourceUser, account, identity, err := mailboxEntCalendarDeliverySourceIdentity(ctx, transaction, event.OrganizationID, *event.OrganizerID)
	if err != nil {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, err
	}
	recipient, err := NormalizeAddress(EngineAddress{Email: target.NormalizedEmail, Name: mailboxEntCalendarAttendeeDisplayName(target)})
	if err != nil {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	method, ok := mailboxEntCalendarInviteDeliveryMethod(delivery.Kind)
	if !ok {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	var message ical.ITIPMessage
	if delivery.Kind == calendarinvitedelivery.KindCancelled && target.RemovedAt != nil {
		if !targetParticipantOK || targetParticipant.RSVP {
			return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		message, ok = mailboxEntCalendarDeliveryRemovedAttendeeMasterCancelITIPMessage(
			event,
			mailboxEntCalendarUserDisplayName(sourceUser),
			event.OrganizerEmail,
			targetParticipant,
			stamp,
		)
	} else {
		message, ok = mailboxEntCalendarDeliveryMasterITIPMessage(event, method, mailboxEntCalendarUserDisplayName(sourceUser), event.OrganizerEmail, participants, stamp)
	}
	if !ok || !calendarMailDeliveryMessageHasAttendee(message, recipient.Email) {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return mailboxEntCalendarDeliveryClaimSnapshot{
		organizationID: event.OrganizationID,
		eventID:        event.ID,
		attendeeID:     target.ID,
		organizerID:    *event.OrganizerID,
		accountID:      account.ID,
		senderActorID:  sourceUser.ID,
		fromIdentityID: identity.ID,
		recipient:      recipient,
		method:         method,
		eventSequence:  delivery.EventSequence,
		message:        message,
	}, nil
}

// mailboxEntCalendarInviteOccurrenceCancelDeliveryClaimSnapshot materializes
// only an authoritative cancelled recurrence exception. It intentionally does
// not reuse the master matcher: an active master and a cancelled exception is
// the distinct, valid occurrence-CANCEL shape.
func mailboxEntCalendarInviteOccurrenceCancelDeliveryClaimSnapshot(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery, event *ent.CalendarEvent, attendees []*ent.CalendarAttendee, exception *ent.CalendarEventException, stamp time.Time) (mailboxEntCalendarDeliveryClaimSnapshot, error) {
	if ctx == nil || transaction == nil || delivery == nil || event == nil || len(attendees) == 0 || exception == nil || delivery.RecurrenceID == nil ||
		delivery.Kind != calendarinvitedelivery.KindCancelled || !mailboxEntValidTime(stamp) || stamp.Location() != time.UTC {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	recurrenceID := *delivery.RecurrenceID
	if exception.CalendarEventID != event.ID || exception.RecurrenceID != recurrenceID ||
		exception.Sequence != delivery.EventSequence || exception.Status == nil || *exception.Status != calendareventexception.StatusCancelled ||
		exception.StartsAt != nil || exception.EndsAt != nil || exception.StartDate != nil || exception.EndDateExclusive != nil || exception.Timezone != nil || exception.Location != nil {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if event.DeletedAt != nil || event.OrganizerID == nil || *event.OrganizerID == uuid.Nil || event.Sequence != delivery.EventSequence ||
		(event.Status != calendarevent.StatusConfirmed && event.Status != calendarevent.StatusTentative) {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	var target *ent.CalendarAttendee
	participants := make([]ical.ITIPParticipant, 0, len(attendees))
	for _, attendee := range attendees {
		if attendee == nil || attendee.ID == uuid.Nil || attendee.CalendarEventID != event.ID || attendee.RecurrenceID != nil {
			return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if attendee.ID == delivery.CalendarAttendeeID {
			target = attendee
		}
		if attendee.RemovedAt != nil {
			continue
		}
		participant, ok := mailboxEntCalendarDeliveryITIPParticipant(attendee, false)
		if !ok || participant.RSVP {
			return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		participants = append(participants, participant)
	}
	if target == nil || !mailboxEntCalendarInviteOccurrenceCancelDeliveryFactsMatch(delivery, event, target) || len(participants) == 0 {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	sourceUser, account, identity, err := mailboxEntCalendarDeliverySourceIdentity(ctx, transaction, event.OrganizationID, *event.OrganizerID)
	if err != nil {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, err
	}
	recipient, err := NormalizeAddress(EngineAddress{Email: target.NormalizedEmail, Name: mailboxEntCalendarAttendeeDisplayName(target)})
	if err != nil {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	message, ok := mailboxEntCalendarDeliveryOccurrenceCancelITIPMessage(
		event,
		recurrenceID,
		mailboxEntCalendarUserDisplayName(sourceUser),
		event.OrganizerEmail,
		participants,
		stamp,
	)
	if !ok || len(message.Attendees) != len(participants) || !calendarMailDeliveryMessageHasAttendee(message, recipient.Email) {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return mailboxEntCalendarDeliveryClaimSnapshot{
		organizationID: event.OrganizationID,
		eventID:        event.ID,
		attendeeID:     target.ID,
		organizerID:    *event.OrganizerID,
		accountID:      account.ID,
		senderActorID:  sourceUser.ID,
		fromIdentityID: identity.ID,
		recipient:      recipient,
		method:         CalendarMailDeliveryMethodCancel,
		eventSequence:  delivery.EventSequence,
		recurrenceID:   recurrenceID,
		message:        message,
	}, nil
}

func mailboxEntCalendarInviteOccurrenceCancelDeliveryFactsMatch(delivery *ent.CalendarInviteDelivery, event *ent.CalendarEvent, attendee *ent.CalendarAttendee) bool {
	return delivery != nil && event != nil && attendee != nil && delivery.ID != uuid.Nil && delivery.OrganizationID != uuid.Nil &&
		delivery.Kind == calendarinvitedelivery.KindCancelled && delivery.RecurrenceID != nil && *delivery.RecurrenceID != "" &&
		delivery.CalendarEventID == event.ID && delivery.CalendarAttendeeID == attendee.ID && event.OrganizationID == delivery.OrganizationID &&
		attendee.CalendarEventID == event.ID && attendee.RecurrenceID == nil && attendee.RemovedAt == nil && event.DeletedAt == nil &&
		event.OrganizerID != nil && *event.OrganizerID != uuid.Nil && event.Sequence == delivery.EventSequence &&
		(event.Status == calendarevent.StatusConfirmed || event.Status == calendarevent.StatusTentative)
}

func mailboxEntCalendarReplyDeliveryClaimSnapshot(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarReplyDelivery, facts mailboxEntCalendarReplyDeliveryLockedFacts, databaseNow time.Time) (mailboxEntCalendarDeliveryClaimSnapshot, error) {
	if ctx == nil || transaction == nil || delivery == nil || facts.event == nil || facts.attendee == nil || !mailboxEntValidTime(databaseNow) {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	stamp, ok := mailboxEntCalendarDeliveryStableITIPStamp(delivery.CreatedAt)
	if !ok {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if delivery.RecurrenceID != nil {
		return mailboxEntCalendarReplyOccurrenceDeliveryClaimSnapshot(ctx, transaction, delivery, facts, stamp)
	}

	event := facts.event
	if event.DeletedAt != nil || event.OrganizerID != nil || event.Sequence != delivery.EventSequence ||
		event.ID != delivery.CalendarEventID || event.OrganizationID != delivery.OrganizationID {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	attendee := facts.attendee
	if attendee.RemovedAt != nil || attendee.UserID == nil || *attendee.UserID == uuid.Nil ||
		attendee.CalendarEventID != event.ID || attendee.RecurrenceID != nil || attendee.RsvpStatus != calendarattendee.RsvpStatus(delivery.RsvpStatus) ||
		delivery.AttendeeVersion < 1 {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	sourceUser, account, identity, err := mailboxEntCalendarDeliverySourceIdentity(ctx, transaction, event.OrganizationID, *attendee.UserID)
	if err != nil {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, err
	}
	recipient, err := NormalizeAddress(EngineAddress{Email: event.OrganizerEmail})
	if err != nil {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	participant, ok := mailboxEntCalendarDeliveryITIPParticipant(attendee, false)
	if !ok || participant.Status == ical.RSVPNeedsAction {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	message, ok := mailboxEntCalendarDeliveryMasterITIPMessage(event, CalendarMailDeliveryMethodReply, "", recipient.Email, []ical.ITIPParticipant{participant}, stamp)
	if !ok || len(message.Attendees) != 1 || message.Attendees[0].Email != attendee.NormalizedEmail || message.Organizer.Email != recipient.Email {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return mailboxEntCalendarDeliveryClaimSnapshot{
		organizationID:  event.OrganizationID,
		eventID:         event.ID,
		attendeeID:      attendee.ID,
		accountID:       account.ID,
		senderActorID:   sourceUser.ID,
		fromIdentityID:  identity.ID,
		recipient:       recipient,
		method:          CalendarMailDeliveryMethodReply,
		eventSequence:   delivery.EventSequence,
		attendeeVersion: delivery.AttendeeVersion,
		message:         message,
	}, nil
}

func mailboxEntCalendarReplyOccurrenceDeliveryClaimSnapshot(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarReplyDelivery, facts mailboxEntCalendarReplyDeliveryLockedFacts, stamp time.Time) (mailboxEntCalendarDeliveryClaimSnapshot, error) {
	if ctx == nil || transaction == nil || delivery == nil || facts.event == nil || facts.attendee == nil || delivery.RecurrenceID == nil ||
		!mailboxEntValidTime(stamp) || stamp.Location() != time.UTC {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	occurrence, ok := mailboxEntCalendarReplyDeliveryOccurrenceFacts(delivery, facts)
	if !ok || occurrence.superseded || occurrence.sequence != delivery.EventSequence {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	event := facts.event
	attendee := facts.attendee
	if event.DeletedAt != nil || event.OrganizerID != nil || event.ID != delivery.CalendarEventID || event.OrganizationID != delivery.OrganizationID ||
		attendee.ID != delivery.CalendarAttendeeID || attendee.CalendarEventID != event.ID || attendee.RecurrenceID == nil ||
		!mailboxEntCalendarDeliveryRecurrencePointersEqual(attendee.RecurrenceID, delivery.RecurrenceID) || attendee.RemovedAt != nil ||
		attendee.UserID == nil || *attendee.UserID == uuid.Nil || attendee.Version != delivery.AttendeeVersion ||
		attendee.RsvpStatus != calendarattendee.RsvpStatus(delivery.RsvpStatus) {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	sourceUser, account, identity, err := mailboxEntCalendarDeliverySourceIdentity(ctx, transaction, event.OrganizationID, *attendee.UserID)
	if err != nil {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, err
	}
	recipient, err := NormalizeAddress(EngineAddress{Email: event.OrganizerEmail})
	if err != nil {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	participant, ok := mailboxEntCalendarReplyOccurrenceITIPParticipant(attendee)
	if !ok || participant.Status == ical.RSVPNeedsAction {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	message, ok := mailboxEntCalendarReplyOccurrenceITIPMessage(event, occurrence, recipient.Email, participant, stamp)
	if !ok || len(message.Attendees) != 1 || message.Attendees[0].Email != attendee.NormalizedEmail ||
		message.Attendees[0].Status != participant.Status || message.Organizer.Email != recipient.Email {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return mailboxEntCalendarDeliveryClaimSnapshot{
		organizationID:  event.OrganizationID,
		eventID:         event.ID,
		attendeeID:      attendee.ID,
		accountID:       account.ID,
		senderActorID:   sourceUser.ID,
		fromIdentityID:  identity.ID,
		recipient:       recipient,
		method:          CalendarMailDeliveryMethodReply,
		eventSequence:   occurrence.sequence,
		attendeeVersion: attendee.Version,
		recurrenceID:    *delivery.RecurrenceID,
		message:         message,
	}, nil
}

// mailboxEntCalendarDeliverySourceIdentity chooses the only primary identity
// that can be used for calendar delivery. Multiple or incomplete records are
// intentionally unavailable rather than silently selecting a sender.
func mailboxEntCalendarDeliverySourceIdentity(ctx context.Context, transaction *ent.Tx, organizationID, sourceUserID uuid.UUID) (*ent.User, *ent.MailAccount, *ent.MailIdentity, error) {
	if ctx == nil || transaction == nil || organizationID == uuid.Nil || sourceUserID == uuid.Nil {
		return nil, nil, nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	users, err := transaction.User.Query().
		Where(
			user.IDEQ(sourceUserID),
			user.StatusEQ(user.StatusActive),
			user.HasOrganizationWith(organization.IDEQ(organizationID)),
		).
		Order(ent.Asc(user.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil || len(users) != 1 || users[0] == nil || users[0].ID != sourceUserID {
		return nil, nil, nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	accounts, err := transaction.MailAccount.Query().
		Where(
			mailaccount.OrganizationIDEQ(organizationID),
			mailaccount.UserIDEQ(sourceUserID),
			mailaccount.StatusEQ(mailaccount.StatusActive),
			mailaccount.DesiredEnabledEQ(true),
			mailaccount.EngineAccountIDNotNil(),
		).
		Order(ent.Asc(mailaccount.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil || len(accounts) != 1 || accounts[0] == nil || accounts[0].ID == uuid.Nil || accounts[0].OrganizationID != organizationID ||
		accounts[0].UserID != sourceUserID || accounts[0].Status != mailaccount.StatusActive || !accounts[0].DesiredEnabled ||
		accounts[0].EngineAccountID == nil || *accounts[0].EngineAccountID == "" {
		return nil, nil, nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	identities, err := transaction.MailIdentity.Query().
		Where(
			mailidentity.MailAccountIDEQ(accounts[0].ID),
			mailidentity.KindEQ(mailidentity.KindPrimary),
			mailidentity.StatusEQ(mailidentity.StatusActive),
			mailidentity.SendAllowedEQ(true),
			mailidentity.EngineIdentityIDNotNil(),
		).
		Order(ent.Asc(mailidentity.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil || len(identities) != 1 || identities[0] == nil || identities[0].ID == uuid.Nil || identities[0].MailAccountID != accounts[0].ID ||
		identities[0].Kind != mailidentity.KindPrimary || identities[0].Status != mailidentity.StatusActive || !identities[0].SendAllowed ||
		identities[0].EngineIdentityID == nil || *identities[0].EngineIdentityID == "" {
		return nil, nil, nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return users[0], accounts[0], identities[0], nil
}

func mailboxEntCalendarInviteDeliveryFactsMatch(delivery *ent.CalendarInviteDelivery, event *ent.CalendarEvent, attendee *ent.CalendarAttendee) bool {
	if delivery == nil || event == nil || attendee == nil || delivery.ID == uuid.Nil || delivery.OrganizationID == uuid.Nil ||
		delivery.CalendarEventID != event.ID || delivery.CalendarAttendeeID != attendee.ID || event.OrganizationID != delivery.OrganizationID ||
		attendee.CalendarEventID != event.ID || event.DeletedAt != nil || event.OrganizerID == nil || *event.OrganizerID == uuid.Nil ||
		event.Sequence != delivery.EventSequence || delivery.RecurrenceID != nil || attendee.RecurrenceID != nil {
		return false
	}
	switch delivery.Kind {
	case calendarinvitedelivery.KindCreated, calendarinvitedelivery.KindUpdated:
		return attendee.RemovedAt == nil && (event.Status == calendarevent.StatusConfirmed || event.Status == calendarevent.StatusTentative)
	case calendarinvitedelivery.KindCancelled:
		return (event.Status == calendarevent.StatusCancelled && attendee.RemovedAt == nil) ||
			(attendee.RemovedAt != nil && (event.Status == calendarevent.StatusConfirmed || event.Status == calendarevent.StatusTentative))
	default:
		return false
	}
}

func mailboxEntCalendarInviteDeliveryMethod(kind calendarinvitedelivery.Kind) (CalendarMailDeliveryMethod, bool) {
	switch kind {
	case calendarinvitedelivery.KindCreated, calendarinvitedelivery.KindUpdated:
		return CalendarMailDeliveryMethodRequest, true
	case calendarinvitedelivery.KindCancelled:
		return CalendarMailDeliveryMethodCancel, true
	default:
		return "", false
	}
}

func mailboxEntCalendarDeliveryITIPParticipant(attendee *ent.CalendarAttendee, rsvp bool) (ical.ITIPParticipant, bool) {
	if attendee == nil || attendee.ID == uuid.Nil || attendee.CalendarEventID == uuid.Nil || attendee.RecurrenceID != nil {
		return ical.ITIPParticipant{}, false
	}
	role := ical.ITIPParticipantRequired
	switch attendee.Role {
	case calendarattendee.RoleOptional:
		role = ical.ITIPParticipantOptional
	case calendarattendee.RoleOrganizer, calendarattendee.RoleRequired:
	default:
		return ical.ITIPParticipant{}, false
	}
	status, ok := mailboxEntCalendarDeliveryRSVPStatus(attendee.RsvpStatus)
	if !ok {
		return ical.ITIPParticipant{}, false
	}
	return ical.ITIPParticipant{
		Email:       attendee.NormalizedEmail,
		DisplayName: mailboxEntCalendarAttendeeDisplayName(attendee),
		Role:        role,
		Status:      status,
		RSVP:        rsvp,
	}, true
}

// mailboxEntCalendarReplyOccurrenceITIPParticipant keeps occurrence REPLY
// attendee identity exact. The master mapper deliberately rejects materialized
// attendees, so occurrence delivery uses this separate closed mapper rather
// than widening REQUEST/CANCEL behavior.
func mailboxEntCalendarReplyOccurrenceITIPParticipant(attendee *ent.CalendarAttendee) (ical.ITIPParticipant, bool) {
	if attendee == nil || attendee.ID == uuid.Nil || attendee.CalendarEventID == uuid.Nil || attendee.RecurrenceID == nil ||
		*attendee.RecurrenceID == "" || !validCalendarMailDeliveryRecurrenceID(*attendee.RecurrenceID) {
		return ical.ITIPParticipant{}, false
	}
	role := ical.ITIPParticipantRequired
	switch attendee.Role {
	case calendarattendee.RoleOptional:
		role = ical.ITIPParticipantOptional
	case calendarattendee.RoleRequired:
	default:
		return ical.ITIPParticipant{}, false
	}
	status, ok := mailboxEntCalendarDeliveryRSVPStatus(attendee.RsvpStatus)
	if !ok {
		return ical.ITIPParticipant{}, false
	}
	return ical.ITIPParticipant{
		Email:       attendee.NormalizedEmail,
		DisplayName: mailboxEntCalendarAttendeeDisplayName(attendee),
		Role:        role,
		Status:      status,
	}, true
}

func mailboxEntCalendarDeliveryRSVPStatus(status calendarattendee.RsvpStatus) (ical.RSVPStatus, bool) {
	switch status {
	case calendarattendee.RsvpStatusNeedsAction:
		return ical.RSVPNeedsAction, true
	case calendarattendee.RsvpStatusAccepted:
		return ical.RSVPAccepted, true
	case calendarattendee.RsvpStatusDeclined:
		return ical.RSVPDeclined, true
	case calendarattendee.RsvpStatusTentative:
		return ical.RSVPTentative, true
	default:
		return "", false
	}
}

func mailboxEntCalendarDeliveryStableITIPStamp(createdAt time.Time) (time.Time, bool) {
	stamp := createdAt.UTC().Truncate(time.Second)
	if !mailboxEntValidTime(stamp) || stamp.Location() != time.UTC {
		return time.Time{}, false
	}
	return stamp, true
}

func mailboxEntCalendarDeliveryMasterITIPMessage(event *ent.CalendarEvent, method CalendarMailDeliveryMethod, organizerDisplayName, organizerEmail string, attendees []ical.ITIPParticipant, stamp time.Time) (ical.ITIPMessage, bool) {
	if event == nil || event.ID == uuid.Nil || event.OrganizationID == uuid.Nil || event.DeletedAt != nil || !mailboxEntValidTime(stamp) || stamp.Location() != time.UTC {
		return ical.ITIPMessage{}, false
	}
	itipMethod, ok := mailboxEntCalendarDeliveryITIPMethod(method)
	if !ok {
		return ical.ITIPMessage{}, false
	}
	status, ok := mailboxEntCalendarDeliveryITIPStatus(event.Status)
	if !ok || (method == CalendarMailDeliveryMethodCancel && status != ical.ITIPStatusCancelled) ||
		((method == CalendarMailDeliveryMethodRequest || method == CalendarMailDeliveryMethodReply) && status == ical.ITIPStatusCancelled) {
		return ical.ITIPMessage{}, false
	}
	temporal, exdates, ok := mailboxEntCalendarDeliveryMasterTemporal(event)
	if !ok {
		return ical.ITIPMessage{}, false
	}
	message := ical.ITIPMessage{
		Method:      itipMethod,
		UID:         event.UID,
		Stamp:       stamp,
		Sequence:    event.Sequence,
		Status:      status,
		Summary:     event.Title,
		Description: event.Description,
		Location:    event.Location,
		Temporal:    temporal,
		Organizer:  ical.ITIPOrganizer{Email: organizerEmail, DisplayName: organizerDisplayName},
		Attendees:  append([]ical.ITIPParticipant(nil), attendees...),
		RRULE:      mailboxEntCalendarEventRRULE(event),
		EXDates:    exdates,
	}
	normalized, err := ical.NormalizeITIPMessage(message)
	if err != nil {
		return ical.ITIPMessage{}, false
	}
	identity, err := ical.CanonicalITIPIdentity(normalized)
	if err != nil || identity.Method != itipMethod || identity.UID != normalized.UID || identity.Sequence != event.Sequence ||
		identity.HasRecurrenceID || identity.RecurrenceID != "" {
		return ical.ITIPMessage{}, false
	}
	return calendarMailDeliveryITIPMessageCopy(normalized), true
}

func mailboxEntCalendarDeliveryRemovedAttendeeMasterCancelITIPMessage(event *ent.CalendarEvent, organizerDisplayName, organizerEmail string, targetParticipant ical.ITIPParticipant, stamp time.Time) (ical.ITIPMessage, bool) {
	if event == nil || event.ID == uuid.Nil || event.OrganizationID == uuid.Nil || event.DeletedAt != nil || !mailboxEntValidTime(stamp) || stamp.Location() != time.UTC ||
		targetParticipant.Email == "" || targetParticipant.RSVP ||
		(event.Status != calendarevent.StatusConfirmed && event.Status != calendarevent.StatusTentative) {
		return ical.ITIPMessage{}, false
	}
	temporal, exdates, ok := mailboxEntCalendarDeliveryMasterTemporal(event)
	if !ok {
		return ical.ITIPMessage{}, false
	}
	message := ical.ITIPMessage{
		Method:      ical.ITIPMethodCancel,
		UID:         event.UID,
		Stamp:       stamp,
		Sequence:    event.Sequence,
		Status:      ical.ITIPStatusCancelled,
		Summary:     event.Title,
		Description: event.Description,
		Location:    event.Location,
		Temporal:    temporal,
		Organizer:   ical.ITIPOrganizer{Email: organizerEmail, DisplayName: organizerDisplayName},
		Attendees:   []ical.ITIPParticipant{targetParticipant},
		RRULE:       mailboxEntCalendarEventRRULE(event),
		EXDates:     exdates,
	}
	normalized, err := ical.NormalizeITIPMessage(message)
	if err != nil {
		return ical.ITIPMessage{}, false
	}
	identity, err := ical.CanonicalITIPIdentity(normalized)
	if err != nil || identity.Method != ical.ITIPMethodCancel || identity.UID != normalized.UID || identity.Sequence != event.Sequence ||
		identity.HasRecurrenceID || identity.RecurrenceID != "" || normalized.Status != ical.ITIPStatusCancelled ||
		len(normalized.Attendees) != 1 || normalized.Attendees[0].Email != targetParticipant.Email || normalized.Attendees[0].RSVP {
		return ical.ITIPMessage{}, false
	}
	return calendarMailDeliveryITIPMessageCopy(normalized), true
}

func mailboxEntCalendarDeliveryOccurrenceCancelITIPMessage(event *ent.CalendarEvent, recurrenceID, organizerDisplayName, organizerEmail string, attendees []ical.ITIPParticipant, stamp time.Time) (ical.ITIPMessage, bool) {
	if event == nil || event.ID == uuid.Nil || event.OrganizationID == uuid.Nil || event.DeletedAt != nil || recurrenceID == "" ||
		!mailboxEntValidTime(stamp) || stamp.Location() != time.UTC || (event.Status != calendarevent.StatusConfirmed && event.Status != calendarevent.StatusTentative) || len(attendees) == 0 {
		return ical.ITIPMessage{}, false
	}
	master, exdates, ok := mailboxEntCalendarDeliveryRecurrenceMaster(event)
	if !ok || mailboxEntCalendarDeliveryRecurrenceExcluded(master.AllDay, exdates, recurrenceID) {
		return ical.ITIPMessage{}, false
	}
	nominal, ok := ical.ResolveNominalOccurrence(master, mailboxEntCalendarEventRRULE(event), recurrenceID)
	if !ok || nominal.RecurrenceID != recurrenceID || nominal.Temporal.AllDay != master.AllDay {
		return ical.ITIPMessage{}, false
	}
	temporal, recurrence, ok := mailboxEntCalendarDeliveryOccurrenceITIPTemporal(nominal)
	if !ok {
		return ical.ITIPMessage{}, false
	}
	message := ical.ITIPMessage{
		Method:       ical.ITIPMethodCancel,
		UID:          event.UID,
		Stamp:        stamp,
		Sequence:     event.Sequence,
		Status:       ical.ITIPStatusCancelled,
		Summary:      event.Title,
		Description:  event.Description,
		Location:     event.Location,
		Temporal:     temporal,
		Organizer:    ical.ITIPOrganizer{Email: organizerEmail, DisplayName: organizerDisplayName},
		Attendees:    append([]ical.ITIPParticipant(nil), attendees...),
		RRULE:        "",
		EXDates:      nil,
		RecurrenceID: &recurrence,
	}
	normalized, err := ical.NormalizeITIPMessage(message)
	if err != nil {
		return ical.ITIPMessage{}, false
	}
	identity, err := ical.CanonicalITIPIdentity(normalized)
	if err != nil || identity.Method != ical.ITIPMethodCancel || identity.UID != normalized.UID || identity.Sequence != event.Sequence ||
		!identity.HasRecurrenceID || identity.RecurrenceID != recurrenceID || normalized.Status != ical.ITIPStatusCancelled ||
		normalized.RRULE != "" || len(normalized.EXDates) != 0 || normalized.RecurrenceID == nil || !mailboxEntCalendarDeliveryRecurrenceIDMatches(*normalized.RecurrenceID, recurrence) ||
		len(normalized.Attendees) != len(attendees) || !mailboxEntCalendarDeliveryOccurrenceTemporalMatches(normalized.Temporal, temporal) {
		return ical.ITIPMessage{}, false
	}
	for _, attendee := range normalized.Attendees {
		if attendee.RSVP {
			return ical.ITIPMessage{}, false
		}
	}
	return calendarMailDeliveryITIPMessageCopy(normalized), true
}

// mailboxEntCalendarReplyOccurrenceITIPMessage emits one attendee's current
// PARTSTAT for an external organizer. It deliberately carries neither the
// master RRULE nor EXDATEs: RECURRENCE-ID identifies the immutable nominal
// slot while temporal fields describe that slot's effective override.
func mailboxEntCalendarReplyOccurrenceITIPMessage(event *ent.CalendarEvent, occurrence mailboxEntCalendarReplyOccurrenceFacts, organizerEmail string, attendee ical.ITIPParticipant, stamp time.Time) (ical.ITIPMessage, bool) {
	if event == nil || event.ID == uuid.Nil || event.OrganizationID == uuid.Nil || event.DeletedAt != nil || event.OrganizerID != nil ||
		occurrence.superseded || occurrence.sequence < 0 || attendee.Email == "" || attendee.RSVP || !mailboxEntValidTime(stamp) || stamp.Location() != time.UTC {
		return ical.ITIPMessage{}, false
	}
	status, ok := mailboxEntCalendarReplyOccurrenceStatus(occurrence.status)
	if !ok || attendee.Status == ical.RSVPNeedsAction {
		return ical.ITIPMessage{}, false
	}
	itipStatus, ok := mailboxEntCalendarDeliveryITIPStatus(status)
	if !ok {
		return ical.ITIPMessage{}, false
	}
	if occurrence.recurrence.Date == "" && occurrence.recurrence.Time.IsZero() {
		return ical.ITIPMessage{}, false
	}
	message := ical.ITIPMessage{
		Method:       ical.ITIPMethodReply,
		UID:          event.UID,
		Stamp:        stamp,
		Sequence:     occurrence.sequence,
		Status:       itipStatus,
		Summary:      event.Title,
		Description:  event.Description,
		Location:     event.Location,
		Temporal:     occurrence.temporal,
		Organizer:    ical.ITIPOrganizer{Email: organizerEmail},
		Attendees:    []ical.ITIPParticipant{attendee},
		RecurrenceID: &occurrence.recurrence,
	}
	normalized, err := ical.NormalizeITIPMessage(message)
	if err != nil {
		return ical.ITIPMessage{}, false
	}
	identity, err := ical.CanonicalITIPIdentity(normalized)
	if err != nil || identity.Method != ical.ITIPMethodReply || identity.UID != normalized.UID || identity.Sequence != occurrence.sequence ||
		!identity.HasRecurrenceID || normalized.RecurrenceID == nil || !mailboxEntCalendarDeliveryRecurrenceIDMatches(*normalized.RecurrenceID, occurrence.recurrence) ||
		len(normalized.Attendees) != 1 || normalized.Attendees[0].Email != attendee.Email || normalized.Attendees[0].RSVP ||
		normalized.RRULE != "" || len(normalized.EXDates) != 0 || !mailboxEntCalendarDeliveryOccurrenceTemporalMatches(normalized.Temporal, occurrence.temporal) {
		return ical.ITIPMessage{}, false
	}
	return calendarMailDeliveryITIPMessageCopy(normalized), true
}

func mailboxEntCalendarDeliveryRecurrenceMaster(event *ent.CalendarEvent) (ical.CalendarEvent, []ical.ITIPRecurrenceID, bool) {
	temporal, exdates, ok := mailboxEntCalendarDeliveryMasterTemporal(event)
	if !ok {
		return ical.CalendarEvent{}, nil, false
	}
	if temporal.AllDay != nil && temporal.Timed == nil {
		master := ical.CalendarEvent{
			AllDay:           true,
			StartDate:        temporal.AllDay.StartDate,
			EndDateExclusive: temporal.AllDay.EndDateExclusive,
			Timezone:         temporal.AllDay.Timezone,
		}
		return master, exdates, master.Validate() == nil
	}
	if temporal.Timed != nil && temporal.AllDay == nil {
		master := ical.CalendarEvent{
			StartsAt: temporal.Timed.StartsAt,
			EndsAt:   temporal.Timed.EndsAt,
			Timezone: temporal.Timed.Timezone,
		}
		return master, exdates, master.Validate() == nil
	}
	return ical.CalendarEvent{}, nil, false
}

func mailboxEntCalendarDeliveryRecurrenceExcluded(allDay bool, exdates []ical.ITIPRecurrenceID, recurrenceID string) bool {
	for _, exdate := range exdates {
		if allDay && exdate.Date == recurrenceID {
			return true
		}
		if !allDay && exdate.Time.Format("2006-01-02T15:04:05Z") == recurrenceID {
			return true
		}
	}
	return false
}

func mailboxEntCalendarDeliveryOccurrenceITIPTemporal(nominal ical.NominalOccurrence) (ical.ITIPTemporal, ical.ITIPRecurrenceID, bool) {
	if nominal.RecurrenceID == "" || nominal.Temporal.Validate() != nil {
		return ical.ITIPTemporal{}, ical.ITIPRecurrenceID{}, false
	}
	if nominal.Temporal.AllDay {
		return ical.ITIPTemporal{AllDay: &ical.ITIPAllDay{
			StartDate:        nominal.Temporal.StartDate,
			EndDateExclusive: nominal.Temporal.EndDateExclusive,
			Timezone:         nominal.Temporal.Timezone,
		}}, ical.ITIPRecurrenceID{Date: nominal.RecurrenceID}, true
	}
	return ical.ITIPTemporal{Timed: &ical.ITIPTimed{
		StartsAt: nominal.Temporal.StartsAt,
		EndsAt:   nominal.Temporal.EndsAt,
		Timezone: nominal.Temporal.Timezone,
	}}, ical.ITIPRecurrenceID{Time: nominal.Temporal.StartsAt}, true
}

func mailboxEntCalendarDeliveryRecurrenceIDMatches(actual, expected ical.ITIPRecurrenceID) bool {
	if actual.Date != "" || expected.Date != "" {
		return actual.Date == expected.Date && actual.Time.IsZero() && expected.Time.IsZero()
	}
	return actual.Time.Equal(expected.Time) && actual.Time.Location() == time.UTC && expected.Time.Location() == time.UTC
}

func mailboxEntCalendarDeliveryOccurrenceTemporalMatches(actual, expected ical.ITIPTemporal) bool {
	if actual.AllDay != nil || expected.AllDay != nil {
		return actual.Timed == nil && expected.Timed == nil && actual.AllDay != nil && expected.AllDay != nil &&
			actual.AllDay.StartDate == expected.AllDay.StartDate && actual.AllDay.EndDateExclusive == expected.AllDay.EndDateExclusive && actual.AllDay.Timezone == expected.AllDay.Timezone
	}
	return actual.AllDay == nil && expected.AllDay == nil && actual.Timed != nil && expected.Timed != nil &&
		actual.Timed.StartsAt.Equal(expected.Timed.StartsAt) && actual.Timed.EndsAt.Equal(expected.Timed.EndsAt) && actual.Timed.Timezone == expected.Timed.Timezone
}

func mailboxEntCalendarDeliveryITIPMethod(method CalendarMailDeliveryMethod) (ical.ITIPMethod, bool) {
	switch method {
	case CalendarMailDeliveryMethodRequest:
		return ical.ITIPMethodRequest, true
	case CalendarMailDeliveryMethodCancel:
		return ical.ITIPMethodCancel, true
	case CalendarMailDeliveryMethodReply:
		return ical.ITIPMethodReply, true
	default:
		return "", false
	}
}

func mailboxEntCalendarDeliveryITIPStatus(status calendarevent.Status) (ical.ITIPStatus, bool) {
	switch status {
	case calendarevent.StatusConfirmed:
		return ical.ITIPStatusConfirmed, true
	case calendarevent.StatusTentative:
		return ical.ITIPStatusTentative, true
	case calendarevent.StatusCancelled:
		return ical.ITIPStatusCancelled, true
	default:
		return "", false
	}
}

func mailboxEntCalendarDeliveryMasterTemporal(event *ent.CalendarEvent) (ical.ITIPTemporal, []ical.ITIPRecurrenceID, bool) {
	if event == nil {
		return ical.ITIPTemporal{}, nil, false
	}
	var temporal ical.ITIPTemporal
	if event.AllDay {
		if event.StartDate == nil || event.EndDateExclusive == nil || event.StartsAt != nil || event.EndsAt != nil {
			return ical.ITIPTemporal{}, nil, false
		}
		temporal.AllDay = &ical.ITIPAllDay{
			StartDate:        event.StartDate.Format("2006-01-02"),
			EndDateExclusive: event.EndDateExclusive.Format("2006-01-02"),
			Timezone:         event.Timezone,
		}
	} else {
		if event.StartsAt == nil || event.EndsAt == nil || event.StartDate != nil || event.EndDateExclusive != nil {
			return ical.ITIPTemporal{}, nil, false
		}
		temporal.Timed = &ical.ITIPTimed{
			StartsAt: *event.StartsAt,
			EndsAt:   *event.EndsAt,
			Timezone: event.Timezone,
		}
	}

	exdates := make([]ical.ITIPRecurrenceID, 0, len(event.Exdates))
	for _, value := range event.Exdates {
		if temporal.AllDay != nil {
			parsed, err := ical.ParseCalendarDate(value)
			if err != nil || parsed.Format("2006-01-02") != value {
				return ical.ITIPTemporal{}, nil, false
			}
			exdates = append(exdates, ical.ITIPRecurrenceID{Date: value})
			continue
		}
		parsed, err := time.Parse("2006-01-02T15:04:05Z", value)
		if err != nil || parsed.Location() != time.UTC || parsed.Format("2006-01-02T15:04:05Z") != value {
			return ical.ITIPTemporal{}, nil, false
		}
		exdates = append(exdates, ical.ITIPRecurrenceID{Time: parsed})
	}
	return temporal, exdates, true
}

func mailboxEntCalendarEventRRULE(event *ent.CalendarEvent) string {
	if event == nil || event.Rrule == nil {
		return ""
	}
	return *event.Rrule
}

func mailboxEntCalendarAttendeeDisplayName(attendee *ent.CalendarAttendee) string {
	if attendee == nil || attendee.DisplayName == nil {
		return ""
	}
	return *attendee.DisplayName
}

func mailboxEntCalendarUserDisplayName(sourceUser *ent.User) string {
	if sourceUser == nil {
		return ""
	}
	if sourceUser.DisplayName != "" {
		return sourceUser.DisplayName
	}
	return sourceUser.Name
}

func mailboxEntCalendarInviteDeliveryClaim(delivery *ent.CalendarInviteDelivery, deliveryID CalendarMailDeliveryID, snapshot mailboxEntCalendarDeliveryClaimSnapshot, token CalendarMailDeliveryClaimToken, generation CalendarMailDeliveryClaimGeneration, leaseUntil time.Time) (CalendarMailDeliveryClaim, error) {
	inviteID, err := NewCalendarInviteDeliveryID(delivery.ID)
	if err != nil || deliveryID.Kind() != CalendarMailDeliveryKindInvite || snapshot.organizerID == uuid.Nil || snapshot.attendeeVersion != 0 ||
		(delivery.RecurrenceID == nil && snapshot.recurrenceID != "") || (delivery.RecurrenceID != nil && snapshot.recurrenceID != *delivery.RecurrenceID) {
		return CalendarMailDeliveryClaim{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	claim := CalendarMailDeliveryClaim{
		deliveryID:      deliveryID,
		inviteID:        inviteID,
		kind:            CalendarMailDeliveryKindInvite,
		method:          snapshot.method,
		state:           CalendarMailDeliveryStateSubmitting,
		claimToken:      token,
		claimGeneration: generation,
		organizationID:  snapshot.organizationID,
		leaseUntil:      leaseUntil,
		eventID:         snapshot.eventID,
		attendeeID:      snapshot.attendeeID,
		organizerID:     snapshot.organizerID,
		accountID:       snapshot.accountID,
		senderActorID:   snapshot.senderActorID,
		fromIdentityID:  snapshot.fromIdentityID,
		recipient:       snapshot.recipient,
		materialization: CalendarMailDeliveryMaterializationInput{
			deliveryID:    deliveryID,
			kind:          CalendarMailDeliveryKindInvite,
			method:        snapshot.method,
			eventID:       snapshot.eventID,
			attendeeID:    snapshot.attendeeID,
			organizerID:   snapshot.organizerID,
			eventSequence: snapshot.eventSequence,
			recurrenceID:   snapshot.recurrenceID,
			message:       snapshot.message,
		},
	}
	if !validCalendarMailDeliveryClaim(claim) {
		return CalendarMailDeliveryClaim{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return calendarMailDeliveryClaimCopy(claim), nil
}

func mailboxEntCalendarReplyDeliveryClaim(delivery *ent.CalendarReplyDelivery, deliveryID CalendarMailDeliveryID, snapshot mailboxEntCalendarDeliveryClaimSnapshot, token CalendarMailDeliveryClaimToken, generation CalendarMailDeliveryClaimGeneration, leaseUntil time.Time) (CalendarMailDeliveryClaim, error) {
	replyID, err := NewCalendarReplyDeliveryID(delivery.ID)
	if err != nil || deliveryID.Kind() != CalendarMailDeliveryKindReply || snapshot.organizerID != uuid.Nil || snapshot.attendeeVersion != delivery.AttendeeVersion ||
		(delivery.RecurrenceID == nil && snapshot.recurrenceID != "") || (delivery.RecurrenceID != nil && snapshot.recurrenceID != *delivery.RecurrenceID) {
		return CalendarMailDeliveryClaim{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	claim := CalendarMailDeliveryClaim{
		deliveryID:      deliveryID,
		replyID:         replyID,
		kind:            CalendarMailDeliveryKindReply,
		method:          CalendarMailDeliveryMethodReply,
		state:           CalendarMailDeliveryStateSubmitting,
		claimToken:      token,
		claimGeneration: generation,
		organizationID:  snapshot.organizationID,
		leaseUntil:      leaseUntil,
		eventID:         snapshot.eventID,
		attendeeID:      snapshot.attendeeID,
		accountID:       snapshot.accountID,
		senderActorID:   snapshot.senderActorID,
		fromIdentityID:  snapshot.fromIdentityID,
		recipient:       snapshot.recipient,
		materialization: CalendarMailDeliveryMaterializationInput{
			deliveryID:      deliveryID,
			kind:            CalendarMailDeliveryKindReply,
			method:          CalendarMailDeliveryMethodReply,
			eventID:         snapshot.eventID,
			attendeeID:      snapshot.attendeeID,
			eventSequence:   snapshot.eventSequence,
			attendeeVersion: snapshot.attendeeVersion,
			recurrenceID:    snapshot.recurrenceID,
			message:         snapshot.message,
		},
	}
	if !validCalendarMailDeliveryClaim(claim) {
		return CalendarMailDeliveryClaim{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return calendarMailDeliveryClaimCopy(claim), nil
}

func mailboxEntCASSupersedeCalendarInviteDelivery(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery, databaseNow time.Time, reclaim bool) (int, error) {
	if ctx == nil || transaction == nil || delivery == nil || !mailboxEntValidTime(databaseNow) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if reclaim && (delivery.LeaseToken == nil || delivery.LeaseUntil == nil) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if !reclaim && delivery.NextAttemptAt == nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	update := transaction.CalendarInviteDelivery.Update().Where(
		calendarinvitedelivery.IDEQ(delivery.ID),
		calendarinvitedelivery.OrganizationIDEQ(delivery.OrganizationID),
		calendarinvitedelivery.CalendarEventIDEQ(delivery.CalendarEventID),
		calendarinvitedelivery.CalendarAttendeeIDEQ(delivery.CalendarAttendeeID),
		calendarinvitedelivery.KindEQ(delivery.Kind),
		calendarinvitedelivery.EventSequenceEQ(delivery.EventSequence),
		calendarinvitedelivery.MailSubmissionIDIsNil(),
		calendarinvitedelivery.AttemptsEQ(delivery.Attempts),
		calendarinvitedelivery.ClaimGenerationEQ(delivery.ClaimGeneration),
		calendarinvitedelivery.TerminalAtIsNil(),
	)
	if delivery.RecurrenceID == nil {
		update = update.Where(calendarinvitedelivery.RecurrenceIDIsNil())
	} else {
		update = update.Where(calendarinvitedelivery.RecurrenceIDEQ(*delivery.RecurrenceID))
	}
	if reclaim {
		update = update.Where(
			calendarinvitedelivery.StateEQ(calendarinvitedelivery.StateSubmitting),
			calendarinvitedelivery.LeaseTokenEQ(*delivery.LeaseToken),
			calendarinvitedelivery.LeaseUntilEQ(*delivery.LeaseUntil),
			calendarinvitedelivery.NextAttemptAtIsNil(),
			calendarinvitedelivery.LastErrorCodeIsNil(),
		)
	} else {
		update = update.Where(
			calendarinvitedelivery.StateEQ(calendarinvitedelivery.StatePending),
			calendarinvitedelivery.LeaseTokenIsNil(),
			calendarinvitedelivery.LeaseUntilIsNil(),
			calendarinvitedelivery.NextAttemptAtEQ(*delivery.NextAttemptAt),
		)
		if delivery.LastErrorCode == nil {
			update = update.Where(calendarinvitedelivery.LastErrorCodeIsNil())
		} else {
			update = update.Where(calendarinvitedelivery.LastErrorCodeEQ(*delivery.LastErrorCode))
		}
	}
	updated, err := update.
		SetState(calendarinvitedelivery.StateSuperseded).
		ClearLeaseToken().
		ClearLeaseUntil().
		ClearNextAttemptAt().
		ClearLastErrorCode().
		SetTerminalAt(databaseNow).
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return updated, nil
}

func mailboxEntCASSupersedeCalendarReplyDelivery(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarReplyDelivery, databaseNow time.Time, reclaim bool) (int, error) {
	if ctx == nil || transaction == nil || delivery == nil || !mailboxEntValidTime(databaseNow) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if reclaim && (delivery.LeaseToken == nil || delivery.LeaseUntil == nil) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if !reclaim && delivery.NextAttemptAt == nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	update := transaction.CalendarReplyDelivery.Update().Where(
		calendarreplydelivery.IDEQ(delivery.ID),
		calendarreplydelivery.OrganizationIDEQ(delivery.OrganizationID),
		calendarreplydelivery.CalendarEventIDEQ(delivery.CalendarEventID),
		calendarreplydelivery.CalendarAttendeeIDEQ(delivery.CalendarAttendeeID),
		calendarreplydelivery.EventSequenceEQ(delivery.EventSequence),
		calendarreplydelivery.MailSubmissionIDIsNil(),
		calendarreplydelivery.AttemptsEQ(delivery.Attempts),
		calendarreplydelivery.ClaimGenerationEQ(delivery.ClaimGeneration),
	)
	if delivery.RecurrenceID == nil {
		update = update.Where(calendarreplydelivery.RecurrenceIDIsNil())
	} else {
		update = update.Where(calendarreplydelivery.RecurrenceIDEQ(*delivery.RecurrenceID))
	}
	if reclaim {
		update = update.Where(
			calendarreplydelivery.StateEQ(calendarreplydelivery.StateSubmitting),
			calendarreplydelivery.LeaseTokenEQ(*delivery.LeaseToken),
			calendarreplydelivery.LeaseUntilEQ(*delivery.LeaseUntil),
			calendarreplydelivery.NextAttemptAtIsNil(),
			calendarreplydelivery.LastErrorCodeIsNil(),
		)
	} else {
		update = update.Where(
			calendarreplydelivery.StateEQ(calendarreplydelivery.StatePending),
			calendarreplydelivery.LeaseTokenIsNil(),
			calendarreplydelivery.LeaseUntilIsNil(),
			calendarreplydelivery.NextAttemptAtEQ(*delivery.NextAttemptAt),
		)
		if delivery.LastErrorCode == nil {
			update = update.Where(calendarreplydelivery.LastErrorCodeIsNil())
		} else {
			update = update.Where(calendarreplydelivery.LastErrorCodeEQ(*delivery.LastErrorCode))
		}
	}
	updated, err := update.
		SetState(calendarreplydelivery.StateSuperseded).
		ClearLeaseToken().
		ClearLeaseUntil().
		ClearNextAttemptAt().
		ClearLastErrorCode().
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return updated, nil
}

func mailboxEntCASCalendarInviteDeliveryClaim(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery, databaseNow time.Time, token CalendarMailDeliveryClaimToken, generation CalendarMailDeliveryClaimGeneration, leaseUntil time.Time, reclaim bool) (int, error) {
	if ctx == nil || transaction == nil || delivery == nil || uuid.UUID(token) == uuid.Nil || generation < 1 || !mailboxEntValidTime(databaseNow) ||
		!validCalendarMailDeliveryLeaseUntil(leaseUntil) || !leaseUntil.After(databaseNow) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if reclaim && (delivery.LeaseToken == nil || delivery.LeaseUntil == nil) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if !reclaim && delivery.NextAttemptAt == nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	update := transaction.CalendarInviteDelivery.Update().Where(
		calendarinvitedelivery.IDEQ(delivery.ID),
		calendarinvitedelivery.OrganizationIDEQ(delivery.OrganizationID),
		calendarinvitedelivery.CalendarEventIDEQ(delivery.CalendarEventID),
		calendarinvitedelivery.CalendarAttendeeIDEQ(delivery.CalendarAttendeeID),
		calendarinvitedelivery.MailSubmissionIDIsNil(),
		calendarinvitedelivery.AttemptsEQ(delivery.Attempts),
		calendarinvitedelivery.ClaimGenerationEQ(delivery.ClaimGeneration),
		calendarinvitedelivery.TerminalAtIsNil(),
	)
	if delivery.RecurrenceID == nil {
		update = update.Where(calendarinvitedelivery.RecurrenceIDIsNil())
	} else {
		update = update.Where(calendarinvitedelivery.RecurrenceIDEQ(*delivery.RecurrenceID))
	}
	if reclaim {
		update = update.Where(
			calendarinvitedelivery.StateEQ(calendarinvitedelivery.StateSubmitting),
			calendarinvitedelivery.LeaseTokenEQ(*delivery.LeaseToken),
			calendarinvitedelivery.LeaseUntilEQ(*delivery.LeaseUntil),
			calendarinvitedelivery.NextAttemptAtIsNil(),
			calendarinvitedelivery.LastErrorCodeIsNil(),
		)
	} else {
		update = update.Where(
			calendarinvitedelivery.StateEQ(calendarinvitedelivery.StatePending),
			calendarinvitedelivery.LeaseTokenIsNil(),
			calendarinvitedelivery.LeaseUntilIsNil(),
			calendarinvitedelivery.NextAttemptAtEQ(*delivery.NextAttemptAt),
		)
		if delivery.LastErrorCode == nil {
			update = update.Where(calendarinvitedelivery.LastErrorCodeIsNil())
		} else {
			update = update.Where(calendarinvitedelivery.LastErrorCodeEQ(*delivery.LastErrorCode))
		}
	}
	if !reclaim {
		update.SetAttempts(delivery.Attempts + 1)
	}
	updated, err := update.
		SetState(calendarinvitedelivery.StateSubmitting).
		SetClaimGeneration(int64(generation)).
		SetLeaseToken(uuid.UUID(token)).
		SetLeaseUntil(leaseUntil).
		ClearNextAttemptAt().
		ClearLastErrorCode().
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return updated, nil
}

func mailboxEntCASCalendarReplyDeliveryClaim(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarReplyDelivery, databaseNow time.Time, token CalendarMailDeliveryClaimToken, generation CalendarMailDeliveryClaimGeneration, leaseUntil time.Time, reclaim bool) (int, error) {
	if ctx == nil || transaction == nil || delivery == nil || uuid.UUID(token) == uuid.Nil || generation < 1 || !mailboxEntValidTime(databaseNow) ||
		!validCalendarMailDeliveryLeaseUntil(leaseUntil) || !leaseUntil.After(databaseNow) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if reclaim && (delivery.LeaseToken == nil || delivery.LeaseUntil == nil) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if !reclaim && delivery.NextAttemptAt == nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	update := transaction.CalendarReplyDelivery.Update().Where(
		calendarreplydelivery.IDEQ(delivery.ID),
		calendarreplydelivery.OrganizationIDEQ(delivery.OrganizationID),
		calendarreplydelivery.CalendarEventIDEQ(delivery.CalendarEventID),
		calendarreplydelivery.CalendarAttendeeIDEQ(delivery.CalendarAttendeeID),
		calendarreplydelivery.MailSubmissionIDIsNil(),
		calendarreplydelivery.AttemptsEQ(delivery.Attempts),
		calendarreplydelivery.ClaimGenerationEQ(delivery.ClaimGeneration),
	)
	if delivery.RecurrenceID == nil {
		update = update.Where(calendarreplydelivery.RecurrenceIDIsNil())
	} else {
		update = update.Where(calendarreplydelivery.RecurrenceIDEQ(*delivery.RecurrenceID))
	}
	if reclaim {
		update = update.Where(
			calendarreplydelivery.StateEQ(calendarreplydelivery.StateSubmitting),
			calendarreplydelivery.LeaseTokenEQ(*delivery.LeaseToken),
			calendarreplydelivery.LeaseUntilEQ(*delivery.LeaseUntil),
			calendarreplydelivery.NextAttemptAtIsNil(),
			calendarreplydelivery.LastErrorCodeIsNil(),
		)
	} else {
		update = update.Where(
			calendarreplydelivery.StateEQ(calendarreplydelivery.StatePending),
			calendarreplydelivery.LeaseTokenIsNil(),
			calendarreplydelivery.LeaseUntilIsNil(),
			calendarreplydelivery.NextAttemptAtEQ(*delivery.NextAttemptAt),
		)
		if delivery.LastErrorCode == nil {
			update = update.Where(calendarreplydelivery.LastErrorCodeIsNil())
		} else {
			update = update.Where(calendarreplydelivery.LastErrorCodeEQ(*delivery.LastErrorCode))
		}
	}
	if !reclaim {
		update.SetAttempts(delivery.Attempts + 1)
	}
	updated, err := update.
		SetState(calendarreplydelivery.StateSubmitting).
		SetClaimGeneration(int64(generation)).
		SetLeaseToken(uuid.UUID(token)).
		SetLeaseUntil(leaseUntil).
		ClearNextAttemptAt().
		ClearLastErrorCode().
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return updated, nil
}

func mailboxEntCalendarInviteDeliveryPendingValid(delivery *ent.CalendarInviteDelivery) bool {
	if delivery == nil || CalendarMailDeliveryState(delivery.State) != CalendarMailDeliveryStatePending || delivery.Attempts < 0 ||
		delivery.Attempts > calendarMailDeliveryMaximumAttempts || delivery.ClaimGeneration < 0 || delivery.MailSubmissionID != nil ||
		delivery.LeaseToken != nil || delivery.LeaseUntil != nil || delivery.NextAttemptAt == nil || !mailboxEntCalendarDeliveryTimestampValid(*delivery.NextAttemptAt) ||
		delivery.TerminalAt != nil {
		return false
	}
	if delivery.Attempts == 0 {
		return delivery.LastErrorCode == nil
	}
	return delivery.LastErrorCode != nil && mailboxEntCalendarInviteDeliveryRetryCodeValid(string(*delivery.LastErrorCode))
}

func mailboxEntCalendarInviteDeliveryRetryCodeValid(code string) bool {
	switch code {
	case "mailbox_unavailable", "engine_unavailable", "internal_error":
		return true
	default:
		return false
	}
}

func mailboxEntCalendarInviteDeliverySubmittingValid(delivery *ent.CalendarInviteDelivery) bool {
	return delivery != nil && CalendarMailDeliveryState(delivery.State) == CalendarMailDeliveryStateSubmitting &&
		delivery.Attempts >= 1 && delivery.Attempts <= calendarMailDeliveryMaximumAttempts && delivery.ClaimGeneration >= 1 &&
		delivery.MailSubmissionID == nil && delivery.LeaseToken != nil && *delivery.LeaseToken != uuid.Nil &&
		delivery.LeaseUntil != nil && validCalendarMailDeliveryLeaseUntil(*delivery.LeaseUntil) &&
		delivery.NextAttemptAt == nil && delivery.LastErrorCode == nil && delivery.TerminalAt == nil
}

func mailboxEntCalendarInviteDeliveryUnknownBoundValid(delivery *ent.CalendarInviteDelivery) bool {
	return delivery != nil && CalendarMailDeliveryState(delivery.State) == CalendarMailDeliveryStateUnknown &&
		delivery.Attempts >= 1 && delivery.Attempts <= calendarMailDeliveryMaximumAttempts &&
		delivery.MailSubmissionID != nil && *delivery.MailSubmissionID != uuid.Nil && delivery.LeaseToken == nil && delivery.LeaseUntil == nil &&
		delivery.NextAttemptAt != nil && mailboxEntCalendarDeliveryTimestampValid(*delivery.NextAttemptAt) && delivery.LastErrorCode != nil &&
		delivery.TerminalAt == nil
}

func mailboxEntCalendarInviteDeliveryTerminalValid(delivery *ent.CalendarInviteDelivery) bool {
	if delivery == nil || delivery.LeaseToken != nil || delivery.LeaseUntil != nil || delivery.TerminalAt == nil ||
		!mailboxEntCalendarDeliveryTimestampValid(*delivery.TerminalAt) {
		return false
	}
	switch CalendarMailDeliveryState(delivery.State) {
	case CalendarMailDeliveryStateSent:
		return delivery.Attempts >= 1 && delivery.Attempts <= calendarMailDeliveryMaximumAttempts &&
			delivery.MailSubmissionID != nil && *delivery.MailSubmissionID != uuid.Nil && delivery.NextAttemptAt == nil && delivery.LastErrorCode == nil
	case CalendarMailDeliveryStateSuperseded:
		return delivery.MailSubmissionID == nil && delivery.NextAttemptAt == nil && delivery.LastErrorCode == nil
	case CalendarMailDeliveryStateFailedPermanent:
		return delivery.Attempts >= 1 && delivery.Attempts <= calendarMailDeliveryMaximumAttempts &&
			delivery.MailSubmissionID == nil && delivery.NextAttemptAt == nil && delivery.LastErrorCode != nil
	default:
		return false
	}
}
