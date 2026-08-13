package calendar_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/valerii/hub/backend/internal/calendar"
)

func TestSerializeITIP_RequestLiteralOracle(t *testing.T) {
	message := timedITIPMessage(calendar.ITIPMethodRequest)

	got, err := calendar.SerializeITIP(message)
	if err != nil {
		t.Fatalf("SerializeITIP() error = %v", err)
	}

	want := crlf(
		"BEGIN:VCALENDAR",
		"PRODID:-//Hub//Calendar//EN",
		"VERSION:2.0",
		"CALSCALE:GREGORIAN",
		"METHOD:REQUEST",
		"BEGIN:VEVENT",
		"UID:event-123@hub.example",
		"DTSTAMP:20260301T120000Z",
		"DTSTART;TZID=America/New_York:20260308T013000",
		"DTEND;TZID=America/New_York:20260308T033000",
		"SEQUENCE:4",
		"STATUS:CONFIRMED",
		"SUMMARY:DST planning",
		`DESCRIPTION:Plan\\review\; phase\, one\nLine two`,
		"LOCATION:Conference room",
		`ORGANIZER;CN="Calendar Owner":mailto:owner@example.test`,
		`ATTENDEE;ROLE=REQ-PARTICIPANT;PARTSTAT=NEEDS-ACTION;RSVP=TRUE:mailto:a@x.io`,
		`ATTENDEE;ROLE=OPT-PARTICIPANT;PARTSTAT=TENTATIVE;RSVP=FALSE:mailto:b@x.io`,
		"RRULE:FREQ=WEEKLY;BYDAY=SU",
		"EXDATE;TZID=America/New_York:20260315T023000",
		"END:VEVENT",
		"END:VCALENDAR",
	)
	if string(got) != want {
		t.Fatalf("SerializeITIP() = %q, want %q", got, want)
	}
}

func TestSerializeITIP_CancelCarriesSequenceAndCancelledStatus(t *testing.T) {
	message := timedITIPMessage(calendar.ITIPMethodCancel)
	message.Sequence = 5
	message.Status = calendar.ITIPStatusCancelled

	got, err := calendar.SerializeITIP(message)
	if err != nil {
		t.Fatalf("SerializeITIP() error = %v", err)
	}

	payload := string(got)
	for _, property := range []string{
		"METHOD:CANCEL\r\n",
		"SEQUENCE:5\r\n",
		"STATUS:CANCELLED\r\n",
	} {
		if !strings.Contains(payload, property) {
			t.Fatalf("SerializeITIP() missing %q in %q", property, payload)
		}
	}
}

func TestSerializeITIP_ReplyHasExactlyOneRespondingAttendee(t *testing.T) {
	message := timedITIPMessage(calendar.ITIPMethodReply)
	message.Attendees = []calendar.ITIPParticipant{{
		Email:  "a@x.io",
		Role:        calendar.ITIPParticipantRequired,
		Status:      calendar.RSVPAccepted,
		RSVP:        false,
	}}

	got, err := calendar.SerializeITIP(message)
	if err != nil {
		t.Fatalf("SerializeITIP() error = %v", err)
	}

	payload := string(got)
	if strings.Count(payload, "ATTENDEE;") != 1 {
		t.Fatalf("SerializeITIP() attendee count = %d, want 1 in %q", strings.Count(payload, "ATTENDEE;"), payload)
	}
	for _, property := range []string{
		"METHOD:REPLY\r\n",
		"UID:event-123@hub.example\r\n",
		"DTSTART;TZID=America/New_York:20260308T013000\r\n",
		`ORGANIZER;CN="Calendar Owner":mailto:owner@example.test` + "\r\n",
		"SUMMARY:DST planning\r\n",
		`ATTENDEE;ROLE=REQ-PARTICIPANT;PARTSTAT=ACCEPTED;RSVP=FALSE:mailto:a@x.io` + "\r\n",
	} {
		if !strings.Contains(payload, property) {
			t.Fatalf("SerializeITIP() missing %q in %q", property, payload)
		}
	}
}

func TestSerializeITIP_RetainsUTF8Text(t *testing.T) {
	message := timedITIPMessage(calendar.ITIPMethodRequest)
	message.Summary = "Встреча с командой"

	got, err := calendar.SerializeITIP(message)
	if err != nil {
		t.Fatalf("SerializeITIP() error = %v", err)
	}

	if !strings.Contains(string(got), "SUMMARY:Встреча с командой\r\n") {
		t.Fatalf("SerializeITIP() did not retain UTF-8 summary: %q", got)
	}
}

