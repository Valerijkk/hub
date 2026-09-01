//go:build hubtest

package calls_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/calls"
	"github.com/valerii/hub/backend/internal/platform/config"
)

type registryProviderProbe struct {
	id           calls.ConferenceProviderID
	capabilities calls.ConferenceProviderCapabilities
	providerIDCalls int
	capabilityCalls int
	issueCalls     int
}

func (p *registryProviderProbe) ProviderID() calls.ConferenceProviderID {
	p.providerIDCalls++
	return p.id
}

func (p *registryProviderProbe) Capabilities() calls.ConferenceProviderCapabilities {
	p.capabilityCalls++
	return p.capabilities
}

func (p *registryProviderProbe) IssueCallJoinCredential(
	*ent.Call,
	uuid.UUID,
	string,
) (calls.JoinCredential, error) {
	p.issueCalls++
	return calls.JoinCredential{}, nil
}

type fullRegistryProviderProbe struct {
	*registryProviderProbe
}

type provisionRegistryProviderProbe struct {
	*registryProviderProbe
}

type deleteRegistryProviderProbe struct {
	*registryProviderProbe
}

type guestRegistryProviderProbe struct {
	*registryProviderProbe
}

type metadataRegistryProviderProbe struct {
	*registryProviderProbe
}

type legacyRoomStringCredentialIssuer interface {
	IssueJoinCredential(string, string, string) (calls.JoinCredential, error)
}

func (*fullRegistryProviderProbe) Provision(context.Context, string, uint32) error {
	return nil
}

func (*fullRegistryProviderProbe) Delete(context.Context, string) error {
	return nil
}

func (*fullRegistryProviderProbe) IssueGuestCredential(string, string) (calls.GuestJoinCredential, error) {
	return calls.GuestJoinCredential{}, nil
}

func (*fullRegistryProviderProbe) UpdateParticipantState(context.Context, string, string, string) error {
	return nil
}

func (*provisionRegistryProviderProbe) Provision(context.Context, string, uint32) error {
	return nil
}

func (*deleteRegistryProviderProbe) Delete(context.Context, string) error {
	return nil
}

func (*guestRegistryProviderProbe) IssueGuestCredential(string, string) (calls.GuestJoinCredential, error) {
	return calls.GuestJoinCredential{}, nil
}

func (*metadataRegistryProviderProbe) UpdateParticipantState(context.Context, string, string, string) error {
	return nil
}

func TestConferenceProviderCapabilitiesAreClosedAndQueryable(t *testing.T) {
	assert.Equal(t, calls.ConferenceProviderID(calls.ProviderLiveKit), calls.ConferenceProviderID("livekit"))
	assert.Equal(t, calls.ConferenceProviderID(calls.ProviderJitsi), calls.ConferenceProviderID("jitsi"))
	assert.Equal(t, uint8(1), uint8(calls.CapabilityIssueCallJoinCredential))
	assert.Equal(t, uint8(2), uint8(calls.CapabilityProvisionRoom))
	assert.Equal(t, uint8(4), uint8(calls.CapabilityDeleteRoom))
	assert.Equal(t, uint8(8), uint8(calls.CapabilityIssueGuestCredential))
	assert.Equal(t, uint8(16), uint8(calls.CapabilityUpdateParticipantMetadata))

	all := allConferenceProviderCapabilities()
	for _, capability := range []calls.ConferenceProviderCapability{
		calls.CapabilityIssueCallJoinCredential,
		calls.CapabilityProvisionRoom,
		calls.CapabilityDeleteRoom,
		calls.CapabilityIssueGuestCredential,
		calls.CapabilityUpdateParticipantMetadata,
	} {
		assert.True(t, all.Has(capability))
	}
	assert.False(t, calls.ConferenceProviderCapabilities(calls.CapabilityIssueCallJoinCredential).Has(calls.CapabilityProvisionRoom))
}

