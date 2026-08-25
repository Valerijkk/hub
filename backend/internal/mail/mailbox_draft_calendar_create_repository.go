package mail

import (
	"context"
	"crypto/sha256"

	"github.com/google/uuid"
)

// mailboxDraftCalendarCreateRepository owns only the durable, content-free
// Calendar MIME create transitions. It is deliberately private and separate
// from MailboxDraftRepository until the calendar create path has an accepted
// durable implementation; production calendar creation remains fail-closed.
// Adapters must use the exact fence and compare every mutable identity and
// binding with CAS. Database-clock-owned due checks stay in the adapter.
// No request, record, or result carries raw MIME or ICS content.
type mailboxDraftCalendarCreateRepository interface {
	ReserveMailboxDraftCalendarCreate(context.Context, mailboxDraftCalendarCreateReserveRequest) (MailboxDraftReservation, error)
	AdvanceMailboxDraftCalendarCreateToImport(context.Context, mailboxDraftCalendarCreateAdvanceToImportRequest) (MailboxDraftReservation, error)
	RestartMailboxDraftCalendarCreateUpload(context.Context, mailboxDraftCalendarCreateRestartUploadRequest) (MailboxDraftReservation, error)
	RestartMailboxDraftCalendarCreateUploadAfterBlobNotFound(context.Context, mailboxDraftCalendarCreateRestartAfterBlobNotFoundRequest) (MailboxDraftReservation, error)
	AbandonMailboxDraftCalendarCreateUpload(context.Context, mailboxDraftCalendarCreateAbandonRequest) (mailboxDraftCalendarCreateAbandonResult, error)
}

type mailboxDraftCalendarCreateReserveRequest struct {
	Create MailboxDraftCreateRequest
	State  MailboxDraftCreateMIMEState
}

type mailboxDraftCalendarCreateAdvanceToImportRequest struct {
	Fence                 MailboxEngineCommandFence
	DraftID               uuid.UUID
	MutationID            uuid.UUID
	ExpectedBindingDigest [sha256.Size]byte
	BlobID                string
	SizeBytes             int64
}

type mailboxDraftCalendarCreateRestartUploadRequest struct {
	Fence                 MailboxEngineCommandFence
	DraftID               uuid.UUID
	ExpectedMutationID    uuid.UUID
	ExpectedAttempt       int
	ExpectedBindingDigest [sha256.Size]byte
	NextMutationID        uuid.UUID
	NextBindingDigest     [sha256.Size]byte
	IdempotencyKeyHash    [sha256.Size]byte
	RequestHash           [sha256.Size]byte
}

type mailboxDraftCalendarCreateRestartAfterBlobNotFoundRequest struct {
	Fence                 MailboxEngineCommandFence
	DraftID               uuid.UUID
	ExpectedMutationID    uuid.UUID
	ExpectedAttempt       int
	ExpectedBindingDigest [sha256.Size]byte
	ExpectedBlobID        string
	ExpectedSizeBytes     int64
	NextMutationID        uuid.UUID
	NextBindingDigest     [sha256.Size]byte
	IdempotencyKeyHash    [sha256.Size]byte
	RequestHash           [sha256.Size]byte
}

type mailboxDraftCalendarCreateAbandonRequest struct {
	Fence                 MailboxEngineCommandFence
	DraftID               uuid.UUID
	ExpectedMutationID    uuid.UUID
	ExpectedAttempt       int
	ExpectedBindingDigest [sha256.Size]byte
	IdempotencyKeyHash    [sha256.Size]byte
	RequestHash           [sha256.Size]byte
}

type mailboxDraftCalendarCreateAbandonResult struct {
	Replayed bool
	Draft    MailboxDraftRecord
}