func TestSerializeITIP_AllDayUsesDateValuesAndExclusiveEnd(t *testing.T) {
	message := timedITIPMessage(calendar.ITIPMethodRequest)
	message.Temporal = calendar.ITIPTemporal{AllDay: &calendar.ITIPAllDay{
		StartDate:        "2026-08-20",
		EndDateExclusive: "2026-08-22",
		Timezone:         "Europe/Moscow",
	}}
	message.EXDates = []calendar.ITIPRecurrenceID{{Date: "2026-08-21"}}

	got, err := calendar.SerializeITIP(message)
	if err != nil {
		t.Fatalf("SerializeITIP() error = %v", err)
	}

	payload := string(got)
	for _, property := range []string{
		"DTSTART;VALUE=DATE:20260820\r\n",
		"DTEND;VALUE=DATE:20260822\r\n",
		"EXDATE;VALUE=DATE:20260821\r\n",
	} {
		if !strings.Contains(payload, property) {
			t.Fatalf("SerializeITIP() missing %q in %q", property, payload)
		}
	}
	if strings.Contains(payload, "TZID=") || strings.Contains(payload, "T000000") {
		t.Fatalf("SerializeITIP() inferred timed all-day value in %q", payload)
	}
}

func TestSerializeITIP_OccurrenceCarriesCanonicalRecurrenceID(t *testing.T) {
	message := timedITIPMessage(calendar.ITIPMethodCancel)
	message.Status = calendar.ITIPStatusCancelled
	message.RecurrenceID = &calendar.ITIPRecurrenceID{Time: time.Date(2026, time.March, 15, 6, 30, 0, 0, time.UTC)}

	got, err := calendar.SerializeITIP(message)
	if err != nil {
		t.Fatalf("SerializeITIP() error = %v", err)
	}

	if !strings.Contains(string(got), "RECURRENCE-ID;TZID=America/New_York:20260315T023000\r\n") {
		t.Fatalf("SerializeITIP() = %q, want timed RECURRENCE-ID", got)
	}
}

func TestSerializeITIP_FoldsAtOctetBoundaryWithoutSplittingUTF8(t *testing.T) {
	message := timedITIPMessage(calendar.ITIPMethodRequest)
	message.Summary = strings.Repeat("Ж", 50)

	got, err := calendar.SerializeITIP(message)
	if err != nil {
		t.Fatalf("SerializeITIP() error = %v", err)
	}

	payload := string(got)
	if !strings.HasSuffix(payload, "\r\n") {
		t.Fatalf("SerializeITIP() must end in CRLF: %q", payload)
	}
	withoutCRLF := strings.ReplaceAll(payload, "\r\n", "")
	if strings.ContainsAny(withoutCRLF, "\r\n") {
		t.Fatalf("SerializeITIP() has non-CRLF line ending: %q", payload)
	}
	if !strings.Contains(payload, "SUMMARY:") || !strings.Contains(payload, "\r\n ") {
		t.Fatalf("SerializeITIP() did not fold a long SUMMARY: %q", payload)
	}
	for _, line := range strings.Split(strings.TrimSuffix(payload, "\r\n"), "\r\n") {
		if len([]byte(line)) > 75 {
			t.Fatalf("physical content line is %d octets, want at most 75: %q", len([]byte(line)), line)
		}
		if !utf8.ValidString(line) {
			t.Fatalf("physical content line splits UTF-8: %q", line)
		}
	}
}

