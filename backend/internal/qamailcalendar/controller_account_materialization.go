//go:build qa && !release

package qamailcalendar

import (
	"errors"

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
