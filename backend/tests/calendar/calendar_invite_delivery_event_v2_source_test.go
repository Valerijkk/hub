package calendar_test

import (
	"os"
	"strings"
	"testing"
)

func calendarInviteDeliveryEventV2Source(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile("../../internal/calendar/calendar_invite_delivery_event_v2.go")
	if err != nil {
		t.Fatalf("read Calendar invite delivery v2 source: %v", err)
	}
	return string(source)
}

func requireCalendarInviteDeliveryEventV2Source(t *testing.T, source string, terms ...string) {
	t.Helper()
	for _, term := range terms {
		if !strings.Contains(source, term) {
			t.Fatalf("source is missing %q", term)
		}
	}
}

func TestCalendarInviteDeliveryEventV2SourceContract(t *testing.T) {
	source := calendarInviteDeliveryEventV2Source(t)
	requireCalendarInviteDeliveryEventV2Source(t, source,
		"calendar.invite.created.v2",
		"calendar.invite.updated.v2",
		"calendar.invite.cancelled.v2",
		"type CalendarInviteDeliveryEnvelopeV2Input struct",
		"type CalendarInviteDeliveryV2Payload struct",
		"type CalendarInviteDeliveryV2Recipient struct",
		`json:"event_id"`,
		`json:"event_sequence"`,
		`json:"organizer_id"`,
		`json:"recurrence_id,omitempty"`,
		`json:"deliveries"`,
		`json:"attendee_id"`,
		`json:"invite_id"`,
		"hub.calendar.invite-delivery.id.v1",
		"uuid.NewSHA1(uuid.NameSpaceURL",
		`sourceEnvelopeID.String()+"/"+attendeeID.String()`,
		"func NewCalendarInviteDeliveryEnvelopeV2(",
		"events.NewEnvelopeV1(",
		"func DecodeCalendarInviteDeliveryEnvelopeV2(",
		"decoder.DisallowUnknownFields()",
		"decoder.Decode(&struct{}{}) != io.EOF",
		"calendarInviteDeliveryV2SubjectMethod",
		"validCalendarEventOccurrenceCancelCanonicalKey",
		"len(payload.Deliveries) > calendarCreateMaxAttendees",
		"sort.Slice(deliveries",
		"bytes.Compare(payload.Deliveries[index-1].InviteID[:], delivery.InviteID[:]) >= 0",
		"expectedInviteID != delivery.InviteID",
		"!bytes.Equal(encoded, canonicalJSON)",
		"SchemaVersion != 1",
	)
	for _, forbidden := range []string{"time.Now", "net/http", "http.", "github.com/nats-io/nats.go"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("v2 delivery source contains forbidden dependency %q", forbidden)
		}
	}
	quoted := string(rune(34))
	for _, field := range []string{"uid", "method", "title", "description", "body", "content", "email", "address", "provider"} {
		if strings.Contains(source, "json:"+quoted+field+quoted) {
			t.Fatalf("v2 delivery source contains forbidden payload field %q", field)
		}
	}
}
