package calendar_test

import (
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/valerii/hub/backend/internal/calendar"
)

func TestDefaultITIPParseLimits_FrozenValues(t *testing.T) {
	want := calendar.ITIPParseLimits{
		MaxBytes:             1048576,
		MaxUnfoldedLineBytes: 8192,
		MaxProperties:        4096,
		MaxAttendees:         200,
		MaxExDates:           1000,
	}
	if got := calendar.DefaultITIPParseLimits(); got != want {
		t.Fatalf("DefaultITIPParseLimits() = %#v, want %#v", got, want)
	}
}

func TestParseITIP_RejectsInvalidLimitsBeforePayload(t *testing.T) {
	defaults := calendar.DefaultITIPParseLimits()
	cases := map[string]calendar.ITIPParseLimits{
		"zero bytes":          withLimit(defaults, func(value *calendar.ITIPParseLimits) { value.MaxBytes = 0 }),
		"negative line":       withLimit(defaults, func(value *calendar.ITIPParseLimits) { value.MaxUnfoldedLineBytes = -1 }),
		"zero properties":     withLimit(defaults, func(value *calendar.ITIPParseLimits) { value.MaxProperties = 0 }),
		"zero attendees":      withLimit(defaults, func(value *calendar.ITIPParseLimits) { value.MaxAttendees = 0 }),
		"zero exclusions":     withLimit(defaults, func(value *calendar.ITIPParseLimits) { value.MaxExDates = 0 }),
		"bytes above hard cap": withLimit(defaults, func(value *calendar.ITIPParseLimits) { value.MaxBytes++ }),
		"line above hard cap":  withLimit(defaults, func(value *calendar.ITIPParseLimits) { value.MaxUnfoldedLineBytes++ }),
		"props above hard cap": withLimit(defaults, func(value *calendar.ITIPParseLimits) { value.MaxProperties++ }),
		"people above hard cap": withLimit(defaults, func(value *calendar.ITIPParseLimits) {
			value.MaxAttendees++
		}),
		"exdates above hard cap": withLimit(defaults, func(value *calendar.ITIPParseLimits) {
			value.MaxExDates++
		}),
	}

	for name, limits := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := calendar.ParseITIP([]byte("private-uid-and-body"), limits)
			if !errors.Is(err, calendar.ErrInvalidITIPLimits) {
				t.Fatalf("ParseITIP() error = %v, want ErrInvalidITIPLimits", err)
			}
			assertSafeError(t, err, "private-uid-and-body")
		})
	}
}

func TestParseITIP_AcceptsFrozenRequestCancelReplyFixtures(t *testing.T) {
	cases := []struct {
		name          string
		wire          string
		method        calendar.ITIPMethod
		status        calendar.ITIPStatus
		attendeeCount int
	}{
		{name: "request", wire: timedRequestWire(), method: calendar.ITIPMethodRequest, status: calendar.ITIPStatusConfirmed, attendeeCount: 2},
		{name: "cancel", wire: cancelWire(), method: calendar.ITIPMethodCancel, status: calendar.ITIPStatusCancelled, attendeeCount: 0},
		{name: "reply", wire: replyWire(), method: calendar.ITIPMethodReply, status: calendar.ITIPStatusTentative, attendeeCount: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			message := mustParse(t, tc.wire)
			if message.Method != tc.method || message.Status != tc.status {
				t.Fatalf("method/status = %q/%q, want %q/%q", message.Method, message.Status, tc.method, tc.status)
			}
			if len(message.Attendees) != tc.attendeeCount {
				t.Fatalf("attendees = %d, want %d", len(message.Attendees), tc.attendeeCount)
			}
			if message.UID != "event-1" || message.Sequence != 7 || message.Summary != "Planning, Q2" {
				t.Fatalf("unexpected normalized core: %#v", message)
			}
		})
	}
}

func TestParseITIP_NormalizesTextParametersEmailRRULEAndDefaults(t *testing.T) {
	message := mustParse(t, timedRequestWire())

	if message.Description != "Line 1\nLine 2" || message.Location != "Room; 1" {
		t.Fatalf("decoded text = %q / %q", message.Description, message.Location)
	}
	if message.Organizer.Email != "owner@example.com" || message.Organizer.DisplayName != "Owner\" ^ Ops" {
		t.Fatalf("organizer = %#v", message.Organizer)
	}
	if len(message.Attendees) != 2 || message.Attendees[0].Email != "a@example.com" || message.Attendees[1].Email != "b@example.com" {
		t.Fatalf("attendee normalization/order = %#v", message.Attendees)
	}
	if got := message.Attendees[0]; got.Role != calendar.ITIPParticipantRequired || got.Status != calendar.RSVPNeedsAction || got.RSVP {
		t.Fatalf("default attendee values = %#v", got)
	}
	if got := message.Attendees[1]; got.Role != calendar.ITIPParticipantOptional || got.Status != calendar.RSVPTentative || !got.RSVP {
		t.Fatalf("explicit attendee values = %#v", got)
	}
	if message.RRULE != "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,SU;COUNT=3" {
		t.Fatalf("RRULE = %q", message.RRULE)
	}
}

func TestParseITIP_PreservesDSTInstantsAndAllDayUnion(t *testing.T) {
	timed := mustParse(t, timedRequestWire())
	if timed.Temporal.Timed == nil || timed.Temporal.AllDay != nil {
		t.Fatalf("timed union = %#v", timed.Temporal)
	}
	if timed.Temporal.Timed.Timezone != "America/New_York" {
		t.Fatalf("timezone = %q", timed.Temporal.Timed.Timezone)
	}
	if want := time.Date(2026, 3, 8, 6, 30, 0, 0, time.UTC); !timed.Temporal.Timed.StartsAt.Equal(want) {
		t.Fatalf("start = %s, want %s", timed.Temporal.Timed.StartsAt, want)
	}
	if want := time.Date(2026, 3, 8, 7, 30, 0, 0, time.UTC); !timed.Temporal.Timed.EndsAt.Equal(want) {
		t.Fatalf("end = %s, want %s", timed.Temporal.Timed.EndsAt, want)
	}

	allDay := mustParse(t, allDayOccurrenceWire())
	if allDay.Temporal.AllDay == nil || allDay.Temporal.Timed != nil {
		t.Fatalf("all-day union = %#v", allDay.Temporal)
	}
	if got := *allDay.Temporal.AllDay; got.StartDate != "2026-04-01" || got.EndDateExclusive != "2026-04-03" || got.Timezone != "UTC" {
		t.Fatalf("all-day = %#v", got)
	}
	if allDay.RecurrenceID == nil || allDay.RecurrenceID.Date != "2026-04-02" || !allDay.RecurrenceID.Time.IsZero() {
		t.Fatalf("recurrence = %#v", allDay.RecurrenceID)
	}
}

