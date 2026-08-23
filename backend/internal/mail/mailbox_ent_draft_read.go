package mail

import (
	"context"
	"sort"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/maildraft"
)

// GetMailboxDraft reads one account-scoped durable draft before any engine I/O.
func (repository *MailboxEntRepository) GetMailboxDraft(ctx context.Context, request MailboxDraftGetRequest) (MailboxDraftGetResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || request.AccountID == uuid.Nil || request.DraftID == uuid.Nil {
		return MailboxDraftGetResult{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}

	var result MailboxDraftGetResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		if err := mailboxEntDraftReadLockAccount(ctx, transaction, request.AccountID); err != nil {
			return err
		}

		draft, err := transaction.MailDraft.Query().
			Where(
				maildraft.IDEQ(request.DraftID),
				maildraft.MailAccountIDEQ(request.AccountID),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil
			}
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if !mailboxEntDraftReadRowValid(draft, request.AccountID, request.DraftID) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}

		attachmentIDs, err := mailboxEntDraftReadAttachmentIDs(ctx, draft)
		if err != nil {
			return err
		}
		record, ok := mailboxEntDraftReadRecord(draft, attachmentIDs)
		if !ok {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}
		result = MailboxDraftGetResult{Found: true, Draft: &record}
		return nil
	})
	if err != nil {
		return MailboxDraftGetResult{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}

// RecheckMailboxDraftRead confirms that an active draft still has the exact durable engine pair after engine I/O.
func (repository *MailboxEntRepository) RecheckMailboxDraftRead(ctx context.Context, request MailboxDraftReadRecheckRequest) (MailboxDraftReadRecheckResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || request.AccountID == uuid.Nil || request.DraftID == uuid.Nil || request.Revision < 1 || !syncValidOpaque(request.EngineMessageID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(request.EngineVersion, MaxEngineOpaqueIDBytes) {
		return MailboxDraftReadRecheckResult{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}

	var result MailboxDraftReadRecheckResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		if err := mailboxEntDraftReadLockAccount(ctx, transaction, request.AccountID); err != nil {
			return err
		}

		draft, err := transaction.MailDraft.Query().
			Where(
				maildraft.IDEQ(request.DraftID),
				maildraft.MailAccountIDEQ(request.AccountID),
				maildraft.StatusEQ(maildraft.StatusActive),
				maildraft.RevisionEQ(request.Revision),
				maildraft.EngineMessageIDEQ(request.EngineMessageID),
				maildraft.EngineVersionEQ(request.EngineVersion),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil
			}
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if !mailboxEntDraftReadRowValid(draft, request.AccountID, request.DraftID) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}
		result.Current = true
		return nil
	})
	if err != nil {
		return MailboxDraftReadRecheckResult{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}

func mailboxEntDraftReadLockAccount(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID) error {
	if ctx == nil || transaction == nil || accountID == uuid.Nil {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}

	account, err := transaction.MailAccount.Query().
		Where(mailaccount.IDEQ(accountID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorNotFound)
		}
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	if account == nil || account.ID != accountID {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	switch account.Status {
	case mailaccount.StatusDeleting:
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAccountDeleting)
	case mailaccount.StatusActive:
	case mailaccount.StatusProvisioning, mailaccount.StatusSuspended, mailaccount.StatusError:
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	default:
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if !account.DesiredEnabled {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
	}
	if account.EngineAccountID == nil || !syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) || account.SyncEpoch < 0 || account.SyncLeaseEpoch < 1 || account.MutationEpoch < 0 {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if (account.MutationLeaseToken == nil) != (account.MutationLeaseUntil == nil) ||
		(account.MutationLeaseToken != nil && *account.MutationLeaseToken == uuid.Nil) ||
		(account.MutationLeaseUntil != nil && !mailboxEntValidTime(*account.MutationLeaseUntil)) {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	_, _, syncLeaseValid := mailboxEntSyncLeaseFence(account)
	if !syncLeaseValid {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	return nil
}

func mailboxEntDraftReadAttachmentIDs(ctx context.Context, draft *ent.MailDraft) ([]uuid.UUID, error) {
	if ctx == nil || draft == nil || draft.ID == uuid.Nil || draft.DraftClientID == uuid.Nil {
		return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}

	attachments, err := draft.QueryAttachments().
		Order(ent.Asc(mailattachment.FieldID)).
		Limit(MaxEngineAttachmentItems + 1).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	if len(attachments) > MaxEngineAttachmentItems {
		return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}

	seen := make(map[uuid.UUID]struct{}, len(attachments))
	attachmentIDs := make([]uuid.UUID, 0, len(attachments))
	for _, attachment := range attachments {
		if !mailboxEntDraftReadAttachmentValid(attachment, draft.ID, draft.DraftClientID) {
			return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}
		if _, duplicate := seen[attachment.ID]; duplicate {
			return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}
		seen[attachment.ID] = struct{}{}
		attachmentIDs = append(attachmentIDs, attachment.ID)
	}
	sort.Slice(attachmentIDs, func(left, right int) bool {
		return syncLessUUID(attachmentIDs[left], attachmentIDs[right])
	})
	return attachmentIDs, nil
}

func mailboxEntDraftReadAttachmentValid(attachment *ent.MailAttachment, draftID, draftClientID uuid.UUID) bool {
	if attachment == nil || attachment.ID == uuid.Nil || attachment.MailDraftID == nil || *attachment.MailDraftID != draftID || attachment.MailMessageID != nil || attachment.DraftClientID == nil || *attachment.DraftClientID != draftClientID || attachment.StorageObjectID == nil || *attachment.StorageObjectID == uuid.Nil || attachment.EngineBlobID != nil || attachment.CreatedAt.IsZero() || attachment.UpdatedAt.IsZero() || !mailboxEntApplyAttachmentOptionalTextValid(attachment.ContentID) || !mailboxEntApplyAttachmentTextValid(attachment.Filename, false) || !mailboxEntApplyAttachmentTextValid(attachment.MediaType, false) || attachment.SizeBytes < 1 || attachment.SizeBytes > MaxEngineAttachmentBytes || !mailboxEntApplyAttachmentDispositionValid(attachment.Disposition) || !mailboxEntApplyAttachmentScanStatusValid(attachment.ScanStatus) || (len(attachment.SHA256) != 0 && len(attachment.SHA256) != AttachmentSHA256Bytes) || ((attachment.ScanStatus == mailattachment.ScanStatus("clean") || attachment.ScanStatus == mailattachment.ScanStatus("rejected")) && len(attachment.SHA256) != AttachmentSHA256Bytes) {
		return false
	}
	return attachment.Disposition != mailattachment.Disposition(EngineAttachmentDispositionInline) || attachment.ContentID != nil
}

func mailboxEntDraftReadRowValid(draft *ent.MailDraft, accountID, draftID uuid.UUID) bool {
	if draft == nil || draft.ID != draftID || draftID == uuid.Nil || draft.MailAccountID != accountID || accountID == uuid.Nil || draft.DraftClientID == uuid.Nil || !syncValidOpaque(draft.EngineDraftKey, MaxEngineOpaqueIDBytes) || draft.Revision < 1 || len(draft.CreateIdempotencyKeyHash) != 32 || len(draft.CreateRequestHash) != 32 || !mailboxEntDraftReadUUIDPointerValid(draft.ReplyToMessageID) || !mailboxEntDraftReadUUIDPointerValid(draft.ForwardMessageID) || !mailboxEntDraftReadUUIDPointerValid(draft.LastMutationID) {
		return false
	}

	enginePairPresent, enginePairValid := mailboxEntDraftReadEnginePairValid(draft.EngineMessageID, draft.EngineVersion)
	if !enginePairValid {
		return false
	}
	pendingKind, pendingKindValid := mailboxEntDraftReadPendingKind(draft.PendingMutationKind)
	if !pendingKindValid {
		return false
	}
	return mailboxEntDraftReadStateValid(MailboxDraftStatus(draft.Status), enginePairPresent, draft.LastMutationID, pendingKind)
}

func mailboxEntDraftReadEnginePairValid(messageID, version *string) (bool, bool) {
	if messageID == nil || version == nil {
		return false, messageID == nil && version == nil
	}
	return true, syncValidOpaque(*messageID, MaxEngineOpaqueIDBytes) && syncValidOpaque(*version, MaxEngineOpaqueIDBytes)
}

func mailboxEntDraftReadPendingKind(value *maildraft.PendingMutationKind) (*MailboxDraftMutationKind, bool) {
	if value == nil {
		return nil, true
	}

	kind := MailboxDraftMutationKind(*value)
	switch kind {
	case MailboxDraftMutationCreate, MailboxDraftMutationSave, MailboxDraftMutationDelete:
		return &kind, true
	default:
		return nil, false
	}
}

func mailboxEntDraftReadStateValid(status MailboxDraftStatus, enginePairPresent bool, mutationID *uuid.UUID, pendingKind *MailboxDraftMutationKind) bool {
	mutationPresent := mutationID != nil
	kindPresent := pendingKind != nil
	switch status {
	case MailboxDraftStatusActive, MailboxDraftStatusSubmitting, MailboxDraftStatusSubmitted:
		return enginePairPresent && !mutationPresent && !kindPresent
	case MailboxDraftStatusSaving:
		return mutationPresent && kindPresent && ((*pendingKind == MailboxDraftMutationCreate && !enginePairPresent) || (*pendingKind == MailboxDraftMutationSave && enginePairPresent))
	case MailboxDraftStatusDeleting:
		return mutationPresent && kindPresent && *pendingKind == MailboxDraftMutationDelete && enginePairPresent
	case MailboxDraftStatusUnknown:
		return mutationPresent && kindPresent && ((*pendingKind == MailboxDraftMutationCreate && !enginePairPresent) || ((*pendingKind == MailboxDraftMutationSave || *pendingKind == MailboxDraftMutationDelete) && enginePairPresent))
	case MailboxDraftStatusDeleted:
		return !mutationPresent && !kindPresent
	default:
		return false
	}
}

func mailboxEntDraftReadRecord(draft *ent.MailDraft, attachmentIDs []uuid.UUID) (MailboxDraftRecord, bool) {
	if draft == nil || !mailboxEntDraftReadRowValid(draft, draft.MailAccountID, draft.ID) || len(attachmentIDs) > MaxEngineAttachmentItems {
		return MailboxDraftRecord{}, false
	}
	for index, attachmentID := range attachmentIDs {
		if attachmentID == uuid.Nil || index > 0 && !syncLessUUID(attachmentIDs[index-1], attachmentID) {
			return MailboxDraftRecord{}, false
		}
	}

	pendingKind, ok := mailboxEntDraftReadPendingKind(draft.PendingMutationKind)
	if !ok {
		return MailboxDraftRecord{}, false
	}
	return MailboxDraftRecord{
		ID:                    draft.ID,
		AccountID:             draft.MailAccountID,
		DraftClientID:         draft.DraftClientID,
		EngineDraftKey:        draft.EngineDraftKey,
		EngineMessageID:       mailboxEntStringPointer(draft.EngineMessageID),
		EngineVersion:         mailboxEntStringPointer(draft.EngineVersion),
		Revision:              draft.Revision,
		Status:                MailboxDraftStatus(draft.Status),
		ReplyToMessageID:      mailboxEntDraftReadUUIDPointerCopy(draft.ReplyToMessageID),
		ForwardMessageID:      mailboxEntDraftReadUUIDPointerCopy(draft.ForwardMessageID),
		LastMutationID:        mailboxEntDraftReadUUIDPointerCopy(draft.LastMutationID),
		PendingMutationKind:   mailboxEntDraftReadMutationKindPointerCopy(pendingKind),
		SelectedAttachmentIDs: append([]uuid.UUID(nil), attachmentIDs...),
	}, true
}

func mailboxEntDraftReadUUIDPointerValid(value *uuid.UUID) bool {
	return value == nil || *value != uuid.Nil
}

func mailboxEntDraftReadUUIDPointerCopy(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func mailboxEntDraftReadMutationKindPointerCopy(value *MailboxDraftMutationKind) *MailboxDraftMutationKind {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
