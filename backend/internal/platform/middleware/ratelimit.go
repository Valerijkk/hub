package middleware

import (
	"context"
	"crypto/rand"
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
	Max    int           // Максимальное число запросов
	Window time.Duration // Временное окно для лимита
}

// Предопределённые конфигурации лимитов.
var (
	AuthRateLimit        = RateLimitConfig{Max: 3, Window: 10 * time.Minute}
	APIRateLimit         = RateLimitConfig{Max: 100, Window: 1 * time.Second}
	FileUploadRateLimit  = RateLimitConfig{Max: 10, Window: 1 * time.Minute}
	MessageSendRateLimit = RateLimitConfig{Max: 60, Window: 1 * time.Minute}
	AIRateLimit          = RateLimitConfig{Max: 20, Window: 1 * time.Minute}
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
//
// ME-5: принимает context.Context для graceful shutdown cleanup-горутины.
// При cancel-е ctx (как правило rootCtx из main) cleanup корректно завершается;
// раньше горутина висла до завершения процесса.
//
// Для бэк-совместимости: nil-ctx эквивалентен context.Background() — старый
// бессрочный режим, на случай если кто-то вызовет конструктор из теста.
func RateLimit(ctx context.Context, config RateLimitConfig) fiber.Handler {
	if ctx == nil {
		ctx = context.Background()
	}
	limiter := &rateLimiter{config: config}

	// Периодическая очистка просроченных записей — иначе утечка памяти.
	go limiter.cleanup(ctx)

	return func(c fiber.Ctx) error {
		key := c.IP()

		if !limiter.allow(key) {
			sec := int(limiter.config.Window.Seconds())
			if sec < 1 {
				sec = 1
			}
			return apperrors.NewRateLimitedAfter("too many requests, please try again later", sec)
		}

		return c.Next()
	}
}

// RateLimitByUser возвращает middleware ограничения частоты запросов по userID из Locals.
// Если userID не задан, фолбэк на IP-адрес клиента.
//
// ME-5: см. RateLimit — ctx нужен для graceful shutdown.
func RateLimitByUser(ctx context.Context, config RateLimitConfig) fiber.Handler {
	if ctx == nil {
		ctx = context.Background()
	}
	limiter := &rateLimiter{config: config}
	go limiter.cleanup(ctx)

	return func(c fiber.Ctx) error {
		key := ""
		if uid, ok := c.Locals("userID").(string); ok && uid != "" {
			key = "user:" + uid
		} else {
			key = "ip:" + c.IP()
		}

		if !limiter.allow(key) {
			sec := int(limiter.config.Window.Seconds())
			if sec < 1 {
				sec = 1
			}
			return apperrors.NewRateLimitedAfter("too many requests, please try again later", sec)
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

	// Удаляем просроченные timestamps из sliding window.
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
// HI-10: sliding window rate limiter на Redis — безопасен между инстансами.
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
//
// ME-5: ctx нужен только для fallback-пути на in-memory RateLimit, у которого
// есть фоновая cleanup-горутина. Для Redis-пути ctx запроса берётся из
// fiber.Ctx.Context().
func RateLimitRedis(ctx context.Context, rdb *redis.Client, cfg RedisRateLimitConfig, extractKey RateLimitKeyExtractor) fiber.Handler {
	if rdb == nil {
		// Fallback на in-memory лимитер, если Redis не сконфигурирован.
		// В DevMode это удобно; в production config.validate должен
		// убедиться, что Redis есть.
		slog.Warn("RateLimitRedis: Redis client is nil, falling back to in-memory limiter",
			"prefix", cfg.KeyPrefix)
		return RateLimit(ctx, RateLimitConfig{Max: cfg.Max, Window: cfg.Window})
	}
	prefix := cfg.KeyPrefix
	if prefix == "" {
		prefix = "rl:"
	}

	// In-memory fallback при сбое Redis — иначе fail-open bypass (SEC).
	memFallback := &rateLimiter{config: RateLimitConfig{Max: cfg.Max, Window: cfg.Window}}
	go memFallback.cleanup(ctx)

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
		// Member уникален на запрос — nanosecond alone коллидирует в burst.
		member := strconv.FormatInt(now.UnixNano(), 10) + ":" + c.IP() + ":" + randomHex(8)
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(nowMs), Member: member})
		cardCmd := pipe.ZCard(ctx, key)
		// Гарантируем, что ключ не залипнет навсегда: TTL чуть больше окна.
		pipe.Expire(ctx, key, cfg.Window+time.Minute)

		if _, err := pipe.Exec(ctx); err != nil {
			slog.Warn("RateLimitRedis: pipeline error, falling back to in-memory limiter",
				"key", key, "error", err)
			memKey := raw
			if memKey == "" {
				memKey = "ip:" + hashIP(c.IP())
			}
			if !memFallback.allow(memKey) {
				sec := int(cfg.Window.Seconds())
				if sec < 1 {
					sec = 1
				}
				return apperrors.NewRateLimitedAfter("too many requests, please try again later", sec)
			}
			return c.Next()
		}

		// W4-A1-004: при 429 удаляем member — иначе отклонённые запросы раздувают ZSET.
		if cardCmd.Val() > int64(cfg.Max) {
			if remErr := rdb.ZRem(ctx, key, member).Err(); remErr != nil {
				slog.Warn("RateLimitRedis: ZRem after reject failed", "key", key, "error", remErr)
			}
			sec := int(cfg.Window.Seconds())
			if sec < 1 {
				sec = 1
			}
			return apperrors.NewRateLimitedAfter("too many requests, please try again later", sec)
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

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(b)
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

// cleanup периодически удаляет устаревшие записи sliding-window.
//
// ME-5: реагирует на ctx.Done() — при shutdown горутина завершается
// детерминированно (без 5-минутной задержки на следующий tick). Раньше
// горутина утекала: при graceful shutdown процесс ждал её до тайм-аута.
func (rl *rateLimiter) cleanup(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
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
					// W4-A1-003: CompareAndDelete — иначе параллельный allow мог
					// LoadOrStore тот же pointer, дописать timestamps, а Delete
					// выкинул бы живой bucket (сброс счётчика / bypass).
					rl.entries.CompareAndDelete(key, entry)
					return true
				}

				entry.mu.Unlock()
				return true
			})
		}
	}
}
