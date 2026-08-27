package mail

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailoauthcredential"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

func (repository *MailboxEntRepository) ResolveMailboxOAuthAccount(ctx context.Context, scope MailboxOAuthScope) (MailboxOAuthAccount, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxOAuthScope(scope) {
		return MailboxOAuthAccount{}, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	var account MailboxOAuthAccount
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		row, err := mailboxEntOAuthAccountForUpdate(ctx, transaction, scope)
		if err != nil {
			return err
		}
		generation := int64(1)
		credential, err := transaction.MailOAuthCredential.Query().
			Where(mailoauthcredential.MailAccountIDEQ(row.ID)).
			ForUpdate().
			Only(ctx)
		if err == nil {
			if !mailboxEntOAuthCredentialIdentityValid(credential, row.ID) {
				return mailboxOAuthError(MailboxOAuthErrorUnavailable)
			}
			generation = credential.CredentialGeneration
		} else if !ent.IsNotFound(err) {
			return mailboxEntOAuthSafeError(err)
		}
		account = MailboxOAuthAccount{Scope: scope, EngineAccountID: *row.EngineAccountID, CredentialGeneration: generation}
		return nil
	})
	if err != nil {
		return MailboxOAuthAccount{}, mailboxEntOAuthSafeError(err)
	}
	return account, nil
}

func (repository *MailboxEntRepository) StoreMailboxOAuthActive(ctx context.Context, request MailboxOAuthStoreActiveRequest) error {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxEntOAuthStoreRequestValid(request) {
		return mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	return repository.withMailboxEntOAuthTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := mailboxEntOAuthAccountForUpdate(ctx, transaction, request.Account.Scope)
		if err != nil {
			return err
		}
		if *account.EngineAccountID != request.Account.EngineAccountID {
			return mailboxOAuthError(MailboxOAuthErrorConflict)
		}
		credential, err := transaction.MailOAuthCredential.Query().
			Where(mailoauthcredential.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			if request.Account.CredentialGeneration != 1 {
				return mailboxOAuthError(MailboxOAuthErrorConflict)
			}
			_, err = transaction.MailOAuthCredential.Create().
				SetMailAccountID(account.ID).
				SetProvider(mailoauthcredential.Provider(mailboxOAuthProvider)).
				SetIssuerOrigin(request.IssuerOrigin).
				SetStatus(mailoauthcredential.Status("active")).
				SetCredentialGeneration(1).
				SetEncryptionFormatVersion(mailboxOAuthEncryptionFormat).
				SetAccessTokenCiphertext(copyMailboxOAuthBytes(request.AccessTokenCiphertext)).
				SetRefreshTokenCiphertext(copyMailboxOAuthBytes(request.RefreshTokenCiphertext)).
				SetTokenType(mailoauthcredential.TokenType("bearer")).
				SetGrantedScope(mailboxOAuthCanonicalScope).
				SetAccessTokenExpiresAt(request.AccessTokenExpiresAt).
				SetRefreshTokenExpiresAt(*request.RefreshTokenExpiresAt).
				SetConnectedAt(request.ObservedAt).
				SetCreatedAt(request.ObservedAt).
				SetUpdatedAt(request.ObservedAt).
				SetVersion(1).
				Save(ctx)
			return mailboxEntOAuthSafeError(err)
		}
		if err != nil {
			return mailboxEntOAuthSafeError(err)
		}
		if !mailboxEntOAuthCredentialIdentityValid(credential, account.ID) || credential.CredentialGeneration != request.Account.CredentialGeneration || credential.IssuerOrigin != request.IssuerOrigin || credential.Version == int(^uint(0)>>1) {
			return mailboxOAuthError(MailboxOAuthErrorConflict)
		}
		updated, err := transaction.MailOAuthCredential.Update().
			Where(
				mailoauthcredential.IDEQ(credential.ID),
				mailoauthcredential.MailAccountIDEQ(account.ID),
				mailoauthcredential.CredentialGenerationEQ(request.Account.CredentialGeneration),
				mailoauthcredential.VersionEQ(credential.Version),
			).
			SetStatus(mailoauthcredential.Status("active")).
			SetAccessTokenCiphertext(copyMailboxOAuthBytes(request.AccessTokenCiphertext)).
			SetRefreshTokenCiphertext(copyMailboxOAuthBytes(request.RefreshTokenCiphertext)).
			SetTokenType(mailoauthcredential.TokenType("bearer")).
			SetGrantedScope(mailboxOAuthCanonicalScope).
			SetAccessTokenExpiresAt(request.AccessTokenExpiresAt).
			SetRefreshTokenExpiresAt(*request.RefreshTokenExpiresAt).
			ClearRefreshedAt().
			ClearDisconnectedAt().
			ClearLastErrorCode().
			SetUpdatedAt(request.ObservedAt).
			AddVersion(1).
			Save(ctx)
		if err != nil {
			return mailboxEntOAuthSafeError(err)
		}
		if updated != 1 {
			return mailboxOAuthError(MailboxOAuthErrorConflict)
		}
		return nil
	})
}

