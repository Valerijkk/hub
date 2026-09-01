//go:build hubtest

package calls_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	entcall "github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/internal/calls"
)

var _ calls.DurableCallJoinCredentialIssuer = (*calls.JitsiConferenceProvider)(nil)

type durableCallJoinCredentialIssuerExact interface {
	IssueDurableCallJoinCredential(
		*ent.Call,
		calls.ProviderParticipantIdentity,
		string,
	) (calls.JoinCredential, error)
}

var _ durableCallJoinCredentialIssuerExact = (*calls.JitsiConferenceProvider)(nil)

func TestJitsiDurableJoinCredentialSignsOnlyProviderParticipantIdentity(t *testing.T) {
	cfg := validJitsiIssuerConfig()
	provider, err := calls.NewJitsiConferenceProvider(cfg)
	require.NoError(t, err)

	call := activeJitsiCall()
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	epoch := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	connectionKey, err := calls.NewCallParticipantConnectionKey(call.ID, userID, epoch)
	require.NoError(t, err)
	identity, err := calls.NewProviderParticipantIdentity(connectionKey)
	require.NoError(t, err)

	credential, err := provider.IssueDurableCallJoinCredential(call, identity, "Ada Lovelace")
	require.NoError(t, err)
	require.Equal(t, "jitsi", credential.Provider)
	require.Equal(t, call.RoomName, credential.RoomName)
	require.Equal(t, cfg.PublicURL, credential.ServiceURL)
	require.NotEmpty(t, credential.Token)
	require.False(t, credential.ExpiresAt.IsZero())

	claims, err := inspectJitsiIssuerMapClaims(credential.Token, cfg.AppSecret, time.Now().UTC(), cfg)
	require.NoError(t, err)
	require.Len(t, claims, 7)
	for _, key := range []string{"aud", "exp", "iat", "iss", "sub", "context", "room"} {
		_, exists := claims[key]
		require.True(t, exists, "missing exact durable credential claim %q", key)
	}
	require.Equal(t, call.RoomName, claims["room"])
	contextClaims, ok := claims["context"].(map[string]interface{})
	require.True(t, ok)
	require.Len(t, contextClaims, 1)
	userClaims, ok := contextClaims["user"].(map[string]interface{})
	require.True(t, ok)
	require.Len(t, userClaims, 2)
	for _, key := range []string{"id", "name"} {
		_, exists := userClaims[key]
		require.True(t, exists, "missing exact durable credential user claim %q", key)
	}
	require.Equal(t, identity.Canonical(), userClaims["id"])
	require.Equal(t, "Ada Lovelace", userClaims["name"])
	require.NotEqual(t, connectionKey.Canonical(), userClaims["id"])
	require.NotEqual(t, userID.String(), userClaims["id"])

	issuedAt, ok := claims["iat"].(float64)
	require.True(t, ok)
	expiresAt, ok := claims["exp"].(float64)
	require.True(t, ok)
	require.Equal(t, float64((5 * time.Minute) / time.Second), expiresAt-issuedAt)
	require.Equal(t, time.Unix(int64(expiresAt), 0).UTC(), credential.ExpiresAt)
}

