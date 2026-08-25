//go:build hubtest

package calls_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/livekit/protocol/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/calls"
	"github.com/valerii/hub/backend/internal/platform/config"
)

func TestLiveKitConferenceProviderIssuesBackwardCompatibleUserJoinCredential(t *testing.T) {
	cfg := liveKitProviderConfig()
	provider := calls.NewLiveKitConferenceProvider(cfg)
	userID := uuid.NewString()

	credential, err := provider.IssueJoinCredential(
		"call-7c4d8dc7-34b4-4c21-8af1-97a6b8e08e02",
		userID,
		"Ada Lovelace",
	)
	require.NoError(t, err)

	assert.Equal(t, calls.ProviderLiveKit, credential.Provider)
	assert.Equal(t, "call-7c4d8dc7-34b4-4c21-8af1-97a6b8e08e02", credential.RoomName)
	assert.Equal(t, cfg.URL, credential.ServiceURL)
	require.NotEmpty(t, credential.Token)

	assertLiveKitCredential(t, cfg, credential.Token, userID, "Ada Lovelace", credential.RoomName)
}

func TestLiveKitConferenceProviderIssuesGuestCredentialWithFreshGuestIdentity(t *testing.T) {
	cfg := liveKitProviderConfig()
	provider := calls.NewLiveKitConferenceProvider(cfg)
	roomName := "call-d5d76dd1-7a91-47c4-8778-3d0c95b69ff7"

	first, err := provider.IssueGuestCredential(roomName, "Guest One")
	require.NoError(t, err)
	second, err := provider.IssueGuestCredential(roomName, "Guest Two")
	require.NoError(t, err)

	for _, credential := range []calls.GuestJoinCredential{first, second} {
		require.True(t, strings.HasPrefix(credential.Identity, "guest-"))
		_, err := uuid.Parse(strings.TrimPrefix(credential.Identity, "guest-"))
		require.NoError(t, err, "guest identity must contain a full UUID")

		assert.Equal(t, calls.ProviderLiveKit, credential.Join.Provider)
		assert.Equal(t, roomName, credential.Join.RoomName)
		assert.Equal(t, cfg.URL, credential.Join.ServiceURL)
		require.NotEmpty(t, credential.Join.Token)
	}

	assert.NotEqual(t, first.Identity, second.Identity, "each guest credential must have fresh identity entropy")
	assertLiveKitCredential(t, cfg, first.Join.Token, first.Identity, "Guest One", roomName)
	assertLiveKitCredential(t, cfg, second.Join.Token, second.Identity, "Guest Two", roomName)
}

func TestLiveKitConferenceProviderConstructionAndCredentialIssueAreNetworkFree(t *testing.T) {
	cfg := liveKitProviderConfig()

	require.NotPanics(t, func() {
		provider := calls.NewLiveKitConferenceProvider(cfg)
		require.NotNil(t, provider)

		credential, err := provider.IssueJoinCredential("call-local", uuid.NewString(), "Local User")
		require.NoError(t, err)
		assert.Equal(t, cfg.URL, credential.ServiceURL)
	})
}

func liveKitProviderConfig() config.LiveKitConfig {
	return config.LiveKitConfig{
		URL:       "https://livekit.hub.test",
		APIKey:    "livekit-provider-key",
		APISecret: "livekit-provider-secret",
	}
}

func assertLiveKitCredential(
	t *testing.T,
	cfg config.LiveKitConfig,
	token string,
	wantIdentity string,
	wantName string,
	wantRoom string,
) {
	t.Helper()

	verifier, err := auth.ParseAPIToken(token)
	require.NoError(t, err)
	claims, grants, err := verifier.Verify(cfg.APISecret)
	require.NoError(t, err)
	require.NotNil(t, claims)
	require.NotNil(t, grants)

	assert.Equal(t, cfg.APIKey, claims.Issuer)
	assert.Equal(t, wantIdentity, claims.Subject)
	assert.Equal(t, wantName, grants.Name)
	require.NotNil(t, grants.Video, "credential must contain a LiveKit video grant")
	assert.True(t, grants.Video.RoomJoin)
	assert.Equal(t, wantRoom, grants.Video.Room)
	require.NotNil(t, claims.NotBefore)
	require.NotNil(t, claims.Expiry)
	assert.WithinDuration(t, claims.NotBefore.Time().Add(24*time.Hour), claims.Expiry.Time(), time.Second)
}