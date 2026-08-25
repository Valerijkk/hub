package calls

import (
	"fmt"

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