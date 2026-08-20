package mail

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MailboxSyncPhase uint8

const (
	MailboxSyncPhaseIncremental MailboxSyncPhase = iota + 1
	MailboxSyncPhaseEnumeration
	MailboxSyncPhaseCatchUp
)

type MailboxSyncFence struct {
	AccountID     uuid.UUID
	LeaseID       uuid.UUID
	LeaseEpoch    int64
	SyncEpoch     int64
	MutationEpoch int64
	Phase         MailboxSyncPhase
	Cursor        string
	BaseCursor    string
	Checkpoint    string
}

type MailboxSyncClaimRequest struct {
	AccountID  uuid.UUID
	LeaseID    uuid.UUID
	Now        time.Time
	LeaseUntil time.Time
}

type MailboxSyncClaimResult struct {
	Claimed         bool
	Fence           MailboxSyncFence
	EngineAccountID string
	EventCutoff     time.Time
	NextAttemptAt   time.Time
}

type MailboxSyncEventCandidate struct {
	EngineMessageID string
	ReceivedEventID uuid.UUID
	ChangedEventID  uuid.UUID
	Destroyed       bool
}

type MailboxSyncPrepareRequest struct {
	Fence           MailboxSyncFence
	EngineAccountID string
	Folders         []EngineFolder
	Page            MailboxSyncPolicyPage
	EventCandidates []MailboxSyncEventCandidate
	OccurredAt      time.Time
	CorrelationID   string
}

type MailboxSyncPreparedPage struct {
	Prepared bool
	Input    MailboxSyncPolicyInput
}

type MailboxSyncApplyRequest struct {
	Fence         MailboxSyncFence
	Result        MailboxSyncPolicyResult
	Now           time.Time
	NextAttemptAt time.Time
}

type MailboxSyncApplyResult struct {
	Applied       bool
	Replayed      bool
	Phase         MailboxSyncPhase
	NextAttemptAt time.Time
}

type MailboxSyncFailureKind uint8

const (
	MailboxSyncFailureInvalidCursor MailboxSyncFailureKind = iota + 1
	MailboxSyncFailureRestartRequired
	MailboxSyncFailureTransient
	MailboxSyncFailureRejected
)

type MailboxSyncResolveRequest struct {
	Fence         MailboxSyncFence
	Failure       MailboxSyncFailureKind
	Now           time.Time
	NextAttemptAt time.Time
}

type MailboxSyncResolveResult struct {
	Resolved      bool
	Phase         MailboxSyncPhase
	NextAttemptAt time.Time
}

type MailboxSyncRepository interface {
	ClaimMailboxSyncStep(context.Context, MailboxSyncClaimRequest) (MailboxSyncClaimResult, error)
	PrepareMailboxSyncPage(context.Context, MailboxSyncPrepareRequest) (MailboxSyncPreparedPage, error)
	ApplyMailboxSyncPage(context.Context, MailboxSyncApplyRequest) (MailboxSyncApplyResult, error)
	ResolveMailboxSyncStep(context.Context, MailboxSyncResolveRequest) (MailboxSyncResolveResult, error)
}
