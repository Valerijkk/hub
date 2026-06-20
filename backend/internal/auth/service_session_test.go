package auth

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

func newSessionTestEnv(t *testing.T) (*SessionService, *ent.Client, *miniredis.Miniredis) {
	t.Helper()
	dsn := "file:sessiontest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	svc := NewSessionService(client, rdb, nil, "test-jwt-secret", 15*time.Minute, 30*24*time.Hour)
	return svc, client, mr
}

func TestHashRefreshToken_Deterministic(t *testing.T) {
	token := "550e8400-e29b-41d4-a716-446655440000"
	h1 := hashRefreshToken(token)
	h2 := hashRefreshToken(token)

	assert.Equal(t, h1, h2)
	assert.Len(t, h1, 64)
	assert.NotEqual(t, token, h1)
}

func TestRefreshRedisKey_UsesHashNotPlaintext(t *testing.T) {
	plain := uuid.New().String()
	key := refreshRedisKey(plain)
	assert.Contains(t, key, "refresh:")
	assert.NotContains(t, key, plain)
	assert.Equal(t, "refresh:"+hashRefreshToken(plain), key)
}

func TestLogout_RemovesRedisAndSession(t *testing.T) {
	svc, client, mr := newSessionTestEnv(t)
	ctx := context.Background()

	u, err := client.User.Create().SetEmail("logout-" + uuid.New().String()[:8] + "@example.com").Save(ctx)
	require.NoError(t, err)

	sessionID := uuid.New()
	plainRefresh := uuid.New().String()
	tokenHash := hashRefreshToken(plainRefresh)

	payload, err := json.Marshal(refreshPayload{
		UserID: u.ID.String(), SessionID: sessionID.String(), Device: "test",
	})
	require.NoError(t, err)
	require.NoError(t, mr.Set(refreshRedisKey(plainRefresh), string(payload)))

	_, err = client.Session.Create().
		SetID(sessionID).
		SetRefreshToken(tokenHash).
		SetDevice("test").
		SetExpiresAt(time.Now().UTC().Add(24 * time.Hour)).
		SetUserID(u.ID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.Logout(ctx, sessionID))

	_, err = client.Session.Get(ctx, sessionID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
	assert.False(t, mr.Exists(refreshRedisKeyStored(tokenHash)))
}

func TestLogout_IdempotentNotFound(t *testing.T) {
	svc, _, _ := newSessionTestEnv(t)
	err := svc.Logout(context.Background(), uuid.New())
	require.NoError(t, err)
}

func TestRefreshToken_RotatesHashAndRedisKey(t *testing.T) {
	svc, client, mr := newSessionTestEnv(t)
	ctx := context.Background()

	u, err := client.User.Create().SetEmail("refresh-" + uuid.New().String()[:8] + "@example.com").Save(ctx)
	require.NoError(t, err)

	sessionID := uuid.New()
	oldPlain := uuid.New().String()
	oldHash := hashRefreshToken(oldPlain)

	payload, err := json.Marshal(refreshPayload{
		UserID: u.ID.String(), SessionID: sessionID.String(), Device: "browser",
	})
	require.NoError(t, err)
	require.NoError(t, mr.Set(refreshRedisKey(oldPlain), string(payload)))

	_, err = client.Session.Create().
		SetID(sessionID).
		SetRefreshToken(oldHash).
		SetDevice("browser").
		SetExpiresAt(time.Now().UTC().Add(24 * time.Hour)).
		SetUserID(u.ID).
		Save(ctx)
	require.NoError(t, err)

	pair, err := svc.RefreshToken(ctx, oldPlain)
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)
	require.NotEmpty(t, pair.RefreshToken)
	assert.NotEqual(t, oldPlain, pair.RefreshToken)

	sess, err := client.Session.Get(ctx, sessionID)
	require.NoError(t, err)
	newHash := hashRefreshToken(pair.RefreshToken)
	assert.Equal(t, newHash, sess.RefreshToken)
	assert.False(t, mr.Exists(refreshRedisKey(oldPlain)))
	assert.True(t, mr.Exists(refreshRedisKey(pair.RefreshToken)))
}

func TestStoreRefreshToken_StoresHashedKey(t *testing.T) {
	svc, _, mr := newSessionTestEnv(t)
	ctx := context.Background()

	plain := uuid.New().String()
	userID := uuid.New()
	sessionID := uuid.New()

	require.NoError(t, svc.storeRefreshToken(ctx, plain, refreshPayload{
		UserID: userID.String(), SessionID: sessionID.String(), Device: "mobile",
	}))

	assert.True(t, mr.Exists(refreshRedisKey(plain)))
	assert.False(t, mr.Exists("refresh:"+plain))
}

func TestRefreshToken_NotFound(t *testing.T) {
	svc, _, _ := newSessionTestEnv(t)
	_, err := svc.RefreshToken(context.Background(), uuid.New().String())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "refresh token not found")
}

