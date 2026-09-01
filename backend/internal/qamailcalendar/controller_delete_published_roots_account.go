//go:build qa && !release

package qamailcalendar

import (
	"bytes"
	"context"

	"entgo.io/ent/dialect/sql"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/predicate"
)

// inboundRequestFixtureLockedAccountRoots contains rows locked by the
// caller-owned transaction. It must not escape that transaction or enter a
// Controller record.
type inboundRequestFixtureLockedAccountRoots struct {
	roots           inboundRequestFixtureRootProof
	account         *ent.MailAccount
	primaryIdentity *ent.MailIdentity
	domain          *ent.MailDomain
}

// lockInboundRequestFixturePublishedAccountRoots locks and proves only the
// MailAccount -> MailIdentity -> MailDomain part of the deterministic root
// graph. Its caller owns SERIALIZABLE isolation, commit, and rollback.
func lockInboundRequestFixturePublishedAccountRoots(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	authority inboundRequestFixturePublishedEngineAuthority,
	mailRoots inboundRequestFixtureLockedMailRoots,
) (inboundRequestFixtureLockedAccountRoots, bool) {
	roots, valid := newInboundRequestFixtureRootProof(plan, persisted)
	if ctx == nil ||
		tx == nil ||
		!valid ||
		!inboundRequestFixturePublishedAccountRootsInputValid(plan, persisted, authority, roots, mailRoots) {
		return inboundRequestFixtureLockedAccountRoots{}, false
	}

	accounts, err := tx.MailAccount.Query().
		Where(
			mailaccount.MailDomainIDEQ(roots.mailDomainID),
			inboundRequestFixturePublishedMailAccountNoWait(),
		).
		Order(ent.Asc(mailaccount.FieldID)).
		All(ctx)
	if err != nil || len(accounts) != 1 ||
		!inboundRequestFixturePublishedAccountValid(accounts[0], roots, plan) {
		return inboundRequestFixtureLockedAccountRoots{}, false
	}
	account := accounts[0]

	identities, err := tx.MailIdentity.Query().
		Where(
			mailidentity.MailAccountIDEQ(account.ID),
			inboundRequestFixturePublishedMailIdentityNoWait(),
		).
		Order(ent.Asc(mailidentity.FieldID)).
		All(ctx)
	if err != nil || len(identities) != 1 ||
		!inboundRequestFixturePublishedPrimaryIdentityValid(identities[0], roots, plan) {
		return inboundRequestFixtureLockedAccountRoots{}, false
	}
	primaryIdentity := identities[0]

	domain, err := tx.MailDomain.Query().
		Where(
			maildomain.IDEQ(roots.mailDomainID),
			maildomain.OrganizationIDEQ(roots.organizationID),
			inboundRequestFixturePublishedMailDomainNoWait(),
		).
		Only(ctx)
	if err != nil || !inboundRequestFixturePublishedDomainValid(domain, roots, plan) ||
		account.EngineAccountID == nil ||
		mailRoots.folder.EngineFolderID == nil ||
		primaryIdentity.EngineIdentityID == nil ||
		!authority.matchesAccountFolderIdentity(
			*account.EngineAccountID,
			*mailRoots.folder.EngineFolderID,
			*primaryIdentity.EngineIdentityID,
		) {
		return inboundRequestFixtureLockedAccountRoots{}, false
	}

	return inboundRequestFixtureLockedAccountRoots{
		roots:           roots,
		account:         account,
		primaryIdentity: primaryIdentity,
		domain:          domain,
	}, true
}

