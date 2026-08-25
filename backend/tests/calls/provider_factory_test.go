//go:build hubtest

package calls_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/calls"
	"github.com/valerii/hub/backend/internal/platform/config"
)

func TestConferenceProviderFactoryConstructsLiveKitProvider(t *testing.T) {
	liveKitCfg := providerFactoryLiveKitConfig()

	provider, err := calls.NewConferenceProvider(
		config.CallsConfig{ActiveProvider: calls.ProviderLiveKit},
		liveKitCfg,
	)
	require.NoError(t, err)
	require.NotNil(t, provider)

	credential, err := provider.IssueJoinCredential("call-factory", uuid.NewString(), "Factory User")
	require.NoError(t, err)
	assert.Equal(t, calls.ProviderLiveKit, credential.Provider)
	assert.Equal(t, "call-factory", credential.RoomName)
	assert.Equal(t, liveKitCfg.URL, credential.ServiceURL)
	require.NotEmpty(t, credential.Token)
}

func TestConferenceProviderFactoryRejectsConfiguredJitsiBeforePhase152(t *testing.T) {
	provider, err := calls.NewConferenceProvider(
		config.CallsConfig{
			ActiveProvider: "jitsi",
			Jitsi: config.JitsiConfig{
				Enabled:    true,
				PublicURL:  "https://meet.hub.test",
				Host:       "meet.hub.test",
				AppID:      "hub",
				AppSecret:  strings.Repeat("jitsi-secret-", 3),
				XMPPDomain: "meet.hub.test",
			},
		},
		providerFactoryLiveKitConfig(),
	)

	require.Nil(t, provider)
	require.ErrorContains(t, err, "jitsi provider is not available before phase 152")
}

func TestConferenceProviderFactoryRejectsUnknownProvider(t *testing.T) {
	provider, err := calls.NewConferenceProvider(
		config.CallsConfig{ActiveProvider: "unknown-provider"},
		providerFactoryLiveKitConfig(),
	)

	require.Nil(t, provider)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown-provider")
}

func providerFactoryLiveKitConfig() config.LiveKitConfig {
	return config.LiveKitConfig{
		URL:       "https://livekit.hub.test",
		APIKey:    "provider-factory-key",
		APISecret: "provider-factory-secret",
	}
}
