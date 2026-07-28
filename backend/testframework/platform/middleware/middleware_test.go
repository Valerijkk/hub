package middleware_test
import (
	"github.com/valerii/hub/backend/internal/platform/middleware"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Тесты middleware.RequestID
// ---------------------------------------------------------------------------

func TestRequestID_GeneratesNewID(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(c.Locals("requestID").(string))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	id := string(body)

	// Должен быть валидным UUID v4
	_, parseErr := uuid.Parse(id)
	assert.NoError(t, parseErr, "generated request ID should be a valid UUID")

	// Заголовок ответа должен совпадать
	assert.Equal(t, id, resp.Header.Get("X-Request-ID"))
}

func TestRequestID_PreservesExistingHeader(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(c.Locals("requestID").(string))
	})

	customID := "my-custom-request-id-123"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", customID)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, customID, string(body))
	assert.Equal(t, customID, resp.Header.Get("X-Request-ID"))
}

func TestRequestID_UniquePerRequest(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(c.Locals("requestID").(string))
	})

	ids := make(map[string]bool)
	for i := 0; i < 50; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		ids[string(body)] = true
	}
	assert.Equal(t, 50, len(ids), "all generated IDs must be unique")
}

// ---------------------------------------------------------------------------
// Тесты middleware.ErrorHandler
// ---------------------------------------------------------------------------

func decodeTZError(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	_, hasCode := body["code"]
	_, hasMessage := body["message"]
	assert.False(t, hasCode, "TZ forbids JSON field code")
	assert.False(t, hasMessage, "TZ forbids JSON field message")
	assert.Contains(t, body, "error")
	assert.Contains(t, body, "request_id")
	assert.Len(t, body, 2, "ErrorHandler must emit only error+request_id")
	return body
}

func TestErrorHandler_AppError(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(middleware.RequestID())
	app.Get("/", func(c fiber.Ctx) error {
		return apperrors.NewBadRequest("invalid input")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)
	body := decodeTZError(t, resp)
	assert.Equal(t, "invalid input", body["error"])
	assert.NotEmpty(t, body["request_id"], "request_id must be populated")
}

func TestErrorHandler_FiberError(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(middleware.RequestID())
	app.Get("/", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "page not found")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 404, resp.StatusCode)
	body := decodeTZError(t, resp)
	assert.Equal(t, "page not found", body["error"])
}

