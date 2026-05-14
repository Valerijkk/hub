package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// RateLimitConfig определяет параметры ограничения частоты запросов.
type RateLimitConfig struct {
	Max    int           // Maximum number of requests
	Window time.Duration // Time window for the limit
}

// Предопределённые конфигурации лимитов.
var (
	AuthRateLimit        = RateLimitConfig{Max: 3, Window: 10 * time.Minute}
	APIRateLimit         = RateLimitConfig{Max: 100, Window: 1 * time.Second}
	FileUploadRateLimit  = RateLimitConfig{Max: 10, Window: 1 * time.Minute}
	MessageSendRateLimit = RateLimitConfig{Max: 300, Window: 1 * time.Minute}
)

type slidingWindowEntry struct {
	timestamps []time.Time
	mu         sync.Mutex
}

type rateLimiter struct {
	entries sync.Map
	config  RateLimitConfig
}

// RateLimit возвращает middleware ограничения частоты запросов на основе in-memory sliding window.
// Не зависит от Redis при старте.
func RateLimit(config RateLimitConfig) fiber.Handler {
	limiter := &rateLimiter{config: config}

	// Periodic cleanup of expired entries
	go limiter.cleanup()

	return func(c fiber.Ctx) error {
		key := c.IP()

		if !limiter.allow(key) {
			return apperrors.NewRateLimited("too many requests, please try again later")
		}

		return c.Next()
	}
}

// RateLimitByUser возвращает middleware ограничения частоты запросов по userID из Locals.
// Если userID не задан, фолбэк на IP-адрес клиента.
func RateLimitByUser(config RateLimitConfig) fiber.Handler {
	limiter := &rateLimiter{config: config}
	go limiter.cleanup()

	return func(c fiber.Ctx) error {
		key := ""
		if uid, ok := c.Locals("userID").(string); ok && uid != "" {
			key = "user:" + uid
		} else {
			key = "ip:" + c.IP()
		}

		if !limiter.allow(key) {
			return apperrors.NewRateLimited("too many requests, please try again later")
		}

		return c.Next()
	}
}

func (rl *rateLimiter) allow(key string) bool {
	now := time.Now()
	windowStart := now.Add(-rl.config.Window)

	val, _ := rl.entries.LoadOrStore(key, &slidingWindowEntry{})
	entry := val.(*slidingWindowEntry)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	// Remove expired timestamps
	valid := entry.timestamps[:0]
	for _, ts := range entry.timestamps {
		if ts.After(windowStart) {
			valid = append(valid, ts)
		}
	}
	entry.timestamps = valid

	if len(entry.timestamps) >= rl.config.Max {
		return false
	}

	entry.timestamps = append(entry.timestamps, now)
	return true
}

// ---------------------------------------------------------------------------
// HI-10: Redis-based sliding window rate limiter (cross-instance safe).
// ---------------------------------------------------------------------------
//
// Назначение: для чувствительных endpoint'ов (auth, magic-link verify) нужен
// общий счётчик, который не обходится перезапуском процесса или горизонтальным
// масштабированием. Также важно не доверять IP "как есть" — заголовки
// X-Forwarded-For легко подделываются, если Fiber не настроен с trusted-proxies.
//
// Алгоритм: ZSET с timestamp'ами попыток, ZREMRANGEBYSCORE отрезает старое
// окно, ZCARD считает текущие — классическая sliding-window реализация
// в Redis (3 команды в одном pipeline).

// RedisRateLimitConfig определяет параметры sliding-window лимитера на Redis.
type RedisRateLimitConfig struct {
	Max       int           // максимум запросов в окне
	Window    time.Duration // длительность окна
	KeyPrefix string        // префикс Redis-ключа (например, "rl:auth:")
}

// RateLimitKeyExtractor возвращает строку, идентифицирующую клиента, по
// которой идёт учёт. Возврат пустой строки означает «применить fallback»
// (как правило, к IP). Это позволяет, например, лимитировать по email,
// если он есть в body, и по IP в противном случае.
type RateLimitKeyExtractor func(c fiber.Ctx) string

