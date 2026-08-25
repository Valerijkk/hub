package mail

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/ical"
)

const calendarMIMEPartBindingDomain = "hub.calendar.mail.mime-part.binding.v1"

// calendarMIMEPart is a sealed request-lifetime calendar MIME payload. Its
// serialized iTIP bytes never cross a durable delivery boundary.
type calendarMIMEPart struct {
	sourceDeliveryID CalendarMailDeliveryID
	organizationID   uuid.UUID
	method           CalendarMailDeliveryMethod
	recipient        EngineAddress
	serialized       []byte
	contentHash      [sha256.Size]byte
	bindingHash      [sha256.Size]byte
}

func newCalendarMIMEPart(sourceDeliveryID CalendarMailDeliveryID, organizationID uuid.UUID, recipient EngineAddress, message ical.ITIPMessage) (calendarMIMEPart, error) {
	if !validCalendarMailDeliveryID(sourceDeliveryID) || organizationID == uuid.Nil {
		return calendarMIMEPart{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}

	normalizedRecipient, err := NormalizeAddress(recipient)
	if err != nil {
		return calendarMIMEPart{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	canonicalMessage, identity, ok := calendarMailDeliveryCanonicalITIPMessage(message)
	if !ok {
		return calendarMIMEPart{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	method := calendarMailDeliveryMethodFromITIP(canonicalMessage.Method)
	if !validCalendarMIMEPartMethod(method) {
		return calendarMIMEPart{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	if !calendarMIMEPartRecipientMatchesMessage(method, normalizedRecipient, canonicalMessage) {
		return calendarMIMEPart{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}

	serialized, err := ical.SerializeITIP(canonicalMessage)
	if err != nil || len(serialized) == 0 {
		return calendarMIMEPart{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	if len(serialized) > MaxCalendarPartBytes {
		return calendarMIMEPart{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorTooLarge)
	}
	if !validCalendarMIMEPartSerialized(serialized, method, normalizedRecipient, identity) {
		return calendarMIMEPart{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}

	contentHash := calendarMIMEPartContentHash(method, serialized)
	return calendarMIMEPart{
		sourceDeliveryID: sourceDeliveryID,
		organizationID:   organizationID,
		method:           method,
		recipient:        normalizedRecipient,
		serialized:       append([]byte(nil), serialized...),
		contentHash:      contentHash,
		bindingHash:      calendarMIMEPartBindingHash(organizationID, sourceDeliveryID, method, normalizedRecipient, contentHash),
	}, nil
}

func (part calendarMIMEPart) valid() bool {
	if !validCalendarMailDeliveryID(part.sourceDeliveryID) || part.organizationID == uuid.Nil ||
		!validCalendarMIMEPartMethod(part.method) || len(part.serialized) == 0 || len(part.serialized) > MaxCalendarPartBytes {
		return false
	}
	normalizedRecipient, err := NormalizeAddress(part.recipient)
	if err != nil || normalizedRecipient != part.recipient {
		return false
	}
	if part.contentHash != calendarMIMEPartContentHash(part.method, part.serialized) ||
		part.bindingHash != calendarMIMEPartBindingHash(part.organizationID, part.sourceDeliveryID, part.method, part.recipient, part.contentHash) {
		return false
	}
	parsed, err := ical.ParseITIP(part.serialized, ical.DefaultITIPParseLimits())
	if err != nil {
		return false
	}
	canonical, _, ok := calendarMailDeliveryCanonicalITIPMessage(parsed)
	if !ok || calendarMailDeliveryMethodFromITIP(canonical.Method) != part.method ||
		!calendarMIMEPartRecipientMatchesMessage(part.method, normalizedRecipient, canonical) {
		return false
	}
	reserialized, err := ical.SerializeITIP(canonical)
	return err == nil && bytes.Equal(part.serialized, reserialized)
}

func (part calendarMIMEPart) deliveryID() CalendarMailDeliveryID { return part.sourceDeliveryID }

func (part calendarMIMEPart) organization() uuid.UUID { return part.organizationID }

func (part calendarMIMEPart) mimeMethod() string {
	if !validCalendarMIMEPartMethod(part.method) {
		return ""
	}
	return string(part.method)
}

func (part calendarMIMEPart) recipientAddress() EngineAddress { return part.recipient }

func (part calendarMIMEPart) calendarBytes() []byte { return append([]byte(nil), part.serialized...) }

func (part calendarMIMEPart) contentDigest() [sha256.Size]byte { return part.contentHash }

func (part calendarMIMEPart) bindingDigest() [sha256.Size]byte { return part.bindingHash }

func validCalendarMIMEPartMethod(method CalendarMailDeliveryMethod) bool {
	switch method {
	case CalendarMailDeliveryMethodRequest, CalendarMailDeliveryMethodCancel, CalendarMailDeliveryMethodReply:
		return true
	default:
		return false
	}
}

func calendarMIMEPartRecipientMatchesMessage(method CalendarMailDeliveryMethod, recipient EngineAddress, message ical.ITIPMessage) bool {
	normalizedRecipient, err := NormalizeAddress(recipient)
	if err != nil {
		return false
	}
	switch method {
	case CalendarMailDeliveryMethodRequest, CalendarMailDeliveryMethodCancel:
		return calendarMailDeliveryMessageHasAttendee(message, normalizedRecipient.Email)
	case CalendarMailDeliveryMethodReply:
		return normalizedRecipient.Email == message.Organizer.Email
	default:
		return false
	}
}

func validCalendarMIMEPartSerialized(serialized []byte, method CalendarMailDeliveryMethod, recipient EngineAddress, expectedIdentity ical.ITIPIdentity) bool {
	parsed, err := ical.ParseITIP(serialized, ical.DefaultITIPParseLimits())
	if err != nil {
		return false
	}
	canonical, identity, ok := calendarMailDeliveryCanonicalITIPMessage(parsed)
	if !ok || identity != expectedIdentity || calendarMailDeliveryMethodFromITIP(canonical.Method) != method ||
		!calendarMIMEPartRecipientMatchesMessage(method, recipient, canonical) {
		return false
	}
	reserialized, err := ical.SerializeITIP(canonical)
	return err == nil && bytes.Equal(serialized, reserialized)
}

func calendarMIMEPartContentHash(method CalendarMailDeliveryMethod, serialized []byte) [sha256.Size]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte(method))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(serialized)
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func calendarMIMEPartBindingHash(organizationID uuid.UUID, sourceDeliveryID CalendarMailDeliveryID, method CalendarMailDeliveryMethod, recipient EngineAddress, contentHash [sha256.Size]byte) [sha256.Size]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte(calendarMIMEPartBindingDomain))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(organizationID[:])
	deliveryID := sourceDeliveryID.UUID()
	_, _ = hash.Write(deliveryID[:])
	calendarMIMEPartHashString(hash, string(method))
	calendarMIMEPartHashString(hash, recipient.Email)
	calendarMIMEPartHashString(hash, recipient.Name)
	_, _ = hash.Write(contentHash[:])
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func calendarMIMEPartHashString(hash interface{ Write([]byte) (int, error) }, value string) {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value)))
	_, _ = hash.Write(length[:])
	_, _ = hash.Write([]byte(value))
}
