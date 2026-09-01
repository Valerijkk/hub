//go:build qa && !release

package qamailcalendar

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

var (
	errAccountMaterializationFixtureControllerInput = errors.New("qa mail account materialization controller input is invalid")
	errAccountMaterializationFixtureLifecycle       = errors.New("qa mail account materialization fixture lifecycle failed")
)

const accountMaterializationFixtureScenario Scenario = "account-materialization-v1"

type accountMaterializationFixtureObserveStage string

const (
	accountMaterializationFixtureObserveStagePlan             accountMaterializationFixtureObserveStage = "plan"
	accountMaterializationFixtureObserveStageRecord           accountMaterializationFixtureObserveStage = "record"
	accountMaterializationFixtureObserveStageTransaction      accountMaterializationFixtureObserveStage = "transaction"
	accountMaterializationFixtureObserveStageOrganizationRead accountMaterializationFixtureObserveStage = "organization_read"
	accountMaterializationFixtureObserveStageOwnerRead        accountMaterializationFixtureObserveStage = "owner_read"
	accountMaterializationFixtureObserveStageDomainRead       accountMaterializationFixtureObserveStage = "domain_read"
	accountMaterializationFixtureObserveStageAccountRead      accountMaterializationFixtureObserveStage = "account_read"
	accountMaterializationFixtureObserveStageIdentityRead     accountMaterializationFixtureObserveStage = "identity_read"
	accountMaterializationFixtureObserveStageClassify         accountMaterializationFixtureObserveStage = "classify"
	accountMaterializationFixtureObserveStageCommit           accountMaterializationFixtureObserveStage = "commit"
)

type accountMaterializationFixtureObserveCode string

func accountMaterializationFixtureObserveLogFailure(
	ctx context.Context,
	stage accountMaterializationFixtureObserveStage,
	code accountMaterializationFixtureObserveCode,
) {
	slog.ErrorContext(
		ctx,
		"qa mail account materialization fixture observation failed",
		"stage", string(stage),
		"reason_code", string(code),
	)
}

type accountMaterializationFixtureState string

const (
	accountMaterializationFixtureStateCreating      accountMaterializationFixtureState = "creating"
	accountMaterializationFixtureStateCreated       accountMaterializationFixtureState = "created"
	accountMaterializationFixtureStateQuiescing     accountMaterializationFixtureState = "quiescing"
	accountMaterializationFixtureStateQuiesced      accountMaterializationFixtureState = "quiesced"
	accountMaterializationFixtureStateDeleting      accountMaterializationFixtureState = "deleting"
	accountMaterializationFixtureStateCleanupFailed accountMaterializationFixtureState = "cleanup_failed"
)

type accountMaterializationFixtureCleanupOrigin string

