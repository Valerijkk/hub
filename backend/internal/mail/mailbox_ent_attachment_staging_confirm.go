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

var _ MailboxAttachmentConfirmRepository = (*MailboxEntRepository)(nil)

// ConfirmMailboxAttachment prepares a stream verification or records its preliminary fingerprint.
// The attachment remains scan-pending; confirmation grants no send, read, download, or materialization authority.
func (repository *MailboxEntRepository) ConfirmMailboxAttachment(ctx context.Context, request MailboxAttachmentConfirmRepositoryRequest) (MailboxAttachmentConfirmRepositoryResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxAttachmentConfirmRepositoryRequestValid(request) {
		return MailboxAttachmentConfirmRepositoryResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}

	var result MailboxAttachmentConfirmRepositoryResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		owner, err := mailboxAttachmentConfirmLockOwner(ctx, transaction, request.ActorID)
		if err != nil {
			return err
		}
		attachment, object, err := mailboxAttachmentConfirmLockRows(ctx, transaction, owner, request.AttachmentID)
		if err != nil {
			return err
		}
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxAttachmentConfirmEntError(err)
		}

		switch request.Phase {
		case MailboxAttachmentConfirmPhasePrepare:
			return mailboxAttachmentConfirmPrepare(ctx, attachment, object, owner, databaseNow, request, &result)
		case MailboxAttachmentConfirmPhaseComplete:
			return mailboxAttachmentConfirmComplete(ctx, transaction, attachment, object, owner, databaseNow, request, &result)
		default:
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
		}
	})
	if err != nil {
		return MailboxAttachmentConfirmRepositoryResult{}, mailboxAttachmentConfirmEntError(err)
	}
	return result, nil
}

func mailboxAttachmentConfirmRepositoryRequestValid(request MailboxAttachmentConfirmRepositoryRequest) bool {
	if request.ActorID == uuid.Nil || request.AttachmentID == uuid.Nil || request.ClaimedSizeBytes < 1 || request.ClaimedSizeBytes > MaxEngineAttachmentBytes {
		return false
	}
	switch request.Phase {
	case MailboxAttachmentConfirmPhasePrepare:
		return request.Fence == nil && request.ActualSizeBytes == 0 && request.ActualSHA256 == ([sha256.Size]byte{})
	case MailboxAttachmentConfirmPhaseComplete:
		return mailboxAttachmentConfirmFenceValid(request.Fence) && request.ClaimedSizeBytes == request.ActualSizeBytes && request.ClaimedSHA256 == request.ActualSHA256 && request.Fence.AttachmentID == request.AttachmentID && request.Fence.UploaderID == request.ActorID && request.Fence.ExpectedSizeBytes == request.ClaimedSizeBytes
	default:
		return false
	}
}

func mailboxAttachmentConfirmLockOwner(ctx context.Context, transaction *ent.Tx, actorID uuid.UUID) (MailboxOwnerRecord, error) {
	owner, err := mailboxEntResolveOwnerForUpdate(ctx, transaction, actorID)
	if err != nil {
		return MailboxOwnerRecord{}, mailboxAttachmentConfirmEntError(err)
	}
	if owner.ActorID != actorID || owner.OrganizationID == uuid.Nil || owner.AccountID == uuid.Nil {
		return MailboxOwnerRecord{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorNotFound)
	}
	if owner.State == MailboxOwnerDeleting {
		return MailboxOwnerRecord{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorDeleting)
	}
	if owner.State != MailboxOwnerActive {
		return MailboxOwnerRecord{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	return owner, nil
}

func mailboxAttachmentConfirmLockRows(ctx context.Context, transaction *ent.Tx, owner MailboxOwnerRecord, attachmentID uuid.UUID) (*ent.MailAttachment, *ent.MailObject, error) {
	if ctx == nil || transaction == nil || owner.ActorID == uuid.Nil || owner.OrganizationID == uuid.Nil || owner.AccountID == uuid.Nil || attachmentID == uuid.Nil {
		return nil, nil, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}
	attachment, err := transaction.MailAttachment.Query().Where(mailattachment.IDEQ(attachmentID)).ForUpdate().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorNotFound)
		}
		return nil, nil, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	if attachment == nil || attachment.ID != attachmentID || attachment.StorageObjectID == nil || *attachment.StorageObjectID == uuid.Nil {
		return nil, nil, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
	}
	object, err := transaction.MailObject.Query().Where(mailobject.IDEQ(*attachment.StorageObjectID)).ForUpdate().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
		}
		return nil, nil, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	if object == nil || object.ID != *attachment.StorageObjectID {
		return nil, nil, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
	}
	if object.OrganizationID != owner.OrganizationID || object.MailAccountID != owner.AccountID || object.UploaderUserID != owner.ActorID {
		return nil, nil, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorNotFound)
	}
	if !mailboxAttachmentConfirmStagedRowsValid(attachment, object) {
		return nil, nil, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
	}
	return attachment, object, nil
}

