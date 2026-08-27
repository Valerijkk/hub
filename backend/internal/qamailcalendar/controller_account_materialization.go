//go:build qa && !release

package qamailcalendar

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/user"
)

var (
	errAccountMaterializationFixtureControllerInput = errors.New("qa mail account materialization controller input is invalid")
	errAccountMaterializationFixtureLifecycle       = errors.New("qa mail account materialization fixture lifecycle failed")
)

const accountMaterializationFixtureScenario Scenario = "account-materialization-v1"

type accountMaterializationFixtureState string

const (
	accountMaterializationFixtureStateCreating      accountMaterializationFixtureState = "creating"
	accountMaterializationFixtureStateCreated       accountMaterializationFixtureState = "created"
	accountMaterializationFixtureStateDeleting      accountMaterializationFixtureState = "deleting"
	accountMaterializationFixtureStateCleanupFailed accountMaterializationFixtureState = "cleanup_failed"
)

type accountMaterializationState string

const (
	accountMaterializationStateNotMaterialized    accountMaterializationState = "not_materialized"
	accountMaterializationStateMaterializedPending accountMaterializationState = "materialized_pending"
	accountMaterializationStateFinalized          accountMaterializationState = "finalized"
)

type accountMaterializationFixtureRecord struct {
	nonce     string
	plan      accountMaterializationFixturePlan
	persisted accountMaterializationFixturePersisted
	state     accountMaterializationFixtureState
}

type accountMaterializationFixtureSnapshot struct {
	nonce                 string
	scenario              Scenario
	state                 accountMaterializationFixtureState
	ownerEmail            string
	organizationID        uuid.UUID
	ownerUserID           uuid.UUID
	mailDomainID          uuid.UUID
	expectedMailAccountID uuid.UUID
	materializationState  accountMaterializationState
	mailAccountID         *uuid.UUID
	primaryIdentityID     *uuid.UUID
	primaryIdentityStatus *string
}

func (c *Controller) createAccountMaterializationFixture(
	ctx context.Context,
) (accountMaterializationFixtureSnapshot, error) {
	if c == nil ||
		ctx == nil ||
		c.client == nil ||
		c.random == nil ||
		c.fixtures == nil ||
		c.accountMaterializationFixtures == nil ||
		!validToken(string(c.config.token)) {
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureControllerInput
	}

	nonce, reserved := c.reserveAccountMaterializationFixture()
	if !reserved {
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
	}

	plan, err := newAccountMaterializationFixturePlan(nonce)
	if err != nil {
		if !c.removeCreatingAccountMaterializationFixture(nonce) {
			return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
		}
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
	}
	if !c.storeCreatingAccountMaterializationFixturePlan(nonce, plan) {
		if !c.removeCreatingAccountMaterializationFixture(nonce) {
			return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
		}
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
	}

	persisted, err := persistAccountMaterializationFixture(ctx, c.client, plan)
	if err != nil {
		if !c.removeCreatingAccountMaterializationFixture(nonce) {
			return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
		}
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
	}
	snapshot, completed := c.completeAccountMaterializationFixture(nonce, plan, persisted)
	if !completed {
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
	}
	return snapshot, nil
}

func (c *Controller) reserveAccountMaterializationFixture() (string, bool) {
	for attempt := 0; attempt < 4; attempt++ {
		var randomBytes [16]byte
		if _, err := io.ReadFull(c.random, randomBytes[:]); err != nil {
			return "", false
		}
		nonce := hex.EncodeToString(randomBytes[:])

		c.mu.Lock()
		_, inboundExists := c.fixtures[nonce]
		_, accountMaterializationExists := c.accountMaterializationFixtures[nonce]
		if !inboundExists && !accountMaterializationExists {
			c.accountMaterializationFixtures[nonce] = accountMaterializationFixtureRecord{
				nonce: nonce,
				state: accountMaterializationFixtureStateCreating,
			}
			c.mu.Unlock()
			return nonce, true
		}
		c.mu.Unlock()
	}

	return "", false
}

func (c *Controller) removeCreatingAccountMaterializationFixture(nonce string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	record, found := c.accountMaterializationFixtures[nonce]
	if !found ||
		record.nonce != nonce ||
		record.state != accountMaterializationFixtureStateCreating {
		return false
	}
	delete(c.accountMaterializationFixtures, nonce)
	return true
}

