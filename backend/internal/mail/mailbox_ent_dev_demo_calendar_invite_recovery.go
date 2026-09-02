//go:build !qa && !release

package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendarinvitedelivery"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailsubmission"
)

// DevDemoCalendarInviteRecoveryRequest is the caller-proved immutable master
// identity of the one development Calendar invitation eligible for recovery.
// The concrete repository still locks and rechecks every fact before mutation.
type DevDemoCalendarInviteRecoveryRequest struct {
	OrganizationID uuid.UUID
	EventID        uuid.UUID
	AttendeeID     uuid.UUID
	OrganizerID    uuid.UUID
	EventSequence  int
	Kind           CalendarInviteDeliveryKind
	RecurrenceID   string
}

// RecoverDevDemoCalendarInviteDelivery removes only the exact rejected
// submission and attachment-less Calendar draft bound to one failed fixture
// invite, then returns that same delivery to its initial pending shape.
func (repository *MailboxEntRepository) RecoverDevDemoCalendarInviteDelivery(ctx context.Context, request DevDemoCalendarInviteRecoveryRequest) error {
	if repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || !validDevDemoCalendarInviteRecoveryRequest(request) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		return mailboxEntRecoverDevDemoCalendarInviteDelivery(ctx, transaction, request)
	}); err != nil {
		return mailboxEntCalendarDeliveryRepositoryError(err)
	}
	return nil
}

func validDevDemoCalendarInviteRecoveryRequest(request DevDemoCalendarInviteRecoveryRequest) bool {
	return request.OrganizationID != uuid.Nil && request.EventID != uuid.Nil && request.AttendeeID != uuid.Nil &&
		request.OrganizerID != uuid.Nil && request.EventSequence >= 0 && request.Kind == CalendarInviteDeliveryKindCreated &&
		request.RecurrenceID == ""
}

