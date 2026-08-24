package mail

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/maildraft"
)

const mailboxDraftRestartAuditAction = "restart"

// RestartMailboxDraftCreateUpload rotates an exact, due unknown create/upload mutation
// into fresh upload work. It performs no provider, object-store, or other external I/O.
func (repository *MailboxEntRepository) RestartMailboxDraftCreateUpload(ctx context.Context, request MailboxDraftRestartCreateUploadRequest) (MailboxDraftRestartCreateUploadResult, error) {
	if repository == nil || repository.client == nil || ctx == nil {
		return MailboxDraftRestartCreateUploadResult{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if kind := mailboxEntDraftRestartRequestKind(request); kind != "" {
		return MailboxDraftRestartCreateUploadResult{}, NewMailboxDraftRepositoryError(kind)
	}

	var result MailboxDraftRestartCreateUploadResult
	var postCommitErr error
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseWallClock, err := mailboxEntDraftCreateLockAccount(ctx, transaction, request.Fence)
		if err != nil {
			return err
		}
		if kind := mailboxEntDraftCreateAccountKind(account); kind != "" {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, databaseWallClock, kind, &postCommitErr)
		}
		if !mailboxEntDraftRecoveryReceiptAccountValid(account) {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorInvalid, &postCommitErr)
		}
		draft, kind := mailboxEntDraftReservationLockDraft(ctx, transaction, request.Fence.AccountID, request.DraftID)
		if kind != "" {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, databaseWallClock, kind, &postCommitErr)
		}
		attachments, kind := mailboxEntDraftReadAttachmentState(ctx, transaction, draft, MailboxDraftRepositoryErrorStateConflict)
		if kind != "" {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, databaseWallClock, kind, &postCommitErr)
		}
		record, ok := mailboxEntDraftReadRecord(draft, attachments)
		if !ok {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}
		receiptInput := mailboxDraftRecoveryReceiptInput{
			Operation:          mailboxDraftRecoveryReceiptOperationRestart,
			DraftID:            request.DraftID,
			IdempotencyKeyHash: request.IdempotencyKeyHash,
			RequestHash:        request.RequestHash,
			FingerprintVersion: request.FingerprintVersion,
		}
		receipt, kind := mailboxEntDraftRecoveryReceiptLock(ctx, transaction, account, receiptInput)
		if kind != "" {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, databaseWallClock, kind, &postCommitErr)
		}
		if receipt != nil {
			if !mailboxEntDraftRecoveryReceiptReplayValid(receipt, account, receiptInput, databaseWallClock) || !mailboxEntDraftRestartPostStateValid(record, attachments, request) {
				return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorIdempotencyConflict, &postCommitErr)
			}
			if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, databaseWallClock); err != nil {
				return err
			}
			result = MailboxDraftRestartCreateUploadResult{Replayed: true, Draft: mailboxEntDraftRecordCopy(record)}
			return nil
		}
		if !mailboxEntDraftRestartFreshStateValid(record, attachments, request, databaseWallClock) {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}

		freshDatabaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxEntDraftRepositoryError(err)
		}
		if !mailboxEntDraftCreateFenceMatches(account, request.Fence, freshDatabaseWallClock) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorFenceLost)
		}
		if !mailboxEntDraftRestartFreshStateValid(record, attachments, request, freshDatabaseWallClock) {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, freshDatabaseWallClock, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}

		nextAttempt := request.ExpectedAttempt + 1
		if err := draft.Update().
			SetStatus(maildraft.StatusSaving).
			SetLastMutationID(request.NextMutationID).
			SetPendingMutationKind(maildraft.PendingMutationKind(MailboxDraftMutationCreate)).
			SetPendingCreateAttachmentPhase(maildraft.PendingCreateAttachmentPhase(MailboxDraftCreateAttachmentPhaseUpload)).
			ClearPendingCreateUploadBlobID().
			ClearPendingCreateUploadSizeBytes().
			SetPendingCreateUploadAttempt(nextAttempt).
			ClearPendingCreateUploadRetryAt().
			ClearPendingCreateUploadExhaustedAt().
			SetUpdatedAt(freshDatabaseWallClock).
			Exec(ctx); err != nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		updated := mailboxEntDraftRecordCopy(record)
		phase := MailboxDraftCreateAttachmentPhaseUpload
		updated.Status = MailboxDraftStatusSaving
		updated.LastMutationID = mailboxEntDraftReadUUIDPointerCopy(&request.NextMutationID)
		updated.PendingMutationKind = mailboxEntDraftReadMutationKindPointerCopy(mailboxEntDraftRestartMutationKindPointer(MailboxDraftMutationCreate))
		updated.PendingCreateAttachmentPhase = mailboxEntDraftReadCreateAttachmentPhasePointerCopy(&phase)
		updated.PendingCreateUploadBlobID = nil
		updated.PendingCreateUploadSizeBytes = nil
		updated.PendingCreateUploadAttempt = nextAttempt
		updated.PendingCreateUploadRetryAt = nil
		updated.PendingCreateUploadExhaustedAt = nil
		updated.UpdatedAt = freshDatabaseWallClock.UTC()
		if !mailboxEntDraftRestartPostStateValid(updated, attachments, request) {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, freshDatabaseWallClock, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}
		work := mailboxEntDraftCreateWork(updated, MailboxDraftWorkCreate)
		if work == nil {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, freshDatabaseWallClock, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}
		if _, kind := mailboxEntDraftRecoveryReceiptComplete(ctx, transaction, account, receiptInput, freshDatabaseWallClock); kind != "" {
			return NewMailboxDraftRepositoryError(kind)
		}
		if err := mailboxEntDraftRestartAudit(ctx, transaction, account, updated, nextAttempt, freshDatabaseWallClock); err != nil {
			return err
		}
		fence := mailboxEntDraftCreateFenceCopy(request.Fence)
		reservation := &MailboxDraftReservation{Kind: MailboxDraftReservationDispatch, Draft: mailboxEntDraftRecordCopy(updated), Work: work, Fence: &fence}
		result = MailboxDraftRestartCreateUploadResult{Draft: mailboxEntDraftRecordCopy(updated), Reservation: reservation}
		return nil
	})
	if err != nil {
		return MailboxDraftRestartCreateUploadResult{}, mailboxEntDraftRepositoryError(err)
	}
	if postCommitErr != nil {
		return MailboxDraftRestartCreateUploadResult{}, postCommitErr
	}
	return result, nil
}

