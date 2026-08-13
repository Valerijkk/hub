package calendar_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abramov-hub/hub/backend/internal/platform/events"
)

var (
	fixedEventID        = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	fixedOrganizationID = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	fixedActorID        = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	fixedAttendeeID     = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	fixedOccurredAt     = time.Date(2030, time.May, 6, 7, 8, 9, 123456789, time.UTC)
)

const (
	sensitiveBody           = "SENSITIVE-BODY-04L"
	sensitiveMIME           = "SENSITIVE-MIME-04L"
	sensitiveCredential     = "SENSITIVE-CREDENTIAL-04L"
	sensitiveIdempotencyKey = "SENSITIVE-IDEMPOTENCY-KEY-04L"
)

type envelopePayloadFixture struct {
	NewStatus  string    `json:"new_status"`
	AttendeeID uuid.UUID `json:"attendee_id"`
}

type nestedPayloadFixture struct {
	Status   string         `json:"status"`
	Metadata map[string]any `json:"metadata"`
}

type mutableStructPayload struct {
	Name     string            `json:"name"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
}

type rejectingPayloadFixture struct {
	Body           string    `json:"body"`
	MIME           string    `json:"mime"`
	Credential     string    `json:"credential"`
	IdempotencyKey string    `json:"idempotency_key"`
	AttendeeID     uuid.UUID `json:"attendee_id"`
}

type customPayloadMarshaler struct {
	raw []byte
}

func (payload *customPayloadMarshaler) MarshalJSON() ([]byte, error) {
	return payload.raw, nil
}

type failingPayloadMarshaler struct{}

func (failingPayloadMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errors.New("SENSITIVE-MARSHAL-ERROR-04L")
}

type nilPayloadMarker interface {
	isNilPayloadMarker()
}

type nilPayloadMarkerValue struct{}

func (*nilPayloadMarkerValue) isNilPayloadMarker() {}

func TestNewEnvelopeV1_SerializesCompleteEnvelopeDeterministically(t *testing.T) {
	payload := envelopePayloadFixture{
		NewStatus:  "accepted",
		AttendeeID: fixedAttendeeID,
	}

	envelope, err := events.NewEnvelopeV1(
		"calendar.meeting.rsvp.changed.v1",
		fixedEventID,
		fixedOccurredAt,
		fixedOrganizationID,
		nil,
		"  corr-calendar-04l  ",
		payload,
	)
	if err != nil {
		t.Fatalf("NewEnvelopeV1() error = %v", err)
	}

	want := []byte(`{"subject":"calendar.meeting.rsvp.changed.v1","schema_version":1,"event_id":"11111111-1111-4111-8111-111111111111","occurred_at":"2030-05-06T07:08:09.123456789Z","organization_id":"22222222-2222-4222-8222-222222222222","actor_id":null,"correlation_id":"corr-calendar-04l","payload":{"new_status":"accepted","attendee_id":"44444444-4444-4444-8444-444444444444"}}`)

	first := mustMarshalEnvelope(t, envelope)
	second := mustMarshalEnvelope(t, envelope)
	direct := mustMarshalEnvelopeDirect(t, envelope)
	if !bytes.Equal(first, want) {
		t.Fatalf("json.Marshal(envelope) = %s, want %s", first, want)
	}
	if !bytes.Equal(second, want) || !bytes.Equal(direct, want) {
		t.Fatalf("reused envelope bytes changed: second=%s direct=%s want=%s", second, direct, want)
	}
	if got := envelope.Subject(); got != "calendar.meeting.rsvp.changed.v1" {
		t.Fatalf("Subject() = %q", got)
	}
	if got := envelope.EventID(); got != fixedEventID {
		t.Fatalf("EventID() = %s", got)
	}
}

func TestNewEnvelopeV1_EncodesNullableActorAndNestedPayloadExactly(t *testing.T) {
	actorInput := fixedActorID
	envelope, err := events.NewEnvelopeV1(
		"calendar.meeting.rsvp.responded.v1",
		fixedEventID,
		fixedOccurredAt,
		fixedOrganizationID,
		&actorInput,
		"corr-nested-04l",
		nestedPayloadFixture{
			Status: "tentative",
			Metadata: map[string]any{
				"source":  "calendar",
				"attempt": 2,
			},
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelopeV1() error = %v", err)
	}
	actorInput = uuid.MustParse("55555555-5555-4555-8555-555555555555")

	want := []byte(`{"subject":"calendar.meeting.rsvp.responded.v1","schema_version":1,"event_id":"11111111-1111-4111-8111-111111111111","occurred_at":"2030-05-06T07:08:09.123456789Z","organization_id":"22222222-2222-4222-8222-222222222222","actor_id":"33333333-3333-4333-8333-333333333333","correlation_id":"corr-nested-04l","payload":{"status":"tentative","metadata":{"attempt":2,"source":"calendar"}}}`)
	encoded := mustMarshalEnvelope(t, envelope)
	if !bytes.Equal(encoded, want) {
		t.Fatalf("json.Marshal(envelope) = %s, want %s", encoded, want)
	}

	var topLevel map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &topLevel); err != nil {
		t.Fatalf("json.Unmarshal(envelope) error = %v", err)
	}
	if len(topLevel) != 8 {
		t.Fatalf("top-level field count = %d, want 8", len(topLevel))
	}
	for _, field := range []string{
		"subject",
		"schema_version",
		"event_id",
		"occurred_at",
		"organization_id",
		"actor_id",
		"correlation_id",
		"payload",
	} {
		if _, ok := topLevel[field]; !ok {
			t.Fatalf("top-level field %q is absent", field)
		}
	}
	for _, forbidden := range []string{"data", "envelope"} {
		if _, ok := topLevel[forbidden]; ok {
			t.Fatalf("unexpected payload wrapper %q", forbidden)
		}
	}
	if got := string(topLevel["actor_id"]); got != `"33333333-3333-4333-8333-333333333333"` {
		t.Fatalf("actor_id = %s", got)
	}
	if got := string(topLevel["payload"]); got != `{"status":"tentative","metadata":{"attempt":2,"source":"calendar"}}` {
		t.Fatalf("payload = %s", got)
	}
}

func TestEnvelopeV1_HasNoExportedMutableFields(t *testing.T) {
	envelopeType := reflect.TypeOf(events.EnvelopeV1{})
	for index := 0; index < envelopeType.NumField(); index++ {
		field := envelopeType.Field(index)
		if field.PkgPath == "" {
			t.Fatalf("EnvelopeV1 exposes field %q of type %s", field.Name, field.Type)
		}
	}
}

func TestNewEnvelopeV1_FreezesNestedMapSliceAndStructPointerPayloads(t *testing.T) {
	t.Run("nested map and slice", func(t *testing.T) {
		nestedObject := map[string]any{"count": 1}
		items := []any{"first", nestedObject}
		payload := map[string]any{
			"items": items,
			"state": "before",
		}
		envelope := mustNewValidEnvelope(t, payload)
		want := mustMarshalEnvelopeDirect(t, envelope)

		items[0] = "changed"
		nestedObject["count"] = 99
		payload["state"] = "after"
		payload["extra"] = true

		assertFrozenEnvelopeBytes(t, envelope, want)
	})

	t.Run("struct pointer", func(t *testing.T) {
		payload := &mutableStructPayload{
			Name: "before",
			Tags: []string{"one", "two"},
			Metadata: map[string]string{
				"source": "calendar",
			},
		}
		envelope := mustNewValidEnvelope(t, payload)
		want := mustMarshalEnvelopeDirect(t, envelope)

		payload.Name = "after"
		payload.Tags[0] = "changed"
		payload.Tags = append(payload.Tags, "three")
		payload.Metadata["source"] = "mutated"

		assertFrozenEnvelopeBytes(t, envelope, want)
	})
}

func TestNewEnvelopeV1_AcceptsAndFreezesCustomMarshalerObject(t *testing.T) {
	raw := []byte(` {"kind":"custom","items":[1,2]} `)
	payload := &customPayloadMarshaler{raw: raw}
	envelope := mustNewValidEnvelope(t, payload)
	want := mustMarshalEnvelopeDirect(t, envelope)
	if !bytes.Contains(want, []byte(`"payload":{"kind":"custom","items":[1,2]}`)) {
		t.Fatalf("custom object payload was not embedded directly: %s", want)
	}

	raw[2] = 'X'
	payload.raw = []byte(`null`)
	assertFrozenEnvelopeBytes(t, envelope, want)
}

func TestNewEnvelopeV1_NormalizesZeroOffsetOccurrenceToUTCZ(t *testing.T) {
	zeroOffsetLocation := time.FixedZone("FixedUTC0", 0)
	zeroOffsetOccurrence := time.Date(2030, time.May, 6, 7, 8, 9, 123456789, zeroOffsetLocation)
	envelope, err := events.NewEnvelopeV1(
		"calendar.meeting.rsvp.changed.v1",
		fixedEventID,
		zeroOffsetOccurrence,
		fixedOrganizationID,
		nil,
		"corr-utc0-04l",
		validPayload(),
	)
	if err != nil {
		t.Fatalf("NewEnvelopeV1() error = %v", err)
	}

	encoded := mustMarshalEnvelope(t, envelope)
	if !bytes.Contains(encoded, []byte(`"occurred_at":"2030-05-06T07:08:09.123456789Z"`)) {
		t.Fatalf("occurred_at was not normalized to UTC Z: %s", encoded)
	}
	if bytes.Contains(encoded, []byte(`+00:00`)) {
		t.Fatalf("occurred_at retained a numeric zero offset: %s", encoded)
	}
}

func TestNewEnvelopeV1_SubjectContract(t *testing.T) {
	validASCII255 := strings.Repeat("a", 252) + ".v1"
	validMultibyte255 := strings.Repeat("é", 124) + "😀.v1"
	valid := []string{
		"calendar.meeting.rsvp.changed.v1",
		"a.v1",
		validASCII255,
		validMultibyte255,
	}
	for _, subject := range valid {
		t.Run("valid-"+subjectName(subject), func(t *testing.T) {
			envelope, err := events.NewEnvelopeV1(
				subject,
				fixedEventID,
				fixedOccurredAt,
				fixedOrganizationID,
				nil,
				"corr-subject-04l",
				validPayload(),
			)
			if err != nil {
				t.Fatalf("NewEnvelopeV1() error = %v", err)
			}
			if got := envelope.Subject(); got != subject {
				t.Fatalf("Subject() = %q, want %q", got, subject)
			}
		})
	}

	invalidUTF8 := string([]byte{'c', 'a', 'l', 0xff, '.', 'v', '1'})
	invalid := []struct {
		name    string
		subject string
	}{
		{name: "empty", subject: ""},
		{name: "ascii-256-bytes", subject: strings.Repeat("a", 253) + ".v1"},
		{name: "multibyte-256-bytes", subject: strings.Repeat("é", 126) + "a.v1"},
		{name: "invalid-utf8", subject: invalidUTF8},
		{name: "leading-dot", subject: ".calendar.v1"},
		{name: "trailing-dot", subject: "calendar.v1."},
		{name: "double-dot", subject: "calendar..v1"},
		{name: "star-wildcard", subject: "calendar.*.v1"},
		{name: "greater-wildcard", subject: "calendar.>.v1"},
		{name: "ascii-whitespace", subject: "calendar meeting.v1"},
		{name: "unicode-whitespace", subject: "calendar.\u2003meeting.v1"},
		{name: "control-rune", subject: "calendar.\u0007meeting.v1"},
		{name: "missing-version", subject: "calendar.meeting.changed"},
		{name: "wrong-version-case", subject: "calendar.meeting.V1"},
		{name: "version-not-token", subject: "calendar.meetingv1"},
	}

	for _, testCase := range invalid {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := events.NewEnvelopeV1(
				testCase.subject,
				fixedEventID,
				fixedOccurredAt,
				fixedOrganizationID,
				nil,
				"corr-subject-04l",
				rejectingPayload(),
			)
			assertStableSafeError(t, err, events.ErrInvalidSubject)
		})
	}
}

func TestNewEnvelopeV1_RejectsInvalidIdentifiersOccurrenceAndCorrelation(t *testing.T) {
	nonZeroOffset := time.Date(2030, time.May, 6, 7, 8, 9, 0, time.FixedZone("UTC+01", 60*60))
	nonRFC3339Time := time.Date(10000, time.January, 1, 0, 0, 0, 0, time.UTC)
	zeroActor := uuid.Nil

	tests := []struct {
		name            string
		eventID         uuid.UUID
		occurredAt      time.Time
		organizationID  uuid.UUID
		actorID         *uuid.UUID
		correlationID   string
		want            error
	}{
		{
			name:           "zero event id",
			eventID:        uuid.Nil,
			occurredAt:     fixedOccurredAt,
			organizationID: fixedOrganizationID,
			correlationID:  "corr-valid-04l",
			want:           events.ErrInvalidEventID,
		},
		{
			name:           "zero occurrence",
			eventID:        fixedEventID,
			occurredAt:     time.Time{},
			organizationID: fixedOrganizationID,
			correlationID:  "corr-valid-04l",
			want:           events.ErrInvalidOccurredAt,
		},
		{
			name:           "nonzero occurrence offset",
			eventID:        fixedEventID,
			occurredAt:     nonZeroOffset,
			organizationID: fixedOrganizationID,
			correlationID:  "corr-valid-04l",
			want:           events.ErrInvalidOccurredAt,
		},
		{
			name:           "occurrence outside RFC3339 year range",
			eventID:        fixedEventID,
			occurredAt:     nonRFC3339Time,
			organizationID: fixedOrganizationID,
			correlationID:  "corr-valid-04l",
			want:           events.ErrInvalidOccurredAt,
		},
		{
			name:           "zero organization id",
			eventID:        fixedEventID,
			occurredAt:     fixedOccurredAt,
			organizationID: uuid.Nil,
			correlationID:  "corr-valid-04l",
			want:           events.ErrInvalidOrganizationID,
		},
		{
			name:           "zero actor id",
			eventID:        fixedEventID,
			occurredAt:     fixedOccurredAt,
			organizationID: fixedOrganizationID,
			actorID:        &zeroActor,
			correlationID:  "corr-valid-04l",
			want:           events.ErrInvalidActorID,
		},
		{
			name:           "empty correlation after trim",
			eventID:        fixedEventID,
			occurredAt:     fixedOccurredAt,
			organizationID: fixedOrganizationID,
			correlationID:  " \t\r\n ",
			want:           events.ErrInvalidCorrelationID,
		},
		{
			name:           "correlation over 128 bytes",
			eventID:        fixedEventID,
			occurredAt:     fixedOccurredAt,
			organizationID: fixedOrganizationID,
			correlationID:  strings.Repeat("c", 129),
			want:           events.ErrInvalidCorrelationID,
		},
		{
			name:           "multibyte correlation over 128 bytes",
			eventID:        fixedEventID,
			occurredAt:     fixedOccurredAt,
			organizationID: fixedOrganizationID,
			correlationID:  strings.Repeat("é", 65),
			want:           events.ErrInvalidCorrelationID,
		},
		{
			name:           "invalid utf8 correlation",
			eventID:        fixedEventID,
			occurredAt:     fixedOccurredAt,
			organizationID: fixedOrganizationID,
			correlationID:  string([]byte{'c', 0xff}),
			want:           events.ErrInvalidCorrelationID,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := events.NewEnvelopeV1(
				"calendar.meeting.rsvp.changed.v1",
				testCase.eventID,
				testCase.occurredAt,
				testCase.organizationID,
				testCase.actorID,
				testCase.correlationID,
				rejectingPayload(),
			)
			assertStableSafeError(t, err, testCase.want)
		})
	}
}

func TestNewEnvelopeV1_AcceptsTrimmedCorrelationBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "single byte", input: " x ", want: "x"},
		{name: "exactly 128 ascii bytes", input: strings.Repeat("c", 128), want: strings.Repeat("c", 128)},
		{name: "exactly 128 multibyte bytes", input: strings.Repeat("é", 64), want: strings.Repeat("é", 64)},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			envelope, err := events.NewEnvelopeV1(
				"calendar.meeting.rsvp.changed.v1",
				fixedEventID,
				fixedOccurredAt,
				fixedOrganizationID,
				nil,
				testCase.input,
				validPayload(),
			)
			if err != nil {
				t.Fatalf("NewEnvelopeV1() error = %v", err)
			}

			var decoded struct {
				CorrelationID string `json:"correlation_id"`
			}
			if err := json.Unmarshal(mustMarshalEnvelope(t, envelope), &decoded); err != nil {
				t.Fatalf("json.Unmarshal(envelope) error = %v", err)
			}
			if decoded.CorrelationID != testCase.want {
				t.Fatalf("correlation_id = %q, want %q", decoded.CorrelationID, testCase.want)
			}
		})
	}
}

func TestNewEnvelopeV1_RejectsEveryInvalidPayloadClassWithoutPanic(t *testing.T) {
	var typedNilPointer *envelopePayloadFixture
	var typedNilMap map[string]any
	var typedNilSlice []string
	var typedNilFunction func()
	var typedNilChannel chan int
	var typedNilInterface nilPayloadMarker = (*nilPayloadMarkerValue)(nil)
	var typedNilMarshaler json.Marshaler = (*customPayloadMarshaler)(nil)

	tests := []struct {
		name    string
		payload any
	}{
		{name: "nil interface", payload: nil},
		{name: "typed nil pointer", payload: typedNilPointer},
		{name: "typed nil map", payload: typedNilMap},
		{name: "typed nil slice", payload: typedNilSlice},
		{name: "typed nil function", payload: typedNilFunction},
		{name: "typed nil channel", payload: typedNilChannel},
		{name: "typed nil through interface", payload: typedNilInterface},
		{name: "typed nil json marshaler", payload: typedNilMarshaler},
		{name: "scalar string", payload: sensitiveBody},
		{name: "scalar number", payload: 42},
		{name: "array", payload: [1]string{sensitiveMIME}},
		{name: "slice", payload: []string{sensitiveMIME}},
		{name: "function", payload: func() {}},
		{name: "channel", payload: make(chan int)},
		{name: "unsupported complex", payload: complex(1, 2)},
		{name: "json null", payload: json.RawMessage(`null`)},
		{name: "custom array", payload: &customPayloadMarshaler{raw: []byte(`[1,2]`)}},
		{name: "custom scalar", payload: &customPayloadMarshaler{raw: []byte(`"` + sensitiveBody + `"`)}},
		{name: "custom null", payload: &customPayloadMarshaler{raw: []byte(`null`)}},
		{name: "custom invalid json", payload: &customPayloadMarshaler{raw: []byte(`{"broken":`)}},
		{name: "custom marshal error", payload: failingPayloadMarshaler{}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := newEnvelopePayloadErrorWithoutPanic(t, testCase.payload)
			assertStableSafeError(t, err, events.ErrInvalidPayload)
		})
	}
}

func mustNewValidEnvelope(t *testing.T, payload any) events.EnvelopeV1 {
	t.Helper()
	envelope, err := events.NewEnvelopeV1(
		"calendar.meeting.rsvp.changed.v1",
		fixedEventID,
		fixedOccurredAt,
		fixedOrganizationID,
		nil,
		"corr-freeze-04l",
		payload,
	)
	if err != nil {
		t.Fatalf("NewEnvelopeV1() error = %v", err)
	}
	return envelope
}

func newEnvelopePayloadErrorWithoutPanic(t *testing.T, payload any) (err error) {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("NewEnvelopeV1() panicked for payload type %T: %v", payload, recovered)
		}
	}()
	_, err = events.NewEnvelopeV1(
		"calendar.meeting.rsvp.changed.v1",
		fixedEventID,
		fixedOccurredAt,
		fixedOrganizationID,
		nil,
		"corr-payload-04l",
		payload,
	)
	return err
}

func assertFrozenEnvelopeBytes(t *testing.T, envelope events.EnvelopeV1, want []byte) {
	t.Helper()
	returned := mustMarshalEnvelopeDirect(t, envelope)
	if !bytes.Equal(returned, want) {
		t.Fatalf("MarshalJSON() after input mutation = %s, want frozen %s", returned, want)
	}
	if len(returned) == 0 {
		t.Fatal("MarshalJSON() returned empty bytes")
	}
	returned[0] ^= 0xff

	directAgain := mustMarshalEnvelopeDirect(t, envelope)
	viaJSON := mustMarshalEnvelope(t, envelope)
	if !bytes.Equal(directAgain, want) || !bytes.Equal(viaJSON, want) {
		t.Fatalf("caller mutation changed frozen bytes: direct=%s json=%s want=%s", directAgain, viaJSON, want)
	}
	if got := envelope.Subject(); got != "calendar.meeting.rsvp.changed.v1" {
		t.Fatalf("Subject() after input mutation = %q", got)
	}
	if got := envelope.EventID(); got != fixedEventID {
		t.Fatalf("EventID() after input mutation = %s", got)
	}
}

func mustMarshalEnvelope(t *testing.T, envelope events.EnvelopeV1) []byte {
	t.Helper()
	encoded, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("json.Marshal(envelope) error = %v", err)
	}
	return encoded
}

func mustMarshalEnvelopeDirect(t *testing.T, envelope events.EnvelopeV1) []byte {
	t.Helper()
	encoded, err := envelope.MarshalJSON()
	if err != nil {
		t.Fatalf("envelope.MarshalJSON() error = %v", err)
	}
	return encoded
}

func validPayload() envelopePayloadFixture {
	return envelopePayloadFixture{
		NewStatus:  "accepted",
		AttendeeID: fixedAttendeeID,
	}
}

func rejectingPayload() rejectingPayloadFixture {
	return rejectingPayloadFixture{
		Body:           sensitiveBody,
		MIME:           sensitiveMIME,
		Credential:     sensitiveCredential,
		IdempotencyKey: sensitiveIdempotencyKey,
		AttendeeID:     fixedAttendeeID,
	}
}

func assertStableSafeError(t *testing.T, got, want error) {
	t.Helper()
	if got == nil {
		t.Fatalf("NewEnvelopeV1() error = nil, want %v", want)
	}
	if !errors.Is(got, want) {
		t.Fatalf("NewEnvelopeV1() error = %v, want %v", got, want)
	}
	if got.Error() != want.Error() {
		t.Fatalf("NewEnvelopeV1() error text = %q, want stable %q", got.Error(), want.Error())
	}
	for _, secret := range []string{
		sensitiveBody,
		sensitiveMIME,
		sensitiveCredential,
		sensitiveIdempotencyKey,
		"SENSITIVE-MARSHAL-ERROR-04L",
		fixedAttendeeID.String(),
	} {
		if strings.Contains(got.Error(), secret) {
			t.Fatalf("validation error leaked rejected input %q", secret)
		}
	}
}

func subjectName(subject string) string {
	if len(subject) <= 48 {
		return subject
	}
	return subject[:16] + "-bytes-" + strings.Repeat("x", len(subject)%17)
}
