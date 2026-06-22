package auth_test

// Тесты сервиса аутентификации: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
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
	"github.com/valerii/hub/backend/ent/user"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

func newSessionTestEnv(t *testing.T) (*auth.SessionService, *ent.Client, *miniredis.Miniredis) {
	t.Helper()
	dsn := "file:sessiontest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	svc := auth.NewSessionService(auth.NewEntRepository(client), rdb, nil, "test-jwt-secret", 15*time.Minute, 30*24*time.Hour)
	return svc, client, mr
}

// TestHashRefreshToken_Deterministic проверяет сценарий: хеш refresh token deterministic.
func TestHashRefreshToken_Deterministic(t *testing.T) {
	token := "550e8400-e29b-41d4-a716-446655440000"
	h1 := auth.ExportHashRefreshToken(token)
	h2 := auth.ExportHashRefreshToken(token)

	assert.Equal(t, h1, h2)
	assert.Len(t, h1, 64)
	assert.NotEqual(t, token, h1)
}

// TestRefreshRedisKey_UsesHashNotPlaintext проверяет сценарий: refresh redis key uses хеш not plaintext.
func TestRefreshRedisKey_UsesHashNotPlaintext(t *testing.T) {
	plain := uuid.New().String()
	key := auth.ExportRefreshRedisKey(plain)
	assert.Contains(t, key, "refresh:")
	assert.NotContains(t, key, plain)
	assert.Equal(t, "refresh:"+auth.ExportHashRefreshToken(plain), key)
}

// TestLogout_RemovesRedisAndSession проверяет сценарий: выход removes redis and сессия.
func TestLogout_RemovesRedisAndSession(t *testing.T) {
	svc, client, mr := newSessionTestEnv(t)
	ctx := context.Background()

	u, err := client.User.Create().SetEmail("logout-" + uuid.New().String()[:8] + "@example.com").Save(ctx)
	require.NoError(t, err)

	sessionID := uuid.New()
	plainRefresh := uuid.New().String()
	tokenHash := auth.ExportHashRefreshToken(plainRefresh)

	payload, err := json.Marshal(auth.ExportRefreshPayload{
		UserID: u.ID.String(), SessionID: sessionID.String(), Device: "test",
	})
	require.NoError(t, err)
	require.NoError(t, mr.Set(auth.ExportRefreshRedisKey(plainRefresh), string(payload)))

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
	assert.False(t, mr.Exists(auth.ExportRefreshRedisKeyStored(tokenHash)))
}

// TestLogout_IdempotentNotFound проверяет сценарий: выход idempotent not found.
func TestLogout_IdempotentNotFound(t *testing.T) {
	svc, _, _ := newSessionTestEnv(t)
	err := svc.Logout(context.Background(), uuid.New())
	require.NoError(t, err)
}

// TestRefreshToken_RotatesHashAndRedisKey проверяет сценарий: refresh token rotates хеш and redis key.
func TestRefreshToken_RotatesHashAndRedisKey(t *testing.T) {
	svc, client, mr := newSessionTestEnv(t)
	ctx := context.Background()

	u, err := client.User.Create().SetEmail("refresh-" + uuid.New().String()[:8] + "@example.com").Save(ctx)
	require.NoError(t, err)

	sessionID := uuid.New()
	oldPlain := uuid.New().String()
	oldHash := auth.ExportHashRefreshToken(oldPlain)

	payload, err := json.Marshal(auth.ExportRefreshPayload{
		UserID: u.ID.String(), SessionID: sessionID.String(), Device: "browser",
	})
	require.NoError(t, err)
	require.NoError(t, mr.Set(auth.ExportRefreshRedisKey(oldPlain), string(payload)))

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
	newHash := auth.ExportHashRefreshToken(pair.RefreshToken)
	assert.Equal(t, newHash, sess.RefreshToken)
	assert.False(t, mr.Exists(auth.ExportRefreshRedisKey(oldPlain)))
	assert.True(t, mr.Exists(auth.ExportRefreshRedisKey(pair.RefreshToken)))
}

