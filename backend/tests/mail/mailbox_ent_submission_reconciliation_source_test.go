package mail_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

var _ mail.MailboxSubmissionReconciliationRepository = (*mail.MailboxEntRepository)(nil)
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionReconciliationScanRequest) (mail.MailboxSubmissionReconciliationScanResult, error) = (*mail.MailboxEntRepository).ScanUnknownMailboxSubmissions
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionReconciliationClaimRequest) (mail.MailboxSubmissionReconciliationClaimResult, error) = (*mail.MailboxEntRepository).ClaimUnknownMailboxSubmission
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionReconciliationResolveRequest) error = (*mail.MailboxEntRepository).ResolveUnknownMailboxSubmissionReconciliation

func mailboxEntSubmissionReconciliationSourceSlice(t *testing.T, source, start, end string) string {
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

func requireMailboxEntSubmissionReconciliationOrdered(t *testing.T, source string, fragments ...string) {
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

func requireMailboxEntSubmissionReconciliationAbsent(t *testing.T, source string, fragments ...string) {
	t.Helper()

	for _, fragment := range fragments {
		if strings.Contains(source, fragment) {
			t.Fatalf("source contains forbidden %q", fragment)
		}
	}
}

func TestMailboxEntSubmissionReconciliationScanSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_submission_reconciliation.go")
	if err != nil {
		t.Fatalf("read submission reconciliation source: %v", err)
	}
	scan := mailboxEntSubmissionReconciliationSourceSlice(t, string(sourceBytes),
		"func (repository *MailboxEntRepository) ScanUnknownMailboxSubmissions",
		"func (repository *MailboxEntRepository) ClaimUnknownMailboxSubmission",
	)

	requireMailboxEntSubmissionReconciliationOrdered(t, scan,
		"repository.withMailboxEntTransaction",
		"mailboxEntDatabaseWallClock",
		"transaction.MailSubmission.Query()",
		"Order(ent.Asc(mailsubmission.FieldNextReconcileAt), ent.Asc(mailsubmission.FieldID))",
		"Limit(request.Limit)",
	)
	for _, required := range []string{
		"mailsubmission.Status(\"unknown\")",
		"mailsubmission.NextReconcileAtNotNil()",
		"mailsubmission.NextReconcileAtLTE(databaseNow)",
		"mailsubmission.CommandLeaseTokenIsNil()",
		"mailsubmission.CommandLeaseExpiresAtIsNil()",
		"mailsubmission.DispatchAuthorizedAtNotNil()",
		"mailsubmission.DispatchAuthorizationHashNotNil()",
		"mailsubmission.ProviderSubmissionIDIsNil()",
		"mailsubmission.ProviderMessageIDIsNil()",
		"mailsubmission.ResultMailMessageIDIsNil()",
		"mailboxEntSubmissionReconciliationScanRowValid",
		"mailboxSubmissionReconciliationScanResultCopy",
	} {
		if !strings.Contains(scan, required) {
			t.Fatalf("scan source is missing %q", required)
		}
	}
	requireMailboxEntSubmissionReconciliationAbsent(t, scan,
		"ForUpdate()", "Update()", "BeginEngineCommand(", "MailEngine", ".ReconcileSubmission(",
		".PreflightSubmission(", ".Submit(", "provider.", "nats.", "http.", "net/http", "time.Now",
	)
}

func TestMailboxEntSubmissionReconciliationClaimSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_submission_reconciliation.go")
	if err != nil {
		t.Fatalf("read submission reconciliation source: %v", err)
	}
	claim := mailboxEntSubmissionReconciliationSourceSlice(t, string(sourceBytes),
		"func (repository *MailboxEntRepository) ClaimUnknownMailboxSubmission",
		"func (repository *MailboxEntRepository) ResolveUnknownMailboxSubmissionReconciliation",
	)

	requireMailboxEntSubmissionReconciliationOrdered(t, claim,
		"repository.client.MailSubmission.Query()",
		"repository.withMailboxEntTransaction",
		"transaction.MailAccount.Query()",
		"transaction.MailSubmission.Query()",
		"mailboxEntDatabaseWallClock",
		"transaction.MailDraft.Query()",
		"SetMutationLeaseToken(leaseID)",
		"SetCommandLeaseToken(leaseID)",
		"NewMailboxSubmissionReconciliationClaimedResult",
	)
	for _, required := range []string{
		"MailboxEngineCommandLeaseDuration",
		"mailboxEntCheckedTimeAdd",
		"uuid.New()",
		"mailboxEntSubmissionReconciliationOwnedLeaseBusy",
		"mailboxEntSubmissionClaimBusyUntil",
		"MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusUnknown",
		"submission.ReconcileAttempts < 0 || submission.ReconcileAttempts > 9",
		"submission.NextReconcileAt == nil || submission.NextReconcileAt.After(databaseNow)",
		"mailboxEntSubmissionReconciliationPendingCode",
		"mailboxEntSubmissionClaimDraftMatches",
		"SetMutationEpoch(nextEpoch)",
		"SetCommandLeaseExpiresAt(leaseUntil)",
		"SetCommandFence(nextEpoch)",
		"SetReconcileAttempts(submission.ReconcileAttempts + 1)",
		"ClearNextReconcileAt()",
		"SetUpdatedAt(databaseNow)",
	} {
		if !strings.Contains(claim, required) {
			t.Fatalf("claim source is missing %q", required)
		}
	}
	requireMailboxEntSubmissionReconciliationAbsent(t, claim,
		"BeginEngineCommand(", "MailEngine", ".ReconcileSubmission(", ".PreflightSubmission(", ".Submit(",
		"provider.", "nats.", "http.", "net/http", "time.Now",
	)
}

