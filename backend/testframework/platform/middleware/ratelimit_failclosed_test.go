//go:build hubtest

package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/middleware"
)

// TestRateLimitRedis_RejectedDoesNotInflateZSET — W4-A1-004 / P2-BE-10:
// 429 должен ZRem member; окно не раздувается отказами.
func TestRateLimitRedis_RejectedDoesNotInflateZSET(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	fixedClient := func(c fiber.Ctx) string { return "zset-client" }

	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(middleware.RateLimitRedis(t.Context(), rdb, middleware.RedisRateLimitConfig{
		Max: 2, Window: time.Minute, KeyPrefix: "rl:zset:",
	}, fixedClient))
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)
	}

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, 429, resp.StatusCode, "reject %d", i+1)
	}

	keys := mr.Keys()
	require.Len(t, keys, 1)
	card, err := rdb.ZCard(context.Background(), keys[0]).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(2), card, "ZSET must stay at Max after rejected bursts")
}

// TestRateLimiter_CleanupCompareAndDelete_PreservesLiveBucket — W4-A1-003:
// cleanup не удаляет bucket, в который параллельно успели дописать timestamps.
func TestRateLimiter_CleanupCompareAndDelete_PreservesLiveBucket(t *testing.T) {
	rl := middleware.ExportNewRateLimiterForTest(middleware.RateLimitConfig{
		Max: 10, Window: time.Millisecond, // короткое окно → timestamps сразу «просрочены»
	})
	key := "race-key"
	require.True(t, middleware.ExportRateLimiterAllow(rl, key))
	time.Sleep(2 * time.Millisecond) // timestamps вне окна

	// Параллельно: cleanup видит пустое окно; allow дописывает новые timestamps
	// до CompareAndDelete — живой pointer должен остаться в map.
	done := make(chan struct{})
	go func() {
		defer close(done)
		middleware.ExportRateLimiterCleanupPass(rl)
	}()
	require.True(t, middleware.ExportRateLimiterAllow(rl, key))
	<-done

	assert.True(t, middleware.ExportRateLimiterHasKey(rl, key),
		"CompareAndDelete must not drop a bucket that received new timestamps")
	assert.True(t, middleware.ExportRateLimiterAllow(rl, key))
}
