//go:build qa && !release

package qamailcalendar

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"io"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarreplydelivery"
	"github.com/valerii/hub/backend/ent/calendarmutationreceipt"
	"github.com/valerii/hub/backend/ent/predicate"
)

// inboundRequestFixturePublishedCalendarRSVPResponse is the complete public
// series-RSVP receipt shape. It remains private so C3d can accept only the
// canonical response stored by the completed mutation receipt.
type inboundRequestFixturePublishedCalendarRSVPResponse struct {
	ID                  uuid.UUID  `json:"id"`
	UserID              *uuid.UUID `json:"user_id"`
	Email               string     `json:"email"`
	DisplayName         *string    `json:"display_name"`
	Role                string     `json:"role"`
	RSVPStatus          string     `json:"rsvp_status"`
	RecurrenceID        *string    `json:"recurrence_id"`
	RespondedAt         *time.Time `json:"responded_at"`
	Version             uint       `json:"version"`
	ReplyDeliveryStatus string     `json:"reply_delivery_status"`
}

// observeInboundRequestFixturePublishedCalendarEffects revalidates C3c, then
// locks the sole pending outbound RSVP delivery and completed series-RSVP
// receipt. It returns copied Calendar UUIDs only; it never reads external or
// Outbox state and never mutates, commits, or retains Ent rows.
func observeInboundRequestFixturePublishedCalendarEffects(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	authority inboundRequestFixturePublishedEngineAuthority,
	lockedRoots inboundRequestFixtureLockedPublishedRoots,
	mailClosure inboundRequestFixturePublishedMailClosure,
) (inboundRequestFixtureCalendarProof, bool) {
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
		return inboundRequestFixtureCalendarProof{}, false
	}

	attendee, observed := observeInboundRequestFixturePublishedCalendarAttendee(
		ctx,
		tx,
		plan,
		persisted,
		authority,
		lockedRoots,
		mailClosure,
	)
	if !observed || !attendee.valid() {
		return inboundRequestFixtureCalendarProof{}, false
	}

	attendees, err := tx.CalendarAttendee.Query().
		Where(
			calendarattendee.IDEQ(attendee.attendeeID),
			calendarattendee.CalendarEventIDEQ(attendee.eventID),
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
			attendee.eventID,
		) ||
		attendees[0].ID != attendee.attendeeID ||
		attendees[0].RespondedAt == nil {
		return inboundRequestFixtureCalendarProof{}, false
	}
	respondedAt := *attendees[0].RespondedAt
	if !inboundRequestFixturePublishedCalendarTimestampValid(respondedAt) {
		return inboundRequestFixtureCalendarProof{}, false
	}

	deliveries, err := tx.CalendarReplyDelivery.Query().
		Where(
			calendarreplydelivery.OrganizationIDEQ(roots.organizationID),
			calendarreplydelivery.CalendarEventIDEQ(attendee.eventID),
			calendarreplydelivery.CalendarAttendeeIDEQ(attendee.attendeeID),
			inboundRequestFixturePublishedCalendarReplyDeliveryNoWait(),
		).
		Order(ent.Asc(calendarreplydelivery.FieldID)).
		Limit(2).
		All(ctx)
	if err != nil || len(deliveries) != 1 ||
		!inboundRequestFixturePublishedCalendarReplyDeliveryMatches(
			deliveries[0],
			roots,
			attendee,
		) {
		return inboundRequestFixtureCalendarProof{}, false
	}
	delivery := deliveries[0]

	receipts, err := tx.CalendarMutationReceipt.Query().
		Where(
			calendarmutationreceipt.OrganizationIDEQ(roots.organizationID),
			calendarmutationreceipt.ActorUserIDEQ(roots.ownerUserID),
			calendarmutationreceipt.OperationEQ(calendarmutationreceipt.Operation("rsvp_series")),
			calendarmutationreceipt.CalendarEventIDEQ(attendee.eventID),
			inboundRequestFixturePublishedCalendarMutationReceiptNoWait(),
		).
		Order(ent.Asc(calendarmutationreceipt.FieldID)).
		Limit(2).
		All(ctx)
	if err != nil || len(receipts) != 1 ||
		!inboundRequestFixturePublishedCalendarMutationReceiptMatches(
			receipts[0],
			roots,
			plan,
			attendee,
			respondedAt,
		) {
		return inboundRequestFixtureCalendarProof{}, false
	}
	receipt := receipts[0]

	proof, valid := newInboundRequestFixtureCalendarProof(
		attendee.inboundMessageReceiptID,
		attendee.inboundObjectStateID,
		attendee.eventID,
		attendee.attendeeID,
		delivery.ID,
		receipt.ID,
	)
	if !valid {
		return inboundRequestFixtureCalendarProof{}, false
	}
	return proof, true
}

