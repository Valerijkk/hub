package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/mailobject"
)

var _ mailboxAttachmentScanRepository = (*MailboxEntRepository)(nil)

// prepareMailboxAttachmentScan locks the durable proof for one drain-complete confirmed attachment.
func (repository *MailboxEntRepository) prepareMailboxAttachmentScan(ctx context.Context, request mailboxAttachmentScanPrepareRequest) (mailboxAttachmentScanPrepared, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxAttachmentScanPrepareRequestValid(request) {
		return mailboxAttachmentScanPrepared{}, mailboxAttachmentScanError(MailboxAttachmentScanErrorInvalid)
	}

	var prepared mailboxAttachmentScanPrepared
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		owner, err := mailboxAttachmentScanLockOwner(ctx, transaction, request.ActorID, request.OrganizationID, request.AccountID)
		if err != nil {
			return err
		}
		attachment, object, err := mailboxAttachmentScanLockRows(ctx, transaction, owner, request.AttachmentID)
		if err != nil {
			return err
		}
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxAttachmentScanEntError(err)
		}
		if !mailboxAttachmentScanConfirmedRowsValid(attachment, object) {
			return mailboxAttachmentScanError(MailboxAttachmentScanErrorStateConflict)
		}
		if databaseNow.Before(object.UploadDrainUntil) {
			return mailboxAttachmentScanError(MailboxAttachmentScanErrorNotReady)
		}

		fence, ok := mailboxAttachmentScanFenceFromRows(owner, attachment, object)
		if !ok {
			return mailboxAttachmentScanError(MailboxAttachmentScanErrorStateConflict)
		}
		prepared = mailboxAttachmentScanPrepared{
			Object:        MailboxObjectLocator{Bucket: object.Bucket, Key: object.ObjectKey},
			ContentLength: object.ExpectedSizeBytes,
			Fence:         mailboxAttachmentScanFenceCopy(&fence),
		}
		return nil
	})
	if err != nil {
		return mailboxAttachmentScanPrepared{}, mailboxAttachmentScanEntError(err)
	}
	if !mailboxAttachmentScanPreparedValid(prepared, request.ActorID, request.OrganizationID, request.AccountID, request.AttachmentID) {
		return mailboxAttachmentScanPrepared{}, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	return prepared, nil
}

// completeMailboxAttachmentScan records exactly one clean or rejected scanner verdict under a fresh durable fence.
func (repository *MailboxEntRepository) completeMailboxAttachmentScan(ctx context.Context, request mailboxAttachmentScanCompleteRequest) (mailboxAttachmentScanCompletion, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxAttachmentScanCompleteRequestValid(request) {
		return mailboxAttachmentScanCompletion{}, mailboxAttachmentScanError(MailboxAttachmentScanErrorInvalid)
	}

	var result mailboxAttachmentScanCompletion
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		owner, err := mailboxAttachmentScanLockOwner(ctx, transaction, request.ActorID, request.OrganizationID, request.AccountID)
		if err != nil {
			return err
		}
		attachment, object, err := mailboxAttachmentScanLockRows(ctx, transaction, owner, request.AttachmentID)
		if err != nil {
			return err
		}
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxAttachmentScanEntError(err)
		}
		if !mailboxAttachmentScanConfirmedRowsValid(attachment, object) || !mailboxAttachmentScanFenceMatchesRows(request.Fence, owner, attachment, object) {
			return mailboxAttachmentScanError(MailboxAttachmentScanErrorStateConflict)
		}
		if databaseNow.Before(object.UploadDrainUntil) {
			return mailboxAttachmentScanError(MailboxAttachmentScanErrorNotReady)
		}

		updated, err := transaction.MailAttachment.Update().
			Where(
				mailattachment.IDEQ(attachment.ID),
				mailattachment.DraftClientIDEQ(request.Fence.DraftClientID),
				mailattachment.StorageObjectIDEQ(request.Fence.ObjectID),
				mailattachment.MailMessageIDIsNil(),
				mailattachment.MailDraftIDIsNil(),
				mailattachment.PendingMailDraftIDIsNil(),
				mailattachment.EngineBlobIDIsNil(),
				mailattachment.SizeBytesEQ(request.Fence.ExpectedSizeBytes),
				mailattachment.ScanStatusEQ(mailattachment.ScanStatus("pending")),
				mailattachment.SHA256EQ(request.Fence.ExpectedSHA256[:]),
				mailattachment.UpdatedAtEQ(request.Fence.AttachmentUpdatedAt),
			).
			SetScanStatus(mailattachment.ScanStatus(request.ScanStatus)).
			SetUpdatedAt(databaseNow).
			Save(ctx)
		if err != nil {
			return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
		}
		if updated != 1 {
			return mailboxAttachmentScanError(MailboxAttachmentScanErrorStateConflict)
		}
		result = mailboxAttachmentScanCompletion{attachmentID: attachment.ID, scanStatus: request.ScanStatus}
		return nil
	})
	if err != nil {
		return mailboxAttachmentScanCompletion{}, mailboxAttachmentScanEntError(err)
	}
	if result.attachmentID != request.AttachmentID || result.scanStatus != request.ScanStatus {
		return mailboxAttachmentScanCompletion{}, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	return result, nil
}

