package auth_test

// TASK-109 / A2W3-4-002: Logout/Revoke публикуют auth.session.{uid}.{logout|revoked}.
// Mutation: убрать publishSessionEnded из Logout/Revoke → тест RED (таймаут на subject).

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/auth"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

func attachEmbeddedNATS(t *testing.T, svc *auth.SessionService) *natsclient.Client {
	t.Helper()
	opts := &server.Options{Port: -1}
	ns := natsserver.RunServer(opts)
	t.Cleanup(ns.Shutdown)

	nc, err := natsclient.NewClient(context.Background(), ns.ClientURL())
	require.NoError(t, err)
	t.Cleanup(nc.Close)
	auth.ExportSessionSetNATS(svc, nc)
	return nc
}

// TestLogout_PublishesAuthSessionLogout_TASK109 — Logout → auth.session.{uid}.logout + session_id.
func TestLogout_PublishesAuthSessionLogout_TASK109(t *testing.T) {
	svc, client, _ := newSessionTestEnv(t)
	ctx := context.Background()
	nc := attachEmbeddedNATS(t, svc)

	u, err := client.User.Create().SetEmail("logout-nats-" + uuid.New().String()[:8] + "@example.com").Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)
	claims, err := svc.ParseAccessToken(pair.AccessToken)
	require.NoError(t, err)
	sessionID, err := uuid.Parse(claims.SessionID)
	require.NoError(t, err)

	subject := fmt.Sprintf("auth.session.%s.logout", u.ID.String())
	ch := make(chan *nats.Msg, 1)
	_, err = nc.Conn.Subscribe(subject, func(msg *nats.Msg) {
		ch <- msg
	})
	require.NoError(t, err)
	require.NoError(t, nc.Conn.Flush())

	require.NoError(t, svc.Logout(ctx, sessionID))

	select {
	case msg := <-ch:
		var ev auth.SessionEndedEvent
		require.NoError(t, json.Unmarshal(msg.Data, &ev))
		assert.Equal(t, u.ID.String(), ev.UserID)
		assert.Equal(t, sessionID.String(), ev.SessionID)
		assert.Equal(t, "logout", ev.Reason)
	case <-time.After(2 * time.Second):
		t.Fatalf("expected NATS publish on %s", subject)
	}
}

// TestRevokeSession_PublishesAuthSessionRevoked_TASK109 — Revoke → auth.session.{uid}.revoked + session_id.
func TestRevokeSession_PublishesAuthSessionRevoked_TASK109(t *testing.T) {
	svc, client, _ := newSessionTestEnv(t)
	ctx := context.Background()
	nc := attachEmbeddedNATS(t, svc)

	u, err := client.User.Create().SetEmail("revoke-nats-" + uuid.New().String()[:8] + "@example.com").Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)
	claims, err := svc.ParseAccessToken(pair.AccessToken)
	require.NoError(t, err)
	sessionID, err := uuid.Parse(claims.SessionID)
	require.NoError(t, err)

	subject := fmt.Sprintf("auth.session.%s.revoked", u.ID.String())
	ch := make(chan *nats.Msg, 1)
	_, err = nc.Conn.Subscribe(subject, func(msg *nats.Msg) {
		ch <- msg
	})
	require.NoError(t, err)
	require.NoError(t, nc.Conn.Flush())

	require.NoError(t, svc.RevokeSession(ctx, u.ID, sessionID))

	select {
	case msg := <-ch:
		var ev auth.SessionEndedEvent
		require.NoError(t, json.Unmarshal(msg.Data, &ev))
		assert.Equal(t, u.ID.String(), ev.UserID)
		assert.Equal(t, sessionID.String(), ev.SessionID)
		assert.Equal(t, "revoked", ev.Reason)
	case <-time.After(2 * time.Second):
		t.Fatalf("expected NATS publish on %s", subject)
	}
}

// TestPublishSessionEnded_NilNATS_TASK109 — nil NATS не паникует.
func TestPublishSessionEnded_NilNATS_TASK109(t *testing.T) {
	svc, _, _ := newSessionTestEnv(t)
	assert.NotPanics(t, func() {
		auth.ExportSessionPublishSessionEnded(svc, uuid.New().String(), uuid.New().String(), "logout")
	})
}
