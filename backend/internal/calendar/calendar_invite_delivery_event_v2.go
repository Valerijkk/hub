package calendar

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/platform/events"
)

const (
	CalendarInviteCreatedSubjectV2   = "calendar.invite.created.v2"
	CalendarInviteUpdatedSubjectV2   = "calendar.invite.updated.v2"
	CalendarInviteCancelledSubjectV2 = "calendar.invite.cancelled.v2"

	calendarInviteDeliveryV2NamespaceName = "hub.calendar.invite-delivery.id.v1"
)

var calendarInviteDeliveryV2Namespace = uuid.NewSHA1(uuid.NameSpaceURL, []byte(calendarInviteDeliveryV2NamespaceName))

// CalendarInviteDeliveryEnvelopeV2Input contains only trusted identifiers for one v2 delivery intent.
type CalendarInviteDeliveryEnvelopeV2Input struct {
	EventID       uuid.UUID
	EventSequence uint
	OrganizerID   uuid.UUID
	RecurrenceID  *string
	AttendeeIDs   []uuid.UUID
}

// CalendarInviteDeliveryV2Payload is the content-free durable delivery admission payload.
type CalendarInviteDeliveryV2Payload struct {
	EventID       uuid.UUID                          `json:"event_id"`
	EventSequence uint                               `json:"event_sequence"`
	OrganizerID   uuid.UUID                          `json:"organizer_id"`
	RecurrenceID  *string                            `json:"recurrence_id,omitempty"`
	Deliveries    []CalendarInviteDeliveryV2Recipient `json:"deliveries"`
}

// CalendarInviteDeliveryV2Recipient binds one attendee to a deterministic invite identifier.
type CalendarInviteDeliveryV2Recipient struct {
	AttendeeID uuid.UUID `json:"attendee_id"`
	InviteID   uuid.UUID `json:"invite_id"`
}

type calendarInviteDeliveryV2EnvelopeJSON struct {
	Subject        string                         `json:"subject"`
	SchemaVersion  int                            `json:"schema_version"`
	EventID        uuid.UUID                      `json:"event_id"`
	OccurredAt     time.Time                      `json:"occurred_at"`
	OrganizationID uuid.UUID                      `json:"organization_id"`
	ActorID        *uuid.UUID                     `json:"actor_id"`
	CorrelationID  string                         `json:"correlation_id"`
	Payload        CalendarInviteDeliveryV2Payload `json:"payload"`
}

type calendarInviteDeliveryV2Method uint8

const (
	calendarInviteDeliveryV2MethodRequest calendarInviteDeliveryV2Method = iota + 1
	calendarInviteDeliveryV2MethodCancel
)

// NewCalendarInviteDeliveryEnvelopeV2 builds one immutable content-free v2 admission envelope.
func NewCalendarInviteDeliveryEnvelopeV2(subject string, sourceEnvelopeID uuid.UUID, occurredAt time.Time, organizationID uuid.UUID, actorID *uuid.UUID, correlationID string, input CalendarInviteDeliveryEnvelopeV2Input) (events.EnvelopeV1, error) {
	payload, valid := calendarInviteDeliveryV2PayloadFromInput(subject, sourceEnvelopeID, input)
	if !valid {
		return events.EnvelopeV1{}, ErrEventInvalid
	}
	envelope, err := events.NewEnvelopeV1(subject, sourceEnvelopeID, occurredAt.UTC(), organizationID, copyUUID(actorID), correlationID, payload)
	if err != nil {
		return events.EnvelopeV1{}, ErrEventInvalid
	}
	return envelope, nil
}

func calendarInviteDeliveryV2SubjectMethod(subject string) (calendarInviteDeliveryV2Method, bool) {
	switch subject {
	case CalendarInviteCreatedSubjectV2, CalendarInviteUpdatedSubjectV2:
		return calendarInviteDeliveryV2MethodRequest, true
	case CalendarInviteCancelledSubjectV2:
		return calendarInviteDeliveryV2MethodCancel, true
	default:
		return 0, false
	}
}

func calendarInviteDeliveryV2InviteID(sourceEnvelopeID, attendeeID uuid.UUID) uuid.UUID {
	return uuid.NewSHA1(calendarInviteDeliveryV2Namespace, []byte(sourceEnvelopeID.String()+"/"+attendeeID.String()))
}

