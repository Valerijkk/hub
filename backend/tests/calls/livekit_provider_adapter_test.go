//go:build hubtest

package calls_test

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

	claims := parseLiveKitCredential(t, cfg, credential.Token)
	assert.Equal(t, cfg.APIKey, claims["iss"])
	assert.Equal(t, userID, claims["sub"])
	assert.Equal(t, "Ada Lovelace", claims["name"])
	assertLiveKitRoomGrant(t, claims, credential.RoomName)
	assertLiveKitTwentyFourHourLifetime(t, claims)
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
		assert.Equal(t, credential.Identity, parseLiveKitCredential(t, cfg, credential.Join.Token)["sub"])
	}

	assert.NotEqual(t, first.Identity, second.Identity, "each guest credential must have fresh identity entropy")
	assert.Equal(t, "Guest One", parseLiveKitCredential(t, cfg, first.Join.Token)["name"])
	assert.Equal(t, "Guest Two", parseLiveKitCredential(t, cfg, second.Join.Token)["name"])
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

func parseLiveKitCredential(t *testing.T, cfg config.LiveKitConfig, token string) jwt.MapClaims {
	t.Helper()

	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		require.IsType(t, jwt.SigningMethodHS256, token.Method)
		return []byte(cfg.APISecret), nil
	})
	require.NoError(t, err)
	require.True(t, parsed.Valid)

	claims, ok := parsed.Claims.(jwt.MapClaims)
	require.True(t, ok)
	return claims
}

func assertLiveKitRoomGrant(t *testing.T, claims jwt.MapClaims, roomName string) {
	t.Helper()

	grant, ok := claims["video"].(map[string]interface{})
	require.True(t, ok, "credential must contain a LiveKit video grant")
	roomJoin, ok := grant["room_join"].(bool)
	require.True(t, ok)
	assert.True(t, roomJoin)
	assert.Equal(t, roomName, grant["room"])
}

func assertLiveKitTwentyFourHourLifetime(t *testing.T, claims jwt.MapClaims) {
	t.Helper()

	issuedAt, ok := claims["iat"].(float64)
	require.True(t, ok)
	expiresAt, ok := claims["exp"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64((24*time.Hour).Seconds()), expiresAt-issuedAt)
}