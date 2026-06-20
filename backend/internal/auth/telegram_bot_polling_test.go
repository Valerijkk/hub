package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/mymmrac/telego"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTelegramBotToken = "123456789:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc"

func writeTelegramOK(w http.ResponseWriter, result any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": result})
}

func newMockTelegramBot(t *testing.T, getUpdatesFn func(call int) any) (*telego.Bot, *httptest.Server) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "deleteWebhook"):
			writeTelegramOK(w, true)
		case strings.HasSuffix(path, "getUpdates"):
			n := int(calls.Add(1))
			writeTelegramOK(w, getUpdatesFn(n))
		case strings.HasSuffix(path, "sendMessage"):
			writeTelegramOK(w, map[string]any{"message_id": 1})
		default:
			writeTelegramOK(w, true)
		}
	}))
	t.Cleanup(srv.Close)

	bot, err := telego.NewBot(
		testTelegramBotToken,
		telego.WithAPIServer(srv.URL),
		telego.WithHTTPClient(srv.Client()),
		telego.WithDiscardLogger(),
	)
	require.NoError(t, err)
	return bot, srv
}

func TestTelegramBotAuth_HandleLogin_StoresCode(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	bot, _ := newMockTelegramBot(t, func(int) any { return []any{} })
	tgAuth := &TelegramBotAuth{bot: bot, redis: rdb}

	msg := &telego.Message{
		MessageID: 1,
		Chat:      telego.Chat{ID: 4242, Type: "private"},
		From:      &telego.User{ID: 777, FirstName: "Alice", LastName: "Smith", Username: "alice"},
		Text:      "/login",
	}
	tgAuth.handleLogin(context.Background(), msg)

	keys := mr.Keys()
	require.NotEmpty(t, keys)
	found := false
	for _, k := range keys {
		if strings.HasPrefix(k, tgCodePrefix) {
			found = true
			val, err := mr.Get(k)
			require.NoError(t, err)
			assert.Contains(t, val, "777|Alice|Smith|alice")
		}
	}
	assert.True(t, found, "expected tglogin code in redis")
}

func TestTelegramBotAuth_HandleLogin_ShutdownDuringSend(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	bot, _ := newMockTelegramBot(t, func(int) any { return []any{} })
	tgAuth := &TelegramBotAuth{bot: bot, redis: rdb}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	msg := &telego.Message{
		Chat: telego.Chat{ID: 1},
		From: &telego.User{ID: 2, FirstName: "Bob"},
		Text: "/start",
	}
	assert.NotPanics(t, func() {
		tgAuth.handleLogin(ctx, msg)
	})
}

func TestTelegramBotAuth_StartPolling_Disabled(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	bot, _ := newMockTelegramBot(t, func(int) any { return []any{} })
	tgAuth := &TelegramBotAuth{bot: bot, redis: rdb}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	assert.NotPanics(t, func() {
		tgAuth.StartPolling(ctx)
	})
	assert.Empty(t, mr.Keys(), "polling must not store login codes after AUTH-01")
}

func TestNewTelegramBotAuth_InvalidToken(t *testing.T) {
	assert.Nil(t, NewTelegramBotAuth("not-a-valid-token", nil))
}