func TestErrorHandler_GenericError(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(middleware.RequestID())
	app.Get("/", func(c fiber.Ctx) error {
		return errors.New("something broke")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 500, resp.StatusCode)
	body := decodeTZError(t, resp)
	assert.Equal(t, "internal server error", body["error"])
	assert.NotContains(t, body["error"], "something broke")
}

func TestErrorHandler_ForbiddenNoInternals(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(middleware.RequestID())
	app.Get("/", func(c fiber.Ctx) error {
		return apperrors.NewForbidden("нет доступа к каналу")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 403, resp.StatusCode)
	body := decodeTZError(t, resp)
	assert.Equal(t, "нет доступа к каналу", body["error"])
	errStr, _ := body["error"].(string)
	assert.NotContains(t, errStr, "sql:")
	assert.NotContains(t, errStr, "pq:")
	assert.NotContains(t, errStr, "ent:")
}

// ---------------------------------------------------------------------------
// Тесты middleware.CORS
// ---------------------------------------------------------------------------

func TestCORS_DefaultOrigins(t *testing.T) {
	// Убеждаемся, что переменная окружения не задана
	os.Unsetenv("CORS_ORIGINS")
	t.Setenv("DEV_MODE", "true") // HI-11: prod-гард fail-start вызывает os.Exit

	app := fiber.New()
	app.Use(middleware.CORS())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Preflight с разрешённым origin по умолчанию
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, "http://localhost:5173", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORS_CustomOriginsFromEnv(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "https://hub.example.com,https://admin.example.com")

	app := fiber.New()
	app.Use(middleware.CORS())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://hub.example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, "https://hub.example.com", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	os.Unsetenv("CORS_ORIGINS")
	t.Setenv("DEV_MODE", "true") // HI-11: prod-гард fail-start вызывает os.Exit

	app := fiber.New()
	app.Use(middleware.CORS())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	// Не должен возвращать allow-origin заголовок для запрещённых origin
	assert.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORS_AllowsCredentials(t *testing.T) {
	os.Unsetenv("CORS_ORIGINS")
	t.Setenv("DEV_MODE", "true") // HI-11: prod-гард fail-start вызывает os.Exit

	app := fiber.New()
	app.Use(middleware.CORS())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, "true", resp.Header.Get("Access-Control-Allow-Credentials"))
}

func TestParseCORSOrigins_TrimsAndRejectsWildcardAndEmpty(t *testing.T) {
	cases := []struct {
		name        string
		env         string
		wantOrigins []string
		wantProbs   int
	}{
		{"empty env", "", nil, 0},
		{"spaces", " http://a , http://b ", []string{"http://a", "http://b"}, 0},
		{"wildcard rejected", "http://a,*", []string{"http://a"}, 1},
		{"empty entries", "http://a,, ,http://b", []string{"http://a", "http://b"}, 2},
		{"all bad", "*, ,,", nil, 4},
		// W4-A1-011 / P3-BE-11: shape — только абсолютный origin без path/query.
		{"null string", "null", nil, 1},
		{"path rejected", "https://hub.example.com/app", nil, 1},
		{"query rejected", "https://hub.example.com?x=1", nil, 1},
		{"ftp rejected", "ftp://hub.example.com", nil, 1},
		{"mixed valid invalid", "https://ok.example,null,https://hub.example.com/path", []string{"https://ok.example"}, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			origins, probs := middleware.ExportParseCORSOrigins(tc.env)
			assert.Equal(t, tc.wantOrigins, origins)
			assert.Len(t, probs, tc.wantProbs)
		})
	}
}

// ---------------------------------------------------------------------------
// Тесты middleware.SecurityHeaders
// ---------------------------------------------------------------------------

func TestSecurityHeaders_SetsExpectedHeaders(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.SecurityHeaders())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
	assert.Equal(t, "strict-origin-when-cross-origin", resp.Header.Get("Referrer-Policy"))
	assert.Equal(t, "default-src 'self'", resp.Header.Get("Content-Security-Policy"))
}

// ---------------------------------------------------------------------------
// Константы конфигурации middleware.RateLimit
// ---------------------------------------------------------------------------

func TestRateLimitConfig_Presets(t *testing.T) {
	assert.Equal(t, 3, middleware.AuthRateLimit.Max)
	assert.Equal(t, 10*time.Minute, middleware.AuthRateLimit.Window)

	assert.Equal(t, 100, middleware.APIRateLimit.Max)
	assert.Equal(t, 1*time.Second, middleware.APIRateLimit.Window)

	assert.Equal(t, 10, middleware.FileUploadRateLimit.Max)
	assert.Equal(t, 1*time.Minute, middleware.FileUploadRateLimit.Window)

	assert.Equal(t, 60, middleware.MessageSendRateLimit.Max)
	assert.Equal(t, 1*time.Minute, middleware.MessageSendRateLimit.Window)
}

// ---------------------------------------------------------------------------
// rateLimiter.allow (внутренняя логика, тестируется юнит-тестами)
// ---------------------------------------------------------------------------

func TestRateLimiter_Allow_WithinLimit(t *testing.T) {
	rl := middleware.ExportNewRateLimiterForTest(middleware.RateLimitConfig{Max: 5, Window: time.Minute})

	for i := 0; i < 5; i++ {
		assert.True(t, middleware.ExportRateLimiterAllow(rl, "key1"), "request %d should be allowed", i+1)
	}
}

func TestRateLimiter_Allow_ExceedsLimit(t *testing.T) {
	rl := middleware.ExportNewRateLimiterForTest(middleware.RateLimitConfig{Max: 3, Window: time.Minute})

	for i := 0; i < 3; i++ {
		assert.True(t, middleware.ExportRateLimiterAllow(rl, "key1"))
	}
	assert.False(t, middleware.ExportRateLimiterAllow(rl, "key1"), "4th request should be blocked")
}

