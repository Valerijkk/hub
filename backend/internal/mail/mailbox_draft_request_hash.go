package mail

import (
	"crypto/sha256"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

const mailboxDraftRequestHashDomain = "hub-mail-draft-request-hash-v1"

// ErrMailboxDraftRequestInvalid is the content-free canonicalization failure.
var ErrMailboxDraftRequestInvalid = errors.New("mailbox draft request invalid")

// MailboxDraftRequestHashAttachment is one ordered content-addressed attachment reference.
type MailboxDraftRequestHashAttachment struct {
	AttachmentID uuid.UUID
	SHA256       [sha256.Size]byte
}

// MailboxDraftRequestHashInput is the complete caller-supplied draft request before canonicalization.
type MailboxDraftRequestHashInput struct {
	AccountID        uuid.UUID
	DraftClientID    uuid.UUID
	FromIdentityID   uuid.UUID
	From             EngineAddress
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
	Attachments      []MailboxDraftRequestHashAttachment
	Body             MailboxContentResult
}

// MailboxDraftCanonicalRequest is the sole normalized value that later draft dispatch may use.
type MailboxDraftCanonicalRequest struct {
	AccountID        uuid.UUID
	DraftClientID    uuid.UUID
	FromIdentityID   uuid.UUID
	From             EngineAddress
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
	Attachments      []MailboxDraftRequestHashAttachment
	Body             MailboxContentResult
	RequestHash      [sha256.Size]byte
}

// CanonicalizeMailboxDraftRequest validates, copies, and hashes one complete draft request.
func CanonicalizeMailboxDraftRequest(input MailboxDraftRequestHashInput) (MailboxDraftCanonicalRequest, error) {
	if input.AccountID == uuid.Nil || input.DraftClientID == uuid.Nil || input.FromIdentityID == uuid.Nil || !mailboxDraftRequestHashUUIDPointerValid(input.ReplyToMessageID) || !mailboxDraftRequestHashUUIDPointerValid(input.ForwardMessageID) {
		return MailboxDraftCanonicalRequest{}, ErrMailboxDraftRequestInvalid
	}

	from, err := NormalizeAddress(input.From)
	if err != nil {
		return MailboxDraftCanonicalRequest{}, ErrMailboxDraftRequestInvalid
	}
	to, err := NormalizeAddressSet(input.To, MaxEngineDraftRecipientItems)
	if err != nil {
		return MailboxDraftCanonicalRequest{}, ErrMailboxDraftRequestInvalid
	}
	cc, err := NormalizeAddressSet(input.CC, MaxEngineDraftRecipientItems)
	if err != nil {
		return MailboxDraftCanonicalRequest{}, ErrMailboxDraftRequestInvalid
	}
	bcc, err := NormalizeAddressSet(input.BCC, MaxEngineDraftRecipientItems)
	if err != nil {
		return MailboxDraftCanonicalRequest{}, ErrMailboxDraftRequestInvalid
	}
	replyTo, err := NormalizeAddressSet(input.ReplyTo, MaxEngineDraftRecipientItems)
	if err != nil || len(to)+len(cc)+len(bcc) > MaxEngineDraftTotalRecipientItems || !mailboxDraftRequestHashRecipientsDistinct(to, cc, bcc) {
		return MailboxDraftCanonicalRequest{}, ErrMailboxDraftRequestInvalid
	}
	if !mailboxDraftRequestHashHeaderValid(input.Subject) || !mailboxDraftRequestHashHeaderValid(input.InReplyTo) || input.InReplyTo != "" && !validThreadMessageID(input.InReplyTo) || len(input.References) > MaxEngineReferenceItems || !mailboxDraftRequestHashPriorityValid(input.Priority) {
		return MailboxDraftCanonicalRequest{}, ErrMailboxDraftRequestInvalid
	}
	for _, reference := range input.References {
		if reference == "" || !mailboxDraftRequestHashHeaderValid(reference) || !validThreadMessageID(reference) {
			return MailboxDraftCanonicalRequest{}, ErrMailboxDraftRequestInvalid
		}
	}
	attachments, ok := mailboxDraftRequestHashAttachments(input.Attachments)
	if !ok {
		return MailboxDraftCanonicalRequest{}, ErrMailboxDraftRequestInvalid
	}
	body, ok := mailboxDraftRequestHashBody(input.Body)
	if !ok {
		return MailboxDraftCanonicalRequest{}, ErrMailboxDraftRequestInvalid
	}

	canonical := MailboxDraftCanonicalRequest{
		AccountID:        input.AccountID,
		DraftClientID:    input.DraftClientID,
		FromIdentityID:   input.FromIdentityID,
		From:             mailboxDraftRequestHashAddressCopy(from),
		To:               mailboxDraftRequestHashAddressesCopy(to),
		CC:               mailboxDraftRequestHashAddressesCopy(cc),
		BCC:              mailboxDraftRequestHashAddressesCopy(bcc),
		ReplyTo:          mailboxDraftRequestHashAddressesCopy(replyTo),
		Subject:          strings.Clone(input.Subject),
		InReplyTo:        strings.Clone(input.InReplyTo),
		References:       mailboxDraftRequestHashStringsCopy(input.References),
		Priority:         input.Priority,
		ReplyToMessageID: mailboxDraftRequestHashUUIDPointerCopy(input.ReplyToMessageID),
		ForwardMessageID: mailboxDraftRequestHashUUIDPointerCopy(input.ForwardMessageID),
		Attachments:      attachments,
		Body:             body,
	}
	canonical.RequestHash = sha256.Sum256(mailboxDraftRequestHashFrame(canonical))
	return canonical, nil
}

func mailboxDraftRequestHashUUIDPointerValid(value *uuid.UUID) bool {
	return value == nil || *value != uuid.Nil
}

func mailboxDraftRequestHashRecipientsDistinct(groups ...[]EngineAddress) bool {
	seen := make(map[string]struct{})
	for _, group := range groups {
		for _, address := range group {
			if _, duplicate := seen[address.Email]; duplicate {
				return false
			}
			seen[address.Email] = struct{}{}
		}
	}
	return true
}

// mailboxDraftRequestHashHeaderValid intentionally repeats the production header policy.
func mailboxDraftRequestHashHeaderValid(value string) bool {
	if len(value) > MaxEngineHeaderValueBytes || !utf8.ValidString(value) {
		return false
	}
	for _, runeValue := range value {
		if runeValue < 0x20 || runeValue >= 0x7f && runeValue <= 0x9f {
			return false
		}
	}
	return true
}

func mailboxDraftRequestHashPriorityValid(value EngineMessagePriority) bool {
	return value == EnginePriorityLow || value == EnginePriorityNormal || value == EnginePriorityHigh
}

func mailboxDraftRequestHashAttachments(values []MailboxDraftRequestHashAttachment) ([]MailboxDraftRequestHashAttachment, bool) {
	if len(values) > MaxEngineAttachmentItems {
		return nil, false
	}
	result := make([]MailboxDraftRequestHashAttachment, len(values))
	seen := make(map[uuid.UUID]struct{}, len(values))
	var zero [sha256.Size]byte
	for index, value := range values {
		if value.AttachmentID == uuid.Nil || value.SHA256 == zero {
			return nil, false
		}
		if _, duplicate := seen[value.AttachmentID]; duplicate {
			return nil, false
		}
		seen[value.AttachmentID] = struct{}{}
		result[index] = value
	}
	return result, true
}

func mailboxDraftRequestHashBody(value MailboxContentResult) (MailboxContentResult, bool) {
	if value.SanitizerVersion != MailboxContentSanitizerVersionV1 || !value.RemoteImagesBlocked {
		return MailboxContentResult{}, false
	}

	var source string
	switch value.SourceFormat {
	case MailboxContentFormatPlain:
		if value.SanitizedHTML != "" {
			return MailboxContentResult{}, false
		}
		source = value.PlainBody
	case MailboxContentFormatRich:
		source = value.SanitizedHTML
	default:
		return MailboxContentResult{}, false
	}

	normalized, err := NormalizeMailboxContent(MailboxContentInput{SourceFormat: value.SourceFormat, Body: source})
	if err != nil || normalized != value {
		return MailboxContentResult{}, false
	}
	return MailboxContentResult{
		SourceFormat:        normalized.SourceFormat,
		SanitizerVersion:    normalized.SanitizerVersion,
		PlainBody:           strings.Clone(normalized.PlainBody),
		SanitizedHTML:       strings.Clone(normalized.SanitizedHTML),
		RemoteImagesBlocked: normalized.RemoteImagesBlocked,
	}, true
}

func mailboxDraftRequestHashFrame(value MailboxDraftCanonicalRequest) []byte {
	attachments := make([][]byte, len(value.Attachments))
	for index, attachment := range value.Attachments {
		attachments[index] = canonicalRecord(
			canonicalTLV(1, canonicalUUID(attachment.AttachmentID)),
			canonicalTLV(2, attachment.SHA256[:]),
		)
	}
	plainDigest := sha256.Sum256([]byte(value.Body.PlainBody))
	richDigest := sha256.Sum256([]byte(value.Body.SanitizedHTML))
	return canonicalRecord(
		canonicalTLV(1, []byte(mailboxDraftRequestHashDomain)),
		canonicalTLV(2, canonicalUUID(value.AccountID)),
		canonicalTLV(3, canonicalUUID(value.DraftClientID)),
		canonicalTLV(4, canonicalUUID(value.FromIdentityID)),
		canonicalTLV(5, canonicalAddressRecord(value.From)),
		canonicalTLV(6, canonicalAddressCollection(value.To)),
		canonicalTLV(7, canonicalAddressCollection(value.CC)),
		canonicalTLV(8, canonicalAddressCollection(value.BCC)),
		canonicalTLV(9, canonicalAddressCollection(value.ReplyTo)),
		canonicalTLV(10, []byte(value.Subject)),
		canonicalTLV(11, []byte(value.InReplyTo)),
		canonicalTLV(12, canonicalStringCollection(value.References)),
		canonicalTLV(13, []byte(value.Priority)),
		canonicalTLV(14, mailboxDraftRequestHashOptionalUUID(value.ReplyToMessageID)),
		canonicalTLV(15, mailboxDraftRequestHashOptionalUUID(value.ForwardMessageID)),
		canonicalTLV(16, canonicalCollection(attachments)),
		canonicalTLV(17, canonicalU16(uint16(value.Body.SourceFormat))),
		canonicalTLV(18, canonicalU16(uint16(value.Body.SanitizerVersion))),
		canonicalTLV(19, plainDigest[:]),
		canonicalTLV(20, richDigest[:]),
	)
}

func mailboxDraftRequestHashOptionalUUID(value *uuid.UUID) []byte {
	if value == nil {
		return []byte{0}
	}
	return append([]byte{1}, canonicalUUID(*value)...)
}

func mailboxDraftRequestHashAddressCopy(value EngineAddress) EngineAddress {
	return EngineAddress{Email: strings.Clone(value.Email), Name: strings.Clone(value.Name)}
}

func mailboxDraftRequestHashAddressesCopy(values []EngineAddress) []EngineAddress {
	if values == nil {
		return nil
	}
	result := make([]EngineAddress, len(values))
	for index, value := range values {
		result[index] = mailboxDraftRequestHashAddressCopy(value)
	}
	return result
}

func mailboxDraftRequestHashStringsCopy(values []string) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = strings.Clone(value)
	}
	return result
}

func mailboxDraftRequestHashUUIDPointerCopy(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}
