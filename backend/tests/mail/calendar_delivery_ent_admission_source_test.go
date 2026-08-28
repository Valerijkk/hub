package mail_test

import (
	"os"
	"strings"
	"testing"
)

func calendarDeliveryEntAdmissionSource(t *testing.T) string {
	t.Helper()

	source, err := os.ReadFile("../../internal/mail/mailbox_ent_calendar_delivery_admission.go")
	if err != nil {
		t.Fatalf("read Calendar delivery Ent admission source: %v", err)
	}
	return string(source)
}

func calendarDeliveryEntAdmissionSlice(t *testing.T, source, start, end string) string {
	t.Helper()

	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("source is missing start marker %q", start)
	}
	endIndex := strings.Index(source[startIndex+len(start):], end)
	if endIndex < 0 {
		t.Fatalf("source is missing end marker %q", end)
	}
	return source[startIndex : startIndex+len(start)+endIndex]
}

func requireCalendarDeliveryEntAdmissionSource(t *testing.T, source string, terms ...string) {
	t.Helper()
	for _, term := range terms {
		if !strings.Contains(source, term) {
			t.Fatalf("source is missing %q", term)
		}
	}
}

func requireCalendarDeliveryEntAdmissionSourceOrder(t *testing.T, source string, terms ...string) {
	t.Helper()
	last := -1
	for _, term := range terms {
		index := strings.Index(source, term)
		if index < 0 {
			t.Fatalf("source is missing %q", term)
		}
		if index <= last {
			t.Fatalf("source puts %q before its required predecessor", term)
		}
		last = index
	}
}

