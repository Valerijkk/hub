package mail_test

import (
	"testing"

	"github.com/google/uuid"
	mail "github.com/valerii/hub/backend/internal/mail"
)

const mailboxCalendarRSVPTargetAttendeeVersion = uint(7)

const mailboxCalendarRSVPTargetRSVPStatus = mail.MailboxCalendarRSVPStatusAccepted

func newMailboxCalendarRSVPTargetRequest(t *testing.T) mail.MailboxCalendarCorrelationRequest {
	t.Helper()
	request, err := mail.NewMailboxCalendarCorrelationRequest(
		uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		uuid.MustParse("44444444-4444-4444-8444-444444444444"),
	)
	if err != nil {
		t.Fatalf("NewMailboxCalendarCorrelationRequest() error = %v", err)
	}
	return request
}

func requireMailboxCalendarRSVPTargetInvalid(t *testing.T, err error) {
	t.Helper()
	kind, ok := mail.MailboxCalendarCorrelationErrorKindOf(err)
	if !ok || kind != mail.MailboxCalendarCorrelationErrorKindInvalid {
		t.Fatalf("MailboxCalendarCorrelationErrorKindOf(%v) = (%q, %t), want (invalid, true)", err, kind, ok)
	}
}

func TestMailboxCalendarRSVPTargetAcceptsSeriesWithoutRecurrence(t *testing.T) {
	eventID := uuid.MustParse("55555555-5555-4555-8555-555555555555")
	target, err := mail.NewMailboxCalendarRSVPTarget(eventID, mail.MailboxCalendarRSVPTargetScopeSeries, nil, mailboxCalendarRSVPTargetAttendeeVersion, mailboxCalendarRSVPTargetRSVPStatus)
	if err != nil {
		t.Fatalf("NewMailboxCalendarRSVPTarget() error = %v", err)
	}
	if target.CalendarEventID() != eventID || target.Scope() != mail.MailboxCalendarRSVPTargetScopeSeries || target.RecurrenceID() != nil || target.AttendeeVersion() != mailboxCalendarRSVPTargetAttendeeVersion || target.RSVPStatus() != mailboxCalendarRSVPTargetRSVPStatus {
		t.Fatalf("NewMailboxCalendarRSVPTarget() = %#v, want series target without recurrence", target)
	}
}

func TestMailboxCalendarRSVPTargetAcceptsCanonicalOccurrenceIDs(t *testing.T) {
	eventID := uuid.MustParse("55555555-5555-4555-8555-555555555555")
	for _, recurrenceID := range []string{"2026-09-05", "2026-09-05T12:34:56Z"} {
		t.Run(recurrenceID, func(t *testing.T) {
			target, err := mail.NewMailboxCalendarRSVPTarget(eventID, mail.MailboxCalendarRSVPTargetScopeOccurrence, &recurrenceID, mailboxCalendarRSVPTargetAttendeeVersion, mailboxCalendarRSVPTargetRSVPStatus)
			if err != nil {
				t.Fatalf("NewMailboxCalendarRSVPTarget() error = %v", err)
			}
			got := target.RecurrenceID()
			if target.CalendarEventID() != eventID || target.Scope() != mail.MailboxCalendarRSVPTargetScopeOccurrence || got == nil || *got != recurrenceID || target.AttendeeVersion() != mailboxCalendarRSVPTargetAttendeeVersion || target.RSVPStatus() != mailboxCalendarRSVPTargetRSVPStatus {
				t.Fatalf("NewMailboxCalendarRSVPTarget() = %#v, want occurrence target for %q", target, recurrenceID)
			}
		})
	}
}

func TestMailboxCalendarRSVPTargetCopiesConstructorAndAccessorRecurrence(t *testing.T) {
	eventID := uuid.MustParse("55555555-5555-4555-8555-555555555555")
	recurrenceID := "2026-09-05T12:34:56Z"
	target, err := mail.NewMailboxCalendarRSVPTarget(eventID, mail.MailboxCalendarRSVPTargetScopeOccurrence, &recurrenceID, mailboxCalendarRSVPTargetAttendeeVersion, mailboxCalendarRSVPTargetRSVPStatus)
	if err != nil {
		t.Fatalf("NewMailboxCalendarRSVPTarget() error = %v", err)
	}

	recurrenceID = "2026-09-06T12:34:56Z"
	first := target.RecurrenceID()
	if first == nil || *first != "2026-09-05T12:34:56Z" {
		t.Fatalf("target recurrence after caller mutation = %v, want original canonical recurrence", first)
	}
	*first = "2026-09-07T12:34:56Z"
	second := target.RecurrenceID()
	if second == nil || *second != "2026-09-05T12:34:56Z" {
		t.Fatalf("target recurrence after accessor mutation = %v, want original canonical recurrence", second)
	}
}

