package mail

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"reflect"
	"time"

	"github.com/google/uuid"
)

const (
	MailboxMessageMutationMaximumResolvedMessages = 200
	mailboxMessageMutationFingerprintVersion int32 = 1
)

const mailboxMessageMutationRequestHashDomain = "hub.mail.message-mutation.request.v1"

type MailboxMessageMutationTargetType string

const (
	MailboxMessageMutationTargetMessage MailboxMessageMutationTargetType = "message"
	MailboxMessageMutationTargetThread  MailboxMessageMutationTargetType = "thread"
)

type MailboxMessageMutationAction string

const (
	MailboxMessageMutationActionMove       MailboxMessageMutationAction = "move"
	MailboxMessageMutationActionMarkRead   MailboxMessageMutationAction = "mark_read"
	MailboxMessageMutationActionMarkUnread MailboxMessageMutationAction = "mark_unread"
	MailboxMessageMutationActionFlag       MailboxMessageMutationAction = "flag"
	MailboxMessageMutationActionUnflag     MailboxMessageMutationAction = "unflag"
	MailboxMessageMutationActionTrash      MailboxMessageMutationAction = "trash"
)

// MailboxMessageMutationInput is one client-idempotent, homogeneous bulk request.
// TargetIDs remain in client order because that is the durable receipt target order.
type MailboxMessageMutationInput struct {
	IdempotencyKey uuid.UUID
	TargetType     MailboxMessageMutationTargetType
	TargetIDs      []uuid.UUID
	Action         MailboxMessageMutationAction
	FolderID       *uuid.UUID
}

// MailboxMessageMutationScope is the complete authenticated account scope used by
// every repository operation in a bulk mutation.
type MailboxMessageMutationScope struct {
	ActorID         uuid.UUID
	OrganizationID  uuid.UUID
	AccountID       uuid.UUID
	EngineAccountID string
}

type MailboxMessageMutationResolvedMessage struct {
	MessageID       uuid.UUID
	ThreadID        uuid.UUID
	EngineMessageID string
}

// MailboxMessageMutationResolvedTarget retains one client target in request order
// and its immutable local message membership. An unavailable target has no members.
type MailboxMessageMutationResolvedTarget struct {
	RequestIndex int
	TargetType   MailboxMessageMutationTargetType
	TargetID     uuid.UUID
	MessageIDs   []uuid.UUID
}

// MailboxMessageMutationSnapshot is the immutable, account-scoped expansion
// persisted with the receipt before provider I/O. Messages and every member list
// are strictly ordered by raw UUID bytes, so retry/reconciliation never re-expands
// a current thread.
type MailboxMessageMutationSnapshot struct {
	Targets        []MailboxMessageMutationResolvedTarget
	Messages       []MailboxMessageMutationResolvedMessage
	EngineFolderID string
}

type MailboxMessageMutationResolutionKind string

const (
	MailboxMessageMutationResolutionResolved       MailboxMessageMutationResolutionKind = "resolved"
	MailboxMessageMutationResolutionExpansionLimit MailboxMessageMutationResolutionKind = "expansion_limit"
)

// MailboxMessageMutationResolution is a read-only, owner-scoped target expansion.
// ExpansionLimit deliberately contains no member snapshot: the repository records
// the terminal receipt without persisting any members or issuing provider I/O.
type MailboxMessageMutationResolution struct {
	Kind     MailboxMessageMutationResolutionKind
	Snapshot *MailboxMessageMutationSnapshot
}

type MailboxMessageMutationTargetStatus string

const (
	MailboxMessageMutationTargetSucceeded MailboxMessageMutationTargetStatus = "succeeded"
	MailboxMessageMutationTargetPartial   MailboxMessageMutationTargetStatus = "partial"
	MailboxMessageMutationTargetFailed    MailboxMessageMutationTargetStatus = "failed"
	MailboxMessageMutationTargetUnknown   MailboxMessageMutationTargetStatus = "unknown"
)

type MailboxMessageMutationTrainingStatus string

const MailboxMessageMutationTrainingNotRequested MailboxMessageMutationTrainingStatus = "not_requested"

