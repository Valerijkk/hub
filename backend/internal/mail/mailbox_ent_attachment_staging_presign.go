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
	"github.com/valerii/hub/backend/ent/mailmutationreceipt"
	"github.com/valerii/hub/backend/ent/mailmutationreceipttarget"
	"github.com/valerii/hub/backend/ent/mailobject"
)

var _ MailboxAttachmentPresignRepository = (*MailboxEntRepository)(nil)

// ReserveMailboxAttachmentPresign atomically reserves or validates one pending attachment upload.
func (repository *MailboxEntRepository) ReserveMailboxAttachmentPresign(ctx context.Context, request MailboxAttachmentPresignRepositoryRequest) (MailboxAttachmentPresignRepositoryResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxAttachmentPresignRepositoryRequestValid(request) {
		return MailboxAttachmentPresignRepositoryResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}

	var result MailboxAttachmentPresignRepositoryResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		owner, err := mailboxEntResolveOwnerForUpdate(ctx, transaction, request.ActorID)
		if err != nil {
			return mailboxAttachmentPresignEntError(err)
		}
		if owner.ActorID != request.ActorID || owner.OrganizationID != request.OrganizationID || owner.AccountID != request.AccountID {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
		}
		if owner.State == MailboxOwnerDeleting {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorDeleting)
		}
		if owner.State != MailboxOwnerActive {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxAttachmentPresignEntError(err)
		}

		receipts, err := transaction.MailMutationReceipt.Query().
			Where(
				mailmutationreceipt.OrganizationIDEQ(request.OrganizationID),
				mailmutationreceipt.ActorUserIDEQ(request.ActorID),
				mailmutationreceipt.OperationEQ(mailmutationreceipt.Operation(mailboxAttachmentPresignOperation)),
				mailmutationreceipt.IdempotencyKeyHashEQ(request.IdempotencyKeyHash[:]...),
			).
			ForUpdate().
			All(ctx)
		if err != nil {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
		if len(receipts) > 1 {
			return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
		if len(receipts) == 1 {
			return mailboxAttachmentPresignReplay(ctx, transaction, databaseNow, request, receipts[0], &result)
		}

		return mailboxAttachmentPresignReserveFresh(ctx, transaction, databaseNow, request, &result)
	})
	if err != nil {
		return MailboxAttachmentPresignRepositoryResult{}, mailboxAttachmentPresignEntError(err)
	}
	return result, nil
}

func mailboxAttachmentPresignRepositoryRequestValid(request MailboxAttachmentPresignRepositoryRequest) bool {
	if request.ActorID == uuid.Nil || request.OrganizationID == uuid.Nil || request.AccountID == uuid.Nil || request.DraftClientID == uuid.Nil || request.AttachmentID == uuid.Nil || request.ObjectID == uuid.Nil || request.AttachmentID == request.ObjectID || request.FingerprintVersion != mailboxAttachmentPresignFingerprintVersion || !mailboxAttachmentPresignDigestValid(request.IdempotencyKeyHash[:]) || !mailboxAttachmentPresignDigestValid(request.RequestHash[:]) || request.ObjectKey != mailboxAttachmentPresignObjectKey(request.OrganizationID, request.AccountID, request.ObjectID) || !validMailboxObjectLocator(MailboxObjectLocator{Bucket: request.Bucket, Key: request.ObjectKey}) {
		return false
	}
	contentID := ""
	if request.ContentID != nil {
		contentID = *request.ContentID
	}
	decision, err := EvaluateAttachmentPolicy(AttachmentPolicyInput{
		Operation:          AttachmentPolicyOperationValidate,
		MaxAttachmentBytes: MaxEngineAttachmentBytes,
		Attachments: []AttachmentPolicyAttachment{{
			Origin:      AttachmentByteOriginUpload,
			Filename:    request.Filename,
			MediaType:   request.MediaType,
			ContentID:   contentID,
			Disposition: request.Disposition,
			ScanStatus:  AttachmentScanStatusPending,
			SizeBytes:   request.SizeBytes,
		}},
	})
	if err != nil || len(decision.Attachments) != 1 {
		return false
	}
	normalized := decision.Attachments[0]
	return normalized.Filename == request.Filename && normalized.MediaType == request.MediaType && normalized.Disposition == request.Disposition && normalized.SizeBytes == request.SizeBytes && mailboxEntApplyAttachmentOptionalTextEqual(mailboxAttachmentPresignContentIDPointer(normalized.ContentID), request.ContentID)
}

