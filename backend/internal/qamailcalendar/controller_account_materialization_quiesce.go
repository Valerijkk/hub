//go:build qa && !release

package qamailcalendar

import (
	"context"
	"time"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

func (c *Controller) quiesceAccountMaterializationFixture(ctx context.Context, nonce string) error {
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
		record.state != accountMaterializationFixtureStateCreated ||
		record.cleanupOrigin != accountMaterializationFixtureCleanupOriginNone ||
		record.plan != canonicalPlan ||
		!accountMaterializationFixturePersistedMatchesPlan(record.persisted, canonicalPlan) {
		c.mu.Unlock()
		return errAccountMaterializationFixtureLifecycle
	}
	record.state = accountMaterializationFixtureStateQuiescing
	c.accountMaterializationFixtures[nonce] = record
	c.mu.Unlock()

	transition := func(state accountMaterializationFixtureState, origin accountMaterializationFixtureCleanupOrigin) bool {
		c.mu.Lock()
		defer c.mu.Unlock()

		current, found := c.accountMaterializationFixtures[nonce]
		if !found ||
			current != record ||
			current.state != accountMaterializationFixtureStateQuiescing {
			return false
		}
		current.state = state
		current.cleanupOrigin = origin
		c.accountMaterializationFixtures[nonce] = current
		return true
	}

	tx, err := c.client.Tx(ctx)
	if err != nil || tx == nil {
		if !transition(accountMaterializationFixtureStateCreated, accountMaterializationFixtureCleanupOriginNone) {
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
		if !transition(accountMaterializationFixtureStateCreated, accountMaterializationFixtureCleanupOriginNone) {
			return errAccountMaterializationFixtureLifecycle
		}
		return errAccountMaterializationFixtureLifecycle
	}

	ownerRow, err := tx.User.Query().
		Where(user.IDEQ(record.persisted.ownerUserID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return rollbackAndRestore()
	}
	organizationRow, err := tx.Organization.Query().
		Where(
			organization.IDEQ(record.persisted.organizationID),
			organization.HasUsersWith(user.IDEQ(record.persisted.ownerUserID)),
		).
		Only(ctx)
	if err != nil || !accountMaterializationFixtureQuiesceRootsValid(record, organizationRow, ownerRow) {
		return rollbackAndRestore()
	}

	updated, err := tx.User.Update().
		Where(
			user.IDEQ(record.persisted.ownerUserID),
			user.StatusEQ(user.StatusActive),
		).
		SetStatus(user.StatusDeactivated).
		SetUpdatedAt(accountMaterializationFixtureQuiescedAt(record.plan)).
		Save(ctx)
	if err != nil || updated != 1 {
		return rollbackAndRestore()
	}

	finished = true
	if err := tx.Commit(); err != nil {
		if !transition(accountMaterializationFixtureStateCleanupFailed, accountMaterializationFixtureCleanupOriginQuiesceCommit) {
			return errAccountMaterializationFixtureLifecycle
		}
		return errAccountMaterializationFixtureLifecycle
	}
	if !transition(accountMaterializationFixtureStateQuiesced, accountMaterializationFixtureCleanupOriginNone) {
		return errAccountMaterializationFixtureLifecycle
	}
	return nil
}

func accountMaterializationFixtureQuiesceRootsValid(
	record accountMaterializationFixtureRecord,
	organizationRow *ent.Organization,
	ownerRow *ent.User,
) bool {
	plan := record.plan
	return record.nonce == plan.nonce &&
		record.state == accountMaterializationFixtureStateQuiescing &&
		record.cleanupOrigin == accountMaterializationFixtureCleanupOriginNone &&
		accountMaterializationFixturePlanMatchesCanonical(plan) &&
		accountMaterializationFixturePersistedMatchesPlan(record.persisted, plan) &&
		organizationRow != nil &&
		organizationRow.ID == plan.organizationID &&
		organizationRow.Slug == plan.organizationSlug &&
		organizationRow.Name == "QA Mail Materialization "+plan.nonce &&
		organizationRow.InviteToken == "qa-mail-materialization-invite-"+plan.nonce &&
		accountMaterializationFixtureValidTime(organizationRow.CreatedAt) &&
		accountMaterializationFixtureValidTime(organizationRow.UpdatedAt) &&
		organizationRow.CreatedAt.Equal(plan.fixedAt) &&
		organizationRow.UpdatedAt.Equal(plan.fixedAt) &&
		ownerRow != nil &&
		ownerRow.ID == plan.ownerUserID &&
		ownerRow.Email == plan.ownerAddress &&
		ownerRow.OrganizationID == organizationRow.ID &&
		ownerRow.SystemRole == user.SystemRoleMember &&
		ownerRow.Status == user.StatusActive &&
		ownerRow.Timezone == "UTC" &&
		ownerRow.Language == "ru" &&
		ownerRow.OnboardingCompleted &&
		ownerRow.Name == "QA Owner" &&
		ownerRow.DisplayName == "QA Owner" &&
		accountMaterializationFixtureValidTime(ownerRow.CreatedAt) &&
		accountMaterializationFixtureValidTime(ownerRow.UpdatedAt) &&
		ownerRow.CreatedAt.Equal(plan.fixedAt) &&
		ownerRow.UpdatedAt.Equal(plan.fixedAt)
}

func accountMaterializationFixtureQuiescedAt(plan accountMaterializationFixturePlan) time.Time {
	return plan.fixedAt.Add(time.Second)
}
