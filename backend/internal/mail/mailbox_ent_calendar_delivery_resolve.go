package mail

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarinvitedelivery"
	"github.com/valerii/hub/backend/ent/calendarreplydelivery"
)

// ResolveCalendarMailDelivery releases one exact, still-live unbound claim
// after a pre-bind Mail outcome. It makes no external engine or Mail reads: the
// sealed claim and delivery row are the entire durable resolution boundary.
func (repository *MailboxEntRepository) ResolveCalendarMailDelivery(ctx context.Context, request CalendarMailDeliveryResolveRequest) error {
	if repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || !validCalendarMailDeliveryResolveRequest(request) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil || !mailboxEntValidTime(databaseNow) {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		switch request.Claim.kind {
		case CalendarMailDeliveryKindInvite:
			return mailboxEntResolveCalendarInviteDelivery(ctx, transaction, request.Claim, request.Code, databaseNow)
		case CalendarMailDeliveryKindReply:
			return mailboxEntResolveCalendarReplyDelivery(ctx, transaction, request.Claim, request.Code, databaseNow)
		default:
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
		}
	})
	return mailboxEntCalendarDeliveryRepositoryError(err)
}

func mailboxEntResolveCalendarInviteDelivery(ctx context.Context, transaction *ent.Tx, claim CalendarMailDeliveryClaim, code CalendarMailDeliveryResolveCode, databaseNow time.Time) error {
	if ctx == nil || transaction == nil || !validCalendarMailDeliveryClaim(claim) || claim.kind != CalendarMailDeliveryKindInvite ||
		!validCalendarMailDeliveryPreBindResolveCode(code) || !mailboxEntValidTime(databaseNow) || !claim.leaseUntil.After(databaseNow) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	delivery, err := transaction.CalendarInviteDelivery.Query().
		Where(calendarinvitedelivery.IDEQ(claim.deliveryID.UUID())).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if err != nil || delivery == nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if !mailboxEntCalendarInviteDeliverySubmittingValid(delivery) || !mailboxEntCalendarInviteDeliveryResolveScopeMatches(delivery, claim) ||
		!mailboxEntCalendarDeliveryResolveFenceMatches(delivery.LeaseToken, delivery.LeaseUntil, delivery.ClaimGeneration, claim) || !delivery.LeaseUntil.After(databaseNow) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	resolution, err := mailboxEntCalendarDeliveryResolvePlan(code, delivery.Attempts, databaseNow)
	if err != nil {
		return err
	}
	updated, err := mailboxEntCASCalendarInviteDeliveryResolve(ctx, transaction, delivery, claim, resolution, databaseNow)
	if err != nil {
		return err
	}
	if updated != 1 {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	return nil
}

func mailboxEntResolveCalendarReplyDelivery(ctx context.Context, transaction *ent.Tx, claim CalendarMailDeliveryClaim, code CalendarMailDeliveryResolveCode, databaseNow time.Time) error {
	if ctx == nil || transaction == nil || !validCalendarMailDeliveryClaim(claim) || claim.kind != CalendarMailDeliveryKindReply ||
		!validCalendarMailDeliveryPreBindResolveCode(code) || !mailboxEntValidTime(databaseNow) || !claim.leaseUntil.After(databaseNow) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	delivery, err := transaction.CalendarReplyDelivery.Query().
		Where(calendarreplydelivery.IDEQ(claim.deliveryID.UUID())).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if err != nil || delivery == nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if !mailboxEntCalendarReplyDeliverySubmittingValid(delivery) || !mailboxEntCalendarReplyDeliveryResolveScopeMatches(delivery, claim) ||
		!mailboxEntCalendarDeliveryResolveFenceMatches(delivery.LeaseToken, delivery.LeaseUntil, delivery.ClaimGeneration, claim) || !delivery.LeaseUntil.After(databaseNow) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	resolution, err := mailboxEntCalendarDeliveryResolvePlan(code, delivery.Attempts, databaseNow)
	if err != nil {
		return err
	}
	updated, err := mailboxEntCASCalendarReplyDeliveryResolve(ctx, transaction, delivery, claim, resolution, databaseNow)
	if err != nil {
		return err
	}
	if updated != 1 {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	return nil
}

type mailboxEntCalendarDeliveryResolution struct {
	state           CalendarMailDeliveryState
	lastErrorCode   string
	nextAttemptAt   *time.Time
	terminal        bool
}

func mailboxEntCalendarDeliveryResolvePlan(code CalendarMailDeliveryResolveCode, attempts int, databaseNow time.Time) (mailboxEntCalendarDeliveryResolution, error) {
	if !validCalendarMailDeliveryPreBindResolveCode(code) || attempts < 1 || attempts > calendarMailDeliveryMaximumAttempts || !mailboxEntValidTime(databaseNow) {
		return mailboxEntCalendarDeliveryResolution{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	switch code {
	case CalendarMailDeliveryResolveCodeMailboxUnavailable, CalendarMailDeliveryResolveCodeInternal:
		if attempts == calendarMailDeliveryMaximumAttempts {
			return mailboxEntCalendarDeliveryResolution{
				state:         CalendarMailDeliveryStateFailedPermanent,
				lastErrorCode: "retry_exhausted",
				terminal:      true,
			}, nil
		}
		delay, ok := mailboxEntCalendarDeliveryResolveRetryDelay(attempts)
		if !ok {
			return mailboxEntCalendarDeliveryResolution{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		nextAttemptAt, ok := mailboxEntCheckedTimeAdd(databaseNow, delay)
		if !ok {
			return mailboxEntCalendarDeliveryResolution{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		lastErrorCode := "mailbox_unavailable"
		if code == CalendarMailDeliveryResolveCodeInternal {
			lastErrorCode = "internal_error"
		}
		return mailboxEntCalendarDeliveryResolution{
			state:         CalendarMailDeliveryStatePending,
			lastErrorCode: lastErrorCode,
			nextAttemptAt: &nextAttemptAt,
		}, nil
	case CalendarMailDeliveryResolveCodeInvalid:
		return mailboxEntCalendarDeliveryResolution{
			state:         CalendarMailDeliveryStateFailedPermanent,
			lastErrorCode: "materialization_invalid",
			terminal:      true,
		}, nil
	case CalendarMailDeliveryResolveCodeRejected:
		return mailboxEntCalendarDeliveryResolution{
			state:         CalendarMailDeliveryStateFailedPermanent,
			lastErrorCode: "submission_rejected",
			terminal:      true,
		}, nil
	case CalendarMailDeliveryResolveCodePolicyDenied:
		return mailboxEntCalendarDeliveryResolution{
			state:         CalendarMailDeliveryStateFailedPermanent,
			lastErrorCode: "policy_denied",
			terminal:      true,
		}, nil
	default:
		return mailboxEntCalendarDeliveryResolution{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
}

func mailboxEntCalendarDeliveryResolveRetryDelay(attempts int) (time.Duration, bool) {
	if attempts < 1 || attempts > calendarMailDeliveryMaximumAttempts {
		return 0, false
	}
	delay := 30 * time.Second * time.Duration(1<<uint(attempts - 1))
	if delay > calendarMailDeliveryRetryMaximum {
		delay = calendarMailDeliveryRetryMaximum
	}
	return delay, delay >= calendarMailDeliveryRetryMinimum && delay <= calendarMailDeliveryRetryMaximum
}

func mailboxEntCalendarInviteDeliveryResolveScopeMatches(delivery *ent.CalendarInviteDelivery, claim CalendarMailDeliveryClaim) bool {
	if delivery == nil || claim.kind != CalendarMailDeliveryKindInvite || claim.deliveryID.Kind() != CalendarMailDeliveryKindInvite ||
		delivery.ID != claim.deliveryID.UUID() || delivery.OrganizationID != claim.organizationID || delivery.CalendarEventID != claim.eventID ||
		delivery.CalendarAttendeeID != claim.attendeeID || !mailboxEntCalendarDeliveryRecurrenceMatches(delivery.RecurrenceID, claim.materialization.recurrenceID) ||
		delivery.EventSequence != claim.materialization.eventSequence {
		return false
	}
	method, ok := mailboxEntCalendarInviteDeliveryMethod(delivery.Kind)
	return ok && method == claim.method
}

func mailboxEntCalendarReplyDeliveryResolveScopeMatches(delivery *ent.CalendarReplyDelivery, claim CalendarMailDeliveryClaim) bool {
	return delivery != nil && claim.kind == CalendarMailDeliveryKindReply && claim.deliveryID.Kind() == CalendarMailDeliveryKindReply &&
		delivery.ID == claim.deliveryID.UUID() && delivery.OrganizationID == claim.organizationID && delivery.CalendarEventID == claim.eventID &&
		delivery.CalendarAttendeeID == claim.attendeeID && mailboxEntCalendarDeliveryRecurrenceMatches(delivery.RecurrenceID, claim.materialization.recurrenceID) &&
		delivery.EventSequence == claim.materialization.eventSequence && delivery.AttendeeVersion == claim.materialization.attendeeVersion
}

func mailboxEntCalendarDeliveryResolveFenceMatches(token *uuid.UUID, leaseUntil *time.Time, generation int64, claim CalendarMailDeliveryClaim) bool {
	return token != nil && *token == uuid.UUID(claim.claimToken) && leaseUntil != nil && leaseUntil.Equal(claim.leaseUntil) &&
		generation == int64(claim.claimGeneration)
}

func mailboxEntCASCalendarInviteDeliveryResolve(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery, claim CalendarMailDeliveryClaim, resolution mailboxEntCalendarDeliveryResolution, databaseNow time.Time) (int, error) {
	if ctx == nil || transaction == nil || delivery == nil || !validCalendarMailDeliveryClaim(claim) || !mailboxEntValidTime(databaseNow) ||
		resolution.lastErrorCode == "" || (resolution.terminal == (resolution.nextAttemptAt != nil)) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	update := transaction.CalendarInviteDelivery.Update().Where(
		calendarinvitedelivery.IDEQ(claim.deliveryID.UUID()),
		calendarinvitedelivery.OrganizationIDEQ(claim.organizationID),
		calendarinvitedelivery.CalendarEventIDEQ(claim.eventID),
		calendarinvitedelivery.CalendarAttendeeIDEQ(claim.attendeeID),
		calendarinvitedelivery.EventSequenceEQ(claim.materialization.eventSequence),
		calendarinvitedelivery.StateEQ(calendarinvitedelivery.StateSubmitting),
		calendarinvitedelivery.MailSubmissionIDIsNil(),
		calendarinvitedelivery.AttemptsEQ(delivery.Attempts),
		calendarinvitedelivery.ClaimGenerationEQ(int64(claim.claimGeneration)),
		calendarinvitedelivery.LeaseTokenEQ(uuid.UUID(claim.claimToken)),
		calendarinvitedelivery.LeaseUntilEQ(claim.leaseUntil),
		calendarinvitedelivery.NextAttemptAtIsNil(),
		calendarinvitedelivery.LastErrorCodeIsNil(),
		calendarinvitedelivery.TerminalAtIsNil(),
	)
	if claim.materialization.recurrenceID == "" {
		update.Where(calendarinvitedelivery.RecurrenceIDIsNil())
	} else {
		update.Where(calendarinvitedelivery.RecurrenceIDEQ(claim.materialization.recurrenceID))
	}
	update.SetState(calendarinvitedelivery.State(resolution.state)).
		ClearLeaseToken().
		ClearLeaseUntil().
		SetLastErrorCode(calendarinvitedelivery.LastErrorCode(resolution.lastErrorCode)).
		SetUpdatedAt(databaseNow)
	if resolution.terminal {
		update.ClearNextAttemptAt().SetTerminalAt(databaseNow)
	} else {
		update.SetNextAttemptAt(*resolution.nextAttemptAt)
	}
	updated, err := update.Save(ctx)
	if err != nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return updated, nil
}

func mailboxEntCASCalendarReplyDeliveryResolve(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarReplyDelivery, claim CalendarMailDeliveryClaim, resolution mailboxEntCalendarDeliveryResolution, databaseNow time.Time) (int, error) {
	if ctx == nil || transaction == nil || delivery == nil || !validCalendarMailDeliveryClaim(claim) || !mailboxEntValidTime(databaseNow) ||
		resolution.lastErrorCode == "" || (resolution.terminal == (resolution.nextAttemptAt != nil)) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	update := transaction.CalendarReplyDelivery.Update().Where(
		calendarreplydelivery.IDEQ(claim.deliveryID.UUID()),
		calendarreplydelivery.OrganizationIDEQ(claim.organizationID),
		calendarreplydelivery.CalendarEventIDEQ(claim.eventID),
		calendarreplydelivery.CalendarAttendeeIDEQ(claim.attendeeID),
		calendarreplydelivery.EventSequenceEQ(claim.materialization.eventSequence),
		calendarreplydelivery.AttendeeVersionEQ(claim.materialization.attendeeVersion),
		calendarreplydelivery.StateEQ(calendarreplydelivery.StateSubmitting),
		calendarreplydelivery.MailSubmissionIDIsNil(),
		calendarreplydelivery.AttemptsEQ(delivery.Attempts),
		calendarreplydelivery.ClaimGenerationEQ(int64(claim.claimGeneration)),
		calendarreplydelivery.LeaseTokenEQ(uuid.UUID(claim.claimToken)),
		calendarreplydelivery.LeaseUntilEQ(claim.leaseUntil),
		calendarreplydelivery.NextAttemptAtIsNil(),
		calendarreplydelivery.LastErrorCodeIsNil(),
	)
	if claim.materialization.recurrenceID == "" {
		update.Where(calendarreplydelivery.RecurrenceIDIsNil())
	} else {
		update.Where(calendarreplydelivery.RecurrenceIDEQ(claim.materialization.recurrenceID))
	}
	update.SetState(calendarreplydelivery.State(resolution.state)).
		ClearLeaseToken().
		ClearLeaseUntil().
		SetLastErrorCode(calendarreplydelivery.LastErrorCode(resolution.lastErrorCode)).
		SetUpdatedAt(databaseNow)
	if resolution.terminal {
		update.ClearNextAttemptAt()
	} else {
		update.SetNextAttemptAt(*resolution.nextAttemptAt)
	}
	updated, err := update.Save(ctx)
	if err != nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return updated, nil
}
