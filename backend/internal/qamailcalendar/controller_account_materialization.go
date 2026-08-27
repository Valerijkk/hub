//go:build qa && !release

package qamailcalendar

import (
	"context"
	"encoding/hex"
	"errors"
	"io"

	"github.com/google/uuid"
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