// RestartMailboxDraftCreateUploadAfterBlobNotFound rotates only a deterministic synchronous
// Email/import blobNotFound result back to upload. Unknown/import is deliberately rejected: an
// ambiguous provider result must reconcile instead of risking a duplicate import.
func (repository *MailboxEntRepository) RestartMailboxDraftCreateUploadAfterBlobNotFound(ctx context.Context, request MailboxDraftRestartCreateUploadAfterBlobNotFoundRequest) (MailboxDraftRestartCreateUploadAfterBlobNotFoundResult, error) {
	if repository == nil || repository.client == nil || ctx == nil {
		return MailboxDraftRestartCreateUploadAfterBlobNotFoundResult{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if kind := mailboxEntDraftRestartAfterBlobNotFoundRequestKind(request); kind != "" {
		return MailboxDraftRestartCreateUploadAfterBlobNotFoundResult{}, NewMailboxDraftRepositoryError(kind)
	}

	var result MailboxDraftRestartCreateUploadAfterBlobNotFoundResult
	var postCommitErr error
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseWallClock, err := mailboxEntDraftCreateLockAccount(ctx, transaction, request.Fence)
		if err != nil {
			return err
		}
		if kind := mailboxEntDraftCreateAccountKind(account); kind != "" {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, databaseWallClock, kind, &postCommitErr)
		}
		if account.ID != request.AccountID || account.UserID != request.ActorID {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}
		draft, kind := mailboxEntDraftReservationLockDraft(ctx, transaction, request.Fence.AccountID, request.DraftID)
		if kind != "" {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, databaseWallClock, kind, &postCommitErr)
		}
		attachments, kind := mailboxEntDraftReadAttachmentState(ctx, transaction, draft, MailboxDraftRepositoryErrorStateConflict)
		if kind != "" {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, databaseWallClock, kind, &postCommitErr)
		}
		record, ok := mailboxEntDraftReadRecord(draft, attachments)
		if !ok || !mailboxEntDraftRestartAfterBlobNotFoundFreshStateValid(record, attachments, request) {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}

		freshDatabaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxEntDraftRepositoryError(err)
		}
		if !mailboxEntDraftCreateFenceMatches(account, request.Fence, freshDatabaseWallClock) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorFenceLost)
		}
		if !mailboxEntDraftRestartAfterBlobNotFoundFreshStateValid(record, attachments, request) {
			return mailboxEntDraftRestartLogicalError(ctx, account, request.Fence, freshDatabaseWallClock, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}

		nextAttempt := record.PendingCreateUploadAttempt + 1
		if err := draft.Update().
			SetStatus(maildraft.StatusSaving).
			SetLastMutationID(request.NextMutationID).
			SetPendingMutationKind(maildraft.PendingMutationKind(MailboxDraftMutationCreate)).
			SetPendingCreateAttachmentPhase(maildraft.PendingCreateAttachmentPhase(MailboxDraftCreateAttachmentPhaseUpload)).
			ClearPendingCreateUploadBlobID().
			ClearPendingCreateUploadSizeBytes().
			SetPendingCreateUploadAttempt(nextAttempt).
			ClearPendingCreateUploadRetryAt().
			ClearPendingCreateUploadExhaustedAt().
			SetUpdatedAt(freshDatabaseWallClock).
			Exec(ctx); err != nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}

		updated := mailboxEntDraftRecordCopy(record)
		phase := MailboxDraftCreateAttachmentPhaseUpload
		updated.Status = MailboxDraftStatusSaving
		updated.LastMutationID = mailboxEntDraftReadUUIDPointerCopy(&request.NextMutationID)
		updated.PendingMutationKind = mailboxEntDraftReadMutationKindPointerCopy(mailboxEntDraftRestartMutationKindPointer(MailboxDraftMutationCreate))
		updated.PendingCreateAttachmentPhase = mailboxEntDraftReadCreateAttachmentPhasePointerCopy(&phase)
		updated.PendingCreateUploadBlobID = nil
		updated.PendingCreateUploadSizeBytes = nil
		updated.PendingCreateUploadAttempt = nextAttempt
		updated.PendingCreateUploadRetryAt = nil
		updated.PendingCreateUploadExhaustedAt = nil
		updated.UpdatedAt = freshDatabaseWallClock.UTC()
		if !mailboxEntDraftRestartAfterBlobNotFoundPostStateValid(updated, attachments, request, nextAttempt) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		work := mailboxEntDraftCreateWork(updated, MailboxDraftWorkCreate)
		if work == nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if err := mailboxEntDraftRestartAudit(ctx, transaction, account, updated, nextAttempt, freshDatabaseWallClock); err != nil {
			return err
		}
		fence := mailboxEntDraftCreateFenceCopy(request.Fence)
		reservation := MailboxDraftReservation{Kind: MailboxDraftReservationDispatch, Draft: mailboxEntDraftRecordCopy(updated), Work: work, Fence: &fence}
		result = MailboxDraftRestartCreateUploadAfterBlobNotFoundResult{Draft: mailboxEntDraftRecordCopy(updated), Reservation: reservation}
		return nil
	})
	if err != nil {
		return MailboxDraftRestartCreateUploadAfterBlobNotFoundResult{}, mailboxEntDraftRepositoryError(err)
	}
	if postCommitErr != nil {
		return MailboxDraftRestartCreateUploadAfterBlobNotFoundResult{}, postCommitErr
	}
	return result, nil
}