func TestParseITIP_AcceptsLFAndUnfoldsSpaceOrTab(t *testing.T) {
	wire := strings.ReplaceAll(timedRequestWire(), "\r\n", "\n")
	wire = strings.Replace(wire, "SUMMARY:Planning\\, Q2\n", "SUMMARY:Plan\n ning\\, Q2\n", 1)
	wire = strings.Replace(wire, `ORGANIZER;CN="Owner^' ^^ Ops":`, "ORGANIZER;CN=\"Owner^'\n\t ^^ Ops\":", 1)
	message := mustParse(t, wire)
	if message.Summary != "Planning, Q2" || message.Organizer.DisplayName != "Owner\" ^ Ops" {
		t.Fatalf("unfolded values = %q / %q", message.Summary, message.Organizer.DisplayName)
	}
}

func TestParseITIP_DecodesTextAndCaretExactlyOnce(t *testing.T) {
	wire := replaceLine(timedRequestWire(), `SUMMARY:Planning\, Q2`, `SUMMARY:\\n`)
	wire = replaceLine(wire, `ORGANIZER;CN="Owner^' ^^ Ops":mailto:Owner@Example.COM`, `ORGANIZER;CN="^^n":mailto:Owner@Example.COM`)
	message := mustParse(t, wire)
	if message.Summary != `\n` {
		t.Fatalf("SUMMARY = %q, want literal backslash+n", message.Summary)
	}
	if message.Organizer.DisplayName != "^n" {
		t.Fatalf("CN = %q, want literal caret+n", message.Organizer.DisplayName)
	}
}

func TestParseITIP_AcceptsEveryCaretNewlineEscapeAndFoldedMultibyteText(t *testing.T) {
	wire := strings.Replace(
		timedRequestWire(),
		`ORGANIZER;CN="Owner^' ^^ Ops":mailto:Owner@Example.COM`,
		`ORGANIZER;CN="A^nB^NC":mailto:Owner@Example.COM`,
		1,
	)
	wire = replaceLine(
		wire,
		`SUMMARY:Planning\, Q2`,
		"SUMMARY:"+strings.Repeat("Ж", 50)+"\r\n "+strings.Repeat("Ж", 50),
	)
	message := mustParse(t, wire)
	if message.Organizer.DisplayName != "A\nB\nC" {
		t.Fatalf("caret newlines = %q", message.Organizer.DisplayName)
	}
	if message.Summary != strings.Repeat("Ж", 100) {
		t.Fatalf("folded multibyte SUMMARY = %q", message.Summary)
	}
}

func TestParseITIP_EnforcesTotalLinePropertyAndCollectionLimits(t *testing.T) {
	payload := []byte(timedRequestWire())
	limits := calendar.DefaultITIPParseLimits()
	limits.MaxBytes = len(payload)
	if _, err := calendar.ParseITIP(payload, limits); err != nil {
		t.Fatalf("exact byte limit rejected: %v", err)
	}
	limits.MaxBytes--
	if _, err := calendar.ParseITIP(payload, limits); !errors.Is(err, calendar.ErrITIPLimitExceeded) {
		t.Fatalf("over byte limit error = %v", err)
	}

	limits = calendar.DefaultITIPParseLimits()
	lines := strings.Split(strings.TrimSuffix(timedRequestWire(), "\r\n"), "\r\n")
	limits.MaxProperties = len(lines)
	if _, err := calendar.ParseITIP(payload, limits); err != nil {
		t.Fatalf("exact property limit rejected: %v", err)
	}
	limits.MaxProperties--
	if _, err := calendar.ParseITIP(payload, limits); !errors.Is(err, calendar.ErrITIPLimitExceeded) {
		t.Fatalf("over property limit error = %v", err)
	}

	maxLine := 0
	for _, line := range lines {
		if len(line) > maxLine {
			maxLine = len(line)
		}
	}
	limits = calendar.DefaultITIPParseLimits()
	limits.MaxUnfoldedLineBytes = maxLine
	if _, err := calendar.ParseITIP(payload, limits); err != nil {
		t.Fatalf("exact line limit rejected: %v", err)
	}
	limits.MaxUnfoldedLineBytes--
	if _, err := calendar.ParseITIP(payload, limits); !errors.Is(err, calendar.ErrITIPLimitExceeded) {
		t.Fatalf("over unfolded-line limit error = %v", err)
	}

	limits = calendar.DefaultITIPParseLimits()
	limits.MaxAttendees = 1
	if _, err := calendar.ParseITIP(payload, limits); !errors.Is(err, calendar.ErrITIPLimitExceeded) {
		t.Fatalf("over attendee limit error = %v", err)
	}
	limits = calendar.DefaultITIPParseLimits()
	limits.MaxExDates = 1
	if _, err := calendar.ParseITIP(payload, limits); !errors.Is(err, calendar.ErrITIPLimitExceeded) {
		t.Fatalf("over EXDATE limit error = %v", err)
	}
}

func TestParseITIP_UsesPostNormalizationUTF8ByteMaxima(t *testing.T) {
	acceptedUID := strings.Repeat("Ж", 127) + "x"
	accepted := replaceLine(timedRequestWire(), "UID:event-1", "UID:"+acceptedUID)
	accepted = replaceLine(accepted, `SUMMARY:Planning\, Q2`, "SUMMARY:"+strings.Repeat("Ж", 250))
	accepted = strings.Replace(
		accepted,
		`ORGANIZER;CN="Owner^' ^^ Ops":mailto:Owner@Example.COM`,
		`ORGANIZER;CN="`+strings.Repeat("Ж", 80)+`":mailto:Owner@Example.COM`,
		1,
	)
	if got := mustParse(t, accepted).UID; got != acceptedUID {
		t.Fatalf("accepted UID changed: %q", got)
	}

	cases := map[string]string{
		"UID 256 bytes":     replaceLine(timedRequestWire(), "UID:event-1", "UID:"+strings.Repeat("Ж", 128)),
		"summary 502 bytes": replaceLine(timedRequestWire(), `SUMMARY:Planning\, Q2`, "SUMMARY:"+strings.Repeat("Ж", 251)),
		"CN 162 bytes": strings.Replace(
			timedRequestWire(),
			`ORGANIZER;CN="Owner^' ^^ Ops":mailto:Owner@Example.COM`,
			`ORGANIZER;CN="`+strings.Repeat("Ж", 81)+`":mailto:Owner@Example.COM`,
			1,
		),
	}
	for name, wire := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := calendar.ParseITIP([]byte(wire), calendar.DefaultITIPParseLimits()); err == nil {
				t.Fatal("ParseITIP() unexpectedly accepted over-limit normalized value")
			}
		})
	}
}