func TestCalendarDeliveryEntAdmissionSourceContract(t *testing.T) {
	source := calendarDeliveryEntAdmissionSource(t)
	admit := calendarDeliveryEntAdmissionSlice(t, source,
		"func (repository *MailboxEntRepository) AdmitCalendarInviteDeliveries",
		"func mailboxEntAdmitCalendarInviteDelivery",
	)
	wake := calendarDeliveryEntAdmissionSlice(t, source,
		"func (repository *MailboxEntRepository) WakeCalendarReplyDelivery",
		"func mailboxEntWakeCalendarReplyDelivery",
	)
	admission := calendarDeliveryEntAdmissionSlice(t, source,
		"func mailboxEntAdmitCalendarInviteDelivery",
		"func mailboxEntCalendarInviteDeliveryRows",
	)
	rows := calendarDeliveryEntAdmissionSlice(t, source,
		"func mailboxEntCalendarInviteDeliveryRows",
		"func mailboxEntCalendarInviteDeliveryReplay",
	)
	replay := calendarDeliveryEntAdmissionSlice(t, source,
		"func mailboxEntCalendarInviteDeliveryReplay",
		"func mailboxEntCalendarInviteDeliveryImmutableMatch",
	)
	wakeRows := calendarDeliveryEntAdmissionSlice(t, source,
		"func mailboxEntWakeCalendarReplyDelivery",
		"func mailboxEntCalendarReplyDeliveryCalendarFactsValid",
	)

	requireCalendarDeliveryEntAdmissionSourceOrder(t, admit,
		"repository == nil || repository.client == nil || ctx == nil || !validCalendarInviteDeliveryAdmissionRequest(request)",
		"calendarMailDeliveryAdmissionRequestCopy(request)",
		"sort.SliceStable(request.Invites",
		"repository.withMailboxEntTransaction",
		"mailboxEntDatabaseWallClock(ctx, transaction)",
		"mailboxEntAdmitCalendarInviteDelivery",
	)
	requireCalendarDeliveryEntAdmissionSourceOrder(t, wake,
		"repository == nil || repository.client == nil || ctx == nil || !validCalendarReplyDeliveryWakeRequest(request)",
		"repository.withMailboxEntTransaction",
		"mailboxEntDatabaseWallClock(ctx, transaction)",
		"mailboxEntWakeCalendarReplyDelivery",
	)
	requireCalendarDeliveryEntAdmissionSource(t, admission,
		"transaction.CalendarEvent.Query()",
		"calendarevent.IDEQ(admission.EventID)",
		"ForUpdate()",
		"transaction.CalendarAttendee.Query()",
		"calendarattendee.IDEQ(admission.AttendeeID)",
		"calendarattendee.CalendarEventIDEQ(event.ID)",
		"mailboxEntCalendarInviteDeliveryReplay",
		"transaction.CalendarInviteDelivery.Create()",
		"SetID(admission.InviteID.UUID())",
		"SetOrganizationID(event.OrganizationID)",
		"SetCalendarEventID(event.ID)",
		"SetCalendarAttendeeID(attendee.ID)",
		"SetSourceEnvelopeID(admission.SourceEnvelopeID)",
		"SetKind(calendarinvitedelivery.Kind(admission.Kind))",
		"SetEventSequence(admission.EventSequence)",
		"SetState(calendarinvitedelivery.State(\"pending\"))",
		"SetAttempts(0)",
		"SetNextAttemptAt(databaseNow)",
	)
	requireCalendarDeliveryEntAdmissionSource(t, rows,
		"calendarinvitedelivery.IDEQ(admission.InviteID.UUID())",
		"calendarinvitedelivery.SourceEnvelopeIDEQ(admission.SourceEnvelopeID)",
		"calendarinvitedelivery.CalendarAttendeeIDEQ(attendee.ID)",
		"calendarinvitedelivery.CalendarEventIDEQ(event.ID)",
		"calendarinvitedelivery.EventSequenceEQ(admission.EventSequence)",
		"calendarinvitedelivery.KindEQ(calendarinvitedelivery.Kind(admission.Kind))",
		"calendarinvitedelivery.RecurrenceIDIsNil()",
		"calendarinvitedelivery.RecurrenceIDEQ(admission.RecurrenceID)",
		"ForUpdate()",
	)
	requireCalendarDeliveryEntAdmissionSource(t, replay,
		"len(candidates) != 1",
		"mailboxEntCalendarInviteDeliveryImmutableMatch",
		"CalendarMailDeliveryRepositoryErrorConflict",
	)
	requireCalendarDeliveryEntAdmissionSource(t, wakeRows,
		"transaction.CalendarReplyDelivery.Query()",
		"calendarreplydelivery.IDEQ(request.ReplyID.UUID())",
		"ForUpdate()",
		"transaction.CalendarEvent.Query()",
		"transaction.CalendarAttendee.Query()",
		"CalendarMailDeliveryStateSent",
		"CalendarMailDeliveryStateSuperseded",
		"CalendarMailDeliveryStateFailedPermanent",
		"CalendarMailDeliveryStateUnknown",
		"CalendarMailDeliveryStateSubmitting",
		"mailboxEntCalendarReplyDeliveryScheduledPendingValid(delivery)",
		"mailboxEntCalendarReplyDeliveryInitialPendingUnscheduledValid(delivery)",
		"calendarreplydelivery.NextAttemptAtIsNil()",
		"calendarreplydelivery.AttemptsEQ(0)",
		"calendarreplydelivery.ClaimGenerationEQ(0)",
		"calendarreplydelivery.MailSubmissionIDIsNil()",
		"calendarreplydelivery.LeaseTokenIsNil()",
		"calendarreplydelivery.LeaseUntilIsNil()",
		"calendarreplydelivery.LastErrorCodeIsNil()",
		"SetNextAttemptAt(databaseNow)",
	)
	requireCalendarDeliveryEntAdmissionSource(t, source,
		"attendee.CalendarEventID != event.ID",
		"event.DeletedAt != nil",
		"event.OrganizerID == nil || *event.OrganizerID != admission.OrganizerID",
		"event.Sequence < admission.EventSequence",
		"CalendarInviteDeliveryKindCreated, CalendarInviteDeliveryKindUpdated",
		"return attendee.RemovedAt == nil",
		"mailboxEntCalendarDeliveryRecurrenceMatches",
		"delivery.OrganizationID == event.OrganizationID",
		"delivery.CalendarEventID != event.ID",
		"delivery.CalendarAttendeeID != attendee.ID",
		"mailboxEntCalendarDeliveryRecurrencePointersEqual",
	)

	for _, forbidden := range []string{
		"time.Now",
		"net/http",
		"http.",
		"github.com/nats-io/nats.go",
		"stalwart.",
		"Publish(",
		"SetMethod(",
		"SetOrganizerID(",
		"SetContent(",
		"SetRecipient(",
		"ScanDueCalendarMailDeliveries",
		"ClaimCalendarMailDelivery",
		"BindCalendarMailDeliverySubmission",
		"ResolveCalendarMailDelivery",
		"NextAttemptAtGT(",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("source contains forbidden behavior or payload field %q", forbidden)
		}
	}
}
