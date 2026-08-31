//go:build qa && !release

package qamailcalendar

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
)

var (
	errThreadSearchFixtureControllerInput = errors.New("qa mail thread search fixture controller input is invalid")
	errThreadSearchFixtureLifecycle       = errors.New("qa mail thread search fixture lifecycle failed")
)

const (
	threadSearchFixtureLifecyclePlanBindingDomain     = "lifecycle-plan-binding-v2"
	threadSearchFixtureLifecycleManifestBindingDomain = "lifecycle-manifest-binding-v2"
	threadSearchFixtureLifecyclePrivateBindingDomain  = "lifecycle-private-binding-v2"
)

type threadSearchFixtureSnapshot struct {
	nonce             string
	scenario          Scenario
	state             ThreadSearchFixtureLifecycleState
	ownerEmail        string
	organizationID    uuid.UUID
	ownerUserID       uuid.UUID
	mailAccountID     uuid.UUID
	mailFolderID      uuid.UUID
	matchingThreadIDs [4]uuid.UUID
	decoyThreadIDs    [2]uuid.UUID
}

func (c *Controller) createThreadSearchFixture(
	ctx context.Context,
) (threadSearchFixtureSnapshot, error) {
	if c == nil ||
		ctx == nil ||
		c.client == nil ||
		c.random == nil ||
		c.fixtures == nil ||
		c.accountMaterializationFixtures == nil ||
		c.threadSearchFixtures == nil ||
		!validToken(string(c.config.token)) {
		return threadSearchFixtureSnapshot{}, errThreadSearchFixtureControllerInput
	}

	expected, reserved := c.reserveThreadSearchFixture()
	if !reserved {
		return threadSearchFixtureSnapshot{}, errThreadSearchFixtureLifecycle
	}

	persisted, err := persistThreadSearchFixture(ctx, c.client, expected.plan)
	phase := ThreadSearchCommitPhaseSucceeded
	probe := ThreadSearchFixtureFreshProbeSummary{}
	if err != nil {
		phase = ThreadSearchCommitPhasePreCommit
		var outcomeUnknown *threadSearchFixtureCommitOutcomeUnknownError
		if errors.As(err, &outcomeUnknown) && outcomeUnknown != nil {
			phase = ThreadSearchCommitPhaseOutcomeUnknown
			probe = c.threadSearchFixtureProbeManifest(ctx, expected.plan, expected.persisted)
		}
	}
	if err == nil && persisted != expected.persisted {
		return threadSearchFixtureSnapshot{}, errThreadSearchFixtureLifecycle
	}

	c.mu.Lock()
	reconciled := ReconcileThreadSearchFixtureCommit(
		c.threadSearchFixtures,
		ThreadSearchFixtureCommitReconcileInput{
			Operation: ThreadSearchCommitOperationCreate,
			Phase:     phase,
			Probe:     probe,
			Expected:  expected,
		},
	)
	c.mu.Unlock()
	if reconciled.Status != 201 ||
		reconciled.Disposition != ThreadSearchFixtureLifecycleDispositionCreated ||
		!reconciled.Applied {
		return threadSearchFixtureSnapshot{}, errThreadSearchFixtureLifecycle
	}

	created := expected
	created.State = ThreadSearchFixtureLifecycleStateCreated
	snapshot, valid := threadSearchFixtureSnapshotFromCreatedRecord(created)
	if !valid {
		return threadSearchFixtureSnapshot{}, errThreadSearchFixtureLifecycle
	}
	return snapshot, nil
}

func (c *Controller) reserveThreadSearchFixture() (ThreadSearchFixtureLifecycleRecord, bool) {
	if c == nil ||
		c.random == nil ||
		c.fixtures == nil ||
		c.accountMaterializationFixtures == nil ||
		c.threadSearchFixtures == nil {
		return ThreadSearchFixtureLifecycleRecord{}, false
	}

	for attempt := 0; attempt < 4; attempt++ {
		var randomBytes [16]byte
		if _, err := io.ReadFull(c.random, randomBytes[:]); err != nil {
			return ThreadSearchFixtureLifecycleRecord{}, false
		}
		nonce := hex.EncodeToString(randomBytes[:])
		plan, err := newThreadSearchFixturePlan(nonce)
		if err != nil {
			return ThreadSearchFixtureLifecycleRecord{}, false
		}
		record, valid := newThreadSearchFixtureCreatingRecord(plan)
		if !valid {
			return ThreadSearchFixtureLifecycleRecord{}, false
		}

		c.mu.Lock()
		_, inboundExists := c.fixtures[nonce]
		_, accountMaterializationExists := c.accountMaterializationFixtures[nonce]
		_, threadSearchExists := c.threadSearchFixtures[nonce]
		if !inboundExists && !accountMaterializationExists && !threadSearchExists {
			c.threadSearchFixtures[nonce] = record
			c.mu.Unlock()
			return record, true
		}
		c.mu.Unlock()
	}

	return ThreadSearchFixtureLifecycleRecord{}, false
}