func mailboxAttachmentConfirmStagedRowsValid(attachment *ent.MailAttachment, object *ent.MailObject) bool {
	if attachment == nil || object == nil || attachment.ID == uuid.Nil || object.ID == uuid.Nil || attachment.MailMessageID != nil || attachment.MailDraftID != nil || attachment.PendingMailDraftID != nil || attachment.DraftClientID == nil || *attachment.DraftClientID == uuid.Nil || attachment.StorageObjectID == nil || *attachment.StorageObjectID != object.ID || attachment.EngineBlobID != nil || attachment.SizeBytes < 1 || attachment.SizeBytes > MaxEngineAttachmentBytes || attachment.ScanStatus != mailattachment.ScanStatus("pending") || !mailboxEntValidTime(attachment.CreatedAt) || !mailboxEntValidTime(attachment.UpdatedAt) {
		return false
	}
	if object.OrganizationID == uuid.Nil || object.MailAccountID == uuid.Nil || object.UploaderUserID == uuid.Nil || object.DraftClientID != *attachment.DraftClientID || object.Bucket == "" || object.ObjectKey != mailboxAttachmentPresignObjectKey(object.OrganizationID, object.MailAccountID, object.ID) || !validMailboxObjectLocator(MailboxObjectLocator{Bucket: object.Bucket, Key: object.ObjectKey}) || object.ExpectedSizeBytes != attachment.SizeBytes || object.ExpectedSizeBytes < 1 || object.ExpectedSizeBytes > MaxEngineAttachmentBytes || object.Filename != attachment.Filename || object.MediaType != attachment.MediaType || object.NextDeleteAt != nil || object.DeleteLeaseToken != nil || object.DeleteLeaseUntil != nil || !mailboxEntValidTime(object.ExpiresAt) || !mailboxEntValidTime(object.UploadDrainUntil) || !object.UploadDrainUntil.Equal(object.ExpiresAt.Add(mailboxAttachmentPresignObjectDrainLifetime)) || !mailboxEntValidTime(object.CreatedAt) || !mailboxEntValidTime(object.UpdatedAt) {
		return false
	}
	return mailboxAttachmentConfirmPendingMetadataValid(attachment)
}

// mailboxAttachmentConfirmPendingMetadataValid validates only pending staging metadata; it never requires a fingerprint.
func mailboxAttachmentConfirmPendingMetadataValid(attachment *ent.MailAttachment) bool {
	if attachment == nil || attachment.ScanStatus != mailattachment.ScanStatus("pending") {
		return false
	}
	contentID := ""
	if attachment.ContentID != nil {
		contentID = *attachment.ContentID
	}
	decision, err := EvaluateAttachmentPolicy(AttachmentPolicyInput{
		Operation:          AttachmentPolicyOperationValidate,
		MaxAttachmentBytes: MaxEngineAttachmentBytes,
		Attachments: []AttachmentPolicyAttachment{{
			Origin:      AttachmentByteOriginUpload,
			Filename:    attachment.Filename,
			MediaType:   attachment.MediaType,
			ContentID:   contentID,
			Disposition: EngineAttachmentDisposition(attachment.Disposition),
			ScanStatus:  AttachmentScanStatusPending,
			SizeBytes:   attachment.SizeBytes,
			SHA256:      nil,
		}},
	})
	if err != nil || len(decision.Attachments) != 1 {
		return false
	}
	normalized := decision.Attachments[0]
	return normalized.Filename == attachment.Filename && normalized.MediaType == attachment.MediaType && normalized.ContentID == contentID && normalized.Disposition == EngineAttachmentDisposition(attachment.Disposition) && normalized.ScanStatus == AttachmentScanStatusPending && normalized.SizeBytes == attachment.SizeBytes && len(normalized.SHA256) == 0
}

