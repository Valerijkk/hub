package mail_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

var _ mail.MailboxPasswordEnrollmentRepository = (*mail.MailboxEntRepository)(nil)
var _ func(*mail.MailboxEntRepository, context.Context, uuid.UUID, uuid.UUID) (mail.MailboxPasswordEnrollmentClaimResult, error) = (*mail.MailboxEntRepository).Claim
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxPasswordEnrollmentClaim, time.Time) error = (*mail.MailboxEntRepository).Complete
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxPasswordEnrollmentClaim, mail.MailboxPasswordEnrollmentReleaseCode) error = (*mail.MailboxEntRepository).Release

func mailboxPasswordEnrollmentEntSourceSlice(t *testing.T, source, start, end string) string {
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

func requireMailboxPasswordEnrollmentEntOrdered(t *testing.T, source string, fragments ...string) {
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

func TestMailboxPasswordEnrollmentEntSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_password_enrollment_ent.go")
	if err != nil {
		t.Fatalf("read password enrollment Ent source: %v", err)
	}
	source := string(sourceBytes)
	capabilityBytes, err := os.ReadFile("../../internal/mail/mailbox_password_enrollment.go")
	if err != nil {
		t.Fatalf("read password enrollment capability source: %v", err)
	}
	capability := string(capabilityBytes)
	claim := mailboxPasswordEnrollmentEntSourceSlice(t, source,
		"func (repository *MailboxEntRepository) Claim",
		"func (repository *MailboxEntRepository) Complete",
	)
	markerNoop := mailboxPasswordEnrollmentEntSourceSlice(t, claim,
		"if account.PasswordEnrolledEngineAccountID != nil",
		"if account.PasswordEnrollmentLeaseUntil != nil",
	)
	complete := mailboxPasswordEnrollmentEntSourceSlice(t, source,
		"func (repository *MailboxEntRepository) Complete",
		"func (repository *MailboxEntRepository) Release",
	)
	release := mailboxPasswordEnrollmentEntSourceSlice(t, source,
		"func (repository *MailboxEntRepository) Release",
		"func mailboxPasswordEnrollmentLockOwner",
	)
	lockOwner := mailboxPasswordEnrollmentEntSourceSlice(t, source,
		"func mailboxPasswordEnrollmentLockOwner",
		"func mailboxPasswordEnrollmentAccountIntegrityValid",
	)
	eligibility := mailboxPasswordEnrollmentEntSourceSlice(t, source,
		"func mailboxPasswordEnrollmentEligible",
		"func mailboxPasswordEnrollmentCurrentTupleMatches",
	)
	currentTuple := mailboxPasswordEnrollmentEntSourceSlice(t, source,
		"func mailboxPasswordEnrollmentCurrentTupleMatches",
		"func mailboxPasswordEnrollmentEntUnavailable",
	)
	claimedInput := mailboxPasswordEnrollmentEntSourceSlice(t, capability,
		"type MailboxPasswordEnrollmentClaimedInput struct",
		"type MailboxPasswordEnrollmentClaim struct",
	)
	sealedClaim := mailboxPasswordEnrollmentEntSourceSlice(t, capability,
		"type MailboxPasswordEnrollmentClaim struct",
		"type MailboxPasswordEnrollmentClaimResult struct",
	)
	claimFromInput := mailboxPasswordEnrollmentEntSourceSlice(t, capability,
		"func mailboxPasswordEnrollmentClaimFromInput",
		"func mailboxPasswordEnrollmentClaimValid",
	)
	claimValid := mailboxPasswordEnrollmentEntSourceSlice(t, capability,
		"func mailboxPasswordEnrollmentClaimValid",
		"func mailboxPasswordEnrollmentClaimCopy",
	)

	requireMailboxPasswordEnrollmentEntOrdered(t, claim,
		"repository.withMailboxEntTransaction",
		"mailboxPasswordEnrollmentLockOwner",
		"mailboxEntDatabaseWallClock",
		"mailboxPasswordEnrollmentAccountIntegrityValid",
		"mailboxPasswordEnrollmentEligible",
		"math.MaxInt64",
		"mailboxEntCheckedTimeAdd",
		"SetPasswordEnrollmentGeneration(nextGeneration)",
		"SetPasswordEnrollmentLeaseToken(requestedLease)",
		"SetPasswordEnrollmentLeaseUntil(leaseUntil)",
		"SetUpdatedAt(databaseWallClock)",
	)
	requireMailboxPasswordEnrollmentEntOrdered(t, claim,
		"if account.PasswordEnrolledEngineAccountID != nil",
		"if account.PasswordEnrollmentLeaseToken != nil",
		"NewMailboxPasswordEnrollmentNoopResult()",
		"if account.PasswordEnrollmentLeaseUntil != nil && account.PasswordEnrollmentLeaseUntil.After(databaseWallClock)",
		"NewMailboxPasswordEnrollmentBusyResult()",
		"if account.PasswordEnrollmentGeneration == math.MaxInt64",
		"nextGeneration := account.PasswordEnrollmentGeneration + 1",
		"SetPasswordEnrollmentGeneration(nextGeneration)",
	)
	if strings.Contains(markerNoop, "PasswordEnrollmentLeaseUntil") {
		t.Fatal("an enrolled account without a lease must remain a noop")
	}
	requireMailboxPasswordEnrollmentEntOrdered(t, claim,
		"mailboxPasswordEnrollmentClaimDeployedTuple(account)",
		"NewMailboxPasswordEnrollmentClaimedResult(MailboxPasswordEnrollmentClaimedInput{",
		"DeployedGeneration:  deployedGeneration",
		"DeployedHashVersion: deployedHashVersion",
		"DeployedHash:        deployedHash",
	)
	requireMailboxPasswordEnrollmentEntOrdered(t, lockOwner,
		"transaction.Organization.Query()",
		"user.StatusEQ(user.StatusActive)",
		"ForUpdate()",
		"transaction.User.Query()",
		"transaction.MailAccount.Query()",
		"mailaccount.OrganizationIDEQ(organizationRow.ID)",
		"Order(ent.Asc(mailaccount.FieldID))",
		"len(accounts) != 1",
	)
	for _, required := range []string{
		"30 * time.Second",
		"mailaccount.Status(\"provisioning\")",
		"mailaccount.Status(\"deleting\")",
		"account.DesiredEnabled",
		"account.DeployedGeneration != nil",
		"account.DeployedHashVersion != nil",
		"account.DeployedHash != nil",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("source is missing %q", required)
		}
	}
	requireMailboxPasswordEnrollmentEntOrdered(t, complete,
		"transaction.MailAccount.Query()",
		"mailaccount.IDEQ(claim.accountID)",
		"mailaccount.OrganizationIDEQ(claim.organizationID)",
		"ForUpdate()",
		"mailboxEntDatabaseWallClock",
		"mailboxPasswordEnrollmentCurrentTupleMatches(account, claim, databaseWallClock, true)",
		"SetPasswordEnrolledEngineAccountID(claim.engineAccountID)",
		"SetPasswordEnrolledAt(observedAt)",
		"ClearPasswordEnrollmentLeaseToken()",
		"ClearPasswordEnrollmentLeaseUntil()",
	)
	if !strings.Contains(complete, "return mailboxPasswordEnrollmentEntRetryable()") {
		t.Fatal("Complete must reject a deployed-tuple mismatch before enrollment mutation")
	}
	requireMailboxPasswordEnrollmentEntOrdered(t, release,
		"transaction.MailAccount.Query()",
		"ForUpdate()",
		"mailboxEntDatabaseWallClock",
		"mailboxPasswordEnrollmentCurrentTupleMatches(account, claim, databaseWallClock, false)",
		"ClearPasswordEnrollmentLeaseToken()",
		"ClearPasswordEnrollmentLeaseUntil()",
	)
	requireMailboxPasswordEnrollmentEntOrdered(t, release,
		"if !mailboxPasswordEnrollmentCurrentTupleMatches(account, claim, databaseWallClock, false)",
		"return nil",
		"account.Update()",
		"ClearPasswordEnrollmentLeaseToken()",
	)
	for _, required := range []string{
		"DeployedGeneration  int64",
		"DeployedHashVersion int64",
		"DeployedHash        [sha256.Size]byte",
		"deployedGeneration  int64",
		"deployedHashVersion int64",
		"deployedHash        [sha256.Size]byte",
		"deployedGeneration:  input.DeployedGeneration",
		"deployedHashVersion: input.DeployedHashVersion",
		"deployedHash:        input.DeployedHash",
		"claim.deployedGeneration >= 1",
		"claim.deployedHashVersion >= 1",
		"mailboxPasswordEnrollmentClaimDeployedTupleMatches(account, claim)",
		"return *account.DeployedGeneration == claim.deployedGeneration",
		"*account.DeployedHashVersion == claim.deployedHashVersion",
		"deployedHash == claim.deployedHash",
	} {
		if !strings.Contains(claimedInput, required) && !strings.Contains(sealedClaim, required) && !strings.Contains(claimFromInput, required) && !strings.Contains(claimValid, required) && !strings.Contains(currentTuple, required) {
			t.Fatalf("source is missing deployed tuple capability contract %q", required)
		}
	}
	if strings.Contains(eligibility, "time.Now") || strings.Contains(source, "time.Now") {
		t.Fatal("durable enrollment state must use the database wall clock")
	}
	for _, forbidden := range []string{
		"SetUserPassword(",
		"RegistryPasswordPort",
		"net/http",
		"http.",
		"nats.",
		"Publish(",
		"SetPasswordHash(",
		"SetSecret(",
		"SetControlHash(",
		"SetDeployedHash(",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("Ent enrollment repository contains forbidden external or secret behavior %q", forbidden)
		}
	}
}
