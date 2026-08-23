package mail

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

// MailboxEngineCommandLeaseDuration is the repository-owned duration of one account-wide engine command lease.
// The repository derives its deadline from DB time and generates its token; callers never supply either value.
const MailboxEngineCommandLeaseDuration = 60 * time.Second

// MailboxDraftStatus is the durable lifecycle state of a provider-neutral draft.
type MailboxDraftStatus string

const (
	MailboxDraftStatusActive     MailboxDraftStatus = "active"
	MailboxDraftStatusSaving     MailboxDraftStatus = "saving"
	MailboxDraftStatusDeleting   MailboxDraftStatus = "deleting"
	MailboxDraftStatusUnknown    MailboxDraftStatus = "unknown"
	MailboxDraftStatusSubmitting MailboxDraftStatus = "submitting"
	MailboxDraftStatusSubmitted  MailboxDraftStatus = "submitted"
	MailboxDraftStatusDeleted    MailboxDraftStatus = "deleted"
)

// MailboxDraftMutationKind identifies the durable engine mutation pending for a draft.
type MailboxDraftMutationKind string

const (
	MailboxDraftMutationCreate MailboxDraftMutationKind = "create"
	MailboxDraftMutationSave   MailboxDraftMutationKind = "save"
	MailboxDraftMutationDelete MailboxDraftMutationKind = "delete"
)

// MailboxDraftRepositoryErrorKind is the closed, content-free failure family of MailboxDraftRepository.
type MailboxDraftRepositoryErrorKind string

const (
	MailboxDraftRepositoryErrorInvalid             MailboxDraftRepositoryErrorKind = "invalid"
	MailboxDraftRepositoryErrorNotFound            MailboxDraftRepositoryErrorKind = "not_found"
	MailboxDraftRepositoryErrorUnavailable         MailboxDraftRepositoryErrorKind = "unavailable"
	MailboxDraftRepositoryErrorAccountDeleting     MailboxDraftRepositoryErrorKind = "account_deleting"
	MailboxDraftRepositoryErrorStateConflict       MailboxDraftRepositoryErrorKind = "state_conflict"
	MailboxDraftRepositoryErrorFenceLost           MailboxDraftRepositoryErrorKind = "fence_lost"
	MailboxDraftRepositoryErrorIdempotencyConflict MailboxDraftRepositoryErrorKind = "idempotency_conflict"
	MailboxDraftRepositoryErrorAttachmentConflict  MailboxDraftRepositoryErrorKind = "attachment_conflict"
	MailboxDraftRepositoryErrorAttachmentLimit     MailboxDraftRepositoryErrorKind = "attachment_limit"
)

const mailboxDraftRepositoryErrorText = "mailbox draft repository error"

// MailboxDraftRepositoryError carries one validated repository failure kind and no provider, driver, or draft content.
type MailboxDraftRepositoryError struct {
	kind MailboxDraftRepositoryErrorKind
}

// NewMailboxDraftRepositoryError returns a content-free error with a validated kind.
func NewMailboxDraftRepositoryError(kind MailboxDraftRepositoryErrorKind) *MailboxDraftRepositoryError {
	if !validMailboxDraftRepositoryErrorKind(kind) {
		kind = MailboxDraftRepositoryErrorInvalid
	}
	return &MailboxDraftRepositoryError{kind: kind}
}

// Error returns the same safe text for every repository failure.
func (err *MailboxDraftRepositoryError) Error() string {
	return mailboxDraftRepositoryErrorText
}

// Kind returns a validated repository failure kind.
func (err *MailboxDraftRepositoryError) Kind() MailboxDraftRepositoryErrorKind {
	if err == nil || !validMailboxDraftRepositoryErrorKind(err.kind) {
		return MailboxDraftRepositoryErrorInvalid
	}
	return err.kind
}

func validMailboxDraftRepositoryErrorKind(kind MailboxDraftRepositoryErrorKind) bool {
	switch kind {
	case MailboxDraftRepositoryErrorInvalid,
		MailboxDraftRepositoryErrorNotFound,
		MailboxDraftRepositoryErrorUnavailable,
		MailboxDraftRepositoryErrorAccountDeleting,
		MailboxDraftRepositoryErrorStateConflict,
		MailboxDraftRepositoryErrorFenceLost,
		MailboxDraftRepositoryErrorIdempotencyConflict,
		MailboxDraftRepositoryErrorAttachmentConflict,
		MailboxDraftRepositoryErrorAttachmentLimit:
		return true
	default:
		return false
	}
}