func mailboxAttachmentScanPrepareRequestValid(request mailboxAttachmentScanPrepareRequest) bool {
	return request.ActorID != uuid.Nil && request.OrganizationID != uuid.Nil && request.AccountID != uuid.Nil && request.AttachmentID != uuid.Nil
}

func mailboxAttachmentScanCompleteRequestValid(request mailboxAttachmentScanCompleteRequest) bool {
	return mailboxAttachmentScanPrepareRequestValid(mailboxAttachmentScanPrepareRequest{
		ActorID:        request.ActorID,
		OrganizationID: request.OrganizationID,
		AccountID:      request.AccountID,
		AttachmentID:   request.AttachmentID,
	}) && (request.ScanStatus == AttachmentScanStatusClean || request.ScanStatus == AttachmentScanStatusRejected) && mailboxAttachmentScanFenceValid(request.Fence) && request.Fence.UploaderID == request.ActorID && request.Fence.OrganizationID == request.OrganizationID && request.Fence.AccountID == request.AccountID && request.Fence.AttachmentID == request.AttachmentID
}

func mailboxAttachmentScanLockOwner(ctx context.Context, transaction *ent.Tx, actorID, organizationID, accountID uuid.UUID) (MailboxOwnerRecord, error) {
	owner, err := mailboxEntResolveOwnerForUpdate(ctx, transaction, actorID)
	if err != nil {
		return MailboxOwnerRecord{}, mailboxAttachmentScanEntError(err)
	}
	if owner.ActorID != actorID || owner.OrganizationID != organizationID || owner.AccountID != accountID {
		return MailboxOwnerRecord{}, mailboxAttachmentScanError(MailboxAttachmentScanErrorNotFound)
	}
	if owner.State == MailboxOwnerDeleting {
		return MailboxOwnerRecord{}, mailboxAttachmentScanError(MailboxAttachmentScanErrorDeleting)
	}
	if owner.State != MailboxOwnerActive {
		return MailboxOwnerRecord{}, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	return owner, nil
}

func mailboxAttachmentScanLockRows(ctx context.Context, transaction *ent.Tx, owner MailboxOwnerRecord, attachmentID uuid.UUID) (*ent.MailAttachment, *ent.MailObject, error) {
	if ctx == nil || transaction == nil || owner.ActorID == uuid.Nil || owner.OrganizationID == uuid.Nil || owner.AccountID == uuid.Nil || attachmentID == uuid.Nil {
		return nil, nil, mailboxAttachmentScanError(MailboxAttachmentScanErrorInvalid)
	}
	attachment, err := transaction.MailAttachment.Query().Where(mailattachment.IDEQ(attachmentID)).ForUpdate().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil, mailboxAttachmentScanError(MailboxAttachmentScanErrorNotFound)
		}
		return nil, nil, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	if attachment == nil || attachment.ID != attachmentID || attachment.StorageObjectID == nil || *attachment.StorageObjectID == uuid.Nil {
		return nil, nil, mailboxAttachmentScanError(MailboxAttachmentScanErrorStateConflict)
	}
	object, err := transaction.MailObject.Query().Where(mailobject.IDEQ(*attachment.StorageObjectID)).ForUpdate().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil, mailboxAttachmentScanError(MailboxAttachmentScanErrorStateConflict)
		}
		return nil, nil, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	if object == nil || object.ID != *attachment.StorageObjectID {
		return nil, nil, mailboxAttachmentScanError(MailboxAttachmentScanErrorStateConflict)
	}
	if object.OrganizationID != owner.OrganizationID || object.MailAccountID != owner.AccountID || object.UploaderUserID != owner.ActorID {
		return nil, nil, mailboxAttachmentScanError(MailboxAttachmentScanErrorNotFound)
	}
	return attachment, object, nil
}

