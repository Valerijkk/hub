package auth_test

// Тесты сервиса аутентификации: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
	"context"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthService_FacadeTelegramDelegates проверяет бизнес-логику: auth сервис facade Telegram delegates.
func TestAuthService_FacadeTelegramDelegates(t *testing.T) {
	_, svc, client, _ := newHandlerTestEnv(t)
	ctx := context.Background()

	_, err := client.SystemSetting.Create().
		SetKey("registration_mode").
		SetValue("open").
		SetDescription("test").
		Save(ctx)
	require.NoError(t, err)

	tgUser := &auth.TelegramUser{ID: 424242424, FirstName: "Facade", Username: "facade_user"}
	pair, err := svc.LoginWithTelegramUser(ctx, tgUser, "tg-app", "10.0.0.1")
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
}

// TestAuthService_FacadeLoginWithTelegram проверяет бизнес-логику: auth сервис facade вход with Telegram.
func TestAuthService_FacadeLoginWithTelegram(t *testing.T) {
	_, svc, client, _ := newHandlerTestEnv(t)
	ctx := context.Background()

	_, err := client.SystemSetting.Create().
		SetKey("registration_mode").
		SetValue("open").
		SetDescription("test").
		Save(ctx)
	require.NoError(t, err)

	authDate := strconv.FormatInt(time.Now().Unix(), 10)
	userJSON := `{"id":777888999,"first_name":"ViaFacade"}`
	params := url.Values{}
	params.Set("auth_date", authDate)
	params.Set("user", userJSON)
	params.Set("query_id", "q-facade")
	params.Set("hash", computeTelegramHash(t, "bot", params))

	pair, err := svc.LoginWithTelegram(ctx, params.Encode(), "tg-web", "127.0.0.1")
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
}
