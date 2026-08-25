package mail_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

var _ mail.MailboxSubmissionStaleRecoveryRepository = (*mail.MailboxEntRepository)(nil)
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionRecoveryScanRequest) (mail.MailboxSubmissionRecoveryScanResult, error) = (*mail.MailboxEntRepository).ScanDueMailboxSubmissionRecovery
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionStaleRecoveryRequest) (mail.MailboxSubmissionStaleRecoveryResult, error) = (*mail.MailboxEntRepository).RecoverStaleMailboxSubmission

func mailboxEntSubmissionRecoverySourceSlice(t *testing.T, source, start, end string) string {
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

func requireMailboxEntSubmissionRecoveryOrdered(t *testing.T, source string, fragments ...string) {
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

func requireMailboxEntSubmissionRecoveryContains(t *testing.T, source string, fragments ...string) {
	t.Helper()

	for _, fragment := range fragments {
		if !strings.Contains(source, fragment) {
			t.Fatalf("source is missing %q", fragment)
		}
	}
}

func requireMailboxEntSubmissionRecoveryForbids(t *testing.T, source string, fragments ...string) {
	t.Helper()

	for _, fragment := range fragments {
		if strings.Contains(source, fragment) {
			t.Fatalf("source contains forbidden %q", fragment)
		}
	}
}

func TestMailboxEntSubmissionRecoverySourceContract(t *testing.T) {
	entSourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_submission_recovery.go")
	if err != nil {
		t.Fatalf("read Ent submission recovery source: %v", err)
	}
	entSource := string(entSourceBytes)
	contractSourceBytes, err := os.ReadFile("../../internal/mail/mailbox_submission_recovery.go")
	if err != nil {
		t.Fatalf("read recovery contract source: %v", err)
	}
	contractSource := string(contractSourceBytes)

	scan := mailboxEntSubmissionRecoverySourceSlice(t, entSource,
		"func (repository *MailboxEntRepository) ScanDueMailboxSubmissionRecovery",
		"func (repository *MailboxEntRepository) RecoverStaleMailboxSubmission",
	)
	recover := mailboxEntSubmissionRecoverySourceSlice(t, entSource,
		"func (repository *MailboxEntRepository) RecoverStaleMailboxSubmission",
		"func mailboxEntSubmissionRecoveryScanRequestValid",
	)
	skipLocked := mailboxEntSubmissionRecoverySourceSlice(t, entSource,
		"func mailboxEntSubmissionRecoverySkipLocked",
		"func mailboxEntSubmissionRecoveryDueFromRow",
	)
	requeue := mailboxEntSubmissionRecoverySourceSlice(t, entSource,
		"func mailboxEntSubmissionRecoveryRequeue",
		"func mailboxEntSubmissionRecoveryTerminal",
	)
	terminal := mailboxEntSubmissionRecoverySourceSlice(t, entSource,
		"func mailboxEntSubmissionRecoveryTerminal",
		"func mailboxEntSubmissionRecoveryUnknown",
	)
	unknown := mailboxEntSubmissionRecoverySourceSlice(t, entSource,
		"func mailboxEntSubmissionRecoveryUnknown",
		"func mailboxEntSubmissionRecoveryReleaseUnknown",
	)
	unknownLease := mailboxEntSubmissionRecoverySourceSlice(t, entSource,
		"func mailboxEntSubmissionRecoveryReleaseUnknown",
		"func mailboxEntSubmissionRecoveryPendingCode",
	)
	accepted := mailboxEntSubmissionRecoverySourceSlice(t, entSource,
		"func mailboxEntSubmissionRecoveryReleaseAccepted",
		"func mailboxEntSubmissionRecoveryReleaseExactAccountLease",
	)
	exactAccountLease := mailboxEntSubmissionRecoverySourceSlice(t, entSource,
		"func mailboxEntSubmissionRecoveryReleaseExactAccountLease",
		"func mailboxEntSubmissionRecoveryDraftError",
	)
	leaseState := mailboxEntSubmissionRecoverySourceSlice(t, entSource,
		"func mailboxEntSubmissionRecoveryLeaseState",
		"func mailboxEntSubmissionRecoveryReleasedStateValid",
	)
	releasedState := mailboxEntSubmissionRecoverySourceSlice(t, entSource,
		"func mailboxEntSubmissionRecoveryReleasedStateValid",
		"func mailboxEntSubmissionRecoveryCurrentStateValid",
	)
	releasedNoop := mailboxEntSubmissionRecoverySourceSlice(t, recover,
		"case mailboxEntSubmissionRecoveryLeaseStateReleased:",
		"case mailboxEntSubmissionRecoveryLeaseStateHeld:",
	)
	copyResult := mailboxEntSubmissionRecoverySourceSlice(t, contractSource,
		"func mailboxSubmissionRecoveryScanResultCopy",
		"func mailboxSubmissionStaleRecoveryRequestValid",
	)

	requireMailboxEntSubmissionRecoveryOrdered(t, scan,
		"mailboxEntDatabaseWallClock",
		"transaction.MailSubmission.Query()",
		"mailboxEntSubmissionRecoverySkipLocked()",
		"mailsubmission.CommandLeaseExpiresAtLTE(selectedAt)",
		"Order(ent.Asc(mailsubmission.FieldCommandLeaseExpiresAt), ent.Asc(mailsubmission.FieldID))",
	)
	requireMailboxEntSubmissionRecoveryContains(t, scan,
		"mailsubmission.StatusIn(mailsubmission.Status(\"authorizing\"), mailsubmission.Status(\"submitting\"), mailsubmission.Status(\"unknown\"), mailsubmission.Status(\"accepted\"))",
		"mailboxEntSubmissionRecoveryScanResultValid(request, scanned)",
		"mailboxEntSubmissionRecoveryScanResultCopy(scanned)",
	)
	requireMailboxEntSubmissionRecoveryContains(t, skipLocked, "sql.WithLockAction(sql.SkipLocked)")
	requireMailboxEntSubmissionRecoveryForbids(t, scan,
		"BeginEngineCommand(", ".PreflightSubmission(", ".Submit(", "MailEngine", "provider.", "nats.", "http.", "net/http",
	)
	requireMailboxEntSubmissionRecoveryForbids(t, skipLocked,
		"BeginEngineCommand(", ".PreflightSubmission(", ".Submit(", "MailEngine", "provider.", "nats.", "http.", "net/http",
	)
	requireMailboxEntSubmissionRecoveryOrdered(t, copyResult,
		"submissions := result.Submissions",
		"make([]MailboxSubmissionRecoveryDue, len(submissions))",
		"copy(result.Submissions, submissions)",
	)
	requireMailboxEntSubmissionRecoveryForbids(t, copyResult,
		"append([]MailboxSubmissionRecoveryDue(nil)", "result.Submissions = nil", "submissions = nil",
	)
	requireMailboxEntSubmissionRecoveryContains(t, contractSource,
		"MailboxSubmissionStatusUnknown",
		"MailboxSubmissionStatusAccepted",
		"MailboxSubmissionStaleRecoveryOutcomeLeaseReleased",
		"return result.RetryAt == nil && result.FailureCode == nil",
		"func mailboxSubmissionStaleRecoveryLeaseReleased",
	)

	requireMailboxEntSubmissionRecoveryOrdered(t, recover,
		"transaction.MailAccount.Query()",
		"transaction.MailSubmission.Query()",
		"transaction.MailDraft.Query()",
		"mailboxEntDatabaseWallClock",
		"mailboxEntSubmissionRecoveryLeaseState",
		"mailboxEntSubmissionRecoveryReleasedStateValid",
		"mailboxEntSubmissionRecoveryCurrentStateValid",
		"mailboxEntSubmissionRecoveryDualFenceShapeMatches",
		"mailboxEntSubmissionRecoveryDualFenceMatches",
		"mailboxEntSubmissionRecoveryRequeue",
		"mailboxEntSubmissionRecoveryTerminal",
		"mailboxEntSubmissionRecoveryUnknown",
		"mailboxEntSubmissionRecoveryReleaseUnknown",
		"mailboxEntSubmissionRecoveryReleaseAccepted",
	)
	requireMailboxEntSubmissionRecoveryContains(t, recover,
		"mailboxSubmissionStaleRecoveryNoop()",
		"case MailboxSubmissionStatusUnknown:",
		"case MailboxSubmissionStatusAccepted:",
		"mailboxEntSubmissionRowValid(submission, account.ID)",
		"case mailboxEntSubmissionRecoveryLeaseStateReleased:",
		"case mailboxEntSubmissionRecoveryLeaseStateHeld:",
		"case mailboxEntSubmissionRecoveryLeaseStateInvalid:",
		"case mailboxEntSubmissionRecoveryLeaseStateConflict:",
		"mailboxEntSubmissionRecoveryCurrentStateValid",
	)
	requireMailboxEntSubmissionRecoveryForbids(t, recover,
		"BeginEngineCommand(", ".PreflightSubmission(", ".Submit(", "MailEngine", "provider.", "nats.", "http.", "net/http",
	)

	requireMailboxEntSubmissionRecoveryOrdered(t, requeue,
		"submission.Update()",
		"SetStatus(mailsubmission.Status(\"queued\"))",
		"ClearCommandLeaseToken()",
		"ClearDispatchAuthorizedAt()",
		"ClearDispatchAuthorizationHash()",
		"SetNextAttemptAt(databaseNow)",
		"account.Update()",
		"ClearMutationLeaseToken()",
		"SetMutationEpoch(nextEpoch)",
	)
	requireMailboxEntSubmissionRecoveryContains(t, requeue,
		"ClearProviderSubmissionID()", "ClearProviderMessageID()", "ClearResultMailMessageID()", "ClearLastErrorCode()", "SetAttemptCount(submission.AttemptCount)",
	)
	requireMailboxEntSubmissionRecoveryForbids(t, requeue,
		"SetStatus(mailsubmission.Status(\"submitting\"))", "SetStatus(mailsubmission.Status(\"unknown\"))", "SetStatus(mailsubmission.Status(\"failed\"))",
		"SetProviderSubmissionID(", "SetProviderMessageID(", "mailboxEntSubmissionRecoveryTerminal(", "mailboxEntSubmissionRecoveryUnknown(", "BeginEngineCommand(", ".PreflightSubmission(", ".Submit(", "MailEngine", "provider.", "nats.", "http.", "net/http",
	)

	requireMailboxEntSubmissionRecoveryOrdered(t, terminal,
		"submission.Update()",
		"SetStatus(mailsubmission.Status(\"failed\"))",
		"SetLastErrorCode(mailsubmission.LastErrorCode(MailboxSubmissionErrorCodeRetryExhausted))",
		"draft.Update()",
		"SetStatus(maildraft.Status(\"active\"))",
		"account.Update()",
		"ClearMutationLeaseToken()",
		"SetMutationEpoch(nextEpoch)",
	)
	requireMailboxEntSubmissionRecoveryContains(t, terminal,
		"ClearDispatchAuthorizedAt()", "ClearDispatchAuthorizationHash()", "ClearProviderSubmissionID()", "ClearProviderMessageID()", "ClearResultMailMessageID()", "ClearNextAttemptAt()",
	)
	requireMailboxEntSubmissionRecoveryForbids(t, terminal,
		"SetStatus(mailsubmission.Status(\"queued\"))", "SetStatus(mailsubmission.Status(\"submitting\"))", "SetStatus(mailsubmission.Status(\"unknown\"))",
		"mailboxEntSubmissionRecoveryRequeue(", "mailboxEntSubmissionRecoveryUnknown(", "BeginEngineCommand(", ".PreflightSubmission(", ".Submit(", "MailEngine", "provider.", "nats.", "http.", "net/http",
	)

	requireMailboxEntSubmissionRecoveryOrdered(t, unknown,
		"submission.Update()",
		"SetStatus(mailsubmission.Status(\"unknown\"))",
		"ClearCommandLeaseToken()",
		"ClearCommandLeaseExpiresAt()",
		"ClearProviderSubmissionID()",
		"ClearProviderMessageID()",
		"ClearResultMailMessageID()",
		"SetLastErrorCode(mailsubmission.LastErrorCode(MailboxSubmissionErrorCodeReconciliationRequired))",
		"SetReconcileAttempts(0)",
		"SetNextReconcileAt(databaseNow)",
		"account.Update()",
		"ClearMutationLeaseToken()",
		"ClearMutationLeaseUntil()",
		"SetMutationEpoch(nextEpoch)",
	)
	requireMailboxEntSubmissionRecoveryForbids(t, unknown,
		"ClearDispatchAuthorizedAt()", "ClearDispatchAuthorizationHash()", "SetCommandFence(", "ClearCommandFence(",
		"mailboxEntSubmissionRecoveryRequeue(", "mailboxEntSubmissionRecoveryTerminal(", "draft.Update()", "SetStatus(maildraft.Status(",
		"SetStatus(mailsubmission.Status(\"queued\"))", "SetStatus(mailsubmission.Status(\"failed\"))", "SetStatus(mailsubmission.Status(\"submitting\"))",
		"BeginEngineCommand(", ".PreflightSubmission(", ".Submit(", "MailEngine", "provider.", "nats.", "http.", "net/http",
	)

	requireMailboxEntSubmissionRecoveryOrdered(t, unknownLease,
		"submission.Update()",
		"SetStatus(mailsubmission.Status(\"unknown\"))",
		"ClearCommandLeaseToken()",
		"ClearCommandLeaseExpiresAt()",
		"SetReconcileAttempts(submission.ReconcileAttempts)",
		"MailboxSubmissionErrorCodeReconciliationExhausted",
		"SetNextReconcileAt(databaseNow)",
		"mailboxEntSubmissionRecoveryReleaseExactAccountLease",
	)
	requireMailboxEntSubmissionRecoveryForbids(t, unknownLease,
		"ClearDispatchAuthorizedAt()", "ClearDispatchAuthorizationHash()", "draft.Update()", "SetProviderSubmissionID(", "SetProviderMessageID(",
		"BeginEngineCommand(", ".PreflightSubmission(", ".Submit(", "MailEngine", "provider.", "nats.", "http.", "net/http",
	)

	requireMailboxEntSubmissionRecoveryOrdered(t, accepted,
		"submission.Update()",
		"ClearCommandLeaseToken()",
		"ClearCommandLeaseExpiresAt()",
		"mailboxEntSubmissionRecoveryReleaseExactAccountLease",
	)
	requireMailboxEntSubmissionRecoveryForbids(t, accepted,
		"SetStatus(", "ClearDispatchAuthorizedAt()", "ClearDispatchAuthorizationHash()", "draft.Update()", "SetProviderSubmissionID(", "SetProviderMessageID(",
		"ClearProviderSubmissionID()", "ClearProviderMessageID()", "SetReconcileAttempts(", "SetNextReconcileAt(",
		"BeginEngineCommand(", ".PreflightSubmission(", ".Submit(", "MailEngine", "provider.", "nats.", "http.", "net/http",
	)
	requireMailboxEntSubmissionRecoveryContains(t, exactAccountLease,
		"*account.MutationLeaseToken != *submission.CommandLeaseToken",
		"!account.MutationLeaseUntil.Equal(*submission.CommandLeaseExpiresAt)",
		"account.MutationEpoch != submission.CommandFence",
		"ClearMutationLeaseToken()",
		"ClearMutationLeaseUntil()",
		"SetMutationEpoch(nextEpoch)",
	)
	requireMailboxEntSubmissionRecoveryContains(t, leaseState,
		"(account.MutationLeaseToken == nil) != (account.MutationLeaseUntil == nil)",
		"(submission.CommandLeaseToken == nil) != (submission.CommandLeaseExpiresAt == nil)",
		"case !accountHeld && !submissionHeld:",
		"case accountHeld && submissionHeld:",
		"mailboxEntSubmissionRecoveryLeaseStateReleased",
		"mailboxEntSubmissionRecoveryLeaseStateConflict",
	)
	requireMailboxEntSubmissionRecoveryContains(t, releasedState,
		"case MailboxSubmissionStatusUnknown:",
		"case MailboxSubmissionStatusAccepted:",
		"submission.CommandLeaseToken != nil",
		"submission.ReconcileAttempts == 10 && submission.NextReconcileAt == nil",
	)
	requireMailboxEntSubmissionRecoveryForbids(t, releasedState,
		"draft.Update()", "ClearCommandLeaseToken()", "ClearCommandLeaseExpiresAt()", "SetStatus(",
		"BeginEngineCommand(", ".PreflightSubmission(", ".Submit(", "MailEngine", "provider.", "nats.", "http.", "net/http",
	)
	requireMailboxEntSubmissionRecoveryContains(t, releasedNoop,
		"mailboxEntSubmissionRecoveryReleasedStateValid",
		"return nil",
	)
	requireMailboxEntSubmissionRecoveryForbids(t, releasedNoop,
		"submission.Update()", "account.Update()", "draft.Update()", "ClearCommandLeaseToken()", "ClearMutationLeaseToken()",
		"BeginEngineCommand(", ".PreflightSubmission(", ".Submit(", "MailEngine", "provider.", "nats.", "http.", "net/http",
	)
}
