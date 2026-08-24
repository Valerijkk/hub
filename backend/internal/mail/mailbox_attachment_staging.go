package mail

import (
	"context"
	"crypto/sha256"
	"errors"
	"reflect"
	"time"

	"github.com/google/uuid"
)

type MailboxAttachmentStagingConfig struct {
	Bucket        string
	PresignOrigin string
}

type MailboxAttachmentPresignInput struct {
	IdempotencyKey uuid.UUID
	DraftClientID  uuid.UUID
	Filename       string
	MediaType      string
	ContentID      *string
	Disposition    EngineAttachmentDisposition
	SizeBytes      int64
}

type MailboxAttachmentConfirmInput struct {
	SHA256    [sha256.Size]byte
	SizeBytes int64
}

type MailboxAttachmentView struct {
	ID          uuid.UUID
	Filename    string
	MediaType   string
	SizeBytes   int64
	Disposition EngineAttachmentDisposition
	ContentID   *string
	ScanStatus  AttachmentScanStatus
	SHA256      *[sha256.Size]byte
}

type MailboxAttachmentPresignResult struct {
	AttachmentID    uuid.UUID
	UploadURL       string
	ExpiresAt       time.Time
	RequiredHeaders map[string]string
}

type MailboxAttachmentDownloadResult struct {
	URL       string
	ExpiresAt time.Time
}

type MailboxAttachmentPresignRepositoryRequest struct {
	ActorID            uuid.UUID
	OrganizationID     uuid.UUID
	AccountID          uuid.UUID
	IdempotencyKeyHash [sha256.Size]byte
	RequestHash        [sha256.Size]byte
	DraftClientID      uuid.UUID
	AttachmentID       uuid.UUID
	ObjectID           uuid.UUID
	Bucket             string
	ObjectKey          string
	Filename           string
	MediaType          string
	ContentID          *string
	Disposition        EngineAttachmentDisposition
	SizeBytes          int64
	FingerprintVersion int32
}

type MailboxAttachmentPresignRepositoryResult struct {
	AttachmentID    uuid.UUID
	Object          MailboxObjectLocator
	IssuedAt        time.Time
	ObjectExpiresAt time.Time
}

type MailboxAttachmentConfirmPhase string

const (
	MailboxAttachmentConfirmPhasePrepare  MailboxAttachmentConfirmPhase = "prepare"
	MailboxAttachmentConfirmPhaseComplete MailboxAttachmentConfirmPhase = "complete"
)

type MailboxAttachmentConfirmResultKind string

const (
	MailboxAttachmentConfirmResultVerify    MailboxAttachmentConfirmResultKind = "verify"
	MailboxAttachmentConfirmResultConfirmed MailboxAttachmentConfirmResultKind = "confirmed"
)

type mailboxAttachmentConfirmFence struct {
	OrganizationID     uuid.UUID
	AccountID          uuid.UUID
	UploaderID         uuid.UUID
	DraftClientID      uuid.UUID
	AttachmentID       uuid.UUID
	ObjectID           uuid.UUID
	Bucket             string
	ObjectKey          string
	ExpectedSizeBytes  int64
	ObjectExpiresAt    time.Time
}

type MailboxAttachmentConfirmRepositoryRequest struct {
	Phase           MailboxAttachmentConfirmPhase
	ActorID         uuid.UUID
	AttachmentID    uuid.UUID
	ClaimedSHA256   [sha256.Size]byte
	ClaimedSizeBytes int64
	ActualSHA256    [sha256.Size]byte
	ActualSizeBytes int64
	Fence           *mailboxAttachmentConfirmFence
}

type MailboxAttachmentConfirmRepositoryResult struct {
	Kind              MailboxAttachmentConfirmResultKind
	Object            MailboxObjectLocator
	ExpectedSizeBytes int64
	ObjectExpiresAt   time.Time
	Fence             *mailboxAttachmentConfirmFence
	Attachment        *MailboxAttachmentView
}

type MailboxAttachmentDownloadRepositoryRequest struct {
	ActorID      uuid.UUID
	AttachmentID uuid.UUID
}

type MailboxAttachmentDownloadRepositoryResult struct {
	Object   MailboxObjectLocator
	IssuedAt time.Time
}

type MailboxAttachmentPresignRepository interface {
	ReserveMailboxAttachmentPresign(context.Context, MailboxAttachmentPresignRepositoryRequest) (MailboxAttachmentPresignRepositoryResult, error)
}

type MailboxAttachmentConfirmRepository interface {
	ConfirmMailboxAttachment(context.Context, MailboxAttachmentConfirmRepositoryRequest) (MailboxAttachmentConfirmRepositoryResult, error)
}

type MailboxAttachmentDownloadRepository interface {
	ResolveMailboxAttachmentDownload(context.Context, MailboxAttachmentDownloadRepositoryRequest) (MailboxAttachmentDownloadRepositoryResult, error)
}

type MailboxAttachmentStagingServiceAPI interface {
	PresignMailboxAttachment(context.Context, uuid.UUID, MailboxAttachmentPresignInput) (MailboxAttachmentPresignResult, error)
	ConfirmMailboxAttachment(context.Context, uuid.UUID, uuid.UUID, MailboxAttachmentConfirmInput) (MailboxAttachmentView, error)
	PresignMailboxAttachmentDownload(context.Context, uuid.UUID, uuid.UUID) (MailboxAttachmentDownloadResult, error)
}

