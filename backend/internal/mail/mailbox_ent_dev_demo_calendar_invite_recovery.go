//go:build !qa && !release

package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendarinboundmessagereceipt"
	"github.com/valerii/hub/backend/ent/calendarinboundobjectstate"
	"github.com/valerii/hub/backend/ent/calendarinvitedelivery"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailmessagefolder"
	"github.com/valerii/hub/backend/ent/mailmessagelabel"
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

// DevDemoCalendarInviteSentRecoveryRequest is the caller-proved immutable
// identity of one development Calendar invite eligible for restart recovery.
type DevDemoCalendarInviteSentRecoveryRequest struct {
	OrganizationID     uuid.UUID
	EventID            uuid.UUID
	AttendeeID         uuid.UUID
	OrganizerID        uuid.UUID
	EventSequence      int
	Kind               CalendarInviteDeliveryKind
	RecurrenceID       string
	RecipientUserID    uuid.UUID
	RecipientAccountID uuid.UUID
	RecipientEventID   uuid.UUID
	RecipientMessageID uuid.UUID
	ExpectedUID        string
}

// DevDemoCalendarInviteSentRecoveryOutcome is the explicit proof result.
type DevDemoCalendarInviteSentRecoveryOutcome string

const (
	DevDemoCalendarInviteSentRecoveryOutcomeNoop     DevDemoCalendarInviteSentRecoveryOutcome = "noop"
	DevDemoCalendarInviteSentRecoveryOutcomeRequeued DevDemoCalendarInviteSentRecoveryOutcome = "requeued"
)

// DevDemoCalendarInviteSentRecoveryResult exposes the recovery result.
type DevDemoCalendarInviteSentRecoveryResult struct {
	Outcome DevDemoCalendarInviteSentRecoveryOutcome
}

type devDemoCalendarInviteSentRecoveryStage string

const (
	devDemoCalendarInviteSentRecoveryStageDatabaseClock     devDemoCalendarInviteSentRecoveryStage = "database-clock"
	devDemoCalendarInviteSentRecoveryStageGraphLock         devDemoCalendarInviteSentRecoveryStage = "graph-lock"
	devDemoCalendarInviteSentRecoveryStageMutation          devDemoCalendarInviteSentRecoveryStage = "mutation"
	devDemoCalendarInviteSentRecoveryStageThreadReaggregate devDemoCalendarInviteSentRecoveryStage = "thread-reaggregate"
	devDemoCalendarInviteSentRecoveryStageTransaction       devDemoCalendarInviteSentRecoveryStage = "transaction"
	devDemoCalendarInviteSentRecoveryStageResult            devDemoCalendarInviteSentRecoveryStage = "result"
)

type devDemoCalendarInviteSentRecoveryStageError struct {
	stage devDemoCalendarInviteSentRecoveryStage
	cause *CalendarMailDeliveryRepositoryError
}

func (recoveryError *devDemoCalendarInviteSentRecoveryStageError) Error() string {
	return string(recoveryError.stage) + ": " + string(recoveryError.cause.Kind())
}

func (recoveryError *devDemoCalendarInviteSentRecoveryStageError) Unwrap() error {
	return recoveryError.cause
}

func newDevDemoCalendarInviteSentRecoveryStageError(stage devDemoCalendarInviteSentRecoveryStage, err error) *devDemoCalendarInviteSentRecoveryStageError {
	sanitizedError := mailboxEntCalendarDeliveryRepositoryError(err)
	var repositoryError *CalendarMailDeliveryRepositoryError
	if errors.As(sanitizedError, &repositoryError) && repositoryError != nil {
		return &devDemoCalendarInviteSentRecoveryStageError{stage: stage, cause: repositoryError}
	}
	return &devDemoCalendarInviteSentRecoveryStageError{
		stage: stage,
		cause: newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable),
	}
}

// mailboxEntDevDemoCalendarInviteSentRecoveryGraph holds only identities
// derived from locked durable rows; no child identifier is caller-controlled.
type mailboxEntDevDemoCalendarInviteSentRecoveryGraph struct {
	deliveryID         CalendarMailDeliveryID
	event              *ent.CalendarEvent
	attendee           *ent.CalendarAttendee
	delivery           *ent.CalendarInviteDelivery
	submission         *ent.MailSubmission
	draft              *ent.MailDraft
	providerMessage    *ent.MailMessage
	recipientAccount   *ent.MailAccount
	inboundMessage     *ent.MailMessage
	inboundReceipt     *ent.CalendarInboundMessageReceipt
	inboundObjectState *ent.CalendarInboundObjectState
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

// RecoverDevDemoCalendarInviteSentDelivery requeues one fully proved sent
// development fixture after a local fake-engine restart invalidates only its
// ephemeral provider copy. It never accepts caller-controlled child rows.
func (repository *MailboxEntRepository) RecoverDevDemoCalendarInviteSentDelivery(ctx context.Context, request DevDemoCalendarInviteSentRecoveryRequest) (DevDemoCalendarInviteSentRecoveryResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validDevDemoCalendarInviteSentRecoveryRequest(request) {
		return DevDemoCalendarInviteSentRecoveryResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return DevDemoCalendarInviteSentRecoveryResult{}, err
	}

	var result DevDemoCalendarInviteSentRecoveryResult
	stage := devDemoCalendarInviteSentRecoveryStageTransaction
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		stage = devDemoCalendarInviteSentRecoveryStageDatabaseClock
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil || !mailboxEntValidTime(databaseNow) {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}

		stage = devDemoCalendarInviteSentRecoveryStageGraphLock
		graph, outcome, err := mailboxEntLockDevDemoCalendarInviteSentRecoveryGraph(ctx, transaction, request)
		if err != nil {
			return err
		}
		switch outcome {
		case DevDemoCalendarInviteSentRecoveryOutcomeNoop:
			result = DevDemoCalendarInviteSentRecoveryResult{Outcome: DevDemoCalendarInviteSentRecoveryOutcomeNoop}
			stage = devDemoCalendarInviteSentRecoveryStageTransaction
			return nil
		case DevDemoCalendarInviteSentRecoveryOutcomeRequeued:
			stage = devDemoCalendarInviteSentRecoveryStageMutation
			if err := mailboxEntRequeueDevDemoCalendarInviteSentRecoveryGraph(ctx, transaction, graph, request, databaseNow); err != nil {
				return err
			}
			stage = devDemoCalendarInviteSentRecoveryStageThreadReaggregate
			if err := mailboxEntReaggregateDevDemoCalendarInviteSentRecoveryThreads(ctx, transaction, graph, databaseNow); err != nil {
				return err
			}
			result = DevDemoCalendarInviteSentRecoveryResult{Outcome: DevDemoCalendarInviteSentRecoveryOutcomeRequeued}
			stage = devDemoCalendarInviteSentRecoveryStageTransaction
			return nil
		default:
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
		}
	})
	if contextErr := ctx.Err(); contextErr != nil {
		return DevDemoCalendarInviteSentRecoveryResult{}, contextErr
	}
	if err != nil {
		return DevDemoCalendarInviteSentRecoveryResult{}, newDevDemoCalendarInviteSentRecoveryStageError(stage, err)
	}
	stage = devDemoCalendarInviteSentRecoveryStageResult
	if !validDevDemoCalendarInviteSentRecoveryResult(result) {
		return DevDemoCalendarInviteSentRecoveryResult{}, newDevDemoCalendarInviteSentRecoveryStageError(
			stage,
			newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable),
		)
	}
	return result, nil
}

