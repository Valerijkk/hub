//go:build hubtest

package calls_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	entcall "github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/internal/calls"
)

type callJoinCredentialFacadeContextKey struct{}

// callJoinCredentialFacadeContextTrackingSource preserves the existing
// projection fake's operations while recording the context passed through the
// service facade.
type callJoinCredentialFacadeContextTrackingSource struct {
	fake     *callJoinCredentialProjectionSourceFake
	contexts []context.Context
}

func (s *callJoinCredentialFacadeContextTrackingSource) LoadActiveCall(
	ctx context.Context,
	callID uuid.UUID,
) (*ent.Call, error) {
	s.contexts = append(s.contexts, ctx)
	return s.fake.LoadActiveCall(ctx, callID)
}

func (s *callJoinCredentialFacadeContextTrackingSource) IsChannelMember(
	ctx context.Context,
	channelID uuid.UUID,
	authenticatedUserID uuid.UUID,
) (bool, error) {
	s.contexts = append(s.contexts, ctx)
	return s.fake.IsChannelMember(ctx, channelID, authenticatedUserID)
}

func (s *callJoinCredentialFacadeContextTrackingSource) GetServerKnownCallUserName(
	ctx context.Context,
	authenticatedUserID uuid.UUID,
) (string, error) {
	s.contexts = append(s.contexts, ctx)
	return s.fake.GetServerKnownCallUserName(ctx, authenticatedUserID)
}

var _ calls.CallJoinCredentialProjectionSource = (*callJoinCredentialFacadeContextTrackingSource)(nil)

func TestNewCallServiceWithProviderAndCredentialProjectorRejectsNilProjector(t *testing.T) {
	service, err := calls.NewCallServiceWithProviderAndCredentialProjector(nil, nil, nil, nil, nil)

	require.ErrorIs(t, err, calls.ErrCallJoinCredentialInvalid)
	assert.Nil(t, service)
}

func TestCallServiceIssueJoinCredentialForAuthenticatedUserRejectsNilAndLegacyServiceWithoutProjectorWork(t *testing.T) {
	ctx := context.WithValue(context.Background(), callJoinCredentialFacadeContextKey{}, "facade-context")
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	authenticatedUserID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	source := &callJoinCredentialProjectionSourceFake{}
	issuer := &callJoinCredentialProjectionIssuerFake{}
	projector, err := calls.NewCallJoinCredentialProjector(source, issuer)
	require.NoError(t, err)
	var nilService *calls.CallService

	credential, err := nilService.IssueJoinCredentialForAuthenticatedUser(ctx, callID, authenticatedUserID)

	require.ErrorIs(t, err, calls.ErrCallJoinCredentialInvalid)
	assert.Equal(t, calls.JoinCredential{}, credential)
	assert.Empty(t, source.operations)
	assert.Empty(t, issuer.operations)

	legacyService := calls.NewCallServiceWithProvider(nil, nil, nil, nil)
	credential, err = legacyService.IssueJoinCredentialForAuthenticatedUser(ctx, callID, authenticatedUserID)

	require.ErrorIs(t, err, calls.ErrCallJoinCredentialInvalid)
	assert.Equal(t, calls.JoinCredential{}, credential)
	assert.NotNil(t, projector)
	assert.Empty(t, source.operations)
	assert.Empty(t, issuer.operations)
}

func TestCallServiceIssueJoinCredentialForAuthenticatedUserProjectsCompleteCredentialForChannelMember(t *testing.T) {
	ctx := context.WithValue(context.Background(), callJoinCredentialFacadeContextKey{}, "facade-context")
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	authenticatedUserID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	creatorID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	channelID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	activeCall := &ent.Call{
		ID:        callID,
		RoomName:  "server-owned-channel-room",
		Status:    entcall.StatusActive,
		ChannelID: &channelID,
		CreatedBy: creatorID,
	}
	want := calls.JoinCredential{
		Provider:   "jitsi",
		RoomName:   activeCall.RoomName,
		ServiceURL: "https://meet.hub.test",
		Token:      "provider-issued-token",
		ExpiresAt:  time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC),
	}
	sourceFake := &callJoinCredentialProjectionSourceFake{
		activeCall: activeCall,
		member:     true,
		userName:   "Persisted Display Name",
	}
	source := &callJoinCredentialFacadeContextTrackingSource{fake: sourceFake}
	issuer := &callJoinCredentialProjectionIssuerFake{credential: want}
	projector, err := calls.NewCallJoinCredentialProjector(source, issuer)
	require.NoError(t, err)
	provider := &providerInjectionFake{}
	service, err := calls.NewCallServiceWithProviderAndCredentialProjector(provider, nil, nil, nil, projector)
	require.NoError(t, err)

	credential, err := service.IssueJoinCredentialForAuthenticatedUser(ctx, callID, authenticatedUserID)

	require.NoError(t, err)
	assert.Equal(t, want, credential)
	require.Len(t, source.contexts, 3)
	for _, observedContext := range source.contexts {
		assert.Same(t, ctx, observedContext)
	}
	assert.Equal(t, []callJoinCredentialProjectionOperation{
		{name: "load_active_call", callID: callID},
		{name: "is_channel_member", channelID: channelID, authenticatedUserID: authenticatedUserID},
		{name: "get_server_known_call_user_name", authenticatedUserID: authenticatedUserID},
	}, sourceFake.operations)
	require.Len(t, issuer.operations, 1)
	assert.Same(t, activeCall, issuer.operations[0].call)
	assert.Equal(t, authenticatedUserID, issuer.operations[0].actorID)
	assert.Equal(t, "Persisted Display Name", issuer.operations[0].userName)
	assert.Empty(t, provider.operations)
}

func TestCallServiceIssueJoinCredentialForAuthenticatedUserReturnsZeroCredentialAndExactIssuerError(t *testing.T) {
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	authenticatedUserID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	channelID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	issuerErr := errors.New("credential issuer unavailable")
	source := &callJoinCredentialProjectionSourceFake{
		activeCall: &ent.Call{
			ID:        callID,
			RoomName:  "server-owned-channel-room",
			Status:    entcall.StatusActive,
			ChannelID: &channelID,
			CreatedBy: uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		},
		member:   true,
		userName: "Persisted Display Name",
	}
	issuer := &callJoinCredentialProjectionIssuerFake{
		credential: calls.JoinCredential{
			Provider:   "jitsi",
			RoomName:   "must-not-leak",
			ServiceURL: "https://meet.hub.test",
			Token:      "partial-provider-token",
			ExpiresAt:  time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC),
		},
		err: issuerErr,
	}
	projector, err := calls.NewCallJoinCredentialProjector(source, issuer)
	require.NoError(t, err)
	service, err := calls.NewCallServiceWithProviderAndCredentialProjector(nil, nil, nil, nil, projector)
	require.NoError(t, err)

	credential, err := service.IssueJoinCredentialForAuthenticatedUser(context.Background(), callID, authenticatedUserID)

	require.ErrorIs(t, err, issuerErr)
	assert.True(t, err == issuerErr)
	assert.Same(t, issuerErr, err)
	assert.Equal(t, calls.JoinCredential{}, credential)
	require.Len(t, issuer.operations, 1)
}