func (repository *MailboxEntRepository) LoadMailboxOAuthActive(ctx context.Context, scope MailboxOAuthScope) (MailboxOAuthCredentialSnapshot, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxOAuthScope(scope) {
		return MailboxOAuthCredentialSnapshot{}, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	var snapshot MailboxOAuthCredentialSnapshot
	err := repository.withMailboxEntOAuthTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := mailboxEntOAuthAccountForUpdate(ctx, transaction, scope)
		if err != nil {
			return err
		}
		credential, err := transaction.MailOAuthCredential.Query().
			Where(mailoauthcredential.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			return mailboxOAuthError(MailboxOAuthErrorReconnectRequired)
		}
		if err != nil {
			return mailboxEntOAuthSafeError(err)
		}
		if credential.Status != mailoauthcredential.Status("active") {
			return mailboxOAuthError(MailboxOAuthErrorReconnectRequired)
		}
		if !mailboxEntOAuthActiveRowValid(credential, account.ID) {
			return mailboxOAuthError(MailboxOAuthErrorUnavailable)
		}
		refreshExpiresAt := *credential.RefreshTokenExpiresAt
		snapshot = MailboxOAuthCredentialSnapshot{
			AccountID: account.ID, OrganizationID: account.OrganizationID, ActorID: account.UserID,
			EngineAccountID: *account.EngineAccountID, IssuerOrigin: credential.IssuerOrigin,
			CredentialGeneration: credential.CredentialGeneration, Version: credential.Version,
			AccessTokenCiphertext: copyMailboxOAuthBytes(*credential.AccessTokenCiphertext), RefreshTokenCiphertext: copyMailboxOAuthBytes(*credential.RefreshTokenCiphertext),
			AccessTokenExpiresAt: *credential.AccessTokenExpiresAt, RefreshTokenExpiresAt: &refreshExpiresAt,
		}
		return nil
	})
	if err != nil {
		return MailboxOAuthCredentialSnapshot{}, mailboxEntOAuthSafeError(err)
	}
	return mailboxOAuthSnapshotCopy(snapshot), nil
}

