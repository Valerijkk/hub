//go:build qa && !release

package qamailcalendar

import (
	"context"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

type accountMaterializationFixtureRecoveryOutcome string

const (
	accountMaterializationFixtureRecoveryRestoredCreated  accountMaterializationFixtureRecoveryOutcome = "restored_created"
	accountMaterializationFixtureRecoveryRestoredQuiesced accountMaterializationFixtureRecoveryOutcome = "restored_quiesced"
	accountMaterializationFixtureRecoveryConfirmedDeleted accountMaterializationFixtureRecoveryOutcome = "confirmed_deleted"
	accountMaterializationFixtureRecoveryRetryableContention accountMaterializationFixtureRecoveryOutcome = "retryable_contention"
	accountMaterializationFixtureRecoveryCorruptionDetected  accountMaterializationFixtureRecoveryOutcome = "corruption_detected"
)

type accountMaterializationFixtureRecoveryObservation struct {
	organization      *ent.Organization
	organizationFound bool
	owner             *ent.User
	ownerFound        bool
	domain            *ent.MailDomain
	domainFound       bool
	accounts          []*ent.MailAccount
	identities        []*ent.MailIdentity
}

func (c *Controller) recoverAccountMaterializationFixture(
	ctx context.Context,
	nonce string,
) (accountMaterializationFixtureRecoveryOutcome, error) {
	if c == nil ||
		ctx == nil ||
		c.client == nil ||
		c.accountMaterializationFixtures == nil ||
		!validToken(string(c.config.token)) ||
		!ValidNonce(nonce) {
		return "", errAccountMaterializationFixtureControllerInput
	}

	canonicalPlan, err := newAccountMaterializationFixturePlan(nonce)
	if err != nil {
		return "", errAccountMaterializationFixtureControllerInput
	}

	c.mu.Lock()
	record, found := c.accountMaterializationFixtures[nonce]
	c.mu.Unlock()
	if !found || record.nonce != nonce || record.state != accountMaterializationFixtureStateCleanupFailed {
		return "", errAccountMaterializationFixtureLifecycle
	}
	if record.cleanupOrigin != accountMaterializationFixtureCleanupOriginQuiesceCommit &&
		record.cleanupOrigin != accountMaterializationFixtureCleanupOriginDeleteCommit {
		return accountMaterializationFixtureRecoveryCorruptionDetected, nil
	}
	if record.plan != canonicalPlan ||
		!accountMaterializationFixturePersistedMatchesPlan(record.persisted, canonicalPlan) {
		return accountMaterializationFixtureRecoveryCorruptionDetected, nil
	}

	observation, observed := c.observeAccountMaterializationFixtureRecovery(ctx, record)
	if !observed {
		if !c.accountMaterializationFixtureRecoveryOwns(nonce, record) {
			return "", errAccountMaterializationFixtureLifecycle
		}
		return accountMaterializationFixtureRecoveryRetryableContention, nil
	}
	if !c.accountMaterializationFixtureRecoveryOwns(nonce, record) {
		return "", errAccountMaterializationFixtureLifecycle
	}

	switch record.cleanupOrigin {
	case accountMaterializationFixtureCleanupOriginQuiesceCommit:
		if accountMaterializationFixtureRecoveryActiveGraphValid(record, observation) {
			if !c.completeAccountMaterializationFixtureRecovery(nonce, record, accountMaterializationFixtureStateCreated) {
				return "", errAccountMaterializationFixtureLifecycle
			}
			return accountMaterializationFixtureRecoveryRestoredCreated, nil
		}
		if accountMaterializationFixtureRecoveryDeactivatedGraphValid(record, observation) {
			if !c.completeAccountMaterializationFixtureRecovery(nonce, record, accountMaterializationFixtureStateQuiesced) {
				return "", errAccountMaterializationFixtureLifecycle
			}
			return accountMaterializationFixtureRecoveryRestoredQuiesced, nil
		}
	case accountMaterializationFixtureCleanupOriginDeleteCommit:
		if observation.allRootsAbsent() {
			if !c.removeAccountMaterializationFixtureRecovery(nonce, record) {
				return "", errAccountMaterializationFixtureLifecycle
			}
			return accountMaterializationFixtureRecoveryConfirmedDeleted, nil
		}
		if accountMaterializationFixtureRecoveryDeactivatedGraphValid(record, observation) {
			if !c.completeAccountMaterializationFixtureRecovery(nonce, record, accountMaterializationFixtureStateQuiesced) {
				return "", errAccountMaterializationFixtureLifecycle
			}
			return accountMaterializationFixtureRecoveryRestoredQuiesced, nil
		}
	}

	return accountMaterializationFixtureRecoveryCorruptionDetected, nil
}

func (c *Controller) observeAccountMaterializationFixtureRecovery(
	ctx context.Context,
	record accountMaterializationFixtureRecord,
) (accountMaterializationFixtureRecoveryObservation, bool) {
	tx, err := c.client.Tx(ctx)
	if err != nil || tx == nil {
		return accountMaterializationFixtureRecoveryObservation{}, false
	}
	finished := false
	defer func() {
		if !finished {
			_ = tx.Rollback()
		}
	}()

	observation := accountMaterializationFixtureRecoveryObservation{
		accounts:   make([]*ent.MailAccount, 0),
		identities: make([]*ent.MailIdentity, 0),
	}
	accounts, err := tx.MailAccount.Query().
		Where(
			mailaccount.UserIDEQ(record.persisted.ownerUserID),
			accountMaterializationFixtureMailAccountNoWait(),
		).
		Order(mailaccount.ByID()).
		Limit(2).
		All(ctx)
	if err != nil {
		return accountMaterializationFixtureRecoveryObservation{}, false
	}
	observation.accounts = accounts
	identityAccountID := record.plan.expectedAccountIntentID
	if len(accounts) == 1 {
		identityAccountID = accounts[0].ID
	}
	identities, err := tx.MailIdentity.Query().
		Where(
			mailidentity.MailAccountIDEQ(identityAccountID),
			accountMaterializationFixtureMailIdentityNoWait(),
		).
		Order(mailidentity.ByID()).
		Limit(2).
		All(ctx)
	if err != nil {
		return accountMaterializationFixtureRecoveryObservation{}, false
	}
	observation.identities = identities

	domainRow, err := tx.MailDomain.Query().
		Where(
			maildomain.IDEQ(record.persisted.mailDomainID),
			accountMaterializationFixtureMailDomainNoWait(),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return accountMaterializationFixtureRecoveryObservation{}, false
	}
	if err == nil {
		observation.domain = domainRow
		observation.domainFound = true
	}

	ownerRow, err := tx.User.Query().
		Where(
			user.IDEQ(record.persisted.ownerUserID),
			accountMaterializationFixtureUserNoWait(),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return accountMaterializationFixtureRecoveryObservation{}, false
	}
	if err == nil {
		observation.owner = ownerRow
		observation.ownerFound = true
	}

	organizationRow, err := tx.Organization.Query().
		Where(
			organization.IDEQ(record.persisted.organizationID),
			accountMaterializationFixtureOrganizationNoWait(),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return accountMaterializationFixtureRecoveryObservation{}, false
	}
	if err == nil {
		observation.organization = organizationRow
		observation.organizationFound = true
	}

	if err := tx.Rollback(); err != nil {
		return accountMaterializationFixtureRecoveryObservation{}, false
	}
	finished = true
	return observation, true
}

func (observation accountMaterializationFixtureRecoveryObservation) allRootsAbsent() bool {
	return !observation.organizationFound &&
		!observation.ownerFound &&
		!observation.domainFound &&
		len(observation.accounts) == 0 &&
		len(observation.identities) == 0
}

func accountMaterializationFixtureRecoveryActiveGraphValid(
	record accountMaterializationFixtureRecord,
	observation accountMaterializationFixtureRecoveryObservation,
) bool {
	if !observation.organizationFound || !observation.ownerFound || !observation.domainFound {
		return false
	}
	activeRecord := record
	activeRecord.state = accountMaterializationFixtureStateCreated
	activeRecord.cleanupOrigin = accountMaterializationFixtureCleanupOriginNone
	_, valid := accountMaterializationFixtureClassifyRows(
		activeRecord,
		observation.organization,
		observation.owner,
		observation.domain,
		observation.accounts,
		observation.identities,
	)
	return valid
}

func accountMaterializationFixtureRecoveryDeactivatedGraphValid(
	record accountMaterializationFixtureRecord,
	observation accountMaterializationFixtureRecoveryObservation,
) bool {
	if !observation.organizationFound || !observation.ownerFound || !observation.domainFound {
		return false
	}
	deletingRecord := record
	deletingRecord.state = accountMaterializationFixtureStateDeleting
	deletingRecord.cleanupOrigin = accountMaterializationFixtureCleanupOriginNone
	return accountMaterializationFixtureDeleteGraphValid(
		deletingRecord,
		observation.organization,
		observation.owner,
		observation.domain,
		observation.accounts,
		observation.identities,
	)
}

func (c *Controller) accountMaterializationFixtureRecoveryOwns(
	nonce string,
	record accountMaterializationFixtureRecord,
) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	current, found := c.accountMaterializationFixtures[nonce]
	return found &&
		current == record &&
		current.state == accountMaterializationFixtureStateCleanupFailed &&
		current.cleanupOrigin == record.cleanupOrigin
}

func (c *Controller) completeAccountMaterializationFixtureRecovery(
	nonce string,
	record accountMaterializationFixtureRecord,
	state accountMaterializationFixtureState,
) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	current, found := c.accountMaterializationFixtures[nonce]
	if !found ||
		current != record ||
		current.state != accountMaterializationFixtureStateCleanupFailed ||
		current.cleanupOrigin != record.cleanupOrigin {
		return false
	}
	current.state = state
	current.cleanupOrigin = accountMaterializationFixtureCleanupOriginNone
	c.accountMaterializationFixtures[nonce] = current
	return true
}

func (c *Controller) removeAccountMaterializationFixtureRecovery(
	nonce string,
	record accountMaterializationFixtureRecord,
) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	current, found := c.accountMaterializationFixtures[nonce]
	if !found ||
		current != record ||
		current.state != accountMaterializationFixtureStateCleanupFailed ||
		current.cleanupOrigin != record.cleanupOrigin {
		return false
	}
	delete(c.accountMaterializationFixtures, nonce)
	return true
}
