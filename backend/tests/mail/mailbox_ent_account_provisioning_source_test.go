package mail_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

var _ mail.MailboxAccountProvisioningRepository = (*mail.MailboxEntRepository)(nil)
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error) = (*mail.MailboxEntRepository).Scan
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxAccountProvisioningClaimRequest) (mail.MailboxAccountProvisioningClaimResult, error) = (*mail.MailboxEntRepository).ClaimMailboxAccountProvisioning
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxAccountProvisioningFinalizeRequest) (mail.MailboxAccountProvisioningFinalizeResult, error) = (*mail.MailboxEntRepository).Finalize

func mailboxEntAccountProvisioningSourceSlice(t *testing.T, source, start, end string) string {
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

func requireMailboxEntAccountProvisioningSourceContains(t *testing.T, source string, fragments ...string) {
	t.Helper()

	for _, fragment := range fragments {
		if !strings.Contains(source, fragment) {
			t.Fatalf("source is missing %q", fragment)
		}
	}
}

func requireMailboxEntAccountProvisioningSourceOrdered(t *testing.T, source string, fragments ...string) {
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

func TestMailboxEntAccountProvisioningSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_account_provisioning.go")
	if err != nil {
		t.Fatalf("read mailbox account provisioning Ent source: %v", err)
	}
	source := string(sourceBytes)
	requireMailboxEntAccountProvisioningSourceContains(t, source,
		"len(*account.AttemptedControlHash) != sha256.Size",
		"subtle.ConstantTimeCompare(*account.AttemptedControlHash, account.ControlHash) != 1",
		"mailaccount.AttemptedControlHashEQ(*account.AttemptedControlHash)",
		"subtle.ConstantTimeCompare(*account.AttemptedControlHash, controlHash) != 1",
	)
	scan := mailboxEntAccountProvisioningSourceSlice(t, source,
		"func (repository *MailboxEntRepository) Scan",
		"func (repository *MailboxEntRepository) ClaimMailboxAccountProvisioning",
	)
	claim := mailboxEntAccountProvisioningSourceSlice(t, source,
		"func (repository *MailboxEntRepository) ClaimMailboxAccountProvisioning",
		"func (repository *MailboxEntRepository) Finalize",
	)
	finalize := mailboxEntAccountProvisioningSourceSlice(t, source,
		"func (repository *MailboxEntRepository) Finalize",
		"func mailboxEntAccountProvisioningScanDue",
	)
	due := mailboxEntAccountProvisioningSourceSlice(t, source,
		"func mailboxEntAccountProvisioningScanDue",
		"func mailboxEntAccountProvisioningClaimLocks",
	)
	domainRequired := mailboxEntAccountProvisioningSourceSlice(t, source,
		"func mailboxEntAccountProvisioningSetDomainRequired",
		"func mailboxEntAccountProvisioningCASPredicates",
	)

	requireMailboxEntAccountProvisioningSourceOrdered(t, scan,
		"NormalizeMailboxAccountProvisioningScanRequest(request)",
		"repository.withMailboxEntTransaction",
		"mailboxEntDatabaseWallClock(ctx, transaction)",
		"mailboxEntAccountProvisioningScanDue",
		"mailboxEntAccountProvisioningMaterializeIntents",
	)
	requireMailboxEntAccountProvisioningSourceContains(t, scan,
		"if _, duplicate := seen[accountID]; duplicate",
		"user.StatusEQ(user.StatusActive)",
		"user.Not(user.HasMailAccount())",
		"maildomain.IsDefaultEQ(true)",
		"maildomain.DkimStatusEQ(maildomain.DkimStatus(\"valid\"))",
		"mailboxEntAccountProvisioningDomainLive",
		"mailboxEntAccountProvisioningIntentID",
		"SetStatus(mailaccount.Status(\"provisioning\"))",
		"SetDesiredEnabled(true)",
		"SetProjectionState(mailaccount.ProjectionState(\"bootstrapping\"))",
		"SetControlGeneration(1)",
		"SetControlHashVersion(1)",
		"SetProvisionAttempts(0)",
		"SetNextProvisionAt(databaseNow)",
		"SetEventCutoffAt(databaseNow)",
		"SetSyncScheduleUpdatedAt(databaseNow)",
		"SetCreatedAt(databaseNow)",
		"SetUpdatedAt(databaseNow)",
		"SetKind(mailidentity.KindPrimary)",
		"SetStatus(mailidentity.Status(\"pending\"))",
		"SetSendAllowed(false)",
		"mailboxEntAccountProvisioningLocalPart",
		"validCanonicalRegistryLocalPart",
		"hex.EncodeToString(userID[:])",
		"u-"+" + suffix[:12]",
		"u-"+" + suffix",
	)
	requireMailboxEntAccountProvisioningSourceContains(t, due,
		"mailaccount.StatusEQ(mailaccount.Status(\"provisioning\"))",
		"mailaccount.NextProvisionAtLTE(databaseNow)",
		"mailaccount.ProvisionLeaseUntilLTE(databaseNow)",
		"Order(ent.Asc(mailaccount.FieldNextProvisionAt), ent.Asc(mailaccount.FieldID))",
		"Order(ent.Asc(mailaccount.FieldProvisionLeaseUntil), ent.Asc(mailaccount.FieldID))",
		"sort.SliceStable",
	)
	if got := strings.Count(scan, "repository.withMailboxEntTransaction"); got != 1 {
		t.Fatalf("scan transactions = %d, want exactly one", got)
	}
	if got := strings.Count(scan, "mailboxEntDatabaseWallClock(ctx, transaction)"); got != 1 {
		t.Fatalf("scan database clocks = %d, want exactly one", got)
	}

	requireMailboxEntAccountProvisioningSourceOrdered(t, claim,
		"repository.withMailboxEntTransaction",
		"mailboxEntDatabaseWallClock(ctx, transaction)",
		"mailboxEntAccountProvisioningClaimLocks",
		"mailboxEntAccountProvisioningDomainGate",
		"NewMailboxAccountProvisioningClaim",
	)
	requireMailboxEntAccountProvisioningSourceContains(t, claim,
		"ForUpdate()",
		"mailboxEntAccountProvisioningConverged",
		"mailboxEntAccountProvisioningDue",
		"mailboxEntAccountProvisioningExactPendingPrimary",
		"mailboxEntAccountProvisioningSetDomainRequired(ctx, transaction, account, dueAt, databaseNow)",
		"SetAttemptedControlGeneration(account.ControlGeneration)",
		"SetAttemptedControlHashVersion(account.ControlHashVersion)",
		"SetProvisionLeaseToken(leaseToken)",
		"SetProvisionLeaseUntil(leaseDeadline)",
		"SetProvisionAttempts(account.ProvisionAttempts + 1)",
		"ClearNextProvisionAt()",
		"ClearProvisionErrorCode()",
		"hex.EncodeToString(account.ControlHash)",
		"MailboxEngineCommandLeaseDuration",
		"math.MaxInt32",
		"if updated != 1",
	)
	requireMailboxEntAccountProvisioningSourceContains(t, domainRequired,
		"mailboxEntAccountProvisioningCASPredicates(account)",
		"ClearAttemptedControlGeneration()",
		"ClearAttemptedControlHashVersion()",
		"ClearAttemptedControlHash()",
		"ClearProvisionLeaseToken()",
		"ClearProvisionLeaseUntil()",
		"SetProvisionErrorCode(mailaccount.ProvisionErrorCode(MailboxAccountProvisioningCodeDomainRequired))",
		"if err != nil || updated != 1",
	)
	if strings.Contains(claim, "MailboxAccountProvisioningCodeEngineDisabled") {
		t.Fatal("claim must not add an engine-disabled local gate")
	}
	if got := strings.Count(claim, "repository.withMailboxEntTransaction"); got != 1 {
		t.Fatalf("claim transactions = %d, want exactly one", got)
	}
	if got := strings.Count(claim, "mailboxEntDatabaseWallClock(ctx, transaction)"); got != 1 {
		t.Fatalf("claim database clocks = %d, want exactly one", got)
	}

	requireMailboxEntAccountProvisioningSourceOrdered(t, finalize,
		"request.Valid()",
		"repository.withMailboxEntTransaction",
		"mailboxEntDatabaseWallClock(ctx, transaction)",
		"mailboxEntAccountProvisioningFinalizeLocks",
		"mailboxEntAccountProvisioningFinalizeFenceMatches",
	)
	requireMailboxEntAccountProvisioningSourceContains(t, finalize,
		"MailboxAccountProvisioningRetryDelay(int(account.ProvisionAttempts))",
		"SetEngineAccountID(request.EngineAccountID)",
		"SetDeployedGeneration(fence.ControlGeneration)",
		"SetDeployedHashVersion(fence.ControlHashVersion)",
		"ClearAttemptedControlGeneration()",
		"ClearAttemptedControlHashVersion()",
		"ClearAttemptedControlHash()",
		"ClearProvisionLeaseToken()",
		"ClearProvisionLeaseUntil()",
		"ClearNextProvisionAt()",
		"ClearProvisionErrorCode()",
		"SetNextProvisionAt(retryAt)",
		"SetProvisionErrorCode(mailaccount.ProvisionErrorCode(request.Code))",
		"if updated != 1",
	)
	for _, forbidden := range []string{
		"SetStatus(mailidentity.Status(\"active\"))",
		"SetSendAllowed(true)",
		"MailboxAccountProvisioningCodeEngineDisabled",
		"time.Now",
		"net/http",
		"http.",
		"stalwart.",
		"nats.",
		"Publish(",
		"Commit(",
		"Rollback(",
	} {
		if strings.Contains(finalize, forbidden) {
			t.Fatalf("finalize contains forbidden transition or external I/O %q", forbidden)
		}
	}
	if got := strings.Count(finalize, "repository.withMailboxEntTransaction"); got != 1 {
		t.Fatalf("finalize transactions = %d, want exactly one", got)
	}
	if got := strings.Count(finalize, "mailboxEntDatabaseWallClock(ctx, transaction)"); got != 1 {
		t.Fatalf("finalize database clocks = %d, want exactly one", got)
	}
}
