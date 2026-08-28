package mail_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

var _ mail.MailboxSetupStatusSource = (*mail.MailboxEntRepository)(nil)
var _ func(*mail.MailboxEntRepository, context.Context, uuid.UUID) (mail.MailboxSetupStatus, error) = (*mail.MailboxEntRepository).ResolveMailboxSetupStatus

func mailboxSetupStatusEntSourceSlice(t *testing.T, source, start, end string) string {
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

func requireMailboxSetupStatusEntSourceContains(t *testing.T, source string, fragments ...string) {
	t.Helper()

	for _, fragment := range fragments {
		if !strings.Contains(source, fragment) {
			t.Fatalf("source is missing %q", fragment)
		}
	}
}

func requireMailboxSetupStatusEntSourceOrdered(t *testing.T, source string, fragments ...string) {
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

func TestMailboxSetupStatusEntSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_setup_status_ent.go")
	if err != nil {
		t.Fatalf("read mailbox setup status Ent source: %v", err)
	}
	source := string(sourceBytes)
	resolve := mailboxSetupStatusEntSourceSlice(t, source,
		"func (repository *MailboxEntRepository) ResolveMailboxSetupStatus",
		"func mailboxEntResolveMailboxSetupStatus",
	)
	read := mailboxSetupStatusEntSourceSlice(t, source,
		"func mailboxEntResolveMailboxSetupStatus",
		"func mailboxSetupStatusFromRows",
	)
	classify := mailboxSetupStatusEntSourceSlice(t, source,
		"func mailboxSetupStatusFromRows",
		"func mailboxSetupStatusAccountIntegrityValid",
	)
	public := mailboxSetupStatusEntSourceSlice(t, source,
		"func mailboxSetupStatusWithoutAddress",
		"func mailboxSetupStatusDomainIntegrityValid",
	)

	requireMailboxSetupStatusEntSourceOrdered(t, resolve,
		"repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || actorID == uuid.Nil",
		"repository.withMailboxEntReadSnapshot",
		"mailboxEntResolveMailboxSetupStatus",
		"mailboxSetupStatusCopy(status)",
	)
	requireMailboxSetupStatusEntSourceContains(t, read,
		"transaction.Organization.Query()",
		"organization.HasUsersWith(user.IDEQ(actorID), user.StatusEQ(user.StatusActive))",
		"transaction.User.Query()",
		"user.IDEQ(actorID)",
		"user.HasOrganizationWith(organization.IDEQ(organizationRow.ID))",
		"transaction.MailAccount.Query()",
		"mailaccount.UserIDEQ(actor.ID)",
		"Order(ent.Asc(mailaccount.FieldID))",
		"len(accounts) == 0",
		"len(accounts) != 1",
		"transaction.MailDomain.Query()",
		"maildomain.IDEQ(account.MailDomainID)",
		"maildomain.OrganizationIDEQ(organizationRow.ID)",
		"Order(ent.Asc(maildomain.FieldID))",
		"len(domains) != 1",
		"transaction.MailIdentity.Query()",
		"mailidentity.MailAccountIDEQ(account.ID)",
		"mailidentity.KindEQ(mailidentity.KindPrimary)",
		"Order(ent.Asc(mailidentity.FieldID))",
		"len(identities) != 1",
		"transaction.MailOAuthCredential.Query()",
		"mailoauthcredential.MailAccountIDEQ(account.ID)",
		"Order(ent.Asc(mailoauthcredential.FieldID))",
	)
	requireMailboxSetupStatusEntSourceContains(t, classify,
		"mailboxBootstrapOAuthDomainReady(domain, organizationID)",
		"mailboxPasswordEnrollmentEligible(account)",
		"mailboxBootstrapOAuthAccountReady(account)",
		"mailboxEntOAuthActiveRowValid(credentials[0], account.ID)",
		"MailboxSetupStagePreparing",
		"MailboxSetupStagePasswordRequired",
		"MailboxSetupStageOAuthRequired",
		"MailboxSetupStageSyncing",
		"MailboxSetupStageReady",
		"mailboxSetupStatusActionRequired()",
	)
	requireMailboxSetupStatusEntSourceContains(t, public,
		"MailboxSetupStageActionRequired",
	)
	if strings.Contains(source, "ForUpdate()") {
		t.Fatal("setup-status snapshot must not acquire row locks")
	}
	if strings.Contains(source, "time.Now") || strings.Contains(source, "AccessTokenExpiresAt") {
		t.Fatal("setup-status source must not use a local clock or OAuth token expiry")
	}
	for _, forbidden := range []string{
		"AccessTokenCiphertext",
		"RefreshTokenCiphertext",
		"IssuerOrigin",
		"CredentialGeneration",
		"net/http",
		"http.",
		"stalwart.",
		"nats.",
		"Publish(",
	} {
		if strings.Contains(public, forbidden) {
			t.Fatalf("public setup-status construction exposes forbidden field or behavior %q", forbidden)
		}
	}
}
