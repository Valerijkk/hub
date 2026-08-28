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

// The source is deliberately the only server-owned authority the projector
// can use. It keeps room, name, call entity, and provider data out of caller
// input and out of this projection-only boundary.
type callJoinCredentialProjectionSourceFake struct {
	operations []callJoinCredentialProjectionOperation

	activeCall *ent.Call
	member     bool
	userName   string

	loadErr   error
	memberErr error
	nameErr   error
}

type callJoinCredentialProjectionOperation struct {
	name                string
	callID              uuid.UUID
	channelID           uuid.UUID
	authenticatedUserID uuid.UUID
}

func (s *callJoinCredentialProjectionSourceFake) LoadActiveCall(
	_ context.Context,
	callID uuid.UUID,
) (*ent.Call, error) {
	s.operations = append(s.operations, callJoinCredentialProjectionOperation{
		name:   "load_active_call",
		callID: callID,
	})
	return s.activeCall, s.loadErr
}

func (s *callJoinCredentialProjectionSourceFake) IsChannelMember(
	_ context.Context,
	channelID uuid.UUID,
	authenticatedUserID uuid.UUID,
) (bool, error) {
	s.operations = append(s.operations, callJoinCredentialProjectionOperation{
		name:                "is_channel_member",
		channelID:           channelID,
		authenticatedUserID: authenticatedUserID,
	})
	return s.member, s.memberErr
}

func (s *callJoinCredentialProjectionSourceFake) GetServerKnownCallUserName(
	_ context.Context,
	actorID uuid.UUID,
) (string, error) {
	s.operations = append(s.operations, callJoinCredentialProjectionOperation{
		name:                "get_server_known_call_user_name",
		authenticatedUserID: actorID,
	})
	return s.userName, s.nameErr
}

var _ calls.CallJoinCredentialProjectionSource = (*callJoinCredentialProjectionSourceFake)(nil)

type callJoinCredentialProjectionIssuerFake struct {
	operations []callJoinCredentialProjectionIssuerOperation
	credential calls.JoinCredential
	err        error
}

type callJoinCredentialProjectionIssuerOperation struct {
	call     *ent.Call
	actorID  uuid.UUID
	userName string
}

func (i *callJoinCredentialProjectionIssuerFake) IssueCallJoinCredential(
	callEntity *ent.Call,
	actorID uuid.UUID,
	userName string,
) (calls.JoinCredential, error) {
	i.operations = append(i.operations, callJoinCredentialProjectionIssuerOperation{
		call:     callEntity,
		actorID:  actorID,
		userName: userName,
	})
	return i.credential, i.err
}

var _ calls.CallJoinCredentialIssuer = (*callJoinCredentialProjectionIssuerFake)(nil)

func TestCallJoinCredentialProjectorRejectsNilDependencies(t *testing.T) {
	var typedNilSource *callJoinCredentialProjectionSourceFake
	var typedNilIssuer *callJoinCredentialProjectionIssuerFake

	for _, testCase := range []struct {
		name   string
		source calls.CallJoinCredentialProjectionSource
		issuer calls.CallJoinCredentialIssuer
	}{
		{
			name:   "nil source",
			issuer: &callJoinCredentialProjectionIssuerFake{},
		},
		{
			name:   "nil issuer",
			source: &callJoinCredentialProjectionSourceFake{},
		},
		{
			name:   "typed-nil source",
			source: typedNilSource,
			issuer: &callJoinCredentialProjectionIssuerFake{},
		},
		{
			name:   "typed-nil issuer",
			source: &callJoinCredentialProjectionSourceFake{},
			issuer: typedNilIssuer,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			projector, err := calls.NewCallJoinCredentialProjector(testCase.source, testCase.issuer)

			require.Error(t, err)
			assert.Nil(t, projector)
		})
	}
}

func TestCallJoinCredentialProjectorRejectsNilOrZeroValueReceiverBeforeDependencies(t *testing.T) {
	ctx := context.Background()
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	actorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	var nilProjector *calls.CallJoinCredentialProjector

	for _, testCase := range []struct {
		name      string
		projector *calls.CallJoinCredentialProjector
	}{
		{
			name:      "nil receiver",
			projector: nilProjector,
		},
		{
			name:      "zero-value receiver",
			projector: &calls.CallJoinCredentialProjector{},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var credential calls.JoinCredential
			var err error

			require.NotPanics(t, func() {
				credential, err = testCase.projector.IssueForAuthenticatedUser(ctx, callID, actorID)
			})

			assert.Equal(t, calls.JoinCredential{}, credential)
			assert.Equal(t, calls.ErrCallJoinCredentialInvalid, err)
		})
	}
}