func TestParseITIP_RejectsMalformedPhysicalAndEscapedInput(t *testing.T) {
	invalidUTF8 := append([]byte(timedRequestWire()), 0xff)
	cases := map[string][]byte{
		"lone CR":             []byte(strings.Replace(timedRequestWire(), "\r\n", "\r", 1)),
		"invalid UTF-8":       invalidUTF8,
		"orphan continuation": []byte(" continuation\r\n" + timedRequestWire()),
		"terminal text slash": []byte(replaceLine(timedRequestWire(), `SUMMARY:Planning\, Q2`, `SUMMARY:secret\`)),
		"unknown text escape": []byte(replaceLine(timedRequestWire(), `SUMMARY:Planning\, Q2`, `SUMMARY:secret\q`)),
		"terminal caret": []byte(strings.Replace(
			timedRequestWire(),
			`ORGANIZER;CN="Owner^' ^^ Ops":mailto:Owner@Example.COM`,
			`ORGANIZER;CN="secret^":mailto:Owner@Example.COM`,
			1,
		)),
		"unknown caret": []byte(strings.Replace(
			timedRequestWire(),
			`ORGANIZER;CN="Owner^' ^^ Ops":mailto:Owner@Example.COM`,
			`ORGANIZER;CN="secret^q":mailto:Owner@Example.COM`,
			1,
		)),
		"unclosed quote": []byte(strings.Replace(
			timedRequestWire(),
			`ORGANIZER;CN="Owner^' ^^ Ops":mailto:Owner@Example.COM`,
			`ORGANIZER;CN="secret:mailto:Owner@Example.COM`,
			1,
		)),
		"literal quote": []byte(strings.Replace(
			timedRequestWire(),
			`ORGANIZER;CN="Owner^' ^^ Ops":mailto:Owner@Example.COM`,
			`ORGANIZER;CN="sec"ret":mailto:Owner@Example.COM`,
			1,
		)),
		"encoded UID LF": []byte(replaceLine(timedRequestWire(), "UID:event-1", `UID:private\nuid`)),
		"decoded text NUL": []byte(replaceLine(timedRequestWire(), `SUMMARY:Planning\, Q2`, "SUMMARY:secret\x00body")),
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := calendar.ParseITIP(payload, calendar.DefaultITIPParseLimits())
			if !errors.Is(err, calendar.ErrInvalidITIPSyntax) && !errors.Is(err, calendar.ErrInvalidITIPMessage) {
				t.Fatalf("ParseITIP() error = %v", err)
			}
			assertSafeError(t, err, "secret", "private", "body")
		})
	}
}

func TestParseITIP_RejectsUnknownPropertiesParametersAndComponents(t *testing.T) {
	cases := map[string]string{
		"unknown calendar property": insertBefore(timedRequestWire(), "BEGIN:VEVENT", "X-WR-CALNAME:secret"),
		"unknown event property":    insertBefore(timedRequestWire(), "END:VEVENT", "URL:https://secret.example"),
		"duration":                  insertBefore(timedRequestWire(), "END:VEVENT", "DURATION:PT1H"),
		"TRANSP":                    insertBefore(timedRequestWire(), "END:VEVENT", "TRANSP:OPAQUE"),
		"CLASS":                     insertBefore(timedRequestWire(), "END:VEVENT", "CLASS:PRIVATE"),
		"ATTACH":                    insertBefore(timedRequestWire(), "END:VEVENT", "ATTACH:https://secret.example/file"),
		"RDATE":                     insertBefore(timedRequestWire(), "END:VEVENT", "RDATE:20260309T120000Z"),
		"EXRULE":                    insertBefore(timedRequestWire(), "END:VEVENT", "EXRULE:FREQ=DAILY"),
		"VTIMEZONE": insertBefore(
			timedRequestWire(),
			"BEGIN:VEVENT",
			"BEGIN:VTIMEZONE\r\nTZID:America/New_York\r\nEND:VTIMEZONE",
		),
		"nested VALARM": insertBefore(
			timedRequestWire(),
			"END:VEVENT",
			"BEGIN:VALARM\r\nACTION:DISPLAY\r\nEND:VALARM",
		),
		"unknown UID parameter": replaceLine(timedRequestWire(), "UID:event-1", "UID;X-SECRET=yes:event-1"),
		"SENT-BY": strings.Replace(
			timedRequestWire(),
			`ORGANIZER;CN="Owner^' ^^ Ops":mailto:Owner@Example.COM`,
			`ORGANIZER;SENT-BY="mailto:secret@example.com":mailto:Owner@Example.COM`,
			1,
		),
		"RANGE": insertBefore(timedRequestWire(), "SEQUENCE:7", "RECURRENCE-ID;RANGE=THISANDFUTURE:20260315T053000Z"),
	}
	for name, wire := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := calendar.ParseITIP([]byte(wire), calendar.DefaultITIPParseLimits())
			if !errors.Is(err, calendar.ErrUnsupportedITIPContent) {
				t.Fatalf("ParseITIP() error = %v, want ErrUnsupportedITIPContent", err)
			}
			assertSafeError(t, err, "secret@example.com", "secret.example")
		})
	}
}

