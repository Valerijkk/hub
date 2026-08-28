package mail_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

var _ mail.MailboxSubmissionClaimRepository = (*mail.MailboxEntRepository)(nil)
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionClaimRequest) (mail.MailboxSubmissionClaimResult, error) = (*mail.MailboxEntRepository).ClaimMailboxSubmission
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionAuthorizeRequest) (mail.MailboxSubmissionAuthorizeResult, error) = (*mail.MailboxEntRepository).AuthorizeMailboxSubmission
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionPreflightFailureRequest) (mail.MailboxSubmissionAuthorizeResult, error) = (*mail.MailboxEntRepository).ResolveMailboxSubmissionPreflightFailure

func mailboxEntSubmissionClaimSourceSlice(t *testing.T, source, start, end string) string {
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

func requireMailboxEntSubmissionClaimOrdered(t *testing.T, source string, fragments ...string) {
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

func TestMailboxEntSubmissionClaimSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_submission_claim.go")
	if err != nil {
		t.Fatalf("read submission claim source: %v", err)
	}
	source := string(sourceBytes)
	claim := mailboxEntSubmissionClaimSourceSlice(t, source,
		"func (repository *MailboxEntRepository) ClaimMailboxSubmission",
		"func (repository *MailboxEntRepository) AuthorizeMailboxSubmission",
	)

	requireMailboxEntSubmissionClaimOrdered(t, claim,
		"repository.client.MailSubmission.Query()",
		"repository.withMailboxEntTransaction",
		"transaction.MailAccount.Query()",
		"transaction.MailSubmission.Query()",
		"transaction.MailDraft.Query()",
		"transaction.MailboxPolicy.Query()",
		"transaction.MailDomain.Query()",
		"mailboxEntDatabaseWallClock",
		"mailboxEntSubmissionClaimBusyUntil",
		"SetMutationLeaseToken(leaseID)",
		"if err := submission.Update().",
	)
	for _, required := range []string{
		"MailboxEngineCommandLeaseDuration",
		"mailboxEntCheckedTimeAdd",
		"uuid.New()",
		"mailboxEntSubmissionClaimPolicyBinding",
		"mailboxEntSubmissionClaimDraftMatches",
		"SetStatus(mailsubmission.Status(\"authorizing\"))",
		"SetCommandLeaseToken(leaseID)",
		"SetCommandLeaseExpiresAt(leaseUntil)",
		"SetCommandFence(nextEpoch)",
		"SetAttemptCount(submission.AttemptCount + 1)",
		"Order(ent.Asc(maildomain.FieldID))",
	} {
		if !strings.Contains(claim, required) {
			t.Fatalf("claim source is missing %q", required)
		}
	}
	if strings.Contains(claim, "mailboxEntSubmissionClaimRepositoryError(") {
		t.Fatal("claim path uses deprecated submission repository mapper")
	}
	if got := strings.Count(claim, "mailboxEntSubmissionRepositoryError("); got != 2 {
		t.Fatalf("claim canonical repository mapper calls = %d, want 2", got)
	}
	for _, forbidden := range []string{"BeginEngineCommand(", "MailEngine", ".PreflightSubmission(", ".Submit(", "http.", "net/http", "nats."} {
		if strings.Contains(claim, forbidden) {
			t.Fatalf("claim path contains forbidden call %q", forbidden)
		}
	}
}

func TestMailboxEntSubmissionAuthorizationSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_submission_claim.go")
	if err != nil {
		t.Fatalf("read submission claim source: %v", err)
	}
	source := string(sourceBytes)
	authorize := mailboxEntSubmissionClaimSourceSlice(t, source,
		"func (repository *MailboxEntRepository) AuthorizeMailboxSubmission",
		"func (repository *MailboxEntRepository) ResolveMailboxSubmissionPreflightFailure",
	)

	requireMailboxEntSubmissionClaimOrdered(t, authorize,
		"transaction.MailAccount.Query()",
		"transaction.MailSubmission.Query()",
		"transaction.MailDraft.Query()",
		"transaction.User.Query()",
		"transaction.MailIdentity.Query()",
		"mailboxEntSubmissionLockCleanAttachments",
		"transaction.MailboxPolicy.Query()",
		"transaction.MailDomain.Query()",
		"mailboxEntSubmissionLockRateBucket",
		"mailboxEntDatabaseWallClock",
		"mailboxEntSubmissionAuthorizingFenceMatches",
		"submissionPreflightBindingHash",
		"submissionPreflightPolicyBindingHash",
		"MailboxSubmissionCleanAttachmentProofHash",
		"EvaluateMailboxSubmissionRateLimit",
		"MailboxSubmissionDispatchAuthorizationHash",
		"SetStatus(mailsubmission.Status(\"submitting\"))",
	)
	for _, required := range []string{
		"subtle.ConstantTimeCompare",
		"claim.policy.Hash()",
		"freshPolicy.Hash()",
		"mailboxEntSubmissionTerminal",
		"mailboxEntSubmissionRequeue",
		"MailboxSubmissionErrorCodePolicy",
		"MailboxSubmissionErrorCodeIdentityUnavailable",
		"MailboxSubmissionErrorCodeAttachmentNotClean",
		"MailboxSubmissionErrorCodeRejected",
	} {
		if !strings.Contains(authorize, required) {
			t.Fatalf("authorization source is missing %q", required)
		}
	}
	rateBucketPersistence := mailboxEntSubmissionClaimSourceSlice(t, source,
		"func mailboxEntSubmissionPersistRateBucket",
		"func mailboxEntSubmissionDraftSubmittingMatches",
	)
	if !strings.Contains(rateBucketPersistence, "mailsendratebucket.VersionEQ(row.Version)") {
		t.Fatal("rate bucket persistence is missing version compare-and-swap")
	}
	terminal := mailboxEntSubmissionClaimSourceSlice(t, source,
		"func mailboxEntSubmissionTerminal",
		"func mailboxEntSubmissionFailedResult",
	)
	for _, required := range []string{
		"SetMutationEpoch(nextEpoch)",
		"ClearCommandLeaseToken()",
		"ClearMutationLeaseToken()",
	} {
		if !strings.Contains(terminal, required) {
			t.Fatalf("terminal transition is missing %q", required)
		}
	}
	if got := strings.Count(authorize, "EvaluateMailboxSubmissionRateLimit("); got != 1 {
		t.Fatalf("rate evaluator calls = %d, want exactly 1", got)
	}
	if strings.Contains(authorize, "mailboxEntSubmissionClaimRepositoryError(") {
		t.Fatal("authorization path uses deprecated submission repository mapper")
	}
	if got := strings.Count(authorize, "mailboxEntSubmissionRepositoryError("); got != 2 {
		t.Fatalf("authorization canonical repository mapper calls = %d, want 2", got)
	}
	for _, forbidden := range []string{"BeginEngineCommand(", "MailEngine", ".PreflightSubmission(", ".Submit(", "http.", "net/http", "nats."} {
		if strings.Contains(authorize, forbidden) {
			t.Fatalf("authorization path contains forbidden call %q", forbidden)
		}
	}
}