func (c *Controller) storeCreatingAccountMaterializationFixturePlan(
	nonce string,
	plan accountMaterializationFixturePlan,
) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	record, found := c.accountMaterializationFixtures[nonce]
	if !found ||
		record.nonce != nonce ||
		record.state != accountMaterializationFixtureStateCreating {
		return false
	}
	record.plan = plan
	c.accountMaterializationFixtures[nonce] = record
	return true
}

func (c *Controller) completeAccountMaterializationFixture(
	nonce string,
	plan accountMaterializationFixturePlan,
	persisted accountMaterializationFixturePersisted,
) (accountMaterializationFixtureSnapshot, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	record, found := c.accountMaterializationFixtures[nonce]
	if !found ||
		record.nonce != nonce ||
		record.state != accountMaterializationFixtureStateCreating ||
		record.plan != plan ||
		!accountMaterializationFixturePersistedMatchesPlan(persisted, plan) {
		return accountMaterializationFixtureSnapshot{}, false
	}
	record.persisted = persisted
	record.state = accountMaterializationFixtureStateCreated
	c.accountMaterializationFixtures[nonce] = record

	return accountMaterializationFixtureSnapshot{
		nonce:                 record.nonce,
		scenario:              accountMaterializationFixtureScenario,
		state:                 record.state,
		ownerEmail:            record.plan.ownerAddress,
		organizationID:        record.persisted.organizationID,
		ownerUserID:           record.persisted.ownerUserID,
		mailDomainID:          record.persisted.mailDomainID,
		expectedMailAccountID: record.plan.expectedAccountIntentID,
		materializationState:  accountMaterializationStateNotMaterialized,
		mailAccountID:         nil,
		primaryIdentityID:     nil,
		primaryIdentityStatus: nil,
	}, true
}

func accountMaterializationFixturePersistedMatchesPlan(
	persisted accountMaterializationFixturePersisted,
	plan accountMaterializationFixturePlan,
) bool {
	return persisted.organizationID == plan.organizationID &&
		persisted.ownerUserID == plan.ownerUserID &&
		persisted.mailDomainID == plan.mailDomainID
}

func accountMaterializationFixtureValidTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond() == 0
}

func accountMaterializationFixtureHashEqual(actual, expected []byte) bool {
	return len(actual) == sha256.Size &&
		len(expected) == sha256.Size &&
		subtle.ConstantTimeCompare(actual, expected) == 1
}

func accountMaterializationFixtureOptionalTupleValid(
	generation *int64,
	hashVersion *int32,
	hash []byte,
) bool {
	if generation == nil || hashVersion == nil || hash == nil {
		return generation == nil && hashVersion == nil && hash == nil
	}
	return *generation >= 1 && *hashVersion >= 1 && len(hash) == sha256.Size
}

func accountMaterializationFixtureOptionalTupleMatches(
	generation *int64,
	hashVersion *int32,
	hash []byte,
	expectedGeneration int64,
	expectedHashVersion int32,
	expectedHash []byte,
) bool {
	return accountMaterializationFixtureOptionalTupleValid(generation, hashVersion, hash) &&
		generation != nil &&
		*generation == expectedGeneration &&
		*hashVersion == expectedHashVersion &&
		accountMaterializationFixtureHashEqual(hash, expectedHash)
}

