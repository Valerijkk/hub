//go:build qa && !release

package qamailcalendar

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarinboundreplyreceipt"
	"github.com/valerii/hub/backend/ent/predicate"
)

// inboundRequestFixturePublishedCalendarAttendeeObservation is the private,
// comparable C3c result. Later Calendar shards own reply delivery, mutation
// receipt, external, Outbox, lifecycle, and delete observations.
type inboundRequestFixturePublishedCalendarAttendeeObservation struct {
	inboundMessageReceiptID uuid.UUID
	inboundObjectStateID    uuid.UUID
	eventID                 uuid.UUID
	attendeeID              uuid.UUID
}

func (observation inboundRequestFixturePublishedCalendarAttendeeObservation) valid() bool {
	return observation.inboundMessageReceiptID != uuid.Nil &&
		observation.inboundObjectStateID != uuid.Nil &&
		observation.eventID != uuid.Nil &&
		observation.attendeeID != uuid.Nil
}

// observeInboundRequestFixturePublishedCalendarAttendee revalidates C3b,
// then locks the sole owner attendee and proves that no inbound REPLY receipt
// exists inside the caller-owned transaction. It returns copied UUIDs only;
// it never mutates, commits, or retains Ent rows.
func observeInboundRequestFixturePublishedCalendarAttendee(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	authority inboundRequestFixturePublishedEngineAuthority,
	lockedRoots inboundRequestFixtureLockedPublishedRoots,
	mailClosure inboundRequestFixturePublishedMailClosure,
) (inboundRequestFixturePublishedCalendarAttendeeObservation, bool) {
	roots, valid := newInboundRequestFixtureRootProof(plan, persisted)
	if ctx == nil ||
		tx == nil ||
		!valid ||
		!inboundRequestFixturePlanIsCanonical(plan) ||
		!authority.valid() ||
		authority.nonce != plan.nonce ||
		lockedRoots.roots != roots ||
		!inboundRequestFixtureLockedPublishedRootsValid(
			roots,
			plan,
			persisted,
			authority,
			lockedRoots.mail,
			lockedRoots.account,
			lockedRoots.owner,
		) ||
		!mailClosure.valid() ||
		mailClosure.roots != roots ||
		!mailClosure.falseSent.matchesRoots(roots) {
		return inboundRequestFixturePublishedCalendarAttendeeObservation{}, false
	}

	event, observed := observeInboundRequestFixturePublishedCalendarEvent(
		ctx,
		tx,
		plan,
		persisted,
		authority,
		lockedRoots,
		mailClosure,
	)
	if !observed || !event.valid() {
		return inboundRequestFixturePublishedCalendarAttendeeObservation{}, false
	}

	attendees, err := tx.CalendarAttendee.Query().
		Where(
			calendarattendee.CalendarEventIDEQ(event.eventID),
			inboundRequestFixturePublishedCalendarAttendeeNoWait(),
		).
		Order(ent.Asc(calendarattendee.FieldID)).
		Limit(2).
		All(ctx)
	if err != nil || len(attendees) != 1 ||
		!inboundRequestFixturePublishedCalendarAttendeeMatches(
			attendees[0],
			roots,
			plan,
			event.eventID,
		) {
		return inboundRequestFixturePublishedCalendarAttendeeObservation{}, false
	}
	attendee := attendees[0]

	replyReceipts, err := tx.CalendarInboundReplyReceipt.Query().
		Where(
			calendarinboundreplyreceipt.CalendarAttendeeIDEQ(attendee.ID),
			inboundRequestFixturePublishedCalendarInboundReplyReceiptNoWait(),
		).
		Order(ent.Asc(calendarinboundreplyreceipt.FieldID)).
		Limit(1).
		All(ctx)
	if err != nil || len(replyReceipts) != 0 {
		return inboundRequestFixturePublishedCalendarAttendeeObservation{}, false
	}

	observation := inboundRequestFixturePublishedCalendarAttendeeObservation{
		inboundMessageReceiptID: event.inboundMessageReceiptID,
		inboundObjectStateID:    event.inboundObjectStateID,
		eventID:                 event.eventID,
		attendeeID:              attendee.ID,
	}
	if !observation.valid() {
		return inboundRequestFixturePublishedCalendarAttendeeObservation{}, false
	}
	return observation, true
}

func inboundRequestFixturePublishedCalendarAttendeeMatches(
	attendee *ent.CalendarAttendee,
	roots inboundRequestFixtureRootProof,
	plan inboundRequestFixturePlan,
	expectedEventID uuid.UUID,
) bool {
	return attendee != nil &&
		roots.valid() &&
		expectedEventID != uuid.Nil &&
		attendee.ID != uuid.Nil &&
		attendee.CalendarEventID == expectedEventID &&
		attendee.UserID != nil &&
		*attendee.UserID == roots.ownerUserID &&
		attendee.NormalizedEmail == plan.ownerAddress &&
		attendee.DisplayName == nil &&
		attendee.Role == calendarattendee.Role("required") &&
		attendee.RsvpStatus == calendarattendee.RsvpStatus("accepted") &&
		attendee.RecurrenceID == nil &&
		attendee.RespondedAt != nil &&
		inboundRequestFixturePublishedCalendarTimestampValid(*attendee.RespondedAt) &&
		attendee.RemovedAt == nil &&
		attendee.ReplyDeliveryStatus == calendarattendee.ReplyDeliveryStatus("pending") &&
		attendee.Version == 2 &&
		inboundRequestFixturePublishedCalendarTimestampValid(attendee.CreatedAt) &&
		inboundRequestFixturePublishedCalendarTimestampValid(attendee.UpdatedAt) &&
		!attendee.UpdatedAt.Before(attendee.CreatedAt)
}

func inboundRequestFixturePublishedCalendarAttendeeNoWait() predicate.CalendarAttendee {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedCalendarInboundReplyReceiptNoWait() predicate.CalendarInboundReplyReceipt {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
