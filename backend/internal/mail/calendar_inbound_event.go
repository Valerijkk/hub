package mail

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"

	"github.com/google/uuid"
)

var errCalendarInboundEventInvalid = errors.New("calendar inbound event invalid")

type calendarInboundEventResult struct {
	delivery       CalendarInboundDelivery
	noCalendarPart bool
}

func decodeCalendarInboundEvent(subject string, raw []byte) (calendarInboundEventResult, error) {
	if subject != "mail.received.v1" || len(raw) == 0 || len(raw) > MaxEngineProjectionItemBytes || !json.Valid(raw) {
		return calendarInboundEventResult{}, errCalendarInboundEventInvalid
	}

	var envelope canonicalEnvelopeJSON
	if !calendarInboundEventDecodeJSON(raw, &envelope) || envelope.Subject == nil || envelope.SchemaVersion == nil || envelope.EventID == nil || envelope.OccurredAt == nil || envelope.OrganizationID == nil || envelope.ActorID == nil || envelope.CorrelationID == nil || envelope.Payload == nil || *envelope.Subject != "mail.received.v1" || *envelope.SchemaVersion != 1 || *envelope.EventID == uuid.Nil || *envelope.OrganizationID == uuid.Nil || !syncValidEventOccurredAt(*envelope.OccurredAt) || !syncValidCorrelationID(*envelope.CorrelationID) {
		return calendarInboundEventResult{}, errCalendarInboundEventInvalid
	}

	actorID, valid := calendarInboundEventActorID(envelope.ActorID)
	if !valid {
		return calendarInboundEventResult{}, errCalendarInboundEventInvalid
	}
	payloadJSON := bytes.TrimSpace(envelope.Payload)
	if len(payloadJSON) < 2 || payloadJSON[0] != '{' || payloadJSON[len(payloadJSON)-1] != '}' || !json.Valid(payloadJSON) {
		return calendarInboundEventResult{}, errCalendarInboundEventInvalid
	}
	var received mailReceivedEventJSON
	if !calendarInboundEventDecodeJSON(payloadJSON, &received) || received.MailboxID == uuid.Nil {
		return calendarInboundEventResult{}, errCalendarInboundEventInvalid
	}

	receivedPayload, valid := canonicalReceivedPayload(received.MailboxID, received)
	if !valid {
		return calendarInboundEventResult{}, errCalendarInboundEventInvalid
	}
	rebuilt, err := NewMailReceivedEnvelopeV1(*envelope.EventID, *envelope.OccurredAt, *envelope.OrganizationID, actorID, *envelope.CorrelationID, receivedPayload)
	if err != nil {
		return calendarInboundEventResult{}, errCalendarInboundEventInvalid
	}
	frozen, err := rebuilt.MarshalJSON()
	if err != nil || !bytes.Equal(raw, frozen) {
		return calendarInboundEventResult{}, errCalendarInboundEventInvalid
	}

	if received.CalendarMethod == nil && received.CalendarPartLocatorHash == nil {
		return calendarInboundEventResult{noCalendarPart: true}, nil
	}
	locatorHash, valid := decodeMailCalendarPartLocatorHash(received.CalendarPartLocatorHash)
	if !valid || received.CalendarMethod == nil || len(locatorHash) != sha256.Size {
		return calendarInboundEventResult{}, errCalendarInboundEventInvalid
	}
	var locator [sha256.Size]byte
	copy(locator[:], locatorHash)
	delivery, err := newCalendarInboundDelivery(*envelope.EventID, *envelope.OrganizationID, received.MailboxID, received.MessageID, received.From.Email, received.CalendarMethod, received.AuthenticationClass, locator)
	if err != nil {
		return calendarInboundEventResult{}, errCalendarInboundEventInvalid
	}
	return calendarInboundEventResult{delivery: delivery}, nil
}

func calendarInboundEventDecodeJSON(raw []byte, target any) bool {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return false
	}
	var extra any
	return decoder.Decode(&extra) == io.EOF
}

func calendarInboundEventActorID(raw json.RawMessage) (*uuid.UUID, bool) {
	actorJSON := bytes.TrimSpace(raw)
	if bytes.Equal(actorJSON, []byte("null")) {
		return nil, true
	}
	var actorID uuid.UUID
	if !calendarInboundEventDecodeJSON(actorJSON, &actorID) || actorID == uuid.Nil {
		return nil, false
	}
	return &actorID, true
}