func TestMailboxCalendarRSVPTargetRejectsInvalidShapes(t *testing.T) {
	eventID := uuid.MustParse("55555555-5555-4555-8555-555555555555")
	seriesRecurrence := "2026-09-05"
	offsetRecurrence := "2026-09-05T12:34:56+03:00"
	for _, test := range []struct {
		name         string
		eventID      uuid.UUID
		scope        mail.MailboxCalendarRSVPTargetScope
		recurrenceID *string
	}{
		{name: "nil event", eventID: uuid.Nil, scope: mail.MailboxCalendarRSVPTargetScopeSeries},
		{name: "series recurrence", eventID: eventID, scope: mail.MailboxCalendarRSVPTargetScopeSeries, recurrenceID: &seriesRecurrence},
		{name: "occurrence without recurrence", eventID: eventID, scope: mail.MailboxCalendarRSVPTargetScopeOccurrence},
		{name: "occurrence offset timestamp", eventID: eventID, scope: mail.MailboxCalendarRSVPTargetScopeOccurrence, recurrenceID: &offsetRecurrence},
		{name: "unknown scope", eventID: eventID, scope: mail.MailboxCalendarRSVPTargetScope("other")},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := mail.NewMailboxCalendarRSVPTarget(test.eventID, test.scope, test.recurrenceID, mailboxCalendarRSVPTargetAttendeeVersion, mailboxCalendarRSVPTargetRSVPStatus)
			requireMailboxCalendarRSVPTargetInvalid(t, err)
		})
	}

	_, err := mail.NewMailboxCalendarRSVPTarget(eventID, mail.MailboxCalendarRSVPTargetScopeSeries, nil, 0, mailboxCalendarRSVPTargetRSVPStatus)
	requireMailboxCalendarRSVPTargetInvalid(t, err)
	_, err = mail.NewMailboxCalendarRSVPTarget(eventID, mail.MailboxCalendarRSVPTargetScopeSeries, nil, mailboxCalendarRSVPTargetAttendeeVersion, mail.MailboxCalendarRSVPStatus("other"))
	requireMailboxCalendarRSVPTargetInvalid(t, err)
}

func TestMailboxCalendarCorrelationResultPreservesNavigationWithoutRSVPTarget(t *testing.T) {
	request := newMailboxCalendarRSVPTargetRequest(t)
	eventID := uuid.MustParse("55555555-5555-4555-8555-555555555555")
	result, err := mail.NewMailboxCalendarCorrelationResult(request, eventID, mail.MailboxCalendarCorrelationPathInboundReceipt)
	if err != nil {
		t.Fatalf("NewMailboxCalendarCorrelationResult() error = %v", err)
	}
	if result.CalendarEventID() != eventID || result.CalendarRSVPTarget() != nil {
		t.Fatalf("navigation-only correlation = %#v, want event navigation with nil RSVP target", result)
	}
}

func TestMailboxCalendarCorrelationResultBindsInboundSeriesRSVPTarget(t *testing.T) {
	request := newMailboxCalendarRSVPTargetRequest(t)
	eventID := uuid.MustParse("55555555-5555-4555-8555-555555555555")
	target, err := mail.NewMailboxCalendarRSVPTarget(eventID, mail.MailboxCalendarRSVPTargetScopeSeries, nil, mailboxCalendarRSVPTargetAttendeeVersion, mailboxCalendarRSVPTargetRSVPStatus)
	if err != nil {
		t.Fatalf("NewMailboxCalendarRSVPTarget() error = %v", err)
	}
	result, err := mail.NewMailboxCalendarCorrelationResultWithRSVPTarget(request, eventID, mail.MailboxCalendarCorrelationPathInboundReceipt, &target)
	if err != nil {
		t.Fatalf("NewMailboxCalendarCorrelationResultWithRSVPTarget() error = %v", err)
	}
	got := result.CalendarRSVPTarget()
	if result.CalendarEventID() != eventID || got == nil || got.CalendarEventID() != eventID ||
		got.Scope() != mail.MailboxCalendarRSVPTargetScopeSeries || got.RecurrenceID() != nil || got.AttendeeVersion() != mailboxCalendarRSVPTargetAttendeeVersion || got.RSVPStatus() != mailboxCalendarRSVPTargetRSVPStatus {
		t.Fatalf("CalendarRSVPTarget() = %#v, want bound series target for master-event navigation", got)
	}
}