func (repository *MailboxEntRepository) RefreshMailboxOAuthActive(ctx context.Context, request MailboxOAuthRefreshRequest) error {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxEntOAuthRefreshRequestValid(request) {
		return mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	return repository.withMailboxEntOAuthTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := mailboxEntOAuthAccountForUpdate(ctx, transaction, request.Scope)
		if err != nil {
			return err
		}
		credential, err := transaction.MailOAuthCredential.Query().
			Where(mailoauthcredential.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			return mailboxOAuthError(MailboxOAuthErrorConflict)
		}
		if err != nil {
			return mailboxEntOAuthSafeError(err)
		}
		if !mailboxEntOAuthActiveRowValid(credential, account.ID) || credential.CredentialGeneration != request.ExpectedGeneration || credential.Version != request.ExpectedVersion || credential.Version == int(^uint(0)>>1) {
			return mailboxOAuthError(MailboxOAuthErrorConflict)
		}
		if request.RefreshTokenCiphertext == nil && (credential.RefreshTokenExpiresAt == nil || !credential.RefreshTokenExpiresAt.After(request.AccessTokenExpiresAt)) {
			return mailboxOAuthError(MailboxOAuthErrorConflict)
		}
		builder := transaction.MailOAuthCredential.Update().
			Where(
				mailoauthcredential.IDEQ(credential.ID),
				mailoauthcredential.MailAccountIDEQ(account.ID),
				mailoauthcredential.StatusEQ(mailoauthcredential.Status("active")),
				mailoauthcredential.CredentialGenerationEQ(request.ExpectedGeneration),
				mailoauthcredential.VersionEQ(request.ExpectedVersion),
			).
			SetAccessTokenCiphertext(copyMailboxOAuthBytes(request.AccessTokenCiphertext)).
			SetTokenType(mailoauthcredential.TokenType("bearer")).
			SetGrantedScope(mailboxOAuthCanonicalScope).
			SetAccessTokenExpiresAt(request.AccessTokenExpiresAt).
			SetRefreshedAt(request.ObservedAt).
			ClearDisconnectedAt().
			ClearLastErrorCode().
			SetUpdatedAt(request.ObservedAt).
			AddVersion(1)
		if request.RefreshTokenCiphertext != nil {
			builder.SetRefreshTokenCiphertext(copyMailboxOAuthBytes(request.RefreshTokenCiphertext)).
				SetRefreshTokenExpiresAt(*request.RefreshTokenExpiresAt)
		}
		updated, err := builder.Save(ctx)
		if err != nil {
			return mailboxEntOAuthSafeError(err)
		}
		if updated != 1 {
			return mailboxOAuthError(MailboxOAuthErrorConflict)
		}
		return nil
	})
}

func (repository *MailboxEntRepository) RequireMailboxOAuthReconnect(ctx context.Context, scope MailboxOAuthScope, expectedGeneration int64, expectedVersion int, reason MailboxOAuthReconnectReason, observedAt time.Time) error {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxOAuthScope(scope) || expectedGeneration < 1 || expectedVersion < 1 || expectedVersion == int(^uint(0)>>1) || !validMailboxOAuthReconnectReason(reason) || !mailboxEntValidTime(observedAt) {
		return mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	return repository.withMailboxEntOAuthTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := mailboxEntOAuthAccountForUpdate(ctx, transaction, scope)
		if err != nil {
			return err
		}
		credential, err := transaction.MailOAuthCredential.Query().
			Where(mailoauthcredential.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntOAuthSafeError(err)
		}
		if !mailboxEntOAuthActiveRowValid(credential, account.ID) || credential.CredentialGeneration != expectedGeneration || credential.Version != expectedVersion {
			return mailboxOAuthError(MailboxOAuthErrorConflict)
		}
		updated, err := transaction.MailOAuthCredential.Update().
			Where(
				mailoauthcredential.IDEQ(credential.ID),
				mailoauthcredential.MailAccountIDEQ(account.ID),
				mailoauthcredential.StatusEQ(mailoauthcredential.Status("active")),
				mailoauthcredential.CredentialGenerationEQ(expectedGeneration),
				mailoauthcredential.VersionEQ(expectedVersion),
			).
			SetStatus(mailoauthcredential.Status("reconnect_required")).
			ClearAccessTokenCiphertext().
			ClearRefreshTokenCiphertext().
			ClearTokenType().
			ClearGrantedScope().
			ClearAccessTokenExpiresAt().
			ClearRefreshTokenExpiresAt().
			ClearRefreshedAt().
			ClearDisconnectedAt().
			SetLastErrorCode(mailoauthcredential.LastErrorCode(reason)).
			SetUpdatedAt(observedAt).
			AddVersion(1).
			Save(ctx)
		if err != nil {
			return mailboxEntOAuthSafeError(err)
		}
		if updated != 1 {
			return mailboxOAuthError(MailboxOAuthErrorConflict)
		}
		return nil
	})
}

