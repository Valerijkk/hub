//go:build qa && !release

package qamailcalendar

import (
	"bytes"
	"context"
	"crypto/sha256"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarinboundmessagereceipt"
	"github.com/valerii/hub/backend/ent/calendarinboundobjectstate"
	"github.com/valerii/hub/backend/ent/predicate"
)

// inboundRequestFixturePublishedCalendarReceiptObservation is the private,
// comparable C3a result. Later Calendar shards own all attendee, delivery,
// mutation-receipt, external, Outbox, lifecycle, and delete observations.
type inboundRequestFixturePublishedCalendarReceiptObservation struct {
	inboundMessageReceiptID uuid.UUID
	inboundObjectStateID    uuid.UUID
	eventID                 uuid.UUID
}

func (observation inboundRequestFixturePublishedCalendarReceiptObservation) valid() bool {
	return observation.inboundMessageReceiptID != uuid.Nil &&
		observation.inboundObjectStateID != uuid.Nil &&
		observation.eventID != uuid.Nil
}

// observeInboundRequestFixturePublishedCalendarReceipt locks the exact applied
// receipt and its series object-state evidence inside the caller-owned
// transaction. It only returns copied UUIDs; it does not mutate or retain rows.
func observeInboundRequestFixturePublishedCalendarReceipt(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	authority inboundRequestFixturePublishedEngineAuthority,
	lockedRoots inboundRequestFixtureLockedPublishedRoots,
	mailClosure inboundRequestFixturePublishedMailClosure,
) (inboundRequestFixturePublishedCalendarReceiptObservation, bool) {
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
		!mailClosure.falseSent.matchesRoots(roots) ||
		lockedRoots.mail.message == nil ||
		lockedRoots.mail.message.CalendarEventID != nil {
		return inboundRequestFixturePublishedCalendarReceiptObservation{}, false
	}

	receipts, err := tx.CalendarInboundMessageReceipt.Query().
		Where(
			calendarinboundmessagereceipt.MailMessageIDEQ(roots.mailMessageID),
			calendarinboundmessagereceipt.EventEnvelopeIDEQ(roots.envelopeEventID),
			calendarinboundmessagereceipt.PartLocatorHashEQ(plan.calendarLocatorHash[:]),
			inboundRequestFixturePublishedCalendarReceiptNoWait(),
		).
		Order(ent.Asc(calendarinboundmessagereceipt.FieldID)).
		Limit(2).
		All(ctx)
	if err != nil || len(receipts) != 1 ||
		!inboundRequestFixturePublishedCalendarReceiptMatches(
			receipts[0],
			roots,
			plan,
		) {
		return inboundRequestFixturePublishedCalendarReceiptObservation{}, false
	}

	receipt := receipts[0]
	eventID := *receipt.CalendarEventID
	objectStates, err := tx.CalendarInboundObjectState.Query().
		Where(
			calendarinboundobjectstate.OrganizationIDEQ(roots.organizationID),
			calendarinboundobjectstate.CalendarOwnerIDEQ(roots.ownerUserID),
			calendarinboundobjectstate.UIDEQ(plan.itipUID),
			calendarinboundobjectstate.RecurrenceIDIsNil(),
			calendarinboundobjectstate.CalendarEventIDEQ(eventID),
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
			eventID,
		) {
		return inboundRequestFixturePublishedCalendarReceiptObservation{}, false
	}

	observation := inboundRequestFixturePublishedCalendarReceiptObservation{
		inboundMessageReceiptID: receipt.ID,
		inboundObjectStateID:    objectStates[0].ID,
		eventID:                 eventID,
	}
	if !observation.valid() {
		return inboundRequestFixturePublishedCalendarReceiptObservation{}, false
	}
	return observation, true
}

func inboundRequestFixturePublishedCalendarReceiptMatches(
	receipt *ent.CalendarInboundMessageReceipt,
	roots inboundRequestFixtureRootProof,
	plan inboundRequestFixturePlan,
) bool {
	return receipt != nil &&
		roots.valid() &&
		receipt.ID != uuid.Nil &&
		receipt.OrganizationID == roots.organizationID &&
		receipt.MailboxID == roots.mailAccountID &&
		receipt.MailMessageID == roots.mailMessageID &&
		receipt.EventEnvelopeID == roots.envelopeEventID &&
		bytes.Equal(receipt.PartLocatorHash, plan.calendarLocatorHash[:]) &&
		receipt.State == calendarinboundmessagereceipt.State("applied") &&
		receipt.Attempts == 1 &&
		receipt.LeaseGeneration == 1 &&
		receipt.LeaseToken == nil &&
		receipt.LeaseUntil == nil &&
		receipt.NextAttemptAt == nil &&
		receipt.ErrorCode == nil &&
		receipt.DeferredObjectStateID == nil &&
		receipt.DeferredExpiresAt == nil &&
		receipt.CalendarEventID != nil &&
		*receipt.CalendarEventID != uuid.Nil &&
		inboundRequestFixturePublishedCalendarTimestampValid(receipt.CreatedAt) &&
		inboundRequestFixturePublishedCalendarTimestampValid(receipt.UpdatedAt) &&
		!receipt.UpdatedAt.Before(receipt.CreatedAt)
}

func inboundRequestFixturePublishedCalendarObjectStateMatches(
	state *ent.CalendarInboundObjectState,
	roots inboundRequestFixtureRootProof,
	plan inboundRequestFixturePlan,
	eventID uuid.UUID,
) bool {
	return state != nil &&
		roots.valid() &&
		eventID != uuid.Nil &&
		state.ID != uuid.Nil &&
		state.OrganizationID == roots.organizationID &&
		state.CalendarOwnerID == roots.ownerUserID &&
		state.UID == plan.itipUID &&
		state.RecurrenceID == nil &&
		state.OrganizerEmail == plan.organizerAddress &&
		state.Sequence == 0 &&
		state.Method == calendarinboundobjectstate.Method("request") &&
		state.State == calendarinboundobjectstate.State("applied") &&
		state.PayloadHashVersion == 1 &&
		len(state.PayloadHash) == sha256.Size &&
		state.CalendarEventID != nil &&
		*state.CalendarEventID == eventID &&
		state.ExpiresAt == nil &&
		inboundRequestFixturePublishedCalendarTimestampValid(state.CreatedAt) &&
		inboundRequestFixturePublishedCalendarTimestampValid(state.UpdatedAt) &&
		!state.UpdatedAt.Before(state.CreatedAt)
}

func inboundRequestFixturePublishedCalendarTimestampValid(value time.Time) bool {
	return !value.IsZero()
}

func inboundRequestFixturePublishedCalendarReceiptNoWait() predicate.CalendarInboundMessageReceipt {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedCalendarObjectStateNoWait() predicate.CalendarInboundObjectState {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
