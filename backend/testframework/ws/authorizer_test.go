package ws_test

// Тесты модуля WebSocket: интеграционные и unit-сценарии testframework.
import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/ws"
)

// TestSplitRoom проверяет сценарий: split room.
func TestSplitRoom(t *testing.T) {
	tests := []struct {
		room       string
		wantPrefix string
		wantID     string
		wantOK     bool
	}{
		{"channel:550e8400-e29b-41d4-a716-446655440000", "channel", "550e8400-e29b-41d4-a716-446655440000", true},
		{"project:660e8400-e29b-41d4-a716-446655440000", "project", "660e8400-e29b-41d4-a716-446655440000", true},
		{"", "", "", false},
		{"no-colon", "", "", false},
		{":missing-prefix", "", "", false},
		{"prefix:", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.room, func(t *testing.T) {
			prefix, id, ok := ws.ExportSplitRoom(tt.room)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantPrefix, prefix)
			assert.Equal(t, tt.wantID, id)
		})
	}
}

// TestAuthorizer_CanSubscribe проверяет сценарий: авторизатор can subscribe.
func TestAuthorizer_CanSubscribe(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	channelID := uuid.New()
	callID := uuid.New()
	projectID := uuid.New()
	checkerErr := errors.New("db unavailable")

	allow := func(_ context.Context, entityID, uid uuid.UUID) (bool, error) {
		return entityID == channelID && uid == userID, nil
	}
	deny := func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
		return false, nil
	}
	fail := func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
		return false, checkerErr
	}

	tests := []struct {
		name    string
		cfg     ws.AuthorizerConfig
		room    string
		user    string
		wantOK  bool
		wantErr bool
	}{
		{
			name:    "invalid user id",
			cfg:     ws.AuthorizerConfig{IsChannelMember: allow},
			room:    "channel:" + channelID.String(),
			user:    "not-uuid",
			wantErr: true,
		},
		{
			name:    "malformed room",
			cfg:     ws.AuthorizerConfig{IsChannelMember: allow},
			room:    "channel-only",
			user:    userID.String(),
			wantErr: true,
		},
		{
			name:    "invalid entity uuid",
			cfg:     ws.AuthorizerConfig{IsChannelMember: allow},
			room:    "channel:not-a-uuid",
			user:    userID.String(),
			wantErr: true,
		},
		{
			name:    "unknown prefix fail-closed",
			cfg:     ws.AuthorizerConfig{IsChannelMember: allow},
			room:    "unknown:" + channelID.String(),
			user:    userID.String(),
			wantOK:  false,
			wantErr: false,
		},
		{
			name:    "nil channel checker fail-closed",
			cfg:     ws.AuthorizerConfig{},
			room:    "channel:" + channelID.String(),
			user:    userID.String(),
			wantOK:  false,
		},
		{
			name:    "channel member allowed",
			cfg:     ws.AuthorizerConfig{IsChannelMember: allow},
			room:    "channel:" + channelID.String(),
			user:    userID.String(),
			wantOK:  true,
		},
		{
			name:    "channel non-member denied",
			cfg:     ws.AuthorizerConfig{IsChannelMember: deny},
			room:    "channel:" + channelID.String(),
			user:    userID.String(),
			wantOK:  false,
		},
		{
			name: "call checker error propagates",
			cfg:  ws.AuthorizerConfig{IsCallParticipantOrChannelMember: fail},
			room: "call:" + callID.String(),
			user: userID.String(),
			wantErr: true,
		},
		{
			name:    "nil call checker fail-closed",
			cfg:     ws.AuthorizerConfig{},
			room:    "call:" + callID.String(),
			user:    userID.String(),
			wantOK:  false,
		},
		{
			name: "call participant allowed",
			cfg: ws.AuthorizerConfig{
				IsCallParticipantOrChannelMember: func(_ context.Context, entityID, uid uuid.UUID) (bool, error) {
					return entityID == callID && uid == userID, nil
				},
			},
			room:   "call:" + callID.String(),
			user:   userID.String(),
			wantOK: true,
		},
		{
			name:    "nil project checker fail-closed",
			cfg:     ws.AuthorizerConfig{},
			room:    "project:" + projectID.String(),
			user:    userID.String(),
			wantOK:  false,
		},
		{
			name: "project member allowed",
			cfg: ws.AuthorizerConfig{
				IsProjectMember: func(_ context.Context, entityID, uid uuid.UUID) (bool, error) {
					return entityID == projectID && uid == userID, nil
				},
			},
			room:   "project:" + projectID.String(),
			user:   userID.String(),
			wantOK: true,
		},
		{
			name: "project non-member denied",
			cfg: ws.AuthorizerConfig{
				IsProjectMember: deny,
			},
			room:   "project:" + projectID.String(),
			user:   userID.String(),
			wantOK: false,
		},
		{
			name: "project checker error propagates",
			cfg: ws.AuthorizerConfig{
				IsProjectMember: fail,
			},
			room:    "project:" + projectID.String(),
			user:    userID.String(),
			wantErr: true,
		},
		{
			name: "call non-participant denied",
			cfg: ws.AuthorizerConfig{
				IsCallParticipantOrChannelMember: deny,
			},
			room:   "call:" + callID.String(),
			user:   userID.String(),
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authz := ws.NewAuthorizer(tt.cfg)
			ok, err := authz.CanSubscribe(ctx, tt.user, tt.room)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}