func threadSearchFixtureSnapshotFromCreatedRecord(
	record ThreadSearchFixtureLifecycleRecord,
) (threadSearchFixtureSnapshot, bool) {
	if !threadSearchFixtureLifecycleRecordMatches(record, ThreadSearchFixtureLifecycleStateCreated) {
		return threadSearchFixtureSnapshot{}, false
	}
	return threadSearchFixtureSnapshot{
		nonce:             record.Nonce,
		scenario:          ScenarioThreadSearchV1,
		state:             record.State,
		ownerEmail:        record.plan.ownerAddress,
		organizationID:    record.persisted.organizationID,
		ownerUserID:       record.persisted.ownerUserID,
		mailAccountID:     record.persisted.mailAccountID,
		mailFolderID:      record.persisted.mailFolderID,
		matchingThreadIDs: record.plan.matchingThreadIDs,
		decoyThreadIDs:    record.plan.decoyThreadIDs,
	}, true
}

func (c *Controller) storedFixtureScenario(nonce string) (Scenario, error) {
	if c == nil ||
		c.fixtures == nil ||
		c.accountMaterializationFixtures == nil ||
		c.threadSearchFixtures == nil ||
		!ValidNonce(nonce) {
		return "", errThreadSearchFixtureControllerInput
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var (
		scenario Scenario
		matches  int
	)
	if record, found := c.fixtures[nonce]; found {
		if record.nonce != nonce ||
			(record.state != inboundRequestFixtureStateCreated &&
				record.state != inboundRequestFixtureStatePublished) {
			return "", errThreadSearchFixtureLifecycle
		}
		scenario = ScenarioInboundRequestV1
		matches++
	}
	if record, found := c.accountMaterializationFixtures[nonce]; found {
		if record.nonce != nonce ||
			(record.state != accountMaterializationFixtureStateCreated &&
				record.state != accountMaterializationFixtureStateQuiesced) {
			return "", errThreadSearchFixtureLifecycle
		}
		scenario = ScenarioAccountMaterializationV1
		matches++
	}
	if record, found := c.threadSearchFixtures[nonce]; found {
		if record.Nonce != nonce ||
			record.Scenario != ThreadSearchFixtureLifecycleScenarioThreadSearchV1 ||
			!threadSearchFixtureLifecycleRecordMatches(record, ThreadSearchFixtureLifecycleStateCreated) {
			return "", errThreadSearchFixtureLifecycle
		}
		scenario = ScenarioThreadSearchV1
		matches++
	}
	if matches != 1 {
		return "", errThreadSearchFixtureLifecycle
	}
	return scenario, nil
}

func newThreadSearchFixtureCreatingRecord(
	plan threadSearchFixturePlan,
) (ThreadSearchFixtureLifecycleRecord, bool) {
	planBinding, valid := threadSearchFixturePlanBinding(plan)
	if !valid {
		return ThreadSearchFixtureLifecycleRecord{}, false
	}
	persisted := threadSearchFixturePersistedFromPlan(plan)
	manifestBinding := threadSearchFixtureManifestBinding(persisted)

	return ThreadSearchFixtureLifecycleRecord{
		Nonce:           plan.nonce,
		Scenario:        ThreadSearchFixtureLifecycleScenarioThreadSearchV1,
		State:           ThreadSearchFixtureLifecycleStateCreating,
		PlanBinding:     planBinding,
		ManifestBinding: manifestBinding,
		privateBinding:  threadSearchFixturePrivateBinding(planBinding, manifestBinding),
		plan:            plan,
		persisted:       persisted,
	}, true
}

func threadSearchFixturePlanBinding(plan threadSearchFixturePlan) ([sha256.Size]byte, bool) {
	if !threadSearchFixturePlanMatchesCanonical(plan) {
		return [sha256.Size]byte{}, false
	}

	encoder := newThreadSearchFixtureBindingEncoder(threadSearchFixtureLifecyclePlanBindingDomain)
	encoder.appendString(plan.nonce)
	encoder.appendTime(plan.fixedAt)
	encoder.appendUUID(plan.organizationID)
	encoder.appendUUID(plan.ownerUserID)
	encoder.appendUUID(plan.mailDomainID)
	encoder.appendUUID(plan.mailAccountID)
	encoder.appendUUID(plan.mailIdentityID)
	encoder.appendUUID(plan.mailFolderID)
	encoder.appendString(plan.organizationSlug)
	encoder.appendString(plan.domain)
	encoder.appendString(plan.ownerAddress)
	encoder.appendString(plan.domainProvisioningKey)
	encoder.appendString(plan.accountProvisioningKey)
	encoder.appendString(plan.engineDomainID)
	encoder.appendString(plan.engineAccountID)
	encoder.appendString(plan.engineIdentityID)
	encoder.appendString(plan.engineFolderID)
	encoder.appendString(plan.folderName)
	encoder.appendString(plan.normalizedFolderName)
	encoder.appendString(plan.correlationID)
	encoder.appendDigest(plan.domainControlHash)
	encoder.appendDigest(plan.accountControlHash)
	encoder.appendUint64(uint64(len(plan.rows)))
	for _, row := range plan.rows {
		encoder.appendString(row.role)
		encoder.appendUUID(row.threadID)
		encoder.appendUUID(row.messageID)
		encoder.appendUUID(row.messageFolderID)
		encoder.appendString(row.normalizedSubject)
		encoder.appendString(row.snippet)
		encoder.appendBool(row.isPinned)
		encoder.appendBool(row.isFlagged)
		encoder.appendTime(row.lastMessageAt)
		encoder.appendDigest(row.threadKey)
		encoder.appendString(row.engineMessageID)
		encoder.appendString(row.engineVersion)
	}
	encoder.appendUint64(uint64(len(plan.matchingThreadIDs)))
	for _, id := range plan.matchingThreadIDs {
		encoder.appendUUID(id)
	}
	encoder.appendUint64(uint64(len(plan.decoyThreadIDs)))
	for _, id := range plan.decoyThreadIDs {
		encoder.appendUUID(id)
	}
	return sha256.Sum256(encoder.bytes), true
}

func threadSearchFixtureManifestBinding(persisted threadSearchFixturePersisted) [sha256.Size]byte {
	encoder := newThreadSearchFixtureBindingEncoder(threadSearchFixtureLifecycleManifestBindingDomain)
	encoder.appendUUID(persisted.organizationID)
	encoder.appendUUID(persisted.ownerUserID)
	encoder.appendUUID(persisted.mailDomainID)
	encoder.appendUUID(persisted.mailAccountID)
	encoder.appendUUID(persisted.mailIdentityID)
	encoder.appendUUID(persisted.mailFolderID)
	encoder.appendUint64(uint64(len(persisted.threadIDs)))
	for _, id := range persisted.threadIDs {
		encoder.appendUUID(id)
	}
	encoder.appendUint64(uint64(len(persisted.messageIDs)))
	for _, id := range persisted.messageIDs {
		encoder.appendUUID(id)
	}
	encoder.appendUint64(uint64(len(persisted.messageFolderIDs)))
	for _, id := range persisted.messageFolderIDs {
		encoder.appendUUID(id)
	}
	return sha256.Sum256(encoder.bytes)
}

func threadSearchFixturePrivateBinding(planBinding [sha256.Size]byte, manifestBinding [sha256.Size]byte) [sha256.Size]byte {
	encoder := newThreadSearchFixtureBindingEncoder(threadSearchFixtureLifecyclePrivateBindingDomain)
	encoder.appendDigest(planBinding)
	encoder.appendDigest(manifestBinding)
	return sha256.Sum256(encoder.bytes)
}

type threadSearchFixtureBindingEncoder struct {
	bytes []byte
}

func newThreadSearchFixtureBindingEncoder(domain string) threadSearchFixtureBindingEncoder {
	encoder := threadSearchFixtureBindingEncoder{}
	encoder.appendString(threadSearchFixtureNamespace)
	encoder.appendString(domain)
	return encoder
}

func (encoder *threadSearchFixtureBindingEncoder) appendString(value string) {
	encoder.appendUint64(uint64(len(value)))
	encoder.bytes = append(encoder.bytes, value...)
}

func (encoder *threadSearchFixtureBindingEncoder) appendUUID(value uuid.UUID) {
	encoder.bytes = append(encoder.bytes, value[:]...)
}

func (encoder *threadSearchFixtureBindingEncoder) appendDigest(value [sha256.Size]byte) {
	encoder.bytes = append(encoder.bytes, value[:]...)
}

func (encoder *threadSearchFixtureBindingEncoder) appendBool(value bool) {
	if value {
		encoder.bytes = append(encoder.bytes, 1)
		return
	}
	encoder.bytes = append(encoder.bytes, 0)
}

func (encoder *threadSearchFixtureBindingEncoder) appendTime(value time.Time) {
	encoder.appendString(value.UTC().Format(time.RFC3339Nano))
}

func (encoder *threadSearchFixtureBindingEncoder) appendUint64(value uint64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	encoder.bytes = append(encoder.bytes, encoded[:]...)
}

func threadSearchFixtureLifecycleRecordMatches(
	record ThreadSearchFixtureLifecycleRecord,
	state ThreadSearchFixtureLifecycleState,
) bool {
	expected, valid := newThreadSearchFixtureCreatingRecord(record.plan)
	return valid &&
		record.Nonce == expected.Nonce &&
		record.Scenario == expected.Scenario &&
		record.State == state &&
		record.PlanBinding == expected.PlanBinding &&
		record.ManifestBinding == expected.ManifestBinding &&
		record.privateBinding == expected.privateBinding &&
		record.persisted == expected.persisted
}