func TestJitsiDurableJoinCredentialRejectsInvalidTrustedFactsWithoutDisclosure(t *testing.T) {
	cfg := validJitsiIssuerConfig()
	provider, err := calls.NewJitsiConferenceProvider(cfg)
	require.NoError(t, err)

	validCall := activeJitsiCall()
	validUserID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	validEpoch := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	validKey, err := calls.NewCallParticipantConnectionKey(validCall.ID, validUserID, validEpoch)
	require.NoError(t, err)
	validIdentity, err := calls.NewProviderParticipantIdentity(validKey)
	require.NoError(t, err)
	mismatchedKey, err := calls.NewCallParticipantConnectionKey(
		uuid.MustParse("99999999-9999-4999-8999-999999999999"),
		validUserID,
		validEpoch,
	)
	require.NoError(t, err)
	mismatchedIdentity, err := calls.NewProviderParticipantIdentity(mismatchedKey)
	require.NoError(t, err)

	tests := []struct {
		name     string
		call     *ent.Call
		identity calls.ProviderParticipantIdentity
		userName string
	}{
		{name: "nil call", call: nil, identity: validIdentity, userName: "Ada Lovelace"},
		{name: "zero call ID", call: &ent.Call{RoomName: "durable-room", Status: entcall.StatusActive}, identity: validIdentity, userName: "Ada Lovelace"},
		{name: "inactive call", call: &ent.Call{ID: validCall.ID, RoomName: "durable-room", Status: entcall.Status("")}, identity: validIdentity, userName: "Ada Lovelace"},
		{name: "ended call", call: &ent.Call{ID: validCall.ID, RoomName: "durable-room", Status: entcall.StatusEnded}, identity: validIdentity, userName: "Ada Lovelace"},
		{name: "empty room", call: &ent.Call{ID: validCall.ID, RoomName: "", Status: entcall.StatusActive}, identity: validIdentity, userName: "Ada Lovelace"},
		{name: "whitespace room", call: &ent.Call{ID: validCall.ID, RoomName: " \t\r\n ", Status: entcall.StatusActive}, identity: validIdentity, userName: "Ada Lovelace"},
		{name: "padded room", call: &ent.Call{ID: validCall.ID, RoomName: " durable-room ", Status: entcall.StatusActive}, identity: validIdentity, userName: "Ada Lovelace"},
		{name: "zero identity", call: validCall, identity: calls.ProviderParticipantIdentity{}, userName: "Ada Lovelace"},
		{name: "identity call mismatch", call: validCall, identity: mismatchedIdentity, userName: "Ada Lovelace"},
		{name: "empty server name", call: validCall, identity: validIdentity, userName: ""},
		{name: "whitespace server name", call: validCall, identity: validIdentity, userName: "   "},
		{name: "padded server name", call: validCall, identity: validIdentity, userName: " Ada Lovelace "},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			credential, err := provider.IssueDurableCallJoinCredential(testCase.call, testCase.identity, testCase.userName)
			require.Same(t, calls.ErrCallJoinCredentialInvalid, err)
			require.Equal(t, calls.JoinCredential{}, credential)

			roomName := ""
			if testCase.call != nil {
				roomName = testCase.call.RoomName
			}
				requireCredentialSafeError(
					t,
					err,
					cfg.AppSecret,
					cfg.PublicURL,
					cfg.AppID,
					cfg.XMPPDomain,
					roomName,
				testCase.userName,
				testCase.identity.Canonical(),
				testCase.identity.ConnectionKey().Canonical(),
			)
		})
	}
}

func TestJitsiDurableJoinCredentialPreservesLegacyUUIDIssuanceAndPaddedName(t *testing.T) {
	cfg := validJitsiIssuerConfig()
	provider, err := calls.NewJitsiConferenceProvider(cfg)
	require.NoError(t, err)

	call := activeJitsiCall()
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	credential, err := provider.IssueCallJoinCredential(call, userID, " Ada Lovelace ")
	require.NoError(t, err)

	claims, err := inspectJitsiIssuerMapClaims(credential.Token, cfg.AppSecret, time.Now().UTC(), cfg)
	require.NoError(t, err)
	contextClaims, ok := claims["context"].(map[string]interface{})
	require.True(t, ok)
	userClaims, ok := contextClaims["user"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, userID.String(), userClaims["id"])
	require.Equal(t, " Ada Lovelace ", userClaims["name"])
	require.Equal(t, call.RoomName, credential.RoomName)
	require.Equal(t, cfg.PublicURL, credential.ServiceURL)
	require.Equal(t, "jitsi", credential.Provider)
}

func TestJitsiDurableCredentialFacetDoesNotBroadenProviderSurface(t *testing.T) {
	cfg := validJitsiIssuerConfig()
	provider, err := calls.NewJitsiConferenceProvider(cfg)
	require.NoError(t, err)

	require.Equal(t, calls.ConferenceProviderID(calls.ProviderJitsi), provider.ProviderID())
	require.Equal(
		t,
		calls.ConferenceProviderCapabilities(calls.CapabilityIssueCallJoinCredential),
		provider.Capabilities(),
	)
	_, implementsLegacy := any(provider).(calls.ConferenceProvider)
	require.False(t, implementsLegacy)
	_, implementsGuest := any(provider).(calls.ConferenceGuestCredentialIssuer)
	require.False(t, implementsGuest)
	_, implementsProvisioning := any(provider).(calls.ConferenceRoomProvisioner)
	require.False(t, implementsProvisioning)
	_, implementsDeletion := any(provider).(calls.ConferenceRoomDeleter)
	require.False(t, implementsDeletion)
	_, implementsMetadata := any(provider).(calls.ConferenceParticipantMetadataUpdater)
	require.False(t, implementsMetadata)

	_, liveKitHasDurableFacet := reflect.TypeOf((*calls.LiveKitConferenceProvider)(nil)).MethodByName("IssueDurableCallJoinCredential")
	require.False(t, liveKitHasDurableFacet)
}