func mailboxEntRecoverDevDemoCalendarInviteDelivery(ctx context.Context, transaction *ent.Tx, request DevDemoCalendarInviteRecoveryRequest) error {
	if ctx == nil || transaction == nil || !validDevDemoCalendarInviteRecoveryRequest(request) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
	if err != nil || !mailboxEntValidTime(databaseNow) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	event, err := transaction.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(request.EventID),
			calendarevent.OrganizationIDEQ(request.OrganizationID),
			calendarevent.OrganizerIDEQ(request.OrganizerID),
			calendarevent.SequenceEQ(request.EventSequence),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || event == nil || event.ID != request.EventID || event.OrganizationID != request.OrganizationID ||
		event.OrganizerID == nil || *event.OrganizerID != request.OrganizerID || event.Sequence != request.EventSequence {
		return mailboxEntDevDemoCalendarInviteRecoveryRowError(err)
	}

	attendee, err := transaction.CalendarAttendee.Query().
		Where(
			calendarattendee.IDEQ(request.AttendeeID),
			calendarattendee.CalendarEventIDEQ(event.ID),
			calendarattendee.RecurrenceIDIsNil(),
			calendarattendee.RemovedAtIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || attendee == nil || attendee.ID != request.AttendeeID || attendee.CalendarEventID != event.ID ||
		attendee.RecurrenceID != nil || attendee.RemovedAt != nil {
		return mailboxEntDevDemoCalendarInviteRecoveryRowError(err)
	}

	deliveries, err := transaction.CalendarInviteDelivery.Query().
		Where(
			calendarinvitedelivery.OrganizationIDEQ(request.OrganizationID),
			calendarinvitedelivery.CalendarEventIDEQ(event.ID),
			calendarinvitedelivery.CalendarAttendeeIDEQ(attendee.ID),
			calendarinvitedelivery.KindEQ(calendarinvitedelivery.Kind(request.Kind)),
			calendarinvitedelivery.EventSequenceEQ(request.EventSequence),
			calendarinvitedelivery.RecurrenceIDIsNil(),
		).
		ForUpdate().
		All(ctx)
	if err != nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if len(deliveries) != 1 || !mailboxEntDevDemoCalendarInviteFailedValid(deliveries[0], request) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	delivery := deliveries[0]
	deliveryID, err := NewCalendarMailDeliveryID(CalendarMailDeliveryKindInvite, delivery.ID)
	if err != nil || delivery.MailSubmissionID == nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	submission, err := mailboxEntCalendarDeliveryBoundSubmissionForObservation(ctx, transaction, deliveryID, *delivery.MailSubmissionID)
	if err != nil || !mailboxEntDevDemoCalendarInviteRejectedSubmissionValid(submission, deliveryID) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	draft, record, err := mailboxEntDevDemoCalendarInviteFixtureDraft(ctx, transaction, deliveryID, submission)
	if err != nil {
		return err
	}
	if draft == nil || record.ConfirmedMutationID == nil ||
		!mailboxEntDraftCalendarCreateTerminalRecordValid(record, *record.ConfirmedMutationID) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	if err := mailboxEntDevDemoCalendarInviteDetachSubmission(ctx, transaction, delivery, request, submission.ID, databaseNow); err != nil {
		return err
	}
	if err := mailboxEntDevDemoCalendarInviteDeleteSubmission(ctx, transaction, submission, deliveryID); err != nil {
		return err
	}
	if err := mailboxEntDevDemoCalendarInviteDeleteDraft(ctx, transaction, draft, deliveryID); err != nil {
		return err
	}
	return mailboxEntDevDemoCalendarInviteRequeue(ctx, transaction, delivery, request, databaseNow)
}

func mailboxEntDevDemoCalendarInviteFailedValid(delivery *ent.CalendarInviteDelivery, request DevDemoCalendarInviteRecoveryRequest) bool {
	return delivery != nil && delivery.ID != uuid.Nil && delivery.SourceEnvelopeID != uuid.Nil &&
		delivery.OrganizationID == request.OrganizationID && delivery.CalendarEventID == request.EventID &&
		delivery.CalendarAttendeeID == request.AttendeeID && delivery.Kind == calendarinvitedelivery.Kind(request.Kind) &&
		delivery.EventSequence == request.EventSequence && delivery.RecurrenceID == nil &&
		CalendarMailDeliveryState(delivery.State) == CalendarMailDeliveryStateFailedPermanent &&
		delivery.MailSubmissionID != nil && *delivery.MailSubmissionID != uuid.Nil &&
		delivery.Attempts >= 1 && delivery.Attempts <= calendarMailDeliveryMaximumAttempts &&
		delivery.ClaimGeneration >= 1 && delivery.ClaimGeneration <= mailboxEntCalendarDeliveryMaximumClaimGeneration &&
		delivery.LeaseToken == nil && delivery.LeaseUntil == nil && delivery.NextAttemptAt == nil &&
		delivery.LastErrorCode != nil && *delivery.LastErrorCode == calendarinvitedelivery.LastErrorCode("submission_rejected") &&
		delivery.TerminalAt != nil && mailboxEntCalendarDeliveryTimestampValid(*delivery.TerminalAt)
}

func mailboxEntDevDemoCalendarInviteRejectedSubmissionValid(submission *ent.MailSubmission, deliveryID CalendarMailDeliveryID) bool {
	if submission == nil || !validCalendarMailDeliveryID(deliveryID) || submission.MailDraftID == nil || *submission.MailDraftID == uuid.Nil ||
		!mailboxEntSubmissionRowValid(submission, submission.MailAccountID) || MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusFailed ||
		submission.LastErrorCode == nil || MailboxSubmissionErrorCode(*submission.LastErrorCode) != MailboxSubmissionErrorCodeRejected ||
		submission.ResultMailMessageID != nil || submission.ProviderSubmissionID != nil || submission.ProviderMessageID != nil ||
		submission.CommandLeaseToken != nil || submission.CommandLeaseExpiresAt != nil {
		return false
	}
	want := sha256.Sum256([]byte(calendarMailDeliverySubmissionKey(deliveryID)))
	return bytes.Equal(submission.IdempotencyKeyHash, want[:])
}

func mailboxEntDevDemoCalendarInviteFixtureDraft(ctx context.Context, transaction *ent.Tx, deliveryID CalendarMailDeliveryID, submission *ent.MailSubmission) (*ent.MailDraft, MailboxDraftRecord, error) {
	if ctx == nil || transaction == nil || submission == nil || submission.MailAccountID == uuid.Nil || submission.MailDraftID == nil || *submission.MailDraftID == uuid.Nil || !validCalendarMailDeliveryID(deliveryID) {
		return nil, MailboxDraftRecord{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	draftKeyHash := sha256.Sum256([]byte(calendarMailDeliveryDraftKey(deliveryID)))
	draft, err := transaction.MailDraft.Query().
		Where(
			maildraft.IDEQ(*submission.MailDraftID),
			maildraft.MailAccountIDEQ(submission.MailAccountID),
			maildraft.DraftClientIDEQ(calendarMailDeliveryDraftClientID(deliveryID)),
			maildraft.CreateIdempotencyKeyHashEQ(draftKeyHash[:]),
			maildraft.StatusEQ(maildraft.Status("active")),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || draft == nil || draft.ID != *submission.MailDraftID || draft.MailAccountID != submission.MailAccountID ||
		draft.DraftClientID != calendarMailDeliveryDraftClientID(deliveryID) || !bytes.Equal(draft.CreateIdempotencyKeyHash, draftKeyHash[:]) ||
		draft.EngineDraftKey != submission.EngineDraftKey || draft.EngineMessageID == nil || draft.EngineVersion == nil ||
		*draft.EngineMessageID != submission.EngineMessageID || *draft.EngineVersion != submission.EngineExpectedVersion {
		return nil, MailboxDraftRecord{}, mailboxEntDevDemoCalendarInviteRecoveryRowError(err)
	}
	attachments, kind := mailboxEntDraftReadAttachmentState(ctx, transaction, draft, MailboxDraftRepositoryErrorStateConflict)
	if kind != "" || len(attachments.ConfirmedAttachmentIDs) != 0 || len(attachments.PendingAttachmentIDs) != 0 {
		return nil, MailboxDraftRecord{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	record, ok := mailboxEntDraftCalendarCreateRecord(draft, attachments)
	if !ok {
		return nil, MailboxDraftRecord{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return draft, record, nil
}

func mailboxEntDevDemoCalendarInviteDetachSubmission(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery, request DevDemoCalendarInviteRecoveryRequest, submissionID uuid.UUID, databaseNow time.Time) error {
	updated, err := transaction.CalendarInviteDelivery.Update().
		Where(
			calendarinvitedelivery.IDEQ(delivery.ID),
			calendarinvitedelivery.OrganizationIDEQ(request.OrganizationID),
			calendarinvitedelivery.CalendarEventIDEQ(request.EventID),
			calendarinvitedelivery.CalendarAttendeeIDEQ(request.AttendeeID),
			calendarinvitedelivery.KindEQ(calendarinvitedelivery.Kind(request.Kind)),
			calendarinvitedelivery.EventSequenceEQ(request.EventSequence),
			calendarinvitedelivery.RecurrenceIDIsNil(),
			calendarinvitedelivery.StateEQ(calendarinvitedelivery.State(CalendarMailDeliveryStateFailedPermanent)),
			calendarinvitedelivery.MailSubmissionIDEQ(submissionID),
			calendarinvitedelivery.AttemptsEQ(delivery.Attempts),
			calendarinvitedelivery.ClaimGenerationEQ(delivery.ClaimGeneration),
			calendarinvitedelivery.LeaseTokenIsNil(), calendarinvitedelivery.LeaseUntilIsNil(),
			calendarinvitedelivery.NextAttemptAtIsNil(),
			calendarinvitedelivery.LastErrorCodeEQ(calendarinvitedelivery.LastErrorCode("submission_rejected")),
			calendarinvitedelivery.TerminalAtEQ(*delivery.TerminalAt),
		).
		ClearMailSubmissionID().
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if updated != 1 {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return nil
}

func mailboxEntDevDemoCalendarInviteDeleteSubmission(ctx context.Context, transaction *ent.Tx, submission *ent.MailSubmission, deliveryID CalendarMailDeliveryID) error {
	want := sha256.Sum256([]byte(calendarMailDeliverySubmissionKey(deliveryID)))
	deleted, err := transaction.MailSubmission.Delete().
		Where(
			mailsubmission.IDEQ(submission.ID),
			mailsubmission.MailAccountIDEQ(submission.MailAccountID),
			mailsubmission.MailDraftIDEQ(*submission.MailDraftID),
			mailsubmission.IdempotencyKeyHashEQ(want[:]),
			mailsubmission.StatusEQ(mailsubmission.Status("failed")),
			mailsubmission.LastErrorCodeEQ(mailsubmission.LastErrorCode(MailboxSubmissionErrorCodeRejected)),
			mailsubmission.CommandLeaseTokenIsNil(), mailsubmission.CommandLeaseExpiresAtIsNil(),
			mailsubmission.ProviderSubmissionIDIsNil(), mailsubmission.ProviderMessageIDIsNil(),
			mailsubmission.ResultMailMessageIDIsNil(), mailsubmission.NextAttemptAtIsNil(),
		).
		Exec(ctx)
	if err != nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if deleted != 1 {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return nil
}

func mailboxEntDevDemoCalendarInviteDeleteDraft(ctx context.Context, transaction *ent.Tx, draft *ent.MailDraft, deliveryID CalendarMailDeliveryID) error {
	want := sha256.Sum256([]byte(calendarMailDeliveryDraftKey(deliveryID)))
	deleted, err := transaction.MailDraft.Delete().
		Where(
			maildraft.IDEQ(draft.ID),
			maildraft.MailAccountIDEQ(draft.MailAccountID),
			maildraft.DraftClientIDEQ(calendarMailDeliveryDraftClientID(deliveryID)),
			maildraft.CreateIdempotencyKeyHashEQ(want[:]),
			maildraft.StatusEQ(maildraft.Status("active")),
			maildraft.EngineDraftKeyEQ(draft.EngineDraftKey),
		).
		Exec(ctx)
	if err != nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if deleted != 1 {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return nil
}

func mailboxEntDevDemoCalendarInviteRequeue(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery, request DevDemoCalendarInviteRecoveryRequest, databaseNow time.Time) error {
	updated, err := transaction.CalendarInviteDelivery.Update().
		Where(
			calendarinvitedelivery.IDEQ(delivery.ID),
			calendarinvitedelivery.OrganizationIDEQ(request.OrganizationID),
			calendarinvitedelivery.CalendarEventIDEQ(request.EventID),
			calendarinvitedelivery.CalendarAttendeeIDEQ(request.AttendeeID),
			calendarinvitedelivery.KindEQ(calendarinvitedelivery.Kind(request.Kind)),
			calendarinvitedelivery.EventSequenceEQ(request.EventSequence),
			calendarinvitedelivery.RecurrenceIDIsNil(),
			calendarinvitedelivery.StateEQ(calendarinvitedelivery.State(CalendarMailDeliveryStateFailedPermanent)),
			calendarinvitedelivery.MailSubmissionIDIsNil(),
			calendarinvitedelivery.AttemptsEQ(delivery.Attempts),
			calendarinvitedelivery.ClaimGenerationEQ(delivery.ClaimGeneration),
			calendarinvitedelivery.LeaseTokenIsNil(), calendarinvitedelivery.LeaseUntilIsNil(),
			calendarinvitedelivery.NextAttemptAtIsNil(),
			calendarinvitedelivery.LastErrorCodeEQ(calendarinvitedelivery.LastErrorCode("submission_rejected")),
			calendarinvitedelivery.TerminalAtEQ(*delivery.TerminalAt),
		).
		SetState(calendarinvitedelivery.State(CalendarMailDeliveryStatePending)).
		SetAttempts(0).
		ClearLeaseToken().
		ClearLeaseUntil().
		ClearLastErrorCode().
		ClearTerminalAt().
		SetNextAttemptAt(databaseNow).
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if updated != 1 {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return nil
}

func mailboxEntDevDemoCalendarInviteRecoveryRowError(err error) error {
	if err == nil || ent.IsNotFound(err) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
}
