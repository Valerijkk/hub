//go:build qa && !release

package qamailcalendar

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"sync"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/mail"
)

var (
	errInboundRequestFixtureControllerInput = errors.New("qa mail calendar controller input is invalid")
	errInboundRequestFixtureLifecycle       = errors.New("qa mail calendar fixture lifecycle failed")
	errInboundRequestFixtureCleanup         = errors.New("qa mail calendar fixture cleanup failed")
)

type inboundRequestFixtureState string

const (
	inboundRequestFixtureStateCreating       inboundRequestFixtureState = "creating"
	inboundRequestFixtureStateCreated        inboundRequestFixtureState = "created"
	inboundRequestFixtureStateDeleting       inboundRequestFixtureState = "deleting"
	inboundRequestFixtureStatePublishing     inboundRequestFixtureState = "publishing"
	inboundRequestFixtureStatePublished      inboundRequestFixtureState = "published"
	inboundRequestFixtureStatePublishUnknown inboundRequestFixtureState = "publish_unknown"
	inboundRequestFixtureStateCleanupFailed  inboundRequestFixtureState = "cleanup_failed"
	inboundRequestFixtureStateCleanupClaimed       inboundRequestFixtureState = "cleanup_claimed"
	inboundRequestFixtureStateCleanupCommitUnknown inboundRequestFixtureState = "cleanup_commit_unknown"
	inboundRequestFixtureStateCleanupEngineFailed  inboundRequestFixtureState = "cleanup_engine_failed"
	inboundRequestFixtureStateCleaned               inboundRequestFixtureState = "cleaned"
)

type inboundRequestFixtureRecord struct {
	nonce         string
	plan          inboundRequestFixturePlan
	engineFixture inboundRequestEngineFixture
	persisted     inboundRequestFixturePersisted
	state         inboundRequestFixtureState
	cleanupProof  cleanupDomainProof
}

type inboundRequestFixtureSnapshot struct {
	nonce           string
	scenario        Scenario
	state           inboundRequestFixtureState
	ownerEmail      string
	organizationID  uuid.UUID
	ownerUserID     uuid.UUID
	mailAccountID   uuid.UUID
	mailFolderID    uuid.UUID
	mailThreadID    uuid.UUID
	mailMessageID   uuid.UUID
	envelopeEventID uuid.UUID
}

// Controller owns the private lifecycle of deterministic Mail/Calendar QA fixtures.
type Controller struct {
	client  *ent.Client
	engine  *mail.FakeMailEngine
	config  Config
	random  io.Reader
	mu      sync.Mutex
	fixtures                       map[string]inboundRequestFixtureRecord
	accountMaterializationFixtures map[string]accountMaterializationFixtureRecord
	threadSearchFixtures           map[string]ThreadSearchFixtureLifecycleRecord
	publishedRecoveryRouteInstalled bool
}

func NewController(client *ent.Client, engine *mail.FakeMailEngine, config Config) (*Controller, error) {
	if client == nil || engine == nil {
		return nil, errInboundRequestFixtureControllerInput
	}

	token := append([]byte(nil), config.token...)
	if !validToken(string(token)) {
		return nil, errInboundRequestFixtureControllerInput
	}

	return &Controller{
		client:   client,
		engine:   engine,
		config:   Config{token: token},
		random:   cryptorand.Reader,
		fixtures:                       make(map[string]inboundRequestFixtureRecord),
		accountMaterializationFixtures: make(map[string]accountMaterializationFixtureRecord),
		threadSearchFixtures:           make(map[string]ThreadSearchFixtureLifecycleRecord),
	}, nil
}

func (c *Controller) createInboundRequestFixture(
	ctx context.Context,
) (inboundRequestFixtureSnapshot, error) {
	if c == nil ||
		ctx == nil ||
		c.client == nil ||
		c.engine == nil ||
		c.random == nil ||
		c.fixtures == nil ||
		c.accountMaterializationFixtures == nil ||
		c.threadSearchFixtures == nil ||
		!validToken(string(c.config.token)) {
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureControllerInput
	}

	nonce, reserved := c.reserveInboundRequestFixture()
	if !reserved {
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
	}

	plan, err := newInboundRequestFixturePlan(nonce)
	if err != nil {
		if !c.removeCreatingInboundRequestFixture(nonce) {
			return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
		}
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
	}
	if !c.storeCreatingInboundRequestFixturePlan(nonce, plan) {
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
	}

	engineFixture, err := seedInboundRequestFixtureEngine(ctx, c.engine, plan)
	if err != nil {
		if !c.removeCreatingInboundRequestFixture(nonce) {
			return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
		}
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
	}
	if !c.storeCreatingInboundRequestFixtureEngine(nonce, plan, engineFixture) {
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
	}

	persisted, err := persistInboundRequestFixture(ctx, c.client, plan, engineFixture)
	if err != nil {
		cleanupErr := destroyInboundRequestFixtureEngine(c.engine, engineFixture)
		if cleanupErr != nil {
			if !c.markInboundRequestFixtureCleanupFailed(nonce, plan, engineFixture) {
				return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
			}
			return inboundRequestFixtureSnapshot{}, errors.Join(
				errInboundRequestFixtureLifecycle,
				errInboundRequestFixtureCleanup,
			)
		}
		if !c.removeCreatingInboundRequestFixture(nonce) {
			return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
		}
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
	}
	if !inboundRequestFixturePersistedMatchesPlan(persisted, plan) {
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
	}

	snapshot, completed := c.completeInboundRequestFixture(nonce, plan, engineFixture, persisted)
	if !completed {
		return inboundRequestFixtureSnapshot{}, errInboundRequestFixtureLifecycle
	}
	return snapshot, nil
}

