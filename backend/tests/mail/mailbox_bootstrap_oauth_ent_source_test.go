package mail_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

var _ mail.MailboxBootstrapOAuthRepository = (*mail.MailboxEntRepository)(nil)
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxOAuthScope) (mail.MailboxBootstrapOAuthAccount, error) = (*mail.MailboxEntRepository).ResolveMailboxBootstrapOAuthAccount
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxBootstrapOAuthFinalizeRequest) error = (*mail.MailboxEntRepository).FinalizeMailboxBootstrapOAuth

func mailboxBootstrapOAuthEntSourceSlice(t *testing.T, source, start, end string) string {
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

func requireMailboxBootstrapOAuthEntOrdered(t *testing.T, source string, fragments ...string) {
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

func requireMailboxBootstrapOAuthEntContains(t *testing.T, source string, fragments ...string) {
	t.Helper()

	for _, fragment := range fragments {
		if !strings.Contains(source, fragment) {
			t.Fatalf("source is missing %q", fragment)
		}
	}
}

func TestMailboxBootstrapOAuthEntSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_bootstrap_oauth_ent.go")
	if err != nil {
		t.Fatalf("read bootstrap OAuth Ent source: %v", err)
	}
	source := string(sourceBytes)
	resolve := mailboxBootstrapOAuthEntSourceSlice(t, source,
		"func (repository *MailboxEntRepository) ResolveMailboxBootstrapOAuthAccount",
		"func (repository *MailboxEntRepository) FinalizeMailboxBootstrapOAuth",
	)
	finalize := mailboxBootstrapOAuthEntSourceSlice(t, source,
		"func (repository *MailboxEntRepository) FinalizeMailboxBootstrapOAuth",
		"func mailboxBootstrapOAuthLockSnapshot",
	)
	lockSnapshot := mailboxBootstrapOAuthEntSourceSlice(t, source,
		"func mailboxBootstrapOAuthLockSnapshot",
		"func mailboxBootstrapOAuthAccountSnapshot",
	)
	accountReady := mailboxBootstrapOAuthEntSourceSlice(t, source,
		"func mailboxBootstrapOAuthAccountReady",
		"func mailboxBootstrapOAuthDomainReady",
	)
	domainReady := mailboxBootstrapOAuthEntSourceSlice(t, source,
		"func mailboxBootstrapOAuthDomainReady",
		"func mailboxBootstrapOAuthAccountCopy",
	)

	requireMailboxBootstrapOAuthEntOrdered(t, resolve,
		"repository.withMailboxEntTransaction",
		"mailboxBootstrapOAuthLockSnapshot",
		"mailboxBootstrapOAuthAccountCopy(snapshot)",
	)
	requireMailboxBootstrapOAuthEntContains(t, lockSnapshot,
		"transaction.Organization.Query()",
		"organization.IDEQ(scope.OrganizationID)",
		"organization.HasUsersWith(",
		"user.IDEQ(scope.ActorID)",
		"user.StatusEQ(user.StatusActive)",
		"transaction.User.Query()",
		"user.HasOrganizationWith(organization.IDEQ(organizationRow.ID))",
		"transaction.MailAccount.Query()",
		"mailaccount.IDEQ(scope.AccountID)",
		"mailaccount.OrganizationIDEQ(organizationRow.ID)",
		"mailaccount.UserIDEQ(actor.ID)",
		"mailaccount.StatusEQ(mailaccount.Status(\"provisioning\"))",
		"mailaccount.DesiredEnabledEQ(true)",
		"mailaccount.EngineAccountIDNotNil()",
		"Order(ent.Asc(mailaccount.FieldID))",
		"len(accounts) != 1",
		"transaction.MailDomain.Query()",
		"maildomain.IDEQ(account.MailDomainID)",
		"maildomain.OrganizationIDEQ(organizationRow.ID)",
		"maildomain.DesiredPresentEQ(true)",
		"maildomain.EnabledEQ(true)",
		"maildomain.StateEQ(maildomain.State(\"verified\"))",
		"maildomain.VerificationStatusEQ(maildomain.VerificationStatus(\"verified\"))",
		"maildomain.DeletedAtIsNil()",
		"maildomain.EngineDomainIDNotNil()",
		"transaction.MailIdentity.Query()",
		"mailidentity.MailAccountIDEQ(account.ID)",
		"mailidentity.KindEQ(mailidentity.KindPrimary)",
		"mailidentity.StatusEQ(mailidentity.Status(\"pending\"))",
		"mailidentity.SendAllowedEQ(false)",
		"mailidentity.EngineIdentityIDIsNil()",
		"mailidentity.NormalizedAddressEQ(expectedAddress)",
		"len(identities) != 1",
		"transaction.MailOAuthCredential.Query()",
		"mailoauthcredential.MailAccountIDEQ(account.ID)",
		"len(credentials) != 0",
	)
	if got := strings.Count(lockSnapshot, "ForUpdate()"); got != 6 {
		t.Fatalf("FOR UPDATE lock count = %d, want 6 for organization, actor, account, domain, identity, and credential fence", got)
	}

	requireMailboxBootstrapOAuthEntContains(t, accountReady,
		"account.ControlGeneration < 1",
		"account.ControlHashVersion < 1",
		"len(account.ControlHash) != sha256.Size",
		"account.DeployedGeneration == nil",
		"*account.DeployedGeneration != account.ControlGeneration",
		"*account.DeployedHashVersion != account.ControlHashVersion",
		"subtle.ConstantTimeCompare(*account.DeployedHash, account.ControlHash) != 1",
		"account.AttemptedControlGeneration != nil",
		"account.AttemptedControlHashVersion != nil",
		"account.AttemptedControlHash != nil",
		"account.ProvisionLeaseToken != nil",
		"account.ProvisionLeaseUntil != nil",
		"account.NextProvisionAt != nil",
		"account.ProvisionErrorCode != nil",
		"account.PasswordEnrollmentGeneration < 1",
		"account.PasswordEnrolledEngineAccountID == nil",
		"*account.PasswordEnrolledEngineAccountID != *account.EngineAccountID",
		"account.PasswordEnrolledAt == nil",
		"account.PasswordEnrollmentLeaseToken != nil",
		"account.PasswordEnrollmentLeaseUntil != nil",
	)
	requireMailboxBootstrapOAuthEntContains(t, domainReady,
		"domain.DesiredGeneration >= 1",
		"domain.DesiredHashVersion >= 1",
		"len(domain.DesiredHash) == sha256.Size",
		"domain.DeployedGeneration != nil",
		"domain.DeployedHashVersion != nil",
		"domain.DeployedHash != nil",
		"*domain.DeployedGeneration == domain.DesiredGeneration",
		"*domain.DeployedHashVersion == domain.DesiredHashVersion",
		"len(*domain.DeployedHash) == sha256.Size",
		"subtle.ConstantTimeCompare(*domain.DeployedHash, domain.DesiredHash) == 1",
	)

	requireMailboxBootstrapOAuthEntOrdered(t, finalize,
		"validMailboxBootstrapOAuthFinalizeRequest(request)",
		"copyMailboxOAuthBytes(request.AccessTokenCiphertext)",
		"copyMailboxOAuthBytes(request.RefreshTokenCiphertext)",
		"repository.withMailboxEntTransaction",
		"mailboxBootstrapOAuthLockSnapshot",
		"snapshot != request.Account",
		"account.DeployedHash == nil",
		"transaction.MailOAuthCredential.Create()",
		"transaction.MailIdentity.Update()",
		"transaction.MailAccount.Update()",
	)
	requireMailboxBootstrapOAuthEntContains(t, finalize,
		"defer clear(request.AccessTokenCiphertext)",
		"defer clear(request.RefreshTokenCiphertext)",
		"SetProvider(mailoauthcredential.Provider(mailboxOAuthProvider))",
		"SetIssuerOrigin(request.IssuerOrigin)",
		"SetStatus(mailoauthcredential.Status(\"active\"))",
		"SetCredentialGeneration(1)",
		"SetEncryptionFormatVersion(mailboxOAuthEncryptionFormat)",
		"SetAccessTokenCiphertext(request.AccessTokenCiphertext)",
		"SetRefreshTokenCiphertext(request.RefreshTokenCiphertext)",
		"SetTokenType(mailoauthcredential.TokenType(\"bearer\"))",
		"SetGrantedScope(mailboxOAuthCanonicalScope)",
		"SetAccessTokenExpiresAt(request.AccessTokenExpiresAt)",
		"SetRefreshTokenExpiresAt(request.RefreshTokenExpiresAt)",
		"SetConnectedAt(request.ObservedAt)",
		"SetCreatedAt(request.ObservedAt)",
		"SetUpdatedAt(request.ObservedAt)",
		"SetVersion(1)",
		"mailidentity.StatusEQ(mailidentity.Status(\"pending\"))",
		"mailidentity.SendAllowedEQ(false)",
		"mailidentity.EngineIdentityIDIsNil()",
		"mailidentity.NormalizedAddressEQ(snapshot.PrimaryEmail)",
		"SetStatus(mailidentity.Status(\"active\"))",
		"SetSendAllowed(true)",
		"SetEngineIdentityID(request.EngineIdentityID)",
		"mailaccount.PasswordEnrollmentGenerationEQ(snapshot.PasswordEnrollmentGeneration)",
		"mailaccount.PasswordEnrolledEngineAccountIDEQ(snapshot.EngineAccountID)",
		"mailaccount.PasswordEnrolledAtEQ(*account.PasswordEnrolledAt)",
		"mailaccount.PasswordEnrollmentLeaseTokenIsNil()",
		"mailaccount.PasswordEnrollmentLeaseUntilIsNil()",
		"mailaccount.DeployedHashEQ(*account.DeployedHash)",
		"mailaccount.AttemptedControlGenerationIsNil()",
		"mailaccount.AttemptedControlHashVersionIsNil()",
		"mailaccount.AttemptedControlHashIsNil()",
		"mailaccount.ProvisionLeaseTokenIsNil()",
		"mailaccount.ProvisionLeaseUntilIsNil()",
		"mailaccount.NextProvisionAtIsNil()",
		"mailaccount.ProvisionErrorCodeIsNil()",
		"SetProjectionState(mailaccount.ProjectionState(\"bootstrapping\"))",
		"ClearSyncCursor()",
		"ClearRebuildCheckpoint()",
		"ClearRebuildBaseCursor()",
		"ClearSyncErrorCode()",
		"ClearLastSyncedAt()",
		"ClearProvisionErrorCode()",
		"ClearNextProvisionAt()",
		"SetNextSyncAt(request.ObservedAt)",
		"SetSyncScheduleUpdatedAt(request.ObservedAt)",
	)
	if got := strings.Count(finalize, "if updated != 1"); got != 2 {
		t.Fatalf("identity/account compare-and-swap count checks = %d, want 2", got)
	}
	if got := strings.Count(finalize, "repository.withMailboxEntTransaction"); got != 1 {
		t.Fatalf("finalizer transactions = %d, want exactly one atomic transaction", got)
	}
	for _, forbidden := range []string{
		"net/http",
		"http.",
		"stalwart.",
		"nats.",
		"Publish(",
		"ExchangeAuthorizationCode(",
		"Commit(",
		"Rollback(",
	} {
		if strings.Contains(finalize, forbidden) {
			t.Fatalf("bootstrap OAuth finalizer contains forbidden external or transaction-boundary call %q", forbidden)
		}
	}
}
