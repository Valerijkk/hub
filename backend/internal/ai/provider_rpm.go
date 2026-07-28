package ai

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// ErrProviderRateLimited — превышен MaxRPM провайдера (AI-SPEC S3).
var ErrProviderRateLimited = fmt.Errorf("provider rate limit exceeded")

// ErrProviderRPMUnavailable — Redis RPM-проверка недоступна (fail-closed, A2W5-A2-007).
var ErrProviderRPMUnavailable = fmt.Errorf("provider rate limit check unavailable")

// ProviderRPMLimiter — requests/minute на провайдера.
type ProviderRPMLimiter interface {
	Allow(ctx context.Context, providerID uuid.UUID, maxRPM int) error
}

// RedisProviderRPMLimiter — Redis sliding window на провайдера.
type RedisProviderRPMLimiter struct {
	rdb *redis.Client
}

// NewRedisProviderRPMLimiter — Redis RPM-лимитер.
func NewRedisProviderRPMLimiter(rdb *redis.Client) *RedisProviderRPMLimiter {
	return &RedisProviderRPMLimiter{rdb: rdb}
}

// Allow возвращает ErrProviderRateLimited при превышении maxRPM за минуту.
// При ошибке Redis — ErrProviderRPMUnavailable (не Allow / fail-open).
func (l *RedisProviderRPMLimiter) Allow(ctx context.Context, providerID uuid.UUID, maxRPM int) error {
	if l == nil || l.rdb == nil || maxRPM <= 0 {
		return nil
	}
	key := "ai:provider:rpm:" + providerID.String()
	now := time.Now()
	windowStart := now.Add(-time.Minute).UnixMilli()
	nowMs := now.UnixMilli()

	pipe := l.rdb.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart-1, 10))
	// Member уникален: на Windows UnixNano может совпасть в burst.
	member := strconv.FormatInt(now.UnixNano(), 10) + ":" + uuid.New().String()
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(nowMs), Member: member})
	card := pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, 2*time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		// A2W5-A2-007: Redis err → fail-closed, не return nil.
		return fmt.Errorf("%w: %v", ErrProviderRPMUnavailable, err)
	}
	// Rejected не оставляем в ZSET — иначе окно раздувается.
	if card.Val() > int64(maxRPM) {
		_ = l.rdb.ZRem(ctx, key, member).Err()
		return ErrProviderRateLimited
	}
	return nil
}

// NoopProviderRPMLimiter пропускает RPM (tests / no Redis).
type NoopProviderRPMLimiter struct{}

func (NoopProviderRPMLimiter) Allow(context.Context, uuid.UUID, int) error { return nil }
