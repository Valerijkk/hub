package mail

import (
	"context"
	"crypto/sha256"

	"github.com/google/uuid"
)

// MaxCalendarPartBytes limits one decoded iTIP calendar part.
const MaxCalendarPartBytes = 1_048_576

// CalendarPartReadRequest is the opaque Calendar-to-Mail request for one sealed part.
type CalendarPartReadRequest struct {
	organizationID uuid.UUID
	mailboxID      uuid.UUID
	messageID      uuid.UUID
	locatorHash    [sha256.Size]byte
	maxBytes       int
}

// NewCalendarPartReadRequest validates one opaque calendar-part read request.
func NewCalendarPartReadRequest(organizationID, mailboxID, messageID uuid.UUID, locatorHash [sha256.Size]byte, maxBytes int) (CalendarPartReadRequest, error) {
	if organizationID == uuid.Nil || mailboxID == uuid.Nil || messageID == uuid.Nil || calendarPartLocatorHashZero(locatorHash) || maxBytes < 1 || maxBytes > MaxCalendarPartBytes {
		return CalendarPartReadRequest{}, newCalendarPartSourceError(CalendarPartSourceErrorInvalid)
	}
	return CalendarPartReadRequest{
		organizationID: organizationID,
		mailboxID:      mailboxID,
		messageID:      messageID,
		locatorHash:    locatorHash,
		maxBytes:       maxBytes,
	}, nil
}

// CalendarPartSource reads one sealed decoded iTIP calendar part for Calendar.
type CalendarPartSource interface {
	ReadCalendarPart(context.Context, CalendarPartReadRequest) (CalendarPartReadResult, error)
}

// CalendarPartReadResult is the opaque Mail-to-Calendar result for one admitted part.
type CalendarPartReadResult struct {
	ownerID             uuid.UUID
	sender              string
	method              MailCalendarMethod
	authenticationClass MailAuthenticationClass
	itip                []byte
}

// OwnerID returns the admitted owner identifier.
func (r CalendarPartReadResult) OwnerID() uuid.UUID { return r.ownerID }

// Sender returns the normalized sender address.
func (r CalendarPartReadResult) Sender() string { return r.sender }

// Method returns the admitted iTIP method.
func (r CalendarPartReadResult) Method() MailCalendarMethod { return r.method }

// AuthenticationClass returns the admitted authentication class.
func (r CalendarPartReadResult) AuthenticationClass() MailAuthenticationClass {
	return r.authenticationClass
}

// ITIP returns a defensive copy of the decoded iTIP bytes.
func (r CalendarPartReadResult) ITIP() []byte { return append([]byte(nil), r.itip...) }

func newCalendarPartReadResult(ownerID uuid.UUID, sender string, method MailCalendarMethod, authenticationClass MailAuthenticationClass, itip []byte, maxBytes int) (CalendarPartReadResult, error) {
	normalizedSender, ok := normalizeCalendarPartSender(sender)
	if ownerID == uuid.Nil || !ok || !validCalendarPartMethod(method) || !trustedCalendarPartAuthenticationClass(authenticationClass) || maxBytes < 1 || maxBytes > MaxCalendarPartBytes || len(itip) == 0 {
		return CalendarPartReadResult{}, newCalendarPartSourceError(CalendarPartSourceErrorInvalid)
	}
	if len(itip) > maxBytes {
		return CalendarPartReadResult{}, newCalendarPartSourceError(CalendarPartSourceErrorPayloadTooLarge)
	}
	return CalendarPartReadResult{
		ownerID:             ownerID,
		sender:              normalizedSender,
		method:              method,
		authenticationClass: authenticationClass,
		itip:                append([]byte(nil), itip...),
	}, nil
}

// CalendarInboundDelivery is the opaque sanitized Mail-to-Calendar delivery.
type CalendarInboundDelivery struct {
	envelopeID           uuid.UUID
	organizationID       uuid.UUID
	mailboxID            uuid.UUID
	messageID            uuid.UUID
	sender               string
	method               MailCalendarMethod
	authenticationClass  MailAuthenticationClass
	locatorHash          [sha256.Size]byte
}

// EnvelopeID returns the source event identifier.
func (d CalendarInboundDelivery) EnvelopeID() uuid.UUID { return d.envelopeID }

// OrganizationID returns the owning organization identifier.
func (d CalendarInboundDelivery) OrganizationID() uuid.UUID { return d.organizationID }

// MailboxID returns the owning mailbox identifier.
func (d CalendarInboundDelivery) MailboxID() uuid.UUID { return d.mailboxID }

// MessageID returns the durable Mail message identifier.
func (d CalendarInboundDelivery) MessageID() uuid.UUID { return d.messageID }

// Sender returns the normalized sender address.
func (d CalendarInboundDelivery) Sender() string { return d.sender }

// Method returns the admitted iTIP method.
func (d CalendarInboundDelivery) Method() MailCalendarMethod { return d.method }

// AuthenticationClass returns the admitted authentication class.
func (d CalendarInboundDelivery) AuthenticationClass() MailAuthenticationClass {
	return d.authenticationClass
}

// LocatorHash returns the locator digest by value.
func (d CalendarInboundDelivery) LocatorHash() [sha256.Size]byte { return d.locatorHash }

// CalendarInboundSink receives a sanitized Mail delivery in Calendar.
type CalendarInboundSink interface {
	HandleCalendarInbound(context.Context, CalendarInboundDelivery) error
}

func newCalendarInboundDelivery(envelopeID, organizationID, mailboxID, messageID uuid.UUID, sender string, method *MailCalendarMethod, authenticationClass MailAuthenticationClass, locatorHash [sha256.Size]byte) (CalendarInboundDelivery, error) {
	normalizedSender, ok := normalizeCalendarPartSender(sender)
	if envelopeID == uuid.Nil || organizationID == uuid.Nil || mailboxID == uuid.Nil || messageID == uuid.Nil || !ok || method == nil || !validCalendarPartMethod(*method) || !trustedCalendarPartAuthenticationClass(authenticationClass) || calendarPartLocatorHashZero(locatorHash) {
		return CalendarInboundDelivery{}, newCalendarPartSourceError(CalendarPartSourceErrorInvalid)
	}
	return CalendarInboundDelivery{
		envelopeID:          envelopeID,
		organizationID:      organizationID,
		mailboxID:           mailboxID,
		messageID:           messageID,
		sender:              normalizedSender,
		method:              *method,
		authenticationClass: authenticationClass,
		locatorHash:         locatorHash,
	}, nil
}

func calendarPartLocatorHashZero(value [sha256.Size]byte) bool {
	return value == [sha256.Size]byte{}
}

func validCalendarPartMethod(value MailCalendarMethod) bool {
	return value == MailCalendarMethodRequest || value == MailCalendarMethodCancel || value == MailCalendarMethodReply
}

func trustedCalendarPartAuthenticationClass(value MailAuthenticationClass) bool {
	return value == MailAuthenticationClassTrustedInternal || value == MailAuthenticationClassDMARCAligned
}

func normalizeCalendarPartSender(value string) (string, bool) {
	normalized, err := NormalizeAddress(EngineAddress{Email: value})
	if err != nil || normalized.Name != "" {
		return "", false
	}
	return normalized.Email, true
}