func mailboxAttachmentPresignDigestValid(value []byte) bool {
	if len(value) != sha256.Size {
		return false
	}
	for _, byteValue := range value {
		if byteValue != 0 {
			return true
		}
	}
	return false
}

func mailboxAttachmentPresignReplay(ctx context.Context, transaction *ent.Tx, databaseNow time.Time, request MailboxAttachmentPresignRepositoryRequest, receipt *ent.MailMutationReceipt, result *MailboxAttachmentPresignRepositoryResult) error {
	if result == nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}
	if kind := mailboxAttachmentPresignReceiptReplayKind(receipt, databaseNow, request); kind != "" {
		return mailboxAttachmentStagingError(kind)
	}
	targets, err := transaction.MailMutationReceiptTarget.Query().
		Where(mailmutationreceipttarget.ReceiptIDEQ(receipt.ID)).
		Order(ent.Asc(mailmutationreceipttarget.FieldRequestIndex)).
		ForUpdate().
		All(ctx)
	if err != nil || len(targets) != 1 || !mailboxAttachmentPresignReceiptTargetValid(targets[0], receipt) {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorIdempotencyUnavailable)
	}
	attachment, err := transaction.MailAttachment.Query().Where(mailattachment.IDEQ(*receipt.TargetID)).ForUpdate().Only(ctx)
	if err != nil || attachment == nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorIdempotencyUnavailable)
	}
	if attachment.StorageObjectID == nil || *attachment.StorageObjectID == uuid.Nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorIdempotencyUnavailable)
	}
	object, err := transaction.MailObject.Query().Where(mailobject.IDEQ(*attachment.StorageObjectID)).ForUpdate().Only(ctx)
	if err != nil || object == nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorIdempotencyUnavailable)
	}
	if !mailboxAttachmentPresignAttachmentReplayValid(attachment, request) || !mailboxAttachmentPresignObjectReplayValid(object, request) || attachment.ID != *receipt.TargetID || *attachment.StorageObjectID != object.ID {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorIdempotencyUnavailable)
	}
	if !object.ExpiresAt.After(databaseNow) {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorIdempotencyUnavailable)
	}
	*result = MailboxAttachmentPresignRepositoryResult{
		AttachmentID:    attachment.ID,
		Object:          MailboxObjectLocator{Bucket: object.Bucket, Key: object.ObjectKey},
		IssuedAt:        databaseNow,
		ObjectExpiresAt: object.ExpiresAt,
	}
	return nil
}

func mailboxAttachmentPresignReceiptReplayKind(receipt *ent.MailMutationReceipt, databaseNow time.Time, request MailboxAttachmentPresignRepositoryRequest) MailboxAttachmentStagingErrorKind {
	if receipt == nil || receipt.ID == uuid.Nil || receipt.OrganizationID != request.OrganizationID || receipt.ActorUserID != request.ActorID || receipt.MailAccountID == nil || *receipt.MailAccountID != request.AccountID || receipt.Operation != mailmutationreceipt.Operation(mailboxAttachmentPresignOperation) || receipt.KeySource != mailmutationreceipt.KeySource("client") || !bytes.Equal(receipt.IdempotencyKeyHash, request.IdempotencyKeyHash[:]) {
		return MailboxAttachmentStagingErrorIdempotencyUnavailable
	}
	if !bytes.Equal(receipt.RequestHash, request.RequestHash[:]) {
		return MailboxAttachmentStagingErrorIdempotencyReused
	}
	if receipt.FingerprintVersion != mailboxAttachmentPresignFingerprintVersion {
		return MailboxAttachmentStagingErrorIdempotencyUnavailable
	}
	if receipt.ReplayUntil == nil || !mailboxEntValidTime(*receipt.ReplayUntil) || !databaseNow.Before(*receipt.ReplayUntil) || receipt.CompactedAt != nil {
		return MailboxAttachmentStagingErrorIdempotencyExpired
	}
	if receipt.State != mailmutationreceipt.State("completed") || receipt.TargetKind == nil || *receipt.TargetKind != mailmutationreceipt.TargetKind("attachment") || receipt.TargetID == nil || *receipt.TargetID == uuid.Nil || receipt.OriginalHTTPStatus == nil || *receipt.OriginalHTTPStatus != 201 || receipt.SafeCode != nil || receipt.BulkAction != nil || receipt.ArgumentKind != nil || receipt.ArgumentID != nil || receipt.NextReconcileAt != nil || receipt.ReconcileLeaseToken != nil || receipt.ReconcileLeaseUntil != nil || receipt.TerminalAt == nil || !mailboxEntValidTime(*receipt.TerminalAt) || !receipt.ReplayUntil.Equal(receipt.TerminalAt.Add(mailboxAttachmentPresignReplayLifetime)) {
		return MailboxAttachmentStagingErrorIdempotencyUnavailable
	}
	return ""
}

