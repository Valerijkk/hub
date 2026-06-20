package auth

import (
	"context"
	"errors"
	"net/url"
	"strconv"
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
)

func newLoginTestEnv(t *testing.T) (*LoginService, *ent.Client, *MagicLinkService) {
	t.Helper()
	dsn := "file:logintest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	magic := NewMagicLinkService(rdb, nil, "test@example.com")
	sessions := NewSessionService(client, rdb, nil, "login-jwt-secret", 15*time.Minute, 30*24*time.Hour)
	regMode := NewRegistrationModeService(client, nil)
	org := NewOrganizationService(client)
	login := NewLoginService(client, magic, "telegram-bot-token", sessions, regMode, org)
	return login, client, magic
}

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

func TestLoginWithMagicLink_InvalidCode(t *testing.T) {
	login, _, _ := newLoginTestEnv(t)
	ctx := context.Background()
	email := "bad-" + uuid.New().String()[:8] + "@example.com"

	_, err := login.LoginWithMagicLink(ctx, email, "000000", "", "browser", "127.0.0.1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrMagicLinkInvalid))
}

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

func TestLoginWithTelegramUser_ExistingUser(t *testing.T) {
	login, client, _ := newLoginTestEnv(t)
	ctx := context.Background()
	tgID := "987654321"

	_, err := client.User.Create().
		SetEmail("telegram_" + tgID + "@hub.local").
		SetTelegramID(tgID).
		SetName("Existing TG").
		Save(ctx)
	require.NoError(t, err)

	pair, err := login.LoginWithTelegramUser(ctx, &TelegramUser{
		ID: 987654321, FirstName: "Existing", LastName: "User",
	}, "tg-app", "192.168.1.1")
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)
	require.NotEmpty(t, pair.RefreshToken)
}

func TestLoginWithTelegramUser_NewUser(t *testing.T) {
	login, client, _ := newLoginTestEnv(t)
	ctx := context.Background()
	_, err := client.SystemSetting.Create().
		SetKey("registration_mode").
		SetValue("open").
		SetDescription("test").
		Save(ctx)
	require.NoError(t, err)

	pair, err := login.LoginWithTelegramUser(ctx, &TelegramUser{
		ID: 111222333, FirstName: "New", Username: "newtg",
	}, "tg-app", "192.168.1.2")
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)

	count, err := client.User.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestLoginWithTelegramUser_InviteOnlyBlocked(t *testing.T) {
	login, client, _ := newLoginTestEnv(t)
	ctx := context.Background()

	_, err := client.SystemSetting.Create().
		SetKey("registration_mode").
		SetValue("invite_only").
		SetDescription("test").
		Save(ctx)
	require.NoError(t, err)

	_, err = login.LoginWithTelegramUser(ctx, &TelegramUser{
		ID: 444555666, FirstName: "Blocked",
	}, "tg-app", "10.0.0.3")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrRegistrationInviteOnly))
}

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

func TestLoginWithTelegram_ValidInitData(t *testing.T) {
	login, client, _ := newLoginTestEnv(t)
	ctx := context.Background()
	_, err := client.SystemSetting.Create().
		SetKey("registration_mode").
		SetValue("open").
		SetDescription("test").
		Save(ctx)
	require.NoError(t, err)

	botToken := "telegram-bot-token"

	authDate := strconv.FormatInt(time.Now().Unix(), 10)
	userJSON := `{"id":555666777,"first_name":"TG","last_name":"User"}`
	params := url.Values{}
	params.Set("auth_date", authDate)
	params.Set("user", userJSON)
	params.Set("query_id", "q1")
	params.Set("hash", computeTelegramHash(t, botToken, params))

	pair, err := login.LoginWithTelegram(ctx, params.Encode(), "tg-web", "10.0.0.5")
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)

	count, err := client.User.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