func (repository *MailboxEntRepository) DisconnectMailboxOAuth(ctx context.Context, scope MailboxOAuthScope, issuerOrigin string, disconnectedAt time.Time) error {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxOAuthScope(scope) || !validMailboxOAuthIssuerOrigin(issuerOrigin) || !mailboxEntValidTime(disconnectedAt) {
		return mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	return repository.withMailboxEntOAuthTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := mailboxEntOAuthAccountForUpdate(ctx, transaction, scope)
		if err != nil {
			return err
		}
		credential, err := transaction.MailOAuthCredential.Query().
			Where(mailoauthcredential.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			_, err = transaction.MailOAuthCredential.Create().
				SetMailAccountID(account.ID).
				SetProvider(mailoauthcredential.Provider(mailboxOAuthProvider)).
				SetIssuerOrigin(issuerOrigin).
				SetStatus(mailoauthcredential.Status("disconnected")).
				SetCredentialGeneration(2).
				SetEncryptionFormatVersion(mailboxOAuthEncryptionFormat).
				SetConnectedAt(disconnectedAt).
				SetDisconnectedAt(disconnectedAt).
				SetCreatedAt(disconnectedAt).
				SetUpdatedAt(disconnectedAt).
				SetVersion(1).
				Save(ctx)
			return mailboxEntOAuthSafeError(err)
		}
		if err != nil {
			return mailboxEntOAuthSafeError(err)
		}
		if !mailboxEntOAuthCredentialIdentityValid(credential, account.ID) || credential.IssuerOrigin != issuerOrigin {
			return mailboxOAuthError(MailboxOAuthErrorUnavailable)
		}
		if credential.CredentialGeneration == int64(^uint64(0)>>1) || credential.Version == int(^uint(0)>>1) {
			return mailboxOAuthError(MailboxOAuthErrorConflict)
		}
		updated, err := transaction.MailOAuthCredential.Update().
			Where(
				mailoauthcredential.IDEQ(credential.ID),
				mailoauthcredential.MailAccountIDEQ(account.ID),
				mailoauthcredential.CredentialGenerationEQ(credential.CredentialGeneration),
				mailoauthcredential.VersionEQ(credential.Version),
			).
			SetStatus(mailoauthcredential.Status("disconnected")).
			SetCredentialGeneration(credential.CredentialGeneration + 1).
			ClearAccessTokenCiphertext().
			ClearRefreshTokenCiphertext().
			ClearTokenType().
			ClearGrantedScope().
			ClearAccessTokenExpiresAt().
			ClearRefreshTokenExpiresAt().
			ClearRefreshedAt().
			SetDisconnectedAt(disconnectedAt).
			ClearLastErrorCode().
			SetUpdatedAt(disconnectedAt).
			AddVersion(1).
			Save(ctx)
		if err != nil {
			return mailboxEntOAuthSafeError(err)
		}
		if updated != 1 {
			return mailboxOAuthError(MailboxOAuthErrorConflict)
		}
		return nil
	})
}

func (repository *MailboxEntRepository) withMailboxEntOAuthTransaction(ctx context.Context, callback func(*ent.Tx) error) error {
	err := repository.withMailboxEntTransaction(ctx, callback)
	return mailboxEntOAuthSafeError(err)
}

