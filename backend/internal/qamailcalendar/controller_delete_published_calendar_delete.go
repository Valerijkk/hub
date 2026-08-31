//go:build qa && !release

package qamailcalendar

import (
	"context"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendarinboundmessagereceipt"
	"github.com/valerii/hub/backend/ent/calendarinboundobjectstate"
	"github.com/valerii/hub/backend/ent/calendarmutationreceipt"
	"github.com/valerii/hub/backend/ent/calendarreplydelivery"
)

func deleteInboundRequestFixturePublishedCalendarExact(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) bool {
	if ctx == nil || tx == nil ||
		!inboundRequestFixtureRootProofMatchesPlan(roots, plan, persisted) ||
		!calendar.valid() {
		return false
	}

	affected, err := tx.CalendarReplyDelivery.Delete().
		Where(
			calendarreplydelivery.IDEQ(calendar.replyDeliveryID),
			calendarreplydelivery.OrganizationIDEQ(roots.organizationID),
			calendarreplydelivery.CalendarEventIDEQ(calendar.eventID),
			calendarreplydelivery.CalendarAttendeeIDEQ(calendar.attendeeID),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	affected, err = tx.CalendarMutationReceipt.Delete().
		Where(
			calendarmutationreceipt.IDEQ(calendar.mutationReceiptID),
			calendarmutationreceipt.OrganizationIDEQ(roots.organizationID),
			calendarmutationreceipt.ActorUserIDEQ(roots.ownerUserID),
			calendarmutationreceipt.OperationEQ(calendarmutationreceipt.Operation("rsvp_series")),
			calendarmutationreceipt.CalendarEventIDEQ(calendar.eventID),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	affected, err = tx.CalendarInboundMessageReceipt.Delete().
		Where(
			calendarinboundmessagereceipt.IDEQ(calendar.inboundMessageReceiptID),
			calendarinboundmessagereceipt.MailMessageIDEQ(roots.mailMessageID),
			calendarinboundmessagereceipt.EventEnvelopeIDEQ(roots.envelopeEventID),
			calendarinboundmessagereceipt.PartLocatorHashEQ(plan.calendarLocatorHash[:]),
			calendarinboundmessagereceipt.CalendarEventIDEQ(calendar.eventID),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	affected, err = tx.CalendarInboundObjectState.Delete().
		Where(
			calendarinboundobjectstate.IDEQ(calendar.inboundObjectStateID),
			calendarinboundobjectstate.OrganizationIDEQ(roots.organizationID),
			calendarinboundobjectstate.CalendarOwnerIDEQ(roots.ownerUserID),
			calendarinboundobjectstate.UIDEQ(plan.itipUID),
			calendarinboundobjectstate.RecurrenceIDIsNil(),
			calendarinboundobjectstate.CalendarEventIDEQ(calendar.eventID),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	affected, err = tx.CalendarAttendee.Delete().
		Where(
			calendarattendee.IDEQ(calendar.attendeeID),
			calendarattendee.CalendarEventIDEQ(calendar.eventID),
			calendarattendee.UserIDEQ(roots.ownerUserID),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	affected, err = tx.CalendarEvent.Delete().
		Where(
			calendarevent.IDEQ(calendar.eventID),
			calendarevent.OrganizationIDEQ(roots.organizationID),
			calendarevent.CalendarOwnerIDEQ(roots.ownerUserID),
			calendarevent.UIDEQ(plan.itipUID),
		).
		Exec(ctx)
	return err == nil && affected == 1
}
