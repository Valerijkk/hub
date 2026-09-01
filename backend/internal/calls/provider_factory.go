package calls

import (
	"fmt"
	"reflect"

	"github.com/valerii/hub/backend/internal/platform/config"
)

// NewConferenceProvider constructs the configured supported conference provider.
func NewConferenceProvider(callsCfg config.CallsConfig, liveKitCfg config.LiveKitConfig) (ConferenceProvider, error) {
	switch callsCfg.ActiveProvider {
	case ProviderLiveKit:
		return NewLiveKitConferenceProvider(liveKitCfg), nil
	case "jitsi":
		return nil, fmt.Errorf("jitsi provider is not available before phase 152")
	default:
		return nil, fmt.Errorf("unknown calls provider %q", callsCfg.ActiveProvider)
	}
}

// ConferenceProviderRegistry is an immutable startup registry of honest provider capabilities.
type ConferenceProviderRegistry struct {
	providers map[ConferenceProviderID]RegisteredConferenceProvider
}

// NewConferenceProviderRegistry validates and freezes provider registrations.
func NewConferenceProviderRegistry(providers ...RegisteredConferenceProvider) (*ConferenceProviderRegistry, error) {
	if len(providers) == 0 {
		return nil, ErrConferenceProviderRegistryInvalid
	}

	registry := &ConferenceProviderRegistry{
		providers: make(map[ConferenceProviderID]RegisteredConferenceProvider, len(providers)),
	}
	for _, provider := range providers {
		if isNilRegisteredConferenceProvider(provider) {
			return nil, ErrConferenceProviderRegistryInvalid
		}

		providerID := provider.ProviderID()
		if !isKnownConferenceProviderID(providerID) {
			return nil, ErrConferenceProviderRegistryInvalid
		}

		capabilities := provider.Capabilities()
		if !areValidConferenceProviderCapabilities(capabilities) ||
			!providerCapabilitiesMatchInterfaces(provider, capabilities) {
			return nil, ErrConferenceProviderRegistryInvalid
		}
		if _, exists := registry.providers[providerID]; exists {
			return nil, ErrConferenceProviderRegistryDuplicate
		}

		registry.providers[providerID] = provider
	}

	return registry, nil
}

// Lookup resolves one exact provider without fallback or provider invocation.
func (r *ConferenceProviderRegistry) Lookup(providerID ConferenceProviderID) (RegisteredConferenceProvider, error) {
	if r == nil || r.providers == nil {
		return nil, ErrConferenceProviderRegistryInvalid
	}
	if !isKnownConferenceProviderID(providerID) {
		return nil, ErrConferenceProviderUnknown
	}

	provider, exists := r.providers[providerID]
	if !exists {
		return nil, ErrConferenceProviderNotRegistered
	}
	return provider, nil
}

// Require resolves a provider only when it implements one exact known capability.
func (r *ConferenceProviderRegistry) Require(
	providerID ConferenceProviderID,
	capability ConferenceProviderCapability,
) (RegisteredConferenceProvider, error) {
	provider, err := r.Lookup(providerID)
	if err != nil {
		return nil, err
	}
	if !isKnownConferenceProviderCapability(capability) {
		return nil, ErrConferenceProviderRegistryInvalid
	}
	if !providerImplementsConferenceProviderCapability(provider, capability) {
		return nil, ErrConferenceProviderCapabilityUnavailable
	}
	return provider, nil
}

func isNilRegisteredConferenceProvider(provider RegisteredConferenceProvider) bool {
	if provider == nil {
		return true
	}
	value := reflect.ValueOf(provider)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func isKnownConferenceProviderID(providerID ConferenceProviderID) bool {
	switch providerID {
	case ConferenceProviderID(ProviderLiveKit), ConferenceProviderID(ProviderJitsi):
		return true
	default:
		return false
	}
}

func isKnownConferenceProviderCapability(capability ConferenceProviderCapability) bool {
	switch capability {
	case CapabilityIssueCallJoinCredential,
		CapabilityProvisionRoom,
		CapabilityDeleteRoom,
		CapabilityIssueGuestCredential,
		CapabilityUpdateParticipantMetadata:
		return true
	default:
		return false
	}
}

func areValidConferenceProviderCapabilities(capabilities ConferenceProviderCapabilities) bool {
	return capabilities.Has(CapabilityIssueCallJoinCredential) &&
		capabilities&^allConferenceProviderCapabilities == 0
}

func providerCapabilitiesMatchInterfaces(
	provider RegisteredConferenceProvider,
	capabilities ConferenceProviderCapabilities,
) bool {
	checks := []struct {
		capability  ConferenceProviderCapability
		implemented bool
	}{
		{CapabilityProvisionRoom, implementsConferenceRoomProvisioner(provider)},
		{CapabilityDeleteRoom, implementsConferenceRoomDeleter(provider)},
		{CapabilityIssueGuestCredential, implementsConferenceGuestCredentialIssuer(provider)},
		{CapabilityUpdateParticipantMetadata, implementsConferenceParticipantMetadataUpdater(provider)},
	}
	for _, check := range checks {
		if capabilities.Has(check.capability) != check.implemented {
			return false
		}
	}
	return true
}

func providerImplementsConferenceProviderCapability(
	provider RegisteredConferenceProvider,
	capability ConferenceProviderCapability,
) bool {
	switch capability {
	case CapabilityIssueCallJoinCredential:
		return true
	case CapabilityProvisionRoom:
		return implementsConferenceRoomProvisioner(provider)
	case CapabilityDeleteRoom:
		return implementsConferenceRoomDeleter(provider)
	case CapabilityIssueGuestCredential:
		return implementsConferenceGuestCredentialIssuer(provider)
	case CapabilityUpdateParticipantMetadata:
		return implementsConferenceParticipantMetadataUpdater(provider)
	default:
		return false
	}
}

func implementsConferenceRoomProvisioner(provider RegisteredConferenceProvider) bool {
	_, ok := provider.(ConferenceRoomProvisioner)
	return ok
}

func implementsConferenceRoomDeleter(provider RegisteredConferenceProvider) bool {
	_, ok := provider.(ConferenceRoomDeleter)
	return ok
}

func implementsConferenceGuestCredentialIssuer(provider RegisteredConferenceProvider) bool {
	_, ok := provider.(ConferenceGuestCredentialIssuer)
	return ok
}

func implementsConferenceParticipantMetadataUpdater(provider RegisteredConferenceProvider) bool {
	_, ok := provider.(ConferenceParticipantMetadataUpdater)
	return ok
}
