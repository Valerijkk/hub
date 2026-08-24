package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailobject"
)

// ReserveMailboxDraftCreate persists or recovers one draft-create operation under the exact acquired fence.
func (repository *MailboxEntRepository) ReserveMailboxDraftCreate(ctx context.Context, request MailboxDraftCreateRequest) (MailboxDraftReservation, error) {
	if repository == nil || repository.client == nil || ctx == nil {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if kind := mailboxEntDraftCreateRequestKind(request); kind != "" {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(kind)
	}

	engineDraftKey, ok := mailboxEntDraftCreateKey(request.DraftID)
	if !ok {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}

	var result MailboxDraftReservation
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseWallClock, err := mailboxEntDraftCreateLockAccount(ctx, transaction, request.Fence)
		if err != nil {
			return err
		}
		if kind := mailboxEntDraftCreateAccountKind(account); kind != "" {
			return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, kind)
		}

		rows, err := mailboxEntDraftCreateLockIdentityRows(ctx, transaction, request)
		if err != nil {
			return err
		}
		if len(rows) != 0 {
			return mailboxEntDraftCreateClassifyExisting(ctx, account, databaseWallClock, request, engineDraftKey, rows, &result)
		}

		if err := mailboxEntDraftCreateLockReference(ctx, transaction, request.Fence.AccountID, request.ReplyToMessageID); err != nil {
			return mailboxEntDraftCreateReleaseForError(ctx, account, request.Fence, databaseWallClock, err)
		}
		if err := mailboxEntDraftCreateLockReference(ctx, transaction, request.Fence.AccountID, request.ForwardMessageID); err != nil {
			return mailboxEntDraftCreateReleaseForError(ctx, account, request.Fence, databaseWallClock, err)
		}
		if err := mailboxEntDraftCreateLockAttachments(ctx, transaction, request); err != nil {
			return mailboxEntDraftCreateReleaseForError(ctx, account, request.Fence, databaseWallClock, err)
		}

		draftCreate := transaction.MailDraft.Create().
			SetID(request.DraftID).
			SetMailAccountID(request.Fence.AccountID).
			SetDraftClientID(request.DraftClientID).
			SetEngineDraftKey(engineDraftKey).
			SetRevision(1).
			SetStatus(maildraft.StatusSaving).
			SetCreateIdempotencyKeyHash(append([]byte(nil), request.CreateIdempotencyKeyHash[:]...)).
			SetCreateRequestHash(append([]byte(nil), request.CreateRequestHash[:]...)).
			SetLastMutationID(request.MutationID).
			SetPendingMutationKind(maildraft.PendingMutationKind("create")).
			SetCreatedAt(databaseWallClock).
			SetUpdatedAt(databaseWallClock)
		if request.ReplyToMessageID != nil {
			draftCreate.SetReplyToMessageID(*request.ReplyToMessageID)
		}
		if request.ForwardMessageID != nil {
			draftCreate.SetForwardMessageID(*request.ForwardMessageID)
		}
		if len(request.SelectedAttachmentIDs) != 0 {
			draftCreate.AddAttachmentIDs(request.SelectedAttachmentIDs...)
		}

		draft, err := draftCreate.Save(ctx)
		if err != nil {
			if ent.IsConstraintError(err) {
				racedRows, readErr := mailboxEntDraftCreateLockIdentityRows(ctx, transaction, request)
				if readErr != nil {
					return readErr
				}
				if len(racedRows) != 0 {
					return mailboxEntDraftCreateClassifyExisting(ctx, account, databaseWallClock, request, engineDraftKey, racedRows, &result)
				}
				return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorIdempotencyConflict)
			}
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}

		record, ok := mailboxEntDraftReadRecord(draft, request.SelectedAttachmentIDs)
		if !ok {
			return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
		fence := mailboxEntDraftCreateFenceCopy(request.Fence)
		result = MailboxDraftReservation{
			Kind:  MailboxDraftReservationDispatch,
			Draft: record,
			Work:  mailboxEntDraftCreateWork(record, MailboxDraftWorkCreate),
			Fence: &fence,
		}
		return nil
	})
	if err != nil {
		return MailboxDraftReservation{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}

func mailboxEntDraftCreateRequestKind(request MailboxDraftCreateRequest) MailboxDraftRepositoryErrorKind {
	if !mailboxEntEngineCommandFenceValid(request.Fence) || request.DraftID == uuid.Nil || request.DraftClientID == uuid.Nil || request.MutationID == uuid.Nil || !mailboxEntDraftCreateHashPresent(request.CreateIdempotencyKeyHash) || !mailboxEntDraftCreateHashPresent(request.CreateRequestHash) || !mailboxEntDraftCreateUUIDPointerValid(request.ReplyToMessageID) || !mailboxEntDraftCreateUUIDPointerValid(request.ForwardMessageID) {
		return MailboxDraftRepositoryErrorInvalid
	}
	if len(request.SelectedAttachmentIDs) > MaxEngineAttachmentItems {
		return MailboxDraftRepositoryErrorAttachmentLimit
	}
	for index, attachmentID := range request.SelectedAttachmentIDs {
		if attachmentID == uuid.Nil || (index > 0 && !syncLessUUID(request.SelectedAttachmentIDs[index-1], attachmentID)) {
			return MailboxDraftRepositoryErrorInvalid
		}
	}
	return ""
}

func mailboxEntDraftCreateHashPresent(value [sha256.Size]byte) bool {
	var zero [sha256.Size]byte
	return value != zero
}

func mailboxEntDraftCreateUUIDPointerValid(value *uuid.UUID) bool {
	return value == nil || *value != uuid.Nil
}

func mailboxEntDraftCreateKey(draftID uuid.UUID) (string, bool) {
	if draftID == uuid.Nil {
		return "", false
	}
	material := make([]byte, 0, len("hub-mail-draft-v1\x00")+len(draftID))
	material = append(material, "hub-mail-draft-v1\x00"...)
	material = append(material, draftID[:]...)
	sum := sha256.Sum256(material)
	key := "d1_" + base64.RawURLEncoding.EncodeToString(sum[:])
	return key, syncValidOpaque(key, MaxEngineOpaqueIDBytes)
}

func mailboxEntDraftCreateLockAccount(ctx context.Context, transaction *ent.Tx, fence MailboxEngineCommandFence) (*ent.MailAccount, time.Time, error) {
	if ctx == nil || transaction == nil || !mailboxEntEngineCommandFenceValid(fence) {
		return nil, time.Time{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}

	account, err := transaction.MailAccount.Query().
		Where(mailaccount.IDEQ(fence.AccountID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, time.Time{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorNotFound)
		}
		return nil, time.Time{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	if account == nil || account.ID != fence.AccountID {
		return nil, time.Time{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}

	databaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
	if err != nil {
		return nil, time.Time{}, mailboxEntDraftRepositoryError(err)
	}
	if !mailboxEntDraftCreateFenceMatches(account, fence, databaseWallClock) {
		return nil, time.Time{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorFenceLost)
	}
	return account, databaseWallClock, nil
}

func mailboxEntDraftCreateFenceMatches(account *ent.MailAccount, fence MailboxEngineCommandFence, databaseWallClock time.Time) bool {
	return account != nil &&
		account.ID == fence.AccountID &&
		account.MutationLeaseToken != nil &&
		*account.MutationLeaseToken == fence.LeaseID &&
		account.MutationLeaseUntil != nil &&
		mailboxEntValidTime(*account.MutationLeaseUntil) &&
		account.MutationLeaseUntil.Equal(fence.LeaseUntil) &&
		account.MutationEpoch == fence.MutationEpoch &&
		mailboxEntValidTime(databaseWallClock) &&
		account.MutationLeaseUntil.After(databaseWallClock)
}

func mailboxEntDraftCreateAccountKind(account *ent.MailAccount) MailboxDraftRepositoryErrorKind {
	if account == nil || account.ID == uuid.Nil {
		return MailboxDraftRepositoryErrorInvalid
	}
	switch account.Status {
	case mailaccount.StatusDeleting:
		return MailboxDraftRepositoryErrorAccountDeleting
	case mailaccount.StatusActive:
	default:
		return MailboxDraftRepositoryErrorStateConflict
	}
	if !account.DesiredEnabled {
		return MailboxDraftRepositoryErrorStateConflict
	}
	if account.EngineAccountID == nil || !syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) || account.SyncEpoch < 0 || account.MutationEpoch < 1 || account.SyncLeaseEpoch < 1 || !mailboxEntValidTime(account.SyncScheduleUpdatedAt) {
		return MailboxDraftRepositoryErrorInvalid
	}
	_, _, syncLeaseValid := mailboxEntSyncLeaseFence(account)
	if !syncLeaseValid {
		return MailboxDraftRepositoryErrorInvalid
	}
	return ""
}

func mailboxEntDraftCreateLockIdentityRows(ctx context.Context, transaction *ent.Tx, request MailboxDraftCreateRequest) ([]*ent.MailDraft, error) {
	if ctx == nil || transaction == nil {
		return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	rows, err := transaction.MailDraft.Query().
		Where(
			maildraft.MailAccountIDEQ(request.Fence.AccountID),
			maildraft.Or(
				maildraft.IDEQ(request.DraftID),
				maildraft.DraftClientIDEQ(request.DraftClientID),
				maildraft.CreateIdempotencyKeyHashEQ(request.CreateIdempotencyKeyHash[:]),
			),
		).
		Order(ent.Asc(maildraft.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	return rows, nil
}

func mailboxEntDraftCreateClassifyExisting(ctx context.Context, account *ent.MailAccount, databaseWallClock time.Time, request MailboxDraftCreateRequest, engineDraftKey string, rows []*ent.MailDraft, result *MailboxDraftReservation) error {
	if result == nil || len(rows) != 1 || !mailboxEntDraftCreateSameRequest(rows[0], request, engineDraftKey) {
		return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorIdempotencyConflict)
	}
	draft := rows[0]
	if !mailboxEntDraftReadRowValid(draft, request.Fence.AccountID, request.DraftID) {
		return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
	}
	attachmentIDs, err := mailboxEntDraftReadAttachmentIDs(ctx, draft)
	if err != nil {
		return mailboxEntDraftCreateReleaseForError(ctx, account, request.Fence, databaseWallClock, err)
	}
	record, ok := mailboxEntDraftReadRecord(draft, attachmentIDs)
	if !ok {
		return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
	}

	switch record.Status {
	case MailboxDraftStatusActive, MailboxDraftStatusSubmitting, MailboxDraftStatusSubmitted, MailboxDraftStatusDeleted:
		if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, databaseWallClock); err != nil {
			return err
		}
		*result = MailboxDraftReservation{Kind: MailboxDraftReservationReplay, Draft: record}
		return nil
	case MailboxDraftStatusSaving, MailboxDraftStatusUnknown:
		if record.PendingMutationKind == nil || *record.PendingMutationKind != MailboxDraftMutationCreate || record.LastMutationID == nil || record.EngineMessageID != nil || record.EngineVersion != nil {
			return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
		fence := mailboxEntDraftCreateFenceCopy(request.Fence)
		*result = MailboxDraftReservation{
			Kind:  MailboxDraftReservationReconcileRequired,
			Draft: record,
			Work:  mailboxEntDraftCreateWork(record, MailboxDraftWorkReconcileCreate),
			Fence: &fence,
		}
		return nil
	default:
		return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
	}
}

func mailboxEntDraftCreateSameRequest(draft *ent.MailDraft, request MailboxDraftCreateRequest, engineDraftKey string) bool {
	return draft != nil &&
		draft.ID == request.DraftID &&
		draft.MailAccountID == request.Fence.AccountID &&
		draft.DraftClientID == request.DraftClientID &&
		draft.EngineDraftKey == engineDraftKey &&
		bytes.Equal(draft.CreateIdempotencyKeyHash, request.CreateIdempotencyKeyHash[:]) &&
		bytes.Equal(draft.CreateRequestHash, request.CreateRequestHash[:])
}

func mailboxEntDraftCreateLockReference(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID, messageID *uuid.UUID) error {
	if messageID == nil {
		return nil
	}
	if ctx == nil || transaction == nil || accountID == uuid.Nil || *messageID == uuid.Nil {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	message, err := transaction.MailMessage.Query().
		Where(
			mailmessage.IDEQ(*messageID),
			mailmessage.MailAccountIDEQ(accountID),
			mailmessage.DeletedAtIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	if message == nil || message.ID != *messageID || message.MailAccountID != accountID || message.DeletedAt != nil {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
	}
	return nil
}

func mailboxEntDraftCreateLockAttachments(ctx context.Context, transaction *ent.Tx, request MailboxDraftCreateRequest) error {
	for _, attachmentID := range request.SelectedAttachmentIDs {
		attachment, err := transaction.MailAttachment.Query().
			Where(mailattachment.IDEQ(attachmentID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAttachmentConflict)
			}
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if !mailboxEntDraftCreateAttachmentValid(attachment, request.DraftClientID) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAttachmentConflict)
		}

		object, err := transaction.MailObject.Query().
			Where(mailobject.IDEQ(*attachment.StorageObjectID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAttachmentConflict)
			}
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if !mailboxEntDraftCreateStorageObjectValid(object, request.Fence.AccountID, request.DraftClientID, attachment) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAttachmentConflict)
		}
	}
	return nil
}

func mailboxEntDraftCreateAttachmentValid(attachment *ent.MailAttachment, draftClientID uuid.UUID) bool {
	if attachment == nil || attachment.ID == uuid.Nil || attachment.MailMessageID != nil || attachment.MailDraftID != nil || attachment.DraftClientID == nil || *attachment.DraftClientID != draftClientID || attachment.StorageObjectID == nil || *attachment.StorageObjectID == uuid.Nil || attachment.EngineBlobID != nil || attachment.SizeBytes < 1 || attachment.SizeBytes > MaxEngineAttachmentBytes || attachment.ScanStatus != mailattachment.ScanStatus("clean") || len(attachment.SHA256) != AttachmentSHA256Bytes || !mailboxEntApplyAttachmentOptionalTextValid(attachment.ContentID) || !mailboxEntApplyAttachmentTextValid(attachment.Filename, false) || !mailboxEntApplyAttachmentTextValid(attachment.MediaType, false) || !mailboxEntApplyAttachmentDispositionValid(attachment.Disposition) {
		return false
	}
	return attachment.Disposition != mailattachment.Disposition(EngineAttachmentDispositionInline) || attachment.ContentID != nil
}

func mailboxEntDraftCreateStorageObjectValid(object *ent.MailObject, accountID, draftClientID uuid.UUID, attachment *ent.MailAttachment) bool {
	return object != nil &&
		attachment != nil &&
		attachment.StorageObjectID != nil &&
		object.ID == *attachment.StorageObjectID &&
		object.MailAccountID == accountID &&
		object.DraftClientID == draftClientID &&
		object.UploadState == mailobject.UploadState("confirmed") &&
		object.ActualSizeBytes != nil &&
		*object.ActualSizeBytes == object.ExpectedSizeBytes &&
		object.ExpectedSizeBytes == attachment.SizeBytes &&
		len(object.SHA256) == AttachmentSHA256Bytes &&
		bytes.Equal(object.SHA256, attachment.SHA256) &&
		object.NextDeleteAt == nil &&
		object.DeleteLeaseToken == nil &&
		object.DeleteLeaseUntil == nil
}

func mailboxEntDraftCreateReleaseForError(ctx context.Context, account *ent.MailAccount, fence MailboxEngineCommandFence, databaseWallClock time.Time, err error) error {
	if repositoryError, ok := err.(*MailboxDraftRepositoryError); ok && repositoryError != nil && repositoryError.Kind() == MailboxDraftRepositoryErrorUnavailable {
		return err
	}
	return mailboxEntDraftCreateReleaseError(ctx, account, fence, databaseWallClock, mailboxEntDraftCreateErrorKind(err))
}

func mailboxEntDraftCreateErrorKind(err error) MailboxDraftRepositoryErrorKind {
	if repositoryError, ok := err.(*MailboxDraftRepositoryError); ok && repositoryError != nil {
		return repositoryError.Kind()
	}
	return MailboxDraftRepositoryErrorUnavailable
}

func mailboxEntDraftCreateReleaseError(ctx context.Context, account *ent.MailAccount, fence MailboxEngineCommandFence, databaseWallClock time.Time, kind MailboxDraftRepositoryErrorKind) error {
	if err := mailboxEntDraftCreateRelease(ctx, account, fence, databaseWallClock); err != nil {
		return err
	}
	return NewMailboxDraftRepositoryError(kind)
}

func mailboxEntDraftCreateRelease(ctx context.Context, account *ent.MailAccount, fence MailboxEngineCommandFence, databaseWallClock time.Time) error {
	if !mailboxEntDraftCreateFenceMatches(account, fence, databaseWallClock) {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorFenceLost)
	}
	nextMutationEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
	if !ok {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if err := account.Update().
		ClearMutationLeaseToken().
		ClearMutationLeaseUntil().
		SetMutationEpoch(nextMutationEpoch).
		SetSyncScheduleUpdatedAt(databaseWallClock).
		Exec(ctx); err != nil {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	return nil
}

func mailboxEntDraftCreateWork(record MailboxDraftRecord, kind MailboxDraftWorkKind) *MailboxDraftWork {
	if record.LastMutationID == nil || record.ID == uuid.Nil || record.AccountID == uuid.Nil || record.Revision < 1 || !syncValidOpaque(record.EngineDraftKey, MaxEngineOpaqueIDBytes) {
		return nil
	}
	return &MailboxDraftWork{
		Kind:                  kind,
		EngineDraftKey:        record.EngineDraftKey,
		EngineMessageID:       mailboxEntStringPointer(record.EngineMessageID),
		EngineVersion:         mailboxEntStringPointer(record.EngineVersion),
		MutationID:            *record.LastMutationID,
		Revision:              record.Revision,
		SelectedAttachmentIDs: append([]uuid.UUID(nil), record.SelectedAttachmentIDs...),
	}
}

func mailboxEntDraftCreateFenceCopy(fence MailboxEngineCommandFence) MailboxEngineCommandFence {
	return MailboxEngineCommandFence{
		AccountID:     fence.AccountID,
		LeaseID:       fence.LeaseID,
		MutationEpoch: fence.MutationEpoch,
		LeaseUntil:    fence.LeaseUntil,
	}
}
