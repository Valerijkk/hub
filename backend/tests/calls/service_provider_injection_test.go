//go:build hubtest

package calls_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/calls"
)

// ExportNewCallServiceWithProviderForTest is the only new CallService test
// seam frozen by this packet. It must remain tagged-test-only and
// constructor-shaped:
//
//	func ExportNewCallServiceWithProviderForTest(
//		repo *EntRepository,
//		provider ConferenceProvider,
//		pub *CallEventPublisher,
//		chat *CallChatService,
//	) *CallService
//
// GuestJoin is deliberately not exercised here. Its service flow requires the
// concrete EntRepository lookup, and adding a repository abstraction or a
// guest-specific bypass solely for this unit would widen the production
// boundary. Guest credential semantics are covered by
// livekit_provider_adapter_test.go; Task 3 static review must verify that the
// existing GuestJoin flow delegates to ConferenceProvider.

type providerInjectionOperation struct {
	name            string
	roomName        string
	identity        string
	userName        string
	metadataJSON    string
	maxParticipants uint32
	createdBy       uuid.UUID
	inviteCode      string
	title           string
}

type providerInjectionFake struct {
	operations      []providerInjectionOperation
	joinCredential  calls.JoinCredential
	guestCredential calls.GuestJoinCredential
	provisionErr    error
	deleteErr       error
	joinErr         error
	guestErr        error
	participantErr  error
}

func (p *providerInjectionFake) Provision(_ context.Context, roomName string, maxParticipants uint32) error {
	p.operations = append(p.operations, providerInjectionOperation{
		name:            "provision",
		roomName:        roomName,
		maxParticipants: maxParticipants,
	})
	return p.provisionErr
}

func (p *providerInjectionFake) Delete(_ context.Context, roomName string) error {
	p.operations = append(p.operations, providerInjectionOperation{
		name:     "delete",
		roomName: roomName,
	})
	return p.deleteErr
}

func (p *providerInjectionFake) IssueJoinCredential(roomName, identity, userName string) (calls.JoinCredential, error) {
	p.operations = append(p.operations, providerInjectionOperation{
		name:     "issue_join_credential",
		roomName: roomName,
		identity: identity,
		userName: userName,
	})
	return p.joinCredential, p.joinErr
}

func (p *providerInjectionFake) IssueGuestCredential(roomName, guestName string) (calls.GuestJoinCredential, error) {
	p.operations = append(p.operations, providerInjectionOperation{
		name:     "issue_guest_credential",
		roomName: roomName,
		userName: guestName,
	})
	return p.guestCredential, p.guestErr
}

func (p *providerInjectionFake) UpdateParticipantState(
	_ context.Context,
	roomName, identity, metadataJSON string,
) error {
	p.operations = append(p.operations, providerInjectionOperation{
		name:         "update_participant_state",
		roomName:     roomName,
		identity:     identity,
		metadataJSON: metadataJSON,
	})
	return p.participantErr
}

func TestCallServiceGenerateTokenUsesInjectedProviderJoinCredential(t *testing.T) {
	provider := &providerInjectionFake{
		joinCredential: calls.JoinCredential{
			Provider:   calls.ProviderLiveKit,
			RoomName:   "call-user-credential",
			ServiceURL: "https://calls.hub.test",
			Token:      "provider-issued-user-token",
		},
	}
	service := calls.ExportNewCallServiceWithProviderForTest(nil, provider, nil, nil)
	userID := uuid.New()

	token, err := service.GenerateToken("call-user-credential", userID.String(), "Ada Lovelace")

	require.NoError(t, err)
	assert.Equal(t, "provider-issued-user-token", token, "CallService must preserve the legacy token-string boundary")
	require.Equal(t, []providerInjectionOperation{
		{
			name:     "issue_join_credential",
			roomName: "call-user-credential",
			identity: userID.String(),
			userName: "Ada Lovelace",
		},
	}, provider.operations)
}

func TestCallServiceGenerateJoinCredentialDelegatesExactTypedCredential(t *testing.T) {
	userID := uuid.New()
	want := calls.JoinCredential{
		Provider:   "jitsi",
		RoomName:   "hub-call-room-42",
		ServiceURL: "https://meet.hub.test",
		Token:      "signed-jitsi-jwt",
	}
	provider := &providerInjectionFake{joinCredential: want}
	service := calls.ExportNewCallServiceWithProviderForTest(nil, provider, nil, nil)

	credential, err := service.GenerateJoinCredential(want.RoomName, userID.String(), "Grace Hopper")

	require.NoError(t, err)
	assert.Equal(t, want, credential, "CallService must preserve every provider-issued join credential field")
	require.Equal(t, []providerInjectionOperation{
		{
			name:     "issue_join_credential",
			roomName: want.RoomName,
			identity: userID.String(),
			userName: "Grace Hopper",
		},
	}, provider.operations)
}

