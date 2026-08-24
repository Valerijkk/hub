package mail

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/maildraft"
)

const (
	mailboxDraftAbandonAuditAction    = "update"
	mailboxDraftAbandonAuditOperation = "abandon_create_upload"
)

// AbandonMailboxDraftCreateUpload records a terminal local tombstone for one
// exact unknown attachment upload. It only detaches local state and schedules
// object deletion; provider and object-store I/O are deliberately absent.
func (repository *MailboxEntRepository) AbandonMailboxDraftCreateUpload(ctx context.Context, request MailboxDraftAbandonCreateUploadRequest) (MailboxDraftAbandonCreateUploadResult, error) {
	if repository == nil || repository.client == nil || ctx == nil {
		return MailboxDraftAbandonCreateUploadResult{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if kind := mailboxEntDraftAbandonRequestKind(request); kind != "" {
		return MailboxDraftAbandonCreateUploadResult{}, NewMailboxDraftRepositoryError(kind)
	}

	var result MailboxDraftAbandonCreateUploadResult
	var postCommitErr error
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseWallClock, err := mailboxEntDraftCreateLockAccount(ctx, transaction, request.Fence)
		if err != nil {
			return err
		}
		if kind := mailboxEntDraftCreateAccountKind(account); kind != "" {
			return mailboxEntDraftAbandonLogicalError(ctx, account, request.Fence, databaseWallClock, kind, &postCommitErr)
		}
		if !mailboxEntDraftRecoveryReceiptAccountValid(account) {
			return mailboxEntDraftAbandonLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorInvalid, &postCommitErr)
		}
		draft, kind := mailboxEntDraftReservationLockDraft(ctx, transaction, request.Fence.AccountID, request.DraftID)
		if kind != "" {
			return mailboxEntDraftAbandonLogicalError(ctx, account, request.Fence, databaseWallClock, kind, &postCommitErr)
		}
		locked, kind := mailboxEntDraftAbandonLockAttachments(ctx, transaction, draft)
		if kind != "" {
			return mailboxEntDraftAbandonLogicalError(ctx, account, request.Fence, databaseWallClock, kind, &postCommitErr)
		}
		record, ok := mailboxEntDraftReadRecord(draft, locked.state)
		if !ok {
			return mailboxEntDraftAbandonLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}
		receiptInput := mailboxDraftRecoveryReceiptInput{
			Operation:          mailboxDraftRecoveryReceiptOperationAbandon,
			DraftID:            request.DraftID,
			IdempotencyKeyHash: request.IdempotencyKeyHash,
			RequestHash:        request.RequestHash,
			FingerprintVersion: request.FingerprintVersion,
		}
		receipt, kind := mailboxEntDraftRecoveryReceiptLock(ctx, transaction, account, receiptInput)
		if kind != "" {
			return mailboxEntDraftAbandonLogicalError(ctx, account, request.Fence, databaseWallClock, kind, &postCommitErr)
		}
		if receipt != nil {
			if !mailboxEntDraftRecoveryReceiptReplayValid(receipt, account, receiptInput, databaseWallClock) || !mailboxEntDraftAbandonPostStateValid(record, locked.state, request) {
				return mailboxEntDraftAbandonLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorIdempotencyConflict, &postCommitErr)
			}
			if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, databaseWallClock); err != nil {
				return err
			}
			result = MailboxDraftAbandonCreateUploadResult{Replayed: true, Draft: mailboxEntDraftRecordCopy(record)}
			return nil
		}
		if !mailboxEntDraftAbandonFreshStateValid(record, locked.state, request) {
			return mailboxEntDraftAbandonLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}

		freshDatabaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxEntDraftRepositoryError(err)
		}
		if !mailboxEntDraftCreateFenceMatches(account, request.Fence, freshDatabaseWallClock) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorFenceLost)
		}
		if !mailboxEntDraftAbandonFreshStateValid(record, locked.state, request) {
			return mailboxEntDraftAbandonLogicalError(ctx, account, request.Fence, freshDatabaseWallClock, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}
		if kind := mailboxEntDraftAbandonCleanup(ctx, transaction, account, draft, locked, freshDatabaseWallClock); kind != "" {
			return mailboxEntDraftAbandonLogicalError(ctx, account, request.Fence, freshDatabaseWallClock, kind, &postCommitErr)
		}

		if err := draft.Update().
			SetStatus(maildraft.StatusAbandoned).
			ClearEngineMessageID().
			ClearEngineVersion().
			ClearLastMutationID().
			ClearConfirmedMutationID().
			ClearPendingMutationKind().
			ClearPendingFromIdentityID().
			ClearPendingReplyToMessageID().
			ClearPendingForwardMessageID().
			ClearPendingSaveRequestHash().
			ClearPendingCreateAttachmentPhase().
			SetPendingCreateUploadAttempt(0).
			ClearPendingCreateUploadRetryAt().
			ClearPendingCreateUploadExhaustedAt().
			SetAbandonedAt(freshDatabaseWallClock).
			SetAbandonedMutationID(request.ExpectedMutationID).
			SetUpdatedAt(freshDatabaseWallClock).
			Exec(ctx); err != nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		updated := mailboxEntDraftRecordCopy(record)
		updated.Status = MailboxDraftStatusAbandoned
		updated.EngineMessageID = nil
		updated.EngineVersion = nil
		updated.Confirmed.AttachmentIDs = []uuid.UUID{}
		updated.PendingSave = nil
		updated.LastMutationID = nil
		updated.ConfirmedMutationID = nil
		updated.PendingMutationKind = nil
		updated.PendingCreateAttachmentPhase = nil
		updated.PendingCreateUploadAttempt = 0
		updated.PendingCreateUploadRetryAt = nil
		updated.PendingCreateUploadExhaustedAt = nil
		updated.AbandonedAt = mailboxEntDraftReadTimePointerCopy(&freshDatabaseWallClock)
		updated.AbandonedMutationID = mailboxEntDraftReadUUIDPointerCopy(&request.ExpectedMutationID)
		updated.UpdatedAt = freshDatabaseWallClock.UTC()
		emptyAttachments := mailboxEntDraftAttachmentState{ConfirmedAttachmentIDs: []uuid.UUID{}, PendingAttachmentIDs: []uuid.UUID{}}
		if !mailboxEntDraftAbandonPostStateValid(updated, emptyAttachments, request) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if _, kind := mailboxEntDraftRecoveryReceiptComplete(ctx, transaction, account, receiptInput, freshDatabaseWallClock); kind != "" {
			return NewMailboxDraftRepositoryError(kind)
		}
		if err := mailboxEntDraftAbandonAudit(ctx, transaction, account, updated, request.ExpectedAttempt, freshDatabaseWallClock); err != nil {
			return err
		}
		if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, freshDatabaseWallClock); err != nil {
			return err
		}
		result = MailboxDraftAbandonCreateUploadResult{Draft: mailboxEntDraftRecordCopy(updated)}
		return nil
	})
	if err != nil {
		return MailboxDraftAbandonCreateUploadResult{}, mailboxEntDraftRepositoryError(err)
	}
	if postCommitErr != nil {
		return MailboxDraftAbandonCreateUploadResult{}, postCommitErr
	}
	return result, nil
}

