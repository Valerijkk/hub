package mail_test

import (
	"os"
	"strings"
	"testing"
)

func TestCalendarDeliveryEntClaimSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_calendar_delivery_claim.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	for _, term := range []string{
		"repository.withMailboxEntTransaction",
		"mailboxEntDatabaseWallClock",
		"delivery.RecurrenceID != nil",
		"calendarattendee.RecurrenceIDIsNil()",
		"mailidentity.KindEQ(mailidentity.KindPrimary)",
		"mailidentity.SendAllowedEQ(true)",
		"ical.NormalizeITIPMessage",
		"ical.CanonicalITIPIdentity",
		"mailboxEntCheckedTimeAdd(databaseNow, calendarMailDeliveryLease)",
		"SetClaimGeneration(int64(generation))",
		"SetLeaseToken(uuid.UUID(token))",
		"SetLeaseUntil(leaseUntil)",
		"SetAttempts(delivery.Attempts + 1)",
		"mailboxEntObserveCalendarInviteDeliveryBoundSubmission",
		"mailboxEntObserveCalendarReplyDeliveryBoundSubmission",
		"delivery.NextAttemptAt.After(databaseNow)",
	} {
		if !strings.Contains(source, term) {
			t.Fatalf("missing %q", term)
		}
	}
	for _, forbidden := range []string{
		"time.Now",
		"net/http",
		"http.",
		"github.com/nats-io/nats.go",
		"stalwart.",
		"Publish(",
		"BindCalendarMailDeliverySubmission",
		"ResolveCalendarMailDelivery",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("forbidden %q", forbidden)
		}
	}
}

func TestCalendarDeliveryEntBoundObservationSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_calendar_delivery_observation.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	for _, term := range []string{
		"mailboxEntObserveCalendarInviteDeliveryBoundSubmission",
		"mailboxEntObserveCalendarReplyDeliveryBoundSubmission",
		"mailboxEntCalendarDeliveryBoundSubmissionForObservation",
		"sha256.Sum256([]byte(calendarMailDeliverySubmissionKey(deliveryID)))",
		"mailsubmission.IDEQ(submissionID)",
		"mailsubmission.IdempotencyKeyHashEQ(idempotencyKeyHash[:]...)",
		"transaction.MailSubmission.Query()",
		"ForUpdate()",
		"mailboxEntSubmissionRowValid(submission, submission.MailAccountID)",
		"MailboxSubmissionStatusQueued",
		"MailboxSubmissionStatusAuthorizing",
		"MailboxSubmissionStatusSubmitting",
		"MailboxSubmissionStatusAccepted",
		"MailboxSubmissionStatusUnknown",
		"MailboxSubmissionStatusSent",
		"MailboxSubmissionStatusFailed",
		"MailboxSubmissionErrorCodeOutcomeUnknown",
		"MailboxSubmissionErrorCodeReconciliationRequired",
		"MailboxSubmissionErrorCodeRejected",
		"MailboxSubmissionErrorCodeSizeLimit",
		"MailboxSubmissionErrorCodePolicy",
		"MailboxSubmissionErrorCodeAttachmentNotClean",
		"MailboxSubmissionErrorCodeIdentityUnavailable",
		"MailboxSubmissionErrorCodeRecipientLimit",
		"MailboxSubmissionErrorCodeRetryExhausted",
		"MailboxSubmissionErrorCodeReconciliationExhausted",
		"MailboxSubmissionErrorCodeMailboxUnavailable",
		"submission_rejected",
		"policy_denied",
		"materialization_invalid",
		"retry_exhausted",
		"mailboxEntCheckedTimeAdd(databaseNow, mailboxEntCalendarDeliveryObservationDelay)",
		"calendarinvitedelivery.StateEQ(calendarinvitedelivery.State(CalendarMailDeliveryStateUnknown))",
		"calendarreplydelivery.StateEQ(calendarreplydelivery.State(CalendarMailDeliveryStateUnknown))",
		"MailSubmissionIDEQ(submission.ID)",
		"NextAttemptAtLTE(databaseNow)",
		"ClearNextAttemptAt()",
		"ClearLastErrorCode()",
		"SetTerminalAt(databaseNow)",
		"CalendarMailDeliveryClaimOutcomeObserved",
		"if updated != 1",
	} {
		if !strings.Contains(source, term) {
			t.Fatalf("missing observation term %q", term)
		}
	}
	for _, forbidden := range []string{
		"time.Now",
		"net/http",
		"http.",
		"github.com/nats-io/nats.go",
		"stalwart.",
		"Publish(",
		"ClaimCalendarMailDelivery(",
		"BindCalendarMailDeliverySubmission(",
		"ResolveCalendarMailDelivery(",
		"SetAttempts(",
		"var _ CalendarMailDeliveryRepository",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("forbidden observation term %q", forbidden)
		}
	}
}