func mailboxAttachmentConfirmPrepare(ctx context.Context, attachment *ent.MailAttachment, object *ent.MailObject, owner MailboxOwnerRecord, databaseNow time.Time, request MailboxAttachmentConfirmRepositoryRequest, result *MailboxAttachmentConfirmRepositoryResult) error {
	if ctx == nil || result == nil || !mailboxEntValidTime(databaseNow) || !mailboxAttachmentConfirmStagedRowsValid(attachment, object) {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
	}
	switch object.UploadState {
	case mailobject.UploadState("confirmed"):
		view, ok := mailboxAttachmentConfirmConfirmedView(attachment, object)
		if !ok {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
		}
		if !mailboxAttachmentConfirmViewExact(&view, MailboxAttachmentConfirmInput{SHA256: request.ClaimedSHA256, SizeBytes: request.ClaimedSizeBytes}) {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorConfirmMismatch)
		}
		*result = MailboxAttachmentConfirmRepositoryResult{Kind: MailboxAttachmentConfirmResultConfirmed, Attachment: mailboxAttachmentConfirmViewPointer(view)}
		return nil
	case mailobject.UploadState("pending"):
	default:
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
	}
	if object.ActualSizeBytes != nil || len(object.SHA256) != 0 || len(attachment.SHA256) != 0 {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
	}
	if !object.ExpiresAt.After(databaseNow) {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorExpired)
	}
	if request.ClaimedSizeBytes != object.ExpectedSizeBytes {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorConfirmMismatch)
	}
	fence := mailboxAttachmentConfirmFence{
		OrganizationID:    owner.OrganizationID,
		AccountID:         owner.AccountID,
		UploaderID:        owner.ActorID,
		DraftClientID:     object.DraftClientID,
		AttachmentID:      attachment.ID,
		ObjectID:          object.ID,
		Bucket:            object.Bucket,
		ObjectKey:         object.ObjectKey,
		ExpectedSizeBytes: object.ExpectedSizeBytes,
		ObjectExpiresAt:   object.ExpiresAt,
	}
	if !mailboxAttachmentConfirmFenceValid(&fence) {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
	}
	*result = MailboxAttachmentConfirmRepositoryResult{
		Kind:              MailboxAttachmentConfirmResultVerify,
		Object:            MailboxObjectLocator{Bucket: object.Bucket, Key: object.ObjectKey},
		ExpectedSizeBytes: object.ExpectedSizeBytes,
		ObjectExpiresAt:   object.ExpiresAt,
		Fence:             mailboxAttachmentConfirmFenceCopy(&fence),
	}
	return nil
}

func mailboxAttachmentConfirmComplete(ctx context.Context, transaction *ent.Tx, attachment *ent.MailAttachment, object *ent.MailObject, owner MailboxOwnerRecord, databaseNow time.Time, request MailboxAttachmentConfirmRepositoryRequest, result *MailboxAttachmentConfirmRepositoryResult) error {
	if ctx == nil || transaction == nil || result == nil || !mailboxEntValidTime(databaseNow) || !mailboxAttachmentConfirmStagedRowsValid(attachment, object) || !mailboxAttachmentConfirmFenceMatchesRows(request.Fence, attachment, object, owner) {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
	}
	switch object.UploadState {
	case mailobject.UploadState("confirmed"):
		view, ok := mailboxAttachmentConfirmConfirmedView(attachment, object)
		if !ok {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
		}
		if !mailboxAttachmentConfirmViewExact(&view, MailboxAttachmentConfirmInput{SHA256: request.ClaimedSHA256, SizeBytes: request.ClaimedSizeBytes}) || request.ActualSHA256 != request.ClaimedSHA256 || request.ActualSizeBytes != request.ClaimedSizeBytes {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorConfirmMismatch)
		}
		*result = MailboxAttachmentConfirmRepositoryResult{Kind: MailboxAttachmentConfirmResultConfirmed, Attachment: mailboxAttachmentConfirmViewPointer(view)}
		return nil
	case mailobject.UploadState("pending"):
	default:
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
	}
	if object.ActualSizeBytes != nil || len(object.SHA256) != 0 || len(attachment.SHA256) != 0 {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
	}
	if !object.ExpiresAt.After(databaseNow) {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorExpired)
	}
	if request.ClaimedSizeBytes != object.ExpectedSizeBytes || request.ActualSizeBytes != object.ExpectedSizeBytes || request.ClaimedSHA256 != request.ActualSHA256 {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorConfirmMismatch)
	}

	updatedObjects, err := transaction.MailObject.Update().
		Where(
			mailobject.IDEQ(object.ID),
			mailobject.OrganizationIDEQ(owner.OrganizationID),
			mailobject.MailAccountIDEQ(owner.AccountID),
			mailobject.UploaderUserIDEQ(owner.ActorID),
			mailobject.DraftClientIDEQ(request.Fence.DraftClientID),
			mailobject.BucketEQ(request.Fence.Bucket),
			mailobject.ObjectKeyEQ(request.Fence.ObjectKey),
			mailobject.ExpectedSizeBytesEQ(request.Fence.ExpectedSizeBytes),
			mailobject.ExpiresAtEQ(request.Fence.ObjectExpiresAt),
			mailobject.UploadDrainUntilEQ(object.UploadDrainUntil),
			mailobject.UploadStateEQ(mailobject.UploadState("pending")),
			mailobject.ActualSizeBytesIsNil(),
			mailobject.SHA256IsNil(),
			mailobject.NextDeleteAtIsNil(),
			mailobject.DeleteLeaseTokenIsNil(),
			mailobject.DeleteLeaseUntilIsNil(),
		).
		SetUploadState(mailobject.UploadState("confirmed")).
		SetActualSizeBytes(request.ActualSizeBytes).
		SetSHA256(append([]byte(nil), request.ActualSHA256[:]...)).
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	if updatedObjects != 1 {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
	}

	updatedAttachments, err := transaction.MailAttachment.Update().
		Where(
			mailattachment.IDEQ(attachment.ID),
			mailattachment.DraftClientIDEQ(request.Fence.DraftClientID),
			mailattachment.StorageObjectIDEQ(object.ID),
			mailattachment.MailMessageIDIsNil(),
			mailattachment.MailDraftIDIsNil(),
			mailattachment.PendingMailDraftIDIsNil(),
			mailattachment.EngineBlobIDIsNil(),
			mailattachment.SizeBytesEQ(request.Fence.ExpectedSizeBytes),
			mailattachment.ScanStatusEQ(mailattachment.ScanStatus("pending")),
			mailattachment.SHA256IsNil(),
		).
		SetSHA256(append([]byte(nil), request.ActualSHA256[:]...)).
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	if updatedAttachments != 1 {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
	}

	view, ok := mailboxAttachmentConfirmAttachmentView(attachment, request.ActualSHA256)
	if !ok {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
	}
	*result = MailboxAttachmentConfirmRepositoryResult{Kind: MailboxAttachmentConfirmResultConfirmed, Attachment: mailboxAttachmentConfirmViewPointer(view)}
	return nil
}