func TestConferenceProviderRegistryRejectsInvalidRegistrations(t *testing.T) {
	var nilProvider calls.RegisteredConferenceProvider
	var typedNilProvider *registryProviderProbe

	testCases := []struct {
		name      string
		providers []calls.RegisteredConferenceProvider
	}{
		{name: "empty"},
		{name: "nil interface", providers: []calls.RegisteredConferenceProvider{nilProvider}},
		{name: "typed nil", providers: []calls.RegisteredConferenceProvider{typedNilProvider}},
		{name: "empty id", providers: []calls.RegisteredConferenceProvider{newRegistryProbe("", credentialCapability())}},
		{name: "unknown id", providers: []calls.RegisteredConferenceProvider{newRegistryProbe("other", credentialCapability())}},
		{name: "padded id", providers: []calls.RegisteredConferenceProvider{newRegistryProbe(" jitsi", credentialCapability())}},
		{name: "case drift id", providers: []calls.RegisteredConferenceProvider{newRegistryProbe("LiveKit", credentialCapability())}},
		{name: "missing credential capability", providers: []calls.RegisteredConferenceProvider{newRegistryProbe(calls.ProviderJitsi, 0)}},
		{name: "unknown capability bit", providers: []calls.RegisteredConferenceProvider{newRegistryProbe(calls.ProviderJitsi, credentialCapability()|calls.ConferenceProviderCapabilities(1<<7))}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry, err := calls.NewConferenceProviderRegistry(testCase.providers...)
			require.Nil(t, registry)
			requireExactProviderError(t, calls.ErrConferenceProviderRegistryInvalid, err)
		})
	}
}

func TestConferenceProviderRegistryRejectsDuplicateWithoutReturningRegistry(t *testing.T) {
	first := newRegistryProbe(calls.ProviderLiveKit, credentialCapability())
	second := newRegistryProbe(calls.ProviderLiveKit, credentialCapability())

	registry, err := calls.NewConferenceProviderRegistry(first, second)

	require.Nil(t, registry)
	requireExactProviderError(t, calls.ErrConferenceProviderRegistryDuplicate, err)
	assert.Equal(t, 0, first.issueCalls)
	assert.Equal(t, 0, second.issueCalls)
}

func TestConferenceProviderRegistryRejectsCapabilityInterfaceMismatchBothWays(t *testing.T) {
	testCases := []struct {
		name       string
		capability calls.ConferenceProviderCapability
	}{
		{name: "provision", capability: calls.CapabilityProvisionRoom},
		{name: "delete", capability: calls.CapabilityDeleteRoom},
		{name: "guest credential", capability: calls.CapabilityIssueGuestCredential},
		{name: "participant metadata", capability: calls.CapabilityUpdateParticipantMetadata},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name+" advertised without interface", func(t *testing.T) {
			provider := newRegistryProbe(
				calls.ProviderLiveKit,
				credentialCapability()|calls.ConferenceProviderCapabilities(testCase.capability),
			)
			registry, err := calls.NewConferenceProviderRegistry(provider)
			require.Nil(t, registry)
			requireExactProviderError(t, calls.ErrConferenceProviderRegistryInvalid, err)
		})

		t.Run(testCase.name+" interface omitted from capabilities", func(t *testing.T) {
			provider := newOptionalCapabilityProvider(testCase.capability, credentialCapability())
			registry, err := calls.NewConferenceProviderRegistry(provider)
			require.Nil(t, registry)
			requireExactProviderError(t, calls.ErrConferenceProviderRegistryInvalid, err)
		})
	}
}