func mailboxEntDraftAbandonRequestKind(request MailboxDraftAbandonCreateUploadRequest) MailboxDraftRepositoryErrorKind {
	if !mailboxEntEngineCommandFenceValid(request.Fence) || request.DraftID == uuid.Nil || request.ExpectedMutationID == uuid.Nil || request.ExpectedAttempt < 1 || request.ExpectedAttempt > 3 || request.FingerprintVersion != mailboxDraftRecoveryReceiptFingerprintVersion || !mailboxEntDraftCreateHashPresent(request.IdempotencyKeyHash) || !mailboxEntDraftCreateHashPresent(request.RequestHash) {
		return MailboxDraftRepositoryErrorInvalid
	}
	return ""
}

func mailboxEntDraftAbandonFreshStateValid(record MailboxDraftRecord, attachments mailboxEntDraftAttachmentState, request MailboxDraftAbandonCreateUploadRequest) bool {
	if record.ID != request.DraftID || record.AccountID != request.Fence.AccountID || record.Status != MailboxDraftStatusUnknown || record.Revision != 1 || record.EngineMessageID != nil || record.EngineVersion != nil || record.PendingSave != nil || record.LastMutationID == nil || *record.LastMutationID != request.ExpectedMutationID || record.ConfirmedMutationID != nil || record.PendingMutationKind == nil || *record.PendingMutationKind != MailboxDraftMutationCreate || record.PendingCreateAttachmentPhase == nil || *record.PendingCreateAttachmentPhase != MailboxDraftCreateAttachmentPhaseUpload || record.PendingCreateUploadAttempt != request.ExpectedAttempt || len(record.Confirmed.AttachmentIDs) == 0 || len(record.Confirmed.AttachmentIDs) != len(attachments.ConfirmedAttachmentIDs) || len(attachments.PendingAttachmentIDs) != 0 || record.AbandonedAt != nil || record.AbandonedMutationID != nil {
		return false
	}
	switch request.ExpectedAttempt {
	case 1, 2:
		return record.PendingCreateUploadRetryAt != nil && record.PendingCreateUploadExhaustedAt == nil
	case 3:
		return record.PendingCreateUploadRetryAt == nil && record.PendingCreateUploadExhaustedAt != nil
	default:
		return false
	}
}

