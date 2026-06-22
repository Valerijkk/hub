package auth_test

// Тесты модуля аутентификации: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewResendEmailClient проверяет сценарий: new resend email client.
func TestNewResendEmailClient(t *testing.T) {
	c := auth.NewResendEmailClient("key", "hub@example.com")
	require.NotNil(t, c)
	assert.Equal(t, "key", auth.ExportResendEmailClientAPIKey(c))
	assert.Equal(t, "hub@example.com", auth.ExportResendEmailClientFromEmail(c))
}

// TestResendEmailClient_SendMagicLinkEmail_Success проверяет сценарий: resend email client отправка magic link email успех.
func TestResendEmailClient_SendMagicLinkEmail_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := auth.NewResendEmailClient("test-key", "hub@example.com")
	auth.ExportResendEmailClientInjectHTTP(client, &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Scheme = "http"
			req.URL.Host = srv.Listener.Addr().String()
			req.URL.Path = "/emails"
			return http.DefaultTransport.RoundTrip(req)
		}),
	})

	err := client.SendMagicLinkEmail(context.Background(), "user@example.com", "123456")
	require.NoError(t, err)
}

// TestResendEmailClient_SendMagicLinkEmail_APIError проверяет сценарий: resend email client отправка magic link email api ошибка.
func TestResendEmailClient_SendMagicLinkEmail_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	client := auth.NewResendEmailClient("test-key", "hub@example.com")
	auth.ExportResendEmailClientInjectHTTP(client, &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Scheme = "http"
			req.URL.Host = srv.Listener.Addr().String()
			req.URL.Path = "/emails"
			return http.DefaultTransport.RoundTrip(req)
		}),
	})

	err := client.SendMagicLinkEmail(context.Background(), "user@example.com", "123456")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// TestTelegramBotAuth_VerifyCode_Success проверяет сценарий: Telegram бот auth verify code успех.
func TestTelegramBotAuth_VerifyCode_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	bot := auth.ExportNewTelegramBotAuthForTest(rdb)
	require.NoError(t, mr.Set(auth.ExportTGCodePrefix()+"654321", "42|Alice|Smith|alice"))

	user, err := bot.VerifyCode(context.Background(), "654321", "")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, int64(42), user.ID)
	assert.Equal(t, "Alice", user.FirstName)
}

// TestTelegramBotAuth_VerifyCode_InvalidCode проверяет отклонение при Telegram бот auth verify code некорректный code.
func TestTelegramBotAuth_VerifyCode_InvalidCode(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	bot := auth.ExportNewTelegramBotAuthForTest(rdb)
	_, err = bot.VerifyCode(context.Background(), "000000", "10.0.0.1")
	require.Error(t, err)
}

// TestTelegramBotAuth_VerifyCode_RateLimited проверяет сценарий: Telegram бот auth verify code rate limited.
func TestTelegramBotAuth_VerifyCode_RateLimited(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	clientKey := "10.0.0.2"
	require.NoError(t, mr.Set(auth.ExportTGFailsPrefix()+clientKey, "5"))

	bot := auth.ExportNewTelegramBotAuthForTest(rdb)
	_, err = bot.VerifyCode(context.Background(), "111111", clientKey)
	require.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrTelegramVerifyLimited)
}

// TestGenerateTGCode проверяет сценарий: generate tg code.
func TestGenerateTGCode(t *testing.T) {
	code, err := auth.ExportGenerateTGCode()
	require.NoError(t, err)
	assert.Len(t, code, 6)
}

// TestNewTelegramBotAuth_EmptyToken проверяет сценарий: new Telegram бот auth пустой token.
func TestNewTelegramBotAuth_EmptyToken(t *testing.T) {
	assert.Nil(t, auth.NewTelegramBotAuth("", nil))
}

// TestTelegramBotAuth_RecordVerifyFailure проверяет сценарий: Telegram бот auth record verify failure.
func TestTelegramBotAuth_RecordVerifyFailure(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	bot := auth.ExportNewTelegramBotAuthForTest(rdb)
	clientKey := "10.0.0.9"
	auth.ExportTelegramRecordVerifyFailure(bot, context.Background(), auth.ExportTGFailsPrefix()+clientKey)
	auth.ExportTelegramRecordVerifyFailure(bot, context.Background(), auth.ExportTGFailsPrefix()+clientKey)

	val, err := mr.Get(auth.ExportTGFailsPrefix() + clientKey)
	require.NoError(t, err)
	assert.Equal(t, "2", val)
}

// TestTelegramBotAuth_VerifyCode_CorruptData проверяет сценарий: Telegram бот auth verify code corrupt data.
func TestTelegramBotAuth_VerifyCode_CorruptData(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	bot := auth.ExportNewTelegramBotAuthForTest(rdb)
	require.NoError(t, mr.Set(auth.ExportTGCodePrefix()+"111111", "bad|data"))
	_, err = bot.VerifyCode(context.Background(), "111111", "1.2.3.4")
	require.Error(t, err)

	require.NoError(t, mr.Set(auth.ExportTGCodePrefix()+"222222", "notint|a|b|c"))
	_, err = bot.VerifyCode(context.Background(), "222222", "1.2.3.4")
	require.Error(t, err)
}

// TestTelegramBotAuth_VerifyCode_EmptyClientKeySkipsRateLimit проверяет сценарий: Telegram бот auth verify code пустой client key skips rate limit.
func TestTelegramBotAuth_VerifyCode_EmptyClientKeySkipsRateLimit(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	require.NoError(t, mr.Set(auth.ExportTGFailsPrefix()+"blocked", "99"))
	bot := auth.ExportNewTelegramBotAuthForTest(rdb)
	_, err = bot.VerifyCode(context.Background(), "000000", "")
	require.Error(t, err)
}
