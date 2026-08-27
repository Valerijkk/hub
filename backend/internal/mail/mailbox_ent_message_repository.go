package mail

import (
	"context"
	"encoding/hex"
	"errors"
	"sort"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailthread"
	"github.com/valerii/hub/backend/ent/schema"
)

var _ MailboxMessageRepository = (*MailboxEntRepository)(nil)

// ResolveMailboxMessage seals one live, owner-scoped local message before any engine I/O.
func (repository *MailboxEntRepository) ResolveMailboxMessage(ctx context.Context, scope MailboxMessageReadScope, messageID uuid.UUID) (MailboxMessageSnapshot, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxMessageReadScopeValid(scope) || messageID == uuid.Nil {
		return MailboxMessageSnapshot{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}

	var snapshot MailboxMessageSnapshot
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().
			Where(
				mailaccount.IDEQ(scope.AccountID),
				mailaccount.UserIDEQ(scope.ActorID),
				mailaccount.OrganizationIDEQ(scope.OrganizationID),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntMessageReadSafeError(err)
		}
		if account.ID != scope.AccountID || account.UserID != scope.ActorID || account.OrganizationID != scope.OrganizationID {
			return mailboxMessageReadError(MailboxMessageReadErrorInvalid)
		}
		switch account.Status {
		case mailaccount.StatusDeleting:
			return mailboxMessageReadError(MailboxMessageReadErrorDeleting)
		case mailaccount.StatusProvisioning, mailaccount.StatusSuspended, mailaccount.StatusError:
			return mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
		case mailaccount.StatusActive:
		default:
			return mailboxMessageReadError(MailboxMessageReadErrorInvalid)
		}
		if !account.DesiredEnabled || account.EngineAccountID == nil || *account.EngineAccountID != scope.EngineAccountID ||
			!syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) || MailboxProjectionState(account.ProjectionState) != MailboxProjectionReady ||
			account.LastSyncedAt == nil || !mailboxEntValidTime(*account.LastSyncedAt) || account.SyncEpoch < 0 || account.MutationEpoch < 0 {
			return mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
		}

		row, err := transaction.MailMessage.Query().
			Where(
				mailmessage.IDEQ(messageID),
				mailmessage.MailAccountIDEQ(scope.AccountID),
				mailmessage.DeletedAtIsNil(),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntMessageReadSafeError(err)
		}
		if !mailboxEntApplyMessageRowValid(row, scope.AccountID) || row.ID != messageID || row.DeletedAt != nil {
			return mailboxMessageReadError(MailboxMessageReadErrorNotFound)
		}
		thread, err := transaction.MailThread.Query().
			Where(
				mailthread.IDEQ(row.MailThreadID),
				mailthread.MailAccountIDEQ(scope.AccountID),
				mailthread.DeletedAtIsNil(),
				mailthread.HasAccountWith(
					mailaccount.IDEQ(scope.AccountID),
					mailaccount.UserIDEQ(scope.ActorID),
					mailaccount.OrganizationIDEQ(scope.OrganizationID),
				),
			).
			ForUpdate().
			Only(ctx)
		if err != nil || thread.ID != row.MailThreadID {
			if err != nil {
				return mailboxEntMessageReadSafeError(err)
			}
			return mailboxMessageReadError(MailboxMessageReadErrorNotFound)
		}

		messages, err := mailboxEntThreadDetailProjectionMessages(ctx, transaction, MailboxFolderProjectionScope{
			ActorID: scope.ActorID, OrganizationID: scope.OrganizationID, AccountID: scope.AccountID,
		}, row.MailThreadID, []*ent.MailMessage{row})
		if err != nil || len(messages) != 1 {
			if err != nil {
				return mailboxEntMessageReadSafeError(err)
			}
			return mailboxMessageReadError(MailboxMessageReadErrorInvalid)
		}

		engineFolderIDs, err := mailboxEntMessageEngineFolderIDs(ctx, transaction, scope.AccountID, messages[0].FolderIDs)
		if err != nil {
			return err
		}
		attachments, err := mailboxEntMessageAttachments(ctx, transaction, messageID)
		if err != nil {
			return err
		}

		snapshot = MailboxMessageSnapshot{
			summary:                 messages[0],
			attachments:             attachments,
			engineAccountID:         scope.EngineAccountID,
			engineMessageID:         row.EngineMessageID,
			engineVersion:           row.EngineVersion,
			internetMessageID:       mailboxEntMessageOptionalString(row.InternetMessageID),
			inReplyTo:               mailboxEntMessageOptionalString(row.InReplyTo),
			references:              append([]string(nil), row.References...),
			from:                    EngineAddress{Email: row.From.Email, Name: row.From.Name},
			to:                      mailboxEntMessageEngineAddresses(row.To),
			cc:                      mailboxEntMessageEngineAddresses(row.Cc),
			bcc:                     mailboxEntMessageEngineAddresses(row.Bcc),
			replyTo:                 mailboxEntMessageEngineAddresses(row.ReplyTo),
			projectionOrigin:        EngineProjectionOrigin(row.ProjectionOrigin),
			authenticationClass:     MailAuthenticationClass(row.AuthenticationClass),
			calendarPartLocatorHash: append([]byte(nil), row.CalendarPartLocatorHash...),
			engineFolderIDs:         engineFolderIDs,
			projectionState:         MailboxProjectionState(account.ProjectionState),
			projectionAsOf:          mailboxFolderProjectionTimePointer(account.LastSyncedAt),
			syncEpoch:               account.SyncEpoch,
			mutationEpoch:           account.MutationEpoch,
			updatedAt:               row.UpdatedAt.UTC(),
		}
		if !mailboxMessageSnapshotValid(snapshot) {
			return mailboxMessageReadError(MailboxMessageReadErrorInvalid)
		}
		return nil
	})
	if err != nil {
		return MailboxMessageSnapshot{}, mailboxEntMessageReadSafeError(err)
	}
	return mailboxMessageSnapshotCopy(snapshot), nil
}

