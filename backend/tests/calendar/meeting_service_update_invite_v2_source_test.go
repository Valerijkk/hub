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
	"github.com/valerii/hub/backend/internal/platform/events"
)

func TestMeetingServiceUpdateInviteV2SourceAndEnvelopeContract(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	sourcePath := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", "..", "internal", "calendar", "meeting_service_update.go"))
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read meeting update service source: %v", err)
	}
	text := string(source)
	for _, required := range []string{
		"CalendarInviteUpdatedSubjectV2",
		"CalendarInviteCancelledSubjectV2",
		"NewCalendarInviteDeliveryEnvelopeV2(",
		"sourceEnvelopeID := ids[1]",
		"OrganizerID: current.OwnerID",
		"EventSequence: persisted.Sequence()",
		"RecurrenceID: nil",
		"AttendeeIDs: activeAttendeeIDs",
		"AttendeeIDs: removedIDs",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("meeting update invite V2 source missing %q", required)
		}
	}
	if strings.Count(text, "NewCalendarInviteDeliveryEnvelopeV2(") != 3 {
		t.Fatalf("meeting update invite V2 builder calls = %d, want three", strings.Count(text, "NewCalendarInviteDeliveryEnvelopeV2("))
	}
	if strings.Count(text, "if len(activeAttendeeIDs) != 0 {") != 2 || strings.Count(text, "if len(removedIDs) != 0 {") != 1 {
		t.Fatal("meeting update/cancel V2 envelopes must be guarded for empty recipient partitions")
	}
	for _, forbidden := range []string{
		"calendar.invite.updated.v1",
		"calendar.invite.cancelled.v1",
		"NewCalendarInviteUpdatedEnvelopeV1",
		"NewCalendarInviteCancelledEnvelopeV1",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("meeting update invite V2 source retains %q", forbidden)
		}
	}

	updatedSourceID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	removedCancelSourceID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	fullCancelSourceID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	eventID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	organizationID := uuid.MustParse("55555555-5555-4555-8555-555555555555")
	organizerID := uuid.MustParse("66666666-6666-4666-8666-666666666666")
	activeAttendeeIDs := []uuid.UUID{
		uuid.MustParse("77777777-7777-4777-8777-777777777777"),
		uuid.MustParse("88888888-8888-4888-8888-888888888888"),
	}
	removedAttendeeIDs := []uuid.UUID{uuid.MustParse("99999999-9999-4999-8999-999999999999")}
	occurredAt := time.Date(2031, time.February, 3, 4, 5, 6, 0, time.UTC)
	updated, err := calendar.NewCalendarInviteDeliveryEnvelopeV2(
		calendar.CalendarInviteUpdatedSubjectV2, updatedSourceID, occurredAt, organizationID, &organizerID, eventID.String(),
		calendar.CalendarInviteDeliveryEnvelopeV2Input{EventID: eventID, EventSequence: 0, OrganizerID: organizerID, AttendeeIDs: activeAttendeeIDs},
	)
	if err != nil {
		t.Fatalf("build updated V2 envelope: %v", err)
	}
	removedCancelled, err := calendar.NewCalendarInviteDeliveryEnvelopeV2(
		calendar.CalendarInviteCancelledSubjectV2, removedCancelSourceID, occurredAt, organizationID, &organizerID, eventID.String(),
		calendar.CalendarInviteDeliveryEnvelopeV2Input{EventID: eventID, EventSequence: 0, OrganizerID: organizerID, AttendeeIDs: removedAttendeeIDs},
	)
	if err != nil {
		t.Fatalf("build removed-attendee cancelled V2 envelope: %v", err)
	}
	fullCancelled, err := calendar.NewCalendarInviteDeliveryEnvelopeV2(
		calendar.CalendarInviteCancelledSubjectV2, fullCancelSourceID, occurredAt, organizationID, &organizerID, eventID.String(),
		calendar.CalendarInviteDeliveryEnvelopeV2Input{EventID: eventID, EventSequence: 0, OrganizerID: organizerID, AttendeeIDs: activeAttendeeIDs},
	)
	if err != nil {
		t.Fatalf("build full-cancelled V2 envelope: %v", err)
	}
	if updated.EventID() != updatedSourceID || removedCancelled.EventID() != removedCancelSourceID || fullCancelled.EventID() != fullCancelSourceID ||
		updated.EventID() == removedCancelled.EventID() || updated.EventID() == fullCancelled.EventID() || removedCancelled.EventID() == fullCancelled.EventID() {
		t.Fatal("meeting update V2 envelopes must retain three distinct source IDs")
	}
	assertMeetingInviteV2Payload(t, updated, calendar.CalendarInviteUpdatedSubjectV2, eventID, organizerID, activeAttendeeIDs)
	assertMeetingInviteV2Payload(t, removedCancelled, calendar.CalendarInviteCancelledSubjectV2, eventID, organizerID, removedAttendeeIDs)
	assertMeetingInviteV2Payload(t, fullCancelled, calendar.CalendarInviteCancelledSubjectV2, eventID, organizerID, activeAttendeeIDs)
	for _, activeID := range activeAttendeeIDs {
		for _, removedID := range removedAttendeeIDs {
			if activeID == removedID {
				t.Fatal("active and removed meeting attendee recipient partitions overlap")
			}
		}
	}
}

func assertMeetingInviteV2Payload(t *testing.T, envelope events.EnvelopeV1, wantSubject string, wantEventID, wantOrganizerID uuid.UUID, wantAttendeeIDs []uuid.UUID) {
	t.Helper()
	if envelope.Subject() != wantSubject {
		t.Fatalf("V2 subject = %q, want %q", envelope.Subject(), wantSubject)
	}
	payload, err := calendar.DecodeCalendarInviteDeliveryEnvelopeV2(envelope)
	if err != nil {
		t.Fatalf("decode V2 payload: %v", err)
	}
	if payload.EventID != wantEventID || payload.EventSequence != 0 || payload.OrganizerID != wantOrganizerID || payload.RecurrenceID != nil || len(payload.Deliveries) != len(wantAttendeeIDs) {
		t.Fatalf("V2 payload = %#v", payload)
	}
	namespace := uuid.NewSHA1(uuid.NameSpaceURL, []byte("hub.calendar.invite-delivery.id.v1"))
	want := make(map[uuid.UUID]uuid.UUID, len(wantAttendeeIDs))
	for _, attendeeID := range wantAttendeeIDs {
		want[attendeeID] = uuid.NewSHA1(namespace, []byte(envelope.EventID().String()+"/"+attendeeID.String()))
	}
	for _, delivery := range payload.Deliveries {
		if inviteID, found := want[delivery.AttendeeID]; !found || delivery.InviteID != inviteID {
			t.Fatalf("deterministic meeting update delivery = %#v", delivery)
		}
	}
}