func TestParseITIP_RejectsComponentAndSingletonCardinalityViolations(t *testing.T) {
	cases := map[string]string{
		"missing VERSION":      removeLine(timedRequestWire(), "VERSION:2.0"),
		"duplicate VERSION":    insertBefore(timedRequestWire(), "BEGIN:VEVENT", "VERSION:2.0"),
		"wrong VERSION":        replaceLine(timedRequestWire(), "VERSION:2.0", "VERSION:1.0"),
		"missing PRODID":       removeLine(timedRequestWire(), "PRODID:-//Hub//Test//EN"),
		"missing UID":          removeLine(timedRequestWire(), "UID:event-1"),
		"missing DTSTAMP":      removeLine(timedRequestWire(), "DTSTAMP:20260301T120000Z"),
		"missing DTSTART":      removeLine(timedRequestWire(), "DTSTART;TZID=America/New_York:20260308T013000"),
		"missing DTEND":        removeLine(timedRequestWire(), "DTEND;TZID=America/New_York:20260308T033000"),
		"missing SEQUENCE":     removeLine(timedRequestWire(), "SEQUENCE:7"),
		"missing STATUS":       removeLine(timedRequestWire(), "STATUS:CONFIRMED"),
		"missing SUMMARY":      removeLine(timedRequestWire(), `SUMMARY:Planning\, Q2`),
		"missing ORGANIZER":    removeLine(timedRequestWire(), `ORGANIZER;CN="Owner^' ^^ Ops":mailto:Owner@Example.COM`),
		"duplicate METHOD":     insertBefore(timedRequestWire(), "BEGIN:VEVENT", "METHOD:REQUEST"),
		"missing VEVENT":       removeComponent(timedRequestWire(), "VEVENT"),
		"second VEVENT":        strings.Replace(timedRequestWire(), "END:VCALENDAR\r\n", minimalEventBlock()+"END:VCALENDAR\r\n", 1),
		"mismatched END":       strings.Replace(timedRequestWire(), "END:VEVENT", "END:VCALENDAR", 1),
		"property after root":  timedRequestWire() + "SUMMARY:secret\r\n",
		"duplicate UID":        insertBefore(timedRequestWire(), "DTSTAMP:", "UID:event-2"),
		"duplicate organizer": insertBefore(timedRequestWire(), "END:VEVENT", "ORGANIZER:mailto:other@example.com"),
	}
	for name, wire := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := calendar.ParseITIP([]byte(wire), calendar.DefaultITIPParseLimits()); err == nil {
				t.Fatal("ParseITIP() unexpectedly accepted invalid cardinality")
			}
		})
	}
}

func TestParseITIP_RejectsInvalidTemporalForms(t *testing.T) {
	cases := map[string]string{
		"floating": replaceLine(
			replaceLine(timedRequestWire(), "DTSTART;TZID=America/New_York:20260308T013000", "DTSTART:20260308T013000"),
			"DTEND;TZID=America/New_York:20260308T033000",
			"DTEND:20260308T033000",
		),
		"mixed forms": replaceLine(timedRequestWire(), "DTEND;TZID=America/New_York:20260308T033000", "DTEND:20260308T073000Z"),
		"different zones": replaceLine(timedRequestWire(), "DTEND;TZID=America/New_York:20260308T033000", "DTEND;TZID=Europe/London:20260308T073000"),
		"fractional seconds": replaceLine(timedRequestWire(), "DTSTAMP:20260301T120000Z", "DTSTAMP:20260301T120000.1Z"),
		"DST gap": replaceLine(timedRequestWire(), "DTSTART;TZID=America/New_York:20260308T013000", "DTSTART;TZID=America/New_York:20260308T023000"),
		"unknown timezone": strings.ReplaceAll(timedRequestWire(), "America/New_York", "Mars/Olympus"),
		"reversed": replaceLine(timedRequestWire(), "DTEND;TZID=America/New_York:20260308T033000", "DTEND;TZID=America/New_York:20260308T010000"),
		"DATE with TZID": strings.Replace(
			allDayOccurrenceWire(),
			"DTSTART;VALUE=DATE:20260401",
			"DTSTART;VALUE=DATE;TZID=UTC:20260401",
			1,
		),
	}
	for name, wire := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := calendar.ParseITIP([]byte(wire), calendar.DefaultITIPParseLimits())
			if err == nil {
				t.Fatal("ParseITIP() unexpectedly accepted invalid temporal form")
			}
			assertSafeError(t, err, "Mars/Olympus")
		})
	}
}

func TestParseITIP_NormalizesSupportedRRULEs(t *testing.T) {
	cases := map[string]string{
		"daily default interval":   "FREQ=DAILY;INTERVAL=1;COUNT=4",
		"weekly sorted days":       "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,FR,SU;COUNT=3",
		"monthly day with date end": "FREQ=MONTHLY;INTERVAL=2;BYMONTHDAY=31;UNTIL=20261231T235959Z",
	}
	wires := map[string]string{
		"daily default interval":   replaceLine(timedRequestWire(), "RRULE:COUNT=3;BYDAY=SU,MO;FREQ=WEEKLY", "RRULE:COUNT=4;FREQ=DAILY"),
		"weekly sorted days":       replaceLine(timedRequestWire(), "RRULE:COUNT=3;BYDAY=SU,MO;FREQ=WEEKLY", "RRULE:BYDAY=SU,FR,MO;COUNT=3;FREQ=WEEKLY"),
		"monthly day with date end": replaceLine(timedRequestWire(), "RRULE:COUNT=3;BYDAY=SU,MO;FREQ=WEEKLY", "RRULE:UNTIL=20261231T235959Z;BYMONTHDAY=31;INTERVAL=2;FREQ=MONTHLY"),
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			if got := mustParse(t, wires[name]).RRULE; got != want {
				t.Fatalf("RRULE = %q, want %q", got, want)
			}
		})
	}
}

func TestParseITIP_RejectsUnsupportedRRULEGrammar(t *testing.T) {
	invalid := []string{
		"FREQ=YEARLY;COUNT=2",
		"FREQ=DAILY;BYDAY=MO;COUNT=2",
		"FREQ=WEEKLY;COUNT=2",
		"FREQ=WEEKLY;BYDAY=1MO;COUNT=2",
		"FREQ=WEEKLY;BYDAY=MO,MO;COUNT=2",
		"FREQ=MONTHLY;BYDAY=MO;COUNT=2",
		"FREQ=MONTHLY;BYMONTHDAY=0;COUNT=2",
		"FREQ=MONTHLY;BYMONTHDAY=1,2;COUNT=2",
		"FREQ=DAILY;COUNT=2;UNTIL=20261231T235959Z",
		"FREQ=DAILY;INTERVAL=0;COUNT=2",
		"FREQ=DAILY;COUNT=1001",
		"FREQ=DAILY;BYHOUR=1;COUNT=2",
		"FREQ=DAILY;FREQ=DAILY;COUNT=2",
		"FREQ=DAILY;UNTIL=20261231",
	}
	for _, rule := range invalid {
		t.Run(rule, func(t *testing.T) {
			wire := replaceLine(timedRequestWire(), "RRULE:COUNT=3;BYDAY=SU,MO;FREQ=WEEKLY", "RRULE:"+rule)
			if _, err := calendar.ParseITIP([]byte(wire), calendar.DefaultITIPParseLimits()); !errors.Is(err, calendar.ErrInvalidITIPRecurrence) {
				t.Fatalf("ParseITIP() error = %v, want ErrInvalidITIPRecurrence", err)
			}
		})
	}
}

