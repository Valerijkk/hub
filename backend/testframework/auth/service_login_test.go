package auth_test

// Тесты сервиса аутентификации: интеграционные и unit-сценарии testframework.
import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/auth"
)

func newLoginTestEnv(t *testing.T) (*auth.LoginService, *ent.Client, *auth.MagicLinkService) {
	t.Helper()
	dsn := "file:logintest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	magic := auth.NewMagicLinkService(rdb, nil, "test@example.com")
	repo := auth.NewEntRepository(client)
	sessions := auth.NewSessionService(repo, rdb, nil, "login-jwt-secret", 15*time.Minute, 30*24*time.Hour)
	regMode := auth.NewRegistrationModeService(repo, nil)
	org := auth.NewOrganizationService(repo)
	login := auth.NewLoginService(repo, magic, sessions, regMode, org)
	return login, client, magic
}

// TestLoginWithMagicLink_PendingUserDenied — A1-001 / ORG-06: pending не получает JWT до admin approve.
func TestLoginWithMagicLink_PendingUserDenied(t *testing.T) {
	login, client, magic := newLoginTestEnv(t)
	ctx := context.Background()
	email := "pending-" + uuid.New().String()[:8] + "@example.com"

	_, err := client.User.Create().
		SetEmail(email).
		SetStatus(user.StatusPending).
		Save(ctx)
	require.NoError(t, err)

	code, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)

	pair, err := login.LoginWithMagicLink(ctx, email, code, "", "browser", "127.0.0.1")
	require.ErrorIs(t, err, auth.ErrAccountNotActive)
	assert.Nil(t, pair)

	sessions, err := client.Session.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, sessions)
}

// TestLoginWithMagicLink_Success проверяет сценарий: вход with magic link успех.
func TestLoginWithMagicLink_Success(t *testing.T) {
	login, client, magic := newLoginTestEnv(t)
	ctx := context.Background()
	email := "login-" + uuid.New().String()[:8] + "@example.com"

	code, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)

	pair, err := login.LoginWithMagicLink(ctx, email, code, "", "browser", "127.0.0.1")
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)
	require.NotEmpty(t, pair.RefreshToken)

	count, err := client.User.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestLoginWithMagicLink_InvalidCode проверяет отклонение при вход with magic link некорректный code.
func TestLoginWithMagicLink_InvalidCode(t *testing.T) {
	login, _, _ := newLoginTestEnv(t)
	ctx := context.Background()
	email := "bad-" + uuid.New().String()[:8] + "@example.com"

	_, err := login.LoginWithMagicLink(ctx, email, "000000", "", "browser", "127.0.0.1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, auth.ErrMagicLinkInvalid))
}

// TestLoginWithMagicLink_ExistingUser проверяет сценарий: вход with magic link existing user.
func TestLoginWithMagicLink_ExistingUser(t *testing.T) {
	login, client, magic := newLoginTestEnv(t)
	ctx := context.Background()
	email := "existing-" + uuid.New().String()[:8] + "@example.com"

	_, err := client.User.Create().SetEmail(email).Save(ctx)
	require.NoError(t, err)

	code, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)

	pair, err := login.LoginWithMagicLink(ctx, email, code, "", "mobile", "10.0.0.2")
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)

	count, err := client.User.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestUserExistsByEmail проверяет сценарий: user exists by email.
func TestUserExistsByEmail(t *testing.T) {
	login, client, _ := newLoginTestEnv(t)
	ctx := context.Background()
	email := "exists-" + uuid.New().String()[:8] + "@example.com"

	ok, err := login.UserExistsByEmail(ctx, email)
	require.NoError(t, err)
	assert.False(t, ok)

	_, err = client.User.Create().SetEmail(email).Save(ctx)
	require.NoError(t, err)

	ok, err = login.UserExistsByEmail(ctx, email)
	require.NoError(t, err)
	assert.True(t, ok)
}
