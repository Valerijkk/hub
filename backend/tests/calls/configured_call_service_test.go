//go:build hubtest

package calls_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/livekit/protocol/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	entcall "github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/internal/calls"
	"github.com/valerii/hub/backend/internal/platform/config"
)

func TestNewCallServiceWithConfiguredProviderIssuesServerProjectedLiveKitCredential(t *testing.T) {
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	channelID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	liveKitCfg := config.LiveKitConfig{
		URL:       "https://configured-livekit.hub.test",
		APIKey:    "configured-livekit-test-key",
		APISecret: "configured-livekit-test-secret",
	}
	source := &callJoinCredentialProjectionSourceFake{
		activeCall: &ent.Call{
			ID:        callID,
			RoomName:  "server-loaded-configured-room",
			Status:    entcall.StatusActive,
			ChannelID: &channelID,
			CreatedBy: uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		},
		member:   true,
		userName: "Server Loaded Name",
	}

	service, err := calls.NewCallServiceWithConfiguredProvider(
		config.CallsConfig{ActiveProvider: calls.ProviderLiveKit},
		liveKitCfg,
		nil,
		nil,
		nil,
		source,
	)
	require.NoError(t, err)
	require.NotNil(t, service)

	credential, err := service.IssueJoinCredentialForAuthenticatedUser(context.Background(), callID, userID)

	require.NoError(t, err)
	assert.Equal(t, calls.ProviderLiveKit, credential.Provider)
	assert.Equal(t, "server-loaded-configured-room", credential.RoomName)
	assert.Equal(t, liveKitCfg.URL, credential.ServiceURL)
	require.NotEmpty(t, credential.Token)

	verifier, err := auth.ParseAPIToken(credential.Token)
	require.NoError(t, err)
	claims, grants, err := verifier.Verify(liveKitCfg.APISecret)
	require.NoError(t, err)
	require.NotNil(t, claims)
	require.NotNil(t, grants)
	assert.Equal(t, liveKitCfg.APIKey, claims.Issuer)
	assert.Equal(t, userID.String(), claims.Subject)
	assert.Equal(t, "Server Loaded Name", grants.Name)
	require.NotNil(t, grants.Video)
	assert.True(t, grants.Video.RoomJoin)
	assert.Equal(t, "server-loaded-configured-room", grants.Video.Room)
	require.NotNil(t, claims.Expiry)
	assert.True(t, credential.ExpiresAt.Equal(claims.Expiry.Time()))

	assert.Equal(t, []callJoinCredentialProjectionOperation{
		{name: "load_active_call", callID: callID},
		{name: "is_channel_member", channelID: channelID, authenticatedUserID: userID},
		{name: "get_server_known_call_user_name", authenticatedUserID: userID},
	}, source.operations)
}

func TestNewCallServiceWithConfiguredProviderReturnsFactoryAndProjectorSourceErrors(t *testing.T) {
	liveKitCfg := config.LiveKitConfig{
		URL:       "https://configured-livekit.hub.test",
		APIKey:    "configured-livekit-test-key",
		APISecret: "configured-livekit-test-secret",
	}

	t.Run("jitsi factory error", func(t *testing.T) {
		source := &callJoinCredentialProjectionSourceFake{}

		service, err := calls.NewCallServiceWithConfiguredProvider(
			config.CallsConfig{ActiveProvider: "jitsi"},
			liveKitCfg,
			nil,
			nil,
			nil,
			source,
		)

		assert.Nil(t, service)
		require.EqualError(t, err, "jitsi provider is not available before phase 152")
		assert.Empty(t, source.operations)
	})

	t.Run("typed nil projection source", func(t *testing.T) {
		var source *callJoinCredentialProjectionSourceFake

		service, err := calls.NewCallServiceWithConfiguredProvider(
			config.CallsConfig{ActiveProvider: calls.ProviderLiveKit},
			liveKitCfg,
			nil,
			nil,
			nil,
			source,
		)

		assert.Nil(t, service)
		require.EqualError(t, err, "call join credential projection source is required")
	})
}
