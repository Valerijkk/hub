package mail

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailsubmission"
)

// ReserveMailboxSubmission atomically replays or creates one queued send
// intent. It does not acquire an engine lease and makes no provider call.
func (repository *MailboxEntRepository) ReserveMailboxSubmission(ctx context.Context, request MailboxSubmissionCreateRequest) (MailboxSubmissionReservation, error) {
	if repository == nil || repository.client == nil || ctx == nil {
		return MailboxSubmissionReservation{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	if kind := mailboxEntSubmissionRequestKind(request); kind != "" {
		return MailboxSubmissionReservation{}, NewMailboxSubmissionRepositoryError(kind)
	}

	var result MailboxSubmissionReservation
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, kind := mailboxEntSubmissionLockAccount(ctx, transaction, request.AccountID)
		if kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}

		existing, found, kind := mailboxEntSubmissionLockIdempotencyReceipt(ctx, transaction, request)
		if kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}
		if found {
			result = MailboxSubmissionReservation{Kind: MailboxSubmissionReservationReplay, Submission: existing}
			return nil
		}

		draft, kind := mailboxEntSubmissionLockDraft(ctx, transaction, request.AccountID, request.DraftID)
		if kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}
		active, kind := mailboxEntSubmissionLockActiveByDraft(ctx, transaction, request.AccountID, request.DraftID)
		if kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}
		if len(active) != 0 {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorSubmissionInProgress)
		}
		if kind := mailboxEntSubmissionNewAccountKind(account); kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}
		if kind := mailboxEntSubmissionDraftNewKind(draft, request); kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}

		databaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxEntSubmissionRepositoryError(err)
		}
		submissionKey := uuid.New().String()
		if !syncValidOpaque(submissionKey, MaxEngineOpaqueIDBytes) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		created, err := transaction.MailSubmission.Create().
			SetMailAccountID(request.AccountID).
			SetMailDraftID(request.DraftID).
			SetIdempotencyKeyHash(append([]byte(nil), request.IdempotencyKeyHash[:]...)).
			SetRequestHash(append([]byte(nil), request.RequestHash[:]...)).
			SetEngineSubmissionKey(submissionKey).
			SetStatus(mailsubmission.Status("queued")).
			SetAttemptCount(0).
			Save(ctx)
		if err != nil {
			if ent.IsConstraintError(err) {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorIdempotencyConflict)
			}
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		updated, err := draft.Update().
			SetStatus(maildraft.Status("submitting")).
			SetUpdatedAt(databaseWallClock).
			Save(ctx)
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		if !mailboxEntSubmissionDraftTransitionValid(draft, updated, request, databaseWallClock) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		record, ok := mailboxEntSubmissionRecord(created, request.AccountID)
		if created.AttemptCount != 0 || !ok || record.Status != MailboxSubmissionStatusQueued || record.DraftID == nil || *record.DraftID != request.DraftID ||
			record.ResultMailMessageID != nil || record.LastErrorCode != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		result = MailboxSubmissionReservation{Kind: MailboxSubmissionReservationQueued, Submission: record}
		return nil
	})
	if err != nil {
		return MailboxSubmissionReservation{}, mailboxEntSubmissionRepositoryError(err)
	}
	return result, nil
}

func mailboxEntSubmissionRequestKind(request MailboxSubmissionCreateRequest) MailboxSubmissionRepositoryErrorKind {
	if request.AccountID == uuid.Nil || request.DraftID == uuid.Nil || request.ExpectedRevision < 1 ||
		!mailboxSubmissionHashPresent(request.IdempotencyKeyHash) || !mailboxSubmissionHashPresent(request.RequestHash) {
		return MailboxSubmissionRepositoryErrorInvalid
	}
	return ""
}

// mailboxEntSubmissionLockAccount is intentionally the first durable lock. It
// validates row shape but postpones active/deleting adjudication until after an
// exact idempotency replay has been found.
func mailboxEntSubmissionLockAccount(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID) (*ent.MailAccount, MailboxSubmissionRepositoryErrorKind) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil {
		return nil, MailboxSubmissionRepositoryErrorInvalid
	}
	account, err := transaction.MailAccount.Query().Where(mailaccount.IDEQ(accountID)).ForUpdate().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, MailboxSubmissionRepositoryErrorNotFound
		}
		return nil, MailboxSubmissionRepositoryErrorUnavailable
	}
	if account == nil || account.ID != accountID {
		return nil, MailboxSubmissionRepositoryErrorInvalid
	}
	switch account.Status {
	case mailaccount.StatusActive, mailaccount.StatusDeleting, mailaccount.StatusProvisioning, mailaccount.StatusSuspended, mailaccount.StatusError:
	default:
		return nil, MailboxSubmissionRepositoryErrorInvalid
	}
	return account, ""
}

