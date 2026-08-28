package mail_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

var (
	_ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSyncClaimRequest) (mail.MailboxSyncClaimResult, error) = (*mail.MailboxEntRepository).ClaimMailboxSyncStep
	_ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSyncResolveRequest) (mail.MailboxSyncResolveResult, error) = (*mail.MailboxEntRepository).ResolveMailboxSyncStep
	_ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) = (*mail.MailboxEntRepository).ScanDueMailboxSyncAccounts
	_ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) = (*mail.MailboxEntRepository).RescheduleMailboxSyncWorkerFault
)

func mailboxEntCoordinationSourceSlice(t *testing.T, source, start, end string) string {
	t.Helper()

	startIndex := strings.Index(source, start)
	if startIndex == -1 {
		t.Fatalf("source is missing start marker %q", start)
	}
	endOffset := strings.Index(source[startIndex+len(start):], end)
	if endOffset == -1 {
		t.Fatalf("source is missing end marker %q after %q", end, start)
	}
	return source[startIndex : startIndex+len(start)+endOffset]
}

func requireMailboxEntOrderedFragments(t *testing.T, source string, fragments ...string) {
	t.Helper()

	previous := -1
	for _, fragment := range fragments {
		position := strings.Index(source, fragment)
		if position == -1 {
			t.Fatalf("source is missing %q", fragment)
		}
		if position <= previous {
			t.Fatalf("source fragment %q is out of order", fragment)
		}
		previous = position
	}
}

func requireMailboxEntNoCoordinationIO(t *testing.T, source string) {
	t.Helper()

	for _, forbidden := range []string{"http.", "net/http", "nats.", "NATS", "MailEngine", "stalwart", "os.", "log.", "callback"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("coordination path contains forbidden I/O vocabulary %q", forbidden)
		}
	}
}

