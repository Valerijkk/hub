package mail

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// DailySlotStore — shared счётчик daily_limit (Redis или in-memory для тестов).
// A3W7-1-001: process-local map не должен быть единственным источником истины.
type DailySlotStore interface {
	Reserve(ctx context.Context, email string, limit int) error
	Release(ctx context.Context, email string) error
}

// MemoryDailySlotStore — process-local fallback (тесты без Redis).
type MemoryDailySlotStore struct {
	mu       sync.Mutex
	dayKey   string
	perEmail map[string]int
}

// NewMemoryDailySlotStore создаёт in-memory счётчик.
func NewMemoryDailySlotStore() *MemoryDailySlotStore {
	return &MemoryDailySlotStore{perEmail: make(map[string]int)}
}

// Reserve резервирует слот в текущих сутках UTC.
func (m *MemoryDailySlotStore) Reserve(_ context.Context, email string, limit int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	today := time.Now().UTC().Format("2006-01-02")
	if m.dayKey != today {
		m.dayKey = today
		m.perEmail = make(map[string]int)
	}
	if m.perEmail[email] >= limit {
		return ErrDailyLimitExceeded
	}
	m.perEmail[email]++
	return nil
}

// Release откатывает слот после неуспешной отправки.
func (m *MemoryDailySlotStore) Release(_ context.Context, email string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.perEmail[email] > 0 {
		m.perEmail[email]--
	}
	return nil
}

// RedisDailySlotStore — INCR + EXPIREAT до UTC midnight (общий для N реплик).
type RedisDailySlotStore struct {
	rdb *redis.Client
}

// NewRedisDailySlotStore создаёт Redis-backed daily counter.
func NewRedisDailySlotStore(rdb *redis.Client) *RedisDailySlotStore {
	return &RedisDailySlotStore{rdb: rdb}
}

const dailySlotKeyPrefix = "mail:daily:"

// reserveDailySlotScript — атомарный INCR + EXPIREAT; при превышении — DECR.
// KEYS[1]=key ARGV[1]=expireAtUnix ARGV[2]=limit → n или -1.
var reserveDailySlotScript = redis.NewScript(`
local n = redis.call('INCR', KEYS[1])
if n == 1 then
  redis.call('EXPIREAT', KEYS[1], ARGV[1])
end
if n > tonumber(ARGV[2]) then
  redis.call('DECR', KEYS[1])
  return -1
end
return n
`)

func dailySlotKey(email string, day string) string {
	return dailySlotKeyPrefix + day + ":" + email
}

func utcMidnightUnix(now time.Time) int64 {
	utc := now.UTC()
	next := time.Date(utc.Year(), utc.Month(), utc.Day()+1, 0, 0, 0, 0, time.UTC)
	return next.Unix()
}

// Reserve — Redis INCR; Redis err → fail-closed (не открываем лимит).
func (r *RedisDailySlotStore) Reserve(ctx context.Context, email string, limit int) error {
	if r == nil || r.rdb == nil {
		return fmt.Errorf("mail daily slot: redis unavailable")
	}
	if limit < 1 {
		limit = 20
	}
	now := time.Now().UTC()
	day := now.Format("2006-01-02")
	key := dailySlotKey(email, day)
	n, err := reserveDailySlotScript.Run(ctx, r.rdb, []string{key}, utcMidnightUnix(now), limit).Int64()
	if err != nil {
		return fmt.Errorf("mail daily slot reserve: %w", err)
	}
	if n < 0 {
		return ErrDailyLimitExceeded
	}
	return nil
}

// Release — DECR после ошибки отправки (не ниже 0; KEEPTTL).
func (r *RedisDailySlotStore) Release(ctx context.Context, email string) error {
	if r == nil || r.rdb == nil {
		return fmt.Errorf("mail daily slot: redis unavailable")
	}
	day := time.Now().UTC().Format("2006-01-02")
	key := dailySlotKey(email, day)
	n, err := r.rdb.Decr(ctx, key).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return fmt.Errorf("mail daily slot release: %w", err)
	}
	if n < 0 {
		// KEEPTTL: не сбрасываем EXPIREAT до UTC midnight.
		if err := r.rdb.SetArgs(ctx, key, 0, redis.SetArgs{KeepTTL: true}).Err(); err != nil {
			return fmt.Errorf("mail daily slot release clamp: %w", err)
		}
	}
	return nil
}
