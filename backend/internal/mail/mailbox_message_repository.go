package mail

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"reflect"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxMailboxMessageDetailResponseBytes int64 = MaxEngineMessageBytes
	mailboxMessageEngineReadMaxBytes     int64 = MaxEngineBodyPartBytes + MaxEngineProjectionItemBytes
)

type MailboxMessageReadErrorKind string

const (
	MailboxMessageReadErrorInvalid         MailboxMessageReadErrorKind = "invalid"
	MailboxMessageReadErrorNotFound        MailboxMessageReadErrorKind = "not_found"
	MailboxMessageReadErrorDeleting        MailboxMessageReadErrorKind = "deleting"
	MailboxMessageReadErrorUnavailable     MailboxMessageReadErrorKind = "unavailable"
	MailboxMessageReadErrorPayloadTooLarge MailboxMessageReadErrorKind = "payload_too_large"
	MailboxMessageReadErrorUnprocessable   MailboxMessageReadErrorKind = "unprocessable"
)

const mailboxMessageReadErrorText = "mail message read error"

// MailboxMessageReadError carries only a closed, content-free public failure kind.
type MailboxMessageReadError struct {
	kind MailboxMessageReadErrorKind
}

func (e *MailboxMessageReadError) Error() string {
	return mailboxMessageReadErrorText
}

func (e *MailboxMessageReadError) Kind() MailboxMessageReadErrorKind {
	if e == nil || !validMailboxMessageReadErrorKind(e.kind) {
		return MailboxMessageReadErrorInvalid
	}
	return e.kind
}

func mailboxMessageReadErrorKindOf(err error) (MailboxMessageReadErrorKind, bool) {
	var readError *MailboxMessageReadError
	if !errors.As(err, &readError) || readError == nil {
		return "", false
	}
	return readError.Kind(), true
}

func mailboxMessageReadError(kind MailboxMessageReadErrorKind) error {
	if !validMailboxMessageReadErrorKind(kind) {
		kind = MailboxMessageReadErrorInvalid
	}
	return &MailboxMessageReadError{kind: kind}
}

func validMailboxMessageReadErrorKind(kind MailboxMessageReadErrorKind) bool {
	switch kind {
	case MailboxMessageReadErrorInvalid,
		MailboxMessageReadErrorNotFound,
		MailboxMessageReadErrorDeleting,
		MailboxMessageReadErrorUnavailable,
		MailboxMessageReadErrorPayloadTooLarge,
		MailboxMessageReadErrorUnprocessable:
		return true
	default:
		return false
	}
}

// MailboxMessageReadScope is the complete authenticated local and engine account scope.
type MailboxMessageReadScope struct {
	ActorID         uuid.UUID
	OrganizationID  uuid.UUID
	AccountID       uuid.UUID
	EngineAccountID string
}

// MailboxMessageRepository resolves and rechecks one sealed local message snapshot.
type MailboxMessageRepository interface {
	ResolveMailboxMessage(context.Context, MailboxMessageReadScope, uuid.UUID) (MailboxMessageSnapshot, error)
	RecheckMailboxMessage(context.Context, MailboxMessageReadScope, MailboxMessageSnapshot) error
}

// MailboxMessageAttachment is the public, local attachment metadata projection.
type MailboxMessageAttachment struct {
	ID          uuid.UUID `json:"id"`
	Filename    string    `json:"filename"`
	MediaType   string    `json:"media_type"`
	SizeBytes   int64     `json:"size_bytes"`
	Disposition string    `json:"disposition"`
	ContentID   *string   `json:"content_id"`
	ScanStatus  string    `json:"scan_status"`
	SHA256      string    `json:"sha256"`

	engineBlobID string
	createdAt    time.Time
	updatedAt    time.Time
}

type mailboxMessageCalendarProjection struct{}

// MailboxMessageCalendarRSVPTarget is the closed public projection of the
// inbound terminal Calendar RSVP authority for one message detail.
type MailboxMessageCalendarRSVPTarget struct {
	EventID         uuid.UUID                 `json:"event_id"`
	Scope           string                    `json:"scope"`
	RecurrenceID    *string                   `json:"recurrence_id"`
	AttendeeVersion uint                      `json:"attendee_version"`
	RSVPStatus      MailboxCalendarRSVPStatus `json:"rsvp_status"`
}