func mailboxEntSubmissionNewAccountKind(account *ent.MailAccount) MailboxSubmissionRepositoryErrorKind {
	if account == nil || account.ID == uuid.Nil {
		return MailboxSubmissionRepositoryErrorInvalid
	}
	switch account.Status {
	case mailaccount.StatusDeleting:
		return MailboxSubmissionRepositoryErrorAccountDeleting
	case mailaccount.StatusActive:
	case mailaccount.StatusProvisioning, mailaccount.StatusSuspended, mailaccount.StatusError:
		return MailboxSubmissionRepositoryErrorUnavailable
	default:
		return MailboxSubmissionRepositoryErrorInvalid
	}
	if !account.DesiredEnabled || account.EngineAccountID == nil || !syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) ||
		account.SyncEpoch < 0 || account.MutationEpoch < 1 || account.SyncLeaseEpoch < 1 || !mailboxEntValidTime(account.SyncScheduleUpdatedAt) {
		return MailboxSubmissionRepositoryErrorInvalid
	}
	if _, _, valid := mailboxEntSyncLeaseFence(account); !valid {
		return MailboxSubmissionRepositoryErrorInvalid
	}
	return ""
}

func mailboxEntSubmissionLockIdempotencyReceipt(ctx context.Context, transaction *ent.Tx, request MailboxSubmissionCreateRequest) (MailboxSubmissionRecord, bool, MailboxSubmissionRepositoryErrorKind) {
	if ctx == nil || transaction == nil || request.AccountID == uuid.Nil || !mailboxSubmissionHashPresent(request.IdempotencyKeyHash) || !mailboxSubmissionHashPresent(request.RequestHash) {
		return MailboxSubmissionRecord{}, false, MailboxSubmissionRepositoryErrorInvalid
	}
	row, err := transaction.MailSubmission.Query().
		Where(
			mailsubmission.MailAccountIDEQ(request.AccountID),
			mailsubmission.IdempotencyKeyHashEQ(request.IdempotencyKeyHash[:]...),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return MailboxSubmissionRecord{}, false, ""
		}
		return MailboxSubmissionRecord{}, false, MailboxSubmissionRepositoryErrorUnavailable
	}
	record, ok := mailboxEntSubmissionRecord(row, request.AccountID)
	if !ok || !bytes.Equal(row.IdempotencyKeyHash, request.IdempotencyKeyHash[:]) {
		return MailboxSubmissionRecord{}, false, MailboxSubmissionRepositoryErrorInvalid
	}
	if !bytes.Equal(row.RequestHash, request.RequestHash[:]) {
		return MailboxSubmissionRecord{}, false, MailboxSubmissionRepositoryErrorIdempotencyConflict
	}
	return record, true, ""
}

func mailboxEntSubmissionLockDraft(ctx context.Context, transaction *ent.Tx, accountID, draftID uuid.UUID) (*ent.MailDraft, MailboxSubmissionRepositoryErrorKind) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || draftID == uuid.Nil {
		return nil, MailboxSubmissionRepositoryErrorInvalid
	}
	draft, err := transaction.MailDraft.Query().
		Where(maildraft.IDEQ(draftID), maildraft.MailAccountIDEQ(accountID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, MailboxSubmissionRepositoryErrorNotFound
		}
		return nil, MailboxSubmissionRepositoryErrorUnavailable
	}
	if !mailboxEntDraftReadRowValid(draft, accountID, draftID) {
		return nil, MailboxSubmissionRepositoryErrorInvalid
	}
	return draft, ""
}