func mailboxEntDraftAbandonPostStateValid(record MailboxDraftRecord, attachments mailboxEntDraftAttachmentState, request MailboxDraftAbandonCreateUploadRequest) bool {
	return record.ID == request.DraftID && record.AccountID == request.Fence.AccountID && record.Status == MailboxDraftStatusAbandoned && record.Revision == 1 && record.EngineMessageID == nil && record.EngineVersion == nil && record.PendingSave == nil && len(record.Confirmed.AttachmentIDs) == 0 && len(attachments.ConfirmedAttachmentIDs) == 0 && len(attachments.PendingAttachmentIDs) == 0 && record.LastMutationID == nil && record.ConfirmedMutationID == nil && record.PendingMutationKind == nil && record.PendingCreateAttachmentPhase == nil && record.PendingCreateUploadAttempt == 0 && record.PendingCreateUploadRetryAt == nil && record.PendingCreateUploadExhaustedAt == nil && record.AbandonedAt != nil && mailboxEntValidTime(*record.AbandonedAt) && record.AbandonedMutationID != nil && *record.AbandonedMutationID == request.ExpectedMutationID
}

func mailboxEntDraftAbandonLogicalError(ctx context.Context, account *ent.MailAccount, fence MailboxEngineCommandFence, databaseWallClock time.Time, kind MailboxDraftRepositoryErrorKind, postCommitErr *error) error {
	if postCommitErr == nil || kind == MailboxDraftRepositoryErrorUnavailable {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	if err := mailboxEntDraftCreateRelease(ctx, account, fence, databaseWallClock); err != nil {
		return err
	}
	*postCommitErr = NewMailboxDraftRepositoryError(kind)
	return nil
}

func mailboxEntDraftAbandonAudit(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, updated MailboxDraftRecord, attempt int, databaseWallClock time.Time) error {
	if ctx == nil || transaction == nil || !mailboxEntDraftRecoveryReceiptAccountValid(account) || updated.ID == uuid.Nil || updated.Status != MailboxDraftStatusAbandoned || attempt < 1 || attempt > 3 || !mailboxEntValidTime(databaseWallClock) {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if _, err := transaction.AuditLog.Create().
		SetAction(mailboxDraftAbandonAuditAction).
		SetEntityType("mail_draft").
		SetEntityID(updated.ID.String()).
		SetUserID(account.UserID).
		SetDetails(map[string]interface{}{
			"organization_id": account.OrganizationID.String(),
			"mail_account_id": account.ID.String(),
			"operation":       mailboxDraftAbandonAuditOperation,
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
