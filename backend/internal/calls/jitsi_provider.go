package calls

import (
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/platform/config"
)

// JitsiConferenceProvider is a credential-only dormant registry adapter.
type JitsiConferenceProvider struct {
	issuer *jitsiCredentialIssuer
}

var _ RegisteredConferenceProvider = (*JitsiConferenceProvider)(nil)

// NewJitsiConferenceProvider validates self-hosted config without activating Jitsi.
func NewJitsiConferenceProvider(cfg config.JitsiConfig) (*JitsiConferenceProvider, error) {
	issuer, err := newJitsiCredentialIssuer(cfg, time.Now)
	if err != nil {
		return nil, err
	}
	return &JitsiConferenceProvider{issuer: issuer}, nil
}

// ProviderID returns the immutable Jitsi registry identity.
func (*JitsiConferenceProvider) ProviderID() ConferenceProviderID {
	return ConferenceProviderID(ProviderJitsi)
}

// Capabilities reports only trusted call-aware credential issuance.
func (*JitsiConferenceProvider) Capabilities() ConferenceProviderCapabilities {
	return ConferenceProviderCapabilities(CapabilityIssueCallJoinCredential)
}

// IssueCallJoinCredential delegates trusted server-loaded facts to the local issuer.
func (p *JitsiConferenceProvider) IssueCallJoinCredential(
	call *ent.Call,
	authenticatedUserID uuid.UUID,
	serverKnownUserName string,
) (JoinCredential, error) {
	return p.issuer.IssueCallJoinCredential(call, authenticatedUserID, serverKnownUserName)
}
