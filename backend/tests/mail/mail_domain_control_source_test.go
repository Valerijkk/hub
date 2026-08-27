package mail_test

import (
	"os"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

var _ mail.MailDomainControlRepository = (*mail.MailboxEntRepository)(nil)

func mailDomainControlSourceSlice(t *testing.T, source, start, end string) string {
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

func TestMailDomainControlCoreSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mail_domain_control.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	createRequest := mailDomainControlSourceSlice(t, source, "type MailDomainControlCreateRequest struct", "// MailDomainControlListRequest")
	view := mailDomainControlSourceSlice(t, source, "type MailDomainControlView struct", "// MailDomainControlCreateResult")
	create := mailDomainControlSourceSlice(t, source, "func (control *MailDomainControl) CreateMailDomain", "// ListMailDomains")

	for _, required := range []string{
		"type MailDomainControlErrorKind string",
		"MailDomainControlErrorInvalid",
		"MailDomainControlErrorConflict",
		"MailDomainControlErrorUnavailable",
		"type MailDomainControlRepository interface",
		"CreateMailDomain(context.Context, mailDomainControlCreateRepositoryRequest)",
		"ListMailDomains(context.Context, MailDomainControlListRequest)",
		"SetDefaultMailDomain(context.Context, MailDomainControlSetDefaultRequest)",
		"func canonicalMailDomainControlDomain",
		"strings.ToLower(strings.TrimSpace(value))",
		"func mailDomainControlIdempotencyKeyHash",
		"hub.mail.domain.idempotency.v1",
		"func mailDomainControlRequestHash",
		"hub.mail.domain.create.request.v1",
		"func mailDomainControlDesiredHash",
		"hub.mail.domain.desired.v1",
		"uuid.NewSHA1(uuid.NameSpaceURL",
		"hub.mail.domain.id.v1",
		"hub-mail-domain-v1:",
		"value.Version() == 4 || value.Version() == 7",
		"mailDomainControlFingerprintVersion int32 = 1",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("source is missing %q", required)
		}
	}
	for _, forbidden := range []string{"OrganizationID", "time.Now", "net/http", "http.", "stalwart.", "EnsureDomain(", "ReadLocalDomainDNSZoneFile"} {
		if strings.Contains(createRequest, forbidden) || strings.Contains(create, forbidden) {
			t.Fatalf("create boundary must not contain %q", forbidden)
		}
	}
	for _, forbidden := range []string{"EngineDomainID", "EngineProvisioningKey", "Lease", "LastError", "Provider"} {
		if strings.Contains(view, forbidden) {
			t.Fatalf("safe view must not expose %q", forbidden)
		}
	}
	for _, required := range []string{"ID", "Domain", "State", "VerificationStatus", "DKIMStatus", "RequiredDNSRecords", "IsDefault"} {
		if !strings.Contains(view, required) {
			t.Fatalf("safe view is missing %q", required)
		}
	}
}