func mailboxEntOAuthAccountForUpdate(ctx context.Context, transaction *ent.Tx, scope MailboxOAuthScope) (*ent.MailAccount, error) {
	if transaction == nil || !validMailboxOAuthScope(scope) {
		return nil, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	organizationRow, err := transaction.Organization.Query().
		Where(
			organization.IDEQ(scope.OrganizationID),
			organization.HasUsersWith(
				user.IDEQ(scope.ActorID),
				user.StatusEQ(user.StatusActive),
			),
		).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, mailboxOAuthError(MailboxOAuthErrorNotFound)
	}
	if err != nil {
		return nil, mailboxEntOAuthSafeError(err)
	}
	actor, err := transaction.User.Query().
		Where(
			user.IDEQ(scope.ActorID),
			user.StatusEQ(user.StatusActive),
			user.HasOrganizationWith(organization.IDEQ(organizationRow.ID)),
		).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, mailboxOAuthError(MailboxOAuthErrorNotFound)
	}
	if err != nil {
		return nil, mailboxEntOAuthSafeError(err)
	}
	account, err := transaction.MailAccount.Query().
		Where(
			mailaccount.IDEQ(scope.AccountID),
			mailaccount.OrganizationIDEQ(organizationRow.ID),
			mailaccount.UserIDEQ(actor.ID),
			mailaccount.StatusEQ(mailaccount.StatusActive),
			mailaccount.DesiredEnabledEQ(true),
			mailaccount.EngineAccountIDNotNil(),
		).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, mailboxOAuthError(MailboxOAuthErrorNotFound)
	}
	if err != nil {
		return nil, mailboxEntOAuthSafeError(err)
	}
	if organizationRow.ID != scope.OrganizationID || actor.ID != scope.ActorID || account.ID != scope.AccountID || account.OrganizationID != organizationRow.ID || account.UserID != actor.ID || account.EngineAccountID == nil || !validMailboxOAuthOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) {
		return nil, mailboxOAuthError(MailboxOAuthErrorNotFound)
	}
	return account, nil
}

func mailboxEntOAuthCredentialIdentityValid(row *ent.MailOAuthCredential, accountID uuid.UUID) bool {
	return row != nil && row.ID != uuid.Nil && row.MailAccountID == accountID && row.Provider == mailoauthcredential.Provider(mailboxOAuthProvider) && validMailboxOAuthIssuerOrigin(row.IssuerOrigin) && row.CredentialGeneration >= 1 && row.EncryptionFormatVersion == mailboxOAuthEncryptionFormat && row.Version >= 1 && mailboxEntValidTime(row.ConnectedAt) && mailboxEntValidTime(row.CreatedAt) && mailboxEntValidTime(row.UpdatedAt)
}

func mailboxEntOAuthActiveRowValid(row *ent.MailOAuthCredential, accountID uuid.UUID) bool {
	return mailboxEntOAuthCredentialIdentityValid(row, accountID) && row.Status == mailoauthcredential.Status("active") && row.AccessTokenCiphertext != nil && row.RefreshTokenCiphertext != nil && len(*row.AccessTokenCiphertext) >= mailboxOAuthCiphertextMinimum && len(*row.AccessTokenCiphertext) <= mailboxOAuthCiphertextMaximum && len(*row.RefreshTokenCiphertext) >= mailboxOAuthCiphertextMinimum && len(*row.RefreshTokenCiphertext) <= mailboxOAuthCiphertextMaximum && row.TokenType != nil && *row.TokenType == mailoauthcredential.TokenType("bearer") && row.GrantedScope != nil && *row.GrantedScope == mailboxOAuthCanonicalScope && row.AccessTokenExpiresAt != nil && mailboxEntValidTime(*row.AccessTokenExpiresAt) && row.RefreshTokenExpiresAt != nil && mailboxEntValidTime(*row.RefreshTokenExpiresAt) && row.RefreshTokenExpiresAt.After(*row.AccessTokenExpiresAt) && row.DisconnectedAt == nil && row.LastErrorCode == nil && (row.RefreshedAt == nil || mailboxEntValidTime(*row.RefreshedAt))
}