func TestRateLimiter_Allow_SeparateKeys(t *testing.T) {
	rl := middleware.ExportNewRateLimiterForTest(middleware.RateLimitConfig{Max: 2, Window: time.Minute})

	assert.True(t, middleware.ExportRateLimiterAllow(rl, "a"))
	assert.True(t, middleware.ExportRateLimiterAllow(rl, "a"))
	assert.False(t, middleware.ExportRateLimiterAllow(rl, "a"))

	// Другой ключ должен иметь собственный счётчик
	assert.True(t, middleware.ExportRateLimiterAllow(rl, "b"))
	assert.True(t, middleware.ExportRateLimiterAllow(rl, "b"))
	assert.False(t, middleware.ExportRateLimiterAllow(rl, "b"))
}

func TestRateLimiter_Allow_WindowExpiry(t *testing.T) {
	rl := middleware.ExportNewRateLimiterForTest(middleware.RateLimitConfig{Max: 2, Window: 50 * time.Millisecond})

	assert.True(t, middleware.ExportRateLimiterAllow(rl, "k"))
	assert.True(t, middleware.ExportRateLimiterAllow(rl, "k"))
	assert.False(t, middleware.ExportRateLimiterAllow(rl, "k"))

	// Ждём истечения окна
	time.Sleep(60 * time.Millisecond)

	assert.True(t, middleware.ExportRateLimiterAllow(rl, "k"), "should allow after window expires")
}

// ---------------------------------------------------------------------------
// middleware.RateLimit (интеграция через тестовое приложение Fiber)
// ---------------------------------------------------------------------------

func TestRateLimit_Middleware_Blocks(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(middleware.RateLimit(t.Context(), middleware.RateLimitConfig{Max: 2, Window: time.Minute}))
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

	// 3-й запрос должен быть ограничен по частоте
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 429, resp.StatusCode)

	body := decodeTZError(t, resp)
	assert.Contains(t, body["error"], "too many requests")
}

func TestRateLimitByUser_MessageSendRateLimit_Allows60Blocks61(t *testing.T) {
	userID := uuid.New().String()

	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	})
	app.Use("/messenger", middleware.RateLimitByUser(t.Context(), middleware.MessageSendRateLimit))
	app.Post("/messenger/channels/:channelId/messages", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusCreated)
	})

	for i := 0; i < middleware.MessageSendRateLimit.Max; i++ {
		req := httptest.NewRequest(http.MethodPost, "/messenger/channels/"+uuid.New().String()+"/messages", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode, "request %d should pass", i+1)
	}

	req := httptest.NewRequest(http.MethodPost, "/messenger/channels/"+uuid.New().String()+"/messages", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 429, resp.StatusCode, "61st message in window should be rate-limited")
	_ = decodeTZError(t, resp)
}

func TestRateLimitByUser_MessageSendRateLimit_PerUserIsolation(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(func(c fiber.Ctx) error {
		if uid := c.Get("X-Test-User"); uid != "" {
			c.Locals("userID", uid)
		}
		return c.Next()
	})
	app.Use("/messenger", middleware.RateLimitByUser(t.Context(), middleware.RateLimitConfig{Max: 2, Window: time.Minute}))
	app.Post("/messenger/channels/x/messages", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusCreated)
	})

	exhaust := func(user string) {
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodPost, "/messenger/channels/x/messages", nil)
			req.Header.Set("X-Test-User", user)
			resp, err := app.Test(req)
			require.NoError(t, err)
			resp.Body.Close()
			assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
		}
		req := httptest.NewRequest(http.MethodPost, "/messenger/channels/x/messages", nil)
		req.Header.Set("X-Test-User", user)
		resp, err := app.Test(req)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, 429, resp.StatusCode)
	}

	exhaust("user-a")
	exhaust("user-b")

	req := httptest.NewRequest(http.MethodPost, "/messenger/channels/x/messages", nil)
	req.Header.Set("X-Test-User", "user-c")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode, "unaffected user should still be allowed")
}