// MailboxMessageDetail is the closed public MailMessageDetail response model.
type MailboxMessageDetail struct {
	ID                  uuid.UUID                              `json:"id"`
	ThreadID            uuid.UUID                              `json:"thread_id"`
	From                MailboxThreadDetailProjectionAddress   `json:"from"`
	To                  []MailboxThreadDetailProjectionAddress `json:"to"`
	CC                  []MailboxThreadDetailProjectionAddress `json:"cc"`
	Subject             string                                 `json:"subject"`
	Snippet             string                                 `json:"snippet"`
	SentAt              *time.Time                             `json:"sent_at"`
	ReceivedAt          *time.Time                             `json:"received_at"`
	Priority            string                                 `json:"priority"`
	SizeBytes           int64                                  `json:"size_bytes"`
	HasAttachments      bool                                   `json:"has_attachments"`
	IsDraft             bool                                   `json:"is_draft"`
	FolderIDs           []uuid.UUID                            `json:"folder_ids"`
	Seen                bool                                   `json:"seen"`
	Flagged             bool                                   `json:"flagged"`
	Answered            bool                                   `json:"answered"`
	Junk                bool                                   `json:"junk"`
	CalendarMethod      *MailCalendarMethod                    `json:"calendar_method"`
	CalendarEventID     *uuid.UUID                             `json:"calendar_event_id"`
	CalendarRSVPTarget  *MailboxMessageCalendarRSVPTarget      `json:"calendar_rsvp_target"`
	PlainBody           string                                 `json:"plain_body"`
	SanitizedHTML       *string                                `json:"sanitized_html"`
	RemoteImagesBlocked bool                                   `json:"remote_images_blocked"`
	Attachments         []MailboxMessageAttachment             `json:"attachments"`
	Calendar            *mailboxMessageCalendarProjection      `json:"calendar"`
}

// MailboxMessageSnapshot seals local public metadata together with private engine identity.
type MailboxMessageSnapshot struct {
	summary                 MailboxThreadDetailProjectionMessage
	attachments             []MailboxMessageAttachment
	engineAccountID         string
	engineMessageID         string
	engineVersion           string
	internetMessageID       string
	inReplyTo               string
	references              []string
	from                    EngineAddress
	to                      []EngineAddress
	cc                      []EngineAddress
	bcc                     []EngineAddress
	replyTo                 []EngineAddress
	projectionOrigin        EngineProjectionOrigin
	authenticationClass     MailAuthenticationClass
	calendarPartLocatorHash []byte
	engineFolderIDs         []string
	projectionState         MailboxProjectionState
	projectionAsOf          *time.Time
	syncEpoch               int64
	mutationEpoch           int64
	updatedAt               time.Time
}

func mailboxMessageReadScopeValid(scope MailboxMessageReadScope) bool {
	return scope.ActorID != uuid.Nil && scope.OrganizationID != uuid.Nil && scope.AccountID != uuid.Nil && syncValidOpaque(scope.EngineAccountID, MaxEngineOpaqueIDBytes)
}

func mailboxMessageSnapshotCopy(snapshot MailboxMessageSnapshot) MailboxMessageSnapshot {
	result := snapshot
	result.summary = mailboxThreadDetailProjectionMessageCopy(snapshot.summary)
	result.attachments = make([]MailboxMessageAttachment, len(snapshot.attachments))
	for index, attachment := range snapshot.attachments {
		result.attachments[index] = mailboxMessageAttachmentCopy(attachment)
	}
	result.references = append([]string(nil), snapshot.references...)
	result.to = append([]EngineAddress(nil), snapshot.to...)
	result.cc = append([]EngineAddress(nil), snapshot.cc...)
	result.bcc = append([]EngineAddress(nil), snapshot.bcc...)
	result.replyTo = append([]EngineAddress(nil), snapshot.replyTo...)
	result.calendarPartLocatorHash = mailboxMessageOptionalBytesCopy(snapshot.calendarPartLocatorHash)
	result.engineFolderIDs = append([]string(nil), snapshot.engineFolderIDs...)
	result.projectionAsOf = mailboxFolderProjectionTimePointer(snapshot.projectionAsOf)
	return result
}

