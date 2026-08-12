//go:build hubtest

package middleware

import (
	"context"
	"time"
)

// ExportParseCORSOrigins wraps parseCORSOrigins for hubtest.
func ExportParseCORSOrigins(env string) ([]string, []string) {
	return parseCORSOrigins(env)
}

// ExportStripQuery wraps stripQuery for hubtest.
func ExportStripQuery(path string) string {
	return stripQuery(path)
}

// ExportIsDevMode wraps isDevMode for hubtest.
func ExportIsDevMode() bool {
	return isDevMode()
}

// ExportNormalizeEmail wraps normalizeEmail for hubtest.
func ExportNormalizeEmail(s string) string {
	return normalizeEmail(s)
}

// ExportHashIP wraps hashIP for hubtest.
func ExportHashIP(ip string) string {
	return hashIP(ip)
}

// ExportNewRateLimiterForTest creates a rateLimiter for hubtest.
func ExportNewRateLimiterForTest(cfg RateLimitConfig) *rateLimiter {
	return &rateLimiter{config: cfg}
}

// ExportRateLimiterAllow wraps rateLimiter.allow for hubtest.
func ExportRateLimiterAllow(rl *rateLimiter, key string) bool {
	return rl.allow(key)
}

// ExportRateLimiterCleanup wraps rateLimiter.cleanup for hubtest.
func ExportRateLimiterCleanup(rl *rateLimiter, ctx context.Context) {
	rl.cleanup(ctx)
}

// ExportRateLimiterHasKey reports whether key is present in the in-memory map.
func ExportRateLimiterHasKey(rl *rateLimiter, key string) bool {
	_, ok := rl.entries.Load(key)
	return ok
}

// ExportRateLimiterCleanupPass runs one cleanup sweep (W4-A1-003).
func ExportRateLimiterCleanupPass(rl *rateLimiter) {
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
			rl.entries.CompareAndDelete(key, entry)
			return true
		}
		entry.mu.Unlock()
		return true
	})
}