func TestParseITIP_RejectsMethodStatusAndParticipantViolations(t *testing.T) {
	cases := map[string]string{
		"unsupported method": replaceLine(timedRequestWire(), "METHOD:REQUEST", "METHOD:PUBLISH"),
		"cancel confirmed": replaceLine(
			strings.Replace(timedRequestWire(), "METHOD:REQUEST", "METHOD:CANCEL", 1),
			"STATUS:CONFIRMED",
			"STATUS:CONFIRMED",
		),
		"request cancelled": replaceLine(timedRequestWire(), "STATUS:CONFIRMED", "STATUS:CANCELLED"),
		"reply zero attendees": removeLine(replyWire(), "ATTENDEE;CN=Responder;ROLE=REQ-PARTICIPANT;PARTSTAT=TENTATIVE;RSVP=FALSE:mailto:responder@example.com"),
		"reply two attendees": insertBefore(replyWire(), "END:VEVENT", "ATTENDEE;PARTSTAT=ACCEPTED:mailto:second@example.com"),
		"reply needs action": strings.Replace(replyWire(), "PARTSTAT=TENTATIVE", "PARTSTAT=NEEDS-ACTION", 1),
		"duplicate attendee": insertBefore(timedRequestWire(), "RRULE:", "ATTENDEE:mailto:A@example.com"),
		"bad RSVP": strings.Replace(timedRequestWire(), "RSVP=TRUE", "RSVP=MAYBE", 1),
		"bad ROLE": strings.Replace(timedRequestWire(), "ROLE=OPT-PARTICIPANT", "ROLE=CHAIR", 1),
		"bad PARTSTAT": strings.Replace(timedRequestWire(), "PARTSTAT=TENTATIVE", "PARTSTAT=DELEGATED", 1),
	}
	for name, wire := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := calendar.ParseITIP([]byte(wire), calendar.DefaultITIPParseLimits()); err == nil {
				t.Fatal("ParseITIP() unexpectedly accepted invalid method/status/participant")
			}
		})
	}
}

func TestParseITIP_AcceptsOccurrenceRequestAndRejectsDuplicateRecurrences(t *testing.T) {
	message := mustParse(t, allDayOccurrenceWire())
	if message.Method != calendar.ITIPMethodRequest || message.RecurrenceID == nil {
		t.Fatalf("occurrence REQUEST not preserved: %#v", message)
	}
	timedWire := insertBefore(
		timedRequestWire(),
		"SEQUENCE:7",
		"RECURRENCE-ID;TZID=America/New_York:20260315T013000",
	)
	timed := mustParse(t, timedWire)
	timedID := mustIdentity(t, timed)
	if !timedID.HasRecurrenceID || timedID.RecurrenceID != "2026-03-15T05:30:00Z" {
		t.Fatalf("timed occurrence identity = %#v", timedID)
	}

	duplicateExdate := replaceLine(
		allDayOccurrenceWire(),
		"EXDATE;VALUE=DATE:20260402,20260404",
		"EXDATE;VALUE=DATE:20260402,20260402",
	)
	if _, err := calendar.ParseITIP([]byte(duplicateExdate), calendar.DefaultITIPParseLimits()); !errors.Is(err, calendar.ErrInvalidITIPRecurrence) {
		t.Fatalf("duplicate EXDATE error = %v", err)
	}
}

func TestParseITIP_ReturnsFreshSlicesAndDoesNotAliasInput(t *testing.T) {
	payload := []byte(timedRequestWire())
	first, err := calendar.ParseITIP(payload, calendar.DefaultITIPParseLimits())
	if err != nil {
		t.Fatalf("first ParseITIP(): %v", err)
	}
	second, err := calendar.ParseITIP(payload, calendar.DefaultITIPParseLimits())
	if err != nil {
		t.Fatalf("second ParseITIP(): %v", err)
	}

	for index := range payload {
		payload[index] = 'x'
	}
	first.Attendees[0].Email = "mutated@example.com"
	first.EXDates[0].Date = "mutated"
	if first.UID != "event-1" || first.Summary != "Planning, Q2" {
		t.Fatalf("parsed strings alias input: %#v", first)
	}
	if second.Attendees[0].Email == "mutated@example.com" || second.EXDates[0].Date == "mutated" {
		t.Fatalf("returned slices share backing arrays: %#v / %#v", first, second)
	}
}

func TestCanonicalITIPIdentity_UsesFrozenLiteralCVEV1Vector(t *testing.T) {
	message := calendar.ITIPMessage{
		Method:   calendar.ITIPMethodRequest,
		UID:      "u",
		Stamp:    time.Unix(0, 0).UTC(),
		Sequence: 0,
		Status:   calendar.ITIPStatusConfirmed,
		Summary:  "s",
		Temporal: calendar.ITIPTemporal{AllDay: &calendar.ITIPAllDay{
			StartDate:        "1970-01-01",
			EndDateExclusive: "1970-01-02",
			Timezone:         "UTC",
		}},
		Organizer: calendar.ITIPOrganizer{Email: "a@b.co"},
	}

	literalCVE := []byte(
		"hub.calendar.itip.payload\x00\x01" +
			"\x01\x00\x00\x00\x01\x01" +
			"\x02\x00\x00\x00\x01u" +
			"\x03\x00\x00\x00\x01\x00" +
			"\x04\x00\x00\x00\x08\x00\x00\x00\x00\x00\x00\x00\x00" +
			"\x05\x00\x00\x00\x08\x00\x00\x00\x00\x00\x00\x00\x00" +
			"\x06\x00\x00\x00\x01\x01" +
			"\x07\x00\x00\x00\x24\x02\x00\x00\x00\x03UTC\x00\x00\x00\x0a1970-01-01\x00\x00\x00\x0a1970-01-02" +
			"\x08\x00\x00\x00\x0e\x00\x00\x00\x06a@b.co\x00\x00\x00\x00" +
			"\x09\x00\x00\x00\x04\x00\x00\x00\x00" +
			"\x0a\x00\x00\x00\x01s" +
			"\x0b\x00\x00\x00\x00" +
			"\x0c\x00\x00\x00\x00" +
			"\x0d\x00\x00\x00\x01\x00" +
			"\x0e\x00\x00\x00\x04\x00\x00\x00\x00",
	)
	wantHash := sha256.Sum256(literalCVE)
	identity, err := calendar.CanonicalITIPIdentity(message)
	if err != nil {
		t.Fatalf("CanonicalITIPIdentity(): %v", err)
	}
	if identity.CanonicalVersion != 1 || identity.Method != calendar.ITIPMethodRequest || identity.UID != "u" || identity.Sequence != 0 {
		t.Fatalf("identity fields = %#v", identity)
	}
	if identity.HasRecurrenceID || identity.RecurrenceID != "" {
		t.Fatalf("absent recurrence identity = %#v", identity)
	}
	if identity.PayloadHash != wantHash {
		t.Fatalf("hash = %x, want SHA-256(literal CVE-v1) %x", identity.PayloadHash, wantHash)
	}
}

