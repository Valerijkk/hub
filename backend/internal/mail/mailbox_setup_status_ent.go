package mail

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/mailoauthcredential"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

// ResolveMailboxSetupStatus reads one stable, browser-safe mailbox setup snapshot.
func (repository *MailboxEntRepository) ResolveMailboxSetupStatus(ctx context.Context, actorID uuid.UUID) (MailboxSetupStatus, error) {
	if repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || actorID == uuid.Nil {
		return MailboxSetupStatus{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var status MailboxSetupStatus
	err := repository.withMailboxEntReadSnapshot(ctx, func(transaction *ent.Tx) error {
		resolved, err := mailboxEntResolveMailboxSetupStatus(ctx, transaction, actorID)
		if err != nil {
			return err
		}
		status = mailboxSetupStatusCopy(resolved)
		return nil
	})
	if err != nil {
		return MailboxSetupStatus{}, mailboxEntSafeError(err)
	}
	return mailboxSetupStatusCopy(status), nil
}

func mailboxEntResolveMailboxSetupStatus(ctx context.Context, transaction *ent.Tx, actorID uuid.UUID) (MailboxSetupStatus, error) {
	if ctx == nil || transaction == nil || actorID == uuid.Nil {
		return MailboxSetupStatus{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	organizationRow, err := transaction.Organization.Query().
		Where(organization.HasUsersWith(user.IDEQ(actorID), user.StatusEQ(user.StatusActive))).
		Only(ctx)
	if err != nil {
		return MailboxSetupStatus{}, mailboxSetupStatusActorLookupError(err)
	}
	if organizationRow == nil || organizationRow.ID == uuid.Nil {
		return MailboxSetupStatus{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}

	actor, err := transaction.User.Query().
		Where(
			user.IDEQ(actorID),
			user.StatusEQ(user.StatusActive),
			user.HasOrganizationWith(organization.IDEQ(organizationRow.ID)),
		).
		Only(ctx)
	if err != nil {
		return MailboxSetupStatus{}, mailboxSetupStatusActorLookupError(err)
	}
	if actor == nil || actor.ID != actorID {
		return MailboxSetupStatus{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}

	accounts, err := transaction.MailAccount.Query().
		Where(mailaccount.UserIDEQ(actor.ID)).
		Order(ent.Asc(mailaccount.FieldID)).
		All(ctx)
	if err != nil {
		return MailboxSetupStatus{}, mailboxEntSafeError(err)
	}
	if len(accounts) == 0 {
		return mailboxSetupStatusWithoutAddress(MailboxSetupStagePreparing), nil
	}
	if len(accounts) != 1 {
		return mailboxSetupStatusActionRequired(), nil
	}
	account := accounts[0]
	if account == nil || account.ID == uuid.Nil || account.UserID != actor.ID || account.OrganizationID != organizationRow.ID || account.MailDomainID == uuid.Nil {
		return mailboxSetupStatusActionRequired(), nil
	}

	domains, err := transaction.MailDomain.Query().
		Where(
			maildomain.IDEQ(account.MailDomainID),
			maildomain.OrganizationIDEQ(organizationRow.ID),
		).
		Order(ent.Asc(maildomain.FieldID)).
		All(ctx)
	if err != nil {
		return MailboxSetupStatus{}, mailboxEntSafeError(err)
	}
	if len(domains) != 1 {
		return mailboxSetupStatusActionRequired(), nil
	}
	domain := domains[0]

	identities, err := transaction.MailIdentity.Query().
		Where(
			mailidentity.MailAccountIDEQ(account.ID),
			mailidentity.KindEQ(mailidentity.KindPrimary),
		).
		Order(ent.Asc(mailidentity.FieldID)).
		All(ctx)
	if err != nil {
		return MailboxSetupStatus{}, mailboxEntSafeError(err)
	}
	if len(identities) != 1 {
		return mailboxSetupStatusActionRequired(), nil
	}
	identity := identities[0]

	credentials, err := transaction.MailOAuthCredential.Query().
		Where(mailoauthcredential.MailAccountIDEQ(account.ID)).
		Order(ent.Asc(mailoauthcredential.FieldID)).
		All(ctx)
	if err != nil {
		return MailboxSetupStatus{}, mailboxEntSafeError(err)
	}

	databaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
	if err != nil {
		return MailboxSetupStatus{}, mailboxEntSafeError(err)
	}
	return mailboxSetupStatusFromRows(organizationRow.ID, account, domain, identity, credentials, databaseWallClock), nil
}

func mailboxSetupStatusFromRows(organizationID uuid.UUID, account *ent.MailAccount, domain *ent.MailDomain, identity *ent.MailIdentity, credentials []*ent.MailOAuthCredential, databaseWallClock time.Time) MailboxSetupStatus {
	if organizationID == uuid.Nil || !mailboxEntValidTime(databaseWallClock) || !mailboxSetupStatusAccountIntegrityValid(account) || !mailboxSetupStatusDomainIntegrityValid(domain, organizationID) {
		return mailboxSetupStatusActionRequired()
	}

	expectedAddress, validAddress := mailboxSetupStatusExpectedPrimaryAddress(account, domain)
	if !validAddress || !mailboxSetupStatusPrimaryIdentityValid(identity, account.ID, expectedAddress) {
		return mailboxSetupStatusActionRequired()
	}
	if mailboxSetupStatusDomainHardFailure(domain, organizationID) || mailboxSetupStatusAccountHardFailure(account) {
		return mailboxSetupStatusActionRequired()
	}
	if !domain.IsDefault || !mailboxBootstrapOAuthDomainReady(domain, organizationID) {
		return mailboxSetupStatusWithAddress(MailboxSetupStagePreparing, expectedAddress)
	}

	switch account.Status {
	case mailaccount.Status("provisioning"):
		if account.ProjectionState != mailaccount.ProjectionState("bootstrapping") || !mailboxSetupStatusPendingPrimaryIdentityValid(identity, account.ID, expectedAddress) || len(credentials) != 0 {
			return mailboxSetupStatusActionRequired()
		}
		if mailboxBootstrapOAuthAccountReady(account) {
			return mailboxSetupStatusWithAddress(MailboxSetupStageOAuthRequired, expectedAddress)
		}
		if mailboxPasswordEnrollmentEligible(account) && mailboxSetupStatusAccountCurrentControlValid(account) {
			if account.PasswordEnrolledEngineAccountID != nil || account.PasswordEnrolledAt != nil {
				return mailboxSetupStatusActionRequired()
			}
			if account.PasswordEnrollmentLeaseUntil != nil && account.PasswordEnrollmentLeaseUntil.After(databaseWallClock) {
				return mailboxSetupStatusWithAddress(MailboxSetupStagePreparing, expectedAddress)
			}
			return mailboxSetupStatusWithAddress(MailboxSetupStagePasswordRequired, expectedAddress)
		}
		if mailboxSetupStatusProvisioningSafe(account) {
			return mailboxSetupStatusWithAddress(MailboxSetupStagePreparing, expectedAddress)
		}
		return mailboxSetupStatusActionRequired()
	case mailaccount.StatusActive:
		if !mailboxSetupStatusActiveAccountReady(account) || !mailboxSetupStatusActivePrimaryIdentityValid(identity, account.ID, expectedAddress) || len(credentials) != 1 || !mailboxEntOAuthActiveRowValid(credentials[0], account.ID) || !mailboxSetupStatusActiveProjectionValid(account) {
			return mailboxSetupStatusActionRequired()
		}
		switch account.ProjectionState {
		case mailaccount.ProjectionState("bootstrapping"), mailaccount.ProjectionState("rebuilding"):
			return mailboxSetupStatusWithAddress(MailboxSetupStageSyncing, expectedAddress)
		case mailaccount.ProjectionStateReady:
			return mailboxSetupStatusWithAddress(MailboxSetupStageReady, expectedAddress)
		default:
			return mailboxSetupStatusActionRequired()
		}
	default:
		return mailboxSetupStatusActionRequired()
	}
}

func mailboxSetupStatusAccountIntegrityValid(account *ent.MailAccount) bool {
	if account == nil || account.ID == uuid.Nil || account.OrganizationID == uuid.Nil || account.UserID == uuid.Nil || account.MailDomainID == uuid.Nil || !mailboxSetupStatusAccountStateValid(account.Status) || !validMailboxProjectionState(MailboxProjectionState(account.ProjectionState)) || !validMailboxBootstrapIdentityExpectedEmail(account.LocalPart+"@example.test") || !mailboxEntValidTime(account.EventCutoffAt) || !mailboxEntValidTime(account.SyncScheduleUpdatedAt) || !mailboxEntValidTime(account.CreatedAt) || !mailboxEntValidTime(account.UpdatedAt) || account.SyncEpoch < 0 || account.SyncLeaseEpoch < 1 || account.MutationEpoch < 0 || account.ProvisionAttempts < 0 || account.SyncAttempts < 0 || !mailboxEntSyncApplyErrorCodeValid(account.SyncErrorCode) {
		return false
	}
	if account.EngineAccountID != nil && !validMailboxOAuthOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) {
		return false
	}
	if !mailboxSetupStatusAccountTupleIntegrityValid(account) || !mailboxPasswordEnrollmentAccountIntegrityValid(account) || !mailboxSetupStatusProvisionLeaseValid(account) || !mailboxSetupStatusMutationLeaseValid(account) {
		return false
	}
	if account.NextProvisionAt != nil && !mailboxEntValidTime(*account.NextProvisionAt) || account.NextSyncAt != nil && !mailboxEntValidTime(*account.NextSyncAt) || account.LastSyncedAt != nil && !mailboxEntValidTime(*account.LastSyncedAt) {
		return false
	}
	if account.EngineAccountID != nil {
		if _, err := mailboxEntSyncClaimSnapshot(account); err != nil {
			return false
		}
	}
	fence, hasLease, leaseValid := mailboxEntSyncLeaseFence(account)
	if !leaseValid {
		return false
	}
	if hasLease {
		if _, _, err := mailboxEntSyncApplyProofFromAccount(account, fence); err != nil {
			return false
		}
	}
	return true
}

func mailboxSetupStatusAccountStateValid(status mailaccount.Status) bool {
	switch status {
	case mailaccount.Status("provisioning"), mailaccount.StatusActive, mailaccount.Status("suspended"), mailaccount.Status("error"), mailaccount.Status("deleting"):
		return true
	default:
		return false
	}
}

func mailboxSetupStatusAccountTupleIntegrityValid(account *ent.MailAccount) bool {
	if account == nil || account.ControlGeneration < 1 || account.ControlHashVersion < 1 || len(account.ControlHash) != sha256.Size {
		return false
	}
	if !mailboxSetupStatusControlTupleValid(account.AttemptedControlGeneration, account.AttemptedControlHashVersion, account.AttemptedControlHash, account.ControlGeneration) || !mailboxSetupStatusControlTupleValid(account.DeployedGeneration, account.DeployedHashVersion, account.DeployedHash, account.ControlGeneration) {
		return false
	}
	return account.ProvisionErrorCode == nil || validMailboxOwnerRecoveryCode(MailboxOwnerRecoveryCode(*account.ProvisionErrorCode))
}

func mailboxSetupStatusControlTupleValid(generation *int64, version *int32, hash *[]byte, maximumGeneration int64) bool {
	hasGeneration := generation != nil
	hasVersion := version != nil
	hasHash := !mailboxEntNullableBytesAbsent(hash)
	if hasGeneration != hasVersion || hasGeneration != hasHash {
		return false
	}
	return !hasGeneration || *generation >= 1 && *generation <= maximumGeneration && *version >= 1 && len(*hash) == sha256.Size
}

func mailboxSetupStatusProvisionLeaseValid(account *ent.MailAccount) bool {
	if account == nil || (account.ProvisionLeaseToken == nil) != (account.ProvisionLeaseUntil == nil) {
		return false
	}
	return account.ProvisionLeaseToken == nil || *account.ProvisionLeaseToken != uuid.Nil && mailboxEntValidTime(*account.ProvisionLeaseUntil)
}

func mailboxSetupStatusMutationLeaseValid(account *ent.MailAccount) bool {
	if account == nil || (account.MutationLeaseToken == nil) != (account.MutationLeaseUntil == nil) {
		return false
	}
	return account.MutationLeaseToken == nil || *account.MutationLeaseToken != uuid.Nil && mailboxEntValidTime(*account.MutationLeaseUntil)
}

func mailboxSetupStatusAccountCurrentControlValid(account *ent.MailAccount) bool {
	if account == nil || account.EngineAccountID == nil || !validMailboxOAuthOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) || !mailboxSetupStatusAccountTupleIntegrityValid(account) || account.DeployedGeneration == nil || account.DeployedHashVersion == nil || mailboxEntNullableBytesAbsent(account.DeployedHash) || *account.DeployedGeneration != account.ControlGeneration || *account.DeployedHashVersion != account.ControlHashVersion || subtle.ConstantTimeCompare(*account.DeployedHash, account.ControlHash) != 1 || account.AttemptedControlGeneration != nil || account.AttemptedControlHashVersion != nil || !mailboxEntNullableBytesAbsent(account.AttemptedControlHash) || account.ProvisionLeaseToken != nil || account.ProvisionLeaseUntil != nil || account.NextProvisionAt != nil || account.ProvisionErrorCode != nil {
		return false
	}
	return true
}

func mailboxSetupStatusAccountHardFailure(account *ent.MailAccount) bool {
	if account == nil || !account.DesiredEnabled || account.ProvisionErrorCode != nil || MailboxProjectionState(account.ProjectionState) == MailboxProjectionDegraded {
		return true
	}
	switch account.Status {
	case mailaccount.Status("error"), mailaccount.Status("suspended"), mailaccount.Status("deleting"):
		return true
	default:
		return false
	}
}

func mailboxSetupStatusProvisioningSafe(account *ent.MailAccount) bool {
	if account == nil || account.Status != mailaccount.Status("provisioning") || !account.DesiredEnabled || account.ProjectionState != mailaccount.ProjectionState("bootstrapping") || account.ProvisionErrorCode != nil || account.PasswordEnrolledEngineAccountID != nil || account.PasswordEnrolledAt != nil {
		return false
	}
	if account.EngineAccountID == nil {
		return account.DeployedGeneration == nil && account.DeployedHashVersion == nil && mailboxEntNullableBytesAbsent(account.DeployedHash)
	}
	return !mailboxSetupStatusAccountCurrentControlValid(account) && (account.AttemptedControlGeneration != nil || account.ProvisionLeaseToken != nil || account.NextProvisionAt != nil)
}

func mailboxSetupStatusActiveAccountReady(account *ent.MailAccount) bool {
	if account == nil || account.Status != mailaccount.StatusActive || !account.DesiredEnabled || !mailboxSetupStatusAccountCurrentControlValid(account) || account.PasswordEnrollmentGeneration < 1 || account.PasswordEnrolledEngineAccountID == nil || account.PasswordEnrolledAt == nil || account.PasswordEnrollmentLeaseToken != nil || account.PasswordEnrollmentLeaseUntil != nil || *account.PasswordEnrolledEngineAccountID != *account.EngineAccountID || !mailboxEntValidTime(*account.PasswordEnrolledAt) {
		return false
	}
	copy := *account
	copy.Status = mailaccount.Status("provisioning")
	return mailboxBootstrapOAuthAccountReady(&copy)
}

func mailboxSetupStatusActiveProjectionValid(account *ent.MailAccount) bool {
	if account == nil || account.NextSyncAt == nil || !mailboxEntValidTime(*account.NextSyncAt) || account.SyncAttempts != 0 || account.SyncErrorCode != nil {
		return false
	}
	switch account.ProjectionState {
	case mailaccount.ProjectionState("bootstrapping"):
		return account.SyncCursor == nil && account.RebuildBaseCursor == nil && account.RebuildCheckpoint == nil && account.LastSyncedAt == nil
	case mailaccount.ProjectionState("rebuilding"):
		return account.RebuildBaseCursor != nil && mailboxEntValidTime(account.SyncScheduleUpdatedAt)
	case mailaccount.ProjectionStateReady:
		return account.SyncCursor != nil && account.RebuildBaseCursor == nil && account.RebuildCheckpoint == nil && account.LastSyncedAt != nil && mailboxEntValidTime(*account.LastSyncedAt)
	default:
		return false
	}
}

func mailboxSetupStatusExpectedPrimaryAddress(account *ent.MailAccount, domain *ent.MailDomain) (string, bool) {
	if account == nil || domain == nil {
		return "", false
	}
	address := account.LocalPart + "@" + domain.Domain
	return address, validMailboxBootstrapIdentityExpectedEmail(address) && validMailboxOwnerPrimaryAddress(address)
}

func mailboxSetupStatusPrimaryIdentityValid(identity *ent.MailIdentity, accountID uuid.UUID, expectedAddress string) bool {
	if identity == nil || accountID == uuid.Nil || identity.ID == uuid.Nil || identity.MailAccountID != accountID || identity.Kind != mailidentity.KindPrimary || identity.NormalizedAddress != expectedAddress || !validMailboxOwnerPrimaryAddress(identity.NormalizedAddress) || !mailboxEntValidTime(identity.CreatedAt) || !mailboxEntValidTime(identity.UpdatedAt) {
		return false
	}
	switch identity.Status {
	case mailidentity.Status("pending"):
		return !identity.SendAllowed && identity.EngineIdentityID == nil
	case mailidentity.StatusActive:
		return identity.SendAllowed && identity.EngineIdentityID != nil && validMailboxOAuthOpaque(*identity.EngineIdentityID, MaxEngineOpaqueIDBytes)
	default:
		return false
	}
}

func mailboxSetupStatusPendingPrimaryIdentityValid(identity *ent.MailIdentity, accountID uuid.UUID, expectedAddress string) bool {
	return mailboxSetupStatusPrimaryIdentityValid(identity, accountID, expectedAddress) && identity.Status == mailidentity.Status("pending") && !identity.SendAllowed && identity.EngineIdentityID == nil
}

func mailboxSetupStatusActivePrimaryIdentityValid(identity *ent.MailIdentity, accountID uuid.UUID, expectedAddress string) bool {
	return mailboxSetupStatusPrimaryIdentityValid(identity, accountID, expectedAddress) && identity.Status == mailidentity.StatusActive && identity.SendAllowed && identity.EngineIdentityID != nil && validMailboxOAuthOpaque(*identity.EngineIdentityID, MaxEngineOpaqueIDBytes)
}

func mailboxSetupStatusWithoutAddress(stage MailboxSetupStage) MailboxSetupStatus {
	return MailboxSetupStatus{Stage: stage}
}

func mailboxSetupStatusWithAddress(stage MailboxSetupStage, primaryAddress string) MailboxSetupStatus {
	return MailboxSetupStatus{Stage: stage, PrimaryAddress: primaryAddress, HasPrimaryAddress: true}
}

func mailboxSetupStatusActionRequired() MailboxSetupStatus {
	return mailboxSetupStatusWithoutAddress(MailboxSetupStageActionRequired)
}

func mailboxSetupStatusCopy(status MailboxSetupStatus) MailboxSetupStatus {
	return MailboxSetupStatus{Stage: status.Stage, PrimaryAddress: status.PrimaryAddress, HasPrimaryAddress: status.HasPrimaryAddress}
}

func mailboxSetupStatusDomainIntegrityValid(domain *ent.MailDomain, organizationID uuid.UUID) bool {
	if domain == nil || organizationID == uuid.Nil || domain.ID == uuid.Nil || domain.OrganizationID != organizationID || !validMailboxBootstrapIdentityExpectedEmail("mailbox-status@"+domain.Domain) || !mailboxEntValidTime(domain.CreatedAt) || !mailboxEntValidTime(domain.UpdatedAt) || domain.DesiredGeneration < 1 || domain.DesiredHashVersion < 1 || len(domain.DesiredHash) != sha256.Size || domain.AttemptCount < 0 || !mailboxSetupStatusDomainStateValid(domain.State) || !mailboxSetupStatusDomainVerificationValid(domain.VerificationStatus) {
		return false
	}
	if domain.EngineDomainID != nil && !validMailboxOAuthOpaque(*domain.EngineDomainID, MaxEngineOpaqueIDBytes) {
		return false
	}
	if domain.DeletedAt != nil && !mailboxEntValidTime(*domain.DeletedAt) || domain.NextRetryAt != nil && !mailboxEntValidTime(*domain.NextRetryAt) {
		return false
	}
	if !mailboxSetupStatusDomainTupleValid(domain.AttemptedGeneration, domain.AttemptedHashVersion, domain.AttemptedHash, domain.DesiredGeneration) || !mailboxSetupStatusDomainTupleValid(domain.DeployedGeneration, domain.DeployedHashVersion, domain.DeployedHash, domain.DesiredGeneration) || !mailboxSetupStatusDomainLeaseValid(domain) {
		return false
	}
	return domain.LastErrorCode == nil || mailboxSetupStatusDomainErrorCodeValid(*domain.LastErrorCode)
}

func mailboxSetupStatusDomainStateValid(state maildomain.State) bool {
	switch state {
	case maildomain.State("pending"), maildomain.State("provisioning"), maildomain.State("dns_required"), maildomain.State("verified"), maildomain.State("error"), maildomain.State("deleting"):
		return true
	default:
		return false
	}
}

func mailboxSetupStatusDomainVerificationValid(status maildomain.VerificationStatus) bool {
	return status == maildomain.VerificationStatus("pending") || status == maildomain.VerificationStatus("verified") || status == maildomain.VerificationStatus("failed")
}

func mailboxSetupStatusDomainTupleValid(generation *int64, version *int32, hash *[]byte, maximumGeneration int64) bool {
	return mailboxSetupStatusControlTupleValid(generation, version, hash, maximumGeneration)
}

func mailboxSetupStatusDomainLeaseValid(domain *ent.MailDomain) bool {
	if domain == nil || (domain.LeaseToken == nil) != (domain.LeaseUntil == nil) {
		return false
	}
	return domain.LeaseToken == nil || *domain.LeaseToken != uuid.Nil && mailboxEntValidTime(*domain.LeaseUntil)
}

func mailboxSetupStatusDomainErrorCodeValid(code maildomain.LastErrorCode) bool {
	switch code {
	case maildomain.LastErrorCode("MAIL_DOMAIN_ENGINE_UNAVAILABLE"), maildomain.LastErrorCode("MAIL_DOMAIN_APPLY_REJECTED"), maildomain.LastErrorCode("MAIL_DOMAIN_OUTCOME_UNKNOWN"), maildomain.LastErrorCode("MAIL_DOMAIN_DNS_INVALID"):
		return true
	default:
		return false
	}
}

func mailboxSetupStatusDomainHardFailure(domain *ent.MailDomain, organizationID uuid.UUID) bool {
	if domain == nil || organizationID == uuid.Nil || domain.OrganizationID != organizationID || !domain.DesiredPresent || !domain.Enabled || !domain.IsDefault || domain.DeletedAt != nil || domain.LastErrorCode != nil || domain.VerificationStatus == maildomain.VerificationStatus("failed") {
		return true
	}
	switch domain.State {
	case maildomain.State("error"), maildomain.State("deleting"):
		return true
	default:
		return false
	}
}

func mailboxSetupStatusActorLookupError(err error) error {
	if ent.IsNotFound(err) {
		return mailboxServiceError(MailboxServiceErrorNotFound)
	}
	return mailboxEntSafeError(err)
}

var _ MailboxSetupStatusSource = (*MailboxEntRepository)(nil)
