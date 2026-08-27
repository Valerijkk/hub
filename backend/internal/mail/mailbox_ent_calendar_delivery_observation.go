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
	"github.com/valerii/hub/backend/ent/mailsubmission"
)

const mailboxEntCalendarDeliveryObservationDelay = 30 * time.Second

type mailboxEntCalendarDeliveryObservationKind uint8

const (
	mailboxEntCalendarDeliveryObservationReschedule mailboxEntCalendarDeliveryObservationKind = iota + 1
	mailboxEntCalendarDeliveryObservationSent
	mailboxEntCalendarDeliveryObservationFailedPermanent
)

type mailboxEntCalendarDeliveryObservation struct {
	kind               mailboxEntCalendarDeliveryObservationKind
	permanentErrorCode string
}

func mailboxEntObserveCalendarInviteDeliveryBoundSubmission(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery, deliveryID CalendarMailDeliveryID, databaseNow time.Time) (CalendarMailDeliveryClaimResult, error) {
	if ctx == nil || transaction == nil || delivery == nil || deliveryID.Kind() != CalendarMailDeliveryKindInvite ||
		!mailboxEntValidTime(databaseNow) || !mailboxEntCalendarInviteDeliveryUnknownBoundValid(delivery) ||
		delivery.ID != deliveryID.UUID() || delivery.MailSubmissionID == nil || delivery.NextAttemptAt == nil {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if delivery.NextAttemptAt.After(databaseNow) {
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
	}
	submission, err := mailboxEntCalendarDeliveryBoundSubmissionForObservation(ctx, transaction, deliveryID, *delivery.MailSubmissionID)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	observation, err := mailboxEntCalendarDeliveryObservationFromSubmission(submission)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	updated, err := mailboxEntCASCalendarInviteDeliveryObservation(ctx, transaction, delivery, submission, databaseNow, observation)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	if updated != 1 {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeObserved}, nil
}

func mailboxEntObserveCalendarReplyDeliveryBoundSubmission(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarReplyDelivery, deliveryID CalendarMailDeliveryID, databaseNow time.Time) (CalendarMailDeliveryClaimResult, error) {
	if ctx == nil || transaction == nil || delivery == nil || deliveryID.Kind() != CalendarMailDeliveryKindReply ||
		!mailboxEntValidTime(databaseNow) || !mailboxEntCalendarReplyDeliveryUnknownBoundValid(delivery) ||
		delivery.ID != deliveryID.UUID() || delivery.MailSubmissionID == nil || delivery.NextAttemptAt == nil {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if delivery.NextAttemptAt.After(databaseNow) {
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
	}
	submission, err := mailboxEntCalendarDeliveryBoundSubmissionForObservation(ctx, transaction, deliveryID, *delivery.MailSubmissionID)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	observation, err := mailboxEntCalendarDeliveryObservationFromSubmission(submission)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	updated, err := mailboxEntCASCalendarReplyDeliveryObservation(ctx, transaction, delivery, submission, databaseNow, observation)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	if updated != 1 {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeObserved}, nil
}

func mailboxEntCalendarDeliveryBoundSubmissionForObservation(ctx context.Context, transaction *ent.Tx, deliveryID CalendarMailDeliveryID, submissionID uuid.UUID) (*ent.MailSubmission, error) {
	if ctx == nil || transaction == nil || !validCalendarMailDeliveryID(deliveryID) || submissionID == uuid.Nil {
		return nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	idempotencyKeyHash := sha256.Sum256([]byte(calendarMailDeliverySubmissionKey(deliveryID)))
	submission, err := transaction.MailSubmission.Query().
		Where(
			mailsubmission.IDEQ(submissionID),
			mailsubmission.IdempotencyKeyHashEQ(idempotencyKeyHash[:]...),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || submission == nil || submission.ID != submissionID || submission.MailAccountID == uuid.Nil ||
		!mailboxEntSubmissionRowValid(submission, submission.MailAccountID) ||
		!bytes.Equal(submission.IdempotencyKeyHash, idempotencyKeyHash[:]) {
		return nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return submission, nil
}

func mailboxEntCalendarDeliveryObservationFromSubmission(submission *ent.MailSubmission) (mailboxEntCalendarDeliveryObservation, error) {
	if submission == nil || submission.MailAccountID == uuid.Nil || !mailboxEntSubmissionRowValid(submission, submission.MailAccountID) {
		return mailboxEntCalendarDeliveryObservation{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	switch MailboxSubmissionStatus(submission.Status) {
	case MailboxSubmissionStatusQueued,
		MailboxSubmissionStatusAuthorizing,
		MailboxSubmissionStatusSubmitting,
		MailboxSubmissionStatusAccepted:
		if submission.LastErrorCode != nil || submission.ResultMailMessageID != nil {
			return mailboxEntCalendarDeliveryObservation{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return mailboxEntCalendarDeliveryObservation{kind: mailboxEntCalendarDeliveryObservationReschedule}, nil
	case MailboxSubmissionStatusUnknown:
		if submission.LastErrorCode == nil || submission.ResultMailMessageID != nil {
			return mailboxEntCalendarDeliveryObservation{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		switch MailboxSubmissionErrorCode(*submission.LastErrorCode) {
		case MailboxSubmissionErrorCodeOutcomeUnknown, MailboxSubmissionErrorCodeReconciliationRequired:
			return mailboxEntCalendarDeliveryObservation{kind: mailboxEntCalendarDeliveryObservationReschedule}, nil
		default:
			return mailboxEntCalendarDeliveryObservation{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
	case MailboxSubmissionStatusSent:
		if submission.LastErrorCode != nil || submission.ResultMailMessageID == nil || *submission.ResultMailMessageID == uuid.Nil {
			return mailboxEntCalendarDeliveryObservation{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return mailboxEntCalendarDeliveryObservation{kind: mailboxEntCalendarDeliveryObservationSent}, nil
	case MailboxSubmissionStatusFailed:
		if submission.LastErrorCode == nil || submission.ResultMailMessageID != nil {
			return mailboxEntCalendarDeliveryObservation{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		switch MailboxSubmissionErrorCode(*submission.LastErrorCode) {
		case MailboxSubmissionErrorCodeRejected, MailboxSubmissionErrorCodeSizeLimit:
			return mailboxEntCalendarDeliveryObservation{
				kind:               mailboxEntCalendarDeliveryObservationFailedPermanent,
				permanentErrorCode: "submission_rejected",
			}, nil
		case MailboxSubmissionErrorCodePolicy:
			return mailboxEntCalendarDeliveryObservation{
				kind:               mailboxEntCalendarDeliveryObservationFailedPermanent,
				permanentErrorCode: "policy_denied",
			}, nil
		case MailboxSubmissionErrorCodeAttachmentNotClean,
			MailboxSubmissionErrorCodeIdentityUnavailable,
			MailboxSubmissionErrorCodeRecipientLimit:
			return mailboxEntCalendarDeliveryObservation{
				kind:               mailboxEntCalendarDeliveryObservationFailedPermanent,
				permanentErrorCode: "materialization_invalid",
			}, nil
		case MailboxSubmissionErrorCodeRetryExhausted,
			MailboxSubmissionErrorCodeReconciliationExhausted,
			MailboxSubmissionErrorCodeMailboxUnavailable:
			return mailboxEntCalendarDeliveryObservation{
				kind:               mailboxEntCalendarDeliveryObservationFailedPermanent,
				permanentErrorCode: "retry_exhausted",
			}, nil
		default:
			return mailboxEntCalendarDeliveryObservation{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
	default:
		return mailboxEntCalendarDeliveryObservation{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
}

func mailboxEntCalendarDeliveryObservationPermanentErrorValid(code string) bool {
	switch code {
	case "submission_rejected", "policy_denied", "materialization_invalid", "retry_exhausted":
		return true
	default:
		return false
	}
}

func mailboxEntCASCalendarInviteDeliveryObservation(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery, submission *ent.MailSubmission, databaseNow time.Time, observation mailboxEntCalendarDeliveryObservation) (int, error) {
	if ctx == nil || transaction == nil || delivery == nil || submission == nil || !mailboxEntValidTime(databaseNow) ||
		!mailboxEntCalendarInviteDeliveryUnknownBoundValid(delivery) || delivery.MailSubmissionID == nil ||
		*delivery.MailSubmissionID != submission.ID || delivery.NextAttemptAt == nil || delivery.NextAttemptAt.After(databaseNow) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	switch observation.kind {
	case mailboxEntCalendarDeliveryObservationReschedule:
		nextAttemptAt, ok := mailboxEntCheckedTimeAdd(databaseNow, mailboxEntCalendarDeliveryObservationDelay)
		if !ok || !nextAttemptAt.After(databaseNow) {
			return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		updated, err := transaction.CalendarInviteDelivery.Update().
			Where(
				calendarinvitedelivery.IDEQ(delivery.ID),
				calendarinvitedelivery.OrganizationIDEQ(delivery.OrganizationID),
				calendarinvitedelivery.CalendarEventIDEQ(delivery.CalendarEventID),
				calendarinvitedelivery.CalendarAttendeeIDEQ(delivery.CalendarAttendeeID),
				calendarinvitedelivery.KindEQ(delivery.Kind),
				calendarinvitedelivery.EventSequenceEQ(delivery.EventSequence),
				calendarinvitedelivery.RecurrenceIDIsNil(),
				calendarinvitedelivery.StateEQ(calendarinvitedelivery.State(CalendarMailDeliveryStateUnknown)),
				calendarinvitedelivery.MailSubmissionIDEQ(submission.ID),
				calendarinvitedelivery.AttemptsEQ(delivery.Attempts),
				calendarinvitedelivery.ClaimGenerationEQ(delivery.ClaimGeneration),
				calendarinvitedelivery.LeaseTokenIsNil(),
				calendarinvitedelivery.LeaseUntilIsNil(),
				calendarinvitedelivery.NextAttemptAtNotNil(),
				calendarinvitedelivery.NextAttemptAtLTE(databaseNow),
				calendarinvitedelivery.LastErrorCodeEQ(*delivery.LastErrorCode),
				calendarinvitedelivery.TerminalAtIsNil(),
			).
			SetState(calendarinvitedelivery.State(CalendarMailDeliveryStateUnknown)).
			SetLastErrorCode(calendarinvitedelivery.LastErrorCode("submission_outcome_unknown")).
			SetNextAttemptAt(nextAttemptAt).
			SetUpdatedAt(databaseNow).
			Save(ctx)
		if err != nil {
			return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return updated, nil
	case mailboxEntCalendarDeliveryObservationSent:
		updated, err := transaction.CalendarInviteDelivery.Update().
			Where(
				calendarinvitedelivery.IDEQ(delivery.ID),
				calendarinvitedelivery.OrganizationIDEQ(delivery.OrganizationID),
				calendarinvitedelivery.CalendarEventIDEQ(delivery.CalendarEventID),
				calendarinvitedelivery.CalendarAttendeeIDEQ(delivery.CalendarAttendeeID),
				calendarinvitedelivery.KindEQ(delivery.Kind),
				calendarinvitedelivery.EventSequenceEQ(delivery.EventSequence),
				calendarinvitedelivery.RecurrenceIDIsNil(),
				calendarinvitedelivery.StateEQ(calendarinvitedelivery.State(CalendarMailDeliveryStateUnknown)),
				calendarinvitedelivery.MailSubmissionIDEQ(submission.ID),
				calendarinvitedelivery.AttemptsEQ(delivery.Attempts),
				calendarinvitedelivery.ClaimGenerationEQ(delivery.ClaimGeneration),
				calendarinvitedelivery.LeaseTokenIsNil(),
				calendarinvitedelivery.LeaseUntilIsNil(),
				calendarinvitedelivery.NextAttemptAtNotNil(),
				calendarinvitedelivery.NextAttemptAtLTE(databaseNow),
				calendarinvitedelivery.LastErrorCodeEQ(*delivery.LastErrorCode),
				calendarinvitedelivery.TerminalAtIsNil(),
			).
			SetState(calendarinvitedelivery.State(CalendarMailDeliveryStateSent)).
			ClearLastErrorCode().
			ClearNextAttemptAt().
			SetTerminalAt(databaseNow).
			SetUpdatedAt(databaseNow).
			Save(ctx)
		if err != nil {
			return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return updated, nil
	case mailboxEntCalendarDeliveryObservationFailedPermanent:
		if !mailboxEntCalendarDeliveryObservationPermanentErrorValid(observation.permanentErrorCode) {
			return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		updated, err := transaction.CalendarInviteDelivery.Update().
			Where(
				calendarinvitedelivery.IDEQ(delivery.ID),
				calendarinvitedelivery.OrganizationIDEQ(delivery.OrganizationID),
				calendarinvitedelivery.CalendarEventIDEQ(delivery.CalendarEventID),
				calendarinvitedelivery.CalendarAttendeeIDEQ(delivery.CalendarAttendeeID),
				calendarinvitedelivery.KindEQ(delivery.Kind),
				calendarinvitedelivery.EventSequenceEQ(delivery.EventSequence),
				calendarinvitedelivery.RecurrenceIDIsNil(),
				calendarinvitedelivery.StateEQ(calendarinvitedelivery.State(CalendarMailDeliveryStateUnknown)),
				calendarinvitedelivery.MailSubmissionIDEQ(submission.ID),
				calendarinvitedelivery.AttemptsEQ(delivery.Attempts),
				calendarinvitedelivery.ClaimGenerationEQ(delivery.ClaimGeneration),
				calendarinvitedelivery.LeaseTokenIsNil(),
				calendarinvitedelivery.LeaseUntilIsNil(),
				calendarinvitedelivery.NextAttemptAtNotNil(),
				calendarinvitedelivery.NextAttemptAtLTE(databaseNow),
				calendarinvitedelivery.LastErrorCodeEQ(*delivery.LastErrorCode),
				calendarinvitedelivery.TerminalAtIsNil(),
			).
			SetState(calendarinvitedelivery.State(CalendarMailDeliveryStateFailedPermanent)).
			SetLastErrorCode(calendarinvitedelivery.LastErrorCode(observation.permanentErrorCode)).
			ClearNextAttemptAt().
			SetTerminalAt(databaseNow).
			SetUpdatedAt(databaseNow).
			Save(ctx)
		if err != nil {
			return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return updated, nil
	default:
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
}

func mailboxEntCASCalendarReplyDeliveryObservation(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarReplyDelivery, submission *ent.MailSubmission, databaseNow time.Time, observation mailboxEntCalendarDeliveryObservation) (int, error) {
	if ctx == nil || transaction == nil || delivery == nil || submission == nil || !mailboxEntValidTime(databaseNow) ||
		!mailboxEntCalendarReplyDeliveryUnknownBoundValid(delivery) || delivery.MailSubmissionID == nil ||
		*delivery.MailSubmissionID != submission.ID || delivery.NextAttemptAt == nil || delivery.NextAttemptAt.After(databaseNow) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	switch observation.kind {
	case mailboxEntCalendarDeliveryObservationReschedule:
		nextAttemptAt, ok := mailboxEntCheckedTimeAdd(databaseNow, mailboxEntCalendarDeliveryObservationDelay)
		if !ok || !nextAttemptAt.After(databaseNow) {
			return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		updated, err := transaction.CalendarReplyDelivery.Update().
			Where(
				calendarreplydelivery.IDEQ(delivery.ID),
				calendarreplydelivery.OrganizationIDEQ(delivery.OrganizationID),
				calendarreplydelivery.CalendarEventIDEQ(delivery.CalendarEventID),
				calendarreplydelivery.CalendarAttendeeIDEQ(delivery.CalendarAttendeeID),
				calendarreplydelivery.EventSequenceEQ(delivery.EventSequence),
				calendarreplydelivery.AttendeeVersionEQ(delivery.AttendeeVersion),
				calendarreplydelivery.RecurrenceIDIsNil(),
				calendarreplydelivery.StateEQ(calendarreplydelivery.State(CalendarMailDeliveryStateUnknown)),
				calendarreplydelivery.MailSubmissionIDEQ(submission.ID),
				calendarreplydelivery.AttemptsEQ(delivery.Attempts),
				calendarreplydelivery.ClaimGenerationEQ(delivery.ClaimGeneration),
				calendarreplydelivery.LeaseTokenIsNil(),
				calendarreplydelivery.LeaseUntilIsNil(),
				calendarreplydelivery.NextAttemptAtNotNil(),
				calendarreplydelivery.NextAttemptAtLTE(databaseNow),
				calendarreplydelivery.LastErrorCodeEQ(*delivery.LastErrorCode),
			).
			SetState(calendarreplydelivery.State(CalendarMailDeliveryStateUnknown)).
			SetLastErrorCode(calendarreplydelivery.LastErrorCode("submission_outcome_unknown")).
			SetNextAttemptAt(nextAttemptAt).
			SetUpdatedAt(databaseNow).
			Save(ctx)
		if err != nil {
			return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return updated, nil
	case mailboxEntCalendarDeliveryObservationSent:
		updated, err := transaction.CalendarReplyDelivery.Update().
			Where(
				calendarreplydelivery.IDEQ(delivery.ID),
				calendarreplydelivery.OrganizationIDEQ(delivery.OrganizationID),
				calendarreplydelivery.CalendarEventIDEQ(delivery.CalendarEventID),
				calendarreplydelivery.CalendarAttendeeIDEQ(delivery.CalendarAttendeeID),
				calendarreplydelivery.EventSequenceEQ(delivery.EventSequence),
				calendarreplydelivery.AttendeeVersionEQ(delivery.AttendeeVersion),
				calendarreplydelivery.RecurrenceIDIsNil(),
				calendarreplydelivery.StateEQ(calendarreplydelivery.State(CalendarMailDeliveryStateUnknown)),
				calendarreplydelivery.MailSubmissionIDEQ(submission.ID),
				calendarreplydelivery.AttemptsEQ(delivery.Attempts),
				calendarreplydelivery.ClaimGenerationEQ(delivery.ClaimGeneration),
				calendarreplydelivery.LeaseTokenIsNil(),
				calendarreplydelivery.LeaseUntilIsNil(),
				calendarreplydelivery.NextAttemptAtNotNil(),
				calendarreplydelivery.NextAttemptAtLTE(databaseNow),
				calendarreplydelivery.LastErrorCodeEQ(*delivery.LastErrorCode),
			).
			SetState(calendarreplydelivery.State(CalendarMailDeliveryStateSent)).
			ClearLastErrorCode().
			ClearNextAttemptAt().
			SetUpdatedAt(databaseNow).
			Save(ctx)
		if err != nil {
			return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return updated, nil
	case mailboxEntCalendarDeliveryObservationFailedPermanent:
		if !mailboxEntCalendarDeliveryObservationPermanentErrorValid(observation.permanentErrorCode) {
			return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		updated, err := transaction.CalendarReplyDelivery.Update().
			Where(
				calendarreplydelivery.IDEQ(delivery.ID),
				calendarreplydelivery.OrganizationIDEQ(delivery.OrganizationID),
				calendarreplydelivery.CalendarEventIDEQ(delivery.CalendarEventID),
				calendarreplydelivery.CalendarAttendeeIDEQ(delivery.CalendarAttendeeID),
				calendarreplydelivery.EventSequenceEQ(delivery.EventSequence),
				calendarreplydelivery.AttendeeVersionEQ(delivery.AttendeeVersion),
				calendarreplydelivery.RecurrenceIDIsNil(),
				calendarreplydelivery.StateEQ(calendarreplydelivery.State(CalendarMailDeliveryStateUnknown)),
				calendarreplydelivery.MailSubmissionIDEQ(submission.ID),
				calendarreplydelivery.AttemptsEQ(delivery.Attempts),
				calendarreplydelivery.ClaimGenerationEQ(delivery.ClaimGeneration),
				calendarreplydelivery.LeaseTokenIsNil(),
				calendarreplydelivery.LeaseUntilIsNil(),
				calendarreplydelivery.NextAttemptAtNotNil(),
				calendarreplydelivery.NextAttemptAtLTE(databaseNow),
				calendarreplydelivery.LastErrorCodeEQ(*delivery.LastErrorCode),
			).
			SetState(calendarreplydelivery.State(CalendarMailDeliveryStateFailedPermanent)).
			SetLastErrorCode(calendarreplydelivery.LastErrorCode(observation.permanentErrorCode)).
			ClearNextAttemptAt().
			SetUpdatedAt(databaseNow).
			Save(ctx)
		if err != nil {
			return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return updated, nil
	default:
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
}
