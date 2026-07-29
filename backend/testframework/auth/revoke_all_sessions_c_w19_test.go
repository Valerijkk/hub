package auth_test

// C-W19-001 / TASK-177: RevokeAllUserSessions = invalidateUserSessions + publish revoked.
// Mutation: убрать publishSessionEnded из RevokeAllUserSessions → NATS timeout RED.
// Mutation: убрать invalidateUserSessions → Redis refresh key остаётся / DB session жива.

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/auth"
)

// TestRevokeAllUserSessions_PublishesRevokedAndInvalidatesRedis_C_W19 —
// Deactivate kill-path: NATS auth.session.{uid}.revoked + Redis Del refresh.
func TestRevokeAllUserSessions_PublishesRevokedAndInvalidatesRedis_C_W19(t *testing.T) {
	svc, client, mr := newSessionTestEnv(t)
	ctx := context.Background()
	nc := attachEmbeddedNATS(t, svc)

	u, err := client.User.Create().
		SetEmail("c-w19-revoke-all-" + uuid.New().String()[:8] + "@example.com").
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)
	claims, err := svc.ParseAccessToken(pair.AccessToken)
	require.NoError(t, err)
	sessionID := claims.SessionID

	refreshKey := auth.ExportRefreshRedisKey(pair.RefreshToken)
	require.True(t, mr.Exists(refreshKey), "refresh must exist before revoke-all")

	subject := fmt.Sprintf("auth.session.%s.revoked", u.ID.String())
	ch := make(chan *nats.Msg, 2)
	_, err = nc.Conn.Subscribe(subject, func(msg *nats.Msg) {
		ch <- msg
	})
	require.NoError(t, err)
	require.NoError(t, nc.Conn.Flush())

	require.NoError(t, svc.RevokeAllUserSessions(ctx, u.ID))

	select {
	case msg := <-ch:
		var ev auth.SessionEndedEvent
		require.NoError(t, json.Unmarshal(msg.Data, &ev))
		assert.Equal(t, u.ID.String(), ev.UserID)
		assert.Equal(t, sessionID, ev.SessionID)
		assert.Equal(t, "revoked", ev.Reason)
	case <-time.After(2 * time.Second):
		t.Fatalf("expected NATS publish on %s (mutation: remove publishSessionEnded)", subject)
	}

	assert.False(t, mr.Exists(refreshKey), "invalidateUserSessions must Del Redis refresh")

	sessions, err := svc.ListSessions(ctx, u.ID)
	require.NoError(t, err)
	assert.Empty(t, sessions, "DB sessions must be deleted")
}

// TestRevokeAllUserSessions_EmptySessions_PublishesEmptySID_C_W19 —
// empty session_id → Bridge UnregisterUser (kill all WS).
func TestRevokeAllUserSessions_EmptySessions_PublishesEmptySID_C_W19(t *testing.T) {
	svc, client, _ := newSessionTestEnv(t)
	ctx := context.Background()
	nc := attachEmbeddedNATS(t, svc)

	u, err := client.User.Create().
		SetEmail("c-w19-empty-" + uuid.New().String()[:8] + "@example.com").
		Save(ctx)
	require.NoError(t, err)

	subject := fmt.Sprintf("auth.session.%s.revoked", u.ID.String())
	ch := make(chan *nats.Msg, 1)
	_, err = nc.Conn.Subscribe(subject, func(msg *nats.Msg) {
		ch <- msg
	})
	require.NoError(t, err)
	require.NoError(t, nc.Conn.Flush())

	require.NoError(t, svc.RevokeAllUserSessions(ctx, u.ID))

	select {
	case msg := <-ch:
		var ev auth.SessionEndedEvent
		require.NoError(t, json.Unmarshal(msg.Data, &ev))
		assert.Equal(t, u.ID.String(), ev.UserID)
		assert.Equal(t, "", ev.SessionID)
		assert.Equal(t, "revoked", ev.Reason)
	case <-time.After(2 * time.Second):
		t.Fatalf("expected empty-sid revoked publish on %s", subject)
	}
}