func TestConferenceProviderRegistryLookupIsDeterministicAndSideEffectFree(t *testing.T) {
	for _, reverseOrder := range []bool{false, true} {
		name := "livekit then jitsi"
		if reverseOrder {
			name = "jitsi then livekit"
		}
		t.Run(name, func(t *testing.T) {
			liveKit := newRegistryProbe(calls.ProviderLiveKit, credentialCapability())
			jitsi := newRegistryProbe(calls.ProviderJitsi, credentialCapability())
			providers := []calls.RegisteredConferenceProvider{liveKit, jitsi}
			if reverseOrder {
				providers[0], providers[1] = providers[1], providers[0]
			}
			registry, err := calls.NewConferenceProviderRegistry(providers...)
			require.NoError(t, err)
			require.NotNil(t, registry)
			liveKitIDCalls := liveKit.providerIDCalls
			liveKitCapabilityCalls := liveKit.capabilityCalls
			jitsiIDCalls := jitsi.providerIDCalls
			jitsiCapabilityCalls := jitsi.capabilityCalls

			resolvedLiveKit, err := registry.Lookup(calls.ConferenceProviderID(calls.ProviderLiveKit))
			require.NoError(t, err)
			assert.Same(t, liveKit, resolvedLiveKit)

			resolvedJitsi, err := registry.Lookup(calls.ConferenceProviderID(calls.ProviderJitsi))
			require.NoError(t, err)
			assert.Same(t, jitsi, resolvedJitsi)

			resolvedLiveKit, err = registry.Require(
				calls.ConferenceProviderID(calls.ProviderLiveKit),
				calls.CapabilityIssueCallJoinCredential,
			)
			require.NoError(t, err)
			assert.Same(t, liveKit, resolvedLiveKit)

			resolvedJitsi, err = registry.Require(
				calls.ConferenceProviderID(calls.ProviderJitsi),
				calls.CapabilityIssueCallJoinCredential,
			)
			require.NoError(t, err)
			assert.Same(t, jitsi, resolvedJitsi)

			assert.Equal(t, liveKitIDCalls, liveKit.providerIDCalls)
			assert.Equal(t, liveKitCapabilityCalls, liveKit.capabilityCalls)
			assert.Equal(t, jitsiIDCalls, jitsi.providerIDCalls)
			assert.Equal(t, jitsiCapabilityCalls, jitsi.capabilityCalls)
			assert.Equal(t, 0, liveKit.issueCalls)
			assert.Equal(t, 0, jitsi.issueCalls)
		})
	}
}

func TestConferenceProviderRegistryDistinguishesUnknownAndKnownOmittedWithoutFallback(t *testing.T) {
	liveKit := newRegistryProbe(calls.ProviderLiveKit, credentialCapability())
	registry, err := calls.NewConferenceProviderRegistry(liveKit)
	require.NoError(t, err)

	for _, id := range []calls.ConferenceProviderID{"", "unknown", " livekit", "JITSI"} {
		resolved, lookupErr := registry.Lookup(id)
		require.Nil(t, resolved)
		requireExactProviderError(t, calls.ErrConferenceProviderUnknown, lookupErr)
	}

	resolved, lookupErr := registry.Lookup(calls.ConferenceProviderID(calls.ProviderJitsi))
	require.Nil(t, resolved)
	requireExactProviderError(t, calls.ErrConferenceProviderNotRegistered, lookupErr)
	assert.Equal(t, 0, liveKit.issueCalls)
}

func TestConferenceProviderRegistryRejectsNilAndZeroReceiversBeforeArguments(t *testing.T) {
	var nilRegistry *calls.ConferenceProviderRegistry
	var zeroRegistry calls.ConferenceProviderRegistry
	invalidID := calls.ConferenceProviderID("unknown")
	invalidCapability := calls.ConferenceProviderCapability(0)

	for _, registry := range []*calls.ConferenceProviderRegistry{nilRegistry, &zeroRegistry} {
		resolved, err := registry.Lookup(invalidID)
		require.Nil(t, resolved)
		requireExactProviderError(t, calls.ErrConferenceProviderRegistryInvalid, err)

		resolved, err = registry.Require(invalidID, invalidCapability)
		require.Nil(t, resolved)
		requireExactProviderError(t, calls.ErrConferenceProviderRegistryInvalid, err)
	}
}

func TestConferenceProviderRegistryRequireValidatesIdentityBeforeCapability(t *testing.T) {
	liveKit := newRegistryProbe(calls.ProviderLiveKit, credentialCapability())
	registry, err := calls.NewConferenceProviderRegistry(liveKit)
	require.NoError(t, err)

	resolved, err := registry.Require("unknown", 0)
	require.Nil(t, resolved)
	requireExactProviderError(t, calls.ErrConferenceProviderUnknown, err)

	resolved, err = registry.Require(calls.ConferenceProviderID(calls.ProviderJitsi), 0)
	require.Nil(t, resolved)
	requireExactProviderError(t, calls.ErrConferenceProviderNotRegistered, err)
}