type MailboxMessageMutationSafeCode string

const (
	MailboxMessageMutationCodeTargetUnavailable MailboxMessageMutationSafeCode = "MAIL_TARGET_UNAVAILABLE"
	MailboxMessageMutationCodeActionRejected    MailboxMessageMutationSafeCode = "MAIL_ACTION_REJECTED"
	MailboxMessageMutationCodeFolderUnavailable MailboxMessageMutationSafeCode = "MAIL_FOLDER_UNAVAILABLE"
	MailboxMessageMutationCodeOutcomeUnknown    MailboxMessageMutationSafeCode = "MAIL_ENGINE_OUTCOME_UNKNOWN"
)

type MailboxMessageMutationTargetResult struct {
	TargetType            MailboxMessageMutationTargetType
	TargetID              uuid.UUID
	Status                MailboxMessageMutationTargetStatus
	AffectedMessageCount  int
	SucceededMessageCount int
	FailedMessageCount    int
	TrainingStatus        MailboxMessageMutationTrainingStatus
	Code                  *MailboxMessageMutationSafeCode
}

type MailboxMessageMutationResultState string

const (
	MailboxMessageMutationResultCompleted MailboxMessageMutationResultState = "completed"
	MailboxMessageMutationResultUnknown   MailboxMessageMutationResultState = "unknown"
)

// MailboxMessageMutationResult is receipt-backed. It intentionally carries no
// mutated projection fields: sync/reconciliation remains authoritative for those.
type MailboxMessageMutationResult struct {
	State   MailboxMessageMutationResultState
	Results []MailboxMessageMutationTargetResult
}

type MailboxMessageMutationReplayKind string

const (
	MailboxMessageMutationReplayAbsent         MailboxMessageMutationReplayKind = "absent"
	MailboxMessageMutationReplayCompleted      MailboxMessageMutationReplayKind = "completed"
	MailboxMessageMutationReplayExpansionLimit MailboxMessageMutationReplayKind = "expansion_limit"
	MailboxMessageMutationReplayUnknown        MailboxMessageMutationReplayKind = "unknown"
	MailboxMessageMutationReplayInProgress     MailboxMessageMutationReplayKind = "in_progress"
)

type MailboxMessageMutationReplayRequest struct {
	Scope              MailboxMessageMutationScope
	IdempotencyKeyHash [sha256.Size]byte
	RequestHash        [sha256.Size]byte
	FingerprintVersion int32
}

type MailboxMessageMutationReplay struct {
	Kind      MailboxMessageMutationReplayKind
	Result    *MailboxMessageMutationResult
	BusyUntil *time.Time
}

type MailboxMessageMutationResolveRequest struct {
	Scope MailboxMessageMutationScope
	Input MailboxMessageMutationInput
}

type MailboxMessageMutationReserveRequest struct {
	Scope              MailboxMessageMutationScope
	Input              MailboxMessageMutationInput
	IdempotencyKeyHash [sha256.Size]byte
	RequestHash        [sha256.Size]byte
	FingerprintVersion int32
	Resolution         MailboxMessageMutationResolution
}

type MailboxMessageMutationReservationKind string

const (
	MailboxMessageMutationReservationReserved       MailboxMessageMutationReservationKind = "reserved"
	MailboxMessageMutationReservationCompleted      MailboxMessageMutationReservationKind = "completed"
	MailboxMessageMutationReservationExpansionLimit MailboxMessageMutationReservationKind = "expansion_limit"
	MailboxMessageMutationReservationUnknown        MailboxMessageMutationReservationKind = "unknown"
	MailboxMessageMutationReservationBusy           MailboxMessageMutationReservationKind = "busy"
)

// MailboxMessageMutationReservation is the sole provider-dispatch authority. A
// Reserved result has a committed receipt, immutable snapshot, and live account
// mutation fence. Every other kind must cause zero provider I/O. In particular,
// an existing in-progress receipt for the same request is Busy or Unknown forever;
// expiry of its command lease must never turn it into a second Reserved dispatch.
type MailboxMessageMutationReservation struct {
	Kind            MailboxMessageMutationReservationKind
	ReceiptID       uuid.UUID
	Fence           *MailboxEngineCommandFence
	EngineAccountID string
	Snapshot        *MailboxMessageMutationSnapshot
	Result          *MailboxMessageMutationResult
	BusyUntil       *time.Time
}

