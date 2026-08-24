package mail

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

const mailboxSubmissionMaxIdempotencyKeyBytes = 128

// MailboxSubmissionCreateInput is the public submission-intent input. The
// immutable draft revision prevents a send request from binding newer content.
type MailboxSubmissionCreateInput struct {
	ExpectedRevision int64
	IdempotencyKey   string
}

// MailboxSubmissionView is the owner-safe submission receipt. Opaque provider
// IDs, provider keys and draft content are intentionally absent.
type MailboxSubmissionView struct {
	ID        uuid.UUID
	Status    MailboxSubmissionStatus
	MessageID *uuid.UUID
	ErrorCode *MailboxSubmissionErrorCode
	CreatedAt time.Time
	UpdatedAt time.Time
}

// MailboxSubmissionService creates durable submission intents only. Provider
// dispatch, reconciliation, projection and outbox emission remain separate
// atoms.
type MailboxSubmissionService struct {
	owners     mailboxOwnerResolver
	repository MailboxSubmissionRepository
}

// NewMailboxSubmissionService creates the owner-gated durable submission
// reservation service.
func NewMailboxSubmissionService(owners mailboxOwnerResolver, repository MailboxSubmissionRepository) (*MailboxSubmissionService, error) {
	if mailboxSubmissionDependencyNil(owners) || mailboxSubmissionDependencyNil(repository) {
		return nil, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
	}
	return &MailboxSubmissionService{owners: owners, repository: repository}, nil
}

// CreateMailboxSubmission atomically records or replays one send intent. It
// performs no provider operation and never acquires an engine-command lease.
func (service *MailboxSubmissionService) CreateMailboxSubmission(ctx context.Context, actorID, draftID uuid.UUID, input MailboxSubmissionCreateInput) (MailboxSubmissionView, error) {
	if !mailboxSubmissionServiceValid(service) || ctx == nil || actorID == uuid.Nil || draftID == uuid.Nil || input.ExpectedRevision < 1 {
		return MailboxSubmissionView{}, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
	}
	if ctx.Err() != nil {
		return MailboxSubmissionView{}, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorUnavailable)
	}

	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxSubmissionView{}, mailboxSubmissionOwnerError(err)
	}
	scope, err := mailboxSubmissionScopeFromOwner(actorID, owner)
	if err != nil {
		return MailboxSubmissionView{}, err
	}

	idempotencyKey, err := mailboxSubmissionNormalizeIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return MailboxSubmissionView{}, err
	}
	idempotencyKeyHash := sha256.Sum256([]byte(idempotencyKey))
	requestHash, ok := mailboxSubmissionCanonicalRequestHash(scope.accountID, draftID, input.ExpectedRevision)
	if !ok {
		return MailboxSubmissionView{}, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
	}

	reservation, err := service.repository.ReserveMailboxSubmission(ctx, MailboxSubmissionCreateRequest{
		AccountID:          scope.accountID,
		DraftID:            draftID,
		ExpectedRevision:   input.ExpectedRevision,
		IdempotencyKeyHash: idempotencyKeyHash,
		RequestHash:        requestHash,
	})
	if err != nil {
		return MailboxSubmissionView{}, mailboxSubmissionRepositoryError(err)
	}
	if !mailboxSubmissionReservationValid(reservation, scope.accountID, draftID) {
		return MailboxSubmissionView{}, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
	}
	return mailboxSubmissionViewFromRecord(reservation.Submission)
}

type mailboxSubmissionScope struct {
	accountID uuid.UUID
}

