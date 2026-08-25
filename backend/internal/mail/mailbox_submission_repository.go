package mail

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"time"

	"github.com/google/uuid"
)

// MailboxSubmissionStatus is the durable, provider-neutral submission lifecycle.
type MailboxSubmissionStatus string

const (
	MailboxSubmissionStatusQueued      MailboxSubmissionStatus = "queued"
	MailboxSubmissionStatusAuthorizing MailboxSubmissionStatus = "authorizing"
	MailboxSubmissionStatusSubmitting  MailboxSubmissionStatus = "submitting"
	MailboxSubmissionStatusUnknown     MailboxSubmissionStatus = "unknown"
	MailboxSubmissionStatusAccepted    MailboxSubmissionStatus = "accepted"
	MailboxSubmissionStatusSent        MailboxSubmissionStatus = "sent"
	MailboxSubmissionStatusFailed      MailboxSubmissionStatus = "failed"
)

// MailboxSubmissionErrorCode is the bounded public reason carried by terminal
// and reconciliation-pending submission states.
type MailboxSubmissionErrorCode string

const (
	MailboxSubmissionErrorCodeOutcomeUnknown          MailboxSubmissionErrorCode = "MAIL_SUBMISSION_OUTCOME_UNKNOWN"
	MailboxSubmissionErrorCodeReconciliationRequired  MailboxSubmissionErrorCode = "MAIL_SUBMISSION_RECONCILIATION_REQUIRED"
	MailboxSubmissionErrorCodeReconciliationExhausted MailboxSubmissionErrorCode = "MAIL_SUBMISSION_RECONCILIATION_EXHAUSTED"
	MailboxSubmissionErrorCodeMailboxUnavailable      MailboxSubmissionErrorCode = "MAILBOX_UNAVAILABLE"
	MailboxSubmissionErrorCodeRejected                MailboxSubmissionErrorCode = "MAIL_SUBMISSION_REJECTED"
	MailboxSubmissionErrorCodePolicy                  MailboxSubmissionErrorCode = "MAIL_SUBMISSION_POLICY"
	MailboxSubmissionErrorCodeAttachmentNotClean      MailboxSubmissionErrorCode = "MAIL_SUBMISSION_ATTACHMENT_NOT_CLEAN"
	MailboxSubmissionErrorCodeIdentityUnavailable     MailboxSubmissionErrorCode = "MAIL_SUBMISSION_IDENTITY_UNAVAILABLE"
	MailboxSubmissionErrorCodeRecipientLimit          MailboxSubmissionErrorCode = "MAIL_SUBMISSION_RECIPIENT_LIMIT"
	MailboxSubmissionErrorCodeSizeLimit               MailboxSubmissionErrorCode = "MAIL_SUBMISSION_SIZE_LIMIT"
	MailboxSubmissionErrorCodeRetryExhausted          MailboxSubmissionErrorCode = "MAIL_SUBMISSION_RETRY_EXHAUSTED"
)