// RecheckMailboxMessage rejects any local ownership, projection, version, link, or scan-state drift after engine I/O.
func (repository *MailboxEntRepository) RecheckMailboxMessage(ctx context.Context, scope MailboxMessageReadScope, expected MailboxMessageSnapshot) error {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxMessageReadScopeValid(scope) || !mailboxMessageSnapshotValid(expected) || expected.summary.ID == uuid.Nil || expected.engineAccountID != scope.EngineAccountID {
		return mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	current, err := repository.ResolveMailboxMessage(ctx, scope, expected.summary.ID)
	if err != nil {
		return mailboxEntMessageReadSafeError(err)
	}
	if !mailboxMessageSnapshotEqual(expected, current) {
		return mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
	}
	return nil
}

func mailboxEntMessageEngineFolderIDs(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID, folderIDs []uuid.UUID) ([]string, error) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || len(folderIDs) < 1 || len(folderIDs) > MaxEngineFolderItems {
		return nil, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	rows, err := transaction.MailFolder.Query().
		Where(
			mailfolder.IDIn(folderIDs...),
			mailfolder.MailAccountIDEQ(accountID),
			mailfolder.ControlStateEQ(mailfolder.ControlStateActive),
			mailfolder.EngineFolderIDNotNil(),
			mailfolder.DeletedAtIsNil(),
		).
		Limit(len(folderIDs) + 1).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, mailboxEntMessageReadSafeError(err)
	}
	if len(rows) != len(folderIDs) {
		return nil, mailboxMessageReadError(MailboxMessageReadErrorNotFound)
	}
	seen := make(map[uuid.UUID]struct{}, len(rows))
	engineFolderIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		state, valid := mailboxEntApplyFolderRowState(row, accountID)
		if row == nil || row.EngineFolderID == nil || !valid || state != mailboxEntApplyFolderActive || row.ControlState != mailfolder.ControlStateActive || row.DeletedAt != nil || !syncValidOpaque(*row.EngineFolderID, MaxEngineOpaqueIDBytes) {
			return nil, mailboxMessageReadError(MailboxMessageReadErrorNotFound)
		}
		if _, duplicate := seen[row.ID]; duplicate {
			return nil, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
		}
		seen[row.ID] = struct{}{}
		engineFolderIDs = append(engineFolderIDs, *row.EngineFolderID)
	}
	for _, folderID := range folderIDs {
		if _, found := seen[folderID]; !found {
			return nil, mailboxMessageReadError(MailboxMessageReadErrorNotFound)
		}
	}
	sort.Strings(engineFolderIDs)
	if !validOpaqueIDs(engineFolderIDs) {
		return nil, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	return engineFolderIDs, nil
}

