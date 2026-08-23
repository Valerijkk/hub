package mail

import (
	"context"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailoauthcredential"
)

// LoadMailboxOAuthBearerSnapshot resolves one exact provider account to a
// locked, fully scoped credential snapshot. Ambiguity and relational drift are
// closed as unavailable rather than selecting an arbitrary credential.
func (repository *MailboxEntRepository) LoadMailboxOAuthBearerSnapshot(ctx context.Context, engineAccountID string) (MailboxOAuthCredentialSnapshot, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validAccountID(engineAccountID) {
		return MailboxOAuthCredentialSnapshot{}, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	var snapshot MailboxOAuthCredentialSnapshot
	err := repository.withMailboxEntOAuthTransaction(ctx, func(transaction *ent.Tx) error {
		accounts, err := transaction.MailAccount.Query().
			Where(
				mailaccount.EngineAccountIDEQ(engineAccountID),
				mailaccount.StatusEQ(mailaccount.StatusActive),
				mailaccount.DesiredEnabledEQ(true),
			).
			Order(ent.Asc(mailaccount.FieldID)).
			ForUpdate().
			All(ctx)
		if err != nil {
			return mailboxEntOAuthSafeError(err)
		}
		if len(accounts) == 0 {
			return mailboxOAuthError(MailboxOAuthErrorNotFound)
		}
		if len(accounts) != 1 {
			return mailboxOAuthError(MailboxOAuthErrorUnavailable)
		}
		candidate := accounts[0]
		if candidate == nil || candidate.ID == uuid.Nil || candidate.OrganizationID == uuid.Nil || candidate.UserID == uuid.Nil || candidate.EngineAccountID == nil || *candidate.EngineAccountID != engineAccountID || candidate.Status != mailaccount.StatusActive || !candidate.DesiredEnabled || candidate.DeployedGeneration == nil || *candidate.DeployedGeneration < 1 {
			return mailboxOAuthError(MailboxOAuthErrorUnavailable)
		}

		scope := MailboxOAuthScope{ActorID: candidate.UserID, OrganizationID: candidate.OrganizationID, AccountID: candidate.ID}
		account, err := mailboxEntOAuthAccountForUpdate(ctx, transaction, scope)
		if err != nil {
			return err
		}
		if account == nil || account.ID != candidate.ID || account.OrganizationID != candidate.OrganizationID || account.UserID != candidate.UserID || account.EngineAccountID == nil || *account.EngineAccountID != engineAccountID {
			return mailboxOAuthError(MailboxOAuthErrorUnavailable)
		}

		credentials, err := transaction.MailOAuthCredential.Query().
			Where(mailoauthcredential.MailAccountIDEQ(account.ID)).
			Order(ent.Asc(mailoauthcredential.FieldID)).
			ForUpdate().
			All(ctx)
		if err != nil {
			return mailboxEntOAuthSafeError(err)
		}
		if len(credentials) == 0 {
			return mailboxOAuthError(MailboxOAuthErrorReconnectRequired)
		}
		if len(credentials) != 1 {
			return mailboxOAuthError(MailboxOAuthErrorUnavailable)
		}
		credential := credentials[0]
		if !mailboxEntOAuthActiveRowValid(credential, account.ID) {
			if credential != nil && credential.Status != mailoauthcredential.Status("active") {
				return mailboxOAuthError(MailboxOAuthErrorReconnectRequired)
			}
			return mailboxOAuthError(MailboxOAuthErrorUnavailable)
		}

		refreshExpiresAt := *credential.RefreshTokenExpiresAt
		snapshot = MailboxOAuthCredentialSnapshot{
			AccountID: account.ID, OrganizationID: account.OrganizationID, ActorID: account.UserID,
			EngineAccountID: *account.EngineAccountID, IssuerOrigin: credential.IssuerOrigin,
			CredentialGeneration: credential.CredentialGeneration, Version: credential.Version,
			AccessTokenCiphertext: copyMailboxOAuthBytes(credential.AccessTokenCiphertext),
			RefreshTokenCiphertext: copyMailboxOAuthBytes(credential.RefreshTokenCiphertext),
			AccessTokenExpiresAt: *credential.AccessTokenExpiresAt, RefreshTokenExpiresAt: &refreshExpiresAt,
		}
		if !validMailboxOAuthSnapshot(snapshot, scope) || snapshot.EngineAccountID != engineAccountID {
			clearStalwartMailboxOAuthSnapshot(&snapshot)
			return mailboxOAuthError(MailboxOAuthErrorUnavailable)
		}
		return nil
	})
	if err != nil {
		clearStalwartMailboxOAuthSnapshot(&snapshot)
		return MailboxOAuthCredentialSnapshot{}, mailboxEntOAuthSafeError(err)
	}
	result := mailboxOAuthSnapshotCopy(snapshot)
	clearStalwartMailboxOAuthSnapshot(&snapshot)
	return result, nil
}

var _ MailboxOAuthBearerSnapshotSource = (*MailboxEntRepository)(nil)