func accountMaterializationFixtureExactRootsValid(
	record accountMaterializationFixtureRecord,
	organization *ent.Organization,
	owner *ent.User,
	domain *ent.MailDomain,
) bool {
	plan := record.plan
	if record.nonce != plan.nonce ||
		record.state != accountMaterializationFixtureStateCreated ||
		!accountMaterializationFixturePlanMatchesCanonical(plan) ||
		!accountMaterializationFixturePersistedMatchesPlan(record.persisted, plan) ||
		organization == nil ||
		organization.ID != plan.organizationID ||
		organization.Slug != plan.organizationSlug ||
		organization.Name != "QA Mail Materialization "+plan.nonce ||
		organization.InviteToken != "qa-mail-materialization-invite-"+plan.nonce ||
		!accountMaterializationFixtureValidTime(organization.CreatedAt) ||
		!accountMaterializationFixtureValidTime(organization.UpdatedAt) ||
		!organization.CreatedAt.Equal(plan.fixedAt) ||
		!organization.UpdatedAt.Equal(plan.fixedAt) ||
		owner == nil ||
		owner.ID != plan.ownerUserID ||
		owner.Email != plan.ownerAddress ||
		owner.OrganizationID != plan.organizationID ||
		owner.SystemRole != user.SystemRoleMember ||
		owner.Status != user.StatusActive ||
		owner.Timezone != "UTC" ||
		owner.Language != "ru" ||
		!owner.OnboardingCompleted ||
		owner.Name != "QA Owner" ||
		owner.DisplayName != "QA Owner" ||
		!accountMaterializationFixtureValidTime(owner.CreatedAt) ||
		!accountMaterializationFixtureValidTime(owner.UpdatedAt) ||
		!owner.CreatedAt.Equal(plan.fixedAt) ||
		!owner.UpdatedAt.Equal(plan.fixedAt) ||
		domain == nil ||
		domain.ID != plan.mailDomainID ||
		domain.OrganizationID != plan.organizationID ||
		domain.Domain != plan.domain ||
		domain.EngineProvisioningKey != plan.domainProvisioningKey ||
		domain.EngineDomainID == nil ||
		*domain.EngineDomainID != plan.engineDomainID ||
		!domain.DesiredPresent ||
		!domain.Enabled ||
		!domain.IsDefault ||
		domain.State != maildomain.State("verified") ||
		domain.VerificationStatus != maildomain.VerificationStatus("verified") ||
		domain.DKIMStatus != maildomain.DKIMStatus("valid") ||
		domain.RequiredDNSRecords == nil ||
		len(domain.RequiredDNSRecords) != 0 ||
		domain.DesiredGeneration != 1 ||
		domain.DesiredHashVersion != 1 ||
		!accountMaterializationFixtureHashEqual(domain.DesiredHash, plan.domainControlHash[:]) ||
		!accountMaterializationFixtureOptionalTupleValid(domain.AttemptedGeneration, domain.AttemptedHashVersion, domain.AttemptedHash) ||
		domain.AttemptedGeneration != nil ||
		!accountMaterializationFixtureOptionalTupleMatches(
			domain.DeployedGeneration,
			domain.DeployedHashVersion,
			domain.DeployedHash,
			1,
			1,
			plan.domainControlHash[:],
		) ||
		domain.LastErrorCode != nil ||
		domain.NextRetryAt != nil ||
		domain.AttemptCount != 0 ||
		domain.LeaseToken != nil ||
		domain.LeaseUntil != nil ||
		domain.DeletedAt != nil ||
		!accountMaterializationFixtureValidTime(domain.CreatedAt) ||
		!accountMaterializationFixtureValidTime(domain.UpdatedAt) ||
		!domain.CreatedAt.Equal(plan.fixedAt) ||
		!domain.UpdatedAt.Equal(plan.fixedAt) {
		return false
	}
	return true
}