func validMailboxDraftCalendarCreateReserveRequest(request mailboxDraftCalendarCreateReserveRequest) bool {
	create := request.Create
	if !mailboxEntEngineCommandFenceValid(create.Fence) ||
		create.DraftID == uuid.Nil ||
		create.DraftClientID == uuid.Nil ||
		create.MutationID == uuid.Nil ||
		!mailboxDraftCalendarCreateDigestPresent(create.CreateIdempotencyKeyHash) ||
		!mailboxDraftCalendarCreateDigestPresent(create.CreateRequestHash) ||
		!mailboxDraftGetEffectiveStateValid(create.Confirmed) ||
		create.Confirmed.ReplyToMessageID != nil ||
		create.Confirmed.ForwardMessageID != nil ||
		len(create.Confirmed.AttachmentIDs) != 0 {
		return false
	}

	kind := MailboxDraftMutationCreate
	state := request.State
	return validMailboxDraftCreateMIMEState(&state, MailboxDraftStatusSaving, &kind) &&
		state.Kind == MailboxDraftCreateMIMEKindCalendar &&
		state.Phase == MailboxDraftCreateMIMEPhaseUpload &&
		state.Attempt == 1 &&
		state.BlobID == nil &&
		state.SizeBytes == nil &&
		state.RetryAt == nil &&
		state.ExhaustedAt == nil &&
		mailboxDraftCalendarCreateDigestPresent(state.BindingDigest)
}

func validMailboxDraftCalendarCreateAdvanceToImportRequest(request mailboxDraftCalendarCreateAdvanceToImportRequest) bool {
	return mailboxEntEngineCommandFenceValid(request.Fence) &&
		request.DraftID != uuid.Nil &&
		request.MutationID != uuid.Nil &&
		mailboxDraftCalendarCreateDigestPresent(request.ExpectedBindingDigest) &&
		validMailboxDraftAdvanceCreateToImportHandoff(request.BlobID, request.SizeBytes)
}

func validMailboxDraftCalendarCreateRestartUploadRequest(request mailboxDraftCalendarCreateRestartUploadRequest) bool {
	return mailboxEntEngineCommandFenceValid(request.Fence) &&
		request.DraftID != uuid.Nil &&
		request.ExpectedMutationID != uuid.Nil &&
		request.ExpectedAttempt >= 1 &&
		request.ExpectedAttempt <= 2 &&
		mailboxDraftCalendarCreateDigestPresent(request.ExpectedBindingDigest) &&
		request.NextMutationID != uuid.Nil &&
		request.NextMutationID != request.ExpectedMutationID &&
		mailboxDraftCalendarCreateDigestPresent(request.NextBindingDigest) &&
		request.NextBindingDigest != request.ExpectedBindingDigest &&
		mailboxDraftCalendarCreateDigestPresent(request.IdempotencyKeyHash) &&
		mailboxDraftCalendarCreateDigestPresent(request.RequestHash)
}

func validMailboxDraftCalendarCreateRestartAfterBlobNotFoundRequest(request mailboxDraftCalendarCreateRestartAfterBlobNotFoundRequest) bool {
	return mailboxEntEngineCommandFenceValid(request.Fence) &&
		request.DraftID != uuid.Nil &&
		request.ExpectedMutationID != uuid.Nil &&
		request.ExpectedAttempt >= 1 &&
		request.ExpectedAttempt <= 2 &&
		mailboxDraftCalendarCreateDigestPresent(request.ExpectedBindingDigest) &&
		validMailboxDraftAdvanceCreateToImportHandoff(request.ExpectedBlobID, request.ExpectedSizeBytes) &&
		request.NextMutationID != uuid.Nil &&
		request.NextMutationID != request.ExpectedMutationID &&
		mailboxDraftCalendarCreateDigestPresent(request.NextBindingDigest) &&
		request.NextBindingDigest != request.ExpectedBindingDigest &&
		mailboxDraftCalendarCreateDigestPresent(request.IdempotencyKeyHash) &&
		mailboxDraftCalendarCreateDigestPresent(request.RequestHash)
}

func validMailboxDraftCalendarCreateAbandonRequest(request mailboxDraftCalendarCreateAbandonRequest) bool {
	return mailboxEntEngineCommandFenceValid(request.Fence) &&
		request.DraftID != uuid.Nil &&
		request.ExpectedMutationID != uuid.Nil &&
		request.ExpectedAttempt >= 1 &&
		request.ExpectedAttempt <= 3 &&
		mailboxDraftCalendarCreateDigestPresent(request.ExpectedBindingDigest) &&
		mailboxDraftCalendarCreateDigestPresent(request.IdempotencyKeyHash) &&
		mailboxDraftCalendarCreateDigestPresent(request.RequestHash)
}

func mailboxDraftCalendarCreateDigestPresent(value [sha256.Size]byte) bool {
	return value != ([sha256.Size]byte{})
}