// TestStoreRefreshToken_StoresHashedKey проверяет сценарий: store refresh token stores hashed key.
func TestStoreRefreshToken_StoresHashedKey(t *testing.T) {
	svc, _, mr := newSessionTestEnv(t)
	ctx := context.Background()

	plain := uuid.New().String()
	userID := uuid.New()
	sessionID := uuid.New()

	require.NoError(t, auth.ExportSessionStoreRefreshToken(svc, ctx, plain, auth.ExportRefreshPayload{
		UserID: userID.String(), SessionID: sessionID.String(), Device: "mobile",
	}))

	assert.True(t, mr.Exists(auth.ExportRefreshRedisKey(plain)))
	assert.False(t, mr.Exists("refresh:"+plain))
}

// TestRefreshToken_NotFound проверяет сценарий: refresh token not found.
func TestRefreshToken_NotFound(t *testing.T) {
	svc, _, _ := newSessionTestEnv(t)
	_, err := svc.RefreshToken(context.Background(), uuid.New().String())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "refresh token not found")
}

func setupRefreshFixture(t *testing.T, svc *auth.SessionService, client *ent.Client, mr *miniredis.Miniredis, opts refreshFixtureOpts) (uuid.UUID, string) {
	t.Helper()
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("refresh-fix-" + uuid.New().String()[:8] + "@example.com").
		SetNillableStatus(opts.status).
		Save(ctx)
	require.NoError(t, err)

	sessionID := uuid.New()
	oldPlain := uuid.New().String()
	oldHash := auth.ExportHashRefreshToken(oldPlain)
	if opts.storedHash != "" {
		oldHash = opts.storedHash
	}

	payload, err := json.Marshal(auth.ExportRefreshPayload{
		UserID: u.ID.String(), SessionID: sessionID.String(), Device: "browser",
	})
	require.NoError(t, err)
	require.NoError(t, mr.Set(auth.ExportRefreshRedisKey(oldPlain), string(payload)))

	if opts.createSession {
		expiresAt := opts.expiresAt
		if expiresAt.IsZero() {
			expiresAt = time.Now().UTC().Add(24 * time.Hour)
		}
		_, err = client.Session.Create().
			SetID(sessionID).
			SetRefreshToken(oldHash).
			SetDevice("browser").
			SetExpiresAt(expiresAt).
			SetUserID(u.ID).
			Save(ctx)
		require.NoError(t, err)
	}

	return u.ID, oldPlain
}

type refreshFixtureOpts struct {
	createSession bool
	expiresAt     time.Time
	storedHash    string
	status        *user.Status
}

// TestRefreshToken_RejectsMissingDBSession проверяет отклонение refresh при отсутствии сессии в БД.
func TestRefreshToken_RejectsMissingDBSession(t *testing.T) {
	svc, client, mr := newSessionTestEnv(t)
	ctx := context.Background()

	_, oldPlain := setupRefreshFixture(t, svc, client, mr, refreshFixtureOpts{createSession: false})

	_, err := svc.RefreshToken(ctx, oldPlain)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
	assert.True(t, mr.Exists(auth.ExportRefreshRedisKey(oldPlain)))
}

// TestRefreshToken_RejectsExpiredSession проверяет отклонение refresh для истёкшей сессии.
func TestRefreshToken_RejectsExpiredSession(t *testing.T) {
	svc, client, mr := newSessionTestEnv(t)
	ctx := context.Background()

	_, oldPlain := setupRefreshFixture(t, svc, client, mr, refreshFixtureOpts{
		createSession: true,
		expiresAt:     time.Now().UTC().Add(-time.Hour),
	})

	_, err := svc.RefreshToken(ctx, oldPlain)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session expired")
}

// TestRefreshToken_RejectsHashMismatch проверяет отклонение refresh при несовпадении hash в БД.
func TestRefreshToken_RejectsHashMismatch(t *testing.T) {
	svc, client, mr := newSessionTestEnv(t)
	ctx := context.Background()

	_, oldPlain := setupRefreshFixture(t, svc, client, mr, refreshFixtureOpts{
		createSession: true,
		storedHash:    auth.ExportHashRefreshToken(uuid.New().String()),
	})

	_, err := svc.RefreshToken(ctx, oldPlain)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hash mismatch")
}

// TestRefreshToken_RejectsDeactivatedUser проверяет отклонение refresh для деактивированного пользователя.
func TestRefreshToken_RejectsDeactivatedUser(t *testing.T) {
	svc, client, mr := newSessionTestEnv(t)
	ctx := context.Background()

	deactivated := user.StatusDeactivated
	_, oldPlain := setupRefreshFixture(t, svc, client, mr, refreshFixtureOpts{
		createSession: true,
		status:        &deactivated,
	})

	_, err := svc.RefreshToken(ctx, oldPlain)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user deactivated")
}

