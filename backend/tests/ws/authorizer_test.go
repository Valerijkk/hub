package ws_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/ws"
)

func TestAuthorizer_CanSubscribe_DispatchesMembershipChecker(t *testing.T) {
	userID := uuid.New()
	entityID := uuid.New()

	tests := []struct {
		name string
		room string
		cfg  ws.AuthorizerConfig
	}{
		{
			name: "channel",
			room: "channel:" + entityID.String(),
			cfg: ws.AuthorizerConfig{
				IsChannelMember: assertMembershipCall(t, entityID, userID),
			},
		},
		{
			name: "call",
			room: "call:" + entityID.String(),
			cfg: ws.AuthorizerConfig{
				IsCallParticipantOrChannelMember: assertMembershipCall(t, entityID, userID),
			},
		},
		{
			name: "project",
			room: "project:" + entityID.String(),
			cfg: ws.AuthorizerConfig{
				IsProjectMember: assertMembershipCall(t, entityID, userID),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := ws.NewAuthorizer(tt.cfg).CanSubscribe(context.Background(), userID.String(), tt.room)

			require.NoError(t, err)
			assert.True(t, allowed)
		})
	}
}

func TestAuthorizer_CanSubscribe_ReturnsCheckerDecisionAndError(t *testing.T) {
	userID := uuid.New()
	entityID := uuid.New()
	checkerError := errors.New("membership store unavailable")

	tests := []struct {
		name    string
		allowed bool
		err     error
	}{
		{name: "denied"},
		{name: "checker error", err: checkerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authorizer := ws.NewAuthorizer(ws.AuthorizerConfig{
				IsChannelMember: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
					return tt.allowed, tt.err
				},
			})

			allowed, err := authorizer.CanSubscribe(context.Background(), userID.String(), "channel:"+entityID.String())

			assert.Equal(t, tt.allowed, allowed)
			if tt.err == nil {
				require.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestAuthorizer_CanSubscribe_FailsClosedForNilCheckerAndUnknownRoom(t *testing.T) {
	userID := uuid.New()
	entityID := uuid.New()

	tests := []struct {
		name string
		room string
	}{
		{name: "channel without checker", room: "channel:" + entityID.String()},
		{name: "call without checker", room: "call:" + entityID.String()},
		{name: "project without checker", room: "project:" + entityID.String()},
		{name: "unknown prefix", room: "audit:" + entityID.String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := ws.NewAuthorizer(ws.AuthorizerConfig{}).CanSubscribe(context.Background(), userID.String(), tt.room)

			require.NoError(t, err)
			assert.False(t, allowed)
		})
	}
}

func TestAuthorizer_CanSubscribe_RejectsMalformedIdentifiersAndRooms(t *testing.T) {
	validID := uuid.New().String()
	authorizer := ws.NewAuthorizer(ws.AuthorizerConfig{
		IsChannelMember: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
			return true, nil
		},
	})

	tests := []struct {
		name string
		user string
		room string
	}{
		{name: "invalid user id", user: "not-a-uuid", room: "channel:" + validID},
		{name: "missing room separator", user: validID, room: "channel"},
		{name: "empty room prefix", user: validID, room: ":" + validID},
		{name: "empty entity id", user: validID, room: "channel:"},
		{name: "invalid entity id", user: validID, room: "channel:not-a-uuid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := authorizer.CanSubscribe(context.Background(), tt.user, tt.room)

			assert.False(t, allowed)
			assert.Error(t, err)
		})
	}
}

func assertMembershipCall(t *testing.T, wantEntityID, wantUserID uuid.UUID) ws.MembershipChecker {
	t.Helper()
	return func(_ context.Context, entityID, userID uuid.UUID) (bool, error) {
		assert.Equal(t, wantEntityID, entityID)
		assert.Equal(t, wantUserID, userID)
		return true, nil
	}
}