func TestMailboxEntSubmissionPreflightFailureResolutionSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_submission_claim.go")
	if err != nil {
		t.Fatalf("read submission claim source: %v", err)
	}
	resolve := mailboxEntSubmissionClaimSourceSlice(t, string(sourceBytes),
		"func (repository *MailboxEntRepository) ResolveMailboxSubmissionPreflightFailure",
		"func mailboxEntSubmissionClaimLookupError",
	)

	requireMailboxEntSubmissionClaimOrdered(t, resolve,
		"transaction.MailAccount.Query()",
		"transaction.MailSubmission.Query()",
		"transaction.MailDraft.Query()",
		"mailboxEntDatabaseWallClock",
		"mailboxEntSubmissionAuthorizingFenceMatches",
		"MailboxSubmissionPreflightFailureDispositionFailed",
		"MailboxSubmissionPreflightFailureDispositionRetry",
		"MailboxSubmissionPreflightFailureDispositionRateLimited",
	)
	for _, required := range []string{
		"MailboxSubmissionErrorCodeRetryExhausted",
		"submission.AttemptCount-1",
		"mailboxEntCheckedTimeAdd",
		"mailboxEntSubmissionRateLimitedResult",
		"mailboxEntSubmissionRequeuedResult",
	} {
		if !strings.Contains(resolve, required) {
			t.Fatalf("preflight failure path is missing %q", required)
		}
	}
	if strings.Contains(resolve, "mailboxEntSubmissionClaimRepositoryError(") {
		t.Fatal("preflight failure path uses deprecated submission repository mapper")
	}
	if got := strings.Count(resolve, "mailboxEntSubmissionRepositoryError("); got != 2 {
		t.Fatalf("preflight failure canonical repository mapper calls = %d, want 2", got)
	}
	for _, forbidden := range []string{"BeginEngineCommand(", "MailEngine", ".PreflightSubmission(", ".Submit(", "http.", "net/http", "nats."} {
		if strings.Contains(resolve, forbidden) {
			t.Fatalf("preflight failure path contains forbidden call %q", forbidden)
		}
	}
}

func TestMailboxEntSubmissionTerminalDraftLifecycleSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_submission_claim.go")
	if err != nil {
		t.Fatalf("read submission claim source: %v", err)
	}
	source := string(sourceBytes)
	terminal := mailboxEntSubmissionClaimSourceSlice(t, source,
		"func mailboxEntSubmissionTerminal",
		"func mailboxEntSubmissionRequeue",
	)

	requireMailboxEntSubmissionClaimOrdered(t, terminal,
		"mailboxEntSubmissionDraftSubmittingMatches",
		"submission.Update()",
		"draft.Update()",
		"SetStatus(maildraft.Status(\"active\"))",
		"account.Update()",
	)
	draftUpdate := mailboxEntSubmissionClaimSourceSlice(t, terminal,
		"if err := draft.Update().",
		"if err := account.Update().",
	)
	if !strings.Contains(draftUpdate, "SetUpdatedAt(databaseNow)") {
		t.Fatal("terminal draft update is missing SetUpdatedAt(databaseNow)")
	}
	draftValidator := mailboxEntSubmissionClaimSourceSlice(t, source,
		"func mailboxEntSubmissionDraftSubmittingMatches",
		"func mailboxEntSubmissionTerminal",
	)
	for _, required := range []string{
		"submission.MailDraftID",
		"MailboxDraftStatus(draft.Status) == MailboxDraftStatusSubmitting",
	} {
		if !strings.Contains(draftValidator, required) {
			t.Fatalf("terminal draft validation is missing %q", required)
		}
	}

	requeue := mailboxEntSubmissionClaimSourceSlice(t, source,
		"func mailboxEntSubmissionRequeue",
		"func mailboxEntSubmissionFailedResult",
	)
	if strings.Contains(requeue, "draft.Update()") || strings.Contains(requeue, "maildraft.Status(\"active\")") {
		t.Fatal("requeue must preserve the draft submitting state")
	}
}
