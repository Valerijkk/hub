package calls

import (
	"errors"

	"github.com/valerii/hub/backend/internal/platform/config"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// NewCallServiceWithConfiguredProvider constructs the Calls facade from the configured provider.
func NewCallServiceWithConfiguredProvider(
	callsCfg config.CallsConfig,
	liveKitCfg config.LiveKitConfig,
	repo *EntRepository,
	nc *natsclient.Client,
	chatRepo *ChatRepository,
	source CallJoinCredentialProjectionSource,
) (*CallService, error) {
	provider, err := NewConferenceProvider(callsCfg, liveKitCfg)
	if err != nil {
		return nil, err
	}

	issuer, ok := provider.(CallJoinCredentialIssuer)
	if !ok {
		return nil, errors.New("configured calls provider does not support join credential issuance")
	}

	projector, err := NewCallJoinCredentialProjector(source, issuer)
	if err != nil {
		return nil, err
	}

	return NewCallServiceWithProviderAndCredentialProjector(provider, repo, nc, chatRepo, projector)
}