// RateLimitRedis возвращает middleware sliding-window на Redis. extractKey
// должен вернуть стабильный идентификатор клиента; если он пуст, мы
// фолбэчимся на хеш IP. Все ключи Redis имеют общий префикс cfg.KeyPrefix.
//
// HI-10: использовать вместо in-memory RateLimit для auth, magic-link verify
// и других путей, где важна устойчивость к IP-rotation и горизонтальному
// масштабированию.
func RateLimitRedis(rdb *redis.Client, cfg RedisRateLimitConfig, extractKey RateLimitKeyExtractor) fiber.Handler {
	if rdb == nil {
		// Fallback на in-memory лимитер, если Redis не сконфигурирован.
		// В DevMode это удобно; в production config.validate должен
		// убедиться, что Redis есть.
		slog.Warn("RateLimitRedis: Redis client is nil, falling back to in-memory limiter",
			"prefix", cfg.KeyPrefix)
		return RateLimit(RateLimitConfig{Max: cfg.Max, Window: cfg.Window})
	}
	prefix := cfg.KeyPrefix
	if prefix == "" {
		prefix = "rl:"
	}

	return func(c fiber.Ctx) error {
		ctx := c.Context()

		raw := ""
		if extractKey != nil {
			raw = extractKey(c)
		}
		if raw == "" {
			raw = "ip:" + hashIP(c.IP())
		}
		key := prefix + raw

		now := time.Now()
		nowMs := now.UnixMilli()
		windowStart := now.Add(-cfg.Window).UnixMilli()

		pipe := rdb.Pipeline()
		// Удаляем устаревшие попытки.
		pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart-1, 10))
		// Добавляем текущую попытку. Member — уникальный nanosecond+IP, чтобы
		// одновременные запросы не схлопывались в один ZADD.
		member := strconv.FormatInt(now.UnixNano(), 10) + ":" + c.IP()
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(nowMs), Member: member})
		cardCmd := pipe.ZCard(ctx, key)
		// Гарантируем, что ключ не залипнет навсегда: TTL чуть больше окна.
		pipe.Expire(ctx, key, cfg.Window+time.Minute)

		if _, err := pipe.Exec(ctx); err != nil {
			// Fail-open на ошибке Redis — лучше пропустить запрос, чем
			// заDDoSить себя. Логируем, чтобы это было видно в проде.
			slog.Warn("RateLimitRedis: pipeline error, failing open",
				"key", key, "error", err)
			return c.Next()
		}

		if cardCmd.Val() > int64(cfg.Max) {
			return apperrors.NewRateLimited("too many requests, please try again later")
		}
		return c.Next()
	}
}

// ExtractEmailFromBody возвращает RateLimitKeyExtractor, вытаскивающий
// email из JSON body запроса. Используется для magic-link verify: лимит
// должен быть по email (это атакуемая сущность), не по IP.
//
// Парсинг идёт без c.Bind(), чтобы не нарушать ожидаемого хендлером
// состояния body buffer.
func ExtractEmailFromBody(field string) RateLimitKeyExtractor {
	if field == "" {
		field = "email"
	}
	return func(c fiber.Ctx) string {
		body := c.Body()
		if len(body) == 0 {
			return ""
		}
		var m map[string]any
		if err := json.Unmarshal(body, &m); err != nil {
			return ""
		}
		v, _ := m[field].(string)
		if v == "" {
			return ""
		}
		return "email:" + normalizeEmail(v)
	}
}

func normalizeEmail(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b >= 'A' && b <= 'Z' {
			b += 'a' - 'A'
		}
		if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
			continue
		}
		out = append(out, b)
	}
	return string(out)
}

// hashIP — sha256-первые-16-hex от IP. Чуть уменьшает PII в логах и Redis-
// ключах. Не зависит от того, что вернёт c.IP() — для нас важна
// стабильность fingerprint'а, не его «правильность»; если за прокси без
// EnableTrustedProxyCheck кто-то спуфит X-Forwarded-For, у каждой
// подмены будет свой бакет (что хуже, чем за прокси, но не хуже того,
// что было до HI-10).
func hashIP(ip string) string {
	if ip == "" {
		return "unknown"
	}
	sum := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(sum[:8])
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		windowStart := now.Add(-rl.config.Window)

		rl.entries.Range(func(key, value any) bool {
			entry := value.(*slidingWindowEntry)
			entry.mu.Lock()

			valid := entry.timestamps[:0]
			for _, ts := range entry.timestamps {
				if ts.After(windowStart) {
					valid = append(valid, ts)
				}
			}
			entry.timestamps = valid

			if len(entry.timestamps) == 0 {
				entry.mu.Unlock()
				rl.entries.Delete(key)
				return true
			}

			entry.mu.Unlock()
			return true
		})
	}
}
