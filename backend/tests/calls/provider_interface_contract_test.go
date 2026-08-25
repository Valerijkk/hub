//go:build hubtest

package calls_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/calls"
)

// providerContractProbe must keep satisfying ConferenceProvider. A provider
// boundary that adds a vendor SDK type or drops one of these lifecycle,
// credential, or participant-state operations breaks this contract.
type providerContractProbe struct{}

func (providerContractProbe) Provision(context.Context, string, uint32) error {
	return nil
}

func (providerContractProbe) Delete(context.Context, string) error {
	return nil
}

func (providerContractProbe) IssueJoinCredential(string, string, string) (calls.JoinCredential, error) {
	return calls.JoinCredential{}, nil
}

func (providerContractProbe) IssueGuestCredential(string, string) (calls.GuestJoinCredential, error) {
	return calls.GuestJoinCredential{}, nil
}

func (providerContractProbe) UpdateParticipantState(context.Context, string, string, string) error {
	return nil
}

var _ calls.ConferenceProvider = providerContractProbe{}

func TestConferenceProviderHasOnlyTheProviderNeutralCallOperations(t *testing.T) {
	providerType := reflect.TypeOf((*calls.ConferenceProvider)(nil)).Elem()
	methodNames := make([]string, providerType.NumMethod())
	for index := range providerType.NumMethod() {
		methodNames[index] = providerType.Method(index).Name
	}

	assert.Equal(t, []string{
		"Delete",
		"IssueGuestCredential",
		"IssueJoinCredential",
		"Provision",
		"UpdateParticipantState",
	}, methodNames)
}

func TestJoinCredentialsKeepProviderRoomServiceURLAndTokenTyped(t *testing.T) {
	userID := uuid.NewString()
	join := calls.JoinCredential{
		Provider:   calls.ProviderLiveKit,
		RoomName:   "call-20260825",
		ServiceURL: "https://calls.hub.test",
		Token:      "signed-user-token",
	}
	guest := calls.GuestJoinCredential{
		Identity: "guest-" + userID,
		Join:     join,
	}

	require.NotEmpty(t, join.Provider)
	require.NotEmpty(t, join.RoomName)
	require.NotEmpty(t, join.ServiceURL)
	require.NotEmpty(t, join.Token)
	assert.Equal(t, calls.ProviderLiveKit, join.Provider)
	assert.Equal(t, "call-20260825", join.RoomName)
	assert.Equal(t, "https://calls.hub.test", join.ServiceURL)
	assert.Equal(t, "signed-user-token", join.Token)
	assert.Equal(t, "guest-"+userID, guest.Identity)
	assert.Equal(t, join, guest.Join)
}
