package mail

import (
	"context"
	"errors"
	"reflect"
)

// PublicDNSDomainProvisioner verifies only public DNS observations returned by
// an upstream domain provisioner. It never exposes provider or DNS details.
type PublicDNSDomainProvisioner struct {
	upstream MailboxDomainProvisioner
	verifier *PublicDNSVerifier
}

var _ MailboxDomainProvisioner = (*PublicDNSDomainProvisioner)(nil)

// NewPublicDNSDomainProvisioner binds an upstream domain provisioner and a
// public DNS verifier without performing I/O.
func NewPublicDNSDomainProvisioner(upstream MailboxDomainProvisioner, verifier *PublicDNSVerifier) (*PublicDNSDomainProvisioner, error) {
	if publicDNSDomainNil(upstream) || verifier == nil {
		return nil, publicDNSEngineError(EngineErrorRejectedPolicy)
	}
	return &PublicDNSDomainProvisioner{upstream: upstream, verifier: verifier}, nil
}

// GetDomain returns an exact absence or a public-DNS-projected domain state.
// Upstream verification claims are never trusted.
func (provisioner *PublicDNSDomainProvisioner) GetDomain(ctx context.Context, domain string) (state EngineDomainState, err error) {
	defer func() {
		if recover() != nil {
			state = EngineDomainState{}
			err = publicDNSEngineError(EngineErrorUnavailable)
		}
	}()
	if !publicDNSDomainProvisionerUsable(provisioner) || publicDNSDomainNil(ctx) || !validCanonicalRegistryDomain(domain) {
		return EngineDomainState{}, publicDNSEngineError(EngineErrorRejectedPolicy)
	}
	if ctx.Err() != nil {
		return EngineDomainState{}, publicDNSEngineError(EngineErrorUnavailable)
	}

	upstreamState, upstreamErr := provisioner.upstream.GetDomain(ctx, domain)
	if ctx.Err() != nil {
		return EngineDomainState{}, publicDNSEngineError(EngineErrorUnavailable)
	}
	if upstreamErr != nil {
		return EngineDomainState{}, publicDNSDomainReadError(upstreamErr)
	}
	if absent, valid := publicDNSDomainExactAbsent(upstreamState); valid {
		return absent, nil
	}
	dnsRequired, valid := publicDNSDomainDNSRequired(upstreamState)
	if !valid {
		return EngineDomainState{}, publicDNSEngineError(EngineErrorUnavailable)
	}
	if ctx.Err() != nil {
		return EngineDomainState{}, publicDNSEngineError(EngineErrorUnavailable)
	}
	verification, verifyErr := provisioner.verifier.Verify(ctx, append([]EngineDNSRecord(nil), dnsRequired.DNSRecords...))
	if ctx.Err() != nil || verifyErr != nil {
		return EngineDomainState{}, publicDNSEngineError(EngineErrorUnavailable)
	}
	projected, valid := publicDNSDomainProject(dnsRequired, verification)
	if !valid {
		return EngineDomainState{}, publicDNSEngineError(EngineErrorUnavailable)
	}
	return projected, nil
}

// EnsureDomain dispatches exactly one upstream ensure and then verifies its
// authoritative DNS-required projection. Any post-dispatch uncertainty is
// indeterminate so the durable worker reconciles before retrying.
func (provisioner *PublicDNSDomainProvisioner) EnsureDomain(ctx context.Context, spec DomainSpec) (state EngineDomainState, err error) {
	defer func() {
		if recover() != nil {
			state = EngineDomainState{}
			err = publicDNSEngineError(EngineErrorIndeterminate)
		}
	}()
	if !publicDNSDomainProvisionerUsable(provisioner) || publicDNSDomainNil(ctx) || !validRegistryProvisioningDomainSpec(spec) {
		return EngineDomainState{}, publicDNSEngineError(EngineErrorRejectedPolicy)
	}
	if ctx.Err() != nil {
		return EngineDomainState{}, publicDNSEngineError(EngineErrorUnavailable)
	}

	upstreamState, upstreamErr := provisioner.upstream.EnsureDomain(ctx, spec)
	if ctx.Err() != nil || upstreamErr != nil {
		return EngineDomainState{}, publicDNSEngineError(EngineErrorIndeterminate)
	}
	dnsRequired, valid := publicDNSDomainDNSRequired(upstreamState)
	if !valid {
		return EngineDomainState{}, publicDNSEngineError(EngineErrorIndeterminate)
	}
	if ctx.Err() != nil {
		return EngineDomainState{}, publicDNSEngineError(EngineErrorIndeterminate)
	}
	verification, verifyErr := provisioner.verifier.Verify(ctx, append([]EngineDNSRecord(nil), dnsRequired.DNSRecords...))
	if ctx.Err() != nil || verifyErr != nil {
		return EngineDomainState{}, publicDNSEngineError(EngineErrorIndeterminate)
	}
	projected, valid := publicDNSDomainProject(dnsRequired, verification)
	if !valid {
		return EngineDomainState{}, publicDNSEngineError(EngineErrorIndeterminate)
	}
	return projected, nil
}