type MailboxMessageMutationFinishOutcome string

const (
	MailboxMessageMutationFinishApplied  MailboxMessageMutationFinishOutcome = "applied"
	MailboxMessageMutationFinishRejected MailboxMessageMutationFinishOutcome = "rejected"
)

type MailboxMessageMutationFinishRequest struct {
	Scope     MailboxMessageMutationScope
	ReceiptID uuid.UUID
	Fence     MailboxEngineCommandFence
	Outcome   MailboxMessageMutationFinishOutcome
}

type MailboxMessageMutationUnknownRequest struct {
	Scope     MailboxMessageMutationScope
	ReceiptID uuid.UUID
	Fence     MailboxEngineCommandFence
}

// MailboxMessageMutationRepository owns receipt replay, the immutable expansion,
// receipt target/member rows, and the account mutation lease. Reserve must recheck
// the idempotency key atomically; Lookup is only the no-I/O fast path. Finish and
// MarkUnknown may not write a local message projection as proof of provider success.
// After any provider dispatch, including a rejected provider response whose Finish
// call cannot durably complete, the receipt is fail-closed: Reserve must expose the
// same receipt only as Unknown or Busy, never as a new Reserved command. Recovery
// workers, not this request path, own any later reconciliation attempt.
type MailboxMessageMutationRepository interface {
	LookupMailboxMessageMutationReplay(context.Context, MailboxMessageMutationReplayRequest) (MailboxMessageMutationReplay, error)
	ResolveMailboxMessageMutationTargets(context.Context, MailboxMessageMutationResolveRequest) (MailboxMessageMutationResolution, error)
	ReserveMailboxMessageMutation(context.Context, MailboxMessageMutationReserveRequest) (MailboxMessageMutationReservation, error)
	FinishMailboxMessageMutation(context.Context, MailboxMessageMutationFinishRequest) (MailboxMessageMutationResult, error)
	MarkMailboxMessageMutationUnknown(context.Context, MailboxMessageMutationUnknownRequest) (MailboxMessageMutationResult, error)
}

// MailboxMessageMutationEngine is the intentionally narrow provider surface used
// by this first slice. Archive, restore, labels, classifier calls, and permanent
// delete are absent by construction.
type MailboxMessageMutationEngine interface {
	MoveMessages(context.Context, string, []string, string) error
	SetFlags(context.Context, string, []string, FlagPatch) error
	DeleteMessages(context.Context, string, []string, bool) error
}

var _ MailboxMessageMutationEngine = (*StalwartEngine)(nil)

type MailboxMessageMutationRepositoryErrorKind string

const (
	MailboxMessageMutationRepositoryErrorInvalid             MailboxMessageMutationRepositoryErrorKind = "invalid"
	MailboxMessageMutationRepositoryErrorNotFound            MailboxMessageMutationRepositoryErrorKind = "not_found"
	MailboxMessageMutationRepositoryErrorUnavailable         MailboxMessageMutationRepositoryErrorKind = "unavailable"
	MailboxMessageMutationRepositoryErrorAccountDeleting     MailboxMessageMutationRepositoryErrorKind = "account_deleting"
	MailboxMessageMutationRepositoryErrorStateConflict       MailboxMessageMutationRepositoryErrorKind = "state_conflict"
	MailboxMessageMutationRepositoryErrorFenceLost           MailboxMessageMutationRepositoryErrorKind = "fence_lost"
	MailboxMessageMutationRepositoryErrorIdempotencyConflict MailboxMessageMutationRepositoryErrorKind = "idempotency_conflict"
)

const mailboxMessageMutationRepositoryErrorText = "mailbox message mutation repository error"

type MailboxMessageMutationRepositoryError struct {
	kind MailboxMessageMutationRepositoryErrorKind
}

