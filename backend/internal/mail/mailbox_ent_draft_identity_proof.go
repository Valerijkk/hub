package mail

import (
	"context"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailidentity"
)

var _ MailboxDraftIdentityProofResolver = (*MailboxEntRepository)(nil)

// ResolveMailboxDraftIdentityProof locks and proves the account-scoped identity for a draft request.
func (repository *MailboxEntRepository) ResolveMailboxDraftIdentityProof(ctx context.Context, request mailboxDraftIdentityProofRequest) (mailboxDraftIdentityProof, error) {
	if repository == nil || repository.client == nil || ctx == nil || request.AccountID == uuid.Nil || request.IdentityID == uuid.Nil {
		return mailboxDraftIdentityProof{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}

	var result mailboxDraftIdentityProof
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		if err := mailboxEntDraftReadLockAccount(ctx, transaction, request.AccountID); err != nil {
			return err
		}

		identity, err := transaction.MailIdentity.Query().
			Where(
				mailidentity.IDEQ(request.IdentityID),
				mailidentity.MailAccountIDEQ(request.AccountID),
				mailidentity.StatusEQ(mailidentity.Status("active")),
				mailidentity.SendAllowedEQ(true),
				mailidentity.KindIn(mailidentity.Kind("primary"), mailidentity.Kind("alias")),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
			}
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if identity == nil || identity.ID != request.IdentityID || identity.MailAccountID != request.AccountID || !identity.SendAllowed || identity.Status != mailidentity.Status("active") || identity.Kind != mailidentity.Kind("primary") && identity.Kind != mailidentity.Kind("alias") {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}

		name := ""
		if identity.DisplayName != nil {
			name = *identity.DisplayName
		}
		from, err := NormalizeAddress(EngineAddress{Email: identity.NormalizedAddress, Name: name})
		if err != nil || from.Email != identity.NormalizedAddress {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}

		result = mailboxDraftIdentityProof{
			accountID:  request.AccountID,
			identityID: request.IdentityID,
			from:       EngineAddress{Email: from.Email, Name: from.Name},
		}
		return nil
	})
	if err != nil {
		return mailboxDraftIdentityProof{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}
