package mail

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailoauthcredential"
)

// MailboxOAuthConnectionStatus is the intentionally small browser-safe view
// of one account-bound OAuth credential.
type MailboxOAuthConnectionStatus string

const (
	MailboxOAuthConnectionNotConnected      MailboxOAuthConnectionStatus = "not_connected"
	MailboxOAuthConnectionActive            MailboxOAuthConnectionStatus = "active"
	MailboxOAuthConnectionReconnectRequired MailboxOAuthConnectionStatus = "reconnect_required"
	MailboxOAuthConnectionDisconnected      MailboxOAuthConnectionStatus = "disconnected"
)

// MailboxOAuthConnectionStatusSnapshot is a safe persistence read model. The
// generation remains internal to the HTTP adapter and prevents an incomplete
// row from becoming a public status response.
type MailboxOAuthConnectionStatusSnapshot struct {
	Status               MailboxOAuthConnectionStatus
	CredentialGeneration int64
	AccessTokenExpiresAt *time.Time
}

// MailboxOAuthConnectionStatusSource reads only the safe connection state for
// one owner-bound mailbox. It deliberately does not expose OAuth credentials.
type MailboxOAuthConnectionStatusSource interface {
	ReadMailboxOAuthConnectionStatus(context.Context, MailboxOAuthScope, string) (MailboxOAuthConnectionStatusSnapshot, error)
}

// ReadMailboxOAuthConnectionStatus reads the exact current owner account and
// at most one credential without acquiring transactional locks. The account
// recheck closes the race between HTTP owner resolution and this canonical
// credential read.
func (repository *MailboxEntRepository) ReadMailboxOAuthConnectionStatus(ctx context.Context, scope MailboxOAuthScope, engineAccountID string) (MailboxOAuthConnectionStatusSnapshot, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxOAuthScope(scope) || !validMailboxOAuthOpaque(engineAccountID, MaxEngineOpaqueIDBytes) {
		return MailboxOAuthConnectionStatusSnapshot{}, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}

	account, err := repository.client.MailAccount.Query().
		Where(
			mailaccount.IDEQ(scope.AccountID),
			mailaccount.OrganizationIDEQ(scope.OrganizationID),
			mailaccount.UserIDEQ(scope.ActorID),
			mailaccount.StatusEQ(mailaccount.StatusActive),
			mailaccount.DesiredEnabledEQ(true),
			mailaccount.EngineAccountIDEQ(engineAccountID),
		).
		Only(ctx)
	if err != nil {
		return MailboxOAuthConnectionStatusSnapshot{}, mailboxEntOAuthSafeError(err)
	}
	if account == nil || account.ID != scope.AccountID || account.OrganizationID != scope.OrganizationID || account.UserID != scope.ActorID || account.EngineAccountID == nil || *account.EngineAccountID != engineAccountID || account.Status != mailaccount.StatusActive || !account.DesiredEnabled || account.DeployedGeneration == nil || *account.DeployedGeneration < 1 || MailboxProjectionState(account.ProjectionState) != MailboxProjectionReady {
		return MailboxOAuthConnectionStatusSnapshot{}, mailboxOAuthError(MailboxOAuthErrorNotFound)
	}

	credential, err := repository.client.MailOAuthCredential.Query().
		Where(mailoauthcredential.MailAccountIDEQ(account.ID)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return MailboxOAuthConnectionStatusSnapshot{Status: MailboxOAuthConnectionNotConnected}, nil
	}
	if err != nil {
		return MailboxOAuthConnectionStatusSnapshot{}, mailboxEntOAuthSafeError(err)
	}
	if !mailboxEntOAuthCredentialIdentityValid(credential, account.ID) {
		return MailboxOAuthConnectionStatusSnapshot{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}

	result := MailboxOAuthConnectionStatusSnapshot{CredentialGeneration: credential.CredentialGeneration}
	switch credential.Status {
	case mailoauthcredential.Status("active"):
		if !mailboxEntOAuthActiveRowValid(credential, account.ID) {
			return MailboxOAuthConnectionStatusSnapshot{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
		}
		expiresAt := *credential.AccessTokenExpiresAt
		result.Status = MailboxOAuthConnectionActive
		result.AccessTokenExpiresAt = &expiresAt
	case mailoauthcredential.Status("reconnect_required"):
		if !mailboxEntOAuthReconnectRequiredRowValid(credential, account.ID) {
			return MailboxOAuthConnectionStatusSnapshot{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
		}
		result.Status = MailboxOAuthConnectionReconnectRequired
	case mailoauthcredential.Status("disconnected"):
		if !mailboxEntOAuthDisconnectedRowValid(credential, account.ID) {
			return MailboxOAuthConnectionStatusSnapshot{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
		}
		result.Status = MailboxOAuthConnectionDisconnected
	default:
		return MailboxOAuthConnectionStatusSnapshot{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	return result, nil
}

func mailboxEntOAuthReconnectRequiredRowValid(row *ent.MailOAuthCredential, accountID uuid.UUID) bool {
	return mailboxEntOAuthCredentialIdentityValid(row, accountID) && row.Status == mailoauthcredential.Status("reconnect_required") && row.AccessTokenCiphertext == nil && row.RefreshTokenCiphertext == nil && row.TokenType == nil && row.GrantedScope == nil && row.AccessTokenExpiresAt == nil && row.RefreshTokenExpiresAt == nil && row.RefreshedAt == nil && row.DisconnectedAt == nil && row.LastErrorCode != nil && validMailboxOAuthReconnectReason(MailboxOAuthReconnectReason(*row.LastErrorCode))
}

func mailboxEntOAuthDisconnectedRowValid(row *ent.MailOAuthCredential, accountID uuid.UUID) bool {
	return mailboxEntOAuthCredentialIdentityValid(row, accountID) && row.Status == mailoauthcredential.Status("disconnected") && row.AccessTokenCiphertext == nil && row.RefreshTokenCiphertext == nil && row.TokenType == nil && row.GrantedScope == nil && row.AccessTokenExpiresAt == nil && row.RefreshTokenExpiresAt == nil && row.RefreshedAt == nil && row.DisconnectedAt != nil && mailboxEntValidTime(*row.DisconnectedAt) && row.LastErrorCode == nil
}
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
		if account == nil || account.ID != candidate.ID || account.OrganizationID != candidate.OrganizationID || account.UserID != candidate.UserID ||
			account.EngineAccountID == nil || *account.EngineAccountID != engineAccountID ||
			account.Status != mailaccount.StatusActive || !account.DesiredEnabled ||
			account.DeployedGeneration == nil || *account.DeployedGeneration < 1 {
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
			AccessTokenCiphertext: copyMailboxOAuthBytes(*credential.AccessTokenCiphertext),
			RefreshTokenCiphertext: copyMailboxOAuthBytes(*credential.RefreshTokenCiphertext),
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
var _ MailboxOAuthConnectionStatusSource = (*MailboxEntRepository)(nil)