// ---------------------------------------------------------------------------
// middleware.Logger (smoke-тест — просто проверяем, что нет паники)
// ---------------------------------------------------------------------------

func TestLogger_DoesNotPanic(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("hello")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "hello", string(body))
}

func TestLogger_PropagatesError(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger())
	app.Get("/", func(c fiber.Ctx) error {
		return apperrors.NewNotFound("nope")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 404, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// stripQuery (ME-03) — обрезка query-строки
// ---------------------------------------------------------------------------

func TestStripQuery(t *testing.T) {
	assert.Equal(t, "/ws", middleware.ExportStripQuery("/ws"))
	assert.Equal(t, "/ws", middleware.ExportStripQuery("/ws?token=secret-jwt"))
	assert.Equal(t, "/api/search", middleware.ExportStripQuery("/api/search?q=hello&limit=10"))
}

// ---------------------------------------------------------------------------
// isDevMode — определение dev-режима
// ---------------------------------------------------------------------------

func TestIsDevMode(t *testing.T) {
	t.Setenv("DEV_MODE", "true")
	assert.True(t, middleware.ExportIsDevMode())

	t.Setenv("DEV_MODE", "TRUE")
	assert.True(t, middleware.ExportIsDevMode())

	t.Setenv("DEV_MODE", "1")
	assert.True(t, middleware.ExportIsDevMode())

	t.Setenv("DEV_MODE", "false")
	assert.False(t, middleware.ExportIsDevMode())

	t.Setenv("DEV_MODE", "")
	assert.False(t, middleware.ExportIsDevMode())
}

// ---------------------------------------------------------------------------
// normalizeEmail + middleware.ExtractEmailFromBody — нормализация email
// ---------------------------------------------------------------------------

func TestNormalizeEmail(t *testing.T) {
	assert.Equal(t, "user@example.com", middleware.ExportNormalizeEmail(" User@Example.COM "))
	assert.Equal(t, "a@b.c", middleware.ExportNormalizeEmail("a\n@\tb.c"))
}

func TestExtractEmailFromBody(t *testing.T) {
	extractor := middleware.ExtractEmailFromBody("email")

	app := fiber.New()
	app.Post("/", func(c fiber.Ctx) error {
		key := extractor(c)
		return c.SendString(key)
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"email":" Admin@Hub.Local "}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "email:admin@hub.local", string(body))
}

func TestExtractEmailFromBody_EmptyBody(t *testing.T) {
	extractor := middleware.ExtractEmailFromBody("email")
	app := fiber.New()
	app.Post("/", func(c fiber.Ctx) error {
		return c.SendString(extractor(c))
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Empty(t, string(readBody(resp)))
}

func readBody(resp *http.Response) []byte {
	body, _ := io.ReadAll(resp.Body)
	return body
}

// ---------------------------------------------------------------------------
// hashIP — хеширование IP
// ---------------------------------------------------------------------------

func TestHashIP(t *testing.T) {
	assert.Equal(t, "unknown", middleware.ExportHashIP(""))
	h1 := middleware.ExportHashIP("127.0.0.1")
	h2 := middleware.ExportHashIP("127.0.0.1")
	h3 := middleware.ExportHashIP("192.168.0.1")
	assert.Equal(t, h1, h2)
	assert.NotEqual(t, h1, h3)
	assert.Len(t, h1, 16)
}

// ---------------------------------------------------------------------------
// middleware.RateLimitByUser — откат на IP
// ---------------------------------------------------------------------------

func TestRateLimitByUser_FallsBackToIP(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(middleware.RateLimitByUser(t.Context(), middleware.RateLimitConfig{Max: 1, Window: time.Minute}))
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	resp1, err := app.Test(req1)
	require.NoError(t, err)
	resp1.Body.Close()
	assert.Equal(t, 200, resp1.StatusCode)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, 429, resp2.StatusCode)
}

// ---------------------------------------------------------------------------
// middleware.RateLimitRedis — откат при nil-клиенте
// ---------------------------------------------------------------------------

func TestRateLimitRedis_NilClientFallsBackToInMemory(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(middleware.RateLimitRedis(t.Context(), nil, middleware.RedisRateLimitConfig{Max: 1, Window: time.Minute, KeyPrefix: "rl:test:"}, nil))
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	resp1, err := app.Test(req1)
	require.NoError(t, err)
	resp1.Body.Close()
	assert.Equal(t, 200, resp1.StatusCode)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, 429, resp2.StatusCode)
}

func TestRateLimitRedis_EnforcesLimitWithMiniredis(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	fixedClient := func(c fiber.Ctx) string { return c.Get("X-Client-Key") }

	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(middleware.RateLimitRedis(t.Context(), rdb, middleware.RedisRateLimitConfig{
		Max: 2, Window: time.Minute, KeyPrefix: "rl:redis:",
	}, fixedClient))
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Client-Key", "same-client")
		resp, err := app.Test(req)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode, "request %d", i+1)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Client-Key", "same-client")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 429, resp.StatusCode)
}

func TestRateLimitRedis_RedisErrorFallsBackToInMemory(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	clientKey := func(c fiber.Ctx) string { return c.Get("X-Client-Key") }

	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(middleware.RateLimitRedis(t.Context(), rdb, middleware.RedisRateLimitConfig{
		Max: 1, Window: time.Minute, KeyPrefix: "rl:failover:",
	}, clientKey))
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Set("X-Client-Key", "client-x")
	resp1, err := app.Test(req1)
	require.NoError(t, err)
	resp1.Body.Close()
	assert.Equal(t, 200, resp1.StatusCode)

	mr.Close()
	_ = rdb.Close()

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Client-Key", "client-x")
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	resp2.Body.Close()
	assert.Equal(t, 200, resp2.StatusCode, "first in-memory attempt after Redis outage")

	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.Header.Set("X-Client-Key", "client-x")
	resp3, err := app.Test(req3)
	require.NoError(t, err)
	defer resp3.Body.Close()
	assert.Equal(t, 429, resp3.StatusCode, "in-memory fallback must still enforce limit")
}

func TestRateLimitRedis_UsesExtractKey(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	extract := func(c fiber.Ctx) string {
		return c.Get("X-Client-Key")
	}

	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(middleware.RateLimitRedis(t.Context(), rdb, middleware.RedisRateLimitConfig{
		Max: 1, Window: time.Minute, KeyPrefix: "rl:key:",
	}, extract))
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Set("X-Client-Key", "client-a")
	resp1, err := app.Test(req1)
	require.NoError(t, err)
	resp1.Body.Close()
	assert.Equal(t, 200, resp1.StatusCode)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Client-Key", "client-a")
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, 429, resp2.StatusCode)

	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.Header.Set("X-Client-Key", "client-b")
	resp3, err := app.Test(req3)
	require.NoError(t, err)
	defer resp3.Body.Close()
	assert.Equal(t, 200, resp3.StatusCode)
}

func TestCORS_AllowsConfiguredOrigin(t *testing.T) {
	t.Setenv("DEV_MODE", "true")
	t.Setenv("CORS_ORIGINS", "https://app.example.com")
	app := fiber.New()
	app.Use(middleware.CORS())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://app.example.com")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "https://app.example.com", resp.Header.Get("Access-Control-Allow-Origin"))
}

// ---------------------------------------------------------------------------
// очистка rateLimiter останавливается при отмене контекста
// ---------------------------------------------------------------------------

func TestRateLimiter_CleanupStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	rl := middleware.ExportNewRateLimiterForTest(middleware.RateLimitConfig{Max: 10, Window: time.Minute})
	done := make(chan struct{})
	go func() {
		middleware.ExportRateLimiterCleanup(rl, ctx)
		close(done)
	}()

	middleware.ExportRateLimiterAllow(rl, "stale-key")
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("cleanup goroutine did not exit after context cancel")
	}
}

