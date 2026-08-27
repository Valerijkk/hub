package mail_test

import (
	"os"
	"strings"
	"testing"
)

func calendarDeliveryEntBindSource(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile("../../internal/mail/mailbox_ent_calendar_delivery_bind.go")
	if err != nil {
		t.Fatalf("read Calendar delivery Ent bind source: %v", err)
	}
	return string(source)
}

func requireCalendarDeliveryEntBindSource(t *testing.T, source string, terms ...string) {
	t.Helper()
	for _, term := range terms {
		if !strings.Contains(source, term) {
			t.Fatalf("source is missing %q", term)
		}
	}
}

func TestCalendarDeliveryEntBindSourceContract(t *testing.T) {
	source := calendarDeliveryEntBindSource(t)
	requireCalendarDeliveryEntBindSource(t, source,
		"func (repository *MailboxEntRepository) BindCalendarMailDeliverySubmission",
		"repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || !validCalendarMailDeliverySubmissionBindRequest(request)",
		"repository.withMailboxEntTransaction",
		"mailboxEntBindCalendarInviteDeliverySubmission",
		"mailboxEntBindCalendarReplyDeliverySubmission",
		"mailboxEntDatabaseWallClock(ctx, transaction)",
		"transaction.CalendarInviteDelivery.Query()",
		"transaction.CalendarReplyDelivery.Query()",
		"transaction.MailSubmission.Query()",
		"ForUpdate()",
		"sha256.Sum256([]byte(calendarMailDeliverySubmissionKey(claim.deliveryID)))",
		"mailsubmission.IDEQ(submissionID)",
		"mailsubmission.MailAccountIDEQ(claim.accountID)",
		"mailsubmission.IdempotencyKeyHashEQ(idempotencyKeyHash[:]...)",
		"transaction.MailDraft.Query()",
		"maildraft.DraftClientIDEQ(calendarMailDeliveryDraftClientID(claim.deliveryID))",
		"mailboxEntCalendarInviteDeliveryFreshBindMatches",
		"mailboxEntCalendarReplyDeliveryFreshBindMatches",
		"delivery.LeaseUntil.After(databaseNow)",
		"mailboxEntCASCalendarInviteDeliveryBind",
		"mailboxEntCASCalendarReplyDeliveryBind",
		"calendarinvitedelivery.StateEQ(calendarinvitedelivery.State(CalendarMailDeliveryStateSubmitting))",
		"calendarreplydelivery.StateEQ(calendarreplydelivery.State(CalendarMailDeliveryStateSubmitting))",
		"MailSubmissionIDIsNil()",
		"AttemptsEQ(delivery.Attempts)",
		"ClaimGenerationEQ(int64(claim.claimGeneration))",
		"LeaseTokenEQ(uuid.UUID(claim.claimToken))",
		"LeaseUntilEQ(claim.leaseUntil)",
		"NextAttemptAtIsNil()",
		"LastErrorCodeIsNil()",
		"calendarinvitedelivery.TerminalAtIsNil()",
		"ClearLeaseToken()",
		"ClearLeaseUntil()",
		"SetLastErrorCode(calendarinvitedelivery.LastErrorCode(",
		"SetLastErrorCode(calendarreplydelivery.LastErrorCode(",
		"submission_outcome_unknown",
		"SetNextAttemptAt(databaseNow)",
		"SetUpdatedAt(databaseNow)",
		"if updated != 1",
		"delivery.MailSubmissionID != nil && *delivery.MailSubmissionID != request.SubmissionID",
		"CalendarMailDeliverySubmissionBindOutcomeReplay",
		"CalendarMailDeliverySubmissionBindOutcomeConflict",
		"CalendarMailDeliverySubmissionBindOutcomeBound",
	)
}

func TestCalendarDeliveryEntBindSourceForbidsExternalWorkAndClaimMutation(t *testing.T) {
	source := calendarDeliveryEntBindSource(t)
	for _, forbidden := range []string{
		"time.Now(",
		"net/http",
		"http.",
		"MailEngine",
		"ClaimCalendarMailDelivery(",
		"ResolveCalendarMailDelivery(",
		"ScanDueCalendarMailDeliveries(",
		"SetMethod(",
		"SetOrganizer(",
		"SetContent(",
		"SetAddress(",
		"var _ CalendarMailDeliveryRepository",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("Bind source must not contain %q", forbidden)
		}
	}
}