func TestCanonicalITIPIdentity_EquivalentWireFormsHashEqually(t *testing.T) {
	left := mustParse(t, timedRequestWire())
	right := mustParse(t, equivalentTimedRequestWire())
	leftID, err := calendar.CanonicalITIPIdentity(left)
	if err != nil {
		t.Fatalf("left identity: %v", err)
	}
	rightID, err := calendar.CanonicalITIPIdentity(right)
	if err != nil {
		t.Fatalf("right identity: %v", err)
	}
	if leftID != rightID {
		t.Fatalf("equivalent payload identities differ:\nleft  %#v\nright %#v", leftID, rightID)
	}
}

func TestCanonicalITIPIdentity_DoesNotDecodeNormalizedValuesAgain(t *testing.T) {
	message := baseIdentityMessage()
	message.Summary = `\n`
	message.Organizer.DisplayName = "^n"
	literal, err := calendar.CanonicalITIPIdentity(message)
	if err != nil {
		t.Fatalf("literal identity: %v", err)
	}
	message.Summary = "\n"
	message.Organizer.DisplayName = "\n"
	decoded, err := calendar.CanonicalITIPIdentity(message)
	if err != nil {
		t.Fatalf("decoded identity: %v", err)
	}
	if literal.PayloadHash == decoded.PayloadHash {
		t.Fatal("identity decoded already-normalized backslash/caret sequences again")
	}
}

func TestCanonicalITIPIdentity_EveryIncludedFieldAffectsHash(t *testing.T) {
	base := baseIdentityMessage()
	baseID := mustIdentity(t, base)

	mutations := map[string]func(*calendar.ITIPMessage){
		"UID":         func(value *calendar.ITIPMessage) { value.UID = "other" },
		"recurrence":  func(value *calendar.ITIPMessage) { value.RecurrenceID.Time = value.RecurrenceID.Time.Add(time.Hour) },
		"sequence":    func(value *calendar.ITIPMessage) { value.Sequence++ },
		"stamp":       func(value *calendar.ITIPMessage) { value.Stamp = value.Stamp.Add(time.Second) },
		"status":      func(value *calendar.ITIPMessage) { value.Status = calendar.ITIPStatusTentative },
		"start":       func(value *calendar.ITIPMessage) { value.Temporal.Timed.StartsAt = value.Temporal.Timed.StartsAt.Add(time.Second) },
		"end":         func(value *calendar.ITIPMessage) { value.Temporal.Timed.EndsAt = value.Temporal.Timed.EndsAt.Add(time.Second) },
		"timezone":    func(value *calendar.ITIPMessage) { value.Temporal.Timed.Timezone = "Europe/London" },
		"organizer":   func(value *calendar.ITIPMessage) { value.Organizer.Email = "other@example.com" },
		"organizer CN": func(value *calendar.ITIPMessage) { value.Organizer.DisplayName = "Other" },
		"attendee":    func(value *calendar.ITIPMessage) { value.Attendees[0].Email = "other-attendee@example.com" },
		"attendee CN": func(value *calendar.ITIPMessage) { value.Attendees[0].DisplayName = "Other" },
		"role":        func(value *calendar.ITIPMessage) { value.Attendees[0].Role = calendar.ITIPParticipantOptional },
		"partstat":    func(value *calendar.ITIPMessage) { value.Attendees[0].Status = calendar.RSVPTentative },
		"rsvp":        func(value *calendar.ITIPMessage) { value.Attendees[0].RSVP = false },
		"summary":     func(value *calendar.ITIPMessage) { value.Summary = "Other" },
		"description": func(value *calendar.ITIPMessage) { value.Description = "Other" },
		"location":    func(value *calendar.ITIPMessage) { value.Location = "Other" },
		"rrule":       func(value *calendar.ITIPMessage) { value.RRULE = "FREQ=DAILY;COUNT=3" },
		"exdate":      func(value *calendar.ITIPMessage) { value.EXDates[0].Time = value.EXDates[0].Time.Add(time.Hour) },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			candidate := cloneMessage(base)
			mutate(&candidate)
			identity := mustIdentity(t, candidate)
			if identity.PayloadHash == baseID.PayloadHash {
				t.Fatalf("%s mutation did not change hash", name)
			}
		})
	}

	methodMutation := cloneMessage(base)
	methodMutation.Method = calendar.ITIPMethodCancel
	methodMutation.Status = calendar.ITIPStatusCancelled
	if got := mustIdentity(t, methodMutation); got.PayloadHash == baseID.PayloadHash {
		t.Fatal("method mutation did not change hash")
	}
}

func TestCanonicalITIPIdentity_AllDayTemporalFieldsAffectHash(t *testing.T) {
	base := mustParse(t, allDayOccurrenceWire())
	baseID := mustIdentity(t, base)
	mutations := map[string]func(*calendar.ITIPMessage){
		"start date": func(value *calendar.ITIPMessage) { value.Temporal.AllDay.StartDate = "2026-03-31" },
		"end date":   func(value *calendar.ITIPMessage) { value.Temporal.AllDay.EndDateExclusive = "2026-04-04" },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			candidate := cloneMessage(base)
			mutate(&candidate)
			if got := mustIdentity(t, candidate); got.PayloadHash == baseID.PayloadHash {
				t.Fatalf("%s mutation did not change hash", name)
			}
		})
	}
}

