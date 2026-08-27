package mail

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

func TestMailboxRegistryProvisioningEngineEnsureDomainMapsDNSRequired(t *testing.T) {
	domains := &recordingRegistryDomainController{id: "domain-1"}
	accounts := &recordingRegistryAccountClient{}
	engine, err := newMailboxRegistryProvisioningEngine(domains, accounts)
	if err != nil {
		t.Fatalf("newMailboxRegistryProvisioningEngine() error = %v", err)
	}

	state, err := engine.EnsureDomain(context.Background(), DomainSpec{
		ProvisioningKey:   "domain-key",
		Domain:            "example.test",
		DesiredGeneration: 3,
		DesiredHash:       "domain-hash-3",
		Present:           true,
		Enabled:           true,
	})
	if err != nil {
		t.Fatalf("EnsureDomain() error = %v", err)
	}
	if domains.calls != 1 || domains.domain != "example.test" {
		t.Fatalf("domain calls = %d for %q, want one call for example.test", domains.calls, domains.domain)
	}
	if accounts.calls != 0 {
		t.Fatalf("account calls = %d, want 0", accounts.calls)
	}
	if state.DomainID != "domain-1" || state.State != EngineDomainDNSRequired || !state.Applied || state.Verification != EngineVerificationPending || state.DKIM != EngineVerificationPending {
		t.Fatalf("EnsureDomain() state = %#v, want confirmed DNS-required pending state", state)
	}
	if state.DNSRecords == nil || len(state.DNSRecords) != 0 {
		t.Fatalf("EnsureDomain() DNSRecords = %#v, want non-nil empty records", state.DNSRecords)
	}
}

func TestMailboxRegistryProvisioningEngineEnsureDomainRejectsInvalidBeforeIO(t *testing.T) {
	for _, spec := range []DomainSpec{
		{ProvisioningKey: "domain-key", Domain: "example.test", DesiredHash: "domain-hash", Present: true},
		{ProvisioningKey: "domain-key", Domain: "example.test", DesiredHash: "domain-hash", Enabled: true},
		{ProvisioningKey: "domain-key", Domain: "Example.test", DesiredHash: "domain-hash", Present: true, Enabled: true},
		{ProvisioningKey: "domain-key", Domain: "example.test", Present: true, Enabled: true},
	} {
		domains := &recordingRegistryDomainController{id: "domain-1"}
		accounts := &recordingRegistryAccountClient{}
		engine, err := newMailboxRegistryProvisioningEngine(domains, accounts)
		if err != nil {
			t.Fatalf("newMailboxRegistryProvisioningEngine() error = %v", err)
		}

		_, err = engine.EnsureDomain(context.Background(), spec)
		if !errors.Is(err, ErrEngineRejectedPolicy) {
			t.Fatalf("EnsureDomain(%#v) error = %v, want rejected policy", spec, err)
		}
		if domains.calls != 0 || accounts.calls != 0 {
			t.Fatalf("invalid EnsureDomain() calls = domain %d account %d, want no I/O", domains.calls, accounts.calls)
		}
	}
}

func TestMailboxRegistryProvisioningEngineEnsureAccountUsesDomainIDBeforeAccount(t *testing.T) {
	calls := make([]string, 0, 2)
	domains := &recordingRegistryDomainController{id: "domain-1", callLog: &calls}
	accounts := &recordingRegistryAccountClient{
		account: stalwart.LocalUserAccount{AccountID: "account-1", EmailAddress: "ada@example.test"},
		callLog:  &calls,
	}
	engine, err := newMailboxRegistryProvisioningEngine(domains, accounts)
	if err != nil {
		t.Fatalf("newMailboxRegistryProvisioningEngine() error = %v", err)
	}

	account, err := engine.EnsureAccount(context.Background(), AccountSpec{
		ProvisioningKey:   "account-key",
		LocalPart:         "ada",
		Domain:            "example.test",
		DesiredGeneration: 9,
		DesiredHash:       "account-hash-9",
		Enabled:           true,
		Primary:           AccountIdentitySpec{LocalPart: "ada", Domain: "example.test", Enabled: true},
		Aliases:           []AccountIdentitySpec{},
	})
	if err != nil {
		t.Fatalf("EnsureAccount() error = %v", err)
	}
	if len(calls) != 2 || calls[0] != "domain" || calls[1] != "account" {
		t.Fatalf("call order = %#v, want domain then account", calls)
	}
	wantRequest := stalwart.LocalUserAccountSpec{EmailLocalPart: "ada", DomainID: "domain-1", DomainName: "example.test"}
	if accounts.request != wantRequest {
		t.Fatalf("EnsureLocalUserAccount() request = %#v, want %#v", accounts.request, wantRequest)
	}
	if account.AccountID != "account-1" || account.State != EngineAccountProvisioning || account.AppliedGeneration != 9 || account.AppliedHash != "account-hash-9" || account.PrimaryIdentityID != "" {
		t.Fatalf("EnsureAccount() = %#v, want credential-free provisioning account", account)
	}
	if account.AliasIdentityIDs == nil || len(account.AliasIdentityIDs) != 0 {
		t.Fatalf("EnsureAccount() aliases = %#v, want non-nil empty identity IDs", account.AliasIdentityIDs)
	}
}

