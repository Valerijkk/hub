package mail

import (
	"crypto/sha256"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MailboxDraftCreateMIMEKind identifies the MIME payload bound to a draft-create mutation.
type MailboxDraftCreateMIMEKind string

const (
	MailboxDraftCreateMIMEKindAttachments MailboxDraftCreateMIMEKind = "attachments"
	MailboxDraftCreateMIMEKindCalendar    MailboxDraftCreateMIMEKind = "calendar"
)

// MailboxDraftCreateMIMEPhase identifies the persistence phase for a MIME payload.
type MailboxDraftCreateMIMEPhase string

const (
	MailboxDraftCreateMIMEPhaseUpload MailboxDraftCreateMIMEPhase = "upload"
	MailboxDraftCreateMIMEPhaseImport MailboxDraftCreateMIMEPhase = "import"
)

// MailboxDraftCreateMIMEState is the active generic MIME state of a draft-create mutation.
type MailboxDraftCreateMIMEState struct {
	Kind          MailboxDraftCreateMIMEKind
	Phase         MailboxDraftCreateMIMEPhase
	BlobID        *string
	SizeBytes     *int64
	BindingDigest [sha256.Size]byte
	Attempt       int
	RetryAt       *time.Time
	ExhaustedAt   *time.Time
}

const (
	mailboxDraftCreateCalendarHashFieldDomain uint16 = iota + 1
	mailboxDraftCreateCalendarHashFieldBaseRequestHash
	mailboxDraftCreateCalendarHashFieldBindingDigest
)

const (
	mailboxDraftCreateMIMEBindingFieldDomain uint16 = iota + 1
	mailboxDraftCreateMIMEBindingFieldOrganizationID
	mailboxDraftCreateMIMEBindingFieldAccountID
	mailboxDraftCreateMIMEBindingFieldDraftID
	mailboxDraftCreateMIMEBindingFieldMutationID
	mailboxDraftCreateMIMEBindingFieldKind
	mailboxDraftCreateMIMEBindingFieldRequestHash
	mailboxDraftCreateMIMEBindingFieldCalendarMarker
	mailboxDraftCreateMIMEBindingFieldCalendarDigest
)

func validMailboxDraftCreateMIMEKind(kind MailboxDraftCreateMIMEKind) bool {
	switch kind {
	case MailboxDraftCreateMIMEKindAttachments, MailboxDraftCreateMIMEKindCalendar:
		return true
	default:
		return false
	}
}

func validMailboxDraftCreateMIMEPhase(phase MailboxDraftCreateMIMEPhase) bool {
	switch phase {
	case MailboxDraftCreateMIMEPhaseUpload, MailboxDraftCreateMIMEPhaseImport:
		return true
	default:
		return false
	}
}

func validMailboxDraftCreateMIMEState(
	state *MailboxDraftCreateMIMEState,
	status MailboxDraftStatus,
	pendingKind *MailboxDraftMutationKind,
) bool {
	if state == nil {
		return true
	}
	if pendingKind == nil || *pendingKind != MailboxDraftMutationCreate {
		return false
	}
	if status != MailboxDraftStatusSaving && status != MailboxDraftStatusUnknown {
		return false
	}
	if !validMailboxDraftCreateMIMEKind(state.Kind) ||
		!validMailboxDraftCreateMIMEPhase(state.Phase) ||
		state.BindingDigest == ([sha256.Size]byte{}) ||
		state.Attempt < 1 ||
		state.Attempt > 3 {
		return false
	}

	switch state.Phase {
	case MailboxDraftCreateMIMEPhaseUpload:
		if state.BlobID != nil || state.SizeBytes != nil {
			return false
		}
	case MailboxDraftCreateMIMEPhaseImport:
		if state.BlobID == nil ||
			!validOpaqueID(*state.BlobID) ||
			state.SizeBytes == nil ||
			*state.SizeBytes < 1 ||
			*state.SizeBytes > MaxEngineMessageBytes {
			return false
		}
	}

	if state.Phase == MailboxDraftCreateMIMEPhaseImport ||
		status == MailboxDraftStatusSaving {
		return state.RetryAt == nil && state.ExhaustedAt == nil
	}

	if state.Attempt < 3 {
		return state.RetryAt != nil &&
			mailboxEntValidTime(*state.RetryAt) &&
			state.ExhaustedAt == nil
	}

	return state.RetryAt == nil &&
		state.ExhaustedAt != nil &&
		mailboxEntValidTime(*state.ExhaustedAt)
}

func mailboxDraftCreateMIMEStateCopy(
	state *MailboxDraftCreateMIMEState,
) *MailboxDraftCreateMIMEState {
	if state == nil {
		return nil
	}

	copyState := *state
	if state.BlobID != nil {
		blobID := strings.Clone(*state.BlobID)
		copyState.BlobID = &blobID
	}
	if state.SizeBytes != nil {
		sizeBytes := *state.SizeBytes
		copyState.SizeBytes = &sizeBytes
	}
	if state.RetryAt != nil {
		retryAt := *state.RetryAt
		copyState.RetryAt = &retryAt
	}
	if state.ExhaustedAt != nil {
		exhaustedAt := *state.ExhaustedAt
		copyState.ExhaustedAt = &exhaustedAt
	}

	return &copyState
}

func validMailboxDraftCreateMIMEStateBridge(
	state *MailboxDraftCreateMIMEState,
	status MailboxDraftStatus,
	pendingKind *MailboxDraftMutationKind,
	attachmentPhase *MailboxDraftCreateAttachmentPhase,
	attachmentBlobID *string,
	attachmentSizeBytes *int64,
	attachmentAttempt int,
	attachmentRetryAt *time.Time,
	attachmentExhaustedAt *time.Time,
) bool {
	if !validMailboxDraftCreateMIMEState(state, status, pendingKind) {
		return false
	}
	if state == nil {
		return true
	}

	switch state.Kind {
	case MailboxDraftCreateMIMEKindCalendar:
		return attachmentPhase == nil &&
			attachmentBlobID == nil &&
			attachmentSizeBytes == nil &&
			attachmentAttempt == 0 &&
			attachmentRetryAt == nil &&
			attachmentExhaustedAt == nil
	case MailboxDraftCreateMIMEKindAttachments:
		if attachmentPhase == nil ||
			state.Phase != MailboxDraftCreateMIMEPhase(*attachmentPhase) {
			return false
		}
		return mailboxDraftCreateMIMEStringEqual(state.BlobID, attachmentBlobID) &&
			mailboxDraftCreateMIMEInt64Equal(state.SizeBytes, attachmentSizeBytes) &&
			state.Attempt == attachmentAttempt &&
			mailboxDraftCreateMIMETimeEqual(state.RetryAt, attachmentRetryAt) &&
			mailboxDraftCreateMIMETimeEqual(state.ExhaustedAt, attachmentExhaustedAt)
	default:
		return false
	}
}

func mailboxDraftCreateMIMEStringEqual(left, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func mailboxDraftCreateMIMEInt64Equal(left, right *int64) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func mailboxDraftCreateMIMETimeEqual(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Equal(*right)
}

func mailboxDraftCreateRequestHashWithCalendar(
	base [sha256.Size]byte,
	calendar *calendarMIMEPart,
) ([sha256.Size]byte, bool) {
	if base == ([sha256.Size]byte{}) {
		return [sha256.Size]byte{}, false
	}
	if calendar == nil {
		return base, true
	}
	if !calendar.valid() {
		return [sha256.Size]byte{}, false
	}

	calendarDigest := calendar.bindingDigest()
	return sha256.Sum256(canonicalRecord(
		canonicalTLV(
			mailboxDraftCreateCalendarHashFieldDomain,
			[]byte("hub-mail-draft-calendar-request-hash-v1"),
		),
		canonicalTLV(
			mailboxDraftCreateCalendarHashFieldBaseRequestHash,
			base[:],
		),
		canonicalTLV(
			mailboxDraftCreateCalendarHashFieldBindingDigest,
			calendarDigest[:],
		),
	)), true
}

func mailboxDraftCreateMIMEBindingDigest(
	organizationID uuid.UUID,
	accountID uuid.UUID,
	draftID uuid.UUID,
	mutationID uuid.UUID,
	kind MailboxDraftCreateMIMEKind,
	requestHash [sha256.Size]byte,
	calendar *calendarMIMEPart,
) ([sha256.Size]byte, bool) {
	if organizationID == uuid.Nil ||
		accountID == uuid.Nil ||
		draftID == uuid.Nil ||
		mutationID == uuid.Nil ||
		!validMailboxDraftCreateMIMEKind(kind) ||
		requestHash == ([sha256.Size]byte{}) {
		return [sha256.Size]byte{}, false
	}

	calendarMarker := []byte{0}
	calendarDigest := [sha256.Size]byte{}
	switch kind {
	case MailboxDraftCreateMIMEKindAttachments:
		if calendar != nil {
			return [sha256.Size]byte{}, false
		}
	case MailboxDraftCreateMIMEKindCalendar:
		if calendar == nil ||
			!calendar.valid() ||
			calendar.organization() != organizationID {
			return [sha256.Size]byte{}, false
		}
		calendarMarker = []byte{1}
		calendarDigest = calendar.bindingDigest()
	}

	return sha256.Sum256(canonicalRecord(
		canonicalTLV(
			mailboxDraftCreateMIMEBindingFieldDomain,
			[]byte("hub-mail-draft-create-mime-binding-v1"),
		),
		canonicalTLV(
			mailboxDraftCreateMIMEBindingFieldOrganizationID,
			canonicalUUID(organizationID),
		),
		canonicalTLV(
			mailboxDraftCreateMIMEBindingFieldAccountID,
			canonicalUUID(accountID),
		),
		canonicalTLV(
			mailboxDraftCreateMIMEBindingFieldDraftID,
			canonicalUUID(draftID),
		),
		canonicalTLV(
			mailboxDraftCreateMIMEBindingFieldMutationID,
			canonicalUUID(mutationID),
		),
		canonicalTLV(
			mailboxDraftCreateMIMEBindingFieldKind,
			[]byte(kind),
		),
		canonicalTLV(
			mailboxDraftCreateMIMEBindingFieldRequestHash,
			requestHash[:],
		),
		canonicalTLV(
			mailboxDraftCreateMIMEBindingFieldCalendarMarker,
			calendarMarker,
		),
		canonicalTLV(
			mailboxDraftCreateMIMEBindingFieldCalendarDigest,
			calendarDigest[:],
		),
	)), true
}
