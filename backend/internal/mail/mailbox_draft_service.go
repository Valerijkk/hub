package mail

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/google/uuid"
)

type MailboxDraftService struct {
	owners      mailboxOwnerResolver
	repository  MailboxDraftRepository
	engine      MailboxDraftEngine
	identities  MailboxDraftIdentityProofResolver
	attachments MailboxDraftAttachmentProofResolver
	materializer mailboxDraftAttachmentMaterializer
}

type MailboxDraftEngine interface {
	GetDraft(context.Context, EngineDraftReadRequest) (EngineMessage, error)
	SaveDraft(context.Context, EngineDraft) (EngineMessage, error)
	ReconcileDraft(context.Context, EngineDraftReconciliation) (EngineDraftResult, error)
	DeleteDraft(context.Context, EngineDraftDeletion) error
	ReconcileDraftDeletion(context.Context, EngineDraftDeletion) (EngineDraftDeletionResult, error)
}

var _ MailboxDraftEngine = (*StalwartEngine)(nil)

type MailboxDraftIdentityProofResolver interface {
	ResolveMailboxDraftIdentityProof(
		context.Context,
		mailboxDraftIdentityProofRequest,
	) (mailboxDraftIdentityProof, error)
}

type mailboxDraftIdentityProofRequest struct {
	AccountID  uuid.UUID
	IdentityID uuid.UUID
}

type mailboxDraftIdentityProof struct {
	accountID  uuid.UUID
	identityID uuid.UUID
	from       EngineAddress
}

type MailboxDraftAttachmentProofResolver interface {
	ResolveMailboxDraftAttachmentProof(
		context.Context,
		mailboxDraftAttachmentProofRequest,
	) (mailboxDraftAttachmentProof, error)
}

type mailboxDraftAttachmentProofRequest struct {
	AccountID     uuid.UUID
	DraftClientID uuid.UUID
	DraftID       uuid.UUID
	AttachmentIDs []uuid.UUID
}

type mailboxDraftAttachmentProof struct {
	accountID     uuid.UUID
	draftClientID uuid.UUID
	draftID       uuid.UUID
	attachments   []MailboxDraftRequestHashAttachment
}

type MailboxDraftAttachmentReference struct {
	AttachmentID uuid.UUID
}

type MailboxDraftWriteInput struct {
	DraftClientID    uuid.UUID
	FromIdentityID   uuid.UUID
	To               []EngineAddress
	CC               []EngineAddress
	BCC              []EngineAddress
	ReplyTo          []EngineAddress
	Subject          string
	InReplyTo        string
	References       []string
	Priority         EngineMessagePriority
	ReplyToMessageID *uuid.UUID
	ForwardMessageID *uuid.UUID
	Attachments      []MailboxDraftAttachmentReference
	Content          MailboxContentInput
}

type MailboxDraftCreateInput struct {
	MailboxDraftWriteInput
	IdempotencyKey string
}

type MailboxDraftSaveInput struct {
	MailboxDraftWriteInput
	ExpectedRevision int64
}

type MailboxDraftDeleteInput struct {
	ExpectedRevision int64
}

type MailboxDraftContentView struct {
	From       EngineAddress
	To         []EngineAddress
	CC         []EngineAddress
	BCC        []EngineAddress
	ReplyTo    []EngineAddress
	Subject    string
	InReplyTo  string
	References []string
	Priority   EngineMessagePriority
	Content    MailboxContentResult
}

type MailboxDraftView struct {
	ID            uuid.UUID
	DraftClientID uuid.UUID
	Revision      int64
	Status        MailboxDraftStatus
	Effective     MailboxDraftEffectiveState
	Content       MailboxDraftContentView
	UpdatedAt     time.Time
}

type MailboxDraftDeleteResult struct {
	ID        uuid.UUID
	Revision  int64
	Status    MailboxDraftStatus
	UpdatedAt time.Time
}