func TestConferenceProviderRegistryRequireAcceptsOneKnownCapabilityOnly(t *testing.T) {
	liveKit := newRegistryProbe(calls.ProviderLiveKit, credentialCapability())
	registry, err := calls.NewConferenceProviderRegistry(liveKit)
	require.NoError(t, err)

	resolved, err := registry.Require(
		calls.ConferenceProviderID(calls.ProviderLiveKit),
		calls.CapabilityIssueCallJoinCredential,
	)
	require.NoError(t, err)
	assert.Same(t, liveKit, resolved)
	assert.Equal(t, 0, liveKit.issueCalls)

	for _, capability := range []calls.ConferenceProviderCapability{
		0,
		calls.CapabilityIssueCallJoinCredential | calls.CapabilityProvisionRoom,
		1 << 7,
	} {
		resolved, requireErr := registry.Require(calls.ConferenceProviderID(calls.ProviderLiveKit), capability)
		require.Nil(t, resolved)
		requireExactProviderError(t, calls.ErrConferenceProviderRegistryInvalid, requireErr)
	}

	resolved, err = registry.Require(
		calls.ConferenceProviderID(calls.ProviderLiveKit),
		calls.CapabilityProvisionRoom,
	)
	require.Nil(t, resolved)
	requireExactProviderError(t, calls.ErrConferenceProviderCapabilityUnavailable, err)
	assert.Equal(t, 0, liveKit.issueCalls)
}

func TestLiveKitRegisteredProviderAdvertisesItsCompleteExistingSurface(t *testing.T) {
	provider := calls.NewLiveKitConferenceProvider(config.LiveKitConfig{
		URL:       "https://livekit.hub.test",
		APIKey:    "registry-livekit-key",
		APISecret: "registry-livekit-secret",
	})

	var registered calls.RegisteredConferenceProvider = provider
	assert.Equal(t, calls.ConferenceProviderID(calls.ProviderLiveKit), registered.ProviderID())
	assert.Equal(t, allConferenceProviderCapabilities(), registered.Capabilities())
	assert.Implements(t, (*calls.ConferenceProvider)(nil), provider)
	assert.Implements(t, (*calls.ConferenceRoomProvisioner)(nil), provider)
	assert.Implements(t, (*calls.ConferenceRoomDeleter)(nil), provider)
	assert.Implements(t, (*calls.ConferenceGuestCredentialIssuer)(nil), provider)
	assert.Implements(t, (*calls.ConferenceParticipantMetadataUpdater)(nil), provider)

	registry, err := calls.NewConferenceProviderRegistry(provider)
	require.NoError(t, err)
	for _, capability := range []calls.ConferenceProviderCapability{
		calls.CapabilityIssueCallJoinCredential,
		calls.CapabilityProvisionRoom,
		calls.CapabilityDeleteRoom,
		calls.CapabilityIssueGuestCredential,
		calls.CapabilityUpdateParticipantMetadata,
	} {
		resolved, requireErr := registry.Require(calls.ConferenceProviderID(calls.ProviderLiveKit), capability)
		require.NoError(t, requireErr)
		assert.Same(t, provider, resolved)
	}
}

func TestJitsiRegisteredProviderIsCredentialOnlyWithoutIssuingCredential(t *testing.T) {
	provider, err := calls.NewJitsiConferenceProvider(validRegistryJitsiConfig())
	require.NoError(t, err)
	require.NotNil(t, provider)

	var registered calls.RegisteredConferenceProvider = provider
	assert.Equal(t, calls.ConferenceProviderID(calls.ProviderJitsi), registered.ProviderID())
	assert.Equal(t, credentialCapability(), registered.Capabilities())
	assert.NotImplements(t, (*calls.ConferenceProvider)(nil), provider)
	assert.NotImplements(t, (*calls.ConferenceRoomProvisioner)(nil), provider)
	assert.NotImplements(t, (*calls.ConferenceRoomDeleter)(nil), provider)
	assert.NotImplements(t, (*calls.ConferenceGuestCredentialIssuer)(nil), provider)
	assert.NotImplements(t, (*calls.ConferenceParticipantMetadataUpdater)(nil), provider)
	assert.NotImplements(t, (*legacyRoomStringCredentialIssuer)(nil), provider)

	registry, err := calls.NewConferenceProviderRegistry(provider)
	require.NoError(t, err)
	resolved, err := registry.Require(
		calls.ConferenceProviderID(calls.ProviderJitsi),
		calls.CapabilityIssueCallJoinCredential,
	)
	require.NoError(t, err)
	assert.Same(t, provider, resolved)
}

