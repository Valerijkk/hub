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

var _ MailboxAttachmentDownloadRepository = (*MailboxEntRepository)(nil)

// ResolveMailboxAttachmentDownload returns one scoped object locator without holding a transaction across storage I/O.
func (repository *MailboxEntRepository) ResolveMailboxAttachmentDownload(ctx context.Context, request MailboxAttachmentDownloadRepositoryRequest) (MailboxAttachmentDownloadRepositoryResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxAttachmentDownloadRepositoryRequestValid(request) {
		return MailboxAttachmentDownloadRepositoryResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}

	var result MailboxAttachmentDownloadRepositoryResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		owner, err := mailboxEntResolveOwnerForUpdate(ctx, transaction, request.ActorID)
		if err != nil {
			return mailboxAttachmentDownloadEntError(err)
		}
		if owner.ActorID != request.ActorID || owner.OrganizationID == uuid.Nil || owner.AccountID == uuid.Nil {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
		if owner.State == MailboxOwnerDeleting {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorDeleting)
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
				return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorNotFound)
			}
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
		if attachment == nil || attachment.ID != request.AttachmentID {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
		if attachment.ScanStatus != mailattachment.ScanStatus(AttachmentScanStatusClean) {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorNotClean)
		}
		if !mailboxAttachmentDownloadAttachmentValid(attachment) {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
		}

		object, err := transaction.MailObject.Query().
			Where(
				mailobject.IDEQ(*attachment.StorageObjectID),
				mailobject.OrganizationIDEQ(owner.OrganizationID),
				mailobject.MailAccountIDEQ(owner.AccountID),
				mailobject.UploaderUserIDEQ(owner.ActorID),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorNotFound)
			}
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
		if object == nil {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}

		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxAttachmentDownloadEntError(err)
		}
		if !mailboxAttachmentDownloadObjectValid(object, attachment, owner, databaseNow) {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
		}

		result = MailboxAttachmentDownloadRepositoryResult{
			Object:   MailboxObjectLocator{Bucket: object.Bucket, Key: object.ObjectKey},
			IssuedAt: databaseNow,
		}
		return nil
	})
	if err != nil {
		return MailboxAttachmentDownloadRepositoryResult{}, mailboxAttachmentDownloadEntError(err)
	}
	if !mailboxAttachmentDownloadRepositoryResultValid(result) {
		return MailboxAttachmentDownloadRepositoryResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	return result, nil
}

func mailboxAttachmentDownloadRepositoryRequestValid(request MailboxAttachmentDownloadRepositoryRequest) bool {
	return request.ActorID != uuid.Nil && request.AttachmentID != uuid.Nil
}

func mailboxAttachmentDownloadAttachmentValid(attachment *ent.MailAttachment) bool {
	if attachment == nil || attachment.ID == uuid.Nil || attachment.DraftClientID == nil || *attachment.DraftClientID == uuid.Nil || attachment.StorageObjectID == nil || *attachment.StorageObjectID == uuid.Nil || attachment.EngineBlobID != nil || attachment.SizeBytes < 1 || attachment.SizeBytes > MaxEngineAttachmentBytes || attachment.ScanStatus != mailattachment.ScanStatus(AttachmentScanStatusClean) || attachment.Sha256 == nil || len(*attachment.Sha256) != sha256.Size || attachment.ScanLeaseToken != nil || attachment.ScanLeaseUntil != nil || attachment.ScanLeaseGeneration < 1 || attachment.ScanAttempts < 1 || attachment.ScanAttempts > 10 || attachment.NextScanAt != nil || attachment.ScanTerminalAt == nil || !mailboxEntValidTime(*attachment.ScanTerminalAt) || !mailboxEntValidTime(attachment.CreatedAt) || !mailboxEntValidTime(attachment.UpdatedAt) || attachment.CreatedAt.After(attachment.UpdatedAt) {
		return false
	}
	if attachment.MailMessageID == nil && attachment.MailDraftID == nil && attachment.PendingMailDraftID == nil || attachment.MailMessageID != nil && (attachment.MailDraftID != nil || attachment.PendingMailDraftID != nil) || attachment.MailDraftID != nil && attachment.PendingMailDraftID != nil && *attachment.MailDraftID != *attachment.PendingMailDraftID {
		return false
	}
	contentID := ""
	if attachment.ContentID != nil {
		contentID = *attachment.ContentID
	}
	decision, err := EvaluateAttachmentPolicy(AttachmentPolicyInput{
		Operation:          AttachmentPolicyOperationRead,
		MaxAttachmentBytes: MaxEngineAttachmentBytes,
		Attachments: []AttachmentPolicyAttachment{{
			Origin:      AttachmentByteOriginUpload,
			Filename:    attachment.Filename,
			MediaType:   attachment.MediaType,
			ContentID:   contentID,
			Disposition: EngineAttachmentDisposition(attachment.Disposition),
			ScanStatus:  AttachmentScanStatusClean,
			SizeBytes:   attachment.SizeBytes,
			SHA256:      append([]byte(nil), (*attachment.Sha256)...),
		}},
	})
	if err != nil || len(decision.Attachments) != 1 {
		return false
	}
	normalized := decision.Attachments[0]
	return normalized.Filename == attachment.Filename && normalized.MediaType == attachment.MediaType && normalized.ContentID == contentID && normalized.Disposition == EngineAttachmentDisposition(attachment.Disposition) && normalized.ScanStatus == AttachmentScanStatusClean && normalized.SizeBytes == attachment.SizeBytes && bytes.Equal(normalized.SHA256, *attachment.Sha256)
}