func mailboxMessageOptionalBytesCopy(value []byte) []byte {
	if value == nil {
		return nil
	}
	return append([]byte{}, value...)
}

func mailboxMessageAttachmentCopy(attachment MailboxMessageAttachment) MailboxMessageAttachment {
	attachment.ContentID = mailboxFolderProjectionStringPointer(attachment.ContentID)
	attachment.createdAt = attachment.createdAt.UTC()
	attachment.updatedAt = attachment.updatedAt.UTC()
	return attachment
}

func mailboxMessageSnapshotEqual(left, right MailboxMessageSnapshot) bool {
	if !mailboxMessageSnapshotValid(left) || !mailboxMessageSnapshotValid(right) {
		return false
	}
	return reflect.DeepEqual(mailboxMessageSnapshotCopy(left), mailboxMessageSnapshotCopy(right))
}

func mailboxMessageSnapshotValid(snapshot MailboxMessageSnapshot) bool {
	if snapshot.summary.ID == uuid.Nil || !validMailboxThreadDetailProjectionMessage(snapshot.summary, snapshot.summary.ThreadID) ||
		!syncValidOpaque(snapshot.engineAccountID, MaxEngineOpaqueIDBytes) ||
		!syncValidOpaque(snapshot.engineMessageID, MaxEngineOpaqueIDBytes) ||
		!syncValidOpaque(snapshot.engineVersion, MaxEngineOpaqueIDBytes) ||
		!validSafeText(snapshot.internetMessageID, MaxEngineHeaderValueBytes, false) ||
		!validSafeText(snapshot.inReplyTo, MaxEngineHeaderValueBytes, false) ||
		!validStrings(snapshot.references, MaxEngineReferenceItems, MaxEngineHeaderValueBytes) ||
		!validAddress(snapshot.from, true) || !validAddresses(snapshot.to, MaxEngineAddressItems) ||
		!validAddresses(snapshot.cc, MaxEngineAddressItems) || !validAddresses(snapshot.bcc, MaxEngineAddressItems) ||
		!validAddresses(snapshot.replyTo, MaxEngineAddressItems) || !mailboxMessagePublicAddressMatches(snapshot.summary.From, snapshot.from) ||
		!mailboxMessagePublicAddressesMatch(snapshot.summary.To, snapshot.to) || !mailboxMessagePublicAddressesMatch(snapshot.summary.CC, snapshot.cc) ||
		len(snapshot.engineFolderIDs) < 1 || len(snapshot.engineFolderIDs) > MaxEngineFolderItems || !validOpaqueIDs(snapshot.engineFolderIDs) ||
		len(snapshot.attachments) > MaxEngineAttachmentItems || snapshot.summary.HasAttachments != (len(snapshot.attachments) != 0) ||
		!validMailboxProjectionState(snapshot.projectionState) || !mailboxEntValidTime(snapshot.updatedAt) ||
		snapshot.syncEpoch < 0 || snapshot.mutationEpoch < 0 {
		return false
	}
	if snapshot.projectionState == MailboxProjectionReady && snapshot.projectionAsOf == nil {
		return false
	}
	if snapshot.projectionAsOf != nil && !mailboxEntValidTime(*snapshot.projectionAsOf) {
		return false
	}
	switch snapshot.projectionOrigin {
	case EngineProjectionOriginInboundOrUnknown, EngineProjectionOriginConfirmedOwnerDraft, EngineProjectionOriginConfirmedOwnerSubmission:
	default:
		return false
	}
	switch snapshot.authenticationClass {
	case MailAuthenticationClassTrustedInternal, MailAuthenticationClassDMARCAligned, MailAuthenticationClassUntrusted:
	default:
		return false
	}
	if (snapshot.summary.CalendarMethod == nil) != (snapshot.calendarPartLocatorHash == nil) || snapshot.calendarPartLocatorHash != nil && len(snapshot.calendarPartLocatorHash) != 32 {
		return false
	}
	attachmentIDs := make(map[uuid.UUID]struct{}, len(snapshot.attachments))
	engineBlobIDs := make(map[string]struct{}, len(snapshot.attachments))
	for index, attachment := range snapshot.attachments {
		if !mailboxMessageAttachmentValid(attachment) || index > 0 && !syncLessUUID(snapshot.attachments[index-1].ID, attachment.ID) {
			return false
		}
		if _, duplicate := attachmentIDs[attachment.ID]; duplicate {
			return false
		}
		if _, duplicate := engineBlobIDs[attachment.engineBlobID]; duplicate {
			return false
		}
		attachmentIDs[attachment.ID] = struct{}{}
		engineBlobIDs[attachment.engineBlobID] = struct{}{}
	}
	return sort.StringsAreSorted(snapshot.engineFolderIDs)
}

