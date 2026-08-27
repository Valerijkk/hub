package calendar_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/calendar"
)

func TestMeetingServiceInviteV2SourceAndEnvelopeContract(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	sourcePath := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", "..", "internal", "calendar", "meeting_service.go"))
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read meeting service source: %v", err)
	}
	text := string(source)
	for _, required := range []string{
		"NewCalendarInviteDeliveryEnvelopeV2(",
		"CalendarInviteCreatedSubjectV2",
		"sourceEnvelopeID",
		"eventSnapshot.OrganizerID",
		"eventSnapshot.Sequence",
		"if len(attendeeRecordIDs) != 0 {",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("meeting invite V2 source missing %q", required)
		}
	}
	if strings.Count(text, "NewCalendarInviteDeliveryEnvelopeV2(") != 1 {
		t.Fatalf("meeting invite V2 builder calls = %d, want one", strings.Count(text, "NewCalendarInviteDeliveryEnvelopeV2("))
	}
	inviteGuard := strings.Index(text, "if len(attendeeRecordIDs) != 0 {")
	inviteBuilder := strings.Index(text, "NewCalendarInviteDeliveryEnvelopeV2(")
	if inviteGuard < 0 || inviteBuilder < inviteGuard {
		t.Fatal("meeting invite V2 must be guarded so zero attendees emit no invite outbox event")
	}
	for _, forbidden := range []string{"calendar.invite.created.v1", "calendarCreateInvitePayload"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("meeting invite V2 source retains %q", forbidden)
		}
	}

	sourceEnvelopeID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	eventID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	organizationID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	organizerID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	attendeeIDs := []uuid.UUID{
		uuid.MustParse("55555555-5555-4555-8555-555555555555"),
		uuid.MustParse("66666666-6666-4666-8666-666666666666"),
	}
	envelope, err := calendar.NewCalendarInviteDeliveryEnvelopeV2(
		calendar.CalendarInviteCreatedSubjectV2, sourceEnvelopeID, time.Date(2031, 2, 3, 4, 5, 6, 0, time.UTC), organizationID, &organizerID, eventID.String(),
		calendar.CalendarInviteDeliveryEnvelopeV2Input{EventID: eventID, EventSequence: 0, OrganizerID: organizerID, AttendeeIDs: attendeeIDs},
	)
	if err != nil {
		t.Fatalf("NewCalendarInviteDeliveryEnvelopeV2() error = %v", err)
	}
	if envelope.EventID() != sourceEnvelopeID || envelope.Subject() != calendar.CalendarInviteCreatedSubjectV2 {
		t.Fatalf("meeting source envelope = subject:%q id:%s", envelope.Subject(), envelope.EventID())
	}
	payload, err := calendar.DecodeCalendarInviteDeliveryEnvelopeV2(envelope)
	if err != nil {
		t.Fatalf("DecodeCalendarInviteDeliveryEnvelopeV2() error = %v", err)
	}
	if payload.EventID != eventID || payload.EventSequence != 0 || payload.OrganizerID != organizerID || payload.RecurrenceID != nil || len(payload.Deliveries) != len(attendeeIDs) {
		t.Fatalf("meeting v2 payload = %#v", payload)
	}
	namespace := uuid.NewSHA1(uuid.NameSpaceURL, []byte("hub.calendar.invite-delivery.id.v1"))
	wantInvites := make(map[uuid.UUID]uuid.UUID, len(attendeeIDs))
	for _, attendeeID := range attendeeIDs {
		wantInvites[attendeeID] = uuid.NewSHA1(namespace, []byte(sourceEnvelopeID.String()+"/"+attendeeID.String()))
	}
	for _, delivery := range payload.Deliveries {
		if want, found := wantInvites[delivery.AttendeeID]; !found || delivery.InviteID != want {
			t.Fatalf("deterministic meeting invite delivery = %#v", delivery)
		}
	}
}
