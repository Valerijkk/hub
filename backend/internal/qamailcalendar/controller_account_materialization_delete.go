//go:build qa && !release

package qamailcalendar

import (
	"context"

	"entgo.io/ent/dialect/sql"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/user"
)

func (c *Controller) deleteAccountMaterializationFixture(ctx context.Context, nonce string) error {
	if c == nil ||
		ctx == nil ||
		c.client == nil ||
		c.accountMaterializationFixtures == nil ||
		!validToken(string(c.config.token)) ||
		!ValidNonce(nonce) {
		return errAccountMaterializationFixtureControllerInput
	}

	canonicalPlan, err := newAccountMaterializationFixturePlan(nonce)
	if err != nil {
		return errAccountMaterializationFixtureControllerInput
	}

	c.mu.Lock()
	record, found := c.accountMaterializationFixtures[nonce]
	if !found ||
		record.nonce != nonce ||
		record.state != accountMaterializationFixtureStateQuiesced ||
		record.cleanupOrigin != accountMaterializationFixtureCleanupOriginNone ||
		record.plan != canonicalPlan ||
		!accountMaterializationFixturePersistedMatchesPlan(record.persisted, canonicalPlan) {
		c.mu.Unlock()
		return errAccountMaterializationFixtureLifecycle
	}
	record.state = accountMaterializationFixtureStateDeleting
	c.accountMaterializationFixtures[nonce] = record
	c.mu.Unlock()

	transition := func(state accountMaterializationFixtureState, origin accountMaterializationFixtureCleanupOrigin) bool {
		c.mu.Lock()
		defer c.mu.Unlock()

		current, found := c.accountMaterializationFixtures[nonce]
		if !found ||
			current != record ||
			current.state != accountMaterializationFixtureStateDeleting {
			return false
		}
		current.state = state
		current.cleanupOrigin = origin
		c.accountMaterializationFixtures[nonce] = current
		return true
	}

	tx, err := c.client.Tx(ctx)
	if err != nil || tx == nil {
		if !transition(accountMaterializationFixtureStateQuiesced, accountMaterializationFixtureCleanupOriginNone) {
			return errAccountMaterializationFixtureLifecycle
		}
		return errAccountMaterializationFixtureLifecycle
	}
	finished := false
	defer func() {
		if !finished {
			_ = tx.Rollback()
		}
	}()
	rollbackAndRestore := func() error {
		_ = tx.Rollback()
		finished = true
		if !transition(accountMaterializationFixtureStateQuiesced, accountMaterializationFixtureCleanupOriginNone) {
			return errAccountMaterializationFixtureLifecycle
		}
		return errAccountMaterializationFixtureLifecycle
	}

	accounts, err := tx.MailAccount.Query().
		Where(
			mailaccount.UserIDEQ(record.persisted.ownerUserID),
			accountMaterializationFixtureMailAccountNoWait(),
		).
		Order(mailaccount.ByID()).
		Limit(2).
		All(ctx)
	if err != nil || len(accounts) > 1 {
		return rollbackAndRestore()
	}

	identities := make([]*ent.MailIdentity, 0)
	if len(accounts) == 1 {
		identities, err = tx.MailIdentity.Query().
			Where(
				mailidentity.MailAccountIDEQ(accounts[0].ID),
				accountMaterializationFixtureMailIdentityNoWait(),
			).
			Order(mailidentity.ByID()).
			Limit(2).
			All(ctx)
		if err != nil || len(identities) > 1 {
			return rollbackAndRestore()
		}
	}

	domainRow, err := tx.MailDomain.Query().
		Where(
			maildomain.IDEQ(record.persisted.mailDomainID),
			accountMaterializationFixtureMailDomainNoWait(),
		).
		Only(ctx)
	if err != nil {
		return rollbackAndRestore()
	}
	ownerRow, err := tx.User.Query().
		Where(
			user.IDEQ(record.persisted.ownerUserID),
			accountMaterializationFixtureUserNoWait(),
		).
		Only(ctx)
	if err != nil {
		return rollbackAndRestore()
	}
	organizationRow, err := tx.Organization.Query().
		Where(
			organization.IDEQ(record.persisted.organizationID),
			accountMaterializationFixtureOrganizationNoWait(),
		).
		Only(ctx)
	if err != nil || !accountMaterializationFixtureDeleteGraphValid(record, organizationRow, ownerRow, domainRow, accounts, identities) {
		return rollbackAndRestore()
	}

	if len(identities) == 1 {
		if err := tx.MailIdentity.DeleteOneID(identities[0].ID).Exec(ctx); err != nil {
			return rollbackAndRestore()
		}
		if err := tx.MailAccount.DeleteOneID(accounts[0].ID).Exec(ctx); err != nil {
			return rollbackAndRestore()
		}
	}
	if err := tx.MailDomain.DeleteOneID(domainRow.ID).Exec(ctx); err != nil {
		return rollbackAndRestore()
	}
	if err := tx.User.DeleteOneID(ownerRow.ID).Exec(ctx); err != nil {
		return rollbackAndRestore()
	}
	if err := tx.Organization.DeleteOneID(organizationRow.ID).Exec(ctx); err != nil {
		return rollbackAndRestore()
	}

	finished = true
	if err := tx.Commit(); err != nil {
		if !transition(accountMaterializationFixtureStateCleanupFailed, accountMaterializationFixtureCleanupOriginDeleteCommit) {
			return errAccountMaterializationFixtureLifecycle
		}
		return errAccountMaterializationFixtureLifecycle
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	current, found := c.accountMaterializationFixtures[nonce]
	if !found ||
		current != record ||
		current.state != accountMaterializationFixtureStateDeleting ||
		current.cleanupOrigin != accountMaterializationFixtureCleanupOriginNone {
		return errAccountMaterializationFixtureLifecycle
	}
	delete(c.accountMaterializationFixtures, nonce)
	return nil
}

func accountMaterializationFixtureDeleteGraphValid(
	record accountMaterializationFixtureRecord,
	organizationRow *ent.Organization,
	ownerRow *ent.User,
	domainRow *ent.MailDomain,
	accounts []*ent.MailAccount,
	identities []*ent.MailIdentity,
) bool {
	if !accountMaterializationFixtureDeleteRootsValid(record, organizationRow, ownerRow, domainRow) {
		return false
	}
	if len(accounts) == 0 {
		return len(identities) == 0
	}
	if len(accounts) != 1 || len(identities) != 1 {
		return false
	}
	if !accountMaterializationFixtureCommonAccountIdentityValid(record.plan, accounts[0], identities[0]) {
		return false
	}
	_, valid := accountMaterializationFixtureClassifyAccountState(record.plan, accounts[0])
	return valid
}

func accountMaterializationFixtureDeleteRootsValid(
	record accountMaterializationFixtureRecord,
	organizationRow *ent.Organization,
	ownerRow *ent.User,
	domainRow *ent.MailDomain,
) bool {
	if record.nonce != record.plan.nonce ||
		record.state != accountMaterializationFixtureStateDeleting ||
		record.cleanupOrigin != accountMaterializationFixtureCleanupOriginNone ||
		!accountMaterializationFixturePlanMatchesCanonical(record.plan) ||
		!accountMaterializationFixturePersistedMatchesPlan(record.persisted, record.plan) ||
		ownerRow == nil ||
		ownerRow.Status != user.StatusDeactivated ||
		!accountMaterializationFixtureValidTime(ownerRow.CreatedAt) ||
		!accountMaterializationFixtureValidTime(ownerRow.UpdatedAt) ||
		!ownerRow.CreatedAt.Equal(record.plan.fixedAt) ||
		!ownerRow.UpdatedAt.Equal(accountMaterializationFixtureQuiescedAt(record.plan)) {
		return false
	}

	activeRecord := record
	activeRecord.state = accountMaterializationFixtureStateCreated
	activeOwner := *ownerRow
	activeOwner.Status = user.StatusActive
	activeOwner.UpdatedAt = record.plan.fixedAt
	return accountMaterializationFixtureExactRootsValid(activeRecord, organizationRow, &activeOwner, domainRow)
}

func accountMaterializationFixtureMailAccountNoWait() predicate.MailAccount {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func accountMaterializationFixtureMailIdentityNoWait() predicate.MailIdentity {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func accountMaterializationFixtureMailDomainNoWait() predicate.MailDomain {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func accountMaterializationFixtureUserNoWait() predicate.User {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func accountMaterializationFixtureOrganizationNoWait() predicate.Organization {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
