package mail

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/mailoauthcredential"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

var _ MailboxBootstrapOAuthRepository = (*MailboxEntRepository)(nil)

// ResolveMailboxBootstrapOAuthAccount returns only the exact pending account
// snapshot that may start one password-enrollment bootstrap OAuth flow.
func (repository *MailboxEntRepository) ResolveMailboxBootstrapOAuthAccount(ctx context.Context, scope MailboxOAuthScope) (MailboxBootstrapOAuthAccount, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxOAuthScope(scope) {
		return MailboxBootstrapOAuthAccount{}, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}

	var snapshot MailboxBootstrapOAuthAccount
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		_, _, _, resolved, err := mailboxBootstrapOAuthLockSnapshot(ctx, transaction, scope)
		if err != nil {
			return err
		}
		snapshot = mailboxBootstrapOAuthAccountCopy(resolved)
		return nil
	})
	if err != nil {
		return MailboxBootstrapOAuthAccount{}, mailboxEntOAuthSafeError(err)
	}
	return mailboxBootstrapOAuthAccountCopy(snapshot), nil
}

// FinalizeMailboxBootstrapOAuth persists the encrypted OAuth envelope and
// activates exactly the locked primary identity and mailbox in one transaction.
func (repository *MailboxEntRepository) FinalizeMailboxBootstrapOAuth(ctx context.Context, request MailboxBootstrapOAuthFinalizeRequest) error {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxBootstrapOAuthFinalizeRequest(request) {
		return mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	request.AccessTokenCiphertext = copyMailboxOAuthBytes(request.AccessTokenCiphertext)
	request.RefreshTokenCiphertext = copyMailboxOAuthBytes(request.RefreshTokenCiphertext)
	defer clear(request.AccessTokenCiphertext)
	defer clear(request.RefreshTokenCiphertext)

	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, _, identity, snapshot, err := mailboxBootstrapOAuthLockSnapshot(ctx, transaction, request.Account.Scope)
		if err != nil {
			return err
		}
		if snapshot != request.Account {
			return mailboxOAuthError(MailboxOAuthErrorConflict)
		}
		if account.DeployedHash == nil {
			return mailboxOAuthError(MailboxOAuthErrorConflict)
		}

		if _, err := transaction.MailOAuthCredential.Create().
			SetMailAccountID(account.ID).
			SetProvider(mailoauthcredential.Provider(mailboxOAuthProvider)).
			SetIssuerOrigin(request.IssuerOrigin).
			SetStatus(mailoauthcredential.Status("active")).
			SetCredentialGeneration(1).
			SetEncryptionFormatVersion(mailboxOAuthEncryptionFormat).
			SetAccessTokenCiphertext(request.AccessTokenCiphertext).
			SetRefreshTokenCiphertext(request.RefreshTokenCiphertext).
			SetTokenType(mailoauthcredential.TokenType("bearer")).
			SetGrantedScope(mailboxOAuthCanonicalScope).
			SetAccessTokenExpiresAt(request.AccessTokenExpiresAt).
			SetRefreshTokenExpiresAt(request.RefreshTokenExpiresAt).
			SetConnectedAt(request.ObservedAt).
			SetCreatedAt(request.ObservedAt).
			SetUpdatedAt(request.ObservedAt).
			SetVersion(1).
			Save(ctx); err != nil {
			return mailboxEntOAuthSafeError(err)
		}

		updated, err := transaction.MailIdentity.Update().
			Where(
				mailidentity.IDEQ(identity.ID),
				mailidentity.MailAccountIDEQ(account.ID),
				mailidentity.KindEQ(mailidentity.KindPrimary),
				mailidentity.StatusEQ(mailidentity.Status("pending")),
				mailidentity.SendAllowedEQ(false),
				mailidentity.EngineIdentityIDIsNil(),
				mailidentity.NormalizedAddressEQ(snapshot.PrimaryEmail),
			).
			SetStatus(mailidentity.Status("active")).
			SetSendAllowed(true).
			SetEngineIdentityID(request.EngineIdentityID).
			SetUpdatedAt(request.ObservedAt).
			Save(ctx)
		if err != nil {
			return mailboxEntOAuthSafeError(err)
		}
		if updated != 1 {
			return mailboxOAuthError(MailboxOAuthErrorConflict)
		}

		updated, err = transaction.MailAccount.Update().
			Where(
				mailaccount.IDEQ(account.ID),
				mailaccount.OrganizationIDEQ(snapshot.Scope.OrganizationID),
				mailaccount.UserIDEQ(snapshot.Scope.ActorID),
				mailaccount.MailDomainIDEQ(account.MailDomainID),
				mailaccount.StatusEQ(mailaccount.Status("provisioning")),
				mailaccount.DesiredEnabledEQ(true),
				mailaccount.EngineAccountIDEQ(snapshot.EngineAccountID),
				mailaccount.PasswordEnrollmentGenerationEQ(snapshot.PasswordEnrollmentGeneration),
				mailaccount.PasswordEnrolledEngineAccountIDEQ(snapshot.EngineAccountID),
				mailaccount.PasswordEnrolledAtEQ(*account.PasswordEnrolledAt),
				mailaccount.PasswordEnrollmentLeaseTokenIsNil(),
				mailaccount.PasswordEnrollmentLeaseUntilIsNil(),
				mailaccount.ControlGenerationEQ(account.ControlGeneration),
				mailaccount.ControlHashVersionEQ(account.ControlHashVersion),
				mailaccount.ControlHashEQ(account.ControlHash),
				mailaccount.AttemptedControlGenerationIsNil(),
				mailaccount.AttemptedControlHashVersionIsNil(),
				mailaccount.AttemptedControlHashIsNil(),
				mailaccount.DeployedGenerationEQ(*account.DeployedGeneration),
				mailaccount.DeployedHashVersionEQ(*account.DeployedHashVersion),
				mailaccount.DeployedHashEQ(*account.DeployedHash),
				mailaccount.ProvisionLeaseTokenIsNil(),
				mailaccount.ProvisionLeaseUntilIsNil(),
				mailaccount.NextProvisionAtIsNil(),
				mailaccount.ProvisionErrorCodeIsNil(),
			).
			SetStatus(mailaccount.Status("active")).
			SetProjectionState(mailaccount.ProjectionState("bootstrapping")).
			ClearSyncCursor().
			ClearRebuildCheckpoint().
			ClearRebuildBaseCursor().
			ClearSyncErrorCode().
			ClearLastSyncedAt().
			ClearProvisionErrorCode().
			ClearNextProvisionAt().
			SetNextSyncAt(request.ObservedAt).
			SetSyncScheduleUpdatedAt(request.ObservedAt).
			SetUpdatedAt(request.ObservedAt).
			Save(ctx)
		if err != nil {
			return mailboxEntOAuthSafeError(err)
		}
		if updated != 1 {
			return mailboxOAuthError(MailboxOAuthErrorConflict)
		}
		return nil
	})
	return mailboxEntOAuthSafeError(err)
}