func NewMailboxMessageMutationRepositoryError(kind MailboxMessageMutationRepositoryErrorKind) *MailboxMessageMutationRepositoryError {
	if !validMailboxMessageMutationRepositoryErrorKind(kind) {
		kind = MailboxMessageMutationRepositoryErrorInvalid
	}
	return &MailboxMessageMutationRepositoryError{kind: kind}
}

func (*MailboxMessageMutationRepositoryError) Error() string {
	return mailboxMessageMutationRepositoryErrorText
}

func (err *MailboxMessageMutationRepositoryError) Kind() MailboxMessageMutationRepositoryErrorKind {
	if err == nil || !validMailboxMessageMutationRepositoryErrorKind(err.kind) {
		return MailboxMessageMutationRepositoryErrorInvalid
	}
	return err.kind
}

func MailboxMessageMutationRepositoryErrorKindOf(err error) (MailboxMessageMutationRepositoryErrorKind, bool) {
	var target *MailboxMessageMutationRepositoryError
	if !errors.As(err, &target) || target == nil || !validMailboxMessageMutationRepositoryErrorKind(target.kind) || target.kind == MailboxMessageMutationRepositoryErrorInvalid {
		return MailboxMessageMutationRepositoryErrorInvalid, false
	}
	return target.kind, true
}

func validMailboxMessageMutationRepositoryErrorKind(kind MailboxMessageMutationRepositoryErrorKind) bool {
	switch kind {
	case MailboxMessageMutationRepositoryErrorInvalid,
		MailboxMessageMutationRepositoryErrorNotFound,
		MailboxMessageMutationRepositoryErrorUnavailable,
		MailboxMessageMutationRepositoryErrorAccountDeleting,
		MailboxMessageMutationRepositoryErrorStateConflict,
		MailboxMessageMutationRepositoryErrorFenceLost,
		MailboxMessageMutationRepositoryErrorIdempotencyConflict:
		return true
	default:
		return false
	}
}

type MailboxMessageMutationErrorKind string

const (
	MailboxMessageMutationErrorInvalid             MailboxMessageMutationErrorKind = "invalid"
	MailboxMessageMutationErrorNotFound            MailboxMessageMutationErrorKind = "not_found"
	MailboxMessageMutationErrorDeleting            MailboxMessageMutationErrorKind = "deleting"
	MailboxMessageMutationErrorUnavailable         MailboxMessageMutationErrorKind = "unavailable"
	MailboxMessageMutationErrorConflict            MailboxMessageMutationErrorKind = "conflict"
	MailboxMessageMutationErrorIdempotencyConflict MailboxMessageMutationErrorKind = "idempotency_conflict"
	MailboxMessageMutationErrorExpansionLimit      MailboxMessageMutationErrorKind = "expansion_limit"
	MailboxMessageMutationErrorBusy                MailboxMessageMutationErrorKind = "busy"
	MailboxMessageMutationErrorOutcomeUnknown      MailboxMessageMutationErrorKind = "outcome_unknown"
)

const mailboxMessageMutationErrorText = "mailbox message mutation error"

type MailboxMessageMutationError struct {
	kind MailboxMessageMutationErrorKind
}

func NewMailboxMessageMutationError(kind MailboxMessageMutationErrorKind) *MailboxMessageMutationError {
	if !validMailboxMessageMutationErrorKind(kind) {
		kind = MailboxMessageMutationErrorInvalid
	}
	return &MailboxMessageMutationError{kind: kind}
}

func (*MailboxMessageMutationError) Error() string {
	return mailboxMessageMutationErrorText
}

func (err *MailboxMessageMutationError) Kind() MailboxMessageMutationErrorKind {
	if err == nil || !validMailboxMessageMutationErrorKind(err.kind) {
		return MailboxMessageMutationErrorInvalid
	}
	return err.kind
}

func MailboxMessageMutationErrorKindOf(err error) (MailboxMessageMutationErrorKind, bool) {
	var target *MailboxMessageMutationError
	if !errors.As(err, &target) || target == nil || !validMailboxMessageMutationErrorKind(target.kind) || target.kind == MailboxMessageMutationErrorInvalid {
		return MailboxMessageMutationErrorInvalid, false
	}
	return target.kind, true
}