func TestCanonicalITIPIdentity_NormalizesOrderAndRejectsCanonicalDuplicates(t *testing.T) {
	left := baseIdentityMessage()
	left.Attendees = append(left.Attendees, calendar.ITIPParticipant{
		Email: "b@example.com", Role: calendar.ITIPParticipantOptional, Status: calendar.RSVPTentative,
	})
	left.EXDates = append(left.EXDates, calendar.ITIPRecurrenceID{Time: time.Date(2026, 3, 22, 5, 30, 0, 0, time.UTC)})
	right := cloneMessage(left)
	right.Attendees[0], right.Attendees[1] = right.Attendees[1], right.Attendees[0]
	right.EXDates[0], right.EXDates[1] = right.EXDates[1], right.EXDates[0]
	if leftID, rightID := mustIdentity(t, left), mustIdentity(t, right); leftID != rightID {
		t.Fatalf("order changed identity: %#v / %#v", leftID, rightID)
	}

	duplicateAttendee := cloneMessage(left)
	duplicateAttendee.Attendees[1].Email = "A@EXAMPLE.COM"
	if _, err := calendar.CanonicalITIPIdentity(duplicateAttendee); !errors.Is(err, calendar.ErrInvalidITIPParticipant) {
		t.Fatalf("duplicate attendee error = %v", err)
	}
	duplicateExdate := cloneMessage(left)
	duplicateExdate.EXDates[1].Time = duplicateExdate.EXDates[0].Time
	if _, err := calendar.CanonicalITIPIdentity(duplicateExdate); !errors.Is(err, calendar.ErrInvalidITIPRecurrence) {
		t.Fatalf("duplicate EXDATE error = %v", err)
	}
}

func TestCanonicalITIPIdentity_RecurrenceIdentityIsExplicitAndNonAliasable(t *testing.T) {
	message := baseIdentityMessage()
	identity := mustIdentity(t, message)
	if !identity.HasRecurrenceID || identity.RecurrenceID != "2026-03-15T05:30:00Z" {
		t.Fatalf("timed recurrence identity = %#v", identity)
	}
	message.RecurrenceID.Time = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if identity.RecurrenceID != "2026-03-15T05:30:00Z" {
		t.Fatalf("identity aliases recurrence input: %#v", identity)
	}

	allDay := mustParse(t, allDayOccurrenceWire())
	allDayID := mustIdentity(t, allDay)
	if !allDayID.HasRecurrenceID || allDayID.RecurrenceID != "2026-04-02" {
		t.Fatalf("all-day recurrence identity = %#v", allDayID)
	}
}

func TestCompareITIPSequence_ClassifiesAndRejectsNegativeInputs(t *testing.T) {
	cases := []struct {
		incoming int
		current  int
		want     calendar.ITIPSequenceRelation
	}{
		{incoming: 4, current: 5, want: calendar.ITIPSequenceLower},
		{incoming: 5, current: 5, want: calendar.ITIPSequenceEqual},
		{incoming: 6, current: 5, want: calendar.ITIPSequenceNewer},
	}
	for _, tc := range cases {
		got, err := calendar.CompareITIPSequence(tc.incoming, tc.current)
		if err != nil || got != tc.want {
			t.Fatalf("CompareITIPSequence(%d, %d) = %v, %v; want %v, nil", tc.incoming, tc.current, got, err, tc.want)
		}
	}
	for _, tc := range [][2]int{{-1, 0}, {0, -1}, {-1, -1}} {
		if _, err := calendar.CompareITIPSequence(tc[0], tc[1]); !errors.Is(err, calendar.ErrInvalidITIPSequence) {
			t.Fatalf("CompareITIPSequence(%d, %d) error = %v", tc[0], tc[1], err)
		}
	}
}

func TestParserAndIdentityErrorsAreStableAndContentFree(t *testing.T) {
	privateWire := replaceLine(timedRequestWire(), "UID:event-1", "UID:private-uid@example.test\x00")
	_, parseErr := calendar.ParseITIP([]byte(privateWire), calendar.DefaultITIPParseLimits())
	if parseErr == nil {
		t.Fatal("ParseITIP() unexpectedly accepted private invalid input")
	}
	assertSafeError(t, parseErr, "private-uid", "example.test", "Planning")

	message := baseIdentityMessage()
	message.Organizer.Email = "private-owner@example.test\x00"
	_, identityErr := calendar.CanonicalITIPIdentity(message)
	if !errors.Is(identityErr, calendar.ErrInvalidITIPEmail) {
		t.Fatalf("identity error = %v", identityErr)
	}
	assertSafeError(t, identityErr, "private-owner", "example.test")
}

func withLimit(value calendar.ITIPParseLimits, mutate func(*calendar.ITIPParseLimits)) calendar.ITIPParseLimits {
	mutate(&value)
	return value
}

func mustParse(t *testing.T, wire string) calendar.ITIPMessage {
	t.Helper()
	message, err := calendar.ParseITIP([]byte(wire), calendar.DefaultITIPParseLimits())
	if err != nil {
		t.Fatalf("ParseITIP(): %v", err)
	}
	return message
}

func mustIdentity(t *testing.T, message calendar.ITIPMessage) calendar.ITIPIdentity {
	t.Helper()
	identity, err := calendar.CanonicalITIPIdentity(message)
	if err != nil {
		t.Fatalf("CanonicalITIPIdentity(): %v", err)
	}
	return identity
}

func assertSafeError(t *testing.T, err error, forbidden ...string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	for _, value := range forbidden {
		if value != "" && strings.Contains(err.Error(), value) {
			t.Fatalf("error %q leaks forbidden content %q", err, value)
		}
	}
}

func replaceLine(wire, oldLine, newLine string) string {
	return strings.Replace(wire, oldLine+"\r\n", newLine+"\r\n", 1)
}

func removeLine(wire, line string) string {
	return strings.Replace(wire, line+"\r\n", "", 1)
}

func insertBefore(wire, prefix, line string) string {
	return strings.Replace(wire, prefix, line+"\r\n"+prefix, 1)
}

func removeComponent(wire, component string) string {
	start := strings.Index(wire, "BEGIN:"+component+"\r\n")
	endMarker := "END:" + component + "\r\n"
	end := strings.Index(wire, endMarker)
	if start < 0 || end < start {
		return wire
	}
	return wire[:start] + wire[end+len(endMarker):]
}

func minimalEventBlock() string {
	return strings.Join([]string{
		"BEGIN:VEVENT",
		"UID:second",
		"DTSTAMP:20260301T120000Z",
		"DTSTART:20260308T063000Z",
		"DTEND:20260308T073000Z",
		"SEQUENCE:0",
		"STATUS:CONFIRMED",
		"SUMMARY:Second",
		"ORGANIZER:mailto:owner@example.com",
		"END:VEVENT",
	}, "\r\n") + "\r\n"
}

