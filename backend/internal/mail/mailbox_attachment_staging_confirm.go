package mail

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"

	"github.com/google/uuid"
)

const mailboxAttachmentConfirmBufferBytes = 32 * 1024

// ConfirmMailboxAttachment verifies one staged object outside durable locks and records a preliminary fingerprint.
// The attachment remains scan-pending; confirmation grants no send, read, download, or materialization authority.
func (service *MailboxAttachmentStagingService) ConfirmMailboxAttachment(ctx context.Context, actorID, attachmentID uuid.UUID, input MailboxAttachmentConfirmInput) (MailboxAttachmentView, error) {
	if !mailboxAttachmentConfirmServiceValid(service) || ctx == nil || actorID == uuid.Nil || attachmentID == uuid.Nil || input.SizeBytes < 1 || input.SizeBytes > MaxEngineAttachmentBytes {
		return MailboxAttachmentView{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}

	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxAttachmentView{}, mailboxAttachmentConfirmOwnerError(err)
	}
	if owner.State() == MailboxOwnerDeleting {
		return MailboxAttachmentView{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorDeleting)
	}
	accountID, hasAccount := owner.AccountID()
	if owner.State() != MailboxOwnerActive || owner.OrganizationID() == uuid.Nil || !hasAccount || accountID == uuid.Nil {
		return MailboxAttachmentView{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}

	prepared, err := service.confirms.ConfirmMailboxAttachment(ctx, MailboxAttachmentConfirmRepositoryRequest{
		Phase:            MailboxAttachmentConfirmPhasePrepare,
		ActorID:          actorID,
		AttachmentID:     attachmentID,
		ClaimedSHA256:    input.SHA256,
		ClaimedSizeBytes: input.SizeBytes,
	})
	if err != nil {
		return MailboxAttachmentView{}, mailboxAttachmentConfirmStagingError(err)
	}

	switch prepared.Kind {
	case MailboxAttachmentConfirmResultConfirmed:
		if !mailboxAttachmentConfirmViewExact(prepared.Attachment, input) {
			return MailboxAttachmentView{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorConfirmMismatch)
		}
		return mailboxAttachmentConfirmViewCopy(*prepared.Attachment), nil
	case MailboxAttachmentConfirmResultVerify:
		if !mailboxAttachmentConfirmVerifyResultValid(prepared, actorID, attachmentID, owner.OrganizationID(), accountID) {
			return MailboxAttachmentView{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
		if input.SizeBytes != prepared.ExpectedSizeBytes {
			return MailboxAttachmentView{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorConfirmMismatch)
		}
	default:
		return MailboxAttachmentView{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}

	actualSHA256, actualSizeBytes, err := mailboxAttachmentConfirmStreamSHA256(service.objectStore, ctx, prepared.Object, prepared.ExpectedSizeBytes)
	if err != nil {
		return MailboxAttachmentView{}, err
	}
	if actualSizeBytes != prepared.ExpectedSizeBytes || actualSizeBytes != input.SizeBytes || actualSHA256 != input.SHA256 {
		return MailboxAttachmentView{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorConfirmMismatch)
	}

	completed, err := service.confirms.ConfirmMailboxAttachment(ctx, MailboxAttachmentConfirmRepositoryRequest{
		Phase:            MailboxAttachmentConfirmPhaseComplete,
		ActorID:          actorID,
		AttachmentID:     attachmentID,
		ClaimedSHA256:    input.SHA256,
		ClaimedSizeBytes: input.SizeBytes,
		ActualSHA256:     actualSHA256,
		ActualSizeBytes:  actualSizeBytes,
		Fence:            mailboxAttachmentConfirmFenceCopy(prepared.Fence),
	})
	if err != nil {
		return MailboxAttachmentView{}, mailboxAttachmentConfirmStagingError(err)
	}
	if completed.Kind != MailboxAttachmentConfirmResultConfirmed {
		return MailboxAttachmentView{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	if !mailboxAttachmentConfirmViewExact(completed.Attachment, input) {
		return MailboxAttachmentView{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorConfirmMismatch)
	}
	return mailboxAttachmentConfirmViewCopy(*completed.Attachment), nil
}

func mailboxAttachmentConfirmServiceValid(service *MailboxAttachmentStagingService) bool {
	return service != nil && !mailboxAttachmentStagingDependencyNil(service.owners) && !mailboxAttachmentStagingDependencyNil(service.confirms) && !mailboxAttachmentStagingDependencyNil(service.objectStore)
}

func mailboxAttachmentConfirmVerifyResultValid(result MailboxAttachmentConfirmRepositoryResult, actorID, attachmentID, organizationID, accountID uuid.UUID) bool {
	if result.Kind != MailboxAttachmentConfirmResultVerify || result.Attachment != nil || !validMailboxObjectLocator(result.Object) || result.ExpectedSizeBytes < 1 || result.ExpectedSizeBytes > MaxEngineAttachmentBytes || !mailboxEntValidTime(result.ObjectExpiresAt) || !mailboxAttachmentConfirmFenceValid(result.Fence) {
		return false
	}
	fence := result.Fence
	return fence.UploaderID == actorID && fence.AttachmentID == attachmentID && fence.OrganizationID == organizationID && fence.AccountID == accountID && result.ExpectedSizeBytes == fence.ExpectedSizeBytes && result.ObjectExpiresAt.Equal(fence.ObjectExpiresAt) && result.Object == (MailboxObjectLocator{Bucket: fence.Bucket, Key: fence.ObjectKey})
}

func mailboxAttachmentConfirmFenceValid(fence *mailboxAttachmentConfirmFence) bool {
	return fence != nil && fence.OrganizationID != uuid.Nil && fence.AccountID != uuid.Nil && fence.UploaderID != uuid.Nil && fence.DraftClientID != uuid.Nil && fence.AttachmentID != uuid.Nil && fence.ObjectID != uuid.Nil && fence.AttachmentID != fence.ObjectID && fence.ExpectedSizeBytes >= 1 && fence.ExpectedSizeBytes <= MaxEngineAttachmentBytes && fence.ObjectKey == mailboxAttachmentPresignObjectKey(fence.OrganizationID, fence.AccountID, fence.ObjectID) && validMailboxObjectLocator(MailboxObjectLocator{Bucket: fence.Bucket, Key: fence.ObjectKey}) && mailboxEntValidTime(fence.ObjectExpiresAt)
}

func mailboxAttachmentConfirmFenceCopy(fence *mailboxAttachmentConfirmFence) *mailboxAttachmentConfirmFence {
	if fence == nil {
		return nil
	}
	copied := *fence
	return &copied
}

func mailboxAttachmentConfirmViewExact(view *MailboxAttachmentView, input MailboxAttachmentConfirmInput) bool {
	return view != nil && mailboxAttachmentConfirmViewValid(*view) && view.SizeBytes == input.SizeBytes && view.SHA256 != nil && *view.SHA256 == input.SHA256
}

func mailboxAttachmentConfirmViewValid(view MailboxAttachmentView) bool {
	if view.ID == uuid.Nil || view.SizeBytes < 1 || view.SizeBytes > MaxEngineAttachmentBytes || view.SHA256 == nil || view.ScanStatus != AttachmentScanStatusPending {
		return false
	}
	contentID := ""
	if view.ContentID != nil {
		contentID = *view.ContentID
	}
	decision, err := EvaluateAttachmentPolicy(AttachmentPolicyInput{
		Operation:          AttachmentPolicyOperationValidate,
		MaxAttachmentBytes: MaxEngineAttachmentBytes,
		Attachments: []AttachmentPolicyAttachment{{
			Origin:      AttachmentByteOriginUpload,
			Filename:    view.Filename,
			MediaType:   view.MediaType,
			ContentID:   contentID,
			Disposition: view.Disposition,
			ScanStatus:  view.ScanStatus,
			SizeBytes:   view.SizeBytes,
			SHA256:      view.SHA256[:],
		}},
	})
	if err != nil || len(decision.Attachments) != 1 {
		return false
	}
	normalized := decision.Attachments[0]
	return normalized.Filename == view.Filename && normalized.MediaType == view.MediaType && normalized.ContentID == contentID && normalized.Disposition == view.Disposition && normalized.ScanStatus == view.ScanStatus && normalized.SizeBytes == view.SizeBytes && len(normalized.SHA256) == sha256.Size
}

func mailboxAttachmentConfirmViewCopy(view MailboxAttachmentView) MailboxAttachmentView {
	result := view
	if view.ContentID != nil {
		contentID := *view.ContentID
		result.ContentID = &contentID
	}
	if view.SHA256 != nil {
		digest := *view.SHA256
		result.SHA256 = &digest
	}
	return result
}

func mailboxAttachmentConfirmStreamSHA256(objectStore MailboxObjectStore, ctx context.Context, object MailboxObjectLocator, expectedSizeBytes int64) (digest [sha256.Size]byte, sizeBytes int64, err error) {
	if mailboxAttachmentStagingDependencyNil(objectStore) || ctx == nil || !validMailboxObjectLocator(object) || expectedSizeBytes < 1 || expectedSizeBytes > MaxEngineAttachmentBytes {
		return digest, 0, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}

	stream, openErr := objectStore.Open(ctx, MailboxObjectReadRequest{Object: object})
	if openErr != nil {
		return digest, 0, mailboxAttachmentConfirmObjectStoreError(openErr)
	}
	if stream.Body == nil {
		return digest, 0, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	defer func() {
		if closeErr := stream.Body.Close(); closeErr != nil {
			err = mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
	}()
	if stream.ContentLength != expectedSizeBytes {
		return digest, 0, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorConfirmMismatch)
	}

	hash := sha256.New()
	var buffer [mailboxAttachmentConfirmBufferBytes]byte
	for {
		read, readErr := stream.Body.Read(buffer[:])
		if read < 0 || read > len(buffer) {
			return digest, sizeBytes, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
		if read > 0 {
			if int64(read) > expectedSizeBytes-sizeBytes {
				return digest, sizeBytes, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorConfirmMismatch)
			}
			if _, writeErr := hash.Write(buffer[:read]); writeErr != nil {
				return digest, sizeBytes, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
			}
			sizeBytes += int64(read)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil || read == 0 {
			return digest, sizeBytes, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
		}
	}
	if sizeBytes != expectedSizeBytes {
		return digest, sizeBytes, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorConfirmMismatch)
	}
	copy(digest[:], hash.Sum(nil))
	return digest, sizeBytes, nil
}

func mailboxAttachmentConfirmObjectStoreError(err error) error {
	kind, ok := MailboxObjectStoreErrorKindOf(err)
	if ok && kind == MailboxObjectStoreErrorNotFound {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorNotFound)
	}
	return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
}

func mailboxAttachmentConfirmOwnerError(err error) error {
	var serviceError *MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
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

func mailboxAttachmentConfirmStagingError(err error) error {
	var stagingError *MailboxAttachmentStagingError
	if !errors.As(err, &stagingError) || stagingError == nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	return mailboxAttachmentStagingError(stagingError.Kind())
}
