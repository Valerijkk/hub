package mail

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

// MailboxSubmissionStatus is the durable, provider-neutral submission lifecycle.
type MailboxSubmissionStatus string

const (
	MailboxSubmissionStatusQueued     MailboxSubmissionStatus = "queued"
	MailboxSubmissionStatusSubmitting MailboxSubmissionStatus = "submitting"
	MailboxSubmissionStatusSent       MailboxSubmissionStatus = "sent"
	MailboxSubmissionStatusUnknown    MailboxSubmissionStatus = "unknown"
	MailboxSubmissionStatusFailed     MailboxSubmissionStatus = "failed"
)

// MailboxSubmissionErrorCode is the bounded public reason carried by terminal
// and reconciliation-pending submission states.
type MailboxSubmissionErrorCode string

const (
	MailboxSubmissionErrorCodeOutcomeUnknown          MailboxSubmissionErrorCode = "MAIL_SUBMISSION_OUTCOME_UNKNOWN"
	MailboxSubmissionErrorCodeReconciliationRequired  MailboxSubmissionErrorCode = "MAIL_SUBMISSION_RECONCILIATION_REQUIRED"
	MailboxSubmissionErrorCodeMailboxUnavailable       MailboxSubmissionErrorCode = "MAILBOX_UNAVAILABLE"
	MailboxSubmissionErrorCodeRejected                 MailboxSubmissionErrorCode = "MAIL_SUBMISSION_REJECTED"
	MailboxSubmissionErrorCodePolicy                   MailboxSubmissionErrorCode = "MAIL_SUBMISSION_POLICY"
	MailboxSubmissionErrorCodeAttachmentNotClean       MailboxSubmissionErrorCode = "MAIL_SUBMISSION_ATTACHMENT_NOT_CLEAN"
	MailboxSubmissionErrorCodeIdentityUnavailable      MailboxSubmissionErrorCode = "MAIL_SUBMISSION_IDENTITY_UNAVAILABLE"
	MailboxSubmissionErrorCodeRecipientLimit           MailboxSubmissionErrorCode = "MAIL_SUBMISSION_RECIPIENT_LIMIT"
	MailboxSubmissionErrorCodeSizeLimit                MailboxSubmissionErrorCode = "MAIL_SUBMISSION_SIZE_LIMIT"
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
	AccountID           uuid.UUID
	DraftID             uuid.UUID
	ExpectedRevision    int64
	IdempotencyKeyHash  [sha256.Size]byte
	RequestHash         [sha256.Size]byte
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
		MailboxSubmissionStatusSubmitting,
		MailboxSubmissionStatusSent,
		MailboxSubmissionStatusUnknown,
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
		MailboxSubmissionErrorCodeMailboxUnavailable,
		MailboxSubmissionErrorCodeRejected,
		MailboxSubmissionErrorCodePolicy,
		MailboxSubmissionErrorCodeAttachmentNotClean,
		MailboxSubmissionErrorCodeIdentityUnavailable,
		MailboxSubmissionErrorCodeRecipientLimit,
		MailboxSubmissionErrorCodeSizeLimit:
		return true
	default:
		return false
	}
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