func inboundRequestFixturePublishedAccountRootsInputValid(
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	authority inboundRequestFixturePublishedEngineAuthority,
	roots inboundRequestFixtureRootProof,
	mailRoots inboundRequestFixtureLockedMailRoots,
) bool {
	canonicalPlan, err := newInboundRequestFixturePlan(plan.nonce)
	if err != nil ||
		plan != canonicalPlan ||
		!inboundRequestFixturePersistedMatchesPlan(persisted, plan) ||
		!authority.valid() ||
		authority.nonce != plan.nonce ||
		mailRoots.roots != roots ||
		mailRoots.messageFolder == nil ||
		mailRoots.message == nil ||
		mailRoots.thread == nil ||
		mailRoots.folder == nil ||
		mailRoots.messageFolder.ID != roots.mailMessageFolderID ||
		mailRoots.message.ID != roots.mailMessageID ||
		mailRoots.thread.ID != roots.mailThreadID ||
		mailRoots.folder.ID != roots.mailFolderID ||
		mailRoots.messageFolder.MailMessageID != roots.mailMessageID ||
		mailRoots.messageFolder.MailFolderID != roots.mailFolderID ||
		mailRoots.message.MailAccountID != roots.mailAccountID ||
		mailRoots.message.MailThreadID != roots.mailThreadID ||
		mailRoots.thread.MailAccountID != roots.mailAccountID ||
		mailRoots.folder.MailAccountID != roots.mailAccountID {
		return false
	}
	return true
}

func inboundRequestFixturePublishedAccountValid(
	account *ent.MailAccount,
	roots inboundRequestFixtureRootProof,
	plan inboundRequestFixturePlan,
) bool {
	return account != nil &&
		account.ID == roots.mailAccountID &&
		account.OrganizationID == roots.organizationID &&
		account.UserID == roots.ownerUserID &&
		account.MailDomainID == roots.mailDomainID &&
		account.EngineProvisioningKey == plan.accountProvisioningKey &&
		account.LocalPart == "owner" &&
		account.EngineAccountID != nil &&
		account.PasswordEnrollmentGeneration == 0 &&
		account.PasswordEnrolledEngineAccountID == nil &&
		account.PasswordEnrolledAt == nil &&
		account.PasswordEnrollmentLeaseToken == nil &&
		account.PasswordEnrollmentLeaseUntil == nil &&
		account.DesiredEnabled &&
		account.Status == mailaccount.StatusActive &&
		account.ControlGeneration == 1 &&
		account.ControlHashVersion == 1 &&
		bytes.Equal(account.ControlHash, plan.accountControlHash[:]) &&
		account.AttemptedControlGeneration == nil &&
		account.AttemptedControlHashVersion == nil &&
		inboundRequestFixturePublishedNullableBytesAbsent(account.AttemptedControlHash) &&
		account.DeployedGeneration != nil &&
		*account.DeployedGeneration == 1 &&
		account.DeployedHashVersion != nil &&
		*account.DeployedHashVersion == 1 &&
		account.DeployedHash != nil &&
		bytes.Equal(*account.DeployedHash, plan.accountControlHash[:]) &&
		account.ProjectionState == mailaccount.ProjectionState("ready") &&
		account.SyncCursor != nil &&
		*account.SyncCursor == plan.correlationID &&
		account.RebuildCheckpoint == nil &&
		account.RebuildBaseCursor == nil &&
		account.SyncEpoch == 0 &&
		account.SyncLeaseEpoch == 1 &&
		account.MutationEpoch == 0 &&
		account.ProvisionAttempts == 0 &&
		account.SyncAttempts == 0 &&
		account.ProvisionLeaseToken == nil &&
		account.SyncLeaseToken == nil &&
		account.SyncLeaseSyncEpoch == nil &&
		account.SyncLeaseMutationEpoch == nil &&
		account.SyncLeasePhase == nil &&
		account.SyncLeaseCursor == nil &&
		account.SyncLeaseBaseCursor == nil &&
		account.SyncLeaseCheckpoint == nil &&
		account.SyncLeaseResultHashVersion == nil &&
		inboundRequestFixturePublishedNullableBytesAbsent(account.SyncLeaseResultHash) &&
		account.SyncLeaseAppliedPhase == nil &&
		account.SyncLeaseAppliedNextAt == nil &&
		account.SyncLeaseAppliedAt == nil &&
		account.MutationLeaseToken == nil &&
		account.ProvisionLeaseUntil == nil &&
		account.SyncLeaseUntil == nil &&
		account.MutationLeaseUntil == nil &&
		account.NextProvisionAt == nil &&
		account.NextSyncAt == nil &&
		account.SyncScheduleUpdatedAt.Equal(plan.fixedAt) &&
		account.LastSyncedAt != nil &&
		account.LastSyncedAt.Equal(plan.fixedAt) &&
		account.ProvisionErrorCode == nil &&
		account.SyncErrorCode == nil &&
		account.EventCutoffAt.Equal(plan.fixedAt) &&
		account.CreatedAt.Equal(plan.fixedAt) &&
		account.UpdatedAt.Equal(plan.fixedAt)
}

