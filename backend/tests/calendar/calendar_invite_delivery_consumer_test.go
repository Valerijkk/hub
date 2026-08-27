package calendar_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCalendarInviteDeliveryConsumerSourceContract(t *testing.T) {
	source := readCalendarInviteDeliveryConsumerSource(t)
	compactSource := strings.NewReplacer(" ", "", "\t", "").Replace(source)

	requireCalendarInviteDeliveryConsumerContains(t, source,
		"CalendarInviteCreatedSubjectV2",
		"CalendarInviteUpdatedSubjectV2",
		"CalendarInviteCancelledSubjectV2",
		"calendar-mail-outbound-invite-v2",
		"FilterSubjects:",
		"jetstream.AckExplicitPolicy",
		"envelope.EventID()",
		"payload.OrganizerID",
		"EventSequence: int(payload.EventSequence)",
		"recipient.AttendeeID",
		"recipient.InviteID",
		"calendarInviteDeliveryConsumerMaximumDeliveries  = 200",
		"calendarInviteDeliveryConsumerAdmissionBatchSize = 50",
		"len(payload.Deliveries) > calendarInviteDeliveryConsumerMaximumDeliveries",
		"DecodeCalendarInviteDeliveryEnvelopeV2(envelope)",
		"value.Subject != subject",
		"bytes.Equal(encoded, canonical)",
		"decoder.DisallowUnknownFields()",
		"CalendarMailDeliveryRepositoryErrorInvalid",
		"CalendarMailDeliveryRepositoryErrorConflict",
		"calendarInviteDeliveryConsumerTerm(message)",
		"calendarInviteDeliveryConsumerNak(message)",
	)
	requireCalendarInviteDeliveryConsumerContains(t, compactSource,
		"calendarInviteDeliveryConsumerPullWindow=8",
		"calendarInviteDeliveryConsumerWorkerConcurrency=4",
		"calendarInviteDeliveryConsumerMaxAckPending=16",
		"calendarInviteDeliveryConsumerAckWait=90*time.Second",
		"calendarInviteDeliveryConsumerDeliveryTimeout=45*time.Second",
		"calendarInviteDeliveryConsumerNakDelay=65*time.Second",
		"SourceEnvelopeID:envelope.EventID()",
		"OrganizerID:payload.OrganizerID",
	)
	if strings.Contains(compactSource, "OrganizerID:value.ActorID") {
		t.Fatal("consumer must derive organizer from payload, never the envelope actor")
	}

	if strings.Contains(source, "MaxDeliver:") {
		t.Fatal("consumer must keep JetStream redelivery unlimited")
	}
	for _, legacySubject := range []string{
		"calendar.invite.created.v1",
		"calendar.invite.updated.v1",
		"calendar.invite.cancelled.v1",
	} {
		if strings.Contains(source, legacySubject) {
			t.Fatalf("consumer must not admit legacy v1 subject: %s", legacySubject)
		}
	}
}

func TestCalendarInviteDeliveryConsumerSourceOrderedAdmission(t *testing.T) {
	source := readCalendarInviteDeliveryConsumerSource(t)

	requireCalendarInviteDeliveryOrder(t, source,
		"if len(admissions) == 0 {",
		"calendarInviteDeliveryConsumerAck(message)",
		"for start := 0; start < len(admissions); start += calendarInviteDeliveryConsumerAdmissionBatchSize {",
		"if err := consumer.repository.AdmitCalendarInviteDeliveries(deliveryContext, mail.CalendarInviteDeliveryAdmissionRequest{Invites: admissions[start:end]}); err != nil {",
		"calendarInviteDeliveryConsumerAck(message)",
	)
}

func readCalendarInviteDeliveryConsumerSource(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("determine test file path")
	}
	source, err := os.ReadFile(filepath.Join(filepath.Dir(file), "..", "..", "internal", "calendar", "calendar_invite_delivery_consumer.go"))
	if err != nil {
		t.Fatalf("read consumer source: %v", err)
	}
	return string(source)
}

func requireCalendarInviteDeliveryConsumerContains(t *testing.T, source string, needles ...string) {
	t.Helper()
	for _, needle := range needles {
		if !strings.Contains(source, needle) {
			t.Fatalf("consumer source missing %q", needle)
		}
	}
}

func requireCalendarInviteDeliveryOrder(t *testing.T, source string, needles ...string) {
	t.Helper()
	offset := 0
	for _, needle := range needles {
		next := strings.Index(source[offset:], needle)
		if next < 0 {
			t.Fatalf("consumer source missing ordered token %q", needle)
		}
		offset += next + len(needle)
	}
}
