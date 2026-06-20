package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_SendMagicLink_RateLimited(t *testing.T) {
	h, _, magic := newHandlerTestEnv(t)
	ctx := t.Context()
	email := "ratelimit-" + uuid.New().String()[:6] + "@example.com"

	for i := 0; i < magicLinkRateMax; i++ {
		_, err := magic.SendMagicLink(ctx, email)
		require.NoError(t, err)
	}

	app := testApp(http.MethodPost, "/auth/magic-link", h.SendMagicLink)
	body, _ := json.Marshal(MagicLinkRequest{Email: email})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}

func TestHandler_LoginTelegramCode_InviteOnlyNewUser(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()
	require.NoError(t, svc.SetRegistrationMode(ctx, "invite_only"))

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	require.NoError(t, mr.Set(tgCodePrefix+"888999", "55|New|User|newuser"))
	h.tgBot = &TelegramBotAuth{redis: rdb}

	app := testApp(http.MethodPost, "/auth/telegram-code", h.LoginTelegramCode)
	body, _ := json.Marshal(map[string]string{"code": "888999"})
	req := httptest.NewRequest(http.MethodPost, "/auth/telegram-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestLoginWithMagicLink_InviteOnlyExistingUser(t *testing.T) {
	login, client, magic := newLoginTestEnv(t)
	ctx := context.Background()

	email := "inv-exist-" + uuid.New().String()[:8] + "@example.com"
	_, err := client.User.Create().SetEmail(email).Save(ctx)
	require.NoError(t, err)

	_, err = client.SystemSetting.Create().
		SetKey("registration_mode").
		SetValue("invite_only").
		SetDescription("test").
		Save(ctx)
	require.NoError(t, err)

	code, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)

	pair, err := login.LoginWithMagicLink(ctx, email, code, "browser", "127.0.0.1")
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)
}