func TestCallServiceGenerateJoinCredentialReturnsZeroValueAndProviderError(t *testing.T) {
	providerErr := errors.New("provider unavailable")
	provider := &providerInjectionFake{
		joinCredential: calls.JoinCredential{
			Provider:   "jitsi",
			RoomName:   "must-not-leak",
			ServiceURL: "https://meet.hub.test",
			Token:      "must-not-leak",
		},
		joinErr: providerErr,
	}
	service := calls.ExportNewCallServiceWithProviderForTest(nil, provider, nil, nil)

	credential, err := service.GenerateJoinCredential("hub-call-room-error", "user-42", "Linus Torvalds")

	require.ErrorIs(t, err, providerErr)
	assert.Equal(t, calls.JoinCredential{}, credential, "a failed provider call must not expose a partial credential")
	require.Equal(t, []providerInjectionOperation{
		{
			name:     "issue_join_credential",
			roomName: "hub-call-room-error",
			identity: "user-42",
			userName: "Linus Torvalds",
		},
	}, provider.operations)
}

func TestCallServiceGenerateTokenPreservesProviderErrorIdentity(t *testing.T) {
	providerErr := errors.New("token issuer unavailable")
	provider := &providerInjectionFake{joinErr: providerErr}
	service := calls.ExportNewCallServiceWithProviderForTest(nil, provider, nil, nil)

	token, err := service.GenerateToken("hub-call-token-error", "user-7", "Margaret Hamilton")

	require.ErrorIs(t, err, providerErr)
	assert.Empty(t, token, "the legacy token boundary must remain empty when credential issuance fails")
	require.Equal(t, []providerInjectionOperation{
		{
			name:     "issue_join_credential",
			roomName: "hub-call-token-error",
			identity: "user-7",
			userName: "Margaret Hamilton",
		},
	}, provider.operations)
}
func TestCallServiceCreateInstantCallRollsBackInjectedProviderAfterPersistenceFailure(t *testing.T) {
	provider := &providerInjectionFake{}
	service := calls.ExportNewCallServiceWithProviderForTest(nil, provider, nil, nil)
	userID := uuid.New()
	calls.ExportCallServiceSetCreateInstantCall(service, func(
		_ context.Context,
		roomName string,
		createdBy uuid.UUID,
		inviteCode string,
		title string,
	) (*ent.Call, error) {
		provider.operations = append(provider.operations, providerInjectionOperation{
			name:       "persist_instant_call",
			roomName:   roomName,
			createdBy:  createdBy,
			inviteCode: inviteCode,
			title:      title,
		})
		return nil, errors.New("persist instant call")
	})

	response, err := service.CreateInstantCall(context.Background(), userID, "Ada Lovelace", "Daily stand-up")

	require.Error(t, err)
	assert.Nil(t, response)
	require.Len(t, provider.operations, 3)
	assert.Equal(t, "provision", provider.operations[0].name)
	assert.Equal(t, uint32(calls.ExportDefaultMaxCallParticipants()), provider.operations[0].maxParticipants)
	assert.True(t, strings.HasPrefix(provider.operations[0].roomName, "call-"))
	assert.Equal(t, "persist_instant_call", provider.operations[1].name)
	assert.Equal(t, provider.operations[0].roomName, provider.operations[1].roomName)
	assert.Equal(t, userID, provider.operations[1].createdBy)
	assert.Equal(t, "Daily stand-up", provider.operations[1].title)
	assert.NotEmpty(t, provider.operations[1].inviteCode, "persistence must receive a generated invite code")
	assert.Equal(t, "delete", provider.operations[2].name)
	assert.Equal(t, provider.operations[0].roomName, provider.operations[2].roomName)
}

func TestCallServiceSetRaiseHandUsesInjectedProviderParticipantState(t *testing.T) {
	provider := &providerInjectionFake{}
	service := calls.ExportNewCallServiceWithProviderForTest(nil, provider, nil, nil)

	err := service.SetRaiseHand(context.Background(), "call-hand-state", "user-42", true)

	require.NoError(t, err)
	require.Len(t, provider.operations, 1, "participant state must be delegated exactly once")
	assert.Equal(t, "update_participant_state", provider.operations[0].name)
	assert.Equal(t, "call-hand-state", provider.operations[0].roomName)
	assert.Equal(t, "user-42", provider.operations[0].identity)

	var metadata struct {
		HandRaised bool   `json:"hand_raised"`
		RaisedAt   string `json:"raised_at"`
	}
	require.NoError(t, json.Unmarshal([]byte(provider.operations[0].metadataJSON), &metadata))
	assert.True(t, metadata.HandRaised)
	_, parseErr := time.Parse(time.RFC3339, metadata.RaisedAt)
	require.NoError(t, parseErr)
}

func TestCallServiceProviderInjectionFakeSatisfiesFrozenProviderContract(t *testing.T) {
	var provider calls.ConferenceProvider = &providerInjectionFake{}
	require.NotNil(t, provider)
}
