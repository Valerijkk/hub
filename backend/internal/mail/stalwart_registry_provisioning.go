package mail

import (
	"context"
	"strings"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

// MailboxRegistryProvisioningEngine adapts the Registry-only domain and local
// account operations to the provider-neutral provisioning surface. It retains
// no state and never receives mailbox authorization or credentials.
type MailboxRegistryProvisioningEngine struct {
	domains  registryProvisioningDomainController
	accounts registryProvisioningAccountClient
}

type registryProvisioningDomainController interface {
	EnsureLocalDomain(context.Context, string) (string, error)
}

type registryProvisioningAccountClient interface {
	EnsureLocalUserAccount(context.Context, stalwart.LocalUserAccountSpec) (stalwart.LocalUserAccount, error)
}

// NewMailboxRegistryProvisioningEngine binds the accepted Registry codecs
// without performing I/O or retaining credential material.
func NewMailboxRegistryProvisioningEngine(domains *stalwart.RegistryDomainController, accounts *stalwart.RegistryAccountClient) (*MailboxRegistryProvisioningEngine, error) {
	if domains == nil || accounts == nil {
		return nil, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	return newMailboxRegistryProvisioningEngine(domains, accounts)
}

func newMailboxRegistryProvisioningEngine(domains registryProvisioningDomainController, accounts registryProvisioningAccountClient) (*MailboxRegistryProvisioningEngine, error) {
	if domains == nil || accounts == nil {
		return nil, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	return &MailboxRegistryProvisioningEngine{domains: domains, accounts: accounts}, nil
}

// EnsureDomain accepts only the enabled, present, canonical local desired
// state. Stalwart does not prove public DNS or DKIM verification, so a
// Registry-confirmed domain is always reported as pending DNS-required state.
func (engine *MailboxRegistryProvisioningEngine) EnsureDomain(ctx context.Context, spec DomainSpec) (EngineDomainState, error) {
	if engine == nil || engine.domains == nil || engine.accounts == nil || ctx == nil || !validRegistryProvisioningDomainSpec(spec) {
		return EngineDomainState{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	if err := ctx.Err(); err != nil {
		return EngineDomainState{}, mapStalwartEngineError(err, true)
	}

	domainID, err := engine.domains.EnsureLocalDomain(ctx, spec.Domain)
	if err != nil {
		return EngineDomainState{}, mapStalwartEngineError(err, true)
	}
	if !validOpaqueID(domainID) {
		return EngineDomainState{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	return EngineDomainState{
		DomainID:     domainID,
		State:        EngineDomainDNSRequired,
		Applied:      true,
		DNSRecords:   []EngineDNSRecord{},
		Verification: EngineVerificationPending,
		DKIM:         EngineVerificationPending,
	}, nil
}

// EnsureAccount admits only one enabled canonical primary identity with no
// aliases. It revalidates the exact domain immediately before the
// credential-free Registry account ensure.
func (engine *MailboxRegistryProvisioningEngine) EnsureAccount(ctx context.Context, spec AccountSpec) (EngineAccount, error) {
	if engine == nil || engine.domains == nil || engine.accounts == nil || ctx == nil || !validRegistryProvisioningAccountSpec(spec) {
		return EngineAccount{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	if err := ctx.Err(); err != nil {
		return EngineAccount{}, mapStalwartEngineError(err, true)
	}

	domainID, err := engine.domains.EnsureLocalDomain(ctx, spec.Domain)
	if err != nil {
		return EngineAccount{}, mapStalwartEngineError(err, true)
	}
	if !validOpaqueID(domainID) {
		return EngineAccount{}, stalwartEngineError(EngineErrorIndeterminate)
	}

	confirmed, err := engine.accounts.EnsureLocalUserAccount(ctx, stalwart.LocalUserAccountSpec{
		EmailLocalPart: spec.LocalPart,
		DomainID:       domainID,
		DomainName:     spec.Domain,
	})
	if err != nil {
		return EngineAccount{}, mapStalwartEngineError(err, true)
	}
	if !validAccountID(confirmed.AccountID) || confirmed.EmailAddress != spec.LocalPart+"@"+spec.Domain {
		return EngineAccount{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	return EngineAccount{
		AccountID:         confirmed.AccountID,
		State:             EngineAccountProvisioning,
		AppliedGeneration: spec.DesiredGeneration,
		AppliedHash:       spec.DesiredHash,
		AliasIdentityIDs:  []string{},
	}, nil
}

func validRegistryProvisioningDomainSpec(spec DomainSpec) bool {
	return spec.Present && spec.Enabled &&
		validSafeText(spec.ProvisioningKey, MaxEngineOpaqueIDBytes, true) &&
		validCanonicalRegistryDomain(spec.Domain) &&
		validSafeText(spec.DesiredHash, MaxEngineOpaqueIDBytes, true)
}

func validRegistryProvisioningAccountSpec(spec AccountSpec) bool {
	return spec.Enabled && len(spec.Aliases) == 0 &&
		validSafeText(spec.ProvisioningKey, MaxEngineOpaqueIDBytes, true) &&
		validCanonicalRegistryLocalPart(spec.LocalPart) &&
		validCanonicalRegistryDomain(spec.Domain) &&
		validSafeText(spec.DesiredHash, MaxEngineOpaqueIDBytes, true) &&
		spec.Primary.Enabled &&
		spec.Primary.LocalPart == spec.LocalPart &&
		spec.Primary.Domain == spec.Domain
}

func validCanonicalRegistryLocalPart(value string) bool {
	if len(value) == 0 || len(value) > 64 {
		return false
	}
	for index := range value {
		character := value[index]
		if !((character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '.' || character == '-' || character == '_') {
			return false
		}
		if (index == 0 || index == len(value)-1) && (character == '.' || character == '-' || character == '_') {
			return false
		}
	}
	return true
}

func validCanonicalRegistryDomain(value string) bool {
	if len(value) == 0 || len(value) > 253 {
		return false
	}
	labels := strings.Split(value, ".")
	if len(labels) < 2 {
		return false
	}
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, character := range []byte(label) {
			if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '-' {
				continue
			}
			return false
		}
	}
	return true
}
