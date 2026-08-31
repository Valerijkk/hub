//go:build !release

package qamailcalendar

import (
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

const threadSearchFixtureRowCount = 6

type threadSearchFixturePlan struct {
	nonce                  string
	fixedAt                time.Time
	organizationID         uuid.UUID
	ownerUserID            uuid.UUID
	mailDomainID           uuid.UUID
	mailAccountID          uuid.UUID
	mailIdentityID         uuid.UUID
	mailFolderID           uuid.UUID
	organizationSlug       string
	domain                 string
	ownerAddress           string
	domainProvisioningKey  string
	accountProvisioningKey string
	engineDomainID         string
	engineAccountID        string
	engineIdentityID       string
	engineFolderID         string
	folderName             string
	normalizedFolderName   string
	correlationID          string
	domainControlHash      [sha256.Size]byte
	accountControlHash     [sha256.Size]byte
	rows                   threadSearchFixtureRows
	matchingThreadIDs      [4]uuid.UUID
	decoyThreadIDs         [2]uuid.UUID
}

type threadSearchFixtureRows [threadSearchFixtureRowCount]threadSearchFixtureRow

type threadSearchFixtureRow struct {
	role              string
	threadID          uuid.UUID
	messageID         uuid.UUID
	messageFolderID   uuid.UUID
	normalizedSubject string
	snippet           string
	isPinned          bool
	isFlagged         bool
	lastMessageAt     time.Time
	threadKey         [sha256.Size]byte
	engineMessageID   string
	engineVersion     string
}

type threadSearchFixtureRowIDs struct {
	label           string
	threadID        uuid.UUID
	messageID       uuid.UUID
	messageFolderID uuid.UUID
}

type threadSearchFixturePersisted struct {
	organizationID   uuid.UUID
	ownerUserID      uuid.UUID
	mailDomainID     uuid.UUID
	mailAccountID    uuid.UUID
	mailIdentityID   uuid.UUID
	mailFolderID     uuid.UUID
	threadIDs        [threadSearchFixtureRowCount]uuid.UUID
	messageIDs       [threadSearchFixtureRowCount]uuid.UUID
	messageFolderIDs [threadSearchFixtureRowCount]uuid.UUID
}

// ThreadSearchFixtureLifecycleScenario identifies the only lifecycle
// scenario admitted by the pure reconciliation seam.
type ThreadSearchFixtureLifecycleScenario string

const (
	ThreadSearchFixtureLifecycleScenarioInvalid        ThreadSearchFixtureLifecycleScenario = ""
	ThreadSearchFixtureLifecycleScenarioThreadSearchV1 ThreadSearchFixtureLifecycleScenario = "thread-search-v1"
)

// ThreadSearchFixtureLifecycleState is the closed state set persisted by the
// pure lifecycle reconciliation seam. A deleted record is removed instead of
// retained as a state.
type ThreadSearchFixtureLifecycleState string

const (
	ThreadSearchFixtureLifecycleStateInvalid     ThreadSearchFixtureLifecycleState = ""
	ThreadSearchFixtureLifecycleStateCreating    ThreadSearchFixtureLifecycleState = "creating"
	ThreadSearchFixtureLifecycleStateCreated     ThreadSearchFixtureLifecycleState = "created"
	ThreadSearchFixtureLifecycleStateDeleting    ThreadSearchFixtureLifecycleState = "deleting"
	ThreadSearchFixtureLifecycleStateInvalidated ThreadSearchFixtureLifecycleState = "invalidated"
)

// ThreadSearchFixtureLifecycleRecord is a fully comparable lifecycle value.
// The private binding remains part of equality so a public-identity match
// cannot weaken a map CAS.
type ThreadSearchFixtureLifecycleRecord struct {
	Nonce           string
	Scenario        ThreadSearchFixtureLifecycleScenario
	State           ThreadSearchFixtureLifecycleState
	PlanBinding     [sha256.Size]byte
	ManifestBinding [sha256.Size]byte
	privateBinding  [sha256.Size]byte
	plan            threadSearchFixturePlan
	persisted       threadSearchFixturePersisted
}

// threadSearchFixtureRecord is retained only as a source-compatible private
// name while the authoritative map stores the exported lifecycle value.
type threadSearchFixtureRecord = ThreadSearchFixtureLifecycleRecord

// NewThreadSearchFixtureLifecycleRecord constructs the sole testable
// lifecycle value. Production derivation of the domain-separated bindings
// remains with the QA-tagged fixture lifecycle.
func NewThreadSearchFixtureLifecycleRecord(
	nonce string,
	scenario ThreadSearchFixtureLifecycleScenario,
	state ThreadSearchFixtureLifecycleState,
	planBinding [sha256.Size]byte,
	manifestBinding [sha256.Size]byte,
	privateBinding [sha256.Size]byte,
) ThreadSearchFixtureLifecycleRecord {
	return ThreadSearchFixtureLifecycleRecord{
		Nonce:           nonce,
		Scenario:        scenario,
		State:           state,
		PlanBinding:     planBinding,
		ManifestBinding: manifestBinding,
		privateBinding:  privateBinding,
	}
}

// ThreadSearchFixtureFreshProbeSummary is the raw result of one fresh probe
// over the fixed fixture manifest. The classifier owns its fail-closed
// interpretation.
type ThreadSearchFixtureFreshProbeSummary struct {
	Probed         bool
	ExactRows      int
	AbsentRows     int
	MismatchedRows int
	Failed         bool
}

// ThreadSearchFixtureCommitReconcileInput is the immutable input to the pure
// lifecycle map-CAS seam.
type ThreadSearchFixtureCommitReconcileInput struct {
	Operation ThreadSearchCommitOperation
	Phase     ThreadSearchCommitPhase
	Probe     ThreadSearchFixtureFreshProbeSummary
	Expected  ThreadSearchFixtureLifecycleRecord
}

// ThreadSearchFixtureLifecycleDisposition describes the exact map effect
// attempted by the reconciliation seam. Its zero value is not actionable.
type ThreadSearchFixtureLifecycleDisposition uint8

const (
	ThreadSearchFixtureLifecycleDispositionInvalid ThreadSearchFixtureLifecycleDisposition = iota
	ThreadSearchFixtureLifecycleDispositionCreated
	ThreadSearchFixtureLifecycleDispositionRemoved
	ThreadSearchFixtureLifecycleDispositionRetained
	ThreadSearchFixtureLifecycleDispositionRestored
	ThreadSearchFixtureLifecycleDispositionCASConflict
	ThreadSearchFixtureLifecycleDispositionInvalidated
)

// ThreadSearchFixtureCommitReconcileResult carries the controller-visible
// HTTP semantics and whether the exact map CAS mutated the lifecycle record.
type ThreadSearchFixtureCommitReconcileResult struct {
	Status      int
	Error       string
	Disposition ThreadSearchFixtureLifecycleDisposition
	Applied     bool
}