func calendarInviteDeliveryV2PayloadFromInput(subject string, sourceEnvelopeID uuid.UUID, input CalendarInviteDeliveryEnvelopeV2Input) (CalendarInviteDeliveryV2Payload, bool) {
	if sourceEnvelopeID == uuid.Nil || input.EventID == uuid.Nil || input.OrganizerID == uuid.Nil || len(input.AttendeeIDs) > calendarCreateMaxAttendees {
		return CalendarInviteDeliveryV2Payload{}, false
	}
	deliveries := make([]CalendarInviteDeliveryV2Recipient, len(input.AttendeeIDs))
	seenAttendeeIDs := make(map[uuid.UUID]struct{}, len(input.AttendeeIDs))
	for index, attendeeID := range input.AttendeeIDs {
		if attendeeID == uuid.Nil {
			return CalendarInviteDeliveryV2Payload{}, false
		}
		if _, duplicate := seenAttendeeIDs[attendeeID]; duplicate {
			return CalendarInviteDeliveryV2Payload{}, false
		}
		seenAttendeeIDs[attendeeID] = struct{}{}
		deliveries[index] = CalendarInviteDeliveryV2Recipient{AttendeeID: attendeeID, InviteID: calendarInviteDeliveryV2InviteID(sourceEnvelopeID, attendeeID)}
	}
	sort.Slice(deliveries, func(left, right int) bool {
		return bytes.Compare(deliveries[left].InviteID[:], deliveries[right].InviteID[:]) < 0
	})
	payload := CalendarInviteDeliveryV2Payload{
		EventID: input.EventID, EventSequence: input.EventSequence, OrganizerID: input.OrganizerID,
		RecurrenceID: calendarInviteDeliveryV2CopyString(input.RecurrenceID), Deliveries: deliveries,
	}
	return payload, validCalendarInviteDeliveryV2Payload(subject, sourceEnvelopeID, payload)
}

func validCalendarInviteDeliveryV2Payload(subject string, sourceEnvelopeID uuid.UUID, payload CalendarInviteDeliveryV2Payload) bool {
	method, valid := calendarInviteDeliveryV2SubjectMethod(subject)
	if !valid || sourceEnvelopeID == uuid.Nil || payload.EventID == uuid.Nil || payload.OrganizerID == uuid.Nil || payload.Deliveries == nil || len(payload.Deliveries) > calendarCreateMaxAttendees {
		return false
	}
	if method == calendarInviteDeliveryV2MethodRequest && payload.RecurrenceID != nil {
		return false
	}
	if payload.RecurrenceID != nil && !validCalendarEventOccurrenceCancelCanonicalKey(*payload.RecurrenceID) {
		return false
	}
	seenAttendeeIDs := make(map[uuid.UUID]struct{}, len(payload.Deliveries))
	seenInviteIDs := make(map[uuid.UUID]struct{}, len(payload.Deliveries))
	for index, delivery := range payload.Deliveries {
		if delivery.AttendeeID == uuid.Nil || delivery.InviteID == uuid.Nil {
			return false
		}
		if _, duplicate := seenAttendeeIDs[delivery.AttendeeID]; duplicate {
			return false
		}
		seenAttendeeIDs[delivery.AttendeeID] = struct{}{}
		if _, duplicate := seenInviteIDs[delivery.InviteID]; duplicate {
			return false
		}
		seenInviteIDs[delivery.InviteID] = struct{}{}
		expectedInviteID := calendarInviteDeliveryV2InviteID(sourceEnvelopeID, delivery.AttendeeID)
		if expectedInviteID != delivery.InviteID {
			return false
		}
		if index > 0 && bytes.Compare(payload.Deliveries[index-1].InviteID[:], delivery.InviteID[:]) >= 0 {
			return false
		}
	}
	return true
}

func calendarInviteDeliveryV2CopyString(value *string) *string {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func calendarInviteDeliveryV2CopyPayload(payload CalendarInviteDeliveryV2Payload) CalendarInviteDeliveryV2Payload {
	copied := payload
	copied.RecurrenceID = calendarInviteDeliveryV2CopyString(payload.RecurrenceID)
	copied.Deliveries = make([]CalendarInviteDeliveryV2Recipient, len(payload.Deliveries))
	copy(copied.Deliveries, payload.Deliveries)
	return copied
}

// DecodeCalendarInviteDeliveryEnvelopeV2 returns only a strictly canonical v2 delivery payload.
func DecodeCalendarInviteDeliveryEnvelopeV2(envelope events.EnvelopeV1) (CalendarInviteDeliveryV2Payload, error) {
	encoded, err := envelope.MarshalJSON()
	if err != nil {
		return CalendarInviteDeliveryV2Payload{}, ErrEventInvalid
	}
	var value calendarInviteDeliveryV2EnvelopeJSON
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&value) != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return CalendarInviteDeliveryV2Payload{}, ErrEventInvalid
	}
	if value.SchemaVersion != 1 || value.Subject != envelope.Subject() || value.EventID != envelope.EventID() || !validCalendarInviteDeliveryV2Payload(value.Subject, value.EventID, value.Payload) {
		return CalendarInviteDeliveryV2Payload{}, ErrEventInvalid
	}
	canonical, err := events.NewEnvelopeV1(value.Subject, value.EventID, value.OccurredAt, value.OrganizationID, calendarInviteDeliveryV2CopyUUID(value.ActorID), value.CorrelationID, value.Payload)
	if err != nil {
		return CalendarInviteDeliveryV2Payload{}, ErrEventInvalid
	}
	canonicalJSON, err := canonical.MarshalJSON()
	if err != nil || !bytes.Equal(encoded, canonicalJSON) {
		return CalendarInviteDeliveryV2Payload{}, ErrEventInvalid
	}
	return calendarInviteDeliveryV2CopyPayload(value.Payload), nil
}

func calendarInviteDeliveryV2CopyUUID(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