func mailboxSubmissionScopeFromOwner(actorID uuid.UUID, owner MailboxOwner) (mailboxSubmissionScope, error) {
	if actorID == uuid.Nil || owner.OrganizationID() == uuid.Nil {
		return mailboxSubmissionScope{}, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
	}
	accountID, hasAccount := owner.AccountID()
	projectionState, hasProjectionState := owner.ProjectionState()
	switch owner.State() {
	case MailboxOwnerNotConfigured:
		if hasAccount || accountID != uuid.Nil || owner.hasEngineAccountID || hasProjectionState {
			return mailboxSubmissionScope{}, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
		}
		return mailboxSubmissionScope{}, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorNotFound)
	case MailboxOwnerDeleting:
		if !hasAccount || accountID == uuid.Nil || owner.hasEngineAccountID || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return mailboxSubmissionScope{}, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
		}
		return mailboxSubmissionScope{}, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorDeleting)
	case MailboxOwnerProvisioning, MailboxOwnerSuspended, MailboxOwnerError:
		if !hasAccount || accountID == uuid.Nil || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return mailboxSubmissionScope{}, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
		}
		return mailboxSubmissionScope{}, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorUnavailable)
	case MailboxOwnerActive:
		if !hasAccount || accountID == uuid.Nil || !owner.hasEngineAccountID || !syncValidOpaque(owner.engineAccountID, MaxEngineOpaqueIDBytes) || !hasProjectionState || projectionState != MailboxProjectionReady {
			return mailboxSubmissionScope{}, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
		}
		return mailboxSubmissionScope{accountID: accountID}, nil
	default:
		return mailboxSubmissionScope{}, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
	}
}

func mailboxSubmissionNormalizeIdempotencyKey(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
	}
	normalized := norm.NFC.String(value)
	if len(normalized) < 1 || len(normalized) > mailboxSubmissionMaxIdempotencyKeyBytes {
		return "", mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
	}
	for _, character := range normalized {
		if unicode.IsControl(character) {
			return "", mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
		}
	}
	return strings.Clone(normalized), nil
}

func mailboxSubmissionCanonicalRequestHash(accountID, draftID uuid.UUID, expectedRevision int64) ([sha256.Size]byte, bool) {
	if accountID == uuid.Nil || draftID == uuid.Nil || expectedRevision < 1 {
		return [sha256.Size]byte{}, false
	}
	const domain = "hub.mail.submission.request.v1"
	framed := make([]byte, 0, 4+len(domain)+16+16+8)
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(domain)))
	framed = append(framed, length[:]...)
	framed = append(framed, domain...)
	framed = append(framed, accountID[:]...)
	framed = append(framed, draftID[:]...)
	var revision [8]byte
	binary.BigEndian.PutUint64(revision[:], uint64(expectedRevision))
	framed = append(framed, revision[:]...)
	return sha256.Sum256(framed), true
}

func mailboxSubmissionServiceValid(service *MailboxSubmissionService) bool {
	return service != nil && !mailboxSubmissionDependencyNil(service.owners) && !mailboxSubmissionDependencyNil(service.repository)
}