func mailboxAttachmentScanConfirmedRowsValid(attachment *ent.MailAttachment, object *ent.MailObject) bool {
	if attachment == nil || object == nil || attachment.ID == uuid.Nil || object.ID == uuid.Nil || attachment.MailMessageID != nil || attachment.MailDraftID != nil || attachment.PendingMailDraftID != nil || attachment.DraftClientID == nil || *attachment.DraftClientID == uuid.Nil || attachment.StorageObjectID == nil || *attachment.StorageObjectID != object.ID || attachment.EngineBlobID != nil || attachment.SizeBytes < 1 || attachment.SizeBytes > MaxEngineAttachmentBytes || attachment.ScanStatus != mailattachment.ScanStatus("pending") || len(attachment.SHA256) != sha256.Size || !mailboxEntValidTime(attachment.CreatedAt) || !mailboxEntValidTime(attachment.UpdatedAt) {
		return false
	}
	if !mailboxAttachmentConfirmPendingMetadataValid(attachment) {
		return false
	}
	if object.OrganizationID == uuid.Nil || object.MailAccountID == uuid.Nil || object.UploaderUserID == uuid.Nil || object.DraftClientID != *attachment.DraftClientID || object.Bucket == "" || object.ObjectKey != mailboxAttachmentPresignObjectKey(object.OrganizationID, object.MailAccountID, object.ID) || !validMailboxObjectLocator(MailboxObjectLocator{Bucket: object.Bucket, Key: object.ObjectKey}) || object.ExpectedSizeBytes != attachment.SizeBytes || object.ExpectedSizeBytes < 1 || object.ExpectedSizeBytes > MaxEngineAttachmentBytes || object.ActualSizeBytes == nil || *object.ActualSizeBytes != object.ExpectedSizeBytes || object.Filename != attachment.Filename || object.MediaType != attachment.MediaType || len(object.SHA256) != sha256.Size || !bytes.Equal(object.SHA256, attachment.SHA256) || object.UploadState != mailobject.UploadState("confirmed") || object.NextDeleteAt != nil || object.DeleteLeaseToken != nil || object.DeleteLeaseUntil != nil || !mailboxEntValidTime(object.ExpiresAt) || !mailboxEntValidTime(object.UploadDrainUntil) || !object.UploadDrainUntil.Equal(object.ExpiresAt.Add(mailboxAttachmentPresignObjectDrainLifetime)) || !mailboxEntValidTime(object.CreatedAt) || !mailboxEntValidTime(object.UpdatedAt) {
		return false
	}
	return true
}

func mailboxAttachmentScanFenceFromRows(owner MailboxOwnerRecord, attachment *ent.MailAttachment, object *ent.MailObject) (mailboxAttachmentScanFence, bool) {
	if owner.ActorID == uuid.Nil || owner.OrganizationID == uuid.Nil || owner.AccountID == uuid.Nil || !mailboxAttachmentScanConfirmedRowsValid(attachment, object) {
		return mailboxAttachmentScanFence{}, false
	}
	var digest [sha256.Size]byte
	copy(digest[:], attachment.SHA256)
	fence := mailboxAttachmentScanFence{
		OrganizationID:      owner.OrganizationID,
		AccountID:           owner.AccountID,
		UploaderID:          owner.ActorID,
		DraftClientID:       object.DraftClientID,
		AttachmentID:        attachment.ID,
		ObjectID:            object.ID,
		Bucket:              object.Bucket,
		ObjectKey:           object.ObjectKey,
		ExpectedSizeBytes:   object.ExpectedSizeBytes,
		ExpectedSHA256:      digest,
		ObjectExpiresAt:     object.ExpiresAt,
		UploadDrainUntil:    object.UploadDrainUntil,
		AttachmentUpdatedAt: attachment.UpdatedAt,
		ObjectUpdatedAt:     object.UpdatedAt,
	}
	return fence, mailboxAttachmentScanFenceValid(&fence)
}

func mailboxAttachmentScanFenceMatchesRows(fence *mailboxAttachmentScanFence, owner MailboxOwnerRecord, attachment *ent.MailAttachment, object *ent.MailObject) bool {
	if !mailboxAttachmentScanFenceValid(fence) || !mailboxAttachmentScanConfirmedRowsValid(attachment, object) {
		return false
	}
	return fence.OrganizationID == owner.OrganizationID && fence.AccountID == owner.AccountID && fence.UploaderID == owner.ActorID && fence.DraftClientID == object.DraftClientID && fence.AttachmentID == attachment.ID && fence.ObjectID == object.ID && fence.Bucket == object.Bucket && fence.ObjectKey == object.ObjectKey && fence.ExpectedSizeBytes == object.ExpectedSizeBytes && bytes.Equal(fence.ExpectedSHA256[:], attachment.SHA256) && fence.ObjectExpiresAt.Equal(object.ExpiresAt) && fence.UploadDrainUntil.Equal(object.UploadDrainUntil) && fence.AttachmentUpdatedAt.Equal(attachment.UpdatedAt) && fence.ObjectUpdatedAt.Equal(object.UpdatedAt)
}

func mailboxAttachmentScanEntError(err error) error {
	if kind, ok := MailboxAttachmentScanErrorKindOf(err); ok {
		return mailboxAttachmentScanError(kind)
	}
	var serviceError *MailboxServiceError
	if errors.As(err, &serviceError) && serviceError != nil {
		switch serviceError.Kind {
		case MailboxServiceErrorInvalid:
			return mailboxAttachmentScanError(MailboxAttachmentScanErrorInvalid)
		case MailboxServiceErrorNotFound:
			return mailboxAttachmentScanError(MailboxAttachmentScanErrorNotFound)
		case MailboxServiceErrorDeleting:
			return mailboxAttachmentScanError(MailboxAttachmentScanErrorDeleting)
		default:
			return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
		}
	}
	return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
}
