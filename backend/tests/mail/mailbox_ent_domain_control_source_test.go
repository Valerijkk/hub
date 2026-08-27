package mail_test

import (
	"os"
	"strings"
	"testing"
)

func mailboxEntDomainControlSourceSlice(t *testing.T, source, start, end string) string {
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

func mailboxEntDomainControlRequireOrdered(t *testing.T, source string, fragments ...string) {
	t.Helper()
	previous := -1
	for _, fragment := range fragments {
		position := strings.Index(source, fragment)
		if position < 0 {
			t.Fatalf("source is missing %q", fragment)
		}
		if position <= previous {
			t.Fatalf("source puts %q before its predecessor", fragment)
		}
		previous = position
	}
}

func TestMailboxEntDomainControlSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_domain_control.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	create := mailboxEntDomainControlSourceSlice(t, source, "func (repository *MailboxEntRepository) CreateMailDomain", "// ListMailDomains")
	list := mailboxEntDomainControlSourceSlice(t, source, "func (repository *MailboxEntRepository) ListMailDomains", "// SetDefaultMailDomain")
	setDefault := mailboxEntDomainControlSourceSlice(t, source, "func (repository *MailboxEntRepository) SetDefaultMailDomain", "func mailDomainControlResolveActorOrganizationForUpdate")
	fresh := mailboxEntDomainControlSourceSlice(t, source, "func mailDomainControlCreateFresh", "func mailDomainControlReceiptReplayKind")
	replay := mailboxEntDomainControlSourceSlice(t, source, "func mailDomainControlReplayCreate", "func mailDomainControlCreateFresh")
	replayTarget := mailboxEntDomainControlSourceSlice(t, source, "func mailDomainControlReplayTargetValid", "func mailDomainControlVerifiedEntRowValid")
	verified := mailboxEntDomainControlSourceSlice(t, source, "func mailDomainControlVerifiedEntRowValid", "func mailDomainControlEntRowValid")

	for _, part := range []string{create, list, setDefault} {
		for _, required := range []string{"repository.withMailboxEntTransaction", "mailboxEntDatabaseWallClock", "mailDomainControlResolveActorOrganizationForUpdate", "ForUpdate()"} {
			if !strings.Contains(part, required) {
				t.Fatalf("transactional method is missing %q", required)
			}
		}
	}
	for _, required := range []string{
		"mailmutationreceipt.Operation(mailDomainControlOperation)",
		"mailmutationreceipt.KeySource(" + "\x22client\x22" + ")",
		"mailmutationreceipt.State(" + "\x22completed\x22" + ")",
		"mailmutationreceipt.TargetKind(" + "\x22domain\x22" + ")",
		"SetOriginalHTTPStatus(202)",
		"mailDomainControlReplayLifetime",
		"mailDomainControlReceiptReplayKind",
		"maildomain.DomainEQ(request.CanonicalDomain)",
		"SetState(maildomain.State(" + "\x22pending\x22" + "))",
		"SetVerificationStatus(maildomain.VerificationStatus(" + "\x22pending\x22" + "))",
		"SetDKIMStatus(maildomain.DKIMStatus(" + "\x22unknown\x22" + "))",
		"SetRequiredDNSRecords(RequiredDNSRecords{})",
		"SetIsDefault(false)",
		"SetDesiredGeneration(1)",
		"SetDesiredHashVersion(mailDomainControlFingerprintVersion)",
	} {
		if !strings.Contains(fresh, required) {
			t.Fatalf("fresh create is missing %q", required)
		}
	}
	for _, required := range []string{
		"!bytes.Equal(receipt.RequestHash, request.RequestHash[:])",
		"MailDomainControlErrorConflict",
		"mailDomainControlReplayTargetValid",
		"maildomain.IDEQ(*receipt.TargetID)",
		"maildomain.OrganizationIDEQ(organizationID)",
	} {
		if !strings.Contains(replay, required) {
			t.Fatalf("receipt replay is missing %q", required)
		}
	}
	for _, required := range []string{
		"mailDomainControlEntRowValid(domain, organizationID)",
		"domain.ID == request.DomainID",
		"domain.Domain == request.CanonicalDomain",
		"domain.EngineProvisioningKey == request.ProvisioningKey",
		"domain.DeletedAt == nil",
	} {
		if !strings.Contains(replayTarget, required) {
			t.Fatalf("advanced replay target validation is missing %q", required)
		}
	}
	for _, forbidden := range []string{"mailDomainControlPendingReplayRowValid", "domain.CreatedAt.Equal(domain.UpdatedAt)", "domain.State !="} {
		if strings.Contains(replayTarget, forbidden) {
			t.Fatalf("advanced replay target validation must not require %q", forbidden)
		}
	}
	for _, required := range []string{
		"maildomain.StateEQ(maildomain.State(" + "\x22verified\x22" + "))",
		"maildomain.VerificationStatusEQ(maildomain.VerificationStatus(" + "\x22verified\x22" + "))",
		"maildomain.DKIMStatusEQ(maildomain.DKIMStatus(" + "\x22valid\x22" + "))",
		"maildomain.DesiredPresentEQ(true)",
		"maildomain.EnabledEQ(true)",
		"maildomain.DeletedAtIsNil()",
		"maildomain.IsDefaultEQ(true)",
		"maildomain.IsDefaultEQ(false)",
		"SetIsDefault(false)",
		"SetIsDefault(true)",
	} {
		if !strings.Contains(setDefault, required) {
			t.Fatalf("set-default is missing %q", required)
		}
	}
	for _, required := range []string{
		"domain.DKIMStatus != maildomain.DKIMStatus(" + "\x22valid\x22" + ")",
		"*domain.DeployedGeneration != domain.DesiredGeneration",
		"*domain.DeployedHashVersion != domain.DesiredHashVersion",
		"subtle.ConstantTimeCompare(domain.DeployedHash, domain.DesiredHash) != 1",
	} {
		if !strings.Contains(verified, required) {
			t.Fatalf("locked default validator is missing %q", required)
		}
	}
	mailboxEntDomainControlRequireOrdered(t, setDefault, "SetIsDefault(false)", "SetIsDefault(true)")
	for _, forbidden := range []string{"time.Now", "net/http", "http.", "stalwart.", "EnsureDomain(", "ReadLocalDomainDNSZoneFile", "Publish("} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("repository must not contain %q", forbidden)
		}
	}
}