func mailboxEntMessageAttachments(ctx context.Context, transaction *ent.Tx, messageID uuid.UUID) ([]MailboxMessageAttachment, error) {
	if ctx == nil || transaction == nil || messageID == uuid.Nil {
		return nil, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	rows, err := transaction.MailAttachment.Query().
		Where(mailattachment.MailMessageIDEQ(messageID)).
		Order(ent.Asc(mailattachment.FieldID)).
		Limit(MaxEngineAttachmentItems + 1).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, mailboxEntMessageReadSafeError(err)
	}
	if len(rows) > MaxEngineAttachmentItems {
		return nil, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	attachments := make([]MailboxMessageAttachment, len(rows))
	for index, row := range rows {
		providerBacked, err := mailboxEntApplyAttachmentRowValid(row, messageID)
		if err != nil || !providerBacked {
			return nil, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
		}
		if row.ScanStatus != mailattachment.ScanStatus("clean") {
			return nil, mailboxMessageReadError(MailboxMessageReadErrorUnprocessable)
		}
		if row.EngineBlobID == nil || len(row.Sha256) != 32 || !mailboxEntValidTime(row.CreatedAt) || !mailboxEntValidTime(row.UpdatedAt) {
			return nil, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
		}
		attachments[index] = MailboxMessageAttachment{
			ID: row.ID, Filename: row.Filename, MediaType: row.MediaType, SizeBytes: row.SizeBytes,
			Disposition: string(row.Disposition), ContentID: mailboxFolderProjectionStringPointer(row.ContentID),
			ScanStatus: "clean", SHA256: hex.EncodeToString(row.Sha256), engineBlobID: *row.EngineBlobID,
			createdAt: row.CreatedAt.UTC(), updatedAt: row.UpdatedAt.UTC(),
		}
	}
	return attachments, nil
}

func mailboxEntMessageOptionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func mailboxEntMessageEngineAddresses(values schema.MailStoredAddresses) []EngineAddress {
	result := make([]EngineAddress, len(values))
	for index, value := range values {
		result[index] = EngineAddress{Email: value.Email, Name: value.Name}
	}
	return result
}

func mailboxEntMessageReadSafeError(err error) error {
	if kind, ok := mailboxMessageReadErrorKindOf(err); ok {
		return mailboxMessageReadError(kind)
	}
	var serviceError *MailboxServiceError
	if errors.As(err, &serviceError) && serviceError != nil {
		switch serviceError.Kind {
		case MailboxServiceErrorNotFound:
			return mailboxMessageReadError(MailboxMessageReadErrorNotFound)
		case MailboxServiceErrorDeleting:
			return mailboxMessageReadError(MailboxMessageReadErrorDeleting)
		case MailboxServiceErrorUnavailable:
			return mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
		case MailboxServiceErrorInvalid:
			return mailboxMessageReadError(MailboxMessageReadErrorInvalid)
		}
	}
	if ent.IsNotFound(err) {
		return mailboxMessageReadError(MailboxMessageReadErrorNotFound)
	}
	if ent.IsConstraintError(err) {
		return mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	return mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
}