// MailboxEngineCommandBeginRequest identifies the account whose single engine command lease is requested.
type MailboxEngineCommandBeginRequest struct {
	AccountID uuid.UUID
}

// MailboxEngineCommandFence is the exact account lease identity and epoch that fences one engine command.
// A stale, expired, or reclaimed fence has zero durable effect.
type MailboxEngineCommandFence struct {
	AccountID     uuid.UUID
	LeaseID       uuid.UUID
	MutationEpoch int64
	LeaseUntil    time.Time
}

// MailboxEngineCommandBeginResult is exclusively a claimed command or a busy result.
// Claimed results have a nonempty validated EngineAccountID, a complete Fence, and no BusyUntil.
// Busy results have an empty EngineAccountID, no Fence, and a DB-derived BusyUntil.
// Begin is a separately committed short transaction: a lease ending at or before DB time is reclaimable,
// while only a lease ending after DB time is live; a claim advances mutation epoch at begin.
type MailboxEngineCommandBeginResult struct {
	Claimed         bool
	EngineAccountID string
	Fence           *MailboxEngineCommandFence
	BusyUntil       *time.Time
}

// MailboxDraftRecord is the content-free, durable draft state. Nil values represent nullable persisted columns.
// Repository reads and finish results deep-copy pointer values and sort/copy SelectedAttachmentIDs in raw UUID order.
type MailboxDraftRecord struct {
	ID                  uuid.UUID
	AccountID           uuid.UUID
	DraftClientID       uuid.UUID
	EngineDraftKey      string
	EngineMessageID     *string
	EngineVersion       *string
	Revision            int64
	Status              MailboxDraftStatus
	ReplyToMessageID    *uuid.UUID
	ForwardMessageID    *uuid.UUID
	LastMutationID      *uuid.UUID
	PendingMutationKind *MailboxDraftMutationKind
	SelectedAttachmentIDs []uuid.UUID
}

// MailboxDraftGetRequest identifies one account-scoped draft.
type MailboxDraftGetRequest struct {
	AccountID uuid.UUID
	DraftID   uuid.UUID
}

// MailboxDraftGetResult contains a deep-copy-safe durable record when Found is true.
type MailboxDraftGetResult struct {
	Found bool
	Draft *MailboxDraftRecord
}

// MailboxDraftReadRecheckRequest seals the local active draft revision and exact engine message/version pair.
type MailboxDraftReadRecheckRequest struct {
	AccountID       uuid.UUID
	DraftID         uuid.UUID
	Revision        int64
	EngineMessageID string
	EngineVersion   string
}

// MailboxDraftReadRecheckResult reports whether the sealed active-draft read remains current.
// An active GetDraft may return content only after Current is true.
type MailboxDraftReadRecheckResult struct {
	Current bool
}

// MailboxDraftCreateRequest reserves a new account-scoped draft mutation under the exact acquired fence.
// The repository validates the account/draft/client/create-key identity matrix, same-account reply/forward references,
// and a complete distinct attachment set of at most 50 storage-backed, same-account, same-client, unclaimed-or-same-draft
// attachments before it mutates durable state. It locks account, draft, then attachments in ascending raw UUID order.
type MailboxDraftCreateRequest struct {
	Fence                    MailboxEngineCommandFence
	DraftID                  uuid.UUID
	DraftClientID            uuid.UUID
	CreateIdempotencyKeyHash [sha256.Size]byte
	CreateRequestHash        [sha256.Size]byte
	ReplyToMessageID         *uuid.UUID
	ForwardMessageID         *uuid.UUID
	SelectedAttachmentIDs    []uuid.UUID
	MutationID               uuid.UUID
}

// MailboxDraftSaveRequest reserves an active draft save mutation under the exact acquired fence.
type MailboxDraftSaveRequest struct {
	Fence                 MailboxEngineCommandFence
	DraftID               uuid.UUID
	DraftClientID         uuid.UUID
	ExpectedRevision      int64
	SelectedAttachmentIDs []uuid.UUID
	MutationID            uuid.UUID
}