func TestCreateSession_SingleSessionPolicy(t *testing.T) {
	svc, client, mr := newSessionTestEnv(t)
	ctx := context.Background()

	u, err := client.User.Create().SetEmail("single-" + uuid.New().String()[:8] + "@example.com").Save(ctx)
	require.NoError(t, err)

	first, err := svc.CreateSession(ctx, u.ID, "device-a", "1.1.1.1")
	require.NoError(t, err)

	sessions, err := client.Session.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, sessions, 1)

	second, err := svc.CreateSession(ctx, u.ID, "device-b", "2.2.2.2")
	require.NoError(t, err)
	assert.NotEqual(t, first.RefreshToken, second.RefreshToken)

	sessions, err = client.Session.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.False(t, mr.Exists(refreshRedisKey(first.RefreshToken)))
	assert.True(t, mr.Exists(refreshRedisKey(second.RefreshToken)))
}

func TestRevokeSession_Ownership(t *testing.T) {
	svc, client, mr := newSessionTestEnv(t)
	ctx := context.Background()

	u, err := client.User.Create().SetEmail("revoke-" + uuid.New().String()[:8] + "@example.com").Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	claims, err := svc.ParseAccessToken(pair.AccessToken)
	require.NoError(t, err)
	sessionID, err := uuid.Parse(claims.SessionID)
	require.NoError(t, err)

	require.NoError(t, svc.RevokeSession(ctx, u.ID, sessionID))

	_, err = client.Session.Get(ctx, sessionID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
	assert.False(t, mr.Exists(refreshRedisKey(pair.RefreshToken)))
}

func TestRevokeSession_NotOwned(t *testing.T) {
	svc, client, _ := newSessionTestEnv(t)
	ctx := context.Background()

	owner, err := client.User.Create().SetEmail("owner-" + uuid.New().String()[:8] + "@example.com").Save(ctx)
	require.NoError(t, err)
	other, err := client.User.Create().SetEmail("other-" + uuid.New().String()[:8] + "@example.com").Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, owner.ID, "browser", "127.0.0.1")
	require.NoError(t, err)
	claims, err := svc.ParseAccessToken(pair.AccessToken)
	require.NoError(t, err)
	sessionID, err := uuid.Parse(claims.SessionID)
	require.NoError(t, err)

	err = svc.RevokeSession(ctx, other.ID, sessionID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not found or not owned by user")
}

func TestParseAccessToken_Invalid(t *testing.T) {
	svc, _, _ := newSessionTestEnv(t)
	_, err := svc.ParseAccessToken("not-a-jwt")
	require.Error(t, err)
}

func TestListSessions_ReturnsUserSessions(t *testing.T) {
	svc, client, _ := newSessionTestEnv(t)
	ctx := context.Background()

	u, err := client.User.Create().SetEmail("list-" + uuid.New().String()[:8] + "@example.com").Save(ctx)
	require.NoError(t, err)

	_, err = svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	sessions, err := svc.ListSessions(ctx, u.ID)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, "browser", sessions[0].Device)
}

func TestInvalidateUserSessions_ClearsMultiple(t *testing.T) {
	svc, client, mr := newSessionTestEnv(t)
	ctx := context.Background()

	u, err := client.User.Create().SetEmail("multi-" + uuid.New().String()[:8] + "@example.com").Save(ctx)
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		sessionID := uuid.New()
		plain := uuid.New().String()
		tokenHash := hashRefreshToken(plain)
		payload, err := json.Marshal(refreshPayload{
			UserID: u.ID.String(), SessionID: sessionID.String(), Device: "dev",
		})
		require.NoError(t, err)
		require.NoError(t, mr.Set(refreshRedisKey(plain), string(payload)))
		_, err = client.Session.Create().
			SetID(sessionID).
			SetRefreshToken(tokenHash).
			SetDevice("dev").
			SetExpiresAt(time.Now().UTC().Add(time.Hour)).
			SetUserID(u.ID).
			Save(ctx)
		require.NoError(t, err)
	}

	require.NoError(t, svc.invalidateUserSessions(ctx, u.ID))

	count, err := client.Session.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestPublishSessionCreated_WithNATS(t *testing.T) {
	svc, client, _ := newSessionTestEnv(t)
	ctx := context.Background()

	opts := &server.Options{Port: -1}
	ns := natsserver.RunServer(opts)
	t.Cleanup(ns.Shutdown)

	nc, err := natsclient.NewClient(ctx, ns.ClientURL())
	require.NoError(t, err)
	t.Cleanup(nc.Close)
	svc.nats = nc

	u, err := client.User.Create().SetEmail("nats-" + uuid.New().String()[:8] + "@example.com").Save(ctx)
	require.NoError(t, err)

	ch := make(chan []byte, 1)
	_, err = nc.Conn.Subscribe("auth.session.created", func(msg *nats.Msg) {
		ch <- append([]byte(nil), msg.Data...)
	})
	require.NoError(t, err)
	require.NoError(t, nc.Conn.Flush())

	svc.publishSessionCreated(u.ID.String(), uuid.New().String(), "browser", "127.0.0.1")

	select {
	case data := <-ch:
		assert.Contains(t, string(data), u.ID.String())
	case <-time.After(2 * time.Second):
		t.Fatal("expected auth.session.created event")
	}
}

func TestPublishSessionCreated_NilNATS(t *testing.T) {
	svc, _, _ := newSessionTestEnv(t)
	assert.NotPanics(t, func() {
		svc.publishSessionCreated(uuid.New().String(), uuid.New().String(), "browser", "127.0.0.1")
	})
}