func TestSerializeITIP_ValidationErrorsAreStableSentinels(t *testing.T) {
	valid := timedITIPMessage(calendar.ITIPMethodRequest)
	invalidTemporal := valid
	invalidTemporal.Temporal.AllDay = &calendar.ITIPAllDay{
		StartDate:        "2026-08-20",
		EndDateExclusive: "2026-08-21",
		Timezone:         "Europe/Moscow",
	}
	invalidReply := valid
	invalidReply.Method = calendar.ITIPMethodReply
	invalidReply.Attendees = nil
	invalidTimezone := valid
	invalidTimezone.Temporal.Timed = &calendar.ITIPTimed{
		StartsAt: valid.Temporal.Timed.StartsAt,
		EndsAt:   valid.Temporal.Timed.EndsAt,
		Timezone: "Mars/Olympus_Mons",
	}

	tests := []struct {
		name    string
		message calendar.ITIPMessage
		wantErr error
	}{
		{
			name:    "missing stamp",
			message: calendar.ITIPMessage{},
			wantErr: calendar.ErrInvalidITIPStamp,
		},
		{
			name:    "negative sequence",
			message: withSequence(valid, -1),
			wantErr: calendar.ErrInvalidITIPSequence,
		},
		{
			name:    "missing uid",
			message: withUID(valid, ""),
			wantErr: calendar.ErrInvalidITIPMessage,
		},
		{
			name:    "invalid temporal union",
			message: invalidTemporal,
			wantErr: calendar.ErrInvalidITIPTemporal,
		},
		{
			name:    "invalid reply attendee combination",
			message: invalidReply,
			wantErr: calendar.ErrInvalidITIPParticipant,
		},
		{
			name:    "invalid organizer email",
			message: withOrganizerEmail(valid, "not-an-email"),
			wantErr: calendar.ErrInvalidITIPEmail,
		},
		{
			name:    "invalid timezone",
			message: invalidTimezone,
			wantErr: calendar.ErrInvalidTimezone,
		},
		{
			name:    "cancel requires cancelled status",
			message: timedITIPMessage(calendar.ITIPMethodCancel),
			wantErr: calendar.ErrInvalidITIPStatus,
		},
		{
			name:    "series request cannot carry recurrence id",
			message: withRecurrenceID(valid, &calendar.ITIPRecurrenceID{Time: time.Date(2026, time.March, 15, 6, 30, 0, 0, time.UTC)}),
			wantErr: calendar.ErrInvalidITIPRecurrence,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := calendar.SerializeITIP(tc.message)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("SerializeITIP() error = %v, want errors.Is(..., %v)", err, tc.wantErr)
			}
		})
	}
}

func TestSerializeITIP_IsByteIdenticalForTheSameValue(t *testing.T) {
	message := timedITIPMessage(calendar.ITIPMethodRequest)

	first, err := calendar.SerializeITIP(message)
	if err != nil {
		t.Fatalf("first SerializeITIP() error = %v", err)
	}
	second, err := calendar.SerializeITIP(message)
	if err != nil {
		t.Fatalf("second SerializeITIP() error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("SerializeITIP() was not byte-identical: first %q, second %q", first, second)
	}
}

func timedITIPMessage(method calendar.ITIPMethod) calendar.ITIPMessage {
	return calendar.ITIPMessage{
		Method:      method,
		UID:         "event-123@hub.example",
		Stamp:       time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC),
		Sequence:    4,
		Status:      calendar.ITIPStatusConfirmed,
		Summary:     "DST planning",
		Description: "Plan\\review; phase, one\nLine two",
		Location:    "Conference room",
		Temporal: calendar.ITIPTemporal{Timed: &calendar.ITIPTimed{
			StartsAt: time.Date(2026, time.March, 8, 6, 30, 0, 0, time.UTC),
			EndsAt:   time.Date(2026, time.March, 8, 7, 30, 0, 0, time.UTC),
			Timezone: "America/New_York",
		}},
		Organizer: calendar.ITIPOrganizer{
			Email:       "owner@example.test",
			DisplayName: "Calendar Owner",
		},
		Attendees: []calendar.ITIPParticipant{
			{
				Email:  "a@x.io",
				Role:        calendar.ITIPParticipantRequired,
				Status:      calendar.RSVPNeedsAction,
				RSVP:        true,
			},
			{
				Email:  "b@x.io",
				Role:        calendar.ITIPParticipantOptional,
				Status:      calendar.RSVPTentative,
				RSVP:        false,
			},
		},
		RRULE: "FREQ=WEEKLY;BYDAY=SU",
		EXDates: []calendar.ITIPRecurrenceID{{
			Time: time.Date(2026, time.March, 15, 6, 30, 0, 0, time.UTC),
		}},
	}
}

func withSequence(message calendar.ITIPMessage, sequence int) calendar.ITIPMessage {
	message.Sequence = sequence
	return message
}

func withUID(message calendar.ITIPMessage, uid string) calendar.ITIPMessage {
	message.UID = uid
	return message
}

func withOrganizerEmail(message calendar.ITIPMessage, email string) calendar.ITIPMessage {
	message.Organizer.Email = email
	return message
}

func withRecurrenceID(message calendar.ITIPMessage, recurrenceID *calendar.ITIPRecurrenceID) calendar.ITIPMessage {
	message.RecurrenceID = recurrenceID
	return message
}

func crlf(lines ...string) string {
	return strings.Join(lines, "\r\n") + "\r\n"
}