type MailboxDraftServiceAPI interface {
	CreateMailboxDraft(
		context.Context,
		uuid.UUID,
		MailboxDraftCreateInput,
	) (MailboxDraftView, error)

	GetMailboxDraft(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (MailboxDraftView, error)

	SaveMailboxDraft(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		MailboxDraftSaveInput,
	) (MailboxDraftView, error)

	DeleteMailboxDraft(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		MailboxDraftDeleteInput,
	) (MailboxDraftDeleteResult, error)
}

type MailboxDraftServiceErrorKind string

const (
	MailboxDraftServiceErrorInvalid                MailboxDraftServiceErrorKind = "invalid"
	MailboxDraftServiceErrorNotFound               MailboxDraftServiceErrorKind = "not_found"
	MailboxDraftServiceErrorDeleting               MailboxDraftServiceErrorKind = "deleting"
	MailboxDraftServiceErrorUnavailable            MailboxDraftServiceErrorKind = "unavailable"
	MailboxDraftServiceErrorPreconditionFailed     MailboxDraftServiceErrorKind = "precondition_failed"
	MailboxDraftServiceErrorConflict               MailboxDraftServiceErrorKind = "conflict"
	MailboxDraftServiceErrorIdempotencyConflict    MailboxDraftServiceErrorKind = "idempotency_conflict"
	MailboxDraftServiceErrorAttachmentConflict     MailboxDraftServiceErrorKind = "attachment_conflict"
	MailboxDraftServiceErrorAttachmentLimit        MailboxDraftServiceErrorKind = "attachment_limit"
	MailboxDraftServiceErrorPayloadTooLarge        MailboxDraftServiceErrorKind = "payload_too_large"
	MailboxDraftServiceErrorUnprocessable          MailboxDraftServiceErrorKind = "unprocessable"
	MailboxDraftServiceErrorOutcomeUnknown         MailboxDraftServiceErrorKind = "outcome_unknown"
	MailboxDraftServiceErrorReconciliationRequired MailboxDraftServiceErrorKind = "reconciliation_required"
)

const mailboxDraftServiceErrorText = "mailbox draft service error"

type MailboxDraftServiceError struct {
	kind MailboxDraftServiceErrorKind
}

func (err *MailboxDraftServiceError) Error() string {
	return mailboxDraftServiceErrorText
}

func (err *MailboxDraftServiceError) Kind() MailboxDraftServiceErrorKind {
	if err == nil || !validMailboxDraftServiceErrorKind(err.kind) {
		return MailboxDraftServiceErrorInvalid
	}
	return err.kind
}

func MailboxDraftServiceErrorKindOf(err error) (MailboxDraftServiceErrorKind, bool) {
	var target *MailboxDraftServiceError
	if !errors.As(err, &target) || target == nil || !validMailboxDraftServiceErrorKind(target.kind) || target.kind == MailboxDraftServiceErrorInvalid {
		return MailboxDraftServiceErrorInvalid, false
	}
	return target.kind, true
}

func NewMailboxDraftService(
	owners mailboxOwnerResolver,
	repository MailboxDraftRepository,
	engine MailboxDraftEngine,
	identities MailboxDraftIdentityProofResolver,
	attachments MailboxDraftAttachmentProofResolver,
	materializer mailboxDraftAttachmentMaterializer,
) (*MailboxDraftService, error) {
	if mailboxDraftServiceDependencyNil(owners) ||
		mailboxDraftServiceDependencyNil(repository) ||
		mailboxDraftServiceDependencyNil(engine) ||
		mailboxDraftServiceDependencyNil(identities) ||
		mailboxDraftServiceDependencyNil(attachments) ||
		mailboxDraftServiceDependencyNil(materializer) {
		return nil, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}

	return &MailboxDraftService{
		owners:      owners,
		repository:  repository,
		engine:      engine,
		identities:  identities,
		attachments: attachments,
		materializer: materializer,
	}, nil
}

func mailboxDraftServiceDependencyNil(dependency any) bool {
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

func mailboxDraftServiceError(kind MailboxDraftServiceErrorKind) error {
	if !validMailboxDraftServiceErrorKind(kind) {
		kind = MailboxDraftServiceErrorInvalid
	}
	return &MailboxDraftServiceError{kind: kind}
}

func validMailboxDraftServiceErrorKind(kind MailboxDraftServiceErrorKind) bool {
	switch kind {
	case MailboxDraftServiceErrorInvalid,
		MailboxDraftServiceErrorNotFound,
		MailboxDraftServiceErrorDeleting,
		MailboxDraftServiceErrorUnavailable,
		MailboxDraftServiceErrorPreconditionFailed,
		MailboxDraftServiceErrorConflict,
		MailboxDraftServiceErrorIdempotencyConflict,
		MailboxDraftServiceErrorAttachmentConflict,
		MailboxDraftServiceErrorAttachmentLimit,
		MailboxDraftServiceErrorPayloadTooLarge,
		MailboxDraftServiceErrorUnprocessable,
		MailboxDraftServiceErrorOutcomeUnknown,
		MailboxDraftServiceErrorReconciliationRequired:
		return true
	default:
		return false
	}
}