func publicDNSDomainProvisionerUsable(provisioner *PublicDNSDomainProvisioner) bool {
	return provisioner != nil && !publicDNSDomainNil(provisioner.upstream) && provisioner.verifier != nil
}

func publicDNSDomainExactAbsent(state EngineDomainState) (EngineDomainState, bool) {
	if state.DomainID != "" || state.State != EngineDomainAbsent || state.Applied ||
		state.DNSRecords == nil || len(state.DNSRecords) != 0 ||
		state.Verification != EngineVerificationUnknown || state.DKIM != EngineVerificationUnknown {
		return EngineDomainState{}, false
	}
	return EngineDomainState{
		State:        EngineDomainAbsent,
		DNSRecords:   []EngineDNSRecord{},
		Verification: EngineVerificationUnknown,
		DKIM:         EngineVerificationUnknown,
	}, true
}

func publicDNSDomainDNSRequired(state EngineDomainState) (EngineDomainState, bool) {
	if !validOpaqueID(state.DomainID) || state.State != EngineDomainDNSRequired || !state.Applied ||
		state.DNSRecords == nil || len(state.DNSRecords) == 0 ||
		state.Verification != EngineVerificationPending || state.DKIM != EngineVerificationPending {
		return EngineDomainState{}, false
	}
	records := make([]EngineDNSRecord, len(state.DNSRecords))
	copy(records, state.DNSRecords)
	required := make(RequiredDNSRecords, len(records))
	for index, record := range records {
		required[index] = RequiredDNSRecord{
			Type:  record.Type,
			Name:  record.Name,
			Value: record.Value,
			TTL:   record.TTL,
		}
	}
	if required.Validate() != nil {
		return EngineDomainState{}, false
	}
	if _, valid := publicDNSExpectedGroups(records); !valid {
		return EngineDomainState{}, false
	}
	return EngineDomainState{
		DomainID:     state.DomainID,
		State:        EngineDomainDNSRequired,
		Applied:      true,
		DNSRecords:   records,
		Verification: EngineVerificationPending,
		DKIM:         EngineVerificationPending,
	}, true
}

func publicDNSDomainProject(state EngineDomainState, verification PublicDNSVerification) (EngineDomainState, bool) {
	if !verification.Valid() {
		return EngineDomainState{}, false
	}
	projected := state
	projected.DNSRecords = make([]EngineDNSRecord, len(state.DNSRecords))
	copy(projected.DNSRecords, state.DNSRecords)
	projected.Verification = verification.Verification
	projected.DKIM = verification.DKIM
	switch verification.Verification {
	case EngineVerificationVerified:
		if verification.DKIM != EngineVerificationVerified {
			return EngineDomainState{}, false
		}
		projected.State = EngineDomainVerified
	case EngineVerificationPending, EngineVerificationFailed:
		projected.State = EngineDomainDNSRequired
	default:
		return EngineDomainState{}, false
	}
	return projected, true
}

func publicDNSDomainReadError(err error) error {
	var engineErr *EngineError
	if errors.As(err, &engineErr) && engineErr != nil && publicDNSDomainKnownEngineErrorKind(engineErr.Kind) {
		return publicDNSEngineError(engineErr.Kind)
	}
	return publicDNSEngineError(EngineErrorUnavailable)
}

func publicDNSDomainKnownEngineErrorKind(kind EngineErrorKind) bool {
	switch kind {
	case EngineErrorNotFound,
		EngineErrorConflict,
		EngineErrorInvalidCursor,
		EngineErrorUnavailable,
		EngineErrorIndeterminate,
		EngineErrorRejectedPolicy,
		EngineErrorUnauthorizedAccount,
		EngineErrorRestartRequired,
		EngineErrorPayloadTooLarge,
		EngineErrorRateLimited,
		EngineErrorIdentityUnavailable,
		EngineErrorRecipientLimit:
		return true
	default:
		return false
	}
}

func publicDNSDomainNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface, reflect.Func, reflect.Chan:
		return reflected.IsNil()
	default:
		return false
	}
}