func mailboxAttachmentPresignReceiptTargetValid(target *ent.MailMutationReceiptTarget, receipt *ent.MailMutationReceipt) bool {
	return target != nil && receipt != nil && target.ID != uuid.Nil && target.ReceiptID == receipt.ID && target.RequestIndex == 0 && target.TargetType == mailmutationreceipttarget.TargetType("attachment") && receipt.TargetID != nil && target.TargetID == *receipt.TargetID && target.Status == mailmutationreceipttarget.Status("succeeded") && target.AffectedCount == 1 && target.SucceededCount == 1 && target.FailedCount == 0 && target.TrainingStatus == mailmutationreceipttarget.TrainingStatus("not_requested") && target.SafeCode == nil && mailboxEntValidTime(target.CreatedAt) && mailboxEntValidTime(target.UpdatedAt)
}

func mailboxAttachmentPresignAttachmentReplayValid(attachment *ent.MailAttachment, request MailboxAttachmentPresignRepositoryRequest) bool {
	return attachment != nil && attachment.ID != uuid.Nil && attachment.MailMessageID == nil && attachment.MailDraftID == nil && attachment.PendingMailDraftID == nil && attachment.DraftClientID != nil && *attachment.DraftClientID == request.DraftClientID && attachment.StorageObjectID != nil && *attachment.StorageObjectID != uuid.Nil && attachment.EngineBlobID == nil && mailboxEntApplyAttachmentOptionalTextEqual(attachment.ContentID, request.ContentID) && attachment.Filename == request.Filename && attachment.MediaType == request.MediaType && attachment.SizeBytes == request.SizeBytes && attachment.Disposition == mailattachment.Disposition(request.Disposition) && attachment.ScanStatus == mailattachment.ScanStatus("pending") && attachment.Sha256 == nil && mailboxEntValidTime(attachment.CreatedAt) && mailboxEntValidTime(attachment.UpdatedAt)
}

func mailboxAttachmentPresignObjectReplayValid(object *ent.MailObject, request MailboxAttachmentPresignRepositoryRequest) bool {
	if object == nil || object.ID == uuid.Nil || object.OrganizationID != request.OrganizationID || object.MailAccountID != request.AccountID || object.UploaderUserID != request.ActorID || object.DraftClientID != request.DraftClientID || object.Bucket != request.Bucket || object.ObjectKey != mailboxAttachmentPresignObjectKey(request.OrganizationID, request.AccountID, object.ID) || object.ExpectedSizeBytes != request.SizeBytes || object.ActualSizeBytes != nil || object.Filename != request.Filename || object.MediaType != request.MediaType || object.Sha256 != nil || object.UploadState != mailobject.UploadState("pending") || object.NextDeleteAt != nil || object.DeleteLeaseToken != nil || object.DeleteLeaseUntil != nil || !mailboxEntValidTime(object.ExpiresAt) || !mailboxEntValidTime(object.UploadDrainUntil) || !object.UploadDrainUntil.Equal(object.ExpiresAt.Add(mailboxAttachmentPresignObjectDrainLifetime)) || !mailboxEntValidTime(object.CreatedAt) || !mailboxEntValidTime(object.UpdatedAt) {
		return false
	}
	return validMailboxObjectLocator(MailboxObjectLocator{Bucket: object.Bucket, Key: object.ObjectKey})
}