func TestMailboxRegistryProvisioningEngineEnsureAccountRejectsInvalidBeforeIO(t *testing.T) {
	valid := AccountSpec{
		ProvisioningKey: "account-key",
		LocalPart:       "ada",
		Domain:          "example.test",
		DesiredHash:     "account-hash",
		Enabled:         true,
		Primary:         AccountIdentitySpec{LocalPart: "ada", Domain: "example.test", Enabled: true},
	}
	for _, spec := range []AccountSpec{
		func() AccountSpec { value := valid; value.Enabled = false; return value }(),
		func() AccountSpec { value := valid; value.Primary.Enabled = false; return value }(),
		func() AccountSpec { value := valid; value.Primary.LocalPart = "other"; return value }(),
		func() AccountSpec { value := valid; value.Aliases = []AccountIdentitySpec{{LocalPart: "alias", Domain: "example.test", Enabled: true}}; return value }(),
		func() AccountSpec { value := valid; value.LocalPart = "Ada"; value.Primary.LocalPart = "Ada"; return value }(),
	} {
		domains := &recordingRegistryDomainController{id: "domain-1"}
		accounts := &recordingRegistryAccountClient{}
		engine, err := newMailboxRegistryProvisioningEngine(domains, accounts)
		if err != nil {
			t.Fatalf("newMailboxRegistryProvisioningEngine() error = %v", err)
		}

		_, err = engine.EnsureAccount(context.Background(), spec)
		if !errors.Is(err, ErrEngineRejectedPolicy) {
			t.Fatalf("EnsureAccount(%#v) error = %v, want rejected policy", spec, err)
		}
		if domains.calls != 0 || accounts.calls != 0 {
			t.Fatalf("invalid EnsureAccount() calls = domain %d account %d, want no I/O", domains.calls, accounts.calls)
		}
	}
}

func TestMailboxRegistryProvisioningEngineFailsClosedAndRedactsProviderDetails(t *testing.T) {
	t.Run("provider error", func(t *testing.T) {
		domains := &recordingRegistryDomainController{id: "domain-1"}
		accounts := &recordingRegistryAccountClient{err: fmt.Errorf("provider-secret: %w", stalwart.ErrUnauthorized)}
		engine, err := newMailboxRegistryProvisioningEngine(domains, accounts)
		if err != nil {
			t.Fatalf("newMailboxRegistryProvisioningEngine() error = %v", err)
		}

		_, err = engine.EnsureAccount(context.Background(), validRegistryProvisioningAccountInput())
		if !errors.Is(err, ErrEngineUnauthorizedAccount) || strings.Contains(err.Error(), "provider-secret") {
			t.Fatalf("EnsureAccount() error = %q, want redacted unauthorized account error", err)
		}
	})

	t.Run("server email mismatch", func(t *testing.T) {
		domains := &recordingRegistryDomainController{id: "domain-1"}
		accounts := &recordingRegistryAccountClient{account: stalwart.LocalUserAccount{AccountID: "account-1", EmailAddress: "wrong@example.test"}}
		engine, err := newMailboxRegistryProvisioningEngine(domains, accounts)
		if err != nil {
			t.Fatalf("newMailboxRegistryProvisioningEngine() error = %v", err)
		}

		_, err = engine.EnsureAccount(context.Background(), validRegistryProvisioningAccountInput())
		if !errors.Is(err, ErrEngineIndeterminate) || strings.Contains(err.Error(), "wrong@example.test") {
			t.Fatalf("EnsureAccount() error = %q, want redacted indeterminate mismatch", err)
		}
	})
}

func validRegistryProvisioningAccountInput() AccountSpec {
	return AccountSpec{
		ProvisioningKey:   "account-key",
		LocalPart:         "ada",
		Domain:            "example.test",
		DesiredGeneration: 9,
		DesiredHash:       "account-hash-9",
		Enabled:           true,
		Primary:           AccountIdentitySpec{LocalPart: "ada", Domain: "example.test", Enabled: true},
		Aliases:           []AccountIdentitySpec{},
	}
}

type recordingRegistryDomainController struct {
	calls   int
	domain  string
	id      string
	err     error
	callLog *[]string
}

func (controller *recordingRegistryDomainController) EnsureLocalDomain(_ context.Context, domain string) (string, error) {
	controller.calls++
	controller.domain = domain
	if controller.callLog != nil {
		*controller.callLog = append(*controller.callLog, "domain")
	}
	return controller.id, controller.err
}

type recordingRegistryAccountClient struct {
	calls   int
	request stalwart.LocalUserAccountSpec
	account stalwart.LocalUserAccount
	err     error
	callLog *[]string
}

func (client *recordingRegistryAccountClient) EnsureLocalUserAccount(_ context.Context, request stalwart.LocalUserAccountSpec) (stalwart.LocalUserAccount, error) {
	client.calls++
	client.request = request
	if client.callLog != nil {
		*client.callLog = append(*client.callLog, "account")
	}
	return client.account, client.err
}