func timedRequestWire() string {
	return strings.Join([]string{
		"BEGIN:VCALENDAR",
		"PRODID:-//Hub//Test//EN",
		"VERSION:2.0",
		"CALSCALE:GREGORIAN",
		"METHOD:REQUEST",
		"BEGIN:VEVENT",
		"UID:event-1",
		"DTSTAMP:20260301T120000Z",
		"DTSTART;TZID=America/New_York:20260308T013000",
		"DTEND;TZID=America/New_York:20260308T033000",
		"SEQUENCE:7",
		"STATUS:CONFIRMED",
		`SUMMARY:Planning\, Q2`,
		`DESCRIPTION:Line 1\nLine 2`,
		`LOCATION:Room\; 1`,
		`ORGANIZER;CN="Owner^' ^^ Ops":mailto:Owner@Example.COM`,
		`ATTENDEE;CN="Bee";ROLE=OPT-PARTICIPANT;PARTSTAT=TENTATIVE;RSVP=TRUE:mailto:B@example.com`,
		"ATTENDEE:mailto:a@example.com",
		"RRULE:COUNT=3;BYDAY=SU,MO;FREQ=WEEKLY",
		"EXDATE;TZID=America/New_York:20260315T013000,20260322T013000",
		"END:VEVENT",
		"END:VCALENDAR",
	}, "\r\n") + "\r\n"
}

func equivalentTimedRequestWire() string {
	return strings.Join([]string{
		"begin:vcalendar",
		"method:request",
		"calscale:gregorian",
		"version:2.0",
		"prodid:-//Hub//Test//EN",
		"begin:vevent",
		"organizer;cn=\"Owner^' ^^ Ops\":MAILTO:owner@example.com",
		"summary:Plan",
		" ning\\, Q2",
		"status:confirmed",
		"sequence:7",
		"dtend;tzid=America/New_York:20260308T033000",
		"dtstart;tzid=America/New_York:20260308T013000",
		"dtstamp:20260301T120000Z",
		"uid:event-1",
		"exdate;tzid=America/New_York:20260322T013000,20260315T013000",
		"rrule:INTERVAL=1;FREQ=weekly;COUNT=3;BYDAY=MO,SU",
		"attendee:mailto:A@EXAMPLE.COM",
		"attendee;partstat=tentative;rsvp=true;role=opt-participant;cn=Bee:MAILTO:b@example.com",
		`location:Room\; 1`,
		`description:Line 1\NLine 2`,
		"end:vevent",
		"end:vcalendar",
	}, "\n") + "\n"
}

func cancelWire() string {
	wire := timedRequestWire()
	wire = replaceLine(wire, "METHOD:REQUEST", "METHOD:CANCEL")
	wire = replaceLine(wire, "STATUS:CONFIRMED", "STATUS:CANCELLED")
	wire = removeLine(wire, `ATTENDEE;CN="Bee";ROLE=OPT-PARTICIPANT;PARTSTAT=TENTATIVE;RSVP=TRUE:mailto:B@example.com`)
	wire = removeLine(wire, "ATTENDEE:mailto:a@example.com")
	return wire
}

func replyWire() string {
	wire := timedRequestWire()
	wire = replaceLine(wire, "METHOD:REQUEST", "METHOD:REPLY")
	wire = replaceLine(wire, "STATUS:CONFIRMED", "STATUS:TENTATIVE")
	wire = removeLine(wire, `ATTENDEE;CN="Bee";ROLE=OPT-PARTICIPANT;PARTSTAT=TENTATIVE;RSVP=TRUE:mailto:B@example.com`)
	wire = replaceLine(
		wire,
		"ATTENDEE:mailto:a@example.com",
		"ATTENDEE;CN=Responder;ROLE=REQ-PARTICIPANT;PARTSTAT=TENTATIVE;RSVP=FALSE:mailto:responder@example.com",
	)
	return wire
}

func allDayOccurrenceWire() string {
	return strings.Join([]string{
		"BEGIN:VCALENDAR",
		"PRODID:-//Hub//Test//EN",
		"VERSION:2.0",
		"METHOD:REQUEST",
		"BEGIN:VEVENT",
		"UID:event-1",
		"DTSTAMP:20260301T120000Z",
		"DTSTART;VALUE=DATE:20260401",
		"DTEND;VALUE=DATE:20260403",
		"RECURRENCE-ID;VALUE=DATE:20260402",
		"SEQUENCE:7",
		"STATUS:CONFIRMED",
		"SUMMARY:Planning",
		"ORGANIZER:mailto:owner@example.com",
		"EXDATE;VALUE=DATE:20260402,20260404",
		"END:VEVENT",
		"END:VCALENDAR",
	}, "\r\n") + "\r\n"
}

func baseIdentityMessage() calendar.ITIPMessage {
	return calendar.ITIPMessage{
		Method:       calendar.ITIPMethodRequest,
		UID:          "event-1",
		Stamp:        time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		Sequence:     7,
		Status:       calendar.ITIPStatusConfirmed,
		Summary:      "Planning",
		Description:  "Description",
		Location:     "Room 1",
		Temporal: calendar.ITIPTemporal{Timed: &calendar.ITIPTimed{
			StartsAt: time.Date(2026, 3, 8, 6, 30, 0, 0, time.UTC),
			EndsAt:   time.Date(2026, 3, 8, 7, 30, 0, 0, time.UTC),
			Timezone: "America/New_York",
		}},
		Organizer: calendar.ITIPOrganizer{Email: "owner@example.com", DisplayName: "Owner"},
		Attendees: []calendar.ITIPParticipant{{
			Email:       "a@example.com",
			DisplayName: "A",
			Role:        calendar.ITIPParticipantRequired,
			Status:      calendar.RSVPAccepted,
			RSVP:        true,
		}},
		RRULE: "COUNT=2;FREQ=DAILY",
		EXDates: []calendar.ITIPRecurrenceID{{
			Time: time.Date(2026, 3, 15, 5, 30, 0, 0, time.UTC),
		}},
		RecurrenceID: &calendar.ITIPRecurrenceID{
			Time: time.Date(2026, 3, 15, 5, 30, 0, 0, time.UTC),
		},
	}
}

func cloneMessage(value calendar.ITIPMessage) calendar.ITIPMessage {
	clone := value
	if value.Temporal.Timed != nil {
		timed := *value.Temporal.Timed
		clone.Temporal.Timed = &timed
	}
	if value.Temporal.AllDay != nil {
		allDay := *value.Temporal.AllDay
		clone.Temporal.AllDay = &allDay
	}
	if value.RecurrenceID != nil {
		recurrence := *value.RecurrenceID
		clone.RecurrenceID = &recurrence
	}
	clone.Attendees = append([]calendar.ITIPParticipant(nil), value.Attendees...)
	clone.EXDates = append([]calendar.ITIPRecurrenceID(nil), value.EXDates...)
	return clone
}