func TestMailboxCalendarCorrelationResultCopiesBoundInboundRSVPTarget(t *testing.T) {
	request := newMailboxCalendarRSVPTargetRequest(t)
	eventID := uuid.MustParse("55555555-5555-4555-8555-555555555555")
	recurrenceID := "2026-09-05T12:34:56Z"
	target, err := mail.NewMailboxCalendarRSVPTarget(eventID, mail.MailboxCalendarRSVPTargetScopeOccurrence, &recurrenceID, mailboxCalendarRSVPTargetAttendeeVersion, mailboxCalendarRSVPTargetRSVPStatus)
	if err != nil {
		t.Fatalf("NewMailboxCalendarRSVPTarget() error = %v", err)
	}
	result, err := mail.NewMailboxCalendarCorrelationResultWithRSVPTarget(request, eventID, mail.MailboxCalendarCorrelationPathInboundReceipt, &target)
	if err != nil {
		t.Fatalf("NewMailboxCalendarCorrelationResultWithRSVPTarget() error = %v", err)
	}

	first := result.CalendarRSVPTarget()
	if first == nil || first.CalendarEventID() != eventID || first.Scope() != mail.MailboxCalendarRSVPTargetScopeOccurrence || first.AttendeeVersion() != mailboxCalendarRSVPTargetAttendeeVersion || first.RSVPStatus() != mailboxCalendarRSVPTargetRSVPStatus {
		t.Fatalf("CalendarRSVPTarget() = %#v, want bound occurrence target", first)
	}
	firstRecurrence := first.RecurrenceID()
	if firstRecurrence == nil || *firstRecurrence != recurrenceID {
		t.Fatalf("CalendarRSVPTarget().RecurrenceID() = %v, want %q", firstRecurrence, recurrenceID)
	}
	*firstRecurrence = "2026-09-06T12:34:56Z"
	second := result.CalendarRSVPTarget()
	secondRecurrence := second.RecurrenceID()
	if secondRecurrence == nil || *secondRecurrence != recurrenceID {
		t.Fatalf("CalendarRSVPTarget() leaked accessor mutation: got %v, want %q", secondRecurrence, recurrenceID)
	}
}

func TestMailboxCalendarCorrelationResultRejectsMismatchedOrOutboundRSVPTarget(t *testing.T) {
	request := newMailboxCalendarRSVPTargetRequest(t)
	eventID := uuid.MustParse("55555555-5555-4555-8555-555555555555")
	otherEventID := uuid.MustParse("66666666-6666-4666-8666-666666666666")
	target, err := mail.NewMailboxCalendarRSVPTarget(eventID, mail.MailboxCalendarRSVPTargetScopeSeries, nil, mailboxCalendarRSVPTargetAttendeeVersion, mailboxCalendarRSVPTargetRSVPStatus)
	if err != nil {
		t.Fatalf("NewMailboxCalendarRSVPTarget() error = %v", err)
	}
	otherTarget, err := mail.NewMailboxCalendarRSVPTarget(otherEventID, mail.MailboxCalendarRSVPTargetScopeSeries, nil, mailboxCalendarRSVPTargetAttendeeVersion, mailboxCalendarRSVPTargetRSVPStatus)
	if err != nil {
		t.Fatalf("NewMailboxCalendarRSVPTarget() error = %v", err)
	}

	for _, test := range []struct {
		name   string
		path   mail.MailboxCalendarCorrelationPath
		target *mail.MailboxCalendarRSVPTarget
	}{
		{name: "mismatched event", path: mail.MailboxCalendarCorrelationPathInboundReceipt, target: &otherTarget},
		{name: "outbound path", path: mail.MailboxCalendarCorrelationPathOutboundInvite, target: &target},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := mail.NewMailboxCalendarCorrelationResultWithRSVPTarget(request, eventID, test.path, test.target)
			requireMailboxCalendarRSVPTargetInvalid(t, err)
		})
	}
}