func mailboxAttachmentPresignReserveFresh(ctx context.Context, transaction *ent.Tx, databaseNow time.Time, request MailboxAttachmentPresignRepositoryRequest, result *MailboxAttachmentPresignRepositoryResult) error {
	if result == nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}
	unclaimed, err := transaction.MailAttachment.Query().
		Where(
			mailattachment.DraftClientIDEQ(request.DraftClientID),
			mailattachment.MailMessageIDIsNil(),
			mailattachment.MailDraftIDIsNil(),
			mailattachment.PendingMailDraftIDIsNil(),
			mailattachment.HasStorageObjectWith(mailobject.OrganizationIDEQ(request.OrganizationID), mailobject.MailAccountIDEQ(request.AccountID), mailobject.UploaderUserIDEQ(request.ActorID)),
		).
		Order(ent.Asc(mailattachment.FieldID)).
		Limit(mailboxAttachmentPresignMaximumUnclaimedAttachments).
		ForUpdate().
		All(ctx)
	if err != nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	if len(unclaimed) >= mailboxAttachmentPresignMaximumUnclaimedAttachments {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorLimit)
	}

	objectExpiresAt, ok := mailboxEntCheckedTimeAdd(databaseNow, mailboxAttachmentPresignObjectLifetime)
	if !ok {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	uploadDrainUntil, ok := mailboxEntCheckedTimeAdd(objectExpiresAt, mailboxAttachmentPresignObjectDrainLifetime)
	if !ok {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	replayUntil, ok := mailboxEntCheckedTimeAdd(databaseNow, mailboxAttachmentPresignReplayLifetime)
	if !ok {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}

	receipt, err := transaction.MailMutationReceipt.Create().
		SetOrganizationID(request.OrganizationID).
		SetActorUserID(request.ActorID).
		SetMailAccountID(request.AccountID).
		SetOperation(mailmutationreceipt.Operation(mailboxAttachmentPresignOperation)).
		SetKeySource(mailmutationreceipt.KeySource("client")).
		SetState(mailmutationreceipt.State("completed")).
		SetIdempotencyKeyHash(append([]byte(nil), request.IdempotencyKeyHash[:]...)).
		SetRequestHash(append([]byte(nil), request.RequestHash[:]...)).
		SetFingerprintVersion(request.FingerprintVersion).
		SetTargetKind(mailmutationreceipt.TargetKind("attachment")).
		SetTargetID(request.AttachmentID).
		SetOriginalHTTPStatus(201).
		SetTerminalAt(databaseNow).
		SetReplayUntil(replayUntil).
		SetCreatedAt(databaseNow).
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil || receipt == nil || receipt.ID == uuid.Nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}

	object, err := transaction.MailObject.Create().
		SetID(request.ObjectID).
		SetOrganizationID(request.OrganizationID).
		SetMailAccountID(request.AccountID).
		SetUploaderUserID(request.ActorID).
		SetDraftClientID(request.DraftClientID).
		SetBucket(request.Bucket).
		SetObjectKey(request.ObjectKey).
		SetExpectedSizeBytes(request.SizeBytes).
		SetFilename(request.Filename).
		SetMediaType(request.MediaType).
		SetUploadState(mailobject.UploadState("pending")).
		SetExpiresAt(objectExpiresAt).
		SetUploadDrainUntil(uploadDrainUntil).
		SetCreatedAt(databaseNow).
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil || object == nil || object.ID != request.ObjectID || object.ObjectKey != request.ObjectKey || !mailboxAttachmentPresignObjectReplayValid(object, request) {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}

	attachmentCreate := transaction.MailAttachment.Create().
		SetID(request.AttachmentID).
		SetDraftClientID(request.DraftClientID).
		SetStorageObjectID(request.ObjectID).
		SetFilename(request.Filename).
		SetMediaType(request.MediaType).
		SetSizeBytes(request.SizeBytes).
		SetDisposition(mailattachment.Disposition(request.Disposition)).
		SetScanStatus(mailattachment.ScanStatus("pending")).
		SetCreatedAt(databaseNow).
		SetUpdatedAt(databaseNow)
	if request.ContentID != nil {
		attachmentCreate.SetContentID(*request.ContentID)
	}
	attachment, err := attachmentCreate.Save(ctx)
	if err != nil || attachment == nil || attachment.ID != request.AttachmentID || attachment.StorageObjectID == nil || *attachment.StorageObjectID != object.ID || !mailboxAttachmentPresignAttachmentReplayValid(attachment, request) {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}

	target, err := transaction.MailMutationReceiptTarget.Create().
		SetReceiptID(receipt.ID).
		SetRequestIndex(0).
		SetTargetType(mailmutationreceipttarget.TargetType("attachment")).
		SetTargetID(request.AttachmentID).
		SetStatus(mailmutationreceipttarget.Status("succeeded")).
		SetAffectedCount(1).
		SetSucceededCount(1).
		SetFailedCount(0).
		SetTrainingStatus(mailmutationreceipttarget.TrainingStatus("not_requested")).
		SetCreatedAt(databaseNow).
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil || !mailboxAttachmentPresignReceiptTargetValid(target, receipt) {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}

	*result = MailboxAttachmentPresignRepositoryResult{
		AttachmentID:    attachment.ID,
		Object:          MailboxObjectLocator{Bucket: object.Bucket, Key: object.ObjectKey},
		IssuedAt:        databaseNow,
		ObjectExpiresAt: object.ExpiresAt,
	}
	return nil
}

func mailboxAttachmentPresignEntError(err error) error {
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
		}
	}
	return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
}