func mailboxEntDraftRestartAfterBlobNotFoundRequestKind(request MailboxDraftRestartCreateUploadAfterBlobNotFoundRequest) MailboxDraftRepositoryErrorKind {
	if !mailboxEntEngineCommandFenceValid(request.Fence) || request.ActorID == uuid.Nil || request.AccountID == uuid.Nil || request.Fence.AccountID != request.AccountID || request.DraftID == uuid.Nil || request.CurrentMutationID == uuid.Nil || request.NextMutationID == uuid.Nil || request.CurrentMutationID == request.NextMutationID || !validMailboxDraftAdvanceCreateToImportHandoff(request.ExpectedBlobID, request.ExpectedSizeBytes) {
		return MailboxDraftRepositoryErrorInvalid
	}
	return ""
}

func mailboxEntDraftRestartAfterBlobNotFoundFreshStateValid(record MailboxDraftRecord, attachments mailboxEntDraftAttachmentState, request MailboxDraftRestartCreateUploadAfterBlobNotFoundRequest) bool {
	return record.ID == request.DraftID && record.AccountID == request.AccountID && record.AccountID == request.Fence.AccountID && record.Status == MailboxDraftStatusSaving && record.Revision == 1 && record.EngineMessageID == nil && record.EngineVersion == nil && record.PendingSave == nil && record.PendingMutationKind != nil && *record.PendingMutationKind == MailboxDraftMutationCreate && record.PendingCreateAttachmentPhase != nil && *record.PendingCreateAttachmentPhase == MailboxDraftCreateAttachmentPhaseImport && record.PendingCreateUploadBlobID != nil && *record.PendingCreateUploadBlobID == request.ExpectedBlobID && record.PendingCreateUploadSizeBytes != nil && *record.PendingCreateUploadSizeBytes == request.ExpectedSizeBytes && record.LastMutationID != nil && *record.LastMutationID == request.CurrentMutationID && record.ConfirmedMutationID == nil && record.PendingCreateUploadAttempt >= 1 && record.PendingCreateUploadAttempt < 3 && record.PendingCreateUploadRetryAt == nil && record.PendingCreateUploadExhaustedAt == nil && len(record.Confirmed.AttachmentIDs) != 0 && len(record.Confirmed.AttachmentIDs) == len(attachments.ConfirmedAttachmentIDs) && len(attachments.PendingAttachmentIDs) == 0 && record.AbandonedAt == nil && record.AbandonedMutationID == nil && validMailboxDraftPendingCreateUploadHandoff(record.PendingCreateAttachmentPhase, record.PendingCreateUploadBlobID, record.PendingCreateUploadSizeBytes)
}

