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
	MailboxDraftStatusAbandoned  MailboxDraftStatus = "abandoned"
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

// MailboxDraftEffectiveState is one complete content-free draft graph. AttachmentIDs are
// distinct, sorted in raw UUID order, and are always copied across repository boundaries.
type MailboxDraftEffectiveState struct {
	FromIdentityID   uuid.UUID
	ReplyToMessageID *uuid.UUID
	ForwardMessageID *uuid.UUID
	AttachmentIDs    []uuid.UUID
}

// MailboxDraftPendingSave is the requested D graph durably reserved for a save. Its request
// hash is the exact canonical input binding; it never contains draft content.
type MailboxDraftPendingSave struct {
	EffectiveState MailboxDraftEffectiveState
	RequestHash    [sha256.Size]byte
}

// MailboxDraftCreateAttachmentPhase is the closed pre-engine attachment phase for a pending create.
type MailboxDraftCreateAttachmentPhase string

const (
	MailboxDraftCreateAttachmentPhaseUpload MailboxDraftCreateAttachmentPhase = "upload"
	MailboxDraftCreateAttachmentPhaseImport MailboxDraftCreateAttachmentPhase = "import"
)

// MailboxDraftRecord is the content-free, durable draft state. Nil values represent nullable
// persisted columns. PendingCreateUploadBlobID and PendingCreateUploadSizeBytes form one exact
// upload handoff pair: they are present only while PendingCreateAttachmentPhase is import.
// Repository reads and finish results deep-copy every nested state, pointer, hash and attachment slice.
type MailboxDraftRecord struct {
	ID                             uuid.UUID
	AccountID                      uuid.UUID
	DraftClientID                  uuid.UUID
	EngineDraftKey                 string
	EngineMessageID                *string
	EngineVersion                  *string
	Revision                       int64
	Status                         MailboxDraftStatus
	Confirmed                      MailboxDraftEffectiveState
	PendingSave                    *MailboxDraftPendingSave
	CreateRequestHash              [sha256.Size]byte
	LastMutationID                 *uuid.UUID
	PendingMutationKind            *MailboxDraftMutationKind
	PendingCreateAttachmentPhase   *MailboxDraftCreateAttachmentPhase
	PendingCreateUploadBlobID      *string
	PendingCreateUploadSizeBytes   *int64
	PendingCreateUploadAttempt     int
	PendingCreateUploadRetryAt     *time.Time
	PendingCreateUploadExhaustedAt *time.Time
	ConfirmedMutationID            *uuid.UUID
	AbandonedAt                    *time.Time
	AbandonedMutationID            *uuid.UUID
	UpdatedAt                      time.Time
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

// MailboxDraftCreateReplayLookupRequest identifies one account-scoped prior draft-create request.
type MailboxDraftCreateReplayLookupRequest struct {
	AccountID                uuid.UUID
	DraftClientID            uuid.UUID
	CreateIdempotencyKeyHash [sha256.Size]byte
	CreateRequestHash        [sha256.Size]byte
}

// MailboxDraftCreateReplayLookupResult contains the exact durable replay identity when Found is true.
type MailboxDraftCreateReplayLookupResult struct {
	Found          bool
	DraftID        uuid.UUID
	LastMutationID *uuid.UUID
}

// MailboxDraftCreatePreHashLocatorRequest identifies one account-scoped create before request-hash adjudication.
type MailboxDraftCreatePreHashLocatorRequest struct {
	AccountID                uuid.UUID
	DraftClientID            uuid.UUID
	CreateIdempotencyKeyHash [sha256.Size]byte
}

// MailboxDraftCreatePreHashLocatorResult contains one compatible durable create candidate when Found is true.
type MailboxDraftCreatePreHashLocatorResult struct {
	Found            bool
	DraftID          uuid.UUID
	Status           MailboxDraftStatus
	FromIdentityID   uuid.UUID
	ReplyToMessageID *uuid.UUID
	ForwardMessageID *uuid.UUID
	BusyUntil        *time.Time
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
	Confirmed                MailboxDraftEffectiveState
	MutationID               uuid.UUID
}

// MailboxDraftReplySourceProof binds an active Reply save to the exact owned
// source-message projection from which its server-authoritative threading was
// derived. It contains only local and opaque engine identifiers; header values
// and message content never cross the repository boundary.
type MailboxDraftReplySourceProof struct {
	MessageID       uuid.UUID
	EngineMessageID string
	EngineVersion   string
}

// MailboxDraftSaveRequest reserves an active draft save mutation under the exact acquired fence.
type MailboxDraftSaveRequest struct {
	Fence                 MailboxEngineCommandFence
	DraftID               uuid.UUID
	DraftClientID         uuid.UUID
	ExpectedRevision      int64
	PendingSave           MailboxDraftPendingSave
	ReplySourceProof      *MailboxDraftReplySourceProof
	MutationID            uuid.UUID
}

// MailboxDraftDeleteRequest reserves an active draft delete mutation under the exact acquired fence.
type MailboxDraftDeleteRequest struct {
	Fence            MailboxEngineCommandFence
	DraftID          uuid.UUID
	ExpectedRevision int64
	MutationID       uuid.UUID
}

// MailboxDraftAbortRequest aborts the exact pending draft mutation under the exact acquired fence.
type MailboxDraftAbortRequest struct {
	Fence      MailboxEngineCommandFence
	DraftID    uuid.UUID
	MutationID uuid.UUID
}

// MailboxDraftAdvanceCreateToImportRequest atomically stores the exact opaque upload blob handoff and
// its byte size while advancing an attachment-backed create to its import command.
type MailboxDraftAdvanceCreateToImportRequest struct {
	Fence      MailboxEngineCommandFence
	DraftID    uuid.UUID
	MutationID uuid.UUID
	BlobID     string
	SizeBytes  int64
}

// MailboxDraftAbortResult reports whether abort removed a provisional create or restored a durable active draft.
type MailboxDraftAbortResult struct {
	DraftDeleted bool
	Draft        *MailboxDraftRecord
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
// exact nullable message/version pair, an effective O or D graph, its request hash, and the last confirmed mutation
// binding. PendingCreateUploadBlobID and PendingCreateUploadSizeBytes are present only for an import-phase create
// and identify the exact already-uploaded MIME blob; work never contains draft content or MIME bytes.
type MailboxDraftWork struct {
	Kind                         MailboxDraftWorkKind
	EngineDraftKey               string
	EngineMessageID              *string
	EngineVersion                *string
	MutationID                   uuid.UUID
	Revision                     int64
	EffectiveState               MailboxDraftEffectiveState
	RequestHash                  [sha256.Size]byte
	PendingCreateAttachmentPhase *MailboxDraftCreateAttachmentPhase
	PendingCreateUploadBlobID    *string
	PendingCreateUploadSizeBytes *int64
	PendingCreateUploadAttempt   int
	PendingCreateUploadRetryAt   *time.Time
	PendingCreateUploadExhaustedAt *time.Time
	ConfirmedMutationID          *uuid.UUID
}

// validMailboxDraftPendingCreateUploadHandoff validates the nullable durable upload handoff pair.
// Upload and no attachment-create phase carry no blob. Import carries one nonempty opaque blob ID and
// an exact MIME byte count within the provider-neutral message limit.
func validMailboxDraftPendingCreateUploadHandoff(phase *MailboxDraftCreateAttachmentPhase, blobID *string, sizeBytes *int64) bool {
	if phase == nil {
		return blobID == nil && sizeBytes == nil
	}

	switch *phase {
	case MailboxDraftCreateAttachmentPhaseUpload:
		return blobID == nil && sizeBytes == nil
	case MailboxDraftCreateAttachmentPhaseImport:
		return blobID != nil && sizeBytes != nil && validOpaqueID(*blobID) && *sizeBytes >= 1 && *sizeBytes <= MaxEngineMessageBytes
	default:
		return false
	}
}

// validMailboxDraftAdvanceCreateToImportHandoff validates the required immutable upload handoff
// passed to the transition that makes an attachment-backed create ready for Email/import.
func validMailboxDraftAdvanceCreateToImportHandoff(blobID string, sizeBytes int64) bool {
	return validOpaqueID(blobID) && sizeBytes >= 1 && sizeBytes <= MaxEngineMessageBytes
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
	Confirmed                MailboxDraftEffectiveState
}

// MailboxDraftRestartCreateUploadRequest authorizes one controlled retry of an unknown attachment upload.
// The repository validates the exact recovered mutation, attempt, idempotency binding, and acquired fence before
// rotating the mutation identity and returning new upload work.
type MailboxDraftRestartCreateUploadRequest struct {
	Fence              MailboxEngineCommandFence
	DraftID            uuid.UUID
	ExpectedMutationID uuid.UUID
	ExpectedAttempt    int
	NextMutationID     uuid.UUID
	IdempotencyKeyHash [sha256.Size]byte
	RequestHash        [sha256.Size]byte
	FingerprintVersion int32
}

// MailboxDraftRestartCreateUploadAfterBlobNotFoundRequest authorizes only the deterministic
// synchronous recovery from an Email/import blobNotFound result. The persisted draft must still be
// saving/create/import under the exact fence and mutation, and its opaque blob handoff must match
// ExpectedBlobID and ExpectedSizeBytes exactly. NextMutationID is caller-generated and distinct.
type MailboxDraftRestartCreateUploadAfterBlobNotFoundRequest struct {
	Fence             MailboxEngineCommandFence
	ActorID           uuid.UUID
	AccountID         uuid.UUID
	DraftID           uuid.UUID
	CurrentMutationID uuid.UUID
	ExpectedBlobID    string
	ExpectedSizeBytes int64
	NextMutationID    uuid.UUID
}

// MailboxDraftAbandonCreateUploadRequest records a terminal local tombstone for an unknown attachment upload.
// It requires the exact recovered mutation, attempt, idempotency binding, and acquired fence.
type MailboxDraftAbandonCreateUploadRequest struct {
	Fence              MailboxEngineCommandFence
	DraftID            uuid.UUID
	ExpectedMutationID uuid.UUID
	ExpectedAttempt    int
	IdempotencyKeyHash [sha256.Size]byte
	RequestHash        [sha256.Size]byte
	FingerprintVersion int32
}

// MailboxDraftRestartCreateUploadResult is a replayed durable draft or a fresh restart reservation.
// Reservation is nonnil only for a fresh restart and nil for replay.
type MailboxDraftRestartCreateUploadResult struct {
	Replayed    bool
	Draft       MailboxDraftRecord
	Reservation *MailboxDraftReservation
}

// MailboxDraftRestartCreateUploadAfterBlobNotFoundResult contains the exact durable upload
// restart and retained command fence. This deterministic transition has no replay branch:
// a repeated or stale request fails closed because the import handoff no longer exists.
type MailboxDraftRestartCreateUploadAfterBlobNotFoundResult struct {
	Draft       MailboxDraftRecord
	Reservation MailboxDraftReservation
}

// MailboxDraftAbandonCreateUploadResult is a replayed or newly abandoned durable draft.
type MailboxDraftAbandonCreateUploadResult struct {
	Replayed bool
	Draft    MailboxDraftRecord
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

// MailboxDraftSaveRetained carries only a pair observation. Finish intentionally rejects it until
// a separately authorized durable retained-completion proof path exists.
type MailboxDraftSaveRetained struct {
	EngineMessageID string
	EngineVersion   string
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
// unknown/create on Unknown. Save preserves confirmed O while it reserves desired D, then promotes D to active N+1 on
// Applied or preserves both graphs as unknown/save N on Unknown. SaveRetained remains reject-only in this packet.
// Delete transitions active N to deleting/delete and then deleted N on Deleted, active N on delete-retained, or
// unknown/delete N on Unknown. The outcome matrix is closed: Applied is valid only for pending create or save;
// save-retained only for save; delete-retained and deleted only for delete; Unknown only for the exact pending kind.
// Other outcome/kind/status combinations are invalid, leave the draft unchanged, and release an owned lease. Deleted
// atomically clears confirmed attachment associations and returns an empty confirmed attachment set; it never deletes
// attachment bytes or rows.
// Invalid dependencies, input, or state and all generic DB or driver failures map to MailboxDraftRepositoryError only.
type MailboxDraftRepository interface {
	BeginEngineCommand(context.Context, MailboxEngineCommandBeginRequest) (MailboxEngineCommandBeginResult, error)
	GetMailboxDraft(context.Context, MailboxDraftGetRequest) (MailboxDraftGetResult, error)
	LookupMailboxDraftCreatePreHashLocator(context.Context, MailboxDraftCreatePreHashLocatorRequest) (MailboxDraftCreatePreHashLocatorResult, error)
	LookupMailboxDraftCreateReplay(context.Context, MailboxDraftCreateReplayLookupRequest) (MailboxDraftCreateReplayLookupResult, error)
	ReserveMailboxDraftCreate(context.Context, MailboxDraftCreateRequest) (MailboxDraftReservation, error)
	AdvanceMailboxDraftCreateToImport(context.Context, MailboxDraftAdvanceCreateToImportRequest) (MailboxDraftReservation, error)
	ReserveMailboxDraftSave(context.Context, MailboxDraftSaveRequest) (MailboxDraftReservation, error)
	ReserveMailboxDraftDelete(context.Context, MailboxDraftDeleteRequest) (MailboxDraftReservation, error)
	AbortMailboxDraftReservation(context.Context, MailboxDraftAbortRequest) (MailboxDraftAbortResult, error)
	PrepareMailboxDraftReconciliation(context.Context, MailboxDraftReconcileRequest) (MailboxDraftReconcilePlan, error)
	ResumeMailboxDraftCreateAfterAbsence(context.Context, MailboxDraftResumeCreateRequest) (MailboxDraftReservation, error)
	RestartMailboxDraftCreateUpload(context.Context, MailboxDraftRestartCreateUploadRequest) (MailboxDraftRestartCreateUploadResult, error)
	RestartMailboxDraftCreateUploadAfterBlobNotFound(context.Context, MailboxDraftRestartCreateUploadAfterBlobNotFoundRequest) (MailboxDraftRestartCreateUploadAfterBlobNotFoundResult, error)
	AbandonMailboxDraftCreateUpload(context.Context, MailboxDraftAbandonCreateUploadRequest) (MailboxDraftAbandonCreateUploadResult, error)
	FinishEngineCommand(context.Context, MailboxDraftFinishRequest) (MailboxDraftFinishResult, error)
	RecheckMailboxDraftRead(context.Context, MailboxDraftReadRecheckRequest) (MailboxDraftReadRecheckResult, error)
}