func TestMailboxEntSyncCoordinationClaimAndResolveSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_sync_claim.go")
	if err != nil {
		t.Fatalf("read claim source: %v", err)
	}
	source := string(sourceBytes)

	claim := mailboxEntCoordinationSourceSlice(t, source,
		"func (repository *MailboxEntRepository) ClaimMailboxSyncStep",
		"func (repository *MailboxEntRepository) ResolveMailboxSyncStep",
	)
	requireMailboxEntOrderedFragments(t, claim,
		"request.AccountID == uuid.Nil",
		"request.LeaseID == uuid.Nil",
		"!mailboxEntValidTime(request.Now)",
		"!mailboxEntValidTime(request.LeaseUntil)",
		"!request.LeaseUntil.After(request.Now)",
		"repository.withMailboxEntTransaction",
		"ForUpdate()",
		"databaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)",
		"!request.LeaseUntil.After(databaseWallClock)",
		"SetSyncScheduleUpdatedAt(databaseWallClock)",
		"mailboxEntSyncClaimSnapshot",
	)
	if got := strings.Count(claim, "mailboxEntDatabaseWallClock(ctx, transaction)"); got != 1 {
		t.Fatalf("claim transaction DB clock reads = %d, want 1", got)
	}
	if !strings.Contains(source, "account.SyncLeaseUntil != nil && !mailboxEntValidTime(*account.SyncLeaseUntil)") {
		t.Fatal("persisted sync lease expiry must be exact-second UTC before any claim or resolve comparison")
	}
	for _, required := range []string{
		"ClearMutationLeaseToken()", "ClearMutationLeaseUntil()", "SetMutationEpoch(nextMutationEpoch)",
		"ClearSyncLeaseResultHashVersion()", "ClearSyncLeaseResultHash()", "ClearSyncLeaseAppliedPhase()", "ClearSyncLeaseAppliedNextAt()", "ClearSyncLeaseAppliedAt()",
		"SetSyncLeaseToken(request.LeaseID)", "SetSyncLeaseUntil(request.LeaseUntil)", "SetSyncLeaseEpoch(leaseEpoch)",
	} {
		if !strings.Contains(claim, required) {
			t.Fatalf("claim is missing durable lease transition %q", required)
		}
	}
	reclaim := mailboxEntCoordinationSourceSlice(t, claim,
		"nextMutationEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)",
		"persistedFence, hasLease, leaseValid := mailboxEntSyncLeaseFence(account)",
	)
	requireMailboxEntOrderedFragments(t, reclaim,
		"ClearMutationLeaseToken()",
		"ClearMutationLeaseUntil()",
		"SetMutationEpoch(nextMutationEpoch)",
		"SetSyncScheduleUpdatedAt(databaseWallClock)",
		"account.MutationLeaseToken = nil",
		"account.MutationLeaseUntil = nil",
		"account.MutationEpoch = nextMutationEpoch",
	)
	liveRenewal := mailboxEntCoordinationSourceSlice(t, claim,
		"if hasLease && account.SyncLeaseUntil.After(databaseWallClock)",
		"if account.NextSyncAt.After(databaseWallClock)",
	)
	requireMailboxEntOrderedFragments(t, liveRenewal,
		"persistedFence.LeaseID != request.LeaseID",
		"mailboxEntSyncNonClaim(*account.SyncLeaseUntil)",
		"persistedFence.SyncEpoch != account.SyncEpoch || persistedFence.MutationEpoch != account.MutationEpoch",
		"leaseUntil := request.LeaseUntil",
		"account.SyncLeaseUntil.After(leaseUntil)",
		"SetSyncLeaseUntil(leaseUntil)",
		"SetSyncScheduleUpdatedAt(databaseWallClock)",
	)
	if !strings.Contains(claim, "result = mailboxEntSyncNonClaim(*account.MutationLeaseUntil)") {
		t.Fatal("live mutation lease must return its exact persisted expiry")
	}
	requireMailboxEntNoCoordinationIO(t, claim)

	resolve := mailboxEntCoordinationSourceSlice(t, source,
		"func (repository *MailboxEntRepository) ResolveMailboxSyncStep",
		"func mailboxEntSyncClaimSnapshot",
	)
	requireMailboxEntOrderedFragments(t, resolve,
		"validMailboxSyncFence(request.Fence)",
		"!mailboxEntValidTime(request.Now)",
		"!mailboxEntValidTime(request.NextAttemptAt)",
		"request.NextAttemptAt.Before(request.Now)",
		"repository.withMailboxEntTransaction",
		"ForUpdate()",
		"databaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)",
		"persistedFence != request.Fence",
		"account.SyncLeaseUntil.After(databaseWallClock)",
		"SetNextSyncAt(request.NextAttemptAt)",
		"SetSyncScheduleUpdatedAt(databaseWallClock)",
		"ClearSyncLeaseToken()",
		"ClearSyncLeaseUntil()",
	)
	if got := strings.Count(resolve, "mailboxEntDatabaseWallClock(ctx, transaction)"); got != 1 {
		t.Fatalf("resolve transaction DB clock reads = %d, want 1", got)
	}
	for _, required := range []string{
		"MAIL_SYNC_CURSOR_INVALID", "MAIL_SYNC_RESTART_REQUIRED", "MAIL_SYNC_ENGINE_UNAVAILABLE", "MAIL_SYNC_PAGE_REJECTED",
		"ClearSyncLeaseSyncEpoch()", "ClearSyncLeaseMutationEpoch()", "ClearSyncLeasePhase()", "ClearSyncLeaseCursor()", "ClearSyncLeaseBaseCursor()", "ClearSyncLeaseCheckpoint()",
	} {
		if !strings.Contains(resolve, required) {
			t.Fatalf("resolve is missing failure/fence transition %q", required)
		}
	}
	requireMailboxEntNoCoordinationIO(t, resolve)
}