func mailboxAttachmentConfirmFenceMatchesRows(fence *mailboxAttachmentConfirmFence, attachment *ent.MailAttachment, object *ent.MailObject, owner MailboxOwnerRecord) bool {
	return mailboxAttachmentConfirmFenceValid(fence) && attachment != nil && object != nil && fence.OrganizationID == owner.OrganizationID && fence.AccountID == owner.AccountID && fence.UploaderID == owner.ActorID && fence.DraftClientID == object.DraftClientID && fence.AttachmentID == attachment.ID && fence.ObjectID == object.ID && fence.Bucket == object.Bucket && fence.ObjectKey == object.ObjectKey && fence.ExpectedSizeBytes == object.ExpectedSizeBytes && fence.ObjectExpiresAt.Equal(object.ExpiresAt)
}

func mailboxAttachmentConfirmConfirmedView(attachment *ent.MailAttachment, object *ent.MailObject) (MailboxAttachmentView, bool) {
	if !mailboxAttachmentConfirmStagedRowsValid(attachment, object) || object.UploadState != mailobject.UploadState("confirmed") || object.ActualSizeBytes == nil || *object.ActualSizeBytes != object.ExpectedSizeBytes || len(object.SHA256) != sha256.Size || len(attachment.SHA256) != sha256.Size || !bytes.Equal(object.SHA256, attachment.SHA256) {
		return MailboxAttachmentView{}, false
	}
	var digest [sha256.Size]byte
	copy(digest[:], attachment.SHA256)
	return mailboxAttachmentConfirmAttachmentView(attachment, digest)
}

func mailboxAttachmentConfirmAttachmentView(attachment *ent.MailAttachment, digest [sha256.Size]byte) (MailboxAttachmentView, bool) {
	if attachment == nil {
		return MailboxAttachmentView{}, false
	}
	view := MailboxAttachmentView{
		ID:          attachment.ID,
		Filename:    attachment.Filename,
		MediaType:   attachment.MediaType,
		SizeBytes:   attachment.SizeBytes,
		Disposition: EngineAttachmentDisposition(attachment.Disposition),
		ContentID:   mailboxAttachmentConfirmContentIDCopy(attachment.ContentID),
		ScanStatus:  AttachmentScanStatusPending,
		SHA256:      &digest,
	}
	if !mailboxAttachmentConfirmViewValid(view) {
		return MailboxAttachmentView{}, false
	}
	return mailboxAttachmentConfirmViewCopy(view), true
}

func mailboxAttachmentConfirmContentIDCopy(contentID *string) *string {
	if contentID == nil {
		return nil
	}
	copied := *contentID
	return &copied
}

func mailboxAttachmentConfirmViewPointer(view MailboxAttachmentView) *MailboxAttachmentView {
	copied := mailboxAttachmentConfirmViewCopy(view)
	return &copied
}

func mailboxAttachmentConfirmEntError(err error) error {
	var stagingError *MailboxAttachmentStagingError
	if errors.As(err, &stagingError) && stagingError != nil {
		return mailboxAttachmentStagingError(stagingError.Kind())
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