func mailboxEntSubmissionLockActiveByDraft(ctx context.Context, transaction *ent.Tx, accountID, draftID uuid.UUID) ([]MailboxSubmissionRecord, MailboxSubmissionRepositoryErrorKind) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || draftID == uuid.Nil {
		return nil, MailboxSubmissionRepositoryErrorInvalid
	}
	rows, err := transaction.MailSubmission.Query().
		Where(
			mailsubmission.MailAccountIDEQ(accountID),
			mailsubmission.MailDraftIDEQ(draftID),
			mailsubmission.StatusIn(
				mailsubmission.Status("queued"),
				mailsubmission.Status("submitting"),
				mailsubmission.Status("unknown"),
			),
		).
		Order(ent.Asc(mailsubmission.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, MailboxSubmissionRepositoryErrorUnavailable
	}
	records := make([]MailboxSubmissionRecord, 0, len(rows))
	for _, row := range rows {
		record, ok := mailboxEntSubmissionRecord(row, accountID)
		if !ok || record.DraftID == nil || *record.DraftID != draftID {
			return nil, MailboxSubmissionRepositoryErrorInvalid
		}
		switch record.Status {
		case MailboxSubmissionStatusQueued, MailboxSubmissionStatusSubmitting, MailboxSubmissionStatusUnknown:
		default:
			return nil, MailboxSubmissionRepositoryErrorInvalid
		}
		records = append(records, record)
	}
	return records, ""
}