func mailboxBootstrapOAuthLockSnapshot(ctx context.Context, transaction *ent.Tx, scope MailboxOAuthScope) (*ent.MailAccount, *ent.MailDomain, *ent.MailIdentity, MailboxBootstrapOAuthAccount, error) {
	if ctx == nil || transaction == nil || !validMailboxOAuthScope(scope) {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxOAuthError(MailboxOAuthErrorInvalid)
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
	if err != nil {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxBootstrapOAuthLookupError(err)
	}
	if organizationRow == nil || organizationRow.ID != scope.OrganizationID {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxOAuthError(MailboxOAuthErrorNotFound)
	}

	actor, err := transaction.User.Query().
		Where(
			user.IDEQ(scope.ActorID),
			user.StatusEQ(user.StatusActive),
			user.HasOrganizationWith(organization.IDEQ(organizationRow.ID)),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxBootstrapOAuthLookupError(err)
	}
	if actor == nil || actor.ID != scope.ActorID {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxOAuthError(MailboxOAuthErrorNotFound)
	}

	accounts, err := transaction.MailAccount.Query().
		Where(
			mailaccount.IDEQ(scope.AccountID),
			mailaccount.OrganizationIDEQ(organizationRow.ID),
			mailaccount.UserIDEQ(actor.ID),
			mailaccount.StatusEQ(mailaccount.Status("provisioning")),
			mailaccount.DesiredEnabledEQ(true),
			mailaccount.EngineAccountIDNotNil(),
		).
		Order(ent.Asc(mailaccount.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxEntOAuthSafeError(err)
	}
	if len(accounts) != 1 {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxBootstrapOAuthCardinalityError(len(accounts))
	}
	account := accounts[0]
	if account == nil || account.ID != scope.AccountID || account.OrganizationID != organizationRow.ID || account.UserID != actor.ID || !mailboxBootstrapOAuthAccountReady(account) {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxOAuthError(MailboxOAuthErrorConflict)
	}

	domain, err := transaction.MailDomain.Query().
		Where(
			maildomain.IDEQ(account.MailDomainID),
			maildomain.OrganizationIDEQ(organizationRow.ID),
			maildomain.DesiredPresentEQ(true),
			maildomain.EnabledEQ(true),
			maildomain.StateEQ(maildomain.State("verified")),
			maildomain.VerificationStatusEQ(maildomain.VerificationStatus("verified")),
			maildomain.DeletedAtIsNil(),
			maildomain.EngineDomainIDNotNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxBootstrapOAuthLookupError(err)
	}
	if !mailboxBootstrapOAuthDomainReady(domain, organizationRow.ID) {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxOAuthError(MailboxOAuthErrorConflict)
	}

	expectedAddress := account.LocalPart + "@" + domain.Domain
	if !validMailboxBootstrapIdentityExpectedEmail(expectedAddress) {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	identities, err := transaction.MailIdentity.Query().
		Where(
			mailidentity.MailAccountIDEQ(account.ID),
			mailidentity.KindEQ(mailidentity.KindPrimary),
			mailidentity.StatusEQ(mailidentity.Status("pending")),
			mailidentity.SendAllowedEQ(false),
			mailidentity.EngineIdentityIDIsNil(),
			mailidentity.NormalizedAddressEQ(expectedAddress),
		).
		Order(ent.Asc(mailidentity.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxEntOAuthSafeError(err)
	}
	if len(identities) != 1 {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxBootstrapOAuthCardinalityError(len(identities))
	}
	identity := identities[0]
	if identity == nil || identity.ID == uuid.Nil || identity.MailAccountID != account.ID || identity.Kind != mailidentity.KindPrimary || identity.Status != mailidentity.Status("pending") || identity.SendAllowed || identity.EngineIdentityID != nil || identity.NormalizedAddress != expectedAddress {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxOAuthError(MailboxOAuthErrorConflict)
	}

	credentials, err := transaction.MailOAuthCredential.Query().
		Where(mailoauthcredential.MailAccountIDEQ(account.ID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxEntOAuthSafeError(err)
	}
	if len(credentials) != 0 {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, mailboxOAuthError(MailboxOAuthErrorConflict)
	}

	snapshot, err := mailboxBootstrapOAuthAccountSnapshot(scope, account, expectedAddress)
	if err != nil {
		return nil, nil, nil, MailboxBootstrapOAuthAccount{}, err
	}
	return account, domain, identity, snapshot, nil
}

func mailboxBootstrapOAuthAccountSnapshot(scope MailboxOAuthScope, account *ent.MailAccount, primaryEmail string) (MailboxBootstrapOAuthAccount, error) {
	if !validMailboxOAuthScope(scope) || !mailboxBootstrapOAuthAccountReady(account) || !validMailboxBootstrapIdentityExpectedEmail(primaryEmail) {
		return MailboxBootstrapOAuthAccount{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	snapshot := MailboxBootstrapOAuthAccount{
		Scope:                        scope,
		EngineAccountID:              *account.EngineAccountID,
		PrimaryEmail:                 primaryEmail,
		PasswordEnrollmentGeneration: account.PasswordEnrollmentGeneration,
	}
	if !validMailboxBootstrapOAuthAccount(snapshot) {
		return MailboxBootstrapOAuthAccount{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	return mailboxBootstrapOAuthAccountCopy(snapshot), nil
}

func mailboxBootstrapOAuthAccountReady(account *ent.MailAccount) bool {
	if account == nil || account.ID == uuid.Nil || account.OrganizationID == uuid.Nil || account.UserID == uuid.Nil || account.MailDomainID == uuid.Nil || account.Status != mailaccount.Status("provisioning") || !account.DesiredEnabled || account.EngineAccountID == nil || !validMailboxOAuthOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) {
		return false
	}
	if account.ControlGeneration < 1 || account.ControlHashVersion < 1 || len(account.ControlHash) != sha256.Size || account.DeployedGeneration == nil || account.DeployedHashVersion == nil || account.DeployedHash == nil || *account.DeployedGeneration != account.ControlGeneration || *account.DeployedHashVersion != account.ControlHashVersion || len(*account.DeployedHash) != sha256.Size || subtle.ConstantTimeCompare(*account.DeployedHash, account.ControlHash) != 1 {
		return false
	}
	if account.AttemptedControlGeneration != nil || account.AttemptedControlHashVersion != nil || !mailboxEntNullableBytesAbsent(account.AttemptedControlHash) || account.ProvisionLeaseToken != nil || account.ProvisionLeaseUntil != nil || account.NextProvisionAt != nil || account.ProvisionErrorCode != nil {
		return false
	}
	if account.PasswordEnrollmentGeneration < 1 || account.PasswordEnrolledEngineAccountID == nil || *account.PasswordEnrolledEngineAccountID != *account.EngineAccountID || account.PasswordEnrolledAt == nil || !mailboxEntValidTime(*account.PasswordEnrolledAt) || account.PasswordEnrollmentLeaseToken != nil || account.PasswordEnrollmentLeaseUntil != nil {
		return false
	}
	return true
}

func mailboxBootstrapOAuthDomainReady(domain *ent.MailDomain, organizationID uuid.UUID) bool {
	if domain == nil || organizationID == uuid.Nil || domain.ID == uuid.Nil || domain.OrganizationID != organizationID || !domain.DesiredPresent || !domain.Enabled || domain.State != maildomain.State("verified") || domain.VerificationStatus != maildomain.VerificationStatus("verified") || domain.DeletedAt != nil || domain.EngineDomainID == nil || !validMailboxOAuthOpaque(*domain.EngineDomainID, MaxEngineOpaqueIDBytes) {
		return false
	}
	return domain.DesiredGeneration >= 1 && domain.DesiredHashVersion >= 1 && len(domain.DesiredHash) == sha256.Size && domain.DeployedGeneration != nil && domain.DeployedHashVersion != nil && domain.DeployedHash != nil && *domain.DeployedGeneration == domain.DesiredGeneration && *domain.DeployedHashVersion == domain.DesiredHashVersion && len(*domain.DeployedHash) == sha256.Size && subtle.ConstantTimeCompare(*domain.DeployedHash, domain.DesiredHash) == 1
}

func mailboxBootstrapOAuthAccountCopy(snapshot MailboxBootstrapOAuthAccount) MailboxBootstrapOAuthAccount {
	return MailboxBootstrapOAuthAccount{
		Scope:                        snapshot.Scope,
		EngineAccountID:              snapshot.EngineAccountID,
		PrimaryEmail:                 snapshot.PrimaryEmail,
		PasswordEnrollmentGeneration: snapshot.PasswordEnrollmentGeneration,
	}
}

func mailboxBootstrapOAuthCardinalityError(count int) error {
	if count == 0 {
		return mailboxOAuthError(MailboxOAuthErrorNotFound)
	}
	return mailboxOAuthError(MailboxOAuthErrorConflict)
}

func mailboxBootstrapOAuthLookupError(err error) error {
	if ent.IsNotFound(err) {
		return mailboxOAuthError(MailboxOAuthErrorNotFound)
	}
	return mailboxEntOAuthSafeError(err)
}