func TestCallJoinCredentialProjectorRejectsMalformedLoadedCallBeforeAuthorization(t *testing.T) {
	ctx := context.Background()
	requestedCallID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	actorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	channelID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	creatorID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	invalidChannelID := uuid.Nil

	for _, testCase := range []struct {
		name       string
		activeCall *ent.Call
	}{
		{
			name: "nil active call",
		},
		{
			name: "different active call ID",
			activeCall: &ent.Call{
				ID:        uuid.MustParse("55555555-5555-4555-8555-555555555555"),
				RoomName:  "wrong-server-loaded-room",
				Status:    entcall.StatusActive,
				ChannelID: &channelID,
				CreatedBy: creatorID,
			},
		},
		{
			name: "inactive call",
			activeCall: &ent.Call{
				ID:        requestedCallID,
				RoomName:  "ended-server-loaded-room",
				Status:    entcall.StatusEnded,
				ChannelID: &channelID,
				CreatedBy: creatorID,
			},
		},
		{
			name: "blank room name",
			activeCall: &ent.Call{
				ID:        requestedCallID,
				RoomName:  "",
				Status:    entcall.StatusActive,
				ChannelID: &channelID,
				CreatedBy: creatorID,
			},
		},
		{
			name: "whitespace room name",
			activeCall: &ent.Call{
				ID:        requestedCallID,
				RoomName:  "   ",
				Status:    entcall.StatusActive,
				ChannelID: &channelID,
				CreatedBy: creatorID,
			},
		},
		{
			name: "padded room name",
			activeCall: &ent.Call{
				ID:        requestedCallID,
				RoomName:  " server-owned-room ",
				Status:    entcall.StatusActive,
				ChannelID: &channelID,
				CreatedBy: creatorID,
			},
		},
		{
			name: "channel call with nil channel ID",
			activeCall: &ent.Call{
				ID:        requestedCallID,
				RoomName:  "server-owned-room",
				Status:    entcall.StatusActive,
				ChannelID: &invalidChannelID,
				CreatedBy: creatorID,
			},
		},
		{
			name: "instant call with nil creator ID",
			activeCall: &ent.Call{
				ID:       requestedCallID,
				RoomName: "server-owned-room",
				Status:   entcall.StatusActive,
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			source := &callJoinCredentialProjectionSourceFake{
				activeCall: testCase.activeCall,
				member:     true,
				userName:   "must not be loaded",
			}
			issuer := &callJoinCredentialProjectionIssuerFake{
				credential: calls.JoinCredential{Token: "must not be issued"},
			}
			projector, err := calls.NewCallJoinCredentialProjector(source, issuer)
			require.NoError(t, err)

			credential, err := projector.IssueForAuthenticatedUser(ctx, requestedCallID, actorID)

			require.ErrorIs(t, err, calls.ErrCallJoinCredentialInvalid)
			assert.Equal(t, calls.JoinCredential{}, credential)
			assert.Equal(t, []callJoinCredentialProjectionOperation{
				{name: "load_active_call", callID: requestedCallID},
			}, source.operations)
			assert.Empty(t, issuer.operations)
		})
	}
}