func inboundRequestFixturePublishedCalendarReplyDeliveryMatches(
	delivery *ent.CalendarReplyDelivery,
	roots inboundRequestFixtureRootProof,
	attendee inboundRequestFixturePublishedCalendarAttendeeObservation,
) bool {
	return delivery != nil &&
		roots.valid() &&
		attendee.valid() &&
		delivery.ID != uuid.Nil &&
		delivery.OrganizationID == roots.organizationID &&
		delivery.CalendarEventID == attendee.eventID &&
		delivery.CalendarAttendeeID == attendee.attendeeID &&
		delivery.EventSequence == 0 &&
		delivery.AttendeeVersion == 2 &&
		delivery.RecurrenceID == nil &&
		delivery.RsvpStatus == calendarreplydelivery.RsvpStatus("accepted") &&
		delivery.State == calendarreplydelivery.State("pending") &&
		delivery.MailSubmissionID == nil &&
		delivery.Attempts == 0 &&
		delivery.LeaseToken == nil &&
		delivery.LeaseUntil == nil &&
		delivery.ClaimGeneration == 0 &&
		delivery.NextAttemptAt != nil &&
		inboundRequestFixturePublishedCalendarWholeSecondUTC(*delivery.NextAttemptAt) &&
		delivery.LastErrorCode == nil &&
		inboundRequestFixturePublishedCalendarTimestampValid(delivery.CreatedAt) &&
		inboundRequestFixturePublishedCalendarTimestampValid(delivery.UpdatedAt) &&
		!delivery.UpdatedAt.Before(delivery.CreatedAt)
}

func inboundRequestFixturePublishedCalendarMutationReceiptMatches(
	receipt *ent.CalendarMutationReceipt,
	roots inboundRequestFixtureRootProof,
	plan inboundRequestFixturePlan,
	attendee inboundRequestFixturePublishedCalendarAttendeeObservation,
	respondedAt time.Time,
) bool {
	if receipt == nil ||
		!roots.valid() ||
		!inboundRequestFixturePlanIsCanonical(plan) ||
		!attendee.valid() ||
		!inboundRequestFixturePublishedCalendarTimestampValid(respondedAt) ||
		receipt.ID == uuid.Nil ||
		receipt.OrganizationID != roots.organizationID ||
		receipt.ActorUserID != roots.ownerUserID ||
		receipt.Operation != calendarmutationreceipt.Operation("rsvp_series") ||
		len(receipt.IdempotencyKeyHash) != sha256.Size ||
		len(receipt.RequestHash) != sha256.Size ||
		receipt.State != calendarmutationreceipt.State("completed") ||
		receipt.ResponseStatus == nil || *receipt.ResponseStatus != 200 ||
		receipt.ResponseJSON == nil ||
		receipt.CalendarEventID == nil || *receipt.CalendarEventID != attendee.eventID ||
		!inboundRequestFixturePublishedCalendarTimestampValid(receipt.CreatedAt) ||
		!inboundRequestFixturePublishedCalendarTimestampValid(receipt.UpdatedAt) ||
		!inboundRequestFixturePublishedCalendarWholeSecondUTC(receipt.ExpiresAt) ||
		receipt.UpdatedAt.Before(receipt.CreatedAt) ||
		!receipt.ExpiresAt.After(receipt.UpdatedAt) {
		return false
	}

	return inboundRequestFixturePublishedCalendarRSVPResponseMatches(
		*receipt.ResponseJSON,
		roots,
		plan,
		attendee,
		respondedAt,
	)
}

func inboundRequestFixturePublishedCalendarRSVPResponseMatches(
	raw []byte,
	roots inboundRequestFixtureRootProof,
	plan inboundRequestFixturePlan,
	attendee inboundRequestFixturePublishedCalendarAttendeeObservation,
	respondedAt time.Time,
) bool {
	if !roots.valid() ||
		!inboundRequestFixturePlanIsCanonical(plan) ||
		!attendee.valid() ||
		!inboundRequestFixturePublishedCalendarTimestampValid(respondedAt) {
		return false
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var response inboundRequestFixturePublishedCalendarRSVPResponse
	if err := decoder.Decode(&response); err != nil {
		return false
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return false
	}

	canonical, err := json.Marshal(response)
	if err != nil || !bytes.Equal(raw, canonical) ||
		response.ID != attendee.attendeeID ||
		response.UserID == nil || *response.UserID != roots.ownerUserID ||
		response.Email != plan.ownerAddress ||
		response.DisplayName != nil ||
		response.Role != "required" ||
		response.RSVPStatus != "accepted" ||
		response.RecurrenceID != nil ||
		response.RespondedAt == nil ||
		!inboundRequestFixturePublishedCalendarTimestampValid(*response.RespondedAt) ||
		response.Version != 2 ||
		response.ReplyDeliveryStatus != "pending" {
		return false
	}

	return response.RespondedAt.Equal(respondedAt)
}

func inboundRequestFixturePublishedCalendarWholeSecondUTC(value time.Time) bool {
	return inboundRequestFixturePublishedCalendarTimestampValid(value) && value.Nanosecond() == 0
}

func inboundRequestFixturePublishedCalendarReplyDeliveryNoWait() predicate.CalendarReplyDelivery {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedCalendarMutationReceiptNoWait() predicate.CalendarMutationReceipt {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