// MailboxDraftDeleteRequest reserves an active draft delete mutation under the exact acquired fence.
type MailboxDraftDeleteRequest struct {
	Fence            MailboxEngineCommandFence
	DraftID          uuid.UUID
	ExpectedRevision int64
	MutationID       uuid.UUID
}

// MailboxDraftWorkKind identifies one provider-neutral draft operation or recovery operation.
type MailboxDraftWorkKind string

const (
	MailboxDraftWorkCreate          MailboxDraftWorkKind = "create"
	MailboxDraftWorkSave            MailboxDraftWorkKind = "save"
	MailboxDraftWorkDelete          MailboxDraftWorkKind = "delete"
	MailboxDraftWorkReconcileCreate MailboxDraftWorkKind = "reconcile_create"
	MailboxDraftWorkReconcileSave   MailboxDraftWorkKind = "reconcile_save"
	MailboxDraftWorkReconcileDelete MailboxDraftWorkKind = "reconcile_delete"
)

// MailboxDraftWork is the persisted, provider-neutral plan for an engine operation.
// Every dispatch or reconciliation plan contains the stable bounded EngineDraftKey, persisted mutation ID, revision,
// exact nullable message/version pair, and a deep-copied sorted attachment set; it never contains draft content or bytes.
type MailboxDraftWork struct {
	Kind                  MailboxDraftWorkKind
	EngineDraftKey        string
	EngineMessageID       *string
	EngineVersion         *string
	MutationID            uuid.UUID
	Revision              int64
	SelectedAttachmentIDs []uuid.UUID
}

// MailboxDraftReservationKind distinguishes a new engine dispatch, an idempotent replay, and recovery that must reconcile.
type MailboxDraftReservationKind string

const (
	MailboxDraftReservationDispatch          MailboxDraftReservationKind = "dispatch"
	MailboxDraftReservationReplay            MailboxDraftReservationKind = "replay"
	MailboxDraftReservationReconcileRequired MailboxDraftReservationKind = "reconcile_required"
)

// MailboxDraftReservation is the result of a draft mutation reservation.
// Dispatch and reconcile-required results retain the exact Fence and include Work. Replay has nil Work and Fence and
// atomically releases the acquired account lease. Conflict, error, and other no-work branches also release that lease,
// unless a lost fence has already been replaced by another owner.
type MailboxDraftReservation struct {
	Kind  MailboxDraftReservationKind
	Draft MailboxDraftRecord
	Work  *MailboxDraftWork
	Fence *MailboxEngineCommandFence
}

// MailboxDraftReconcileRequest identifies one draft for recovery under a freshly acquired exact fence.
type MailboxDraftReconcileRequest struct {
	Fence   MailboxEngineCommandFence
	DraftID uuid.UUID
}

// MailboxDraftReconcilePlan returns only durable work and mutation state for recovery.
// It may turn reclaimed saving or deleting state into the matching unknown state and retain the fresh fence. A result
// that requires no reconciliation releases the fence.
type MailboxDraftReconcilePlan struct {
	Reconcile bool
	Draft     MailboxDraftRecord
	Work      *MailboxDraftWork
	Fence     *MailboxEngineCommandFence
}

// MailboxDraftResumeCreateRequest authorizes one controlled SaveDraft after authoritative absence for unknown/create.
// It must match the live fence, draft and client identities, canonical request hash, and complete attachment set while
// preserving the persisted mutation ID; no generic absence result authorizes a save or means an existing draft was retained.
type MailboxDraftResumeCreateRequest struct {
	Fence                    MailboxEngineCommandFence
	DraftID                  uuid.UUID
	DraftClientID            uuid.UUID
	CreateIdempotencyKeyHash [sha256.Size]byte
	CreateRequestHash        [sha256.Size]byte
	SelectedAttachmentIDs    []uuid.UUID
}

// MailboxDraftFinishOutcome is sealed to the closed outcomes that may finish a pending engine command.
type MailboxDraftFinishOutcome interface {
	mailboxDraftFinishOutcome()
}