const (
	accountMaterializationFixtureCleanupOriginNone          accountMaterializationFixtureCleanupOrigin = ""
	accountMaterializationFixtureCleanupOriginQuiesceCommit accountMaterializationFixtureCleanupOrigin = "quiesce_commit"
	accountMaterializationFixtureCleanupOriginDeleteCommit  accountMaterializationFixtureCleanupOrigin = "delete_commit"
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
	cleanupOrigin accountMaterializationFixtureCleanupOrigin
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
		c.threadSearchFixtures == nil ||
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

func (c *Controller) observeAccountMaterializationFixture(
	ctx context.Context,
	nonce string,
) (accountMaterializationFixtureSnapshot, error) {
	if c == nil ||
		ctx == nil ||
		c.client == nil ||
		c.fixtures == nil ||
		c.accountMaterializationFixtures == nil ||
		!validToken(string(c.config.token)) ||
		!ValidNonce(nonce) {
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureControllerInput
	}

	canonicalPlan, err := newAccountMaterializationFixturePlan(nonce)
	if err != nil {
		accountMaterializationFixtureObserveLogFailure(
			ctx,
			accountMaterializationFixtureObserveStagePlan,
			"canonical_plan",
		)
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureControllerInput
	}

	c.mu.Lock()
	record, found := c.accountMaterializationFixtures[nonce]
	c.mu.Unlock()
	if code := accountMaterializationFixtureObserveRecordMismatchCode(record, found, nonce, canonicalPlan); code != "" {
		accountMaterializationFixtureObserveLogFailure(
			ctx,
			accountMaterializationFixtureObserveStageRecord,
			code,
		)
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
	}

	tx, err := c.client.Tx(ctx)
	if err != nil || tx == nil {
		accountMaterializationFixtureObserveLogFailure(
			ctx,
			accountMaterializationFixtureObserveStageTransaction,
			"open_failed",
		)
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	organizationRow, err := tx.Organization.Query().
		Where(
			organization.IDEQ(record.persisted.organizationID),
			organization.HasUsersWith(user.IDEQ(record.persisted.ownerUserID)),
		).
		Only(ctx)
	if err != nil {
		accountMaterializationFixtureObserveLogFailure(
			ctx,
			accountMaterializationFixtureObserveStageOrganizationRead,
			"lookup_failed",
		)
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
	}
	ownerRow, err := tx.User.Query().
		Where(user.IDEQ(record.persisted.ownerUserID)).
		Only(ctx)
	if err != nil {
		accountMaterializationFixtureObserveLogFailure(
			ctx,
			accountMaterializationFixtureObserveStageOwnerRead,
			"lookup_failed",
		)
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
	}
	domainRow, err := tx.MailDomain.Query().
		Where(maildomain.IDEQ(record.persisted.mailDomainID)).
		Only(ctx)
	if err != nil {
		accountMaterializationFixtureObserveLogFailure(
			ctx,
			accountMaterializationFixtureObserveStageDomainRead,
			"lookup_failed",
		)
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
	}
	accounts, err := tx.MailAccount.Query().
		Where(mailaccount.UserIDEQ(record.persisted.ownerUserID)).
		Order(mailaccount.ByID()).
		Limit(2).
		All(ctx)
	if err != nil {
		accountMaterializationFixtureObserveLogFailure(
			ctx,
			accountMaterializationFixtureObserveStageAccountRead,
			"lookup_failed",
		)
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
	}
	if len(accounts) > 1 {
		accountMaterializationFixtureObserveLogFailure(
			ctx,
			accountMaterializationFixtureObserveStageAccountRead,
			"multiple_rows",
		)
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
	}

	identities := make([]*ent.MailIdentity, 0)
	if len(accounts) == 1 {
		identities, err = tx.MailIdentity.Query().
			Where(mailidentity.MailAccountIDEQ(accounts[0].ID)).
			Order(mailidentity.ByID()).
			Limit(2).
			All(ctx)
		if err != nil {
			accountMaterializationFixtureObserveLogFailure(
				ctx,
				accountMaterializationFixtureObserveStageIdentityRead,
				"lookup_failed",
			)
			return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
		}
	}

	snapshot, code := accountMaterializationFixtureClassifyRowsCode(
		record,
		organizationRow,
		ownerRow,
		domainRow,
		accounts,
		identities,
	)
	if code != "" {
		accountMaterializationFixtureObserveLogFailure(
			ctx,
			accountMaterializationFixtureObserveStageClassify,
			code,
		)
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
	}
	if err := tx.Commit(); err != nil {
		accountMaterializationFixtureObserveLogFailure(
			ctx,
			accountMaterializationFixtureObserveStageCommit,
			"failed",
		)
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
	}
	committed = true

	c.mu.Lock()
	defer c.mu.Unlock()
	current, found := c.accountMaterializationFixtures[nonce]
	if !found ||
		current != record ||
		current.state != accountMaterializationFixtureStateCreated {
		accountMaterializationFixtureObserveLogFailure(
			ctx,
			accountMaterializationFixtureObserveStageRecord,
			"changed_after_commit",
		)
		return accountMaterializationFixtureSnapshot{}, errAccountMaterializationFixtureLifecycle
	}
	return snapshot, nil
}

func accountMaterializationFixtureObserveRecordMismatchCode(
	record accountMaterializationFixtureRecord,
	found bool,
	nonce string,
	canonicalPlan accountMaterializationFixturePlan,
) accountMaterializationFixtureObserveCode {
	switch {
	case !found:
		return "missing"
	case record.nonce != nonce:
		return "nonce"
	case record.state != accountMaterializationFixtureStateCreated:
		return "state"
	case record.plan != canonicalPlan:
		return "plan"
	case !accountMaterializationFixturePersistedMatchesPlan(record.persisted, canonicalPlan):
		return "persisted_roots"
	default:
		return ""
	}
}

func (c *Controller) reserveAccountMaterializationFixture() (string, bool) {
	if c == nil ||
		c.random == nil ||
		c.fixtures == nil ||
		c.accountMaterializationFixtures == nil ||
		c.threadSearchFixtures == nil {
		return "", false
	}

	for attempt := 0; attempt < 4; attempt++ {
		var randomBytes [16]byte
		if _, err := io.ReadFull(c.random, randomBytes[:]); err != nil {
			return "", false
		}
		nonce := hex.EncodeToString(randomBytes[:])

		c.mu.Lock()
		_, inboundExists := c.fixtures[nonce]
		_, accountMaterializationExists := c.accountMaterializationFixtures[nonce]
		_, threadSearchExists := c.threadSearchFixtures[nonce]
		if !inboundExists && !accountMaterializationExists && !threadSearchExists {
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
	return !value.IsZero() && value.Nanosecond() == 0
}

func accountMaterializationFixtureHashEqual(actual, expected []byte) bool {
	return len(actual) == sha256.Size &&
		len(expected) == sha256.Size &&
		subtle.ConstantTimeCompare(actual, expected) == 1
}

func accountMaterializationFixtureOptionalTupleValid(
	generation *int64,
	hashVersion *int32,
	hash *[]byte,
) bool {
	if generation == nil || hashVersion == nil || hash == nil || *hash == nil {
		return generation == nil && hashVersion == nil && (hash == nil || *hash == nil)
	}
	return *generation >= 1 && *hashVersion >= 1 && len(*hash) == sha256.Size
}

func accountMaterializationFixtureOptionalTupleMatches(
	generation *int64,
	hashVersion *int32,
	hash *[]byte,
	expectedGeneration int64,
	expectedHashVersion int32,
	expectedHash []byte,
) bool {
	return accountMaterializationFixtureOptionalTupleValid(generation, hashVersion, hash) &&
		generation != nil &&
		*generation == expectedGeneration &&
		*hashVersion == expectedHashVersion &&
		accountMaterializationFixtureHashEqual(*hash, expectedHash)
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
		domain.DkimStatus != maildomain.DkimStatus("valid") ||
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

func accountMaterializationFixtureClassifyRowsCode(
	record accountMaterializationFixtureRecord,
	organization *ent.Organization,
	owner *ent.User,
	domain *ent.MailDomain,
	accounts []*ent.MailAccount,
	identities []*ent.MailIdentity,
) (accountMaterializationFixtureSnapshot, accountMaterializationFixtureObserveCode) {
	snapshot, valid := accountMaterializationFixtureClassifyRows(
		record,
		organization,
		owner,
		domain,
		accounts,
		identities,
	)
	if valid {
		return snapshot, ""
	}
	if code := accountMaterializationFixtureExactRootsFailureCode(record, organization, owner, domain); code != "" {
		return accountMaterializationFixtureSnapshot{}, code
	}
	switch {
	case len(accounts) == 0 && len(identities) != 0:
		return accountMaterializationFixtureSnapshot{}, "identity_rows_without_account"
	case len(accounts) != 1:
		return accountMaterializationFixtureSnapshot{}, "mail_account_cardinality"
	case len(identities) != 1:
		return accountMaterializationFixtureSnapshot{}, "mail_identity_cardinality"
	case !accountMaterializationFixtureCommonAccountIdentityValid(record.plan, accounts[0], identities[0]):
		return accountMaterializationFixtureSnapshot{}, accountMaterializationFixtureCommonAccountIdentityFailureCode(
			record.plan,
			accounts[0],
			identities[0],
		)
	default:
		return accountMaterializationFixtureSnapshot{}, "mail_account_materialization_state"
	}
}

func accountMaterializationFixtureExactRootsFailureCode(
	record accountMaterializationFixtureRecord,
	organization *ent.Organization,
	owner *ent.User,
	domain *ent.MailDomain,
) accountMaterializationFixtureObserveCode {
	plan := record.plan
	switch {
	case record.nonce != plan.nonce:
		return "record_nonce"
	case record.state != accountMaterializationFixtureStateCreated:
		return "record_state"
	case !accountMaterializationFixturePlanMatchesCanonical(plan):
		return "plan_canonical"
	case !accountMaterializationFixturePersistedMatchesPlan(record.persisted, plan):
		return "persisted_roots"
	case organization == nil:
		return "organization_missing"
	case organization.ID != plan.organizationID:
		return "organization_id"
	case organization.Slug != plan.organizationSlug:
		return "organization_slug"
	case organization.Name != "QA Mail Materialization "+plan.nonce:
		return "organization_name"
	case organization.InviteToken != "qa-mail-materialization-invite-"+plan.nonce:
		return "organization_invite_token"
	case !accountMaterializationFixtureValidTime(organization.CreatedAt):
		return "organization_created_at_shape"
	case !accountMaterializationFixtureValidTime(organization.UpdatedAt):
		return "organization_updated_at_shape"
	case !organization.CreatedAt.Equal(plan.fixedAt):
		return "organization_created_at"
	case !organization.UpdatedAt.Equal(plan.fixedAt):
		return "organization_updated_at"
	case owner == nil:
		return "owner_missing"
	case owner.ID != plan.ownerUserID:
		return "owner_id"
	case owner.Email != plan.ownerAddress:
		return "owner_email"
	case owner.SystemRole != user.SystemRoleMember:
		return "owner_system_role"
	case owner.Status != user.StatusActive:
		return "owner_status"
	case owner.Timezone != "UTC":
		return "owner_timezone"
	case owner.Language != "ru":
		return "owner_language"
	case !owner.OnboardingCompleted:
		return "owner_onboarding_completed"
	case owner.Name != "QA Owner":
		return "owner_name"
	case owner.DisplayName != "QA Owner":
		return "owner_display_name"
	case !accountMaterializationFixtureValidTime(owner.CreatedAt):
		return "owner_created_at_shape"
	case !accountMaterializationFixtureValidTime(owner.UpdatedAt):
		return "owner_updated_at_shape"
	case !owner.CreatedAt.Equal(plan.fixedAt):
		return "owner_created_at"
	case !owner.UpdatedAt.Equal(plan.fixedAt):
		return "owner_updated_at"
	case domain == nil:
		return "domain_missing"
	case domain.ID != plan.mailDomainID:
		return "domain_id"
	case domain.OrganizationID != plan.organizationID:
		return "domain_organization_id"
	case domain.Domain != plan.domain:
		return "domain_name"
	case domain.EngineProvisioningKey != plan.domainProvisioningKey:
		return "domain_engine_provisioning_key"
	case domain.EngineDomainID == nil || *domain.EngineDomainID != plan.engineDomainID:
		return "domain_engine_id"
	case !domain.DesiredPresent:
		return "domain_desired_present"
	case !domain.Enabled:
		return "domain_enabled"
	case !domain.IsDefault:
		return "domain_default"
	case domain.State != maildomain.State("verified"):
		return "domain_state"
	case domain.VerificationStatus != maildomain.VerificationStatus("verified"):
		return "domain_verification_status"
	case domain.DkimStatus != maildomain.DkimStatus("valid"):
		return "domain_dkim_status"
	case domain.RequiredDNSRecords == nil || len(domain.RequiredDNSRecords) != 0:
		return "domain_required_dns_records"
	case domain.DesiredGeneration != 1:
		return "domain_desired_generation"
	case domain.DesiredHashVersion != 1:
		return "domain_desired_hash_version"
	case !accountMaterializationFixtureHashEqual(domain.DesiredHash, plan.domainControlHash[:]):
		return "domain_desired_hash"
	case !accountMaterializationFixtureOptionalTupleValid(domain.AttemptedGeneration, domain.AttemptedHashVersion, domain.AttemptedHash):
		return "domain_attempted_tuple_shape"
	case domain.AttemptedGeneration != nil:
		return "domain_attempted_generation"
	case !accountMaterializationFixtureOptionalTupleMatches(
		domain.DeployedGeneration,
		domain.DeployedHashVersion,
		domain.DeployedHash,
		1,
		1,
		plan.domainControlHash[:],
	):
		return "domain_deployed_tuple"
	case domain.LastErrorCode != nil:
		return "domain_last_error_code"
	case domain.NextRetryAt != nil:
		return "domain_next_retry_at"
	case domain.AttemptCount != 0:
		return "domain_attempt_count"
	case domain.LeaseToken != nil:
		return "domain_lease_token"
	case domain.LeaseUntil != nil:
		return "domain_lease_until"
	case domain.DeletedAt != nil:
		return "domain_deleted_at"
	case !accountMaterializationFixtureValidTime(domain.CreatedAt):
		return "domain_created_at_shape"
	case !accountMaterializationFixtureValidTime(domain.UpdatedAt):
		return "domain_updated_at_shape"
	case !domain.CreatedAt.Equal(plan.fixedAt):
		return "domain_created_at"
	case !domain.UpdatedAt.Equal(plan.fixedAt):
		return "domain_updated_at"
	default:
		return ""
	}
}

func accountMaterializationFixtureCommonAccountIdentityFailureCode(
	plan accountMaterializationFixturePlan,
	account *ent.MailAccount,
	identity *ent.MailIdentity,
) accountMaterializationFixtureObserveCode {
	switch {
	case !accountMaterializationFixturePlanMatchesCanonical(plan):
		return "plan_canonical"
	case account == nil:
		return "mail_account_missing"
	case account.ID != plan.expectedAccountIntentID:
		return "mail_account_id"
	case account.OrganizationID != plan.organizationID:
		return "mail_account_organization_id"
	case account.UserID != plan.ownerUserID:
		return "mail_account_user_id"
	case account.MailDomainID != plan.mailDomainID:
		return "mail_account_domain_id"
	case account.EngineProvisioningKey != plan.expectedAccountProvisioningKey:
		return "mail_account_engine_provisioning_key"
	case account.LocalPart != plan.ownerLocalPart:
		return "mail_account_local_part"
	case !account.DesiredEnabled:
		return "mail_account_desired_enabled"
	case account.Status != mailaccount.Status("provisioning"):
		return "mail_account_status"
	case account.ControlGeneration != 1:
		return "mail_account_control_generation"
	case account.ControlHashVersion != 1:
		return "mail_account_control_hash_version"
	case !accountMaterializationFixtureHashEqual(account.ControlHash, plan.expectedAccountControlHash[:]):
		return "mail_account_control_hash"
	case account.ProjectionState != mailaccount.ProjectionState("bootstrapping"):
		return "mail_account_projection_state"
	case !accountMaterializationFixtureValidTime(account.EventCutoffAt):
		return "mail_account_event_cutoff_at"
	case !accountMaterializationFixtureValidTime(account.SyncScheduleUpdatedAt):
		return "mail_account_sync_schedule_updated_at"
	case !accountMaterializationFixtureValidTime(account.CreatedAt):
		return "mail_account_created_at"
	case !accountMaterializationFixtureValidTime(account.UpdatedAt):
		return "mail_account_updated_at"
	case account.ProvisionAttempts < 0:
		return "mail_account_provision_attempts"
	case account.PasswordEnrollmentGeneration != 0:
		return "mail_account_password_enrollment_generation"
	case account.PasswordEnrolledEngineAccountID != nil:
		return "mail_account_password_enrolled_engine_account_id"
	case account.PasswordEnrolledAt != nil:
		return "mail_account_password_enrolled_at"
	case account.PasswordEnrollmentLeaseToken != nil:
		return "mail_account_password_enrollment_lease_token"
	case account.PasswordEnrollmentLeaseUntil != nil:
		return "mail_account_password_enrollment_lease_until"
	case !accountMaterializationFixtureOptionalTupleValid(
		account.AttemptedControlGeneration,
		account.AttemptedControlHashVersion,
		account.AttemptedControlHash,
	):
		return "mail_account_attempted_control_tuple"
	case !accountMaterializationFixtureOptionalTupleValid(
		account.DeployedGeneration,
		account.DeployedHashVersion,
		account.DeployedHash,
	):
		return "mail_account_deployed_tuple"
	case identity == nil:
		return "mail_identity_missing"
	case identity.ID == uuid.Nil:
		return "mail_identity_id"
	case identity.MailAccountID != account.ID:
		return "mail_identity_account_id"
	case identity.NormalizedAddress != plan.ownerAddress:
		return "mail_identity_normalized_address"
	case identity.Kind != mailidentity.KindPrimary:
		return "mail_identity_kind"
	case identity.Status != mailidentity.Status("pending"):
		return "mail_identity_status"
	case identity.SendAllowed:
		return "mail_identity_send_allowed"
	case identity.EngineIdentityID != nil:
		return "mail_identity_engine_id"
	default:
		return "mail_account_identity_classifier"
	}
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
	hash *[]byte,
) bool {
	return generation == nil && hashVersion == nil && (hash == nil || *hash == nil)
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
