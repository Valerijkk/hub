package mail

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PresignMailboxAttachmentDownload returns one bounded download URL for a clean, linked attachment.
func (service *MailboxAttachmentStagingService) PresignMailboxAttachmentDownload(ctx context.Context, actorID, attachmentID uuid.UUID) (MailboxAttachmentDownloadResult, error) {
	if !mailboxAttachmentDownloadServiceValid(service) || ctx == nil || actorID == uuid.Nil || attachmentID == uuid.Nil {
		return MailboxAttachmentDownloadResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}

	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxAttachmentDownloadResult{}, mailboxAttachmentPresignOwnerError(err)
	}
	if owner.State() == MailboxOwnerDeleting {
		return MailboxAttachmentDownloadResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorDeleting)
	}
	accountID, hasAccount := owner.AccountID()
	if owner.State() != MailboxOwnerActive || owner.OrganizationID() == uuid.Nil || !hasAccount || accountID == uuid.Nil {
		return MailboxAttachmentDownloadResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}

	resolved, err := service.downloads.ResolveMailboxAttachmentDownload(ctx, MailboxAttachmentDownloadRepositoryRequest{
		ActorID:      actorID,
		AttachmentID: attachmentID,
	})
	if err != nil {
		return MailboxAttachmentDownloadResult{}, mailboxAttachmentPresignStagingError(err)
	}
	if !mailboxAttachmentDownloadRepositoryResultValid(resolved) {
		return MailboxAttachmentDownloadResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}

	expiresAt, ok := mailboxAttachmentDownloadExpiry(resolved.IssuedAt)
	if !ok {
		return MailboxAttachmentDownloadResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	presignRequest := MailboxObjectPresignGetRequest{
		Object:    resolved.Object,
		IssuedAt:  resolved.IssuedAt,
		ExpiresAt: expiresAt,
	}
	if !validMailboxObjectPresignGetRequest(presignRequest) {
		return MailboxAttachmentDownloadResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	presigned, err := service.objectStore.PresignGet(ctx, presignRequest)
	if err != nil {
		return MailboxAttachmentDownloadResult{}, mailboxAttachmentDownloadObjectStoreError(err)
	}
	if !validMailboxObjectPresignGetResult(presignRequest, presigned) {
		return MailboxAttachmentDownloadResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	return MailboxAttachmentDownloadResult{URL: presigned.URL, ExpiresAt: presigned.ExpiresAt}, nil
}

func mailboxAttachmentDownloadServiceValid(service *MailboxAttachmentStagingService) bool {
	return service != nil && !mailboxAttachmentStagingDependencyNil(service.owners) && !mailboxAttachmentStagingDependencyNil(service.downloads) && !mailboxAttachmentStagingDependencyNil(service.objectStore)
}

func mailboxAttachmentDownloadRepositoryResultValid(result MailboxAttachmentDownloadRepositoryResult) bool {
	return validMailboxObjectLocator(result.Object) && mailboxEntValidTime(result.IssuedAt)
}

func mailboxAttachmentDownloadExpiry(issuedAt time.Time) (time.Time, bool) {
	expiresAt, ok := mailboxEntCheckedTimeAdd(issuedAt, mailboxObjectStoreMaximumGetLifetime)
	if !ok || !validMailboxObjectStoreLifetime(issuedAt, expiresAt, mailboxObjectStoreMaximumGetLifetime) {
		return time.Time{}, false
	}
	return expiresAt, true
}

func mailboxAttachmentDownloadObjectStoreError(err error) error {
	if kind, ok := MailboxObjectStoreErrorKindOf(err); ok && kind == MailboxObjectStoreErrorNotFound {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorNotFound)
	}
	return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
}