func TestJitsiRegisteredProviderConstructorErrorsDoNotDiscloseConfiguration(t *testing.T) {
	cfg := validRegistryJitsiConfig()
	cfg.Enabled = false

	provider, err := calls.NewJitsiConferenceProvider(cfg)

	require.Nil(t, provider)
	require.Error(t, err)
	for _, secretFact := range []string{
		cfg.AppSecret,
		cfg.AppID,
		cfg.PublicURL,
		cfg.Host,
		cfg.XMPPDomain,
	} {
		assert.NotContains(t, err.Error(), secretFact)
	}
}

func TestLegacyFactoryStillRejectsJitsiActivationExactly(t *testing.T) {
	provider, err := calls.NewConferenceProvider(
		config.CallsConfig{
			ActiveProvider: calls.ProviderJitsi,
			Jitsi:          validRegistryJitsiConfig(),
		},
		config.LiveKitConfig{
			URL:       "https://livekit.hub.test",
			APIKey:    "registry-factory-key",
			APISecret: "registry-factory-secret",
		},
	)

	require.Nil(t, provider)
	assert.EqualError(t, err, "jitsi provider is not available before phase 152")
}

func newRegistryProbe(id string, capabilities calls.ConferenceProviderCapabilities) *registryProviderProbe {
	return &registryProviderProbe{
		id:           calls.ConferenceProviderID(id),
		capabilities: capabilities,
	}
}

func newOptionalCapabilityProvider(
	capability calls.ConferenceProviderCapability,
	capabilities calls.ConferenceProviderCapabilities,
) calls.RegisteredConferenceProvider {
	probe := newRegistryProbe(calls.ProviderLiveKit, capabilities)
	switch capability {
	case calls.CapabilityProvisionRoom:
		return &provisionRegistryProviderProbe{registryProviderProbe: probe}
	case calls.CapabilityDeleteRoom:
		return &deleteRegistryProviderProbe{registryProviderProbe: probe}
	case calls.CapabilityIssueGuestCredential:
		return &guestRegistryProviderProbe{registryProviderProbe: probe}
	case calls.CapabilityUpdateParticipantMetadata:
		return &metadataRegistryProviderProbe{registryProviderProbe: probe}
	default:
		panic("test helper received an unsupported capability")
	}
}

func credentialCapability() calls.ConferenceProviderCapabilities {
	return calls.ConferenceProviderCapabilities(calls.CapabilityIssueCallJoinCredential)
}

func allConferenceProviderCapabilities() calls.ConferenceProviderCapabilities {
	return calls.ConferenceProviderCapabilities(
		calls.CapabilityIssueCallJoinCredential |
			calls.CapabilityProvisionRoom |
			calls.CapabilityDeleteRoom |
			calls.CapabilityIssueGuestCredential |
			calls.CapabilityUpdateParticipantMetadata,
	)
}

func validRegistryJitsiConfig() config.JitsiConfig {
	return config.JitsiConfig{
		Enabled:    true,
		Host:       "meet.registry.hub.test",
		PublicURL:  "https://meet.registry.hub.test",
		AppID:      "registry-hub",
		AppSecret:  "registry-jitsi-secret-0123456789-abcdef",
		XMPPDomain: "meet.registry.hub.test",
	}
}

func requireExactProviderError(t *testing.T, want, got error) {
	t.Helper()
	require.Error(t, got)
	assert.True(t, got == want, "expected exact sentinel %q, got %q", want, got)
}