func mailboxMessageAttachmentValid(attachment MailboxMessageAttachment) bool {
	if attachment.ID == uuid.Nil || !syncValidOpaque(attachment.engineBlobID, MaxEngineOpaqueIDBytes) ||
		!syncValidSafeText(attachment.Filename, MaxEngineMIMEFieldBytes, false) ||
		!syncValidSafeText(attachment.MediaType, MaxEngineMIMEFieldBytes, false) ||
		utf8.RuneCountInString(attachment.Filename) > 255 || utf8.RuneCountInString(attachment.MediaType) > 255 ||
		attachment.SizeBytes < 0 || attachment.SizeBytes > MaxEngineAttachmentBytes ||
		attachment.ScanStatus != "clean" || !mailboxEntValidTime(attachment.createdAt) || !mailboxEntValidTime(attachment.updatedAt) {
		return false
	}
	if attachment.Disposition != string(EngineAttachmentDispositionAttachment) && attachment.Disposition != string(EngineAttachmentDispositionInline) {
		return false
	}
	if attachment.ContentID != nil && (!syncValidSafeText(*attachment.ContentID, MaxEngineMIMEFieldBytes, false) || utf8.RuneCountInString(*attachment.ContentID) > 255) || attachment.Disposition == string(EngineAttachmentDispositionInline) && attachment.ContentID == nil {
		return false
	}
	decoded, err := hex.DecodeString(attachment.SHA256)
	return err == nil && len(decoded) == 32 && hex.EncodeToString(decoded) == attachment.SHA256
}

func mailboxMessagePublicAddressMatches(public MailboxThreadDetailProjectionAddress, private EngineAddress) bool {
	if !validMailboxThreadDetailProjectionAddress(public) || !validAddress(private, true) || public.Email != private.Email {
		return false
	}
	if private.Name == "" {
		return public.Name == nil
	}
	return public.Name != nil && *public.Name == private.Name
}

func mailboxMessagePublicAddressesMatch(public []MailboxThreadDetailProjectionAddress, private []EngineAddress) bool {
	if len(public) != len(private) {
		return false
	}
	for index := range public {
		if !mailboxMessagePublicAddressMatches(public[index], private[index]) {
			return false
		}
	}
	return true
}

func mailboxMessageProjectionMatches(snapshot MailboxMessageSnapshot, projection EngineMessageProjection) bool {
	if !mailboxMessageSnapshotValid(snapshot) || validateProjection(projection) != nil ||
		projection.AccountID != snapshot.engineAccountID || projection.MessageID != snapshot.engineMessageID || projection.Version != snapshot.engineVersion ||
		projection.InternetMessageID != snapshot.internetMessageID || projection.InReplyTo != snapshot.inReplyTo ||
		!mailboxMessageStringsEqual(projection.References, snapshot.references) || !mailboxMessageEngineAddressEqual(projection.From, snapshot.from) ||
		!mailboxMessageEngineAddressesEqual(projection.To, snapshot.to) || !mailboxMessageEngineAddressesEqual(projection.CC, snapshot.cc) ||
		!mailboxMessageEngineAddressesEqual(projection.BCC, snapshot.bcc) || !mailboxMessageEngineAddressesEqual(projection.ReplyTo, snapshot.replyTo) ||
		projection.Subject != snapshot.summary.Subject || projection.Snippet != snapshot.summary.Snippet ||
		!mailboxMessageTimePointerEqual(projection.SentAt, snapshot.summary.SentAt) || !mailboxMessageTimePointerEqual(projection.ReceivedAt, snapshot.summary.ReceivedAt) ||
		projection.SizeBytes != snapshot.summary.SizeBytes || string(projection.Priority) != snapshot.summary.Priority ||
		projection.Flags.Seen != snapshot.summary.Seen || projection.Flags.Flagged != snapshot.summary.Flagged ||
		projection.Flags.Answered != snapshot.summary.Answered || projection.Flags.Junk != snapshot.summary.Junk || projection.Flags.Draft != snapshot.summary.IsDraft ||
		projection.Origin != snapshot.projectionOrigin || projection.AuthenticationClass != snapshot.authenticationClass ||
		!mailboxMessageCalendarMethodEqual(projection.CalendarMethod, snapshot.summary.CalendarMethod) ||
		!mailboxMessageOptionalBytesEqual(projection.CalendarPartLocatorHash, snapshot.calendarPartLocatorHash) ||
		!mailboxMessageOpaqueSetEqual(projection.FolderIDs, snapshot.engineFolderIDs) ||
		!mailboxMessageProjectionAttachmentsMatch(projection.Attachments, snapshot.attachments) {
		return false
	}
	return true
}

