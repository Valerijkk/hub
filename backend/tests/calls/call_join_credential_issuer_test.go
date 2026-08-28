//go:build hubtest

package calls_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	entcall "github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/internal/calls"
)

// Call and identity facts here are trusted server-loaded inputs. A future service
// boundary owns loading, membership, and authorization before it invokes an issuer.
func TestCallJoinCredentialIssuersSatisfyCallAwareBoundary(t *testing.T) {
	liveKitIssuer := calls.NewLiveKitConferenceProvider(liveKitProviderConfig())
	jitsiIssuer, err := calls.ExportNewJitsiCredentialIssuerForTest(validJitsiIssuerConfig(), func() time.Time {
		return fixedJitsiIssueTime
	})
	require.NoError(t, err)

	var liveKitContract calls.CallJoinCredentialIssuer = liveKitIssuer
	var jitsiContract calls.CallJoinCredentialIssuer = jitsiIssuer

	require.NotNil(t, liveKitContract)
	require.NotNil(t, jitsiContract)
}

func TestLiveKitCallJoinCredentialIssuerUsesTrustedCallAndIdentity(t *testing.T) {
	cfg := liveKitProviderConfig()
	var issuer calls.CallJoinCredentialIssuer = calls.NewLiveKitConferenceProvider(cfg)
	callEntity := callJoinCredentialIssuerActiveCall()
	callEntity.RoomName = "call-aware-livekit-room"
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	serverKnownName := "  Grace Hopper  "

	credential, err := issuer.IssueCallJoinCredential(callEntity, userID, serverKnownName)

	require.NoError(t, err)
	assert.Equal(t, calls.ProviderLiveKit, credential.Provider)
	assert.Equal(t, callEntity.RoomName, credential.RoomName)
	assert.Equal(t, cfg.URL, credential.ServiceURL)
	assertLiveKitCredential(t, cfg, credential.Token, userID.String(), serverKnownName, callEntity.RoomName, credential.ExpiresAt)
}

func TestJitsiCallJoinCredentialIssuerUsesTrustedCallAndIdentity(t *testing.T) {
	cfg := validJitsiIssuerConfig()
	clockCalls := 0
	jitsiIssuer, err := calls.ExportNewJitsiCredentialIssuerForTest(cfg, func() time.Time {
		clockCalls++
		return fixedJitsiIssueTime
	})
	require.NoError(t, err)

	var issuer calls.CallJoinCredentialIssuer = jitsiIssuer
	callEntity := callJoinCredentialIssuerActiveCall()
	callEntity.RoomName = "call-aware-jitsi-room"
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	serverKnownName := "  Dorothy Vaughan  "

	credential, err := issuer.IssueCallJoinCredential(callEntity, userID, serverKnownName)

	require.NoError(t, err)
	assert.Equal(t, "jitsi", credential.Provider)
	assert.Equal(t, callEntity.RoomName, credential.RoomName)
	assert.Equal(t, cfg.PublicURL, credential.ServiceURL)
	assert.Equal(t, 1, clockCalls)

	parsedToken, claims, err := inspectJitsiIssuerToken(
		credential.Token,
		cfg.AppSecret,
		fixedJitsiIssueTime.Add(time.Minute),
		cfg,
		callEntity.RoomName,
		userID,
		serverKnownName,
	)
	require.NoError(t, err)
	require.True(t, parsedToken.Valid)
	require.NotNil(t, claims.ExpiresAt)
	assert.True(t, credential.ExpiresAt.Equal(claims.ExpiresAt.Time))
	assert.True(t, credential.ExpiresAt.Equal(fixedJitsiIssueTime.Add(5*time.Minute)))
}

func TestCallJoinCredentialIssuersRejectInvalidTrustedInputs(t *testing.T) {
	validUserID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	jitsiIssuer, err := calls.ExportNewJitsiCredentialIssuerForTest(validJitsiIssuerConfig(), func() time.Time {
		return fixedJitsiIssueTime
	})
	require.NoError(t, err)

	issuers := []struct {
		name   string
		issuer calls.CallJoinCredentialIssuer
	}{
		{name: "LiveKit", issuer: calls.NewLiveKitConferenceProvider(liveKitProviderConfig())},
		{name: "Jitsi", issuer: jitsiIssuer},
	}
	tests := []struct {
		name     string
		newCall  func() *ent.Call
		userID   uuid.UUID
		userName string
	}{
		{name: "nil call", newCall: func() *ent.Call { return nil }, userID: validUserID, userName: "Ada Lovelace"},
		{
			name: "zero call ID",
			newCall: func() *ent.Call {
				callEntity := callJoinCredentialIssuerActiveCall()
				callEntity.ID = uuid.Nil
				return callEntity
			},
			userID: validUserID, userName: "Ada Lovelace",
		},
		{
			name: "inactive call",
			newCall: func() *ent.Call {
				callEntity := callJoinCredentialIssuerActiveCall()
				callEntity.Status = entcall.Status("")
				return callEntity
			},
			userID: validUserID, userName: "Ada Lovelace",
		},
		{
			name: "ended call",
			newCall: func() *ent.Call {
				callEntity := callJoinCredentialIssuerActiveCall()
				callEntity.Status = entcall.StatusEnded
				return callEntity
			},
			userID: validUserID, userName: "Ada Lovelace",
		},
		{
			name: "empty room name",
			newCall: func() *ent.Call {
				callEntity := callJoinCredentialIssuerActiveCall()
				callEntity.RoomName = ""
				return callEntity
			},
			userID: validUserID, userName: "Ada Lovelace",
		},
		{
			name: "whitespace room name",
			newCall: func() *ent.Call {
				callEntity := callJoinCredentialIssuerActiveCall()
				callEntity.RoomName = " \t\r\n "
				return callEntity
			},
			userID: validUserID, userName: "Ada Lovelace",
		},
		{
			name: "padded room name",
			newCall: func() *ent.Call {
				callEntity := callJoinCredentialIssuerActiveCall()
				callEntity.RoomName = " canonical-room "
				return callEntity
			},
			userID: validUserID, userName: "Ada Lovelace",
		},
		{name: "zero authenticated user", newCall: callJoinCredentialIssuerActiveCall, userID: uuid.Nil, userName: "Ada Lovelace"},
		{name: "empty server-known name", newCall: callJoinCredentialIssuerActiveCall, userID: validUserID, userName: ""},
		{name: "whitespace server-known name", newCall: callJoinCredentialIssuerActiveCall, userID: validUserID, userName: " \t\r\n "},
	}

	for _, issuerCase := range issuers {
		issuerCase := issuerCase
		t.Run(issuerCase.name, func(t *testing.T) {
			for _, testCase := range tests {
				testCase := testCase
				t.Run(testCase.name, func(t *testing.T) {
					credential, err := issuerCase.issuer.IssueCallJoinCredential(testCase.newCall(), testCase.userID, testCase.userName)

					assert.Equal(t, calls.JoinCredential{}, credential)
					require.ErrorIs(t, err, calls.ErrCallJoinCredentialInvalid)
				})
			}
		})
	}
}

func callJoinCredentialIssuerActiveCall() *ent.Call {
	return &ent.Call{
		ID:       uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		RoomName: "canonical-call-room",
		Status:   entcall.StatusActive,
	}
}
