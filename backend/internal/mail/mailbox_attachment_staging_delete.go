package mail

import (
	"context"

	"github.com/google/uuid"
)

// MailboxAttachmentAbandonServiceAPI is the domain boundary for idempotently abandoning one staged attachment.
type MailboxAttachmentAbandonServiceAPI interface {
	AbandonMailboxAttachment(context.Context, uuid.UUID, uuid.UUID) error
}

// MailboxAttachmentAbandonService keeps attachment abandonment separate from upload, confirmation, and object-store I/O.
type MailboxAttachmentAbandonService struct {
	owners   mailboxOwnerResolver
	abandons MailboxAttachmentAbandonRepository
}

func NewMailboxAttachmentAbandonService(
	owners mailboxOwnerResolver,
	abandons MailboxAttachmentAbandonRepository,
) (*MailboxAttachmentAbandonService, error) {
	if mailboxAttachmentStagingDependencyNil(owners) || mailboxAttachmentStagingDependencyNil(abandons) {
		return nil, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}

	return &MailboxAttachmentAbandonService{owners: owners, abandons: abandons}, nil
}

// AbandonMailboxAttachment removes an unlinked attachment and leaves its object for the deletion worker.
// Missing and foreign attachments are deliberately indistinguishable from an already-abandoned attachment.
func (service *MailboxAttachmentAbandonService) AbandonMailboxAttachment(ctx context.Context, actorID, attachmentID uuid.UUID) error {
	if !mailboxAttachmentAbandonServiceValid(service) || ctx == nil || actorID == uuid.Nil || attachmentID == uuid.Nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}

	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return mailboxAttachmentAbandonOwnerError(err)
	}
	if owner.State() == MailboxOwnerDeleting {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorDeleting)
	}
	accountID, hasAccount := owner.AccountID()
	if owner.State() != MailboxOwnerActive || owner.OrganizationID() == uuid.Nil || !hasAccount || accountID == uuid.Nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}

	result, err := service.abandons.AbandonMailboxAttachment(ctx, MailboxAttachmentAbandonRepositoryRequest{
		ActorID:                  actorID,
		OrganizationID:           owner.OrganizationID(),
		AccountID:                accountID,
		AttachmentID:             attachmentID,
		MinimumObjectDeleteDelay: mailboxObjectStoreMaximumGetLifetime,
	})
	if err != nil {
		return mailboxAttachmentAbandonRepositoryError(err)
	}
	if !mailboxAttachmentAbandonRepositoryResultValid(result) {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}

	switch result.Kind {
	case MailboxAttachmentAbandonRepositoryResultAbandoned, MailboxAttachmentAbandonRepositoryResultAbsent:
		return nil
	case MailboxAttachmentAbandonRepositoryResultConflict:
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorStateConflict)
	case MailboxAttachmentAbandonRepositoryResultDeleting:
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorDeleting)
	default:
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
}

func mailboxAttachmentAbandonServiceValid(service *MailboxAttachmentAbandonService) bool {
	return service != nil && !mailboxAttachmentStagingDependencyNil(service.owners) && !mailboxAttachmentStagingDependencyNil(service.abandons)
}

func mailboxAttachmentAbandonRepositoryResultValid(result MailboxAttachmentAbandonRepositoryResult) bool {
	switch result.Kind {
	case MailboxAttachmentAbandonRepositoryResultAbandoned,
		MailboxAttachmentAbandonRepositoryResultAbsent,
		MailboxAttachmentAbandonRepositoryResultDeleting:
		return result.Conflict == ""
	case MailboxAttachmentAbandonRepositoryResultConflict:
		switch result.Conflict {
		case MailboxAttachmentAbandonRepositoryConflictMessage,
			MailboxAttachmentAbandonRepositoryConflictDraft,
			MailboxAttachmentAbandonRepositoryConflictPendingDraft,
			MailboxAttachmentAbandonRepositoryConflictStorage:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func mailboxAttachmentAbandonOwnerError(err error) error {
	return mailboxAttachmentPresignOwnerError(err)
}

func mailboxAttachmentAbandonRepositoryError(err error) error {
	kind, ok := MailboxAttachmentStagingErrorKindOf(err)
	if !ok {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}

	switch kind {
	case MailboxAttachmentStagingErrorInvalid,
		MailboxAttachmentStagingErrorDeleting,
		MailboxAttachmentStagingErrorUnavailable,
		MailboxAttachmentStagingErrorStateConflict:
		return mailboxAttachmentStagingError(kind)
	default:
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
}
