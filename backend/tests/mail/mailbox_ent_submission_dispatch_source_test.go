package mail_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

var _ mail.MailboxSubmissionDispatchRepository = (*mail.MailboxEntRepository)(nil)
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionDispatchResolutionRequest) error = (*mail.MailboxEntRepository).ResolveMailboxSubmissionDispatch

func mailboxEntSubmissionDispatchSourceSlice(t *testing.T, source, start, end string) string {
	t.Helper()

	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("source is missing start marker %q", start)
	}
	endOffset := strings.Index(source[startIndex+len(start):], end)
	if endOffset < 0 {
		t.Fatalf("source is missing end marker %q after %q", end, start)
	}
	return source[startIndex : startIndex+len(start)+endOffset]
}

func requireMailboxEntSubmissionDispatchOrdered(t *testing.T, source string, fragments ...string) {
	t.Helper()

	previous := -1
	for _, fragment := range fragments {
		position := strings.Index(source, fragment)
		if position < 0 {
			t.Fatalf("source is missing %q", fragment)
		}
		if position <= previous {
			t.Fatalf("source puts %q before its required predecessor", fragment)
		}
		previous = position
	}
}

func TestMailboxEntSubmissionDispatchSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_submission_dispatch.go")
	if err != nil {
		t.Fatalf("read submission dispatch source: %v", err)
	}
	resolve := mailboxEntSubmissionDispatchSourceSlice(t, string(sourceBytes),
		"func (repository *MailboxEntRepository) ResolveMailboxSubmissionDispatch",
		"func mailboxEntSubmissionDispatchRequestValid",
	)

	requireMailboxEntSubmissionDispatchOrdered(t, resolve,
		"transaction.MailAccount.Query()",
		"transaction.MailSubmission.Query()",
		"transaction.MailDraft.Query()",
		"mailboxEntDatabaseWallClock",
		"mailboxEntSubmissionDispatchFenceMatches",
		"SetStatus(mailsubmission.Status(\"accepted\"))",
		"SetStatus(mailsubmission.Status(\"unknown\"))",
		"SetReconcileAttempts(0)",
		"SetNextReconcileAt(databaseNow)",
		"SetMutationEpoch(nextEpoch)",
	)
	for _, required := range []string{
		"subtle.ConstantTimeCompare",
		"submission.EngineMessageID",
		"ClearCommandLeaseToken()",
		"ClearMutationLeaseToken()",
		"SetProviderSubmissionID(request.ProviderSubmissionID)",
		"SetProviderMessageID(request.ProviderMessageID)",
		"SetLastErrorCode(mailsubmission.LastErrorCode(MailboxSubmissionErrorCodeOutcomeUnknown))",
		"submission.ReconcileAttempts != 0 || submission.NextReconcileAt != nil",
	} {
		if !strings.Contains(resolve, required) {
			t.Fatalf("dispatch resolution source is missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"SetStatus(mailsubmission.Status(\"queued\"))",
		"BeginEngineCommand(",
		"MailEngine",
		".PreflightSubmission(",
		".Submit(",
		"http.",
		"net/http",
		"nats.",
	} {
		if strings.Contains(resolve, forbidden) {
			t.Fatalf("dispatch resolution path contains forbidden %q", forbidden)
		}
	}
	if !strings.Contains(string(sourceBytes), `subtle.ConstantTimeCompare(*submission.DispatchAuthorizationHash, authorization.authorizationHash[:]) != 1`) {
		t.Fatal(`dispatch resolution must bind the stored authorization hash to the sealed capability`)
	}
}