func mailboxEntDraftRestartAfterBlobNotFoundPostStateValid(record MailboxDraftRecord, attachments mailboxEntDraftAttachmentState, request MailboxDraftRestartCreateUploadAfterBlobNotFoundRequest, expectedAttempt int) bool {
	return record.ID == request.DraftID && record.AccountID == request.AccountID && record.AccountID == request.Fence.AccountID && record.Status == MailboxDraftStatusSaving && record.Revision == 1 && record.EngineMessageID == nil && record.EngineVersion == nil && record.PendingSave == nil && record.PendingMutationKind != nil && *record.PendingMutationKind == MailboxDraftMutationCreate && record.PendingCreateAttachmentPhase != nil && *record.PendingCreateAttachmentPhase == MailboxDraftCreateAttachmentPhaseUpload && record.PendingCreateUploadBlobID == nil && record.PendingCreateUploadSizeBytes == nil && record.LastMutationID != nil && *record.LastMutationID == request.NextMutationID && record.ConfirmedMutationID == nil && expectedAttempt >= 2 && expectedAttempt <= 3 && record.PendingCreateUploadAttempt == expectedAttempt && record.PendingCreateUploadRetryAt == nil && record.PendingCreateUploadExhaustedAt == nil && len(record.Confirmed.AttachmentIDs) != 0 && len(record.Confirmed.AttachmentIDs) == len(attachments.ConfirmedAttachmentIDs) && len(attachments.PendingAttachmentIDs) == 0 && record.AbandonedAt == nil && record.AbandonedMutationID == nil
}

func mailboxEntDraftRestartRequestKind(request MailboxDraftRestartCreateUploadRequest) MailboxDraftRepositoryErrorKind {
	if !mailboxEntEngineCommandFenceValid(request.Fence) || request.DraftID == uuid.Nil || request.ExpectedMutationID == uuid.Nil || request.NextMutationID == uuid.Nil || request.ExpectedMutationID == request.NextMutationID || request.ExpectedAttempt != 1 && request.ExpectedAttempt != 2 || request.FingerprintVersion != mailboxDraftRecoveryReceiptFingerprintVersion || !mailboxEntDraftCreateHashPresent(request.IdempotencyKeyHash) || !mailboxEntDraftCreateHashPresent(request.RequestHash) {
		return MailboxDraftRepositoryErrorInvalid
	}
	return ""
}