func TestCallJoinCredentialProjectorReloadsAuthorizesAndIssuesOnceForChannelMember(t *testing.T) {
	ctx := context.Background()
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	actorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	creatorID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	channelID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	activeCall := &ent.Call{
		ID:        callID,
		RoomName:  "server-owned-room",
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
	source := &callJoinCredentialProjectionSourceFake{
		activeCall: activeCall,
		member:     true,
		userName:   "Persisted Display Name",
	}
	issuer := &callJoinCredentialProjectionIssuerFake{credential: want}

	projector, err := calls.NewCallJoinCredentialProjector(source, issuer)
	require.NoError(t, err)

	credential, err := projector.IssueForAuthenticatedUser(ctx, callID, actorID)

	require.NoError(t, err)
	assert.Equal(t, want, credential)
	assert.Equal(t, []callJoinCredentialProjectionOperation{
		{name: "load_active_call", callID: callID},
		{name: "is_channel_member", channelID: channelID, authenticatedUserID: actorID},
		{name: "get_server_known_call_user_name", authenticatedUserID: actorID},
	}, source.operations)
	require.Len(t, issuer.operations, 1)
	assert.Same(t, activeCall, issuer.operations[0].call)
	assert.Equal(t, actorID, issuer.operations[0].actorID)
	assert.Equal(t, "Persisted Display Name", issuer.operations[0].userName)
}

func TestCallJoinCredentialProjectorAllowsOnlyInstantCallCreator(t *testing.T) {
	ctx := context.Background()
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	creatorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	activeCall := &ent.Call{
		ID:        callID,
		RoomName:  "instant-server-owned-room",
		Status:    entcall.StatusActive,
		CreatedBy: creatorID,
	}
	source := &callJoinCredentialProjectionSourceFake{
		activeCall: activeCall,
		userName:   "Instant Creator",
	}
	want := calls.JoinCredential{Provider: calls.ProviderLiveKit, RoomName: activeCall.RoomName, Token: "instant-token"}
	issuer := &callJoinCredentialProjectionIssuerFake{credential: want}

	projector, err := calls.NewCallJoinCredentialProjector(source, issuer)
	require.NoError(t, err)

	credential, err := projector.IssueForAuthenticatedUser(ctx, callID, creatorID)

	require.NoError(t, err)
	assert.Equal(t, want, credential)
	assert.Equal(t, []callJoinCredentialProjectionOperation{
		{name: "load_active_call", callID: callID},
		{name: "get_server_known_call_user_name", authenticatedUserID: creatorID},
	}, source.operations)
	require.Len(t, issuer.operations, 1)
	assert.Same(t, activeCall, issuer.operations[0].call)
	assert.Equal(t, creatorID, issuer.operations[0].actorID)
	assert.Equal(t, "Instant Creator", issuer.operations[0].userName)
}

func TestCallJoinCredentialProjectorDeniesUnauthorizedActorWithoutNameOrIssuer(t *testing.T) {
	ctx := context.Background()
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	actorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	creatorID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	channelID := uuid.MustParse("44444444-4444-4444-8444-444444444444")

	for _, testCase := range []struct {
		name           string
		activeCall     *ent.Call
		member         bool
		wantOperations []callJoinCredentialProjectionOperation
	}{
		{
			name: "channel non-member",
			activeCall: &ent.Call{
				ID:        callID,
				RoomName:  "channel-room",
				Status:    entcall.StatusActive,
				ChannelID: &channelID,
				CreatedBy: creatorID,
			},
			wantOperations: []callJoinCredentialProjectionOperation{
				{name: "load_active_call", callID: callID},
				{name: "is_channel_member", channelID: channelID, authenticatedUserID: actorID},
			},
		},
		{
			name: "instant non-creator",
			activeCall: &ent.Call{
				ID:        callID,
				RoomName:  "instant-room",
				Status:    entcall.StatusActive,
				CreatedBy: creatorID,
			},
			wantOperations: []callJoinCredentialProjectionOperation{
				{name: "load_active_call", callID: callID},
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			source := &callJoinCredentialProjectionSourceFake{
				activeCall: testCase.activeCall,
				member:     testCase.member,
				userName:   "must not be loaded",
			}
			issuer := &callJoinCredentialProjectionIssuerFake{
				credential: calls.JoinCredential{Token: "must not be issued"},
			}
			projector, err := calls.NewCallJoinCredentialProjector(source, issuer)
			require.NoError(t, err)

			credential, err := projector.IssueForAuthenticatedUser(ctx, callID, actorID)

			require.ErrorIs(t, err, calls.ErrCallForbidden)
			assert.Equal(t, calls.JoinCredential{}, credential)
			assert.Equal(t, testCase.wantOperations, source.operations)
			assert.Empty(t, issuer.operations)
		})
	}
}

func TestCallJoinCredentialProjectorStopsAtSourceErrors(t *testing.T) {
	ctx := context.Background()
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	actorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	channelID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	activeChannelCall := &ent.Call{
		ID:        callID,
		RoomName:  "server-owned-room",
		Status:    entcall.StatusActive,
		ChannelID: &channelID,
		CreatedBy: uuid.MustParse("33333333-3333-4333-8333-333333333333"),
	}
	loadErr := errors.New("load active call")
	membershipErr := errors.New("check channel membership")
	nameErr := errors.New("load persisted display name")

	for _, testCase := range []struct {
		name           string
		source         *callJoinCredentialProjectionSourceFake
		wantErr        error
		wantOperations []callJoinCredentialProjectionOperation
	}{
		{
			name: "active reload",
			source: &callJoinCredentialProjectionSourceFake{
				loadErr: loadErr,
			},
			wantErr: loadErr,
			wantOperations: []callJoinCredentialProjectionOperation{
				{name: "load_active_call", callID: callID},
			},
		},
		{
			name: "channel membership",
			source: &callJoinCredentialProjectionSourceFake{
				activeCall: activeChannelCall,
				memberErr:  membershipErr,
			},
			wantErr: membershipErr,
			wantOperations: []callJoinCredentialProjectionOperation{
				{name: "load_active_call", callID: callID},
				{name: "is_channel_member", channelID: channelID, authenticatedUserID: actorID},
			},
		},
		{
			name: "persisted display name",
			source: &callJoinCredentialProjectionSourceFake{
				activeCall: activeChannelCall,
				member:     true,
				nameErr:    nameErr,
			},
			wantErr: nameErr,
			wantOperations: []callJoinCredentialProjectionOperation{
				{name: "load_active_call", callID: callID},
				{name: "is_channel_member", channelID: channelID, authenticatedUserID: actorID},
				{name: "get_server_known_call_user_name", authenticatedUserID: actorID},
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			issuer := &callJoinCredentialProjectionIssuerFake{
				credential: calls.JoinCredential{Token: "must not be issued"},
			}
			projector, err := calls.NewCallJoinCredentialProjector(testCase.source, issuer)
			require.NoError(t, err)

			credential, err := projector.IssueForAuthenticatedUser(ctx, callID, actorID)

			require.ErrorIs(t, err, testCase.wantErr)
			assert.Equal(t, calls.JoinCredential{}, credential)
			assert.Equal(t, testCase.wantOperations, testCase.source.operations)
			assert.Empty(t, issuer.operations)
		})
	}
}

func TestCallJoinCredentialProjectorPassesBlankPersistedNameToIssuer(t *testing.T) {
	ctx := context.Background()
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	actorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	channelID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	activeCall := &ent.Call{
		ID:        callID,
		RoomName:  "server-owned-room",
		Status:    entcall.StatusActive,
		ChannelID: &channelID,
		CreatedBy: uuid.MustParse("33333333-3333-4333-8333-333333333333"),
	}
	source := &callJoinCredentialProjectionSourceFake{
		activeCall: activeCall,
		member:     true,
		userName:   "",
	}
	issuer := &callJoinCredentialProjectionIssuerFake{
		credential: calls.JoinCredential{Token: "must not leak"},
		err:        calls.ErrCallJoinCredentialInvalid,
	}
	projector, err := calls.NewCallJoinCredentialProjector(source, issuer)
	require.NoError(t, err)

	credential, err := projector.IssueForAuthenticatedUser(ctx, callID, actorID)

	require.ErrorIs(t, err, calls.ErrCallJoinCredentialInvalid)
	assert.Equal(t, calls.JoinCredential{}, credential)
	assert.Equal(t, []callJoinCredentialProjectionOperation{
		{name: "load_active_call", callID: callID},
		{name: "is_channel_member", channelID: channelID, authenticatedUserID: actorID},
		{name: "get_server_known_call_user_name", authenticatedUserID: actorID},
	}, source.operations)
	require.Len(t, issuer.operations, 1)
	assert.Same(t, activeCall, issuer.operations[0].call)
	assert.Equal(t, actorID, issuer.operations[0].actorID)
	assert.Empty(t, issuer.operations[0].userName)
}

func TestCallJoinCredentialProjectorReturnsIssuerErrorUnchangedWithZeroCredential(t *testing.T) {
	ctx := context.Background()
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	actorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	channelID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	activeCall := &ent.Call{
		ID:        callID,
		RoomName:  "server-owned-room",
		Status:    entcall.StatusActive,
		ChannelID: &channelID,
		CreatedBy: uuid.MustParse("33333333-3333-4333-8333-333333333333"),
	}
	issuerErr := errors.New("issuer unavailable")
	source := &callJoinCredentialProjectionSourceFake{
		activeCall: activeCall,
		member:     true,
		userName:   "Persisted Display Name",
	}
	issuer := &callJoinCredentialProjectionIssuerFake{
		credential: calls.JoinCredential{Token: "must not leak"},
		err:        issuerErr,
	}
	projector, err := calls.NewCallJoinCredentialProjector(source, issuer)
	require.NoError(t, err)

	credential, err := projector.IssueForAuthenticatedUser(ctx, callID, actorID)

	require.ErrorIs(t, err, issuerErr)
	assert.Equal(t, issuerErr, err)
	assert.Equal(t, calls.JoinCredential{}, credential)
	assert.Equal(t, []callJoinCredentialProjectionOperation{
		{name: "load_active_call", callID: callID},
		{name: "is_channel_member", channelID: channelID, authenticatedUserID: actorID},
		{name: "get_server_known_call_user_name", authenticatedUserID: actorID},
	}, source.operations)
	require.Len(t, issuer.operations, 1)
	assert.Same(t, activeCall, issuer.operations[0].call)
	assert.Equal(t, actorID, issuer.operations[0].actorID)
	assert.Equal(t, "Persisted Display Name", issuer.operations[0].userName)
}
