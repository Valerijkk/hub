package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarinvitedelivery"
	"github.com/valerii/hub/backend/ent/calendarreplydelivery"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailsubmission"
)

// BindCalendarMailDeliverySubmission atomically binds one durable Mail submission to its sealed claim.
func (repository *MailboxEntRepository) BindCalendarMailDeliverySubmission(ctx context.Context, request CalendarMailDeliverySubmissionBindRequest) (CalendarMailDeliverySubmissionBindResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || !validCalendarMailDeliverySubmissionBindRequest(request) {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	var result CalendarMailDeliverySubmissionBindResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		var bindErr error
		switch request.Claim.deliveryID.Kind() {
		case CalendarMailDeliveryKindInvite:
			result, bindErr = mailboxEntBindCalendarInviteDeliverySubmission(ctx, transaction, request)
		case CalendarMailDeliveryKindReply:
			result, bindErr = mailboxEntBindCalendarReplyDeliverySubmission(ctx, transaction, request)
		default:
			bindErr = newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
		}
		return bindErr
	})
	if err != nil {
		return CalendarMailDeliverySubmissionBindResult{}, mailboxEntCalendarDeliveryRepositoryError(err)
	}
	if !validCalendarMailDeliverySubmissionBindResult(result) {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return result, nil
}