func accountMaterializationFixtureCommonAccountIdentityValid(
	plan accountMaterializationFixturePlan,
	account *ent.MailAccount,
	identity *ent.MailIdentity,
) bool {
	if !accountMaterializationFixturePlanMatchesCanonical(plan) ||
		account == nil ||
		account.ID != plan.expectedAccountIntentID ||
		account.OrganizationID != plan.organizationID ||
		account.UserID != plan.ownerUserID ||
		account.MailDomainID != plan.mailDomainID ||
		account.EngineProvisioningKey != plan.expectedAccountProvisioningKey ||
		account.LocalPart != plan.ownerLocalPart ||
		!account.DesiredEnabled ||
		account.Status != mailaccount.Status("provisioning") ||
		account.ControlGeneration != 1 ||
		account.ControlHashVersion != 1 ||
		!accountMaterializationFixtureHashEqual(account.ControlHash, plan.expectedAccountControlHash[:]) ||
		account.ProjectionState != mailaccount.ProjectionState("bootstrapping") ||
		!accountMaterializationFixtureValidTime(account.EventCutoffAt) ||
		!accountMaterializationFixtureValidTime(account.SyncScheduleUpdatedAt) ||
		!accountMaterializationFixtureValidTime(account.CreatedAt) ||
		!accountMaterializationFixtureValidTime(account.UpdatedAt) ||
		account.ProvisionAttempts < 0 ||
		account.PasswordEnrollmentGeneration != 0 ||
		account.PasswordEnrolledEngineAccountID != nil ||
		account.PasswordEnrolledAt != nil ||
		account.PasswordEnrollmentLeaseToken != nil ||
		account.PasswordEnrollmentLeaseUntil != nil ||
		!accountMaterializationFixtureOptionalTupleValid(
			account.AttemptedControlGeneration,
			account.AttemptedControlHashVersion,
			account.AttemptedControlHash,
		) ||
		!accountMaterializationFixtureOptionalTupleValid(
			account.DeployedGeneration,
			account.DeployedHashVersion,
			account.DeployedHash,
		) ||
		identity == nil ||
		identity.ID == uuid.Nil ||
		identity.MailAccountID != account.ID ||
		identity.NormalizedAddress != plan.ownerAddress ||
		identity.Kind != mailidentity.KindPrimary ||
		identity.Status != mailidentity.Status("pending") ||
		identity.SendAllowed ||
		identity.EngineIdentityID != nil {
		return false
	}
	return true
}

func accountMaterializationFixtureClassifyRows(
	record accountMaterializationFixtureRecord,
	organization *ent.Organization,
	owner *ent.User,
	domain *ent.MailDomain,
	accounts []*ent.MailAccount,
	identities []*ent.MailIdentity,
) (accountMaterializationFixtureSnapshot, bool) {
	if !accountMaterializationFixtureExactRootsValid(record, organization, owner, domain) {
		return accountMaterializationFixtureSnapshot{}, false
	}
	if len(accounts) == 0 {
		if len(identities) != 0 {
			return accountMaterializationFixtureSnapshot{}, false
		}
		return accountMaterializationFixtureSnapshot{
			nonce:                 record.nonce,
			scenario:              accountMaterializationFixtureScenario,
			state:                 record.state,
			ownerEmail:            record.plan.ownerAddress,
			organizationID:        record.persisted.organizationID,
			ownerUserID:           record.persisted.ownerUserID,
			mailDomainID:          record.persisted.mailDomainID,
			expectedMailAccountID: record.plan.expectedAccountIntentID,
			materializationState:  accountMaterializationStateNotMaterialized,
		}, true
	}
	if len(accounts) != 1 || len(identities) != 1 {
		return accountMaterializationFixtureSnapshot{}, false
	}

	account := accounts[0]
	identity := identities[0]
	if !accountMaterializationFixtureCommonAccountIdentityValid(record.plan, account, identity) {
		return accountMaterializationFixtureSnapshot{}, false
	}
	materializationState, valid := accountMaterializationFixtureClassifyAccountState(record.plan, account)
	if !valid {
		return accountMaterializationFixtureSnapshot{}, false
	}

	mailAccountID := account.ID
	primaryIdentityID := identity.ID
	primaryIdentityStatus := string(identity.Status)
	return accountMaterializationFixtureSnapshot{
		nonce:                 record.nonce,
		scenario:              accountMaterializationFixtureScenario,
		state:                 record.state,
		ownerEmail:            record.plan.ownerAddress,
		organizationID:        record.persisted.organizationID,
		ownerUserID:           record.persisted.ownerUserID,
		mailDomainID:          record.persisted.mailDomainID,
		expectedMailAccountID: record.plan.expectedAccountIntentID,
		materializationState:  materializationState,
		mailAccountID:         &mailAccountID,
		primaryIdentityID:     &primaryIdentityID,
		primaryIdentityStatus: &primaryIdentityStatus,
	}, true
}