func validDevDemoCalendarInviteRecoveryRequest(request DevDemoCalendarInviteRecoveryRequest) bool {
	return request.OrganizationID != uuid.Nil && request.EventID != uuid.Nil && request.AttendeeID != uuid.Nil &&
		request.OrganizerID != uuid.Nil && request.EventSequence >= 0 && request.Kind == CalendarInviteDeliveryKindCreated &&
		request.RecurrenceID == ""
}

func validDevDemoCalendarInviteSentRecoveryResult(result DevDemoCalendarInviteSentRecoveryResult) bool {
	switch result.Outcome {
	case DevDemoCalendarInviteSentRecoveryOutcomeNoop, DevDemoCalendarInviteSentRecoveryOutcomeRequeued:
		return true
	default:
		return false
	}
}

func validDevDemoCalendarInviteSentRecoveryRequest(request DevDemoCalendarInviteSentRecoveryRequest) bool {
	return request.OrganizationID != uuid.Nil && request.EventID != uuid.Nil && request.AttendeeID != uuid.Nil &&
		request.OrganizerID != uuid.Nil && request.EventSequence >= 0 && request.Kind == CalendarInviteDeliveryKindCreated &&
		request.RecurrenceID == "" && request.RecipientUserID != uuid.Nil && request.RecipientAccountID != uuid.Nil &&
		request.RecipientEventID != uuid.Nil && request.RecipientMessageID != uuid.Nil &&
		request.ExpectedUID == "hub-"+request.EventID.String()
}