// TestCreateSession_SingleSessionPolicy проверяет сценарий: создание сессия single сессия policy.
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
	assert.False(t, mr.Exists(auth.ExportRefreshRedisKey(first.RefreshToken)))
	assert.True(t, mr.Exists(auth.ExportRefreshRedisKey(second.RefreshToken)))
}

// TestRevokeSession_Ownership проверяет сценарий: revoke сессия ownership.
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
	assert.False(t, mr.Exists(auth.ExportRefreshRedisKey(pair.RefreshToken)))
}

// TestRevokeSession_NotOwned проверяет сценарий: revoke сессия not owned.
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

// TestParseAccessToken_Invalid проверяет отклонение при разбор access token некорректный.
func TestParseAccessToken_Invalid(t *testing.T) {
	svc, _, _ := newSessionTestEnv(t)
	_, err := svc.ParseAccessToken("not-a-jwt")
	require.Error(t, err)
}

// TestListSessions_ReturnsUserSessions проверяет сценарий: список sessions returns user sessions.
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

// TestInvalidateUserSessions_ClearsMultiple проверяет сценарий: invalidate user sessions clears multiple.
func TestInvalidateUserSessions_ClearsMultiple(t *testing.T) {
	svc, client, mr := newSessionTestEnv(t)
	ctx := context.Background()

	u, err := client.User.Create().SetEmail("multi-" + uuid.New().String()[:8] + "@example.com").Save(ctx)
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		sessionID := uuid.New()
		plain := uuid.New().String()
		tokenHash := auth.ExportHashRefreshToken(plain)
		payload, err := json.Marshal(auth.ExportRefreshPayload{
			UserID: u.ID.String(), SessionID: sessionID.String(), Device: "dev",
		})
		require.NoError(t, err)
		require.NoError(t, mr.Set(auth.ExportRefreshRedisKey(plain), string(payload)))
		_, err = client.Session.Create().
			SetID(sessionID).
			SetRefreshToken(tokenHash).
			SetDevice("dev").
			SetExpiresAt(time.Now().UTC().Add(time.Hour)).
			SetUserID(u.ID).
			Save(ctx)
		require.NoError(t, err)
	}

	require.NoError(t, auth.ExportSessionInvalidateUserSessions(svc, ctx, u.ID))

	count, err := client.Session.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestPublishSessionCreated_WithNATS проверяет сценарий: publish сессия created with nats.
func TestPublishSessionCreated_WithNATS(t *testing.T) {
	svc, client, _ := newSessionTestEnv(t)
	ctx := context.Background()

	opts := &server.Options{Port: -1}
	ns := natsserver.RunServer(opts)
	t.Cleanup(ns.Shutdown)

	nc, err := natsclient.NewClient(ctx, ns.ClientURL())
	require.NoError(t, err)
	t.Cleanup(nc.Close)
	auth.ExportSessionSetNATS(svc, nc)

	u, err := client.User.Create().SetEmail("nats-" + uuid.New().String()[:8] + "@example.com").Save(ctx)
	require.NoError(t, err)

	ch := make(chan []byte, 1)
	_, err = nc.Conn.Subscribe("auth.session.created", func(msg *nats.Msg) {
		ch <- append([]byte(nil), msg.Data...)
	})
	require.NoError(t, err)
	require.NoError(t, nc.Conn.Flush())

	auth.ExportSessionPublishSessionCreated(svc, u.ID.String(), uuid.New().String(), "browser", "127.0.0.1")

	select {
	case data := <-ch:
		assert.Contains(t, string(data), u.ID.String())
	case <-time.After(2 * time.Second):
		t.Fatal("expected auth.session.created event")
	}
}

// TestPublishSessionCreated_NilNATS проверяет сценарий: publish сессия created nil nats.
func TestPublishSessionCreated_NilNATS(t *testing.T) {
	svc, _, _ := newSessionTestEnv(t)
	assert.NotPanics(t, func() {
		auth.ExportSessionPublishSessionCreated(svc, uuid.New().String(), uuid.New().String(), "browser", "127.0.0.1")
	})
}