func mailboxSubmissionDependencyNil(dependency any) bool {
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

func mailboxSubmissionReservationValid(reservation MailboxSubmissionReservation, accountID, draftID uuid.UUID) bool {
	if accountID == uuid.Nil || draftID == uuid.Nil || !validMailboxSubmissionStatus(reservation.Submission.Status) ||
		reservation.Submission.ID == uuid.Nil || reservation.Submission.AccountID != accountID ||
		!mailboxEntValidTime(reservation.Submission.CreatedAt) || !mailboxEntValidTime(reservation.Submission.UpdatedAt) ||
		reservation.Submission.UpdatedAt.Before(reservation.Submission.CreatedAt) ||
		!mailboxSubmissionUUIDPointerValid(reservation.Submission.DraftID) ||
		!mailboxSubmissionUUIDPointerValid(reservation.Submission.ResultMailMessageID) ||
		!mailboxSubmissionErrorCodePointerValid(reservation.Submission.LastErrorCode) {
		return false
	}
	switch reservation.Kind {
	case MailboxSubmissionReservationQueued:
		return reservation.Submission.Status == MailboxSubmissionStatusQueued &&
			reservation.Submission.DraftID != nil && *reservation.Submission.DraftID == draftID &&
			reservation.Submission.ResultMailMessageID == nil && reservation.Submission.LastErrorCode == nil
	case MailboxSubmissionReservationReplay:
		return mailboxSubmissionReplayRecordValid(reservation.Submission, accountID, draftID)
	default:
		return false
	}
}

func mailboxSubmissionReplayRecordValid(record MailboxSubmissionRecord, accountID, draftID uuid.UUID) bool {
	if accountID == uuid.Nil || draftID == uuid.Nil || record.AccountID != accountID ||
		record.DraftID == nil || *record.DraftID != draftID || record.ID == uuid.Nil ||
		!mailboxEntValidTime(record.CreatedAt) || !mailboxEntValidTime(record.UpdatedAt) ||
		record.UpdatedAt.Before(record.CreatedAt) {
		return false
	}
	switch record.Status {
	case MailboxSubmissionStatusQueued, MailboxSubmissionStatusSubmitting:
		return record.ResultMailMessageID == nil && record.LastErrorCode == nil
	case MailboxSubmissionStatusSent:
		return record.ResultMailMessageID != nil && *record.ResultMailMessageID != uuid.Nil && record.LastErrorCode == nil
	case MailboxSubmissionStatusUnknown:
		if record.ResultMailMessageID != nil || record.LastErrorCode == nil {
			return false
		}
		switch *record.LastErrorCode {
		case MailboxSubmissionErrorCodeOutcomeUnknown,
			MailboxSubmissionErrorCodeReconciliationRequired,
			MailboxSubmissionErrorCodeMailboxUnavailable:
			return true
		default:
			return false
		}
	case MailboxSubmissionStatusFailed:
		if record.ResultMailMessageID != nil || record.LastErrorCode == nil {
			return false
		}
		switch *record.LastErrorCode {
		case MailboxSubmissionErrorCodeRejected,
			MailboxSubmissionErrorCodePolicy,
			MailboxSubmissionErrorCodeAttachmentNotClean,
			MailboxSubmissionErrorCodeIdentityUnavailable,
			MailboxSubmissionErrorCodeRecipientLimit,
			MailboxSubmissionErrorCodeSizeLimit,
			MailboxSubmissionErrorCodeMailboxUnavailable:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func mailboxSubmissionUUIDPointerValid(value *uuid.UUID) bool {
	return value == nil || *value != uuid.Nil
}

func mailboxSubmissionErrorCodePointerValid(value *MailboxSubmissionErrorCode) bool {
	return value == nil || validMailboxSubmissionErrorCode(*value)
}

func mailboxSubmissionViewFromRecord(record MailboxSubmissionRecord) (MailboxSubmissionView, error) {
	if record.ID == uuid.Nil || !validMailboxSubmissionStatus(record.Status) || !mailboxEntValidTime(record.CreatedAt) ||
		!mailboxEntValidTime(record.UpdatedAt) || record.UpdatedAt.Before(record.CreatedAt) ||
		!mailboxSubmissionUUIDPointerValid(record.ResultMailMessageID) ||
		!mailboxSubmissionErrorCodePointerValid(record.LastErrorCode) {
		return MailboxSubmissionView{}, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
	}
	return MailboxSubmissionView{
		ID:        record.ID,
		Status:    record.Status,
		MessageID: mailboxSubmissionUUIDPointerCopy(record.ResultMailMessageID),
		ErrorCode: mailboxSubmissionErrorCodePointerCopy(record.LastErrorCode),
		CreatedAt: record.CreatedAt.UTC(),
		UpdatedAt: record.UpdatedAt.UTC(),
	}, nil
}

// MailboxSubmissionServiceErrorKind is the safe caller-visible failure family
// for a durable send intent.
type MailboxSubmissionServiceErrorKind string

const (
	MailboxSubmissionServiceErrorInvalid              MailboxSubmissionServiceErrorKind = "invalid"
	MailboxSubmissionServiceErrorNotFound             MailboxSubmissionServiceErrorKind = "not_found"
	MailboxSubmissionServiceErrorDeleting             MailboxSubmissionServiceErrorKind = "deleting"
	MailboxSubmissionServiceErrorUnavailable          MailboxSubmissionServiceErrorKind = "unavailable"
	MailboxSubmissionServiceErrorVersionConflict      MailboxSubmissionServiceErrorKind = "version_conflict"
	MailboxSubmissionServiceErrorSubmissionInProgress MailboxSubmissionServiceErrorKind = "submission_in_progress"
	MailboxSubmissionServiceErrorDraftNotSendable     MailboxSubmissionServiceErrorKind = "draft_not_sendable"
	MailboxSubmissionServiceErrorIdempotencyConflict  MailboxSubmissionServiceErrorKind = "idempotency_conflict"
)

const mailboxSubmissionServiceErrorText = "mailbox submission service error"

// MailboxSubmissionServiceError contains a closed safe error kind and no
// provider, draft-content or storage data.
type MailboxSubmissionServiceError struct {
	kind MailboxSubmissionServiceErrorKind
}

func (err *MailboxSubmissionServiceError) Error() string {
	return mailboxSubmissionServiceErrorText
}

// Kind returns the validated service failure kind.
func (err *MailboxSubmissionServiceError) Kind() MailboxSubmissionServiceErrorKind {
	if err == nil || !validMailboxSubmissionServiceErrorKind(err.kind) {
		return MailboxSubmissionServiceErrorInvalid
	}
	return err.kind
}

// MailboxSubmissionServiceErrorKindOf safely extracts a submission service
// failure kind for a future HTTP adapter.
func MailboxSubmissionServiceErrorKindOf(err error) (MailboxSubmissionServiceErrorKind, bool) {
	var target *MailboxSubmissionServiceError
	if !errors.As(err, &target) || target == nil || !validMailboxSubmissionServiceErrorKind(target.kind) || target.kind == MailboxSubmissionServiceErrorInvalid {
		return MailboxSubmissionServiceErrorInvalid, false
	}
	return target.kind, true
}

func mailboxSubmissionServiceError(kind MailboxSubmissionServiceErrorKind) error {
	if !validMailboxSubmissionServiceErrorKind(kind) {
		kind = MailboxSubmissionServiceErrorInvalid
	}
	return &MailboxSubmissionServiceError{kind: kind}
}

func validMailboxSubmissionServiceErrorKind(kind MailboxSubmissionServiceErrorKind) bool {
	switch kind {
	case MailboxSubmissionServiceErrorInvalid,
		MailboxSubmissionServiceErrorNotFound,
		MailboxSubmissionServiceErrorDeleting,
		MailboxSubmissionServiceErrorUnavailable,
		MailboxSubmissionServiceErrorVersionConflict,
		MailboxSubmissionServiceErrorSubmissionInProgress,
		MailboxSubmissionServiceErrorDraftNotSendable,
		MailboxSubmissionServiceErrorIdempotencyConflict:
		return true
	default:
		return false
	}
}

func mailboxSubmissionOwnerError(err error) error {
	var ownerError *MailboxServiceError
	if !errors.As(err, &ownerError) || ownerError == nil {
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
	}
	switch ownerError.Kind {
	case MailboxServiceErrorNotFound:
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorNotFound)
	case MailboxServiceErrorDeleting:
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorDeleting)
	case MailboxServiceErrorUnavailable:
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorUnavailable)
	default:
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
	}
}

func mailboxSubmissionRepositoryError(err error) error {
	var repositoryError *MailboxSubmissionRepositoryError
	if !errors.As(err, &repositoryError) || repositoryError == nil {
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorUnavailable)
	}
	switch repositoryError.Kind() {
	case MailboxSubmissionRepositoryErrorInvalid:
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
	case MailboxSubmissionRepositoryErrorNotFound:
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorNotFound)
	case MailboxSubmissionRepositoryErrorUnavailable:
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorUnavailable)
	case MailboxSubmissionRepositoryErrorAccountDeleting:
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorDeleting)
	case MailboxSubmissionRepositoryErrorVersionConflict:
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorVersionConflict)
	case MailboxSubmissionRepositoryErrorSubmissionInProgress:
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorSubmissionInProgress)
	case MailboxSubmissionRepositoryErrorDraftNotSendable:
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorDraftNotSendable)
	case MailboxSubmissionRepositoryErrorIdempotencyConflict:
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorIdempotencyConflict)
	default:
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
	}
}