func TestMailboxEntSubmissionReconciliationResolveSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_submission_reconciliation.go")
	if err != nil {
		t.Fatalf("read submission reconciliation source: %v", err)
	}
	source := string(sourceBytes)
	resolve := mailboxEntSubmissionReconciliationSourceSlice(t, source,
		"func (repository *MailboxEntRepository) ResolveUnknownMailboxSubmissionReconciliation",
		"func mailboxEntSubmissionReconciliationScanRowValid",
	)

	requireMailboxEntSubmissionReconciliationOrdered(t, resolve,
		"transaction.MailAccount.Query()",
		"transaction.MailSubmission.Query()",
		"transaction.MailDraft.Query()",
		"mailboxEntDatabaseWallClock",
		"mailboxEntSubmissionReconciliationFenceMatches",
		"MailboxSubmissionReconciliationObservationAccepted",
		"MailboxSubmissionReconciliationObservationRetry",
		"SetMutationEpoch(nextEpoch)",
	)
	for _, required := range []string{
		"subtle.ConstantTimeCompare",
		"request.Claim.submission.MessageID",
		"SetStatus(mailsubmission.Status(\"accepted\"))",
		"SetProviderSubmissionID(request.EngineSubmissionID)",
		"SetProviderMessageID(request.EngineMessageID)",
		"ClearCommandLeaseToken()",
		"ClearCommandLeaseExpiresAt()",
		"ClearLastErrorCode()",
		"ClearNextAttemptAt()",
		"ClearNextReconcileAt()",
		"mailboxEntSubmissionReconciliationBackoff",
		"SetNextReconcileAt(retryAt)",
		"MailboxSubmissionErrorCodeReconciliationExhausted",
		"ClearMutationLeaseToken()",
		"ClearMutationLeaseUntil()",
		"SetSyncScheduleUpdatedAt(databaseNow)",
	} {
		if !strings.Contains(resolve, required) {
			t.Fatalf("resolve source is missing %q", required)
		}
	}
	fence := mailboxEntSubmissionReconciliationSourceSlice(t, source,
		"func mailboxEntSubmissionReconciliationFenceMatches",
		"func mailboxEntSubmissionReconciliationBackoff",
	)
	for _, required := range []string{
		"request.Claim.fence.LeaseID",
		"request.Claim.fence.LeaseUntil",
		"request.Claim.fence.MutationEpoch",
		"submission.EngineAccountID",
		"mailboxEntSubmissionClaimDraftMatches",
		"submission.ReconcileAttempts < 1 || submission.ReconcileAttempts > 10",
	} {
		if !strings.Contains(fence, required) {
			t.Fatalf("resolve fence source is missing %q", required)
		}
	}
	backoff := mailboxEntSubmissionReconciliationSourceSlice(t, source,
		"func mailboxEntSubmissionReconciliationBackoff",
		"func mailboxEntSubmissionReconciliationRepositoryError",
	)
	for _, required := range []string{
		"case 1:", "time.Minute", "case 2:", "5 * time.Minute", "case 3:", "15 * time.Minute",
		"case 4:", "30 * time.Minute", "case 5:", "time.Hour", "case 6:", "2 * time.Hour",
		"case 7:", "4 * time.Hour", "case 8, 9:", "6 * time.Hour",
	} {
		if !strings.Contains(backoff, required) {
			t.Fatalf("resolve backoff source is missing %q", required)
		}
	}
	requireMailboxEntSubmissionReconciliationAbsent(t, resolve,
		"SetStatus(mailsubmission.Status(\"queued\"))", "draft.Update()", "BeginEngineCommand(", "MailEngine",
		".ReconcileSubmission(", ".PreflightSubmission(", ".Submit(", "provider.", "nats.", "http.", "net/http", "time.Now",
	)
}
