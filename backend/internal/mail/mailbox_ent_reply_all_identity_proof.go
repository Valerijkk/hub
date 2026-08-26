package mail

import (
	"context"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailidentity"
)

var _ MailboxReplyAllIdentityProofResolver = (*MailboxEntRepository)(nil)

// ResolveMailboxReplyAllIdentityProof locks and proves the complete bounded set of active owner addresses.
func (repository *MailboxEntRepository) ResolveMailboxReplyAllIdentityProof(
	ctx context.Context,
	request mailboxReplyAllIdentityProofRequest,
) (mailboxReplyAllIdentityProof, error) {
	if repository == nil || repository.client == nil || ctx == nil || request.ActorID == uuid.Nil ||
		request.OrganizationID == uuid.Nil || request.AccountID == uuid.Nil || request.SelectedIdentityID == uuid.Nil {
		return mailboxReplyAllIdentityProof{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if ctx.Err() != nil {
		return mailboxReplyAllIdentityProof{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}

	var proof mailboxReplyAllIdentityProof
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().
			Where(
				mailaccount.IDEQ(request.AccountID),
				mailaccount.UserIDEQ(request.ActorID),
				mailaccount.OrganizationIDEQ(request.OrganizationID),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorNotFound)
			}
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if account == nil || account.ID != request.AccountID || account.UserID != request.ActorID || account.OrganizationID != request.OrganizationID {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}
		if err := mailboxEntDraftReadLockAccount(ctx, transaction, request.AccountID); err != nil {
			return err
		}
		if account.ProjectionState != mailaccount.ProjectionStateReady {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}

		selected, err := transaction.MailIdentity.Query().
			Where(
				mailidentity.IDEQ(request.SelectedIdentityID),
				mailidentity.MailAccountIDEQ(request.AccountID),
				mailidentity.StatusEQ(mailidentity.StatusActive),
				mailidentity.SendAllowedEQ(true),
				mailidentity.KindIn(mailidentity.KindPrimary, mailidentity.KindAlias),
				mailidentity.HasAccountWith(
					mailaccount.IDEQ(request.AccountID),
					mailaccount.UserIDEQ(request.ActorID),
					mailaccount.OrganizationIDEQ(request.OrganizationID),
				),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
			}
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		selectedAddress, err := mailboxEntReplyAllIdentityAddress(selected)
		if err != nil || selected == nil || selected.ID != request.SelectedIdentityID || selected.MailAccountID != request.AccountID ||
			selected.Status != mailidentity.StatusActive || !selected.SendAllowed ||
			selected.Kind != mailidentity.KindPrimary && selected.Kind != mailidentity.KindAlias {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}

		rows, err := transaction.MailIdentity.Query().
			Where(
				mailidentity.MailAccountIDEQ(request.AccountID),
				mailidentity.StatusEQ(mailidentity.StatusActive),
				mailidentity.HasAccountWith(
					mailaccount.IDEQ(request.AccountID),
					mailaccount.UserIDEQ(request.ActorID),
					mailaccount.OrganizationIDEQ(request.OrganizationID),
				),
			).
			Order(ent.Asc(mailidentity.FieldNormalizedAddress), ent.Asc(mailidentity.FieldID)).
			Limit(maxMailboxIdentityLimit + 1).
			ForUpdate().
			All(ctx)
		if err != nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if len(rows) == 0 {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}
		if len(rows) > maxMailboxIdentityLimit {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}

		owned := make([]EngineAddress, 0, len(rows))
		seen := make(map[string]struct{}, len(rows))
		selectedFound := false
		for _, row := range rows {
			if row == nil || row.ID == uuid.Nil || row.MailAccountID != request.AccountID || row.Status != mailidentity.StatusActive {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
			}
			switch row.Kind {
			case mailidentity.KindPrimary, mailidentity.KindAlias:
			case mailidentity.KindGroup:
				if row.SendAllowed {
					return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
				}
			default:
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
			}
			address, err := mailboxEntReplyAllIdentityAddress(row)
			if err != nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
			}
			if _, duplicate := seen[address.Email]; duplicate {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
			}
			seen[address.Email] = struct{}{}
			owned = append(owned, address)
			if row.ID == request.SelectedIdentityID {
				selectedFound = true
			}
		}
		if !selectedFound {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}
		if _, present := seen[selectedAddress.Email]; !present {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}

		proof = mailboxReplyAllIdentityProof{
			actorID: request.ActorID, organizationID: request.OrganizationID, accountID: request.AccountID,
			selectedIdentityID: request.SelectedIdentityID, selected: selectedAddress, owned: owned,
		}
		return nil
	})
	if err != nil {
		return mailboxReplyAllIdentityProof{}, mailboxEntDraftRepositoryError(err)
	}
	return proof, nil
}

func mailboxEntReplyAllIdentityAddress(identity *ent.MailIdentity) (EngineAddress, error) {
	if identity == nil {
		return EngineAddress{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	name := ""
	if identity.DisplayName != nil {
		name = *identity.DisplayName
	}
	address, err := NormalizeAddress(EngineAddress{Email: identity.NormalizedAddress, Name: name})
	if err != nil || address.Email != identity.NormalizedAddress {
		return EngineAddress{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	return address, nil
}