// MailboxDraftApplied records a nonempty opaque engine message/version pair for a create or save.
type MailboxDraftApplied struct {
	EngineMessageID string
	EngineVersion   string
}

// MailboxDraftSaveRetained records the exact provider-confirmed attachment set when a pending save retained its old pair.
type MailboxDraftSaveRetained struct {
	ProviderConfirmedAttachmentIDs []uuid.UUID
}

// MailboxDraftDeleteRetained records that a pending delete retained the active draft and its revision/pair.
type MailboxDraftDeleteRetained struct{}

// MailboxDraftDeleted records that a pending delete completed while retaining its old valid pair as a tombstone.
type MailboxDraftDeleted struct{}

// MailboxDraftUnknown records an indeterminate result for the exact current pending mutation.
type MailboxDraftUnknown struct{}

func (MailboxDraftApplied) mailboxDraftFinishOutcome()        {}
func (MailboxDraftSaveRetained) mailboxDraftFinishOutcome()   {}
func (MailboxDraftDeleteRetained) mailboxDraftFinishOutcome() {}
func (MailboxDraftDeleted) mailboxDraftFinishOutcome()        {}
func (MailboxDraftUnknown) mailboxDraftFinishOutcome()        {}

// MailboxDraftFinishRequest completes exactly one pending mutation under the exact acquired fence.
type MailboxDraftFinishRequest struct {
	Fence      MailboxEngineCommandFence
	DraftID    uuid.UUID
	MutationID uuid.UUID
	Kind       MailboxDraftMutationKind
	Outcome    MailboxDraftFinishOutcome
}

// MailboxDraftFinishResult contains the deep-copy-safe durable draft after a successful finish.
type MailboxDraftFinishResult struct {
	Draft MailboxDraftRecord
}

// MailboxDraftRepository owns provider-neutral draft persistence and the account-wide engine-command fence.
// It validates the complete account/token/begin-epoch/unexpired-persisted-deadline fence and exact pending draft state,
// then clears the lease and advances epoch atomically with a transition. A stale, expired, or reclaimed fence mutates no
// row and triggers no compensating I/O. Create transitions to saving/create and then active revision 1 on Applied or
// unknown/create on Unknown. Save transitions active N to saving/save and then active N+1 on Applied or unknown/save N
// on Unknown; only exact GetDraft plus read recheck can finish save-retained back to active N with its confirmed set.
// Delete transitions active N to deleting/delete and then deleted N on Deleted, active N on delete-retained, or
// unknown/delete N on Unknown. The outcome matrix is closed: Applied is valid only for pending create or save;
// save-retained only for save; delete-retained and deleted only for delete; Unknown only for the exact pending kind.
// Other outcome/kind/status combinations are invalid, leave the draft unchanged, and release an owned lease. Deleted
// atomically clears selected attachment associations and returns an empty set; it never deletes attachment bytes or rows.
// Invalid dependencies, input, or state and all generic DB or driver failures map to MailboxDraftRepositoryError only.
type MailboxDraftRepository interface {
	BeginEngineCommand(context.Context, MailboxEngineCommandBeginRequest) (MailboxEngineCommandBeginResult, error)
	GetMailboxDraft(context.Context, MailboxDraftGetRequest) (MailboxDraftGetResult, error)
	ReserveMailboxDraftCreate(context.Context, MailboxDraftCreateRequest) (MailboxDraftReservation, error)
	ReserveMailboxDraftSave(context.Context, MailboxDraftSaveRequest) (MailboxDraftReservation, error)
	ReserveMailboxDraftDelete(context.Context, MailboxDraftDeleteRequest) (MailboxDraftReservation, error)
	PrepareMailboxDraftReconciliation(context.Context, MailboxDraftReconcileRequest) (MailboxDraftReconcilePlan, error)
	ResumeMailboxDraftCreateAfterAbsence(context.Context, MailboxDraftResumeCreateRequest) (MailboxDraftReservation, error)
	FinishEngineCommand(context.Context, MailboxDraftFinishRequest) (MailboxDraftFinishResult, error)
	RecheckMailboxDraftRead(context.Context, MailboxDraftReadRecheckRequest) (MailboxDraftReadRecheckResult, error)
}