type MailboxAttachmentStagingService struct {
	owners      mailboxOwnerResolver
	presigns    MailboxAttachmentPresignRepository
	confirms    MailboxAttachmentConfirmRepository
	downloads   MailboxAttachmentDownloadRepository
	objectStore MailboxObjectStore
	config      MailboxAttachmentStagingConfig
}

func NewMailboxAttachmentStagingService(
	owners mailboxOwnerResolver,
	presigns MailboxAttachmentPresignRepository,
	confirms MailboxAttachmentConfirmRepository,
	downloads MailboxAttachmentDownloadRepository,
	objectStore MailboxObjectStore,
	config MailboxAttachmentStagingConfig,
) (*MailboxAttachmentStagingService, error) {
	if mailboxAttachmentStagingDependencyNil(owners) ||
		mailboxAttachmentStagingDependencyNil(presigns) ||
		mailboxAttachmentStagingDependencyNil(confirms) ||
		mailboxAttachmentStagingDependencyNil(downloads) ||
		mailboxAttachmentStagingDependencyNil(objectStore) ||
		config.Bucket == "" || config.PresignOrigin == "" {
		return nil, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}

	return &MailboxAttachmentStagingService{
		owners:      owners,
		presigns:    presigns,
		confirms:    confirms,
		downloads:   downloads,
		objectStore: objectStore,
		config:      config,
	}, nil
}

type MailboxAttachmentStagingErrorKind string

const (
	MailboxAttachmentStagingErrorInvalid                MailboxAttachmentStagingErrorKind = "invalid"
	MailboxAttachmentStagingErrorNotFound               MailboxAttachmentStagingErrorKind = "not_found"
	MailboxAttachmentStagingErrorDeleting               MailboxAttachmentStagingErrorKind = "deleting"
	MailboxAttachmentStagingErrorUnavailable            MailboxAttachmentStagingErrorKind = "unavailable"
	MailboxAttachmentStagingErrorLimit                  MailboxAttachmentStagingErrorKind = "limit"
	MailboxAttachmentStagingErrorIdempotencyReused      MailboxAttachmentStagingErrorKind = "idempotency_reused"
	MailboxAttachmentStagingErrorIdempotencyUnavailable MailboxAttachmentStagingErrorKind = "idempotency_unavailable"
	MailboxAttachmentStagingErrorIdempotencyExpired     MailboxAttachmentStagingErrorKind = "idempotency_expired"
	MailboxAttachmentStagingErrorExpired                MailboxAttachmentStagingErrorKind = "expired"
	MailboxAttachmentStagingErrorConfirmMismatch        MailboxAttachmentStagingErrorKind = "confirm_mismatch"
	MailboxAttachmentStagingErrorStateConflict          MailboxAttachmentStagingErrorKind = "state_conflict"
	MailboxAttachmentStagingErrorNotClean               MailboxAttachmentStagingErrorKind = "not_clean"
)

const mailboxAttachmentStagingErrorText = "mailbox attachment staging error"

type MailboxAttachmentStagingError struct {
	kind MailboxAttachmentStagingErrorKind
}

func NewMailboxAttachmentStagingError(kind MailboxAttachmentStagingErrorKind) *MailboxAttachmentStagingError {
	if !validMailboxAttachmentStagingErrorKind(kind) {
		kind = MailboxAttachmentStagingErrorInvalid
	}
	return &MailboxAttachmentStagingError{kind: kind}
}

func (*MailboxAttachmentStagingError) Error() string {
	return mailboxAttachmentStagingErrorText
}

func (err *MailboxAttachmentStagingError) Kind() MailboxAttachmentStagingErrorKind {
	if err == nil || !validMailboxAttachmentStagingErrorKind(err.kind) {
		return MailboxAttachmentStagingErrorInvalid
	}
	return err.kind
}

func MailboxAttachmentStagingErrorKindOf(err error) (MailboxAttachmentStagingErrorKind, bool) {
	var target *MailboxAttachmentStagingError
	if !errors.As(err, &target) || target == nil || !validMailboxAttachmentStagingErrorKind(target.kind) || target.kind == MailboxAttachmentStagingErrorInvalid {
		return MailboxAttachmentStagingErrorInvalid, false
	}
	return target.kind, true
}

func mailboxAttachmentStagingDependencyNil(dependency any) bool {
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

func mailboxAttachmentStagingError(kind MailboxAttachmentStagingErrorKind) error {
	return NewMailboxAttachmentStagingError(kind)
}

func validMailboxAttachmentStagingErrorKind(kind MailboxAttachmentStagingErrorKind) bool {
	switch kind {
	case MailboxAttachmentStagingErrorInvalid,
		MailboxAttachmentStagingErrorNotFound,
		MailboxAttachmentStagingErrorDeleting,
		MailboxAttachmentStagingErrorUnavailable,
		MailboxAttachmentStagingErrorLimit,
		MailboxAttachmentStagingErrorIdempotencyReused,
		MailboxAttachmentStagingErrorIdempotencyUnavailable,
		MailboxAttachmentStagingErrorIdempotencyExpired,
		MailboxAttachmentStagingErrorExpired,
		MailboxAttachmentStagingErrorConfirmMismatch,
		MailboxAttachmentStagingErrorStateConflict,
		MailboxAttachmentStagingErrorNotClean:
		return true
	default:
		return false
	}
}
