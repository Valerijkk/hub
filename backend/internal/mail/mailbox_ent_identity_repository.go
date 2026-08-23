package mail

import (
	"context"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/mailusersettings"
	"github.com/valerii/hub/backend/ent/predicate"
)

var _ MailboxIdentityRepository = (*MailboxEntRepository)(nil)

// ListMailboxIdentities reads a stable owner-scoped page of authorized From identities.
func (repository *MailboxEntRepository) ListMailboxIdentities(ctx context.Context, scope MailboxIdentityScope, query MailboxIdentityListQuery) (MailboxIdentityList, error) {
	if repository == nil || repository.client == nil || ctx == nil || scope.ActorID == uuid.Nil || scope.OrganizationID == uuid.Nil || scope.AccountID == uuid.Nil || !mailboxIdentityListQueryValid(query) {
		return MailboxIdentityList{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	page := MailboxIdentityList{Items: []MailIdentity{}}
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
			if account.ProjectionState != mailaccount.ProjectionStateReady || account.EngineAccountID == nil || !validMailboxOwnerOpaque(*account.EngineAccountID) {
				return mailboxServiceError(MailboxServiceErrorUnavailable)
			}
		case mailaccount.StatusDeleting:
			return mailboxServiceError(MailboxServiceErrorDeleting)
		case mailaccount.StatusProvisioning, mailaccount.StatusSuspended, mailaccount.StatusError:
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		default:
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		settings, err := transaction.MailUserSettings.Query().
			Where(
				mailusersettings.MailAccountIDEQ(scope.AccountID),
				mailusersettings.HasAccountWith(
					mailaccount.IDEQ(scope.AccountID),
					mailaccount.UserIDEQ(scope.ActorID),
					mailaccount.OrganizationIDEQ(scope.OrganizationID),
				),
				mailusersettings.HasDefaultIdentityWith(
					mailidentity.MailAccountIDEQ(scope.AccountID),
				),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSafeError(err)
		}
		if settings.MailAccountID != scope.AccountID || settings.DefaultIdentityID == uuid.Nil {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		identityQuery := transaction.MailIdentity.Query().
			Where(
				mailidentity.MailAccountIDEQ(scope.AccountID),
				mailidentity.StatusEQ(mailidentity.StatusActive),
				mailidentity.SendAllowedEQ(true),
				mailidentity.KindIn(mailidentity.KindPrimary, mailidentity.KindAlias),
				mailidentity.HasAccountWith(
					mailaccount.IDEQ(scope.AccountID),
					mailaccount.UserIDEQ(scope.ActorID),
					mailaccount.OrganizationIDEQ(scope.OrganizationID),
				),
			)
		if query.Cursor != nil {
			identityQuery = identityQuery.Where(mailboxIdentityAfterCursor(*query.Cursor))
		}
		rows, err := identityQuery.
			Order(
				ent.Asc(mailidentity.FieldNormalizedAddress),
				ent.Asc(mailidentity.FieldID),
			).
			Limit(query.Limit + 1).
			All(ctx)
		if err != nil {
			return mailboxEntSafeError(err)
		}

		hasNext := len(rows) > query.Limit
		if hasNext {
			rows = rows[:query.Limit]
		}
		items := make([]MailIdentity, len(rows))
		for index, row := range rows {
			item, err := mailboxEntIdentityItem(scope.AccountID, settings.DefaultIdentityID, row)
			if err != nil {
				return err
			}
			items[index] = item
		}
		page.Items = items
		if hasNext {
			last := page.Items[len(page.Items)-1]
			cursor, err := mailboxIdentityEncodeCursor(MailboxIdentityCursor{
				accountID:         scope.AccountID,
				normalizedAddress: last.normalizedAddress,
				id:                last.ID,
			})
			if err != nil {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			page.NextCursor = &cursor
		}
		return nil
	})
	if err != nil {
		return MailboxIdentityList{}, mailboxEntSafeError(err)
	}
	if !mailboxIdentityListValid(page, query, scope.AccountID) {
		return MailboxIdentityList{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return mailboxIdentityListCopy(page), nil
}

func mailboxIdentityAfterCursor(cursor MailboxIdentityCursor) predicate.MailIdentity {
	return mailidentity.Or(
		mailidentity.NormalizedAddressGT(cursor.normalizedAddress),
		mailidentity.And(
			mailidentity.NormalizedAddressEQ(cursor.normalizedAddress),
			mailidentity.IDGT(cursor.id),
		),
	)
}

func mailboxEntIdentityItem(accountID, defaultIdentityID uuid.UUID, row *ent.MailIdentity) (MailIdentity, error) {
	if accountID == uuid.Nil || defaultIdentityID == uuid.Nil || row == nil || row.ID == uuid.Nil || row.MailAccountID != accountID || row.Status != mailidentity.StatusActive || !row.SendAllowed || row.Kind == mailidentity.KindGroup {
		return MailIdentity{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	displayName := ""
	if row.DisplayName != nil {
		displayName = *row.DisplayName
	}
	item := MailIdentity{
		ID:                row.ID,
		Address:           row.NormalizedAddress,
		DisplayName:       displayName,
		Kind:              string(row.Kind),
		SendAllowed:       row.SendAllowed,
		IsDefault:         row.ID == defaultIdentityID,
		Status:            string(row.Status),
		normalizedAddress: row.NormalizedAddress,
	}
	if !validMailboxIdentity(item) {
		return MailIdentity{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return item, nil
}