func mailboxEntSubmissionDraftNewKind(draft *ent.MailDraft, request MailboxSubmissionCreateRequest) MailboxSubmissionRepositoryErrorKind {
	if draft == nil || !mailboxEntDraftReadRowValid(draft, request.AccountID, request.DraftID) {
		return MailboxSubmissionRepositoryErrorInvalid
	}
	if MailboxDraftStatus(draft.Status) != MailboxDraftStatusActive {
		if MailboxDraftStatus(draft.Status) == MailboxDraftStatusSubmitting {
			return MailboxSubmissionRepositoryErrorSubmissionInProgress
		}
		return MailboxSubmissionRepositoryErrorDraftNotSendable
	}
	if draft.Revision != request.ExpectedRevision {
		return MailboxSubmissionRepositoryErrorVersionConflict
	}
	if draft.EngineMessageID == nil || draft.EngineVersion == nil || !syncValidOpaque(draft.EngineDraftKey, MaxEngineOpaqueIDBytes) ||
		!syncValidOpaque(*draft.EngineMessageID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(*draft.EngineVersion, MaxEngineOpaqueIDBytes) {
		return MailboxSubmissionRepositoryErrorInvalid
	}
	return ""
}

func mailboxEntSubmissionDraftTransitionValid(before, after *ent.MailDraft, request MailboxSubmissionCreateRequest, updatedAt time.Time) bool {
	if before == nil || after == nil || !mailboxEntValidTime(updatedAt) || !mailboxEntDraftReadRowValid(before, request.AccountID, request.DraftID) ||
		!mailboxEntDraftReadRowValid(after, request.AccountID, request.DraftID) || before.ID != after.ID ||
		before.MailAccountID != after.MailAccountID || before.Revision != after.Revision || before.EngineDraftKey != after.EngineDraftKey ||
		before.EngineMessageID == nil || before.EngineVersion == nil || after.EngineMessageID == nil || after.EngineVersion == nil ||
		*before.EngineMessageID != *after.EngineMessageID || *before.EngineVersion != *after.EngineVersion ||
		MailboxDraftStatus(before.Status) != MailboxDraftStatusActive || MailboxDraftStatus(after.Status) != MailboxDraftStatusSubmitting ||
		!after.UpdatedAt.Equal(updatedAt) {
		return false
	}
	return before.DraftClientID == after.DraftClientID && before.FromIdentityID == after.FromIdentityID &&
		mailboxEntSubmissionUUIDPointersEqual(before.ReplyToMessageID, after.ReplyToMessageID) &&
		mailboxEntSubmissionUUIDPointersEqual(before.ForwardMessageID, after.ForwardMessageID) &&
		mailboxEntSubmissionUUIDPointersEqual(before.LastMutationID, after.LastMutationID) &&
		mailboxEntSubmissionUUIDPointersEqual(before.ConfirmedMutationID, after.ConfirmedMutationID)
}

func mailboxEntSubmissionUUIDPointersEqual(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func mailboxEntSubmissionRecord(row *ent.MailSubmission, accountID uuid.UUID) (MailboxSubmissionRecord, bool) {
	if !mailboxEntSubmissionRowValid(row, accountID) {
		return MailboxSubmissionRecord{}, false
	}
	var draftID *uuid.UUID
	if row.MailDraftID != nil {
		draftID = mailboxSubmissionUUIDPointerCopy(row.MailDraftID)
	}
	var resultMessageID *uuid.UUID
	if row.ResultMailMessageID != nil {
		resultMessageID = mailboxSubmissionUUIDPointerCopy(row.ResultMailMessageID)
	}
	var errorCode *MailboxSubmissionErrorCode
	if row.LastErrorCode != nil {
		value := MailboxSubmissionErrorCode(*row.LastErrorCode)
		errorCode = &value
	}
	return MailboxSubmissionRecord{
		ID:                  row.ID,
		AccountID:           row.MailAccountID,
		DraftID:             draftID,
		Status:              MailboxSubmissionStatus(row.Status),
		ResultMailMessageID: resultMessageID,
		LastErrorCode:       errorCode,
		CreatedAt:           row.CreatedAt.UTC(),
		UpdatedAt:           row.UpdatedAt.UTC(),
	}, true
}

func mailboxEntSubmissionRowValid(row *ent.MailSubmission, accountID uuid.UUID) bool {
	if row == nil || row.ID == uuid.Nil || row.MailAccountID != accountID || accountID == uuid.Nil ||
		len(row.IdempotencyKeyHash) != 32 || len(row.RequestHash) != 32 || !syncValidOpaque(row.EngineSubmissionKey, MaxEngineOpaqueIDBytes) ||
		row.AttemptCount < 0 || !mailboxEntValidTime(row.CreatedAt) || !mailboxEntValidTime(row.UpdatedAt) || row.UpdatedAt.Before(row.CreatedAt) ||
		!mailboxSubmissionUUIDPointerValid(row.MailDraftID) || !mailboxSubmissionUUIDPointerValid(row.ResultMailMessageID) {
		return false
	}
	enginePairPresent := row.EngineSubmissionID != nil || row.EngineResultMessageID != nil
	if enginePairPresent != (row.EngineSubmissionID != nil && row.EngineResultMessageID != nil) {
		return false
	}
	if enginePairPresent && (!syncValidOpaque(*row.EngineSubmissionID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(*row.EngineResultMessageID, MaxEngineOpaqueIDBytes)) {
		return false
	}
	var errorCode *MailboxSubmissionErrorCode
	if row.LastErrorCode != nil {
		value := MailboxSubmissionErrorCode(*row.LastErrorCode)
		if !validMailboxSubmissionErrorCode(value) {
			return false
		}
		errorCode = &value
	}
	switch MailboxSubmissionStatus(row.Status) {
	case MailboxSubmissionStatusQueued:
		return row.AttemptCount == 0 && row.MailDraftID != nil && !enginePairPresent && row.ResultMailMessageID == nil && errorCode == nil
	case MailboxSubmissionStatusSubmitting:
		return row.AttemptCount >= 0 && row.MailDraftID != nil && !enginePairPresent && row.ResultMailMessageID == nil && errorCode == nil
	case MailboxSubmissionStatusSent:
		return enginePairPresent && row.ResultMailMessageID != nil && errorCode == nil
	case MailboxSubmissionStatusUnknown:
		if row.MailDraftID == nil || row.ResultMailMessageID != nil || errorCode == nil {
			return false
		}
		switch *errorCode {
		case MailboxSubmissionErrorCodeOutcomeUnknown, MailboxSubmissionErrorCodeReconciliationRequired, MailboxSubmissionErrorCodeMailboxUnavailable:
		default:
			return false
		}
		return !enginePairPresent || *errorCode != MailboxSubmissionErrorCodeOutcomeUnknown
	case MailboxSubmissionStatusFailed:
		if enginePairPresent || row.ResultMailMessageID != nil || errorCode == nil {
			return false
		}
		switch *errorCode {
		case MailboxSubmissionErrorCodeRejected,
			MailboxSubmissionErrorCodePolicy,
			MailboxSubmissionErrorCodeAttachmentNotClean,
			MailboxSubmissionErrorCodeIdentityUnavailable,
			MailboxSubmissionErrorCodeRecipientLimit,
			MailboxSubmissionErrorCodeSizeLimit,
			MailboxSubmissionErrorCodeMailboxUnavailable:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func mailboxEntSubmissionRepositoryError(err error) error {
	var repositoryError *MailboxSubmissionRepositoryError
	if errors.As(err, &repositoryError) && repositoryError != nil {
		return NewMailboxSubmissionRepositoryError(repositoryError.Kind())
	}
	switch {
	case errors.Is(err, ErrMailboxServiceInvalid):
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	case errors.Is(err, ErrMailboxServiceNotFound):
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorNotFound)
	case errors.Is(err, ErrMailboxServiceDeleting):
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorAccountDeleting)
	default:
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
}