func mailboxMessageProjectionAttachmentsMatch(provider []EngineProjectionAttachment, stored []MailboxMessageAttachment) bool {
	if len(provider) != len(stored) || len(provider) > MaxEngineAttachmentItems {
		return false
	}
	byBlobID := make(map[string]MailboxMessageAttachment, len(stored))
	for _, attachment := range stored {
		if !mailboxMessageAttachmentValid(attachment) {
			return false
		}
		byBlobID[attachment.engineBlobID] = attachment
	}
	for _, attachment := range provider {
		storedAttachment, found := byBlobID[attachment.BlobID]
		filename := attachment.Filename
		if filename == "" {
			filename = mailboxEntAttachmentDefaultFilename
		}
		mediaType := attachment.MediaType
		if mediaType == "" {
			mediaType = mailboxEntAttachmentDefaultMediaType
		}
		if !found || filename != storedAttachment.Filename || mediaType != storedAttachment.MediaType ||
			attachment.SizeBytes != storedAttachment.SizeBytes || string(attachment.Disposition) != storedAttachment.Disposition ||
			!mailboxMessageOptionalStringEqual(attachment.ContentID, storedAttachment.ContentID) ||
			len(attachment.SHA256) != 0 && hex.EncodeToString(attachment.SHA256) != storedAttachment.SHA256 {
			return false
		}
		delete(byBlobID, attachment.BlobID)
	}
	return len(byBlobID) == 0
}

func mailboxMessageOptionalStringEqual(value string, stored *string) bool {
	if value == "" {
		return stored == nil
	}
	return stored != nil && value == *stored
}

func mailboxMessageOptionalBytesEqual(left, right []byte) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return bytes.Equal(left, right)
}

func mailboxMessageOpaqueSetEqual(left, sortedRight []string) bool {
	if len(left) != len(sortedRight) || !validOpaqueIDs(left) || !validOpaqueIDs(sortedRight) {
		return false
	}
	copyLeft := append([]string(nil), left...)
	sort.Strings(copyLeft)
	return reflect.DeepEqual(copyLeft, sortedRight)
}

func mailboxMessageEngineAddressEqual(left, right EngineAddress) bool {
	return left.Email == right.Email && left.Name == right.Name
}

func mailboxMessageEngineAddressesEqual(left, right []EngineAddress) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if !mailboxMessageEngineAddressEqual(left[index], right[index]) {
			return false
		}
	}
	return true
}

func mailboxMessageStringsEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func mailboxMessageTimePointerEqual(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
}

func mailboxMessageCalendarMethodEqual(left, right *MailCalendarMethod) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func mailboxMessageCalendarRSVPTargetCopy(target *MailboxMessageCalendarRSVPTarget) *MailboxMessageCalendarRSVPTarget {
	if target == nil {
		return nil
	}
	result := *target
	result.RecurrenceID = mailboxFolderProjectionStringPointer(target.RecurrenceID)
	return &result
}

