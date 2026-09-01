//go:build qa && !release

package qamailcalendar

import (
	"bytes"
	"context"
	"crypto/sha256"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendareventexception"
	"github.com/valerii/hub/backend/ent/calendarinboundobjectstate"
	"github.com/valerii/hub/backend/ent/predicate"
)

// inboundRequestFixturePublishedCalendarEventObservation is the private,
// comparable C3b result. Later Calendar shards own attendee, delivery,
// mutation-receipt, external, Outbox, lifecycle, and delete observations.
type inboundRequestFixturePublishedCalendarEventObservation struct {
	inboundMessageReceiptID uuid.UUID
	inboundObjectStateID    uuid.UUID
	eventID                 uuid.UUID
}

func (observation inboundRequestFixturePublishedCalendarEventObservation) valid() bool {
	return observation.inboundMessageReceiptID != uuid.Nil &&
		observation.inboundObjectStateID != uuid.Nil &&
		observation.eventID != uuid.Nil
}

// observeInboundRequestFixturePublishedCalendarEvent revalidates C3a, then
// locks the sole imported series event and its exact object-state binding in
// the caller-owned transaction. It returns copied UUIDs only and never
// mutates, commits, or retains Ent rows.
func observeInboundRequestFixturePublishedCalendarEvent(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	authority inboundRequestFixturePublishedEngineAuthority,
	lockedRoots inboundRequestFixtureLockedPublishedRoots,
	mailClosure inboundRequestFixturePublishedMailClosure,
) (inboundRequestFixturePublishedCalendarEventObservation, bool) {
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
		return inboundRequestFixturePublishedCalendarEventObservation{}, false
	}

	receipt, observed := observeInboundRequestFixturePublishedCalendarReceipt(
		ctx,
		tx,
		plan,
		persisted,
		authority,
		lockedRoots,
		mailClosure,
	)
	if !observed || !receipt.valid() {
		return inboundRequestFixturePublishedCalendarEventObservation{}, false
	}

	events, err := tx.CalendarEvent.Query().
		Where(
			calendarevent.OrganizationIDEQ(roots.organizationID),
			calendarevent.CalendarOwnerIDEQ(roots.ownerUserID),
			calendarevent.UIDEQ(plan.itipUID),
			inboundRequestFixturePublishedCalendarEventNoWait(),
		).
		Order(ent.Asc(calendarevent.FieldID)).
		Limit(2).
		All(ctx)
	if err != nil || len(events) != 1 ||
		!inboundRequestFixturePublishedCalendarEventMatches(events[0], roots, plan, receipt.eventID) {
		return inboundRequestFixturePublishedCalendarEventObservation{}, false
	}
	event := events[0]

	objectStates, err := tx.CalendarInboundObjectState.Query().
		Where(
			calendarinboundobjectstate.IDEQ(receipt.inboundObjectStateID),
			inboundRequestFixturePublishedCalendarObjectStateNoWait(),
		).
		Order(ent.Asc(calendarinboundobjectstate.FieldID)).
		Limit(2).
		All(ctx)
	if err != nil || len(objectStates) != 1 ||
		!inboundRequestFixturePublishedCalendarObjectStateMatches(
			objectStates[0],
			roots,
			plan,
			event.ID,
		) ||
		event.PayloadHash == nil ||
		!bytes.Equal(*event.PayloadHash, objectStates[0].PayloadHash) {
		return inboundRequestFixturePublishedCalendarEventObservation{}, false
	}

	exceptions, err := tx.CalendarEventException.Query().
		Where(
			calendareventexception.CalendarEventIDEQ(event.ID),
			inboundRequestFixturePublishedCalendarEventExceptionNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(exceptions) != 0 {
		return inboundRequestFixturePublishedCalendarEventObservation{}, false
	}

	observation := inboundRequestFixturePublishedCalendarEventObservation{
		inboundMessageReceiptID: receipt.inboundMessageReceiptID,
		inboundObjectStateID:    receipt.inboundObjectStateID,
		eventID:                 event.ID,
	}
	if !observation.valid() {
		return inboundRequestFixturePublishedCalendarEventObservation{}, false
	}
	return observation, true
}

func inboundRequestFixturePublishedCalendarEventMatches(
	event *ent.CalendarEvent,
	roots inboundRequestFixtureRootProof,
	plan inboundRequestFixturePlan,
	expectedEventID uuid.UUID,
) bool {
	return event != nil &&
		roots.valid() &&
		expectedEventID != uuid.Nil &&
		event.ID == expectedEventID &&
		event.OrganizationID == roots.organizationID &&
		event.CalendarOwnerID == roots.ownerUserID &&
		event.UID == plan.itipUID &&
		event.Title == plan.subject &&
		event.Description == "" &&
		event.Location == "" &&
		event.StartsAt != nil &&
		event.StartsAt.Equal(plan.meetingStartsAt) &&
		event.EndsAt != nil &&
		event.EndsAt.Equal(plan.meetingEndsAt) &&
		event.StartDate == nil &&
		event.EndDateExclusive == nil &&
		event.AllDay == false &&
		event.Timezone == "UTC" &&
		event.Rrule == nil &&
		len(event.Exdates) == 0 &&
		event.Sequence == 0 &&
		event.Version == 1 &&
		event.Source == calendarevent.Source("mail_import") &&
		event.Status == calendarevent.Status("confirmed") &&
		event.OrganizerID == nil &&
		event.OrganizerEmail == plan.organizerAddress &&
		event.ScheduledCallID == nil &&
		event.ReminderMinutes == nil &&
		event.PayloadHashVersion != nil &&
		*event.PayloadHashVersion == 1 &&
		event.PayloadHash != nil &&
		len(*event.PayloadHash) == sha256.Size &&
		event.DeletedAt == nil &&
		inboundRequestFixturePublishedCalendarTimestampValid(event.CreatedAt) &&
		inboundRequestFixturePublishedCalendarTimestampValid(event.UpdatedAt) &&
		!event.UpdatedAt.Before(event.CreatedAt)
}

func inboundRequestFixturePublishedCalendarEventNoWait() predicate.CalendarEvent {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedCalendarEventExceptionNoWait() predicate.CalendarEventException {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