func inboundRequestFixturePublishedPrimaryIdentityValid(
	identity *ent.MailIdentity,
	roots inboundRequestFixtureRootProof,
	plan inboundRequestFixturePlan,
) bool {
	return identity != nil &&
		identity.ID == roots.mailIdentityID &&
		identity.MailAccountID == roots.mailAccountID &&
		identity.NormalizedAddress == plan.ownerAddress &&
		identity.DisplayName == nil &&
		identity.Kind == mailidentity.KindPrimary &&
		identity.SendAllowed &&
		identity.Status == mailidentity.StatusActive &&
		identity.EngineIdentityID != nil &&
		identity.CreatedAt.Equal(plan.fixedAt) &&
		identity.UpdatedAt.Equal(plan.fixedAt)
}

func inboundRequestFixturePublishedDomainValid(
	domain *ent.MailDomain,
	roots inboundRequestFixtureRootProof,
	plan inboundRequestFixturePlan,
) bool {
	return domain != nil &&
		domain.ID == roots.mailDomainID &&
		domain.OrganizationID == roots.organizationID &&
		domain.Domain == plan.domain &&
		domain.EngineProvisioningKey == plan.domainProvisioningKey &&
		domain.EngineDomainID == nil &&
		domain.DesiredPresent &&
		domain.Enabled &&
		!domain.IsDefault &&
		domain.State == maildomain.State("pending") &&
		domain.VerificationStatus == maildomain.VerificationStatus("pending") &&
		domain.DkimStatus == maildomain.DkimStatus("unknown") &&
		domain.RequiredDNSRecords != nil &&
		len(domain.RequiredDNSRecords) == 0 &&
		domain.DesiredGeneration == 1 &&
		domain.DesiredHashVersion == 1 &&
		bytes.Equal(domain.DesiredHash, plan.domainControlHash[:]) &&
		domain.AttemptedGeneration == nil &&
		domain.AttemptedHashVersion == nil &&
		inboundRequestFixturePublishedNullableBytesAbsent(domain.AttemptedHash) &&
		domain.DeployedGeneration == nil &&
		domain.DeployedHashVersion == nil &&
		inboundRequestFixturePublishedNullableBytesAbsent(domain.DeployedHash) &&
		domain.LastErrorCode == nil &&
		domain.NextRetryAt == nil &&
		domain.AttemptCount == 0 &&
		domain.LeaseToken == nil &&
		domain.LeaseUntil == nil &&
		domain.DeletedAt == nil &&
		domain.Version == 1 &&
		domain.CreatedAt.Equal(plan.fixedAt) &&
		domain.UpdatedAt.Equal(plan.fixedAt)
}

func inboundRequestFixturePublishedNullableBytesAbsent(value *[]byte) bool {
	return value == nil || *value == nil
}

func inboundRequestFixturePublishedMailAccountNoWait() predicate.MailAccount {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailIdentityNoWait() predicate.MailIdentity {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailDomainNoWait() predicate.MailDomain {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