func mailboxMessageCalendarRSVPTargetValid(target *MailboxMessageCalendarRSVPTarget) bool {
	if target == nil || target.EventID == uuid.Nil || target.AttendeeVersion == 0 || !target.RSVPStatus.valid() {
		return false
	}
	switch target.Scope {
	case string(MailboxCalendarRSVPTargetScopeSeries):
		return target.RecurrenceID == nil
	case string(MailboxCalendarRSVPTargetScopeOccurrence):
		return target.RecurrenceID != nil && validMailboxCalendarRSVPTargetRecurrenceID(*target.RecurrenceID)
	default:
		return false
	}
}

func mailboxMessageDetailCopy(detail MailboxMessageDetail) MailboxMessageDetail {
	detail.From = mailboxThreadDetailProjectionAddressCopy(detail.From)
	detail.To = mailboxThreadDetailProjectionAddressesCopy(detail.To)
	detail.CC = mailboxThreadDetailProjectionAddressesCopy(detail.CC)
	detail.SentAt = mailboxFolderProjectionTimePointer(detail.SentAt)
	detail.ReceivedAt = mailboxFolderProjectionTimePointer(detail.ReceivedAt)
	detail.FolderIDs = append([]uuid.UUID(nil), detail.FolderIDs...)
	detail.CalendarMethod = copyMailCalendarMethod(detail.CalendarMethod)
	detail.CalendarEventID = mailboxFolderProjectionUUIDPointer(detail.CalendarEventID)
	detail.CalendarRSVPTarget = mailboxMessageCalendarRSVPTargetCopy(detail.CalendarRSVPTarget)
	detail.SanitizedHTML = mailboxFolderProjectionStringPointer(detail.SanitizedHTML)
	attachments := make([]MailboxMessageAttachment, len(detail.Attachments))
	for index, attachment := range detail.Attachments {
		attachments[index] = mailboxMessageAttachmentCopy(attachment)
	}
	detail.Attachments = attachments
	detail.Calendar = nil
	return detail
}

func mailboxMessageDetailValid(detail MailboxMessageDetail) bool {
	if detail.CalendarEventID != nil && *detail.CalendarEventID == uuid.Nil {
		return false
	}
	if detail.CalendarRSVPTarget != nil &&
		(!mailboxMessageCalendarRSVPTargetValid(detail.CalendarRSVPTarget) ||
			detail.CalendarEventID == nil ||
			detail.CalendarRSVPTarget.EventID != *detail.CalendarEventID) {
		return false
	}
	summary := MailboxThreadDetailProjectionMessage{
		ID: detail.ID, ThreadID: detail.ThreadID, From: detail.From, To: detail.To, CC: detail.CC,
		Subject: detail.Subject, Snippet: detail.Snippet, SentAt: detail.SentAt, ReceivedAt: detail.ReceivedAt,
		Priority: detail.Priority, SizeBytes: detail.SizeBytes, HasAttachments: detail.HasAttachments,
		IsDraft: detail.IsDraft, FolderIDs: detail.FolderIDs, Seen: detail.Seen, Flagged: detail.Flagged,
		Answered: detail.Answered, Junk: detail.Junk, CalendarMethod: detail.CalendarMethod,
		createdAt: time.Unix(1, 0).UTC(),
	}
	if !validMailboxThreadDetailProjectionMessage(summary, detail.ThreadID) || len(detail.PlainBody) > MailboxContentMaxPlainBytes ||
		utf8.RuneCountInString(detail.PlainBody) > MailboxContentMaxRunes || detail.SanitizedHTML != nil &&
		(len(*detail.SanitizedHTML) > MailboxContentMaxSanitizedHTMLBytes || !utf8.ValidString(*detail.SanitizedHTML) || strings.ContainsRune(*detail.SanitizedHTML, '\x00')) ||
		!detail.RemoteImagesBlocked || len(detail.Attachments) > MaxEngineAttachmentItems || detail.HasAttachments != (len(detail.Attachments) != 0) || detail.Calendar != nil {
		return false
	}
	for index, attachment := range detail.Attachments {
		if !mailboxMessageAttachmentValid(attachment) || index > 0 && !syncLessUUID(detail.Attachments[index-1].ID, attachment.ID) {
			return false
		}
	}
	return utf8.ValidString(detail.PlainBody) && !strings.ContainsRune(detail.PlainBody, '\x00')
}
