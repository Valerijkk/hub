package mail_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

var _ mail.MailboxSubmissionWorkerDueScanner = (*mail.MailboxEntRepository)(nil)
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionWorkerDueScanRequest) (mail.MailboxSubmissionWorkerDueScanResult, error) = (*mail.MailboxEntRepository).ScanDueMailboxSubmissionIDs

func mailboxEntSubmissionWorkerSourceSlice(t *testing.T, source, start, end string) string {
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

func requireMailboxEntSubmissionWorkerOrdered(t *testing.T, source string, fragments ...string) {
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

func TestMailboxEntSubmissionWorkerDueScanSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_submission_worker.go")
	if err != nil {
		t.Fatalf("read submission worker Ent source: %v", err)
	}
	scan := mailboxEntSubmissionWorkerSourceSlice(t, string(sourceBytes),
		"func (repository *MailboxEntRepository) ScanDueMailboxSubmissionIDs",
		"func mailboxEntSubmissionWorkerDueScanRowValid",
	)

	requireMailboxEntSubmissionWorkerOrdered(t, scan,
		"request.Limit < 1 || request.Limit > 50",
		"repository.withMailboxEntTransaction",
		"mailboxEntDatabaseWallClock",
		"transaction.MailSubmission.Query()",
		"mailsubmission.StatusEQ(mailsubmission.Status(\"queued\"))",
		"mailsubmission.NextAttemptAtNotNil()",
		"mailsubmission.NextAttemptAtLTE(selectedAt)",
		"mailsubmission.CommandLeaseTokenIsNil()",
		"mailsubmission.CommandLeaseExpiresAtIsNil()",
		"mailsubmission.DispatchAuthorizedAtIsNil()",
		"mailsubmission.DispatchAuthorizationHashIsNil()",
		"mailsubmission.ProviderSubmissionIDIsNil()",
		"mailsubmission.ProviderMessageIDIsNil()",
		"mailsubmission.ResultMailMessageIDIsNil()",
		"mailsubmission.LastErrorCodeIsNil()",
		"mailsubmission.AttemptCountGTE(0)",
		"mailsubmission.AttemptCountLTE(7)",
		"mailsubmission.MailDraftIDNotNil()",
		"Order(ent.Asc(mailsubmission.FieldNextAttemptAt), ent.Asc(mailsubmission.FieldID))",
		"Limit(request.Limit)",
		"make([]uuid.UUID, 0, len(rows))",
		"mailboxEntSubmissionWorkerDueScanRowValid(submission, selectedAt)",
		"MailboxSubmissionWorkerDueScanResult{SubmissionIDs: submissionIDs}",
	)
	if got := strings.Count(scan, "mailboxEntDatabaseWallClock"); got != 1 {
		t.Fatalf("scan database clock reads = %d, want 1", got)
	}
	for _, required := range []string{
		"mailboxEntSubmissionWorkerDueScanResultValid(request, scanned)",
		"mailboxEntSubmissionWorkerDueScanResultCopy(scanned)",
		"mailboxEntSubmissionWorkerDueScanRepositoryError(err)",
	} {
		if !strings.Contains(scan, required) {
			t.Fatalf("due scan source is missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"ForUpdate()",
		"SkipLocked",
		".Update()",
		".Exec(",
		"BeginEngineCommand(",
		"MailEngine",
		".PreflightSubmission(",
		".Submit(",
		"http.",
		"net/http",
		"nats.",
	} {
		if strings.Contains(scan, forbidden) {
			t.Fatalf("due scan contains forbidden %q", forbidden)
		}
	}
}
