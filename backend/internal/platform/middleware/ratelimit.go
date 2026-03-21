package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"

	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// RateLimitConfig defines the rate limit parameters.
type RateLimitConfig struct {
	Max    int           // Maximum number of requests
	Window time.Duration // Time window for the limit
}

// Предопределённые конфигурации лимитов.
var (
	AuthRateLimit        = RateLimitConfig{Max: 3, Window: 10 * time.Minute}
	APIRateLimit         = RateLimitConfig{Max: 100, Window: 1 * time.Second}
	FileUploadRateLimit  = RateLimitConfig{Max: 10, Window: 1 * time.Minute}
	MessageSendRateLimit = RateLimitConfig{Max: 60, Window: 1 * time.Minute}
)

type slidingWindowEntry struct {
	timestamps []time.Time
	mu         sync.Mutex
}

type rateLimiter struct {
	entries sync.Map
	config  RateLimitConfig
}

// RateLimit returns a rate limiting middleware using in-memory sliding window.
// Does not depend on Redis at startup.
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