func mailboxMessageMutationError(kind MailboxMessageMutationErrorKind) error {
	return NewMailboxMessageMutationError(kind)
}

func validMailboxMessageMutationErrorKind(kind MailboxMessageMutationErrorKind) bool {
	switch kind {
	case MailboxMessageMutationErrorInvalid,
		MailboxMessageMutationErrorNotFound,
		MailboxMessageMutationErrorDeleting,
		MailboxMessageMutationErrorUnavailable,
		MailboxMessageMutationErrorConflict,
		MailboxMessageMutationErrorIdempotencyConflict,
		MailboxMessageMutationErrorExpansionLimit,
		MailboxMessageMutationErrorBusy,
		MailboxMessageMutationErrorOutcomeUnknown:
		return true
	default:
		return false
	}
}

func mailboxMessageMutationDependencyNil(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func mailboxMessageMutationInputValid(input MailboxMessageMutationInput) bool {
	if !mailboxAttachmentPresignIdempotencyKeyValid(input.IdempotencyKey) ||
		!validMailboxMessageMutationTargetType(input.TargetType) ||
		!validMailboxMessageMutationAction(input.Action) ||
		len(input.TargetIDs) < 1 || len(input.TargetIDs) > MailboxMessageMutationMaximumResolvedMessages {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(input.TargetIDs))
	for _, targetID := range input.TargetIDs {
		if targetID == uuid.Nil {
			return false
		}
		if _, exists := seen[targetID]; exists {
			return false
		}
		seen[targetID] = struct{}{}
	}
	if input.Action == MailboxMessageMutationActionMove {
		return input.FolderID != nil && *input.FolderID != uuid.Nil
	}
	return input.FolderID == nil
}

func validMailboxMessageMutationTargetType(value MailboxMessageMutationTargetType) bool {
	return value == MailboxMessageMutationTargetMessage || value == MailboxMessageMutationTargetThread
}

func validMailboxMessageMutationAction(value MailboxMessageMutationAction) bool {
	switch value {
	case MailboxMessageMutationActionMove,
		MailboxMessageMutationActionMarkRead,
		MailboxMessageMutationActionMarkUnread,
		MailboxMessageMutationActionFlag,
		MailboxMessageMutationActionUnflag,
		MailboxMessageMutationActionTrash:
		return true
	default:
		return false
	}
}

func mailboxMessageMutationScopeValid(scope MailboxMessageMutationScope) bool {
	return scope.ActorID != uuid.Nil && scope.OrganizationID != uuid.Nil && scope.AccountID != uuid.Nil && syncValidOpaque(scope.EngineAccountID, MaxEngineOpaqueIDBytes)
}

func mailboxMessageMutationInputCopy(input MailboxMessageMutationInput) MailboxMessageMutationInput {
	copy := input
	copy.TargetIDs = append([]uuid.UUID(nil), input.TargetIDs...)
	if input.FolderID != nil {
		folderID := *input.FolderID
		copy.FolderID = &folderID
	}
	return copy
}

func mailboxMessageMutationRequestHash(input MailboxMessageMutationInput, scope MailboxMessageMutationScope) [sha256.Size]byte {
	targetFrames := make([][]byte, 0, len(input.TargetIDs))
	for _, targetID := range input.TargetIDs {
		targetFrames = append(targetFrames, canonicalTLV(1, canonicalUUID(targetID)))
	}
	folder := []byte{0}
	if input.FolderID != nil {
		folder = append([]byte{1}, canonicalUUID(*input.FolderID)...)
	}
	mac := hmac.New(sha256.New, input.IdempotencyKey[:])
	_, _ = mac.Write(canonicalRecord(
		canonicalTLV(1, []byte(mailboxMessageMutationRequestHashDomain)),
		canonicalTLV(2, canonicalUUID(scope.ActorID)),
		canonicalTLV(3, canonicalUUID(scope.OrganizationID)),
		canonicalTLV(4, canonicalUUID(scope.AccountID)),
		canonicalTLV(5, []byte(input.TargetType)),
		canonicalTLV(6, canonicalRecord(targetFrames...)),
		canonicalTLV(7, []byte(input.Action)),
		canonicalTLV(8, folder),
	))
	var result [sha256.Size]byte
	copy(result[:], mac.Sum(nil))
	return result
}

func mailboxMessageMutationResolutionValid(resolution MailboxMessageMutationResolution, input MailboxMessageMutationInput) bool {
	if !mailboxMessageMutationInputValid(input) {
		return false
	}
	switch resolution.Kind {
	case MailboxMessageMutationResolutionExpansionLimit:
		return resolution.Snapshot == nil
	case MailboxMessageMutationResolutionResolved:
		return resolution.Snapshot != nil && mailboxMessageMutationSnapshotValid(*resolution.Snapshot, input)
	default:
		return false
	}
}

func mailboxMessageMutationSnapshotValid(snapshot MailboxMessageMutationSnapshot, input MailboxMessageMutationInput) bool {
	if len(snapshot.Targets) != len(input.TargetIDs) || len(snapshot.Messages) > MailboxMessageMutationMaximumResolvedMessages {
		return false
	}
	if input.Action == MailboxMessageMutationActionMove {
		if !validOpaqueID(snapshot.EngineFolderID) {
			return false
		}
	} else if snapshot.EngineFolderID != "" {
		return false
	}

	messages := make(map[uuid.UUID]MailboxMessageMutationResolvedMessage, len(snapshot.Messages))
	engineIDs := make(map[string]struct{}, len(snapshot.Messages))
	for index, message := range snapshot.Messages {
		if message.MessageID == uuid.Nil || message.ThreadID == uuid.Nil || !validOpaqueID(message.EngineMessageID) ||
			index > 0 && !mailboxMessageMutationUUIDLess(snapshot.Messages[index-1].MessageID, message.MessageID) {
			return false
		}
		if _, exists := messages[message.MessageID]; exists {
			return false
		}
		if _, exists := engineIDs[message.EngineMessageID]; exists {
			return false
		}
		messages[message.MessageID] = message
		engineIDs[message.EngineMessageID] = struct{}{}
	}

	memberCount := make(map[uuid.UUID]int, len(snapshot.Messages))
	for index, target := range snapshot.Targets {
		if target.RequestIndex != index || target.TargetType != input.TargetType || target.TargetID != input.TargetIDs[index] ||
			!validMailboxMessageMutationTargetType(target.TargetType) || target.TargetID == uuid.Nil ||
			!mailboxMessageMutationUUIDsStrictlySorted(target.MessageIDs) {
			return false
		}
		if target.TargetType == MailboxMessageMutationTargetMessage && len(target.MessageIDs) > 0 && (len(target.MessageIDs) != 1 || target.MessageIDs[0] != target.TargetID) {
			return false
		}
		for _, messageID := range target.MessageIDs {
			message, exists := messages[messageID]
			if !exists || target.TargetType == MailboxMessageMutationTargetThread && message.ThreadID != target.TargetID {
				return false
			}
			memberCount[messageID]++
		}
	}
	for messageID := range messages {
		if memberCount[messageID] == 0 {
			return false
		}
	}
	return true
}

func mailboxMessageMutationSnapshotCopy(snapshot MailboxMessageMutationSnapshot) MailboxMessageMutationSnapshot {
	copy := snapshot
	copy.Targets = make([]MailboxMessageMutationResolvedTarget, len(snapshot.Targets))
	for index, target := range snapshot.Targets {
		copy.Targets[index] = target
		copy.Targets[index].MessageIDs = append([]uuid.UUID(nil), target.MessageIDs...)
	}
	copy.Messages = append([]MailboxMessageMutationResolvedMessage(nil), snapshot.Messages...)
	return copy
}

func mailboxMessageMutationResolutionCopy(resolution MailboxMessageMutationResolution) MailboxMessageMutationResolution {
	copy := resolution
	if resolution.Snapshot != nil {
		snapshot := mailboxMessageMutationSnapshotCopy(*resolution.Snapshot)
		copy.Snapshot = &snapshot
	}
	return copy
}

func mailboxMessageMutationSnapshotEqual(left, right MailboxMessageMutationSnapshot, input MailboxMessageMutationInput) bool {
	return mailboxMessageMutationSnapshotValid(left, input) && mailboxMessageMutationSnapshotValid(right, input) && reflect.DeepEqual(mailboxMessageMutationSnapshotCopy(left), mailboxMessageMutationSnapshotCopy(right))
}

func mailboxMessageMutationUUIDsStrictlySorted(values []uuid.UUID) bool {
	for index, value := range values {
		if value == uuid.Nil || index > 0 && !mailboxMessageMutationUUIDLess(values[index-1], value) {
			return false
		}
	}
	return true
}

func mailboxMessageMutationUUIDLess(left, right uuid.UUID) bool {
	for index := range left {
		if left[index] != right[index] {
			return left[index] < right[index]
		}
	}
	return false
}

func mailboxMessageMutationEngineMessageIDs(snapshot MailboxMessageMutationSnapshot, input MailboxMessageMutationInput) []string {
	if !mailboxMessageMutationSnapshotValid(snapshot, input) || len(snapshot.Messages) == 0 {
		return nil
	}
	ids := make([]string, len(snapshot.Messages))
	for index, message := range snapshot.Messages {
		ids[index] = message.EngineMessageID
	}
	return ids
}

func mailboxMessageMutationResultValid(result MailboxMessageMutationResult, input MailboxMessageMutationInput) bool {
	if !mailboxMessageMutationInputValid(input) || len(result.Results) != len(input.TargetIDs) {
		return false
	}
	unknown := false
	for index, target := range result.Results {
		if target.TargetType != input.TargetType || target.TargetID != input.TargetIDs[index] ||
			target.TrainingStatus != MailboxMessageMutationTrainingNotRequested ||
			!mailboxMessageMutationTargetResultValid(target) {
			return false
		}
		unknown = unknown || target.Status == MailboxMessageMutationTargetUnknown
	}
	switch result.State {
	case MailboxMessageMutationResultCompleted:
		return !unknown
	case MailboxMessageMutationResultUnknown:
		return unknown
	default:
		return false
	}
}

func mailboxMessageMutationTargetResultValid(target MailboxMessageMutationTargetResult) bool {
	if !validMailboxMessageMutationTargetType(target.TargetType) || target.TargetID == uuid.Nil ||
		target.AffectedMessageCount < 0 || target.AffectedMessageCount > MailboxMessageMutationMaximumResolvedMessages ||
		target.SucceededMessageCount < 0 || target.SucceededMessageCount > MailboxMessageMutationMaximumResolvedMessages ||
		target.FailedMessageCount < 0 || target.FailedMessageCount > MailboxMessageMutationMaximumResolvedMessages ||
		target.SucceededMessageCount+target.FailedMessageCount > target.AffectedMessageCount {
		return false
	}
	code := MailboxMessageMutationSafeCode("")
	if target.Code != nil {
		code = *target.Code
	}
	switch target.Status {
	case MailboxMessageMutationTargetSucceeded:
		return target.AffectedMessageCount >= 1 && target.SucceededMessageCount == target.AffectedMessageCount && target.FailedMessageCount == 0 && target.Code == nil
	case MailboxMessageMutationTargetPartial:
		return target.AffectedMessageCount >= 2 && target.SucceededMessageCount >= 1 && target.FailedMessageCount >= 1 && target.SucceededMessageCount+target.FailedMessageCount == target.AffectedMessageCount && code == MailboxMessageMutationCodeActionRejected
	case MailboxMessageMutationTargetFailed:
		if target.AffectedMessageCount == 0 {
			return target.SucceededMessageCount == 0 && target.FailedMessageCount == 0 && (code == MailboxMessageMutationCodeTargetUnavailable || code == MailboxMessageMutationCodeFolderUnavailable)
		}
		return target.SucceededMessageCount == 0 && target.FailedMessageCount == target.AffectedMessageCount && code == MailboxMessageMutationCodeActionRejected
	case MailboxMessageMutationTargetUnknown:
		return target.AffectedMessageCount >= 1 && target.SucceededMessageCount+target.FailedMessageCount < target.AffectedMessageCount && code == MailboxMessageMutationCodeOutcomeUnknown
	default:
		return false
	}
}

func mailboxMessageMutationResultCopy(result MailboxMessageMutationResult) MailboxMessageMutationResult {
	copy := result
	copy.Results = make([]MailboxMessageMutationTargetResult, len(result.Results))
	for index, target := range result.Results {
		copy.Results[index] = target
		if target.Code != nil {
			code := *target.Code
			copy.Results[index].Code = &code
		}
	}
	return copy
}

func mailboxMessageMutationReplayValid(replay MailboxMessageMutationReplay, input MailboxMessageMutationInput) bool {
	switch replay.Kind {
	case MailboxMessageMutationReplayAbsent:
		return replay.Result == nil && replay.BusyUntil == nil
	case MailboxMessageMutationReplayCompleted:
		return replay.Result != nil && replay.Result.State == MailboxMessageMutationResultCompleted && mailboxMessageMutationResultValid(*replay.Result, input) && replay.BusyUntil == nil
	case MailboxMessageMutationReplayExpansionLimit:
		return replay.Result == nil && replay.BusyUntil == nil
	case MailboxMessageMutationReplayUnknown:
		return replay.Result != nil && replay.Result.State == MailboxMessageMutationResultUnknown && mailboxMessageMutationResultValid(*replay.Result, input) && replay.BusyUntil == nil
	case MailboxMessageMutationReplayInProgress:
		return replay.Result == nil && mailboxMessageMutationBusyUntilValid(replay.BusyUntil)
	default:
		return false
	}
}

func mailboxMessageMutationReservationValid(reservation MailboxMessageMutationReservation, request MailboxMessageMutationReserveRequest) bool {
	if !mailboxMessageMutationScopeValid(request.Scope) || !mailboxMessageMutationInputValid(request.Input) || !mailboxMessageMutationResolutionValid(request.Resolution, request.Input) {
		return false
	}
	switch reservation.Kind {
	case MailboxMessageMutationReservationReserved:
		return reservation.ReceiptID != uuid.Nil && reservation.Fence != nil &&
			reservation.EngineAccountID == request.Scope.EngineAccountID && reservation.Fence.AccountID == request.Scope.AccountID &&
			mailboxEntEngineCommandFenceValid(*reservation.Fence) && reservation.Snapshot != nil &&
			request.Resolution.Snapshot != nil && mailboxMessageMutationSnapshotEqual(*reservation.Snapshot, *request.Resolution.Snapshot, request.Input) &&
			reservation.Result == nil && reservation.BusyUntil == nil && len(reservation.Snapshot.Messages) >= 1
	case MailboxMessageMutationReservationCompleted:
		return reservation.ReceiptID == uuid.Nil && reservation.Fence == nil && reservation.EngineAccountID == "" && reservation.Snapshot == nil &&
			reservation.Result != nil && reservation.Result.State == MailboxMessageMutationResultCompleted && mailboxMessageMutationResultValid(*reservation.Result, request.Input) && reservation.BusyUntil == nil
	case MailboxMessageMutationReservationExpansionLimit:
		return reservation.ReceiptID == uuid.Nil && reservation.Fence == nil && reservation.EngineAccountID == "" && reservation.Snapshot == nil && reservation.Result == nil && reservation.BusyUntil == nil
	case MailboxMessageMutationReservationUnknown:
		return reservation.ReceiptID == uuid.Nil && reservation.Fence == nil && reservation.EngineAccountID == "" && reservation.Snapshot == nil &&
			reservation.Result != nil && reservation.Result.State == MailboxMessageMutationResultUnknown && mailboxMessageMutationResultValid(*reservation.Result, request.Input) && reservation.BusyUntil == nil
	case MailboxMessageMutationReservationBusy:
		return reservation.ReceiptID == uuid.Nil && reservation.Fence == nil && reservation.EngineAccountID == "" && reservation.Snapshot == nil && reservation.Result == nil && mailboxMessageMutationBusyUntilValid(reservation.BusyUntil)
	default:
		return false
	}
}

func mailboxMessageMutationBusyUntilValid(value *time.Time) bool {
	return value != nil && !value.IsZero()
}