func mailboxEntCalendarDeliverySubmissionForBind(ctx context.Context, transaction *ent.Tx, claim CalendarMailDeliveryClaim, submissionID uuid.UUID, requireDraft bool) error {
	if ctx == nil || transaction == nil || !validCalendarMailDeliveryClaim(claim) || submissionID == uuid.Nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	idempotencyKey := calendarMailDeliverySubmissionKey(claim.deliveryID)
	if idempotencyKey == "" {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	idempotencyKeyHash := sha256.Sum256([]byte(calendarMailDeliverySubmissionKey(claim.deliveryID)))
	submission, err := transaction.MailSubmission.Query().
		Where(
			mailsubmission.IDEQ(submissionID),
			mailsubmission.MailAccountIDEQ(claim.accountID),
			mailsubmission.IdempotencyKeyHashEQ(idempotencyKeyHash[:]),
		).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if err != nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if !mailboxEntSubmissionRowValid(submission, claim.accountID) || submission.ID != submissionID || submission.MailAccountID != claim.accountID || !bytes.Equal(submission.IdempotencyKeyHash, idempotencyKeyHash[:]) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if !requireDraft {
		return nil
	}
	if submission.MailDraftID == nil || *submission.MailDraftID == uuid.Nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	draft, err := transaction.MailDraft.Query().
		Where(
			maildraft.IDEQ(*submission.MailDraftID),
			maildraft.MailAccountIDEQ(claim.accountID),
			maildraft.DraftClientIDEQ(calendarMailDeliveryDraftClientID(claim.deliveryID)),
		).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if err != nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if !mailboxEntDraftReadRowValid(draft, claim.accountID, *submission.MailDraftID) || draft.DraftClientID != calendarMailDeliveryDraftClientID(claim.deliveryID) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	return nil
}

func mailboxEntBindCalendarInviteDeliverySubmission(ctx context.Context, transaction *ent.Tx, request CalendarMailDeliverySubmissionBindRequest) (CalendarMailDeliverySubmissionBindResult, error) {
	claim := request.Claim
	if ctx == nil || transaction == nil || !validCalendarMailDeliverySubmissionBindRequest(request) || claim.deliveryID.Kind() != CalendarMailDeliveryKindInvite {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	delivery, err := transaction.CalendarInviteDelivery.Query().
		Where(calendarinvitedelivery.IDEQ(claim.deliveryID.UUID())).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if err != nil || delivery == nil || delivery.ID != claim.deliveryID.UUID() {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
	if err != nil || !mailboxEntValidTime(databaseNow) {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	if delivery.MailSubmissionID != nil && *delivery.MailSubmissionID != request.SubmissionID {
		return CalendarMailDeliverySubmissionBindResult{outcome: CalendarMailDeliverySubmissionBindOutcomeConflict}, nil
	}
	if delivery.MailSubmissionID != nil {
		if bindErr := mailboxEntCalendarDeliverySubmissionForBind(ctx, transaction, claim, request.SubmissionID, false); bindErr != nil {
			return CalendarMailDeliverySubmissionBindResult{}, bindErr
		}
		return mailboxEntCalendarInviteDeliveryBindReplay(delivery, claim, request.SubmissionID)
	}
	if bindErr := mailboxEntCalendarDeliverySubmissionForBind(ctx, transaction, claim, request.SubmissionID, true); bindErr != nil {
		return CalendarMailDeliverySubmissionBindResult{}, bindErr
	}
	if !mailboxEntCalendarInviteDeliveryFreshBindMatches(delivery, claim, databaseNow) {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	updated, err := mailboxEntCASCalendarInviteDeliveryBind(ctx, transaction, delivery, claim, request.SubmissionID, databaseNow)
	if err != nil {
		return CalendarMailDeliverySubmissionBindResult{}, err
	}
	if updated != 1 {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return CalendarMailDeliverySubmissionBindResult{outcome: CalendarMailDeliverySubmissionBindOutcomeBound}, nil
}

func mailboxEntBindCalendarReplyDeliverySubmission(ctx context.Context, transaction *ent.Tx, request CalendarMailDeliverySubmissionBindRequest) (CalendarMailDeliverySubmissionBindResult, error) {
	claim := request.Claim
	if ctx == nil || transaction == nil || !validCalendarMailDeliverySubmissionBindRequest(request) || claim.deliveryID.Kind() != CalendarMailDeliveryKindReply {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	delivery, err := transaction.CalendarReplyDelivery.Query().
		Where(calendarreplydelivery.IDEQ(claim.deliveryID.UUID())).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if err != nil || delivery == nil || delivery.ID != claim.deliveryID.UUID() {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
	if err != nil || !mailboxEntValidTime(databaseNow) {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	if delivery.MailSubmissionID != nil && *delivery.MailSubmissionID != request.SubmissionID {
		return CalendarMailDeliverySubmissionBindResult{outcome: CalendarMailDeliverySubmissionBindOutcomeConflict}, nil
	}
	if delivery.MailSubmissionID != nil {
		if bindErr := mailboxEntCalendarDeliverySubmissionForBind(ctx, transaction, claim, request.SubmissionID, false); bindErr != nil {
			return CalendarMailDeliverySubmissionBindResult{}, bindErr
		}
		return mailboxEntCalendarReplyDeliveryBindReplay(delivery, claim, request.SubmissionID)
	}
	if bindErr := mailboxEntCalendarDeliverySubmissionForBind(ctx, transaction, claim, request.SubmissionID, true); bindErr != nil {
		return CalendarMailDeliverySubmissionBindResult{}, bindErr
	}
	if !mailboxEntCalendarReplyDeliveryFreshBindMatches(delivery, claim, databaseNow) {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	updated, err := mailboxEntCASCalendarReplyDeliveryBind(ctx, transaction, delivery, claim, request.SubmissionID, databaseNow)
	if err != nil {
		return CalendarMailDeliverySubmissionBindResult{}, err
	}
	if updated != 1 {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return CalendarMailDeliverySubmissionBindResult{outcome: CalendarMailDeliverySubmissionBindOutcomeBound}, nil
}

func mailboxEntCalendarInviteDeliveryFreshBindMatches(delivery *ent.CalendarInviteDelivery, claim CalendarMailDeliveryClaim, databaseNow time.Time) bool {
	if delivery == nil || !validCalendarMailDeliveryClaim(claim) || claim.deliveryID.Kind() != CalendarMailDeliveryKindInvite ||
		!mailboxEntValidTime(databaseNow) || !mailboxEntCalendarInviteDeliverySubmittingValid(delivery) ||
		delivery.LeaseUntil == nil || !delivery.LeaseUntil.After(databaseNow) {
		return false
	}
	method, ok := mailboxEntCalendarInviteDeliveryMethod(delivery.Kind)
	return ok && method == claim.method &&
		CalendarInviteDeliveryID(delivery.ID) == claim.inviteID &&
		delivery.OrganizationID == claim.organizationID &&
		delivery.CalendarEventID == claim.eventID &&
		delivery.CalendarAttendeeID == claim.attendeeID &&
		delivery.EventSequence == claim.materialization.eventSequence &&
		mailboxEntCalendarDeliveryRecurrenceMatches(delivery.RecurrenceID, claim.materialization.recurrenceID) &&
		delivery.ClaimGeneration == int64(claim.claimGeneration) &&
		delivery.LeaseToken != nil && *delivery.LeaseToken == uuid.UUID(claim.claimToken) &&
		delivery.LeaseUntil.Equal(claim.leaseUntil)
}

func mailboxEntCalendarReplyDeliveryFreshBindMatches(delivery *ent.CalendarReplyDelivery, claim CalendarMailDeliveryClaim, databaseNow time.Time) bool {
	if delivery == nil || !validCalendarMailDeliveryClaim(claim) || claim.deliveryID.Kind() != CalendarMailDeliveryKindReply ||
		!mailboxEntValidTime(databaseNow) || !mailboxEntCalendarReplyDeliverySubmittingValid(delivery) ||
		delivery.LeaseUntil == nil || !delivery.LeaseUntil.After(databaseNow) {
		return false
	}
	return claim.method == CalendarMailDeliveryMethodReply &&
		CalendarReplyDeliveryID(delivery.ID) == claim.replyID &&
		delivery.OrganizationID == claim.organizationID &&
		delivery.CalendarEventID == claim.eventID &&
		delivery.CalendarAttendeeID == claim.attendeeID &&
		delivery.EventSequence == claim.materialization.eventSequence &&
		delivery.AttendeeVersion == claim.materialization.attendeeVersion &&
		mailboxEntCalendarDeliveryRecurrenceMatches(delivery.RecurrenceID, claim.materialization.recurrenceID) &&
		delivery.ClaimGeneration == int64(claim.claimGeneration) &&
		delivery.LeaseToken != nil && *delivery.LeaseToken == uuid.UUID(claim.claimToken) &&
		delivery.LeaseUntil.Equal(claim.leaseUntil)
}

func mailboxEntCalendarInviteDeliveryBindReplay(delivery *ent.CalendarInviteDelivery, claim CalendarMailDeliveryClaim, submissionID uuid.UUID) (CalendarMailDeliverySubmissionBindResult, error) {
	if delivery == nil || !validCalendarMailDeliveryClaim(claim) || claim.deliveryID.Kind() != CalendarMailDeliveryKindInvite || delivery.MailSubmissionID == nil {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if *delivery.MailSubmissionID != submissionID {
		return CalendarMailDeliverySubmissionBindResult{outcome: CalendarMailDeliverySubmissionBindOutcomeConflict}, nil
	}
	if !mailboxEntCalendarInviteDeliveryBoundScopeMatches(delivery, claim) {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	switch CalendarMailDeliveryState(delivery.State) {
	case CalendarMailDeliveryStateUnknown:
		if !mailboxEntCalendarInviteDeliveryUnknownBoundValid(delivery) {
			return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
	case CalendarMailDeliveryStateSent:
		if !mailboxEntCalendarInviteDeliveryTerminalValid(delivery) {
			return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
	default:
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return CalendarMailDeliverySubmissionBindResult{outcome: CalendarMailDeliverySubmissionBindOutcomeReplay}, nil
}

func mailboxEntCalendarReplyDeliveryBindReplay(delivery *ent.CalendarReplyDelivery, claim CalendarMailDeliveryClaim, submissionID uuid.UUID) (CalendarMailDeliverySubmissionBindResult, error) {
	if delivery == nil || !validCalendarMailDeliveryClaim(claim) || claim.deliveryID.Kind() != CalendarMailDeliveryKindReply || delivery.MailSubmissionID == nil {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if *delivery.MailSubmissionID != submissionID {
		return CalendarMailDeliverySubmissionBindResult{outcome: CalendarMailDeliverySubmissionBindOutcomeConflict}, nil
	}
	if !mailboxEntCalendarReplyDeliveryBoundScopeMatches(delivery, claim) {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	switch CalendarMailDeliveryState(delivery.State) {
	case CalendarMailDeliveryStateUnknown:
		if !mailboxEntCalendarReplyDeliveryUnknownBoundValid(delivery) {
			return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
	case CalendarMailDeliveryStateSent:
		if !mailboxEntCalendarReplyDeliveryTerminalValid(delivery) {
			return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
	default:
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return CalendarMailDeliverySubmissionBindResult{outcome: CalendarMailDeliverySubmissionBindOutcomeReplay}, nil
}

func mailboxEntCalendarInviteDeliveryBoundScopeMatches(delivery *ent.CalendarInviteDelivery, claim CalendarMailDeliveryClaim) bool {
	if delivery == nil || !validCalendarMailDeliveryClaim(claim) || claim.deliveryID.Kind() != CalendarMailDeliveryKindInvite {
		return false
	}
	method, ok := mailboxEntCalendarInviteDeliveryMethod(delivery.Kind)
	return ok && method == claim.method &&
		CalendarInviteDeliveryID(delivery.ID) == claim.inviteID &&
		delivery.OrganizationID == claim.organizationID &&
		delivery.CalendarEventID == claim.eventID &&
		delivery.CalendarAttendeeID == claim.attendeeID &&
		delivery.EventSequence == claim.materialization.eventSequence &&
		mailboxEntCalendarDeliveryRecurrenceMatches(delivery.RecurrenceID, claim.materialization.recurrenceID) &&
		delivery.ClaimGeneration == int64(claim.claimGeneration)
}

func mailboxEntCalendarReplyDeliveryBoundScopeMatches(delivery *ent.CalendarReplyDelivery, claim CalendarMailDeliveryClaim) bool {
	return delivery != nil && validCalendarMailDeliveryClaim(claim) && claim.deliveryID.Kind() == CalendarMailDeliveryKindReply &&
		claim.method == CalendarMailDeliveryMethodReply &&
		CalendarReplyDeliveryID(delivery.ID) == claim.replyID &&
		delivery.OrganizationID == claim.organizationID &&
		delivery.CalendarEventID == claim.eventID &&
		delivery.CalendarAttendeeID == claim.attendeeID &&
		delivery.EventSequence == claim.materialization.eventSequence &&
		delivery.AttendeeVersion == claim.materialization.attendeeVersion &&
		mailboxEntCalendarDeliveryRecurrenceMatches(delivery.RecurrenceID, claim.materialization.recurrenceID) &&
		delivery.ClaimGeneration == int64(claim.claimGeneration)
}

func mailboxEntCASCalendarInviteDeliveryBind(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery, claim CalendarMailDeliveryClaim, submissionID uuid.UUID, databaseNow time.Time) (int, error) {
	if ctx == nil || transaction == nil || delivery == nil || !validCalendarMailDeliveryClaim(claim) ||
		claim.deliveryID.Kind() != CalendarMailDeliveryKindInvite || submissionID == uuid.Nil ||
		!mailboxEntValidTime(databaseNow) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	updated, err := transaction.CalendarInviteDelivery.Update().
		Where(
			calendarinvitedelivery.IDEQ(claim.deliveryID.UUID()),
			calendarinvitedelivery.OrganizationIDEQ(claim.organizationID),
			calendarinvitedelivery.CalendarEventIDEQ(claim.eventID),
			calendarinvitedelivery.CalendarAttendeeIDEQ(claim.attendeeID),
			calendarinvitedelivery.KindEQ(delivery.Kind),
			calendarinvitedelivery.EventSequenceEQ(claim.materialization.eventSequence),
			calendarinvitedelivery.RecurrenceIDIsNil(),
			calendarinvitedelivery.StateEQ(calendarinvitedelivery.State(CalendarMailDeliveryStateSubmitting)),
			calendarinvitedelivery.MailSubmissionIDIsNil(),
			calendarinvitedelivery.AttemptsEQ(delivery.Attempts),
			calendarinvitedelivery.ClaimGenerationEQ(int64(claim.claimGeneration)),
			calendarinvitedelivery.LeaseTokenEQ(uuid.UUID(claim.claimToken)),
			calendarinvitedelivery.LeaseUntilEQ(claim.leaseUntil),
			calendarinvitedelivery.NextAttemptAtIsNil(),
			calendarinvitedelivery.LastErrorCodeIsNil(),
			calendarinvitedelivery.TerminalAtIsNil(),
		).
		SetMailSubmissionID(submissionID).
		SetState(calendarinvitedelivery.State(CalendarMailDeliveryStateUnknown)).
		ClearLeaseToken().
		ClearLeaseUntil().
		SetLastErrorCode(calendarinvitedelivery.LastErrorCode("submission_outcome_unknown")).
		SetNextAttemptAt(databaseNow).
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return updated, nil
}

func mailboxEntCASCalendarReplyDeliveryBind(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarReplyDelivery, claim CalendarMailDeliveryClaim, submissionID uuid.UUID, databaseNow time.Time) (int, error) {
	if ctx == nil || transaction == nil || delivery == nil || !validCalendarMailDeliveryClaim(claim) ||
		claim.deliveryID.Kind() != CalendarMailDeliveryKindReply || submissionID == uuid.Nil ||
		!mailboxEntValidTime(databaseNow) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	updated, err := transaction.CalendarReplyDelivery.Update().
		Where(
			calendarreplydelivery.IDEQ(claim.deliveryID.UUID()),
			calendarreplydelivery.OrganizationIDEQ(claim.organizationID),
			calendarreplydelivery.CalendarEventIDEQ(claim.eventID),
			calendarreplydelivery.CalendarAttendeeIDEQ(claim.attendeeID),
			calendarreplydelivery.EventSequenceEQ(claim.materialization.eventSequence),
			calendarreplydelivery.AttendeeVersionEQ(claim.materialization.attendeeVersion),
			calendarreplydelivery.RecurrenceIDIsNil(),
			calendarreplydelivery.StateEQ(calendarreplydelivery.State(CalendarMailDeliveryStateSubmitting)),
			calendarreplydelivery.MailSubmissionIDIsNil(),
			calendarreplydelivery.AttemptsEQ(delivery.Attempts),
			calendarreplydelivery.ClaimGenerationEQ(int64(claim.claimGeneration)),
			calendarreplydelivery.LeaseTokenEQ(uuid.UUID(claim.claimToken)),
			calendarreplydelivery.LeaseUntilEQ(claim.leaseUntil),
			calendarreplydelivery.NextAttemptAtIsNil(),
			calendarreplydelivery.LastErrorCodeIsNil(),
		).
		SetMailSubmissionID(submissionID).
		SetState(calendarreplydelivery.State(CalendarMailDeliveryStateUnknown)).
		ClearLeaseToken().
		ClearLeaseUntil().
		SetLastErrorCode(calendarreplydelivery.LastErrorCode("submission_outcome_unknown")).
		SetNextAttemptAt(databaseNow).
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return updated, nil
}