func TestMailboxEntSyncWorkerCoordinationSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_worker.go")
	if err != nil {
		t.Fatalf("read worker source: %v", err)
	}
	source := string(sourceBytes)

	scan := mailboxEntCoordinationSourceSlice(t, source,
		"func (repository *MailboxEntRepository) ScanDueMailboxSyncAccounts",
		"func (repository *MailboxEntRepository) RescheduleMailboxSyncWorkerFault",
	)
	requireMailboxEntOrderedFragments(t, scan,
		"request.Limit < 1 || request.Limit > 50",
		"repository.withMailboxEntTransaction",
		"now, err := repository.mailboxEntNow()",
		"mailboxEntMailboxSyncSkipLocked()",
		"mailaccount.NextSyncAtLTE(now)",
		"ent.Asc(mailaccount.FieldNextSyncAt)",
		"ent.Asc(mailaccount.FieldID)",
		"Limit(request.Limit)",
		"make([]MailboxSyncWorkerDueAccount, 0, len(rows))",
	)
	if got := strings.Count(scan, "repository.mailboxEntNow()"); got != 1 {
		t.Fatalf("scan durable clock reads = %d, want 1", got)
	}
	if !strings.Contains(source, "sql.WithLockAction(sql.SkipLocked)") {
		t.Fatal("due scan must use Ent selector SKIP LOCKED predicate")
	}
	rowMapping := mailboxEntCoordinationSourceSlice(t, source,
		"func mailboxEntWorkerDueAccount",
		"func mailboxEntValidWorkerFaultRequest",
	)
	for _, required := range []string{
		"(account.SyncLeaseToken == nil) != (account.SyncLeaseUntil == nil)",
		"(account.MutationLeaseToken == nil) != (account.MutationLeaseUntil == nil)",
		"!mailboxEntValidTime(*account.SyncLeaseUntil)",
		"account.SyncLeaseUntil.After(selectedAt)",
		"!mailboxEntValidTime(*account.MutationLeaseUntil)",
		"account.MutationLeaseUntil.After(selectedAt)",
		"UpdatedAt:     account.SyncScheduleUpdatedAt",
	} {
		if !strings.Contains(rowMapping, required) {
			t.Fatalf("due-row mapping is missing fail-closed lease validation %q", required)
		}
	}
	requireMailboxEntNoCoordinationIO(t, scan)

	fault := mailboxEntCoordinationSourceSlice(t, source,
		"func (repository *MailboxEntRepository) RescheduleMailboxSyncWorkerFault",
		"func mailboxEntMailboxSyncSkipLocked",
	)
	requireMailboxEntOrderedFragments(t, fault,
		"!mailboxEntValidWorkerFaultRequest(request)",
		"repository.withMailboxEntTransaction",
		"now, err := repository.mailboxEntNow()",
		"now.Before(request.SelectedAt)",
		"mailboxEntCheckedTimeAdd(now, request.RetryDelay)",
		"mailaccount.IDEQ(request.Account.AccountID)",
		"mailaccount.NextSyncAtEQ(request.Account.DueAt)",
		"mailaccount.SyncScheduleUpdatedAtEQ(request.Account.UpdatedAt)",
		"mailaccount.SyncEpochEQ(request.Account.SyncEpoch)",
		"mailaccount.MutationEpochEQ(request.Account.MutationEpoch)",
		"mailaccount.SyncLeaseEpochEQ(request.Account.LeaseEpoch)",
		"mailaccount.SyncAttemptsEQ(request.Account.SyncAttempts)",
		"SetSyncAttempts(request.Account.SyncAttempts + 1)",
		"SetNextSyncAt(nextAttemptAt)",
		"SetSyncScheduleUpdatedAt(now)",
	)
	if got := strings.Count(fault, "repository.mailboxEntNow()"); got != 1 {
		t.Fatalf("fault durable clock reads = %d, want 1", got)
	}
	for _, required := range []string{
		"(account.SyncLeaseToken == nil) != (account.SyncLeaseUntil == nil)",
		"(account.MutationLeaseToken == nil) != (account.MutationLeaseUntil == nil)",
		"!mailboxEntValidTime(*account.SyncLeaseUntil)",
		"account.SyncLeaseUntil.After(now)",
		"!mailboxEntValidTime(*account.MutationLeaseUntil)",
		"account.MutationLeaseUntil.After(now)",
	} {
		if !strings.Contains(fault, required) {
			t.Fatalf("fault is missing fail-closed lease guard %q", required)
		}
	}
	for _, forbidden := range []string{"ClearSyncLease", "SetProjectionState", "SetSyncCursor", "SetSyncEpoch", "SetMutationEpoch"} {
		if strings.Contains(fault, forbidden) {
			t.Fatalf("fault must not alter lease/projection fence state: found %q", forbidden)
		}
	}
	requireMailboxEntNoCoordinationIO(t, fault)
}
