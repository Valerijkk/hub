package auth_test

// Тесты HTTP-обработчика аутентификации: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_SendMagicLink_RateLimited проверяет HTTP-слой: HTTP-обработчик отправка magic link rate limited.
func TestHandler_SendMagicLink_RateLimited(t *testing.T) {
	h, _, _, magic := newHandlerTestEnv(t)
	ctx := t.Context()
	email := "ratelimit-" + uuid.New().String()[:6] + "@example.com"

	for i := 0; i < auth.ExportMagicLinkRateMax(); i++ {
		_, err := magic.SendMagicLink(ctx, email)
		require.NoError(t, err)
	}

	app := testApp(http.MethodPost, "/auth/magic-link", h.SendMagicLink)
	body, _ := json.Marshal(auth.MagicLinkRequest{Email: email})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}

// TestLoginWithMagicLink_InviteOnlyExistingUser проверяет сценарий: вход with magic link приглашение only existing user.
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

	pair, err := login.LoginWithMagicLink(ctx, email, code, "", "browser", "127.0.0.1")
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)
}