// MailboxSubmissionRecord is the content-free durable receipt returned across
// the submission repository boundary. It never exposes opaque provider IDs or
// the provider submission key.
type MailboxSubmissionRecord struct {
	ID                  uuid.UUID
	AccountID           uuid.UUID
	DraftID             *uuid.UUID
	Status              MailboxSubmissionStatus
	ResultMailMessageID *uuid.UUID
	LastErrorCode       *MailboxSubmissionErrorCode
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// MailboxSubmissionCreateRequest is one owner-scoped durable submission
// intent. The provider key is deliberately absent: only the repository creates
// it after it has ruled out an idempotency replay.
type MailboxSubmissionCreateRequest struct {
	AccountID          uuid.UUID
	DraftID            uuid.UUID
	ExpectedRevision   int64
	IdempotencyKeyHash [sha256.Size]byte
	RequestHash        [sha256.Size]byte
}

// MailboxSubmissionReservationKind distinguishes an exact idempotency replay
// from a newly persisted queued intent.
type MailboxSubmissionReservationKind string

const (
	MailboxSubmissionReservationReplay MailboxSubmissionReservationKind = "replay"
	MailboxSubmissionReservationQueued MailboxSubmissionReservationKind = "queued"
)

// MailboxSubmissionReservation contains exactly one validated receipt.
type MailboxSubmissionReservation struct {
	Kind       MailboxSubmissionReservationKind
	Submission MailboxSubmissionRecord
}

// MailboxSubmissionRepository persists and replays a submission intent without
// provider I/O. Implementations must serialize the account, receipt, draft and
// any active submission before creating a new queued record.
type MailboxSubmissionRepository interface {
	ReserveMailboxSubmission(context.Context, MailboxSubmissionCreateRequest) (MailboxSubmissionReservation, error)
}

// MailboxSubmissionReadRequest scopes one durable submission receipt to the
// authenticated mailbox owner. It carries identifiers only; provider and
// draft content never cross the read boundary.
type MailboxSubmissionReadRequest struct {
	ActorID        uuid.UUID
	OrganizationID uuid.UUID
	AccountID      uuid.UUID
	SubmissionID   uuid.UUID
}

// MailboxSubmissionReader reads one owner-scoped durable submission receipt.
// It is deliberately separate from MailboxSubmissionRepository so existing
// reservation fakes do not need to broaden their contract.
type MailboxSubmissionReader interface {
	ReadMailboxSubmission(context.Context, MailboxSubmissionReadRequest) (MailboxSubmissionRecord, error)
}

// MailboxSubmissionRepositoryErrorKind is a closed, content-free failure
// family for durable submission reservation.
type MailboxSubmissionRepositoryErrorKind string

const (
	MailboxSubmissionRepositoryErrorInvalid              MailboxSubmissionRepositoryErrorKind = "invalid"
	MailboxSubmissionRepositoryErrorNotFound             MailboxSubmissionRepositoryErrorKind = "not_found"
	MailboxSubmissionRepositoryErrorUnavailable          MailboxSubmissionRepositoryErrorKind = "unavailable"
	MailboxSubmissionRepositoryErrorAccountDeleting      MailboxSubmissionRepositoryErrorKind = "account_deleting"
	MailboxSubmissionRepositoryErrorVersionConflict      MailboxSubmissionRepositoryErrorKind = "version_conflict"
	MailboxSubmissionRepositoryErrorSubmissionInProgress MailboxSubmissionRepositoryErrorKind = "submission_in_progress"
	MailboxSubmissionRepositoryErrorDraftNotSendable     MailboxSubmissionRepositoryErrorKind = "draft_not_sendable"
	MailboxSubmissionRepositoryErrorIdempotencyConflict  MailboxSubmissionRepositoryErrorKind = "idempotency_conflict"
)

const mailboxSubmissionRepositoryErrorText = "mailbox submission repository error"

// MailboxSubmissionRepositoryError carries only a validated failure kind.
type MailboxSubmissionRepositoryError struct {
	kind MailboxSubmissionRepositoryErrorKind
}

// NewMailboxSubmissionRepositoryError creates a content-free submission
// repository error.
func NewMailboxSubmissionRepositoryError(kind MailboxSubmissionRepositoryErrorKind) *MailboxSubmissionRepositoryError {
	if !validMailboxSubmissionRepositoryErrorKind(kind) {
		kind = MailboxSubmissionRepositoryErrorInvalid
	}
	return &MailboxSubmissionRepositoryError{kind: kind}
}

func (err *MailboxSubmissionRepositoryError) Error() string {
	return mailboxSubmissionRepositoryErrorText
}

// Kind returns the validated repository failure kind.
func (err *MailboxSubmissionRepositoryError) Kind() MailboxSubmissionRepositoryErrorKind {
	if err == nil || !validMailboxSubmissionRepositoryErrorKind(err.kind) {
		return MailboxSubmissionRepositoryErrorInvalid
	}
	return err.kind
}

func validMailboxSubmissionRepositoryErrorKind(kind MailboxSubmissionRepositoryErrorKind) bool {
	switch kind {
	case MailboxSubmissionRepositoryErrorInvalid,
		MailboxSubmissionRepositoryErrorNotFound,
		MailboxSubmissionRepositoryErrorUnavailable,
		MailboxSubmissionRepositoryErrorAccountDeleting,
		MailboxSubmissionRepositoryErrorVersionConflict,
		MailboxSubmissionRepositoryErrorSubmissionInProgress,
		MailboxSubmissionRepositoryErrorDraftNotSendable,
		MailboxSubmissionRepositoryErrorIdempotencyConflict:
		return true
	default:
		return false
	}
}

func mailboxSubmissionHashPresent(value [sha256.Size]byte) bool {
	var zero [sha256.Size]byte
	return value != zero
}

func validMailboxSubmissionStatus(status MailboxSubmissionStatus) bool {
	switch status {
	case MailboxSubmissionStatusQueued,
		MailboxSubmissionStatusAuthorizing,
		MailboxSubmissionStatusSubmitting,
		MailboxSubmissionStatusUnknown,
		MailboxSubmissionStatusAccepted,
		MailboxSubmissionStatusSent,
		MailboxSubmissionStatusFailed:
		return true
	default:
		return false
	}
}

func validMailboxSubmissionErrorCode(code MailboxSubmissionErrorCode) bool {
	switch code {
	case MailboxSubmissionErrorCodeOutcomeUnknown,
		MailboxSubmissionErrorCodeReconciliationRequired,
		MailboxSubmissionErrorCodeReconciliationExhausted,
		MailboxSubmissionErrorCodeMailboxUnavailable,
		MailboxSubmissionErrorCodeRejected,
		MailboxSubmissionErrorCodePolicy,
		MailboxSubmissionErrorCodeAttachmentNotClean,
		MailboxSubmissionErrorCodeIdentityUnavailable,
		MailboxSubmissionErrorCodeRecipientLimit,
		MailboxSubmissionErrorCodeSizeLimit,
		MailboxSubmissionErrorCodeRetryExhausted:
		return true
	default:
		return false
	}
}

const (
	mailboxEngineSubmissionBindingVersion = 1
	mailboxEngineSubmissionBindingDomain  = "hub.mail.engine_submission.binding.v1"
)

// mailboxEngineSubmissionBinding is the private immutable engine command
// snapshot. It is persisted with a hash so later dispatch never derives a
// command from mutable account, draft, or policy rows.
type mailboxEngineSubmissionBinding struct {
	engineBindingVersion  int
	engineAccountID       string
	engineDraftKey        string
	engineMessageID       string
	engineExpectedVersion string
	engineMaxMessageBytes int64
	engineSubmissionKey   string
	engineBindingHash     [sha256.Size]byte
}

func newMailboxEngineSubmissionBinding(accountID, draftKey, messageID, expectedVersion string, maxMessageBytes int64, submissionKey string) (mailboxEngineSubmissionBinding, bool) {
	binding := mailboxEngineSubmissionBinding{
		engineBindingVersion:  mailboxEngineSubmissionBindingVersion,
		engineAccountID:       accountID,
		engineDraftKey:        draftKey,
		engineMessageID:       messageID,
		engineExpectedVersion: expectedVersion,
		engineMaxMessageBytes: maxMessageBytes,
		engineSubmissionKey:   submissionKey,
	}
	if !mailboxEngineSubmissionBindingFieldsValid(binding) {
		return mailboxEngineSubmissionBinding{}, false
	}
	hash, ok := mailboxEngineSubmissionBindingHash(binding)
	if !ok {
		return mailboxEngineSubmissionBinding{}, false
	}
	binding.engineBindingHash = hash
	return binding, true
}

func mailboxEngineSubmissionBindingValid(binding mailboxEngineSubmissionBinding) bool {
	if !mailboxEngineSubmissionBindingFieldsValid(binding) {
		return false
	}
	expected, ok := mailboxEngineSubmissionBindingHash(binding)
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare(binding.engineBindingHash[:], expected[:]) == 1
}

func mailboxEngineSubmissionBindingFieldsValid(binding mailboxEngineSubmissionBinding) bool {
	return binding.engineBindingVersion == mailboxEngineSubmissionBindingVersion &&
		syncValidOpaque(binding.engineAccountID, MaxEngineOpaqueIDBytes) &&
		syncValidOpaque(binding.engineDraftKey, MaxEngineOpaqueIDBytes) &&
		syncValidOpaque(binding.engineMessageID, MaxEngineOpaqueIDBytes) &&
		syncValidOpaque(binding.engineExpectedVersion, MaxEngineOpaqueIDBytes) &&
		binding.engineMaxMessageBytes >= 1 && binding.engineMaxMessageBytes <= MaxEngineMessageBytes &&
		syncValidOpaque(binding.engineSubmissionKey, MaxEngineOpaqueIDBytes)
}

// mailboxEngineSubmissionBindingHash is SHA-256 over the exact V1 binary
// encoding: frame(domain), U32BE(version), frame(account), frame(draft),
// frame(message), frame(expected version), frame(submission key), and
// U64BE(max message bytes). A frame is U32BE byte length followed by bytes.
func mailboxEngineSubmissionBindingHash(binding mailboxEngineSubmissionBinding) ([sha256.Size]byte, bool) {
	if !mailboxEngineSubmissionBindingFieldsValid(binding) {
		return [sha256.Size]byte{}, false
	}
	encoded := make([]byte, 0, 4+len(mailboxEngineSubmissionBindingDomain)+4+
		4+len(binding.engineAccountID)+4+len(binding.engineDraftKey)+4+len(binding.engineMessageID)+
		4+len(binding.engineExpectedVersion)+4+len(binding.engineSubmissionKey)+8)
	var ok bool
	encoded, ok = mailboxEngineSubmissionAppendFrame(encoded, mailboxEngineSubmissionBindingDomain)
	if !ok {
		return [sha256.Size]byte{}, false
	}
	var version [4]byte
	binary.BigEndian.PutUint32(version[:], uint32(binding.engineBindingVersion))
	encoded = append(encoded, version[:]...)
	for _, value := range []string{binding.engineAccountID, binding.engineDraftKey, binding.engineMessageID, binding.engineExpectedVersion, binding.engineSubmissionKey} {
		encoded, ok = mailboxEngineSubmissionAppendFrame(encoded, value)
		if !ok {
			return [sha256.Size]byte{}, false
		}
	}
	var maxMessageBytes [8]byte
	binary.BigEndian.PutUint64(maxMessageBytes[:], uint64(binding.engineMaxMessageBytes))
	encoded = append(encoded, maxMessageBytes[:]...)
	return sha256.Sum256(encoded), true
}

func mailboxEngineSubmissionAppendFrame(encoded []byte, value string) ([]byte, bool) {
	return mailboxEngineSubmissionAppendFrameBytes(encoded, []byte(value))
}

func mailboxEngineSubmissionAppendFrameBytes(encoded, value []byte) ([]byte, bool) {
	if uint64(len(value)) > uint64(^uint32(0)) {
		return nil, false
	}
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value)))
	encoded = append(encoded, length[:]...)
	return append(encoded, value...), true
}

func mailboxSubmissionUUIDPointerCopy(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func mailboxSubmissionErrorCodePointerCopy(value *MailboxSubmissionErrorCode) *MailboxSubmissionErrorCode {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func mailboxSubmissionRecordCopy(record MailboxSubmissionRecord) MailboxSubmissionRecord {
	return MailboxSubmissionRecord{
		ID:                  record.ID,
		AccountID:           record.AccountID,
		DraftID:             mailboxSubmissionUUIDPointerCopy(record.DraftID),
		Status:              record.Status,
		ResultMailMessageID: mailboxSubmissionUUIDPointerCopy(record.ResultMailMessageID),
		LastErrorCode:       mailboxSubmissionErrorCodePointerCopy(record.LastErrorCode),
		CreatedAt:           record.CreatedAt.UTC(),
		UpdatedAt:           record.UpdatedAt.UTC(),
	}
}