func mailboxEntDraftRestartFreshStateValid(record MailboxDraftRecord, attachments mailboxEntDraftAttachmentState, request MailboxDraftRestartCreateUploadRequest, databaseNow time.Time) bool {
	return record.ID == request.DraftID && record.AccountID == request.Fence.AccountID && record.Status == MailboxDraftStatusUnknown && record.Revision == 1 && record.EngineMessageID == nil && record.EngineVersion == nil && record.PendingSave == nil && record.PendingMutationKind != nil && *record.PendingMutationKind == MailboxDraftMutationCreate && record.PendingCreateAttachmentPhase != nil && *record.PendingCreateAttachmentPhase == MailboxDraftCreateAttachmentPhaseUpload && record.PendingCreateUploadBlobID == nil && record.PendingCreateUploadSizeBytes == nil && record.LastMutationID != nil && *record.LastMutationID == request.ExpectedMutationID && record.PendingCreateUploadAttempt == request.ExpectedAttempt && record.PendingCreateUploadRetryAt != nil && !databaseNow.Before(*record.PendingCreateUploadRetryAt) && record.PendingCreateUploadExhaustedAt == nil && len(record.Confirmed.AttachmentIDs) != 0 && len(attachments.ConfirmedAttachmentIDs) != 0 && len(attachments.PendingAttachmentIDs) == 0 && record.AbandonedAt == nil && record.AbandonedMutationID == nil
}

func mailboxEntDraftRestartPostStateValid(record MailboxDraftRecord, attachments mailboxEntDraftAttachmentState, request MailboxDraftRestartCreateUploadRequest) bool {
	return record.ID == request.DraftID && record.AccountID == request.Fence.AccountID && record.Status == MailboxDraftStatusSaving && record.Revision == 1 && record.EngineMessageID == nil && record.EngineVersion == nil && record.PendingSave == nil && record.PendingMutationKind != nil && *record.PendingMutationKind == MailboxDraftMutationCreate && record.PendingCreateAttachmentPhase != nil && *record.PendingCreateAttachmentPhase == MailboxDraftCreateAttachmentPhaseUpload && record.PendingCreateUploadBlobID == nil && record.PendingCreateUploadSizeBytes == nil && record.LastMutationID != nil && *record.LastMutationID == request.NextMutationID && record.PendingCreateUploadAttempt == request.ExpectedAttempt+1 && record.PendingCreateUploadRetryAt == nil && record.PendingCreateUploadExhaustedAt == nil && len(record.Confirmed.AttachmentIDs) != 0 && len(attachments.ConfirmedAttachmentIDs) != 0 && len(attachments.PendingAttachmentIDs) == 0 && record.AbandonedAt == nil && record.AbandonedMutationID == nil
}

func mailboxEntDraftRestartLogicalError(ctx context.Context, account *ent.MailAccount, fence MailboxEngineCommandFence, databaseWallClock time.Time, kind MailboxDraftRepositoryErrorKind, postCommitErr *error) error {
	if postCommitErr == nil || kind == MailboxDraftRepositoryErrorUnavailable {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	if err := mailboxEntDraftCreateRelease(ctx, account, fence, databaseWallClock); err != nil {
		return err
	}
	*postCommitErr = NewMailboxDraftRepositoryError(kind)
	return nil
}

func mailboxEntDraftRestartAudit(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, updated MailboxDraftRecord, attempt int, databaseWallClock time.Time) error {
	if ctx == nil || transaction == nil || !mailboxEntDraftRecoveryReceiptAccountValid(account) || updated.ID == uuid.Nil || attempt < 2 || attempt > 3 || !mailboxEntValidTime(databaseWallClock) {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if _, err := transaction.AuditLog.Create().
		SetAction(mailboxDraftRestartAuditAction).
		SetEntityType("mail_draft").
		SetEntityID(updated.ID.String()).
		SetUserID(account.UserID).
		SetDetails(map[string]interface{}{
			"organization_id": account.OrganizationID.String(),
			"mail_account_id": account.ID.String(),
			"operation":       string(mailboxDraftRecoveryReceiptOperationRestart),
			"status":          string(updated.Status),
			"attempt":         attempt,
		}).
		SetIPAddress("").
		SetUserAgent("").
		SetCreatedAt(databaseWallClock).
		Save(ctx); err != nil {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	return nil
}

func mailboxEntDraftRestartMutationKindPointer(value MailboxDraftMutationKind) *MailboxDraftMutationKind {
	copied := value
	return &copied
}
