package events

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
)

type validationError string

const (
	envelopeSchemaVersionV1 = 1
	maxSubjectBytes         = 255
	maxCorrelationIDBytes   = 128
	ErrInvalidSubject        validationError = "events: invalid subject"
	ErrInvalidEventID        validationError = "events: invalid event id"
	ErrInvalidOccurredAt     validationError = "events: invalid occurred at"
	ErrInvalidOrganizationID validationError = "events: invalid organization id"
	ErrInvalidActorID        validationError = "events: invalid actor id"
	ErrInvalidCorrelationID  validationError = "events: invalid correlation id"
	ErrInvalidPayload        validationError = "events: invalid payload"
)

func (err validationError) Error() string {
	return string(err)
}

// EnvelopeV1 is an immutable, complete validated JSON value for a version-one event.
type EnvelopeV1 struct {
	subject    string
	eventID    uuid.UUID
	frozenJSON string
}

type envelopeJSONV1 struct {
	Subject        string          `json:"subject"`
	SchemaVersion  int             `json:"schema_version"`
	EventID        uuid.UUID       `json:"event_id"`
	OccurredAt     time.Time       `json:"occurred_at"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	ActorID        *uuid.UUID      `json:"actor_id"`
	CorrelationID  string          `json:"correlation_id"`
	Payload        json.RawMessage `json:"payload"`
}

// NewEnvelopeV1 validates and snapshots explicit inputs without I/O or generated values.
func NewEnvelopeV1(
	subject string,
	eventID uuid.UUID,
	occurredAt time.Time,
	organizationID uuid.UUID,
	actorID *uuid.UUID,
	correlationID string,
	payload any,
) (EnvelopeV1, error) {
	if !validSubject(subject) {
		return EnvelopeV1{}, ErrInvalidSubject
	}
	if eventID == uuid.Nil {
		return EnvelopeV1{}, ErrInvalidEventID
	}
	if !validOccurredAt(occurredAt) {
		return EnvelopeV1{}, ErrInvalidOccurredAt
	}
	if organizationID == uuid.Nil {
		return EnvelopeV1{}, ErrInvalidOrganizationID
	}
	if actorID != nil && *actorID == uuid.Nil {
		return EnvelopeV1{}, ErrInvalidActorID
	}

	correlationID = strings.TrimSpace(correlationID)
	if !utf8.ValidString(correlationID) || correlationID == "" || len(correlationID) > maxCorrelationIDBytes {
		return EnvelopeV1{}, ErrInvalidCorrelationID
	}

	payloadJSON, valid := marshalObjectPayload(payload)
	if !valid {
		return EnvelopeV1{}, ErrInvalidPayload
	}

	var actorIDSnapshot *uuid.UUID
	if actorID != nil {
		actorIDCopy := *actorID
		actorIDSnapshot = &actorIDCopy
	}

	encoded, err := json.Marshal(envelopeJSONV1{
		Subject:        subject,
		SchemaVersion:  envelopeSchemaVersionV1,
		EventID:        eventID,
		OccurredAt:     occurredAt.UTC(),
		OrganizationID: organizationID,
		ActorID:        actorIDSnapshot,
		CorrelationID:  correlationID,
		Payload:        payloadJSON,
	})
	if err != nil {
		return EnvelopeV1{}, ErrInvalidPayload
	}

	return EnvelopeV1{
		subject:    subject,
		eventID:    eventID,
		frozenJSON: string(encoded),
	}, nil
}

// Subject returns a copy of the validated event subject.
func (envelope EnvelopeV1) Subject() string {
	return envelope.subject
}

// EventID returns a copy of the validated event identifier.
func (envelope EnvelopeV1) EventID() uuid.UUID {
	return envelope.eventID
}

// MarshalJSON returns a fresh copy of the complete frozen envelope bytes.
func (envelope EnvelopeV1) MarshalJSON() ([]byte, error) {
	return []byte(envelope.frozenJSON), nil
}

func validSubject(subject string) bool {
	if !utf8.ValidString(subject) || len(subject) == 0 || len(subject) > maxSubjectBytes {
		return false
	}
	if !strings.HasSuffix(subject, ".v1") {
		return false
	}
	for _, token := range strings.Split(subject, ".") {
		if token == "" {
			return false
		}
	}
	for _, r := range subject {
		if r == '*' || r == '>' || unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func validOccurredAt(occurredAt time.Time) bool {
	if occurredAt.IsZero() {
		return false
	}
	_, offset := occurredAt.Zone()
	if offset != 0 {
		return false
	}
	_, err := occurredAt.UTC().MarshalJSON()
	return err == nil
}

func marshalObjectPayload(payload any) ([]byte, bool) {
	if nilValue(payload) {
		return nil, false
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, false
	}
	encoded = bytes.TrimSpace(encoded)
	if len(encoded) < 2 || encoded[0] != '{' || encoded[len(encoded)-1] != '}' {
		return nil, false
	}
	return encoded, true
}

func nilValue(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