// mailboxEntLockDevDemoCalendarInviteSentRecoveryGraph proves the one exact
// sent demo invite graph. It has no mutation authority. Noop is returned only
// for the exact pending and unbound delivery shape.
func mailboxEntLockDevDemoCalendarInviteSentRecoveryGraph(ctx context.Context, transaction *ent.Tx, request DevDemoCalendarInviteSentRecoveryRequest) (mailboxEntDevDemoCalendarInviteSentRecoveryGraph, DevDemoCalendarInviteSentRecoveryOutcome, error) {
	graph := mailboxEntDevDemoCalendarInviteSentRecoveryGraph{}
	if ctx == nil || transaction == nil || !validDevDemoCalendarInviteSentRecoveryRequest(request) {
		return graph, "", newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return graph, "", err
	}

	event, err := transaction.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(request.EventID),
			calendarevent.OrganizationIDEQ(request.OrganizationID),
			calendarevent.OrganizerIDEQ(request.OrganizerID),
			calendarevent.SequenceEQ(request.EventSequence),
			calendarevent.UIDEQ(request.ExpectedUID),
			calendarevent.DeletedAtIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || event == nil || event.ID != request.EventID || event.OrganizationID != request.OrganizationID ||
		event.OrganizerID == nil || *event.OrganizerID != request.OrganizerID || event.Sequence != request.EventSequence ||
		event.UID != request.ExpectedUID || event.DeletedAt != nil {
		return graph, "", mailboxEntDevDemoCalendarInviteRecoveryRowError(err)
	}

	attendee, err := transaction.CalendarAttendee.Query().
		Where(
			calendarattendee.IDEQ(request.AttendeeID),
			calendarattendee.CalendarEventIDEQ(event.ID),
			calendarattendee.UserIDEQ(request.RecipientUserID),
			calendarattendee.RecurrenceIDIsNil(),
			calendarattendee.RemovedAtIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || attendee == nil || attendee.ID != request.AttendeeID || attendee.CalendarEventID != event.ID ||
		attendee.UserID == nil || *attendee.UserID != request.RecipientUserID || attendee.RecurrenceID != nil || attendee.RemovedAt != nil {
		return graph, "", mailboxEntDevDemoCalendarInviteRecoveryRowError(err)
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
		Order(ent.Asc(calendarinvitedelivery.FieldID)).
		Limit(2).
		ForUpdate().
		All(ctx)
	if err != nil {
		return graph, "", newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if len(deliveries) != 1 || !mailboxEntDevDemoCalendarInviteSentImmutableValid(deliveries[0], request) {
		return graph, "", newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	delivery := deliveries[0]
	if mailboxEntDevDemoCalendarInviteSentPendingValid(delivery) {
		return graph, DevDemoCalendarInviteSentRecoveryOutcomeNoop, nil
	}
	if !mailboxEntDevDemoCalendarInviteSentDeliveryValid(delivery) || delivery.MailSubmissionID == nil {
		return graph, "", newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	deliveryID, err := NewCalendarMailDeliveryID(CalendarMailDeliveryKindInvite, delivery.ID)
	if err != nil {
		return graph, "", newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	submission, err := mailboxEntDevDemoCalendarInviteSentSubmission(ctx, transaction, deliveryID, *delivery.MailSubmissionID)
	if err != nil {
		return graph, "", err
	}
	if !mailboxEntDevDemoCalendarInviteSentSubmissionValid(submission, deliveryID) {
		return graph, "", newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	draft, err := mailboxEntDevDemoCalendarInviteSentDraft(ctx, transaction, deliveryID, submission)
	if err != nil {
		return graph, "", err
	}
	providerMessage, err := mailboxEntDevDemoCalendarInviteSentProviderMessage(ctx, transaction, submission)
	if err != nil {
		return graph, "", err
	}

	recipientAccount, err := transaction.MailAccount.Query().
		Where(
			mailaccount.IDEQ(request.RecipientAccountID),
			mailaccount.OrganizationIDEQ(request.OrganizationID),
			mailaccount.UserIDEQ(request.RecipientUserID),
			mailaccount.DesiredEnabledEQ(true),
			mailaccount.StatusEQ(mailaccount.Status("active")),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || !mailboxEntDevDemoCalendarInviteRecipientAccountValid(recipientAccount, request) {
		return graph, "", mailboxEntDevDemoCalendarInviteRecoveryRowError(err)
	}

	inboundMessage, err := transaction.MailMessage.Query().
		Where(
			mailmessage.IDEQ(request.RecipientMessageID),
			mailmessage.MailAccountIDEQ(recipientAccount.ID),
			mailmessage.DeletedAtIsNil(),
			mailmessage.IsDraftEQ(false),
			mailmessage.ProjectionOriginEQ(mailmessage.ProjectionOrigin(EngineProjectionOriginInboundOrUnknown)),
			mailmessage.CalendarMethodEQ(mailmessage.CalendarMethod(MailCalendarMethodRequest)),
			mailmessage.CalendarEventIDIsNil(),
			mailmessage.CalendarPartLocatorHashNotNil(),
			mailmessage.HasAttachmentsEQ(false),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || !mailboxEntDevDemoCalendarInviteInboundMessageValid(inboundMessage, recipientAccount.ID, request) {
		return graph, "", mailboxEntDevDemoCalendarInviteRecoveryRowError(err)
	}
	if err := mailboxEntDevDemoCalendarInviteMessageAttachmentsEmpty(ctx, transaction, inboundMessage.ID); err != nil {
		return graph, "", err
	}

	locatorHash := append([]byte(nil), (*inboundMessage.CalendarPartLocatorHash)...)
	receipts, err := transaction.CalendarInboundMessageReceipt.Query().
		Where(
			calendarinboundmessagereceipt.OrganizationIDEQ(request.OrganizationID),
			calendarinboundmessagereceipt.MailboxIDEQ(recipientAccount.ID),
			calendarinboundmessagereceipt.MailMessageIDEQ(inboundMessage.ID),
			calendarinboundmessagereceipt.CalendarEventIDEQ(request.RecipientEventID),
			calendarinboundmessagereceipt.PartLocatorHashEQ(locatorHash),
		).
		Order(ent.Asc(calendarinboundmessagereceipt.FieldID)).
		Limit(2).
		ForUpdate().
		All(ctx)
	if err != nil {
		return graph, "", newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if len(receipts) != 1 || !mailboxEntDevDemoCalendarInviteInboundReceiptValid(receipts[0], request, recipientAccount.ID, inboundMessage.ID, locatorHash) {
		return graph, "", newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	objectStates, err := transaction.CalendarInboundObjectState.Query().
		Where(
			calendarinboundobjectstate.OrganizationIDEQ(request.OrganizationID),
			calendarinboundobjectstate.CalendarOwnerIDEQ(request.RecipientUserID),
			calendarinboundobjectstate.UIDEQ(request.ExpectedUID),
			calendarinboundobjectstate.RecurrenceIDIsNil(),
			calendarinboundobjectstate.CalendarEventIDEQ(request.RecipientEventID),
		).
		Order(ent.Asc(calendarinboundobjectstate.FieldID)).
		Limit(2).
		ForUpdate().
		All(ctx)
	if err != nil {
		return graph, "", newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if len(objectStates) != 1 || !mailboxEntDevDemoCalendarInviteInboundObjectStateValid(objectStates[0], event, request) {
		return graph, "", newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	if err := mailboxEntDevDemoCalendarInviteSentDraftReferencesAbsent(ctx, transaction, providerMessage.ID, inboundMessage.ID); err != nil {
		return graph, "", err
	}

	graph.deliveryID = deliveryID
	graph.event = event
	graph.attendee = attendee
	graph.delivery = delivery
	graph.submission = submission
	graph.draft = draft
	graph.providerMessage = providerMessage
	graph.recipientAccount = recipientAccount
	graph.inboundMessage = inboundMessage
	graph.inboundReceipt = receipts[0]
	graph.inboundObjectState = objectStates[0]
	return graph, DevDemoCalendarInviteSentRecoveryOutcomeRequeued, nil
}

func mailboxEntDevDemoCalendarInviteSentImmutableValid(delivery *ent.CalendarInviteDelivery, request DevDemoCalendarInviteSentRecoveryRequest) bool {
	return delivery != nil && delivery.ID != uuid.Nil && delivery.SourceEnvelopeID != uuid.Nil &&
		delivery.OrganizationID == request.OrganizationID && delivery.CalendarEventID == request.EventID &&
		delivery.CalendarAttendeeID == request.AttendeeID && delivery.Kind == calendarinvitedelivery.Kind(request.Kind) &&
		delivery.EventSequence == request.EventSequence && delivery.RecurrenceID == nil &&
		delivery.Attempts >= 0 && delivery.Attempts <= calendarMailDeliveryMaximumAttempts &&
		delivery.ClaimGeneration >= 0 && delivery.ClaimGeneration <= mailboxEntCalendarDeliveryMaximumClaimGeneration
}

func mailboxEntDevDemoCalendarInviteSentPendingValid(delivery *ent.CalendarInviteDelivery) bool {
	return delivery != nil && CalendarMailDeliveryState(delivery.State) == CalendarMailDeliveryStatePending &&
		delivery.MailSubmissionID == nil && delivery.Attempts == 0 && delivery.LeaseToken == nil && delivery.LeaseUntil == nil &&
		delivery.NextAttemptAt != nil && mailboxEntCalendarDeliveryTimestampValid(*delivery.NextAttemptAt) &&
		delivery.LastErrorCode == nil && delivery.TerminalAt == nil
}

func mailboxEntDevDemoCalendarInviteSentDeliveryValid(delivery *ent.CalendarInviteDelivery) bool {
	return delivery != nil && CalendarMailDeliveryState(delivery.State) == CalendarMailDeliveryStateSent &&
		delivery.MailSubmissionID != nil && *delivery.MailSubmissionID != uuid.Nil &&
		delivery.Attempts >= 1 && delivery.Attempts <= calendarMailDeliveryMaximumAttempts &&
		delivery.ClaimGeneration >= 1 && delivery.ClaimGeneration <= mailboxEntCalendarDeliveryMaximumClaimGeneration &&
		delivery.LeaseToken == nil && delivery.LeaseUntil == nil && delivery.NextAttemptAt == nil &&
		delivery.LastErrorCode == nil && delivery.TerminalAt != nil && mailboxEntCalendarDeliveryTimestampValid(*delivery.TerminalAt)
}

func mailboxEntDevDemoCalendarInviteSentSubmission(ctx context.Context, transaction *ent.Tx, deliveryID CalendarMailDeliveryID, submissionID uuid.UUID) (*ent.MailSubmission, error) {
	if ctx == nil || transaction == nil || !validCalendarMailDeliveryID(deliveryID) || submissionID == uuid.Nil {
		return nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	want := sha256.Sum256([]byte(calendarMailDeliverySubmissionKey(deliveryID)))
	submission, err := transaction.MailSubmission.Query().
		Where(
			mailsubmission.IDEQ(submissionID),
			mailsubmission.IdempotencyKeyHashEQ(want[:]),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || submission == nil || submission.ID != submissionID || !bytes.Equal(submission.IdempotencyKeyHash, want[:]) {
		return nil, mailboxEntDevDemoCalendarInviteRecoveryRowError(err)
	}
	return submission, nil
}

func mailboxEntDevDemoCalendarInviteSentSubmissionValid(submission *ent.MailSubmission, deliveryID CalendarMailDeliveryID) bool {
	if submission == nil || submission.ID == uuid.Nil || submission.MailAccountID == uuid.Nil || !validCalendarMailDeliveryID(deliveryID) ||
		!mailboxEntSubmissionRowValid(submission, submission.MailAccountID) || MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusSent ||
		submission.MailDraftID == nil || *submission.MailDraftID == uuid.Nil || submission.ResultMailMessageID == nil || *submission.ResultMailMessageID == uuid.Nil ||
		submission.CommandLeaseToken != nil || submission.CommandLeaseExpiresAt != nil || submission.DispatchAuthorizedAt == nil || submission.DispatchAuthorizationHash == nil ||
		len(*submission.DispatchAuthorizationHash) != sha256.Size || submission.ProviderSubmissionID == nil || submission.ProviderMessageID == nil ||
		*submission.ProviderSubmissionID == "" || *submission.ProviderMessageID == "" || submission.LastErrorCode != nil || submission.NextAttemptAt != nil {
		return false
	}
	want := sha256.Sum256([]byte(calendarMailDeliverySubmissionKey(deliveryID)))
	return bytes.Equal(submission.IdempotencyKeyHash, want[:])
}

func mailboxEntDevDemoCalendarInviteSentDraft(ctx context.Context, transaction *ent.Tx, deliveryID CalendarMailDeliveryID, submission *ent.MailSubmission) (*ent.MailDraft, error) {
	if ctx == nil || transaction == nil || !mailboxEntDevDemoCalendarInviteSentSubmissionValid(submission, deliveryID) || submission.MailDraftID == nil {
		return nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	want := sha256.Sum256([]byte(calendarMailDeliveryDraftKey(deliveryID)))
	draft, err := transaction.MailDraft.Query().
		Where(
			maildraft.IDEQ(*submission.MailDraftID),
			maildraft.MailAccountIDEQ(submission.MailAccountID),
			maildraft.DraftClientIDEQ(calendarMailDeliveryDraftClientID(deliveryID)),
			maildraft.CreateIdempotencyKeyHashEQ(want[:]),
			maildraft.StatusEQ(maildraft.Status("submitted")),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || draft == nil || !mailboxEntDraftReadRowValid(draft, submission.MailAccountID, *submission.MailDraftID) ||
		draft.DraftClientID != calendarMailDeliveryDraftClientID(deliveryID) || !bytes.Equal(draft.CreateIdempotencyKeyHash, want[:]) ||
		draft.Status != maildraft.Status("submitted") || draft.EngineDraftKey != submission.EngineDraftKey || draft.EngineMessageID == nil || draft.EngineVersion == nil ||
		*draft.EngineMessageID != submission.EngineMessageID || *draft.EngineVersion != submission.EngineExpectedVersion {
		return nil, mailboxEntDevDemoCalendarInviteRecoveryRowError(err)
	}
	attachments, kind := mailboxEntDraftReadAttachmentState(ctx, transaction, draft, MailboxDraftRepositoryErrorStateConflict)
	if kind != "" || len(attachments.ConfirmedAttachmentIDs) != 0 || len(attachments.PendingAttachmentIDs) != 0 {
		return nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return draft, nil
}

func mailboxEntDevDemoCalendarInviteSentProviderMessage(ctx context.Context, transaction *ent.Tx, submission *ent.MailSubmission) (*ent.MailMessage, error) {
	if ctx == nil || transaction == nil || submission == nil || submission.MailAccountID == uuid.Nil || submission.ResultMailMessageID == nil ||
		*submission.ResultMailMessageID == uuid.Nil || submission.ProviderMessageID == nil || *submission.ProviderMessageID == "" {
		return nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	message, err := transaction.MailMessage.Query().
		Where(
			mailmessage.IDEQ(*submission.ResultMailMessageID),
			mailmessage.MailAccountIDEQ(submission.MailAccountID),
			mailmessage.DeletedAtIsNil(),
			mailmessage.IsDraftEQ(false),
			mailmessage.ProjectionOriginEQ(mailmessage.ProjectionOrigin(EngineProjectionOriginConfirmedOwnerSubmission)),
			mailmessage.EngineMessageIDEQ(*submission.ProviderMessageID),
			mailmessage.CalendarMethodIsNil(),
			mailmessage.CalendarPartLocatorHashIsNil(),
			mailmessage.CalendarEventIDIsNil(),
			mailmessage.HasAttachmentsEQ(false),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || message == nil || !mailboxEntApplyMessageRowValid(message, submission.MailAccountID) || message.ID != *submission.ResultMailMessageID ||
		message.DeletedAt != nil || message.IsDraft || message.ProjectionOrigin != mailmessage.ProjectionOrigin(EngineProjectionOriginConfirmedOwnerSubmission) ||
		message.EngineMessageID != *submission.ProviderMessageID || message.CalendarMethod != nil ||
		!mailboxEntNullableBytesAbsent(message.CalendarPartLocatorHash) || message.CalendarEventID != nil || message.HasAttachments {
		return nil, mailboxEntDevDemoCalendarInviteRecoveryRowError(err)
	}
	if err := mailboxEntDevDemoCalendarInviteMessageAttachmentsEmpty(ctx, transaction, message.ID); err != nil {
		return nil, err
	}
	return message, nil
}

func mailboxEntDevDemoCalendarInviteRecipientAccountValid(account *ent.MailAccount, request DevDemoCalendarInviteSentRecoveryRequest) bool {
	return account != nil && account.ID == request.RecipientAccountID && account.OrganizationID == request.OrganizationID &&
		account.UserID == request.RecipientUserID && account.DesiredEnabled && account.Status == mailaccount.Status("active")
}

func mailboxEntDevDemoCalendarInviteInboundMessageValid(message *ent.MailMessage, accountID uuid.UUID, request DevDemoCalendarInviteSentRecoveryRequest) bool {
	return message != nil && accountID != uuid.Nil && message.ID == request.RecipientMessageID && message.MailAccountID == accountID &&
		mailboxEntApplyMessageRowValid(message, accountID) && message.DeletedAt == nil && !message.IsDraft &&
		message.ProjectionOrigin == mailmessage.ProjectionOrigin(EngineProjectionOriginInboundOrUnknown) && message.CalendarMethod != nil &&
		*message.CalendarMethod == mailmessage.CalendarMethod(MailCalendarMethodRequest) && message.CalendarEventID == nil &&
		message.CalendarPartLocatorHash != nil &&
		len(*message.CalendarPartLocatorHash) == sha256.Size && !message.HasAttachments
}

func mailboxEntDevDemoCalendarInviteMessageAttachmentsEmpty(ctx context.Context, transaction *ent.Tx, messageID uuid.UUID) error {
	if ctx == nil || transaction == nil || messageID == uuid.Nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	attachments, err := transaction.MailAttachment.Query().
		Where(mailattachment.MailMessageIDEQ(messageID)).
		Order(ent.Asc(mailattachment.FieldID)).
		Limit(1).
		ForUpdate().
		All(ctx)
	if err != nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if len(attachments) != 0 {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return nil
}

func mailboxEntDevDemoCalendarInviteInboundReceiptValid(receipt *ent.CalendarInboundMessageReceipt, request DevDemoCalendarInviteSentRecoveryRequest, accountID, messageID uuid.UUID, locatorHash []byte) bool {
	if receipt == nil || receipt.ID == uuid.Nil || accountID == uuid.Nil || messageID == uuid.Nil || len(locatorHash) != sha256.Size ||
		receipt.OrganizationID != request.OrganizationID || receipt.MailboxID != accountID || receipt.MailMessageID != messageID ||
		!bytes.Equal(receipt.PartLocatorHash, locatorHash) || receipt.EventEnvelopeID == uuid.Nil || receipt.CalendarEventID == nil ||
		*receipt.CalendarEventID != request.RecipientEventID || receipt.Attempts < 1 || receipt.Attempts > 8 || receipt.LeaseGeneration < 1 ||
		receipt.LeaseToken != nil || receipt.LeaseUntil != nil || receipt.NextAttemptAt != nil || receipt.ErrorCode != nil ||
		receipt.DeferredObjectStateID != nil || receipt.DeferredExpiresAt != nil || !mailboxEntValidTime(receipt.CreatedAt) || !mailboxEntValidTime(receipt.UpdatedAt) {
		return false
	}
	return receipt.State == calendarinboundmessagereceipt.State("applied") || receipt.State == calendarinboundmessagereceipt.State("duplicate")
}

func mailboxEntDevDemoCalendarInviteInboundObjectStateValid(state *ent.CalendarInboundObjectState, event *ent.CalendarEvent, request DevDemoCalendarInviteSentRecoveryRequest) bool {
	return state != nil && event != nil && event.ID == request.EventID && event.OrganizationID == request.OrganizationID &&
		event.UID == request.ExpectedUID && event.OrganizerEmail != "" && state.ID != uuid.Nil && state.OrganizationID == request.OrganizationID &&
		state.CalendarOwnerID == request.RecipientUserID && state.UID == request.ExpectedUID && state.RecurrenceID == nil &&
		state.State == calendarinboundobjectstate.State("applied") && state.Method == calendarinboundobjectstate.Method("request") &&
		state.CalendarEventID != nil && *state.CalendarEventID == request.RecipientEventID && state.ExpiresAt == nil &&
		state.Sequence == request.EventSequence && state.OrganizerEmail == event.OrganizerEmail &&
		state.PayloadHashVersion >= 1 && len(state.PayloadHash) == sha256.Size &&
		!state.CreatedAt.IsZero() && !state.UpdatedAt.IsZero() && !state.UpdatedAt.Before(state.CreatedAt)
}

// mailboxEntDevDemoCalendarInviteSentDraftReferencesAbsent locks every draft
// reference that would make the later hard delete RESTRICT-fail. Recovery must
// never unlink another draft just to recycle this one exact demo fixture.
func mailboxEntDevDemoCalendarInviteSentDraftReferencesAbsent(ctx context.Context, transaction *ent.Tx, providerMessageID, inboundMessageID uuid.UUID) error {
	if ctx == nil || transaction == nil || providerMessageID == uuid.Nil || inboundMessageID == uuid.Nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	rows, err := transaction.MailDraft.Query().
		Where(
			maildraft.Or(
				maildraft.ReplyToMessageIDIn(providerMessageID, inboundMessageID),
				maildraft.ForwardMessageIDIn(providerMessageID, inboundMessageID),
				maildraft.PendingReplyToMessageIDIn(providerMessageID, inboundMessageID),
				maildraft.PendingForwardMessageIDIn(providerMessageID, inboundMessageID),
			),
		).
		Order(ent.Asc(maildraft.FieldID)).
		Limit(1).
		ForUpdate().
		All(ctx)
	if err != nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if len(rows) != 0 {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return nil
}

// mailboxEntDevDemoCalendarInviteSentDeleteMessageLinks removes only the
// locked join rows owned by the two fixture messages. It never mutates the
// folders or labels themselves.
func mailboxEntDevDemoCalendarInviteSentDeleteMessageLinks(ctx context.Context, transaction *ent.Tx, providerMessage, inboundMessage *ent.MailMessage) error {
	if ctx == nil || transaction == nil || providerMessage == nil || inboundMessage == nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	for _, message := range []*ent.MailMessage{providerMessage, inboundMessage} {
		if message.ID == uuid.Nil || message.MailAccountID == uuid.Nil {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
		}

		folderLinks, err := transaction.MailMessageFolder.Query().
			Where(mailmessagefolder.MailMessageIDEQ(message.ID)).
			Order(ent.Asc(mailmessagefolder.FieldID)).
			Limit(MaxEngineFolderItems + 1).
			ForUpdate().
			All(ctx)
		if err != nil {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if len(folderLinks) > MaxEngineFolderItems {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
		}
		folderLinkIDs := make([]uuid.UUID, 0, len(folderLinks))
		for _, link := range folderLinks {
			if link == nil || link.ID == uuid.Nil || link.MailMessageID != message.ID || link.MailFolderID == uuid.Nil {
				return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
			}
			folderLinkIDs = append(folderLinkIDs, link.ID)
		}
		if len(folderLinkIDs) != 0 {
			deleted, err := transaction.MailMessageFolder.Delete().
				Where(
					mailmessagefolder.IDIn(folderLinkIDs...),
					mailmessagefolder.MailMessageIDEQ(message.ID),
				).
				Exec(ctx)
			if err != nil {
				return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
			}
			if deleted != len(folderLinkIDs) {
				return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
			}
		}

		labelLinks, err := transaction.MailMessageLabel.Query().
			Where(mailmessagelabel.MailMessageIDEQ(message.ID)).
			Order(ent.Asc(mailmessagelabel.FieldID)).
			Limit(MaxEngineKeywordItems + 1).
			ForUpdate().
			All(ctx)
		if err != nil {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if len(labelLinks) > MaxEngineKeywordItems {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
		}
		labelLinkIDs := make([]uuid.UUID, 0, len(labelLinks))
		for _, link := range labelLinks {
			if link == nil || link.ID == uuid.Nil || link.MailMessageID != message.ID || link.MailLabelID == uuid.Nil {
				return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
			}
			labelLinkIDs = append(labelLinkIDs, link.ID)
		}
		if len(labelLinkIDs) != 0 {
			deleted, err := transaction.MailMessageLabel.Delete().
				Where(
					mailmessagelabel.IDIn(labelLinkIDs...),
					mailmessagelabel.MailMessageIDEQ(message.ID),
				).
				Exec(ctx)
			if err != nil {
				return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
			}
			if deleted != len(labelLinkIDs) {
				return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
			}
		}
	}
	return nil
}

// mailboxEntRequeueDevDemoCalendarInviteSentRecoveryGraph performs the one
// bounded mutation after the graph proof atom. It deliberately preserves the
// Calendar event and inbound object state.
func mailboxEntRequeueDevDemoCalendarInviteSentRecoveryGraph(ctx context.Context, transaction *ent.Tx, graph mailboxEntDevDemoCalendarInviteSentRecoveryGraph, request DevDemoCalendarInviteSentRecoveryRequest, databaseNow time.Time) error {
	if ctx == nil || transaction == nil || !mailboxEntValidTime(databaseNow) || !mailboxEntDevDemoCalendarInviteSentRecoveryGraphMutationValid(graph, request) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	delivery := graph.delivery
	submission := graph.submission
	draft := graph.draft
	inboundMessage := graph.inboundMessage
	providerMessage := graph.providerMessage
	receipt := graph.inboundReceipt
	locatorHash := append([]byte(nil), (*inboundMessage.CalendarPartLocatorHash)...)

	updated, err := transaction.CalendarInviteDelivery.Update().
		Where(
			calendarinvitedelivery.IDEQ(delivery.ID),
			calendarinvitedelivery.OrganizationIDEQ(request.OrganizationID),
			calendarinvitedelivery.CalendarEventIDEQ(request.EventID),
			calendarinvitedelivery.CalendarAttendeeIDEQ(request.AttendeeID),
			calendarinvitedelivery.SourceEnvelopeIDEQ(delivery.SourceEnvelopeID),
			calendarinvitedelivery.KindEQ(calendarinvitedelivery.Kind(request.Kind)),
			calendarinvitedelivery.EventSequenceEQ(request.EventSequence),
			calendarinvitedelivery.RecurrenceIDIsNil(),
			calendarinvitedelivery.StateEQ(calendarinvitedelivery.State(CalendarMailDeliveryStateSent)),
			calendarinvitedelivery.MailSubmissionIDEQ(submission.ID),
			calendarinvitedelivery.AttemptsEQ(delivery.Attempts),
			calendarinvitedelivery.ClaimGenerationEQ(delivery.ClaimGeneration),
			calendarinvitedelivery.LeaseTokenIsNil(),
			calendarinvitedelivery.LeaseUntilIsNil(),
			calendarinvitedelivery.NextAttemptAtIsNil(),
			calendarinvitedelivery.LastErrorCodeIsNil(),
			calendarinvitedelivery.TerminalAtEQ(*delivery.TerminalAt),
		).
		SetState(calendarinvitedelivery.State(CalendarMailDeliveryStatePending)).
		ClearMailSubmissionID().
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

	deleted, err := transaction.MailSubmission.Delete().
		Where(
			mailsubmission.IDEQ(submission.ID),
			mailsubmission.MailAccountIDEQ(submission.MailAccountID),
			mailsubmission.MailDraftIDEQ(*submission.MailDraftID),
			mailsubmission.IdempotencyKeyHashEQ(submission.IdempotencyKeyHash),
			mailsubmission.StatusEQ(mailsubmission.Status(MailboxSubmissionStatusSent)),
			mailsubmission.CommandLeaseTokenIsNil(),
			mailsubmission.CommandLeaseExpiresAtIsNil(),
			mailsubmission.ProviderSubmissionIDEQ(*submission.ProviderSubmissionID),
			mailsubmission.ProviderMessageIDEQ(*submission.ProviderMessageID),
			mailsubmission.ResultMailMessageIDEQ(*submission.ResultMailMessageID),
			mailsubmission.LastErrorCodeIsNil(),
			mailsubmission.NextAttemptAtIsNil(),
		).
		Exec(ctx)
	if err != nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if deleted != 1 {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	deleted, err = transaction.MailDraft.Delete().
		Where(
			maildraft.IDEQ(draft.ID),
			maildraft.MailAccountIDEQ(draft.MailAccountID),
			maildraft.DraftClientIDEQ(draft.DraftClientID),
			maildraft.CreateIdempotencyKeyHashEQ(draft.CreateIdempotencyKeyHash),
			maildraft.StatusEQ(maildraft.Status("submitted")),
			maildraft.EngineDraftKeyEQ(draft.EngineDraftKey),
			maildraft.EngineMessageIDEQ(*draft.EngineMessageID),
			maildraft.EngineVersionEQ(*draft.EngineVersion),
		).
		Exec(ctx)
	if err != nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if deleted != 1 {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	deleted, err = transaction.CalendarInboundMessageReceipt.Delete().
		Where(
			calendarinboundmessagereceipt.IDEQ(receipt.ID),
			calendarinboundmessagereceipt.OrganizationIDEQ(request.OrganizationID),
			calendarinboundmessagereceipt.MailboxIDEQ(graph.recipientAccount.ID),
			calendarinboundmessagereceipt.MailMessageIDEQ(inboundMessage.ID),
			calendarinboundmessagereceipt.PartLocatorHashEQ(locatorHash),
			calendarinboundmessagereceipt.EventEnvelopeIDEQ(receipt.EventEnvelopeID),
			calendarinboundmessagereceipt.StateEQ(receipt.State),
			calendarinboundmessagereceipt.CalendarEventIDEQ(request.RecipientEventID),
			calendarinboundmessagereceipt.LeaseTokenIsNil(),
			calendarinboundmessagereceipt.LeaseUntilIsNil(),
			calendarinboundmessagereceipt.NextAttemptAtIsNil(),
			calendarinboundmessagereceipt.ErrorCodeIsNil(),
			calendarinboundmessagereceipt.DeferredObjectStateIDIsNil(),
			calendarinboundmessagereceipt.DeferredExpiresAtIsNil(),
		).
		Exec(ctx)
	if err != nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if deleted != 1 {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	if err := mailboxEntDevDemoCalendarInviteSentDeleteMessageLinks(ctx, transaction, providerMessage, inboundMessage); err != nil {
		return err
	}

	deleted, err = transaction.MailMessage.Delete().
		Where(
			mailmessage.IDEQ(inboundMessage.ID),
			mailmessage.MailAccountIDEQ(inboundMessage.MailAccountID),
			mailmessage.MailThreadIDEQ(inboundMessage.MailThreadID),
			mailmessage.EngineMessageIDEQ(inboundMessage.EngineMessageID),
			mailmessage.DeletedAtIsNil(),
			mailmessage.IsDraftEQ(false),
			mailmessage.ProjectionOriginEQ(mailmessage.ProjectionOrigin(EngineProjectionOriginInboundOrUnknown)),
			mailmessage.CalendarMethodEQ(mailmessage.CalendarMethod(MailCalendarMethodRequest)),
			mailmessage.CalendarPartLocatorHashEQ(locatorHash),
			mailmessage.CalendarEventIDIsNil(),
			mailmessage.HasAttachmentsEQ(false),
		).
		Exec(ctx)
	if err != nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if deleted != 1 {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	deleted, err = transaction.MailMessage.Delete().
		Where(
			mailmessage.IDEQ(providerMessage.ID),
			mailmessage.MailAccountIDEQ(providerMessage.MailAccountID),
			mailmessage.MailThreadIDEQ(providerMessage.MailThreadID),
			mailmessage.EngineMessageIDEQ(providerMessage.EngineMessageID),
			mailmessage.DeletedAtIsNil(),
			mailmessage.IsDraftEQ(false),
			mailmessage.ProjectionOriginEQ(mailmessage.ProjectionOrigin(EngineProjectionOriginConfirmedOwnerSubmission)),
			mailmessage.CalendarMethodIsNil(),
			mailmessage.CalendarPartLocatorHashIsNil(),
			mailmessage.CalendarEventIDIsNil(),
			mailmessage.HasAttachmentsEQ(false),
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

// mailboxEntReaggregateDevDemoCalendarInviteSentRecoveryThreads repairs only
// thread projections affected by the two just-deleted fixture messages. It
// never deletes a thread: empty-thread tombstoning remains centralized in the
// canonical aggregate helper.
func mailboxEntReaggregateDevDemoCalendarInviteSentRecoveryThreads(ctx context.Context, transaction *ent.Tx, graph mailboxEntDevDemoCalendarInviteSentRecoveryGraph, databaseNow time.Time) error {
	if ctx == nil || transaction == nil || !mailboxEntValidTime(databaseNow) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	providerMessage := graph.providerMessage
	inboundMessage := graph.inboundMessage
	if providerMessage == nil || inboundMessage == nil || graph.recipientAccount == nil ||
		providerMessage.ID == uuid.Nil || inboundMessage.ID == uuid.Nil ||
		providerMessage.MailAccountID == uuid.Nil || inboundMessage.MailAccountID == uuid.Nil ||
		providerMessage.MailThreadID == uuid.Nil || inboundMessage.MailThreadID == uuid.Nil ||
		providerMessage.MailAccountID == inboundMessage.MailAccountID ||
		inboundMessage.MailAccountID != graph.recipientAccount.ID ||
		providerMessage.LastSeenSyncEpoch < 0 || providerMessage.LastSeenSyncEpoch > mailboxEntApplySyncEpochMaximum ||
		inboundMessage.LastSeenSyncEpoch < 0 || inboundMessage.LastSeenSyncEpoch > mailboxEntApplySyncEpochMaximum ||
		!mailboxEntApplyMessageRowValid(providerMessage, providerMessage.MailAccountID) ||
		!mailboxEntApplyMessageRowValid(inboundMessage, inboundMessage.MailAccountID) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	if err := mailboxEntApplySyncThreadAggregates(
		ctx,
		transaction,
		providerMessage.MailAccountID,
		providerMessage.LastSeenSyncEpoch,
		[]uuid.UUID{providerMessage.MailThreadID},
		databaseNow,
	); err != nil {
		return err
	}
	return mailboxEntApplySyncThreadAggregates(
		ctx,
		transaction,
		inboundMessage.MailAccountID,
		inboundMessage.LastSeenSyncEpoch,
		[]uuid.UUID{inboundMessage.MailThreadID},
		databaseNow,
	)
}

func mailboxEntDevDemoCalendarInviteSentRecoveryGraphMutationValid(graph mailboxEntDevDemoCalendarInviteSentRecoveryGraph, request DevDemoCalendarInviteSentRecoveryRequest) bool {
	if !validDevDemoCalendarInviteSentRecoveryRequest(request) || graph.deliveryID.Kind() != CalendarMailDeliveryKindInvite ||
		graph.deliveryID.UUID() == uuid.Nil || graph.event == nil || graph.attendee == nil || graph.delivery == nil ||
		graph.submission == nil || graph.draft == nil || graph.providerMessage == nil || graph.recipientAccount == nil ||
		graph.inboundMessage == nil || graph.inboundReceipt == nil || graph.inboundObjectState == nil ||
		graph.delivery.ID != graph.deliveryID.UUID() || graph.delivery.MailSubmissionID == nil ||
		*graph.delivery.MailSubmissionID != graph.submission.ID || graph.submission.MailDraftID == nil ||
		*graph.submission.MailDraftID != graph.draft.ID || graph.submission.ResultMailMessageID == nil ||
		*graph.submission.ResultMailMessageID != graph.providerMessage.ID {
		return false
	}
	if !mailboxEntDevDemoCalendarInviteSentImmutableValid(graph.delivery, request) ||
		!mailboxEntDevDemoCalendarInviteSentDeliveryValid(graph.delivery) ||
		!mailboxEntDevDemoCalendarInviteSentSubmissionValid(graph.submission, graph.deliveryID) ||
		!mailboxEntDevDemoCalendarInviteRecipientAccountValid(graph.recipientAccount, request) ||
		!mailboxEntDevDemoCalendarInviteInboundMessageValid(graph.inboundMessage, graph.recipientAccount.ID, request) {
		return false
	}
	if graph.draft.ID != *graph.submission.MailDraftID || graph.draft.MailAccountID != graph.submission.MailAccountID ||
		graph.draft.DraftClientID != calendarMailDeliveryDraftClientID(graph.deliveryID) ||
		graph.draft.Status != maildraft.Status("submitted") || graph.draft.EngineDraftKey != graph.submission.EngineDraftKey ||
		graph.draft.EngineMessageID == nil || graph.draft.EngineVersion == nil ||
		*graph.draft.EngineMessageID != graph.submission.EngineMessageID ||
		*graph.draft.EngineVersion != graph.submission.EngineExpectedVersion ||
		graph.providerMessage.ID != *graph.submission.ResultMailMessageID ||
		graph.providerMessage.MailAccountID != graph.submission.MailAccountID ||
		graph.providerMessage.EngineMessageID != *graph.submission.ProviderMessageID {
		return false
	}
	locatorHash := *graph.inboundMessage.CalendarPartLocatorHash
	return mailboxEntDevDemoCalendarInviteInboundReceiptValid(
		graph.inboundReceipt,
		request,
		graph.recipientAccount.ID,
		graph.inboundMessage.ID,
		locatorHash,
	) && mailboxEntDevDemoCalendarInviteInboundObjectStateValid(graph.inboundObjectState, graph.event, request)
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