func TestCalendarDeliveryEntBoundObservationTerminalShapes(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_calendar_delivery_observation.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	inviteStart := strings.Index(source, "func mailboxEntCASCalendarInviteDeliveryObservation")
	replyStart := strings.Index(source, "func mailboxEntCASCalendarReplyDeliveryObservation")
	if inviteStart < 0 || replyStart < 0 || replyStart <= inviteStart {
		t.Fatal("missing ordered invite/reply observation CAS functions")
	}
	if !strings.Contains(source[inviteStart:replyStart], "SetTerminalAt(databaseNow)") {
		t.Fatal("invite observation must terminalize durable sent and permanent outcomes")
	}
	if strings.Contains(source[replyStart:], "SetTerminalAt(") {
		t.Fatal("reply observation must not write an invite-only terminal timestamp")
	}
}

func TestCalendarDeliveryEntRemovedAttendeeMasterCancelSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_calendar_delivery_claim.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	for _, term := range []string{
		"func mailboxEntCalendarInviteDeliveryFactsMatch",
		"event.Status == calendarevent.StatusCancelled && attendee.RemovedAt == nil",
		"attendee.RemovedAt != nil && (event.Status == calendarevent.StatusConfirmed || event.Status == calendarevent.StatusTentative)",
		"var targetParticipant ical.ITIPParticipant",
		"targetParticipantOK = true",
		"if delivery.Kind == calendarinvitedelivery.KindCancelled && target.RemovedAt != nil",
		"mailboxEntCalendarDeliveryRemovedAttendeeMasterCancelITIPMessage",
		"targetParticipant.RSVP",
		"func mailboxEntCalendarDeliveryRemovedAttendeeMasterCancelITIPMessage",
		"Method:      ical.ITIPMethodCancel",
		"Status:      ical.ITIPStatusCancelled",
		"Attendees:   []ical.ITIPParticipant{targetParticipant}",
		"RRULE:       mailboxEntCalendarEventRRULE(event)",
		"EXDates:     exdates",
		"ical.NormalizeITIPMessage(message)",
		"ical.CanonicalITIPIdentity(normalized)",
		"identity.HasRecurrenceID || identity.RecurrenceID !=",
		"len(normalized.Attendees) != 1",
	} {
		if !strings.Contains(source, term) {
			t.Fatalf("missing removed-attendee CANCEL term %q", term)
		}
	}
	factsStart := strings.Index(source, "func mailboxEntCalendarInviteDeliveryFactsMatch")
	methodStart := strings.Index(source, "func mailboxEntCalendarInviteDeliveryMethod")
	if factsStart < 0 || methodStart <= factsStart {
		t.Fatal("missing ordered invite fact matcher")
	}
	facts := source[factsStart:methodStart]
	for _, forbidden := range []string{
		"event.Status == calendarevent.StatusConfirmed && attendee.RemovedAt == nil",
		"event.Status == calendarevent.StatusTentative && attendee.RemovedAt == nil",
	} {
		if strings.Contains(facts, forbidden) {
			t.Fatalf("active attendee must not receive a partial-event CANCEL: %q", forbidden)
		}
	}
}