func (c *Controller) reserveInboundRequestFixture() (string, bool) {
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
			c.fixtures[nonce] = inboundRequestFixtureRecord{
				nonce: nonce,
				state: inboundRequestFixtureStateCreating,
			}
			c.mu.Unlock()
			return nonce, true
		}
		c.mu.Unlock()
	}

	return "", false
}

func (c *Controller) removeCreatingInboundRequestFixture(nonce string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	record, found := c.fixtures[nonce]
	if !found || record.nonce != nonce || record.state != inboundRequestFixtureStateCreating {
		return false
	}
	delete(c.fixtures, nonce)
	return true
}

func (c *Controller) storeCreatingInboundRequestFixturePlan(
	nonce string,
	plan inboundRequestFixturePlan,
) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	record, found := c.fixtures[nonce]
	if !found || record.nonce != nonce || record.state != inboundRequestFixtureStateCreating {
		return false
	}
	record.plan = plan
	c.fixtures[nonce] = record
	return true
}

func (c *Controller) storeCreatingInboundRequestFixtureEngine(
	nonce string,
	plan inboundRequestFixturePlan,
	engineFixture inboundRequestEngineFixture,
) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	record, found := c.fixtures[nonce]
	if !found ||
		record.nonce != nonce ||
		record.state != inboundRequestFixtureStateCreating ||
		record.plan != plan {
		return false
	}
	record.engineFixture = engineFixture
	c.fixtures[nonce] = record
	return true
}

func (c *Controller) markInboundRequestFixtureCleanupFailed(
	nonce string,
	plan inboundRequestFixturePlan,
	engineFixture inboundRequestEngineFixture,
) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	record, found := c.fixtures[nonce]
	if !found ||
		record.nonce != nonce ||
		record.state != inboundRequestFixtureStateCreating ||
		record.plan != plan ||
		record.engineFixture != engineFixture {
		return false
	}
	record.state = inboundRequestFixtureStateCleanupFailed
	c.fixtures[nonce] = record
	return true
}

func (c *Controller) completeInboundRequestFixture(
	nonce string,
	plan inboundRequestFixturePlan,
	engineFixture inboundRequestEngineFixture,
	persisted inboundRequestFixturePersisted,
) (inboundRequestFixtureSnapshot, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	record, found := c.fixtures[nonce]
	if !found ||
		record.nonce != nonce ||
		record.state != inboundRequestFixtureStateCreating ||
		record.plan != plan ||
		record.engineFixture != engineFixture {
		return inboundRequestFixtureSnapshot{}, false
	}
	record.persisted = persisted
	record.state = inboundRequestFixtureStateCreated
	c.fixtures[nonce] = record

	return inboundRequestFixtureSnapshot{
		nonce:           record.nonce,
		scenario:        ScenarioInboundRequestV1,
		state:           record.state,
		ownerEmail:      record.plan.ownerAddress,
		organizationID:  record.persisted.organizationID,
		ownerUserID:     record.persisted.ownerUserID,
		mailAccountID:   record.persisted.mailAccountID,
		mailFolderID:    record.persisted.mailFolderID,
		mailThreadID:    record.persisted.mailThreadID,
		mailMessageID:   record.persisted.mailMessageID,
		envelopeEventID: record.persisted.envelopeEventID,
	}, true
}

func inboundRequestFixturePersistedMatchesPlan(
	persisted inboundRequestFixturePersisted,
	plan inboundRequestFixturePlan,
) bool {
	return persisted.organizationID == plan.organizationID &&
		persisted.ownerUserID == plan.ownerUserID &&
		persisted.mailAccountID == plan.mailAccountID &&
		persisted.mailIdentityID == plan.mailIdentityID &&
		persisted.mailFolderID == plan.mailFolderID &&
		persisted.mailThreadID == plan.mailThreadID &&
		persisted.mailMessageID == plan.mailMessageID &&
		persisted.envelopeEventID == plan.envelopeEventID &&
		persisted.calendarLocatorHash == plan.calendarLocatorHash
}