func mailboxAttachmentDownloadObjectValid(object *ent.MailObject, attachment *ent.MailAttachment, owner MailboxOwnerRecord, databaseNow time.Time) bool {
	if object == nil || attachment == nil || owner.ActorID == uuid.Nil || owner.OrganizationID == uuid.Nil || owner.AccountID == uuid.Nil || object.ID == uuid.Nil || object.DeletionID == uuid.Nil || attachment.StorageObjectID == nil || *attachment.StorageObjectID != object.ID || object.OrganizationID != owner.OrganizationID || object.MailAccountID != owner.AccountID || object.UploaderUserID != owner.ActorID || object.DraftClientID != *attachment.DraftClientID || object.Bucket == "" || object.ObjectKey != mailboxAttachmentPresignObjectKey(owner.OrganizationID, owner.AccountID, object.ID) || !validMailboxObjectLocator(MailboxObjectLocator{Bucket: object.Bucket, Key: object.ObjectKey}) || object.ExpectedSizeBytes != attachment.SizeBytes || object.ExpectedSizeBytes < 1 || object.ExpectedSizeBytes > MaxEngineAttachmentBytes || object.ActualSizeBytes == nil || *object.ActualSizeBytes != object.ExpectedSizeBytes || object.Filename != attachment.Filename || object.MediaType != attachment.MediaType || object.Sha256 == nil || attachment.Sha256 == nil || len(*object.Sha256) != sha256.Size || !bytes.Equal(*object.Sha256, *attachment.Sha256) || object.UploadState != mailobject.UploadState("confirmed") || object.NextDeleteAt != nil || object.DeleteLeaseToken != nil || object.DeleteLeaseUntil != nil || object.DeleteLeaseGeneration != 0 || object.DeleteAttempts != 0 || object.DeleteTerminalAt != nil || !mailboxEntValidTime(object.ExpiresAt) || !mailboxEntValidTime(object.UploadDrainUntil) || !object.UploadDrainUntil.Equal(object.ExpiresAt.Add(mailboxAttachmentPresignObjectDrainLifetime)) || !mailboxEntValidTime(object.CreatedAt) || !mailboxEntValidTime(object.UpdatedAt) || object.CreatedAt.After(object.UpdatedAt) || !mailboxEntValidTime(databaseNow) || databaseNow.Before(object.UploadDrainUntil) || databaseNow.Before(*attachment.ScanTerminalAt) {
		return false
	}
	return true
}

func mailboxAttachmentDownloadEntError(err error) error {
	if kind, ok := MailboxAttachmentStagingErrorKindOf(err); ok {
		return mailboxAttachmentStagingError(kind)
	}
	var serviceError *MailboxServiceError
	if errors.As(err, &serviceError) && serviceError != nil {
		switch serviceError.Kind {
		case MailboxServiceErrorInvalid:
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
		case MailboxServiceErrorNotFound:
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorNotFound)
		case MailboxServiceErrorDeleting:
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorDeleting)
		default:
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
	}
	return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
}