func accountMaterializationFixtureClassifyAccountState(
	plan accountMaterializationFixturePlan,
	account *ent.MailAccount,
) (accountMaterializationState, bool) {
	if account == nil {
		return "", false
	}
	if account.EngineAccountID == nil &&
		accountMaterializationFixtureOptionalTupleAbsent(
			account.DeployedGeneration,
			account.DeployedHashVersion,
			account.DeployedHash,
		) {
		switch {
		case accountMaterializationFixtureFreshAccountState(account):
			return accountMaterializationStateMaterializedPending, true
		case accountMaterializationFixtureClaimedAccountState(plan, account):
			return accountMaterializationStateMaterializedPending, true
		case accountMaterializationFixtureRetryAccountState(plan, account):
			return accountMaterializationStateMaterializedPending, true
		}
	}
	if account.EngineAccountID != nil &&
		*account.EngineAccountID == plan.expectedEngineAccountID &&
		accountMaterializationFixtureOptionalTupleMatches(
			account.DeployedGeneration,
			account.DeployedHashVersion,
			account.DeployedHash,
			account.ControlGeneration,
			account.ControlHashVersion,
			account.ControlHash,
		) &&
		accountMaterializationFixtureOptionalTupleAbsent(
			account.AttemptedControlGeneration,
			account.AttemptedControlHashVersion,
			account.AttemptedControlHash,
		) &&
		account.ProvisionLeaseToken == nil &&
		account.ProvisionLeaseUntil == nil &&
		account.NextProvisionAt == nil &&
		account.ProvisionErrorCode == nil &&
		account.ProvisionAttempts >= 1 {
		return accountMaterializationStateFinalized, true
	}
	return "", false
}

func accountMaterializationFixtureOptionalTupleAbsent(
	generation *int64,
	hashVersion *int32,
	hash []byte,
) bool {
	return generation == nil && hashVersion == nil && hash == nil
}

func accountMaterializationFixtureFreshAccountState(account *ent.MailAccount) bool {
	return accountMaterializationFixtureOptionalTupleAbsent(
		account.AttemptedControlGeneration,
		account.AttemptedControlHashVersion,
		account.AttemptedControlHash,
	) &&
		account.ProvisionLeaseToken == nil &&
		account.ProvisionLeaseUntil == nil &&
		account.NextProvisionAt != nil &&
		accountMaterializationFixtureValidTime(*account.NextProvisionAt) &&
		account.ProvisionErrorCode == nil &&
		account.ProvisionAttempts == 0
}

func accountMaterializationFixtureClaimedAccountState(
	plan accountMaterializationFixturePlan,
	account *ent.MailAccount,
) bool {
	return accountMaterializationFixtureOptionalTupleMatches(
		account.AttemptedControlGeneration,
		account.AttemptedControlHashVersion,
		account.AttemptedControlHash,
		account.ControlGeneration,
		account.ControlHashVersion,
		account.ControlHash,
	) &&
		account.ProvisionLeaseToken != nil &&
		*account.ProvisionLeaseToken != uuid.Nil &&
		account.ProvisionLeaseUntil != nil &&
		accountMaterializationFixtureValidTime(*account.ProvisionLeaseUntil) &&
		account.NextProvisionAt == nil &&
		account.ProvisionErrorCode == nil &&
		account.ProvisionAttempts >= 1 &&
		accountMaterializationFixturePlanMatchesCanonical(plan)
}

func accountMaterializationFixtureRetryAccountState(
	plan accountMaterializationFixturePlan,
	account *ent.MailAccount,
) bool {
	if !accountMaterializationFixtureOptionalTupleMatches(
		account.AttemptedControlGeneration,
		account.AttemptedControlHashVersion,
		account.AttemptedControlHash,
		account.ControlGeneration,
		account.ControlHashVersion,
		account.ControlHash,
	) ||
		account.ProvisionLeaseToken != nil ||
		account.ProvisionLeaseUntil != nil ||
		account.NextProvisionAt == nil ||
		!accountMaterializationFixtureValidTime(*account.NextProvisionAt) ||
		account.ProvisionErrorCode == nil ||
		account.ProvisionAttempts < 1 ||
		!accountMaterializationFixturePlanMatchesCanonical(plan) {
		return false
	}
	switch *account.ProvisionErrorCode {
	case mailaccount.ProvisionErrorCode("MAILBOX_PROVISIONING_RETRY"),
		mailaccount.ProvisionErrorCode("MAILBOX_UNAVAILABLE"):
		return true
	default:
		return false
	}
}
