package mail_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCalendarDeliveryEntResolveSourceContract(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	sourcePath := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", "..", "internal", "mail", "mailbox_ent_calendar_delivery_resolve.go"))
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read calendar delivery Ent resolve source: %v", err)
	}
	text := string(source)

	for _, required := range []string{
		"func (repository *MailboxEntRepository) ResolveCalendarMailDelivery",
		"withMailboxEntTransaction",
		"mailboxEntDatabaseWallClock",
		"mailboxEntCalendarDeliveryResolveRetryDelay",
		"30 * time.Second",
		"calendarMailDeliveryRetryMaximum",
		"attempts - 1",
		"CalendarMailDeliveryResolveCodeMailboxUnavailable",
		"CalendarMailDeliveryResolveCodeInternal",
		"CalendarMailDeliveryResolveCodeInvalid",
		"CalendarMailDeliveryResolveCodeRejected",
		"CalendarMailDeliveryResolveCodePolicyDenied",
		"materialization_invalid",
		"submission_rejected",
		"policy_denied",
		"retry_exhausted",
		"ClearLeaseToken()",
		"ClearLeaseUntil()",
		"ClearNextAttemptAt()",
		"SetTerminalAt(databaseNow)",
		"mailboxEntCASCalendarInviteDeliveryResolve",
		"mailboxEntCASCalendarReplyDeliveryResolve",
		"AttemptsEQ(delivery.Attempts)",
		"ClaimGenerationEQ(int64(claim.claimGeneration))",
		"LeaseTokenEQ(uuid.UUID(claim.claimToken))",
		"LeaseUntilEQ(claim.leaseUntil)",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("calendar delivery Ent resolve source missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"time.Now",
		"MailEngine",
		"provider",
		"invalid_recipient",
		"BindCalendarMailDeliverySubmission",
		"ClaimCalendarMailDelivery",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("calendar delivery Ent resolve source contains forbidden %q", forbidden)
		}
	}

	inviteStart := strings.Index(text, "func mailboxEntCASCalendarInviteDeliveryResolve")
	replyStart := strings.Index(text, "func mailboxEntCASCalendarReplyDeliveryResolve")
	if inviteStart < 0 || replyStart < 0 || replyStart <= inviteStart {
		t.Fatal("calendar delivery Ent resolve CAS functions are missing or out of order")
	}
	inviteCAS := text[inviteStart:replyStart]
	replyCAS := text[replyStart:]
	if !strings.Contains(inviteCAS, "SetTerminalAt(databaseNow)") {
		t.Fatal("invite terminal resolve must set terminal_at from database time")
	}
	if strings.Contains(replyCAS, "TerminalAt") {
		t.Fatal("reply resolve must not address absent terminal_at")
	}
}
