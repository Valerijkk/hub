//go:build hubtest

package middleware

import "context"

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
