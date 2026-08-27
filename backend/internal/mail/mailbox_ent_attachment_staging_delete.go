package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/mailobject"
)

var _ MailboxAttachmentAbandonRepository = (*MailboxEntRepository)(nil)

// AbandonMailboxAttachment removes one actor-owned, unlinked storage attachment and
// retains a durable intent for the object-deletion worker. It does not perform S3 I/O.
func (repository *MailboxEntRepository) AbandonMailboxAttachment(ctx context.Context, request MailboxAttachmentAbandonRepositoryRequest) (MailboxAttachmentAbandonRepositoryResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxAttachmentAbandonRepositoryRequestValid(request) {
		return MailboxAttachmentAbandonRepositoryResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}

	var result MailboxAttachmentAbandonRepositoryResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		owner, err := mailboxEntResolveOwnerForUpdate(ctx, transaction, request.ActorID)
		if err != nil {
			return mailboxAttachmentAbandonEntError(err)
		}
		if owner.ActorID != request.ActorID || owner.OrganizationID != request.OrganizationID || owner.AccountID != request.AccountID {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
		if owner.State == MailboxOwnerDeleting {
			result = MailboxAttachmentAbandonRepositoryResult{Kind: MailboxAttachmentAbandonRepositoryResultDeleting}
			return nil
		}
		if owner.State != MailboxOwnerActive {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}

		attachment, err := transaction.MailAttachment.Query().
			Where(mailattachment.IDEQ(request.AttachmentID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				result = MailboxAttachmentAbandonRepositoryResult{Kind: MailboxAttachmentAbandonRepositoryResultAbsent}
				return nil
			}
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
		if attachment == nil || attachment.ID != request.AttachmentID || attachment.StorageObjectID == nil || *attachment.StorageObjectID == uuid.Nil || attachment.EngineBlobID != nil {
			result = MailboxAttachmentAbandonRepositoryResult{Kind: MailboxAttachmentAbandonRepositoryResultAbsent}
			return nil
		}

		object, err := transaction.MailObject.Query().
			Where(mailobject.IDEQ(*attachment.StorageObjectID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				result = MailboxAttachmentAbandonRepositoryResult{Kind: MailboxAttachmentAbandonRepositoryResultAbsent}
				return nil
			}
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
		if object == nil || object.ID != *attachment.StorageObjectID {
			result = MailboxAttachmentAbandonRepositoryResult{Kind: MailboxAttachmentAbandonRepositoryResultAbsent}
			return nil
		}
		if object.OrganizationID != request.OrganizationID || object.MailAccountID != request.AccountID || object.UploaderUserID != request.ActorID {
			result = MailboxAttachmentAbandonRepositoryResult{Kind: MailboxAttachmentAbandonRepositoryResultAbsent}
			return nil
		}

		if conflict, linked := mailboxAttachmentAbandonLinkConflict(attachment); linked {
			result = MailboxAttachmentAbandonRepositoryResult{
				Kind:     MailboxAttachmentAbandonRepositoryResultConflict,
				Conflict: conflict,
			}
			return nil
		}

		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxAttachmentAbandonEntError(err)
		}
		if !mailboxAttachmentAbandonStorageRowsValid(attachment, object, request, databaseNow) {
			result = mailboxAttachmentAbandonStorageConflict()
			return nil
		}
		nextDeleteAt, ok := mailboxEntAttachmentObjectDeleteDueAt(object.UploadDrainUntil, databaseNow, request.MinimumObjectDeleteDelay)
		if !ok {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}

		deletedAttachments, err := transaction.MailAttachment.Delete().
			Where(
				mailattachment.IDEQ(attachment.ID),
				mailattachment.StorageObjectIDEQ(object.ID),
			).
			Save(ctx)
		if err != nil {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
		if deletedAttachments != 1 {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}

		nextState := mailobject.UploadState("orphaned")
		if object.UploadState == mailobject.UploadState("confirmed") {
			nextState = mailobject.UploadState("deleted")
		}
		updatedObjects, err := transaction.MailObject.Update().
			Where(
				mailobject.IDEQ(object.ID),
				mailobject.OrganizationIDEQ(request.OrganizationID),
				mailobject.MailAccountIDEQ(request.AccountID),
				mailobject.UploaderUserIDEQ(request.ActorID),
				mailobject.UploadStateEQ(object.UploadState),
			).
			SetUploadState(nextState).
			ClearDeleteLeaseToken().
			SetDeleteLeaseGeneration(0).
			ClearDeleteLeaseUntil().
			SetDeleteAttempts(0).
			SetNextDeleteAt(nextDeleteAt).
			ClearDeleteTerminalAt().
			SetUpdatedAt(databaseNow).
			Exec(ctx)
		if err != nil {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
		if updatedObjects != 1 {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}

		result = MailboxAttachmentAbandonRepositoryResult{Kind: MailboxAttachmentAbandonRepositoryResultAbandoned}
		return nil
	})
	if err != nil {
		return MailboxAttachmentAbandonRepositoryResult{}, mailboxAttachmentAbandonEntError(err)
	}
	if !mailboxAttachmentAbandonRepositoryResultValid(result) {
		return MailboxAttachmentAbandonRepositoryResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	return result, nil
}

func mailboxAttachmentAbandonRepositoryRequestValid(request MailboxAttachmentAbandonRepositoryRequest) bool {
	return request.ActorID != uuid.Nil &&
		request.OrganizationID != uuid.Nil &&
		request.AccountID != uuid.Nil &&
		request.AttachmentID != uuid.Nil &&
		request.MinimumObjectDeleteDelay >= mailboxObjectStoreMaximumGetLifetime &&
		request.MinimumObjectDeleteDelay%time.Second == 0
}

func mailboxAttachmentAbandonStorageConflict() MailboxAttachmentAbandonRepositoryResult {
	return MailboxAttachmentAbandonRepositoryResult{
		Kind:     MailboxAttachmentAbandonRepositoryResultConflict,
		Conflict: MailboxAttachmentAbandonRepositoryConflictStorage,
	}
}

func mailboxAttachmentAbandonLinkConflict(attachment *ent.MailAttachment) (MailboxAttachmentAbandonRepositoryConflictKind, bool) {
	if attachment == nil {
		return MailboxAttachmentAbandonRepositoryConflictStorage, true
	}
	if attachment.MailMessageID != nil {
		if *attachment.MailMessageID == uuid.Nil || attachment.MailDraftID != nil || attachment.PendingMailDraftID != nil {
			return MailboxAttachmentAbandonRepositoryConflictStorage, true
		}
		return MailboxAttachmentAbandonRepositoryConflictMessage, true
	}
	if attachment.PendingMailDraftID != nil {
		if *attachment.PendingMailDraftID == uuid.Nil || (attachment.MailDraftID != nil && *attachment.MailDraftID != *attachment.PendingMailDraftID) {
			return MailboxAttachmentAbandonRepositoryConflictStorage, true
		}
		return MailboxAttachmentAbandonRepositoryConflictPendingDraft, true
	}
	if attachment.MailDraftID != nil {
		if *attachment.MailDraftID == uuid.Nil {
			return MailboxAttachmentAbandonRepositoryConflictStorage, true
		}
		return MailboxAttachmentAbandonRepositoryConflictDraft, true
	}
	return "", false
}

func mailboxAttachmentAbandonStorageRowsValid(attachment *ent.MailAttachment, object *ent.MailObject, request MailboxAttachmentAbandonRepositoryRequest, databaseNow time.Time) bool {
	if attachment == nil || object == nil || attachment.ID != request.AttachmentID || attachment.StorageObjectID == nil || *attachment.StorageObjectID != object.ID || attachment.EngineBlobID != nil || attachment.MailMessageID != nil || attachment.MailDraftID != nil || attachment.PendingMailDraftID != nil || attachment.DraftClientID == nil || *attachment.DraftClientID == uuid.Nil || attachment.SizeBytes < 1 || attachment.SizeBytes > MaxEngineAttachmentBytes || object.ID == uuid.Nil || object.DeletionID == uuid.Nil || object.OrganizationID != request.OrganizationID || object.MailAccountID != request.AccountID || object.UploaderUserID != request.ActorID || object.DraftClientID != *attachment.DraftClientID || object.ExpectedSizeBytes != attachment.SizeBytes || object.Bucket == "" || object.ObjectKey != mailboxAttachmentPresignObjectKey(request.OrganizationID, request.AccountID, object.ID) || !validMailboxObjectLocator(MailboxObjectLocator{Bucket: object.Bucket, Key: object.ObjectKey}) || object.Filename != attachment.Filename || object.MediaType != attachment.MediaType || !mailboxEntValidTime(databaseNow) || !mailboxEntValidTime(attachment.CreatedAt) || !mailboxEntValidTime(attachment.UpdatedAt) || !mailboxEntValidTime(object.CreatedAt) || !mailboxEntValidTime(object.UpdatedAt) || !mailboxEntValidTime(object.ExpiresAt) || !mailboxEntValidTime(object.UploadDrainUntil) || attachment.CreatedAt.After(attachment.UpdatedAt) || attachment.UpdatedAt.After(databaseNow) || object.CreatedAt.After(object.UpdatedAt) || object.UpdatedAt.After(databaseNow) {
		return false
	}
	drainUntil, ok := mailboxEntCheckedTimeAdd(object.ExpiresAt, mailboxAttachmentPresignObjectDrainLifetime)
	if !ok || !object.UploadDrainUntil.Equal(drainUntil) {
		return false
	}
	switch attachment.ScanStatus {
	case mailattachment.ScanStatus("pending"):
		if attachment.Sha256 != nil && len(*attachment.Sha256) != sha256.Size {
			return false
		}
	case mailattachment.ScanStatus("clean"), mailattachment.ScanStatus("rejected"), mailattachment.ScanStatus("error"):
		if attachment.Sha256 == nil || len(*attachment.Sha256) != sha256.Size {
			return false
		}
	default:
		return false
	}
	switch object.UploadState {
	case mailobject.UploadState("pending"):
		return object.ActualSizeBytes == nil && object.Sha256 == nil && attachment.Sha256 == nil
	case mailobject.UploadState("confirmed"):
		return object.ActualSizeBytes != nil && *object.ActualSizeBytes == object.ExpectedSizeBytes && object.Sha256 != nil && attachment.Sha256 != nil && len(*object.Sha256) == sha256.Size && len(*attachment.Sha256) == sha256.Size && bytes.Equal(*object.Sha256, *attachment.Sha256)
	default:
		return false
	}
}

func mailboxAttachmentAbandonEntError(err error) error {
	if kind, ok := MailboxAttachmentStagingErrorKindOf(err); ok {
		return mailboxAttachmentStagingError(kind)
	}
	var serviceError *MailboxServiceError
	if errors.As(err, &serviceError) && serviceError != nil {
		switch serviceError.Kind {
		case MailboxServiceErrorInvalid:
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
		case MailboxServiceErrorDeleting:
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorDeleting)
		default:
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
	}
	return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
}