func mailboxEntOAuthStoreRequestValid(request MailboxOAuthStoreActiveRequest) bool {
	return validMailboxOAuthAccount(request.Account) && validMailboxOAuthIssuerOrigin(request.IssuerOrigin) && len(request.AccessTokenCiphertext) >= mailboxOAuthCiphertextMinimum && len(request.AccessTokenCiphertext) <= mailboxOAuthCiphertextMaximum && len(request.RefreshTokenCiphertext) >= mailboxOAuthCiphertextMinimum && len(request.RefreshTokenCiphertext) <= mailboxOAuthCiphertextMaximum && mailboxEntValidTime(request.AccessTokenExpiresAt) && request.RefreshTokenExpiresAt != nil && mailboxEntValidTime(*request.RefreshTokenExpiresAt) && request.RefreshTokenExpiresAt.After(request.AccessTokenExpiresAt) && mailboxEntValidTime(request.ObservedAt) && request.AccessTokenExpiresAt.After(request.ObservedAt) && request.RefreshTokenExpiresAt.After(request.ObservedAt)
}

func mailboxEntOAuthRefreshRequestValid(request MailboxOAuthRefreshRequest) bool {
	if !validMailboxOAuthScope(request.Scope) || request.ExpectedGeneration < 1 || request.ExpectedVersion < 1 || request.ExpectedVersion == int(^uint(0)>>1) || len(request.AccessTokenCiphertext) < mailboxOAuthCiphertextMinimum || len(request.AccessTokenCiphertext) > mailboxOAuthCiphertextMaximum || !mailboxEntValidTime(request.AccessTokenExpiresAt) || !mailboxEntValidTime(request.ObservedAt) || !request.AccessTokenExpiresAt.After(request.ObservedAt) {
		return false
	}
	rotated := request.RefreshTokenCiphertext != nil || request.RefreshTokenExpiresAt != nil
	return !rotated || len(request.RefreshTokenCiphertext) >= mailboxOAuthCiphertextMinimum && len(request.RefreshTokenCiphertext) <= mailboxOAuthCiphertextMaximum && request.RefreshTokenExpiresAt != nil && mailboxEntValidTime(*request.RefreshTokenExpiresAt) && request.RefreshTokenExpiresAt.After(request.AccessTokenExpiresAt)
}

func validMailboxOAuthReconnectReason(reason MailboxOAuthReconnectReason) bool {
	return reason == MailboxOAuthReconnectInvalidGrant || reason == MailboxOAuthReconnectUnauthorizedAccount || reason == MailboxOAuthReconnectAccountMismatch
}

func mailboxEntOAuthSafeError(err error) error {
	if err == nil {
		return nil
	}
	var oauthError *MailboxOAuthError
	if errors.As(err, &oauthError) && oauthError != nil {
		return mailboxOAuthError(oauthError.Kind)
	}
	if ent.IsNotFound(err) {
		return mailboxOAuthError(MailboxOAuthErrorNotFound)
	}
	if ent.IsConstraintError(err) {
		return mailboxOAuthError(MailboxOAuthErrorConflict)
	}
	return mailboxOAuthError(MailboxOAuthErrorUnavailable)
}

func copyMailboxOAuthBytes(value []byte) []byte {
	if value == nil {
		return nil
	}
	return append([]byte(nil), value...)
}

func mailboxOAuthSnapshotCopy(value MailboxOAuthCredentialSnapshot) MailboxOAuthCredentialSnapshot {
	value.AccessTokenCiphertext = copyMailboxOAuthBytes(value.AccessTokenCiphertext)
	value.RefreshTokenCiphertext = copyMailboxOAuthBytes(value.RefreshTokenCiphertext)
	if value.RefreshTokenExpiresAt != nil {
		copied := *value.RefreshTokenExpiresAt
		value.RefreshTokenExpiresAt = &copied
	}
	return value
}

var _ MailboxOAuthRepository = (*MailboxEntRepository)(nil)
