package mail

import (
	"context"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailfolder"
)

var _ MailboxFolderProjectionRepository = (*MailboxEntRepository)(nil)

// ListMailboxFolderProjection reads only the durable Ent projection in the complete authenticated scope.
func (repository *MailboxEntRepository) ListMailboxFolderProjection(ctx context.Context, scope MailboxFolderProjectionScope) (MailboxFolderProjection, error) {
	if repository == nil || repository.client == nil || ctx == nil || scope.ActorID == uuid.Nil || scope.OrganizationID == uuid.Nil || scope.AccountID == uuid.Nil {
		return MailboxFolderProjection{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var projection MailboxFolderProjection
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
			return mailboxEntSafeError(err)
		}
		if account.ID != scope.AccountID || account.UserID != scope.ActorID || account.OrganizationID != scope.OrganizationID {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		switch account.Status {
		case mailaccount.StatusActive:
			if !validMailboxProjectionState(MailboxProjectionState(account.ProjectionState)) || (account.LastSyncedAt != nil && !mailboxEntValidTime(*account.LastSyncedAt)) {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
		case mailaccount.StatusDeleting:
			return mailboxServiceError(MailboxServiceErrorDeleting)
		case mailaccount.StatusProvisioning, mailaccount.StatusSuspended, mailaccount.StatusError:
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		default:
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		rows, err := transaction.MailFolder.Query().
			Where(
				mailfolder.MailAccountIDEQ(scope.AccountID),
				mailfolder.DeletedAtIsNil(),
				mailfolder.HasAccountWith(
					mailaccount.IDEQ(scope.AccountID),
					mailaccount.UserIDEQ(scope.ActorID),
					mailaccount.OrganizationIDEQ(scope.OrganizationID),
				),
			).
			Order(
				ent.Asc(mailfolder.FieldRole),
				ent.Asc(mailfolder.FieldSortOrder),
				ent.Asc(mailfolder.FieldNormalizedName),
				ent.Asc(mailfolder.FieldID),
			).
			Limit(maxMailboxFolderProjectionItems + 1).
			All(ctx)
		if err != nil {
			return mailboxEntSafeError(err)
		}
		if len(rows) > maxMailboxFolderProjectionItems {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		items := make([]MailboxFolderProjectionItem, len(rows))
		for index, row := range rows {
			item, err := mailboxEntFolderProjectionItem(scope, row)
			if err != nil {
				return err
			}
			items[index] = item
		}

		projection = MailboxFolderProjection{
			Items:           items,
			ProjectionState: MailboxProjectionState(account.ProjectionState),
			ProjectionAsOf:  mailboxFolderProjectionTimePointer(account.LastSyncedAt),
		}
		sortMailboxFolderProjectionItems(projection.Items)
		return nil
	})
	if err != nil {
		return MailboxFolderProjection{}, mailboxEntSafeError(err)
	}

	return mailboxFolderProjectionCopy(projection), nil
}

func mailboxEntFolderProjectionItem(scope MailboxFolderProjectionScope, row *ent.MailFolder) (MailboxFolderProjectionItem, error) {
	if row == nil || scope.AccountID == uuid.Nil || row.ID == uuid.Nil || row.MailAccountID != scope.AccountID || row.DeletedAt != nil {
		return MailboxFolderProjectionItem{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if _, valid := mailboxEntApplyFolderRowState(row, scope.AccountID); !valid || !mailboxEntValidTime(row.CreatedAt) || !mailboxEntValidTime(row.UpdatedAt) || !validMailboxFolderProjectionColor(row.Color) {
		return MailboxFolderProjectionItem{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var role *string
	if row.Role != nil {
		roleValue := string(*row.Role)
		role = &roleValue
	}

	return MailboxFolderProjectionItem{
		ID:             row.ID,
		ParentID:       mailboxFolderProjectionUUIDPointer(row.ParentID),
		Role:           role,
		Name:           row.Name,
		Color:          mailboxFolderProjectionStringPointer(row.Color),
		SortOrder:      row.SortOrder,
		UnreadCount:    row.UnreadCount,
		TotalCount:     row.TotalCount,
		State:          string(row.ControlState),
		Version:        row.Version,
		normalizedName: row.NormalizedName,
	}, nil
}
