package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResendEmailClient(t *testing.T) {
	c := NewResendEmailClient("key", "hub@example.com")
	require.NotNil(t, c)
	assert.Equal(t, "key", c.apiKey)
	assert.Equal(t, "hub@example.com", c.fromEmail)
}

func TestResendEmailClient_SendMagicLinkEmail_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := NewResendEmailClient("test-key", "hub@example.com")
	client.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Scheme = "http"
			req.URL.Host = srv.Listener.Addr().String()
			req.URL.Path = "/emails"
			return http.DefaultTransport.RoundTrip(req)
		}),
	}

	err := client.SendMagicLinkEmail(context.Background(), "user@example.com", "123456")
	require.NoError(t, err)
}

func TestResendEmailClient_SendMagicLinkEmail_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	client := NewResendEmailClient("test-key", "hub@example.com")
	client.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Scheme = "http"
			req.URL.Host = srv.Listener.Addr().String()
			req.URL.Path = "/emails"
			return http.DefaultTransport.RoundTrip(req)
		}),
	}

	err := client.SendMagicLinkEmail(context.Background(), "user@example.com", "123456")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestTelegramBotAuth_VerifyCode_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	bot := &TelegramBotAuth{redis: rdb}
	require.NoError(t, mr.Set(tgCodePrefix+"654321", "42|Alice|Smith|alice"))

	user, err := bot.VerifyCode(context.Background(), "654321", "")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, int64(42), user.ID)
	assert.Equal(t, "Alice", user.FirstName)
}

func TestTelegramBotAuth_VerifyCode_InvalidCode(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	bot := &TelegramBotAuth{redis: rdb}
	_, err = bot.VerifyCode(context.Background(), "000000", "10.0.0.1")
	require.Error(t, err)
}

func TestTelegramBotAuth_VerifyCode_RateLimited(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	clientKey := "10.0.0.2"
	require.NoError(t, mr.Set(tgFailsPrefix+clientKey, "5"))

	bot := &TelegramBotAuth{redis: rdb}
	_, err = bot.VerifyCode(context.Background(), "111111", clientKey)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTelegramVerifyLimited)
}

func TestGenerateTGCode(t *testing.T) {
	code, err := generateTGCode()
	require.NoError(t, err)
	assert.Len(t, code, 6)
}

func TestNewTelegramBotAuth_EmptyToken(t *testing.T) {
	assert.Nil(t, NewTelegramBotAuth("", nil))
}

func TestTelegramBotAuth_RecordVerifyFailure(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	bot := &TelegramBotAuth{redis: rdb}
	clientKey := "10.0.0.9"
	bot.recordVerifyFailure(context.Background(), tgFailsPrefix+clientKey)
	bot.recordVerifyFailure(context.Background(), tgFailsPrefix+clientKey)

	val, err := mr.Get(tgFailsPrefix + clientKey)
	require.NoError(t, err)
	assert.Equal(t, "2", val)
}

func TestTelegramBotAuth_VerifyCode_CorruptData(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	bot := &TelegramBotAuth{redis: rdb}
	require.NoError(t, mr.Set(tgCodePrefix+"111111", "bad|data"))
	_, err = bot.VerifyCode(context.Background(), "111111", "1.2.3.4")
	require.Error(t, err)

	require.NoError(t, mr.Set(tgCodePrefix+"222222", "notint|a|b|c"))
	_, err = bot.VerifyCode(context.Background(), "222222", "1.2.3.4")
	require.Error(t, err)
}

func TestTelegramBotAuth_VerifyCode_EmptyClientKeySkipsRateLimit(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	require.NoError(t, mr.Set(tgFailsPrefix+"blocked", "99"))
	bot := &TelegramBotAuth{redis: rdb}
	_, err = bot.VerifyCode(context.Background(), "000000", "")
	require.Error(t, err)
}